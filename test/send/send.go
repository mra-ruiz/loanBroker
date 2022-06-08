package main

import (
	"context"
	"database/sql"
	"e-commerce-app/models"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	fmt.Println("Hi! I am going to send a CloudEvent :)")

	db, err := connectDatabase()

	var allStoredOrders = importDbData(db)

	// Create client
	c, err := cloudevents.NewClientHTTP()
	checkForErrors(err, "Failed to create client")

	// Create an Event.
	event :=  cloudevents.NewEvent()
	event.SetSource("example/uri")
	event.SetType("example.type")
	event.SetData(cloudevents.ApplicationJSON, &allStoredOrders)

	// Set a target.
	ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8080/")

	// Send that Event.
	if result := c.Send(ctx, event); cloudevents.IsUndelivered(result) {
		log.Fatalf("failed to send, %v", result)
	}
}

func connectDatabase() (*sql.DB, error) {
	// connection string
	host := "localhost"
    port := 5432
    user := "mruizcardenas"
    password := "K67u5ye"
    dbname := "postgres"

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	
	// open database
	db, err := sql.Open("postgres", psqlconn)
	checkForErrors(err, "Could not open database")

	// check db
    err = db.Ping()
	checkForErrors(err, "Could not ping database")
	fmt.Println("Connected to databse!")
	return db, err
}

func importDbData(db *sql.DB) []models.StoredOrder {
	var allStoredOrders []models.StoredOrder
	var storedOrder models.StoredOrder
	rows, err := db.Query(`SELECT * FROM stored_orders`)

	checkForErrors(err, "send: Could not query select * from stored_orders")

	for rows.Next() {
		if err = rows.Scan(&storedOrder.OrderID, &storedOrder.Order); err != nil {
			checkForErrors(err, "Error with scan")
		} else {
			fmt.Println("Here's where scan has no error")
		}
		allStoredOrders = append(allStoredOrders, storedOrder)
		fmt.Println(storedOrder)
		fmt.Println(allStoredOrders)
	}

	// Only for restoring database for testing reasons
	// resetDatabase(db)

	// Close database
	defer rows.Close()
	return allStoredOrders
}

func resetDatabase(db *sql.DB) {
	// Resetting after inventory-reserve
	originalInventory := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{inventory}', '{
		"transaction_id": "transactionID7845764", 
		"transaction_date": "01-1-2022", 
		"order_id": "orderID123456", 
		"items": [
			"Pencil", 
			"Paper"
		], 
		"transaction_type": "online"
	}', true) WHERE order_id = 'orderID123456';`

	_, err := db.Exec(originalInventory)
	checkForErrors(err, "Could not reset database")
}

func checkForErrors(err error, s string) {
	if err != nil {
		fmt.Printf("%v\n", err)
		log.Fatalf(s)
	}
}