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

	err = saveOrder(neworder, db)
	if err != nil {
		msg := fmt.Sprintf("Could not save order: %v", err)
		log.Println(msg)
		w.Write([]byte(msg))
		w.WriteHeader(500) 
		return
	}

	log.Printf("[%s] - order status set to new", neworder.OrderID)
}

func saveOrder(neworder models.StoredOrder, db *sql.DB) error {	
	// Converting the new order status into a byte slice
	b, err := json.Marshal(neworder)
	if err != nil {
		msg := fmt.Sprintf("Error with Marshall() in saveOrder(): Could not marshall order status: %v", err)
		log.Println(msg)
		return errors.New(msg)
	}

	// Updating the order status in the database
	updateString := `
	  INSERT stored_orders (order_id, order_info) 
	  VALUES $1`
	_, err = db.Exec(updateString, b)
	if err != nil {
		msg := fmt.Sprintf("Error with Exec() in saveOrder(): Could not update order status to new: %v", err)
		log.Println(msg)
		return errors.New(msg)
	}

	return nil
}