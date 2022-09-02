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

    // Receive new order
    var	neworder models.StoredOrder
    err = json.Unmarshal(body, &neworder)
    if err != nil {
        msg := fmt.Sprintf("Failed to unmarshal body: %v", err)
        log.Println(msg)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }

    log.Printf("[%s] - received new order", neworder.OrderID)

    // persist the order data. Set order status to new
    neworder.Order.OrderStatus = "New"

    // Store the new order in the database
    err = saveOrder(neworder)
    if err != nil {
        msg := fmt.Sprintf("Could not save order: %v", err)
        log.Println(msg)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }

    log.Printf("[%s] - order status set to new", neworder.OrderID)
}

func saveOrder(neworder models.StoredOrder) error {	
    // Converting the new order's order id into a byte slice
    orderIdBytes, err := json.Marshal(neworder.OrderID)
    if err != nil {
        msg := fmt.Sprintf("Error with Marshall() in saveOrder(): Could not marshall order id: %v", err)
        log.Println(msg)
        return errors.New(msg)
    }
    // Converting the new order's order info into a byte slice
    orderInfoBytes, err := json.Marshal(neworder.Order)
    if err != nil {
        msg := fmt.Sprintf("Error with Marshall() in saveOrder(): Could not marshall order: %v", err)
        log.Println(msg)
        return errors.New(msg)
    }
    // Inserting the new order into the database
    insertCommand := `INSERT INTO stored_orders (order_id, order_info) VALUES ($1, $2)`
    _, err = db.Exec(insertCommand, orderIdBytes, orderInfoBytes)
    if err != nil {
        msg := fmt.Sprintf("Error with Exec() in saveOrder(): Could not insert new order into database: %v", err)
        log.Println(msg)
        return errors.New(msg)
    }

    return nil
}