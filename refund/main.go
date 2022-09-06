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
    body, err := io.ReadAll(req.Body)
    if err != nil {
        msg := fmt.Sprintf("Failed to read the request body: %v", err)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }
    defer req.Body.Close()

    // Receive order to refund
    var	order models.StoredOrder
    err = json.Unmarshal(body, &order)
    if err != nil {
        msg := fmt.Sprintf("Failed to unmarshal body: %v", err)
        log.Println(msg)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }

    log.Printf("[%s] - processing refund", order.OrderID)

    // Find Payment transaction for this order
    fetchedPayment, err := getPayment(order.OrderID)
    if err != nil {
        msg := fmt.Sprintf("handler(): Could not get payment: %v", err)
        log.Println(msg)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }

    // Process the refund for the order
    fetchedPayment.Refund()

    // Saves refunded transaction to the database
    err = savePayment(fetchedPayment, order.OrderID)
    if err != nil {
        msg := fmt.Sprintf("handler(): Could not save payment: %v", err)
        log.Println(msg)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }

    log.Printf("[%s] - refund processed", order.OrderID)
}

func getPayment(orderID string) (models.Payment, error) {
    // Searching for order
    resultingOrder, err := db.Query(`select order_info from stored_orders where order_id = $1;`, orderID)
    if err != nil {
        msg := fmt.Sprintf("Error with Query() in getPayment(): Could not get order: %v", err)
        log.Println(msg)
        return models.Payment{}, errors.New(msg)
    }

    // Convert order of type JSONB to type models.StoredOrder.Order
    var order models.Order
    for resultingOrder.Next() {
        resultingOrder.Scan(&order)
    }

    return order.Payment, nil
}

func savePayment(payment models.Payment, orderID string) error {
    // converting payment into a byte slice
    paymentBytes, err := json.Marshal(payment)
    if err != nil {
        msg := fmt.Sprintf("Error with Marshall() in savePayment(): Could not marshall payment: %v", err)
        log.Println(msg)
        return errors.New(msg)
    }

    // Updating payment of specific order
    updateString := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{payment}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
    _, err = db.Exec(updateString, paymentBytes, orderID)
    if err != nil {
        msg := fmt.Sprintf("Error with Exec() in savePayment(): Could not update payment: %v", err)
        log.Println(msg)
        return errors.New(msg)
    }

    return nil
}