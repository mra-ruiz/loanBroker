package main

import (
	"context"
	"e-commerce-app/models"
	"fmt"
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
	event := createEvent()
	
	// Set a target.
	ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8080/")

	// Send that Event.
	send(c, ctx, event)
}

func createEvent()(event cloudevents.Event) {
	e :=  cloudevents.NewEvent()
	e.SetSource("example/uri")
	e.SetType("example.type")
	e.SetData(cloudevents.ApplicationJSON, models.Inventory{}) //inventory
	return e
}

func send(c cloudevents.Client, ctx context.Context, event cloudevents.Event) {
	if result := c.Send(ctx, event); cloudevents.IsUndelivered(result) {
		log.Fatalf("failed to send, %v", result)
	}
}