package main

import (
	"context"
	"database/sql"
	"e-commerce-app/models"
	"e-commerce-app/utils"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	fmt.Println("Hi! I am going to send a CloudEvent :)")

	db, err := utils.ConnectDatabase()
	if err != nil {
		_ = fmt.Errorf("Could not connect to database: %w", err)
		os.Exit(1)
	}

	var allStoredOrders = importDbData(db)

	// Create client
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		_ = fmt.Errorf("Failed to create client: %w", err)
		os.Exit(1)
	}


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

func importDbData(db *sql.DB) []models.StoredOrder {
	var allStoredOrders []models.StoredOrder
	var storedOrder models.StoredOrder
	rows, err := db.Query(`SELECT * FROM stored_orders`)
	if err != nil {
		_ = fmt.Errorf("send: Could not query select * from stored_orders: %w", err)
		return nil
	}

	for rows.Next() {
		if err = rows.Scan(&storedOrder.OrderID, &storedOrder.Order); err != nil {
			if err != nil {
				_ = fmt.Errorf("ImportDBData(): Error with scan: %w", err)
				return nil
			}
		} else {
			// fmt.Println("Here's where scan has no error")
		}
		allStoredOrders = append(allStoredOrders, storedOrder)
		fmt.Println("Original stored orders:")
		fmt.Println(allStoredOrders)
	}

	// Close database
	defer rows.Close()
	return allStoredOrders
}