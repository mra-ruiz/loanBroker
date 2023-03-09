package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"e-commerce-app/models"
	"e-commerce-app/utils"
)

var (
    db *sql.DB
)
 
func main() {
	log.Printf("log: Order new function called :)")

    connectDb()

    http.HandleFunc("/", handler)
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}

func connectDb() {
    var err error
    db, err = utils.ConnectDatabase()
    if err != nil {
        fmt.Printf("Could not connect to database: %v", err)
        log.Fatal(err)
    }
}

func handler(w http.ResponseWriter, req *http.Request) {
	log.Printf("log: handler function called :)")

    body, err := io.ReadAll(req.Body)
    if err != nil {
        msg := fmt.Sprintf("handler(): Failed to read the request body: %v", err)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }
    defer req.Body.Close()

    // Receive request to update order status
    var	order models.StoredOrder
    err = json.Unmarshal(body, &order)
    if err != nil {
        msg := fmt.Sprintf("hadnler(): Failed to unmarshal body: %v", err)
        log.Println(msg)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }

    log.Printf("[%s] - received request to update order status", order.OrderID)

    // Find order in database
    fetchedOrder, err := getOrder(order.OrderID)
    if err != nil {
        msg := fmt.Sprintf("handler(): Could not get order: %v", err)
        log.Println(msg)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }

    // Set order to status to "pending"
    fetchedOrder.OrderStatus = "Pending"

    // Saves order and updates order status to 'Pending'
    err = saveOrder(fetchedOrder, order.OrderID)
    if err != nil {
        msg := fmt.Sprintf("handler(): Could not save order: %v", err)
        log.Println(msg)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }

    log.Printf("[%s] - order status updated to pending", order.OrderID)
	fmt.Fprintf(w, "[%s] - order status updated to pending", order.OrderID)
}

func getOrder(orderID string) (models.Order, error) {
	log.Printf("log: get order function called :)")

    // Searching for order
    resultingOrder, err := db.Query(`select order_info from stored_orders where order_id = $1;`, orderID)
    if err != nil {
        msg := fmt.Sprintf("Error with Query() in getOrder(): Could not get order: %v", err)
        log.Println(msg)
        return models.Order{}, errors.New(msg)
    }

    // Convert order of type JSONB to type models.StoredOrder.Order
    var order models.Order
    for resultingOrder.Next() {
        resultingOrder.Scan(&order)
    }
    
    return order, nil
}

func saveOrder(order models.Order, orderId string) error {
	log.Printf("log: save order function called :)")

    // converting order into a byte slice
    orderStatusBytes, err := json.Marshal(order.OrderStatus)
    if err != nil {
        msg := fmt.Sprintf("Error with Marshall() in saveOrder(): Could not marshall order: %v", err)
        log.Println(msg)
        return errors.New(msg)
    }

    // Updating order of specific order in database
    updateString := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{order_status}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
    _, err = db.Exec(updateString, orderStatusBytes, orderId)
    if err != nil {
        msg := fmt.Sprintf("Error with Exec() in saveOrder(): Could not update order status: %v", err)
        log.Println(msg)
        return errors.New(msg)
    }

    return nil
} 