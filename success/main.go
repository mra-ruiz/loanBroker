package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"e-commerce-app/models"
)

func main() {
	log.Printf("success function called...")
    http.HandleFunc("/", handler)
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}

func handler(w http.ResponseWriter, req *http.Request) {

    log.Printf("success: In handler()...")
	// time.Sleep(10*time.Second)

    body, err := io.ReadAll(req.Body)
    if err != nil {
        msg := fmt.Sprintf("Failed to read the request body: %v", err)
		log.Printf("[%s]", msg)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }
	log.Printf("success: body read successfully...")
    defer req.Body.Close()
	log.Printf("success: defer body close done...")

    // Receive order
    var	neworder models.StoredOrder
    err = json.Unmarshal(body, &neworder)
    if err != nil {
        msg := fmt.Sprintf("Failed to unmarshal body: %v", err)
		log.Printf("[%s]", msg)
        log.Println(msg)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }

    log.Printf("[%s] - Workflow completed successfully!", neworder.OrderID)
    log.Printf("Sending CloudEvent...")

	// ############## CLOUD EVENT ##############
	// Create client
	// c, err := cloudevents.NewClientHTTP()
	// if err != nil {
	// 	msg := fmt.Sprintf("Failed to create client: %v", err)
	// 	w.Write([]byte(msg))
    //     w.WriteHeader(500)
	// 	return
	// }

	// // Create an Event.
	// event :=  cloudevents.NewEvent()
	// event.SetSource("example/uri")
	// event.SetType("example.type")
	// event.SetData(cloudevents.ApplicationJSON, &neworder)

	// // Set a target.
	// ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8080/")

	// // Send that Event.
	// if result := c.Send(ctx, event); cloudevents.IsUndelivered(result) {
	// 	log.Fatalf("failed to send, %v", result)
	// }

    // log.Printf("CloudEvent sent!")
}