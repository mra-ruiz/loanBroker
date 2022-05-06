package main

import (
	"context"
	"e-commerce-app/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	fmt.Println("Hi! I am going to send CloudEvents :)")

	// Create client
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	// Create an Event.
	event :=  cloudevents.NewEvent()
	event.SetSource("example/uri")
	event.SetType("example.type")

	var orders = getOrders()
	event.SetData(cloudevents.ApplicationJSON, &orders)
	
	// Set a target.
	ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8080/")

	// Send that Event.
	if result := c.Send(ctx, event); cloudevents.IsUndelivered(result) {
		log.Fatalf("failed to send, %v", result)
	}
}

func getOrders()(orders []models.Order) {
	// Reading order from JSON file
	fileBytes, err := ioutil.ReadFile("./orders.json")
	if err != nil {
		log.Fatalf("failed to read file, %v", err)
	}

	// Unmarshaling json order slice to the newOrder object
	err = json.Unmarshal(fileBytes, &orders)
	if err != nil {
		log.Fatalf("failed to unmarshal fileBytes into orders, %v", err)
	}

	return orders
}