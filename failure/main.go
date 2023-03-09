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
    http.HandleFunc("/", handler)
    if err := http.ListenAndServe(":8080", nil); err != nil {
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

    // Receive order
    var	neworder models.StoredOrder
    err = json.Unmarshal(body, &neworder)
    if err != nil {
        msg := fmt.Sprintf("Failed to unmarshal body: %v", err)
        log.Println(msg)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }

    log.Printf("[%s] - Workflow failed", neworder.OrderID)
}