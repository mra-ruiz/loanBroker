package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	fmt.Println("Hi! I am going to send and receive CloudEvents :)")
	sendFirstCloudEvent()
	receiveFirstCloudEvent()
}

func sendFirstCloudEvent() {
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	// Create an Event.
	event :=  cloudevents.NewEvent()
	event.SetSource("example/uri")
	event.SetType("example.type")
	event.SetData(cloudevents.ApplicationJSON, map[string]string{"hello": "world"})

	// Set a target.
	ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8080/")

	// Send that Event.
	if result := c.Send(ctx, event); cloudevents.IsUndelivered(result) {
		log.Fatalf("failed to send, %v", result)
	}
}


func receive(event cloudevents.Event) {
	// do something with event.
    fmt.Printf("%s", event)
}

func receiveFirstCloudEvent() {
	// The default client is HTTP.
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	log.Fatal(c.StartReceiver(context.Background(), receive));
}

func marshalCloudEvent()(b []byte, e error) {
	event := cloudevents.NewEvent()
	event.SetID("example-uuid-32943bac6fea")
	event.SetSource("example/uri")
	event.SetType("example.type")
	event.SetData(cloudevents.ApplicationJSON, map[string]string{"hello": "world"})

	bytes, err := json.Marshal(event)

	return bytes, err
}

func unmarshalCloudEvent(bytes []byte)(e error) {
	event :=  cloudevents.NewEvent()

	err := json.Unmarshal(bytes, &event)

	return err
}