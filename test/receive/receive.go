package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	fmt.Println("Hi! I am going to receive CloudEvents :)")

	// The default client is HTTP.
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	log.Fatal(c.StartReceiver(context.Background(), receive));
}

func receive(event cloudevents.Event) {
	// do something with event. Printing for now
    fmt.Printf("%s", event)
	fmt.Printf("%s", event.Context)
}


// ####### Not using the marshal and unmarshal yet ######

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

