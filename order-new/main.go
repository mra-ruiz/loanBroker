// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"e-commerce-app/models"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	fmt.Println("Received a new order ...")

	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Fatal(c.StartReceiver(context.Background(), receive));
}

func receive( ctx context.Context, e cloudevents.Event ) {	
	var orders []models.Order

	err := json.Unmarshal(e.Data(), &orders)
	if err != nil {
		log.Fatalf("Couldn't unmarshal e.Data() into orders, %v", err)
	}

	for i := range orders {
		handler(ctx, orders, orders[i])
	}
}

// handler for the Lambda function
func handler(ctx context.Context, orders []models.Order, ord models.Order) (models.Order, error) {

	log.Printf("[%s] - received new order", ord.OrderID)

	// persist the order data. Set order status to new
	ord.OrderStatus = "New"

	err := saveOrder(ctx, orders, ord)
	if err != nil {
		log.Printf("[%s] - error! %s", ord.OrderID, err.Error())
		return models.Order{}, models.NewErrProcessOrder(err.Error())
	}

	// testing scenario
	if ord.OrderID[0:1] == "1" {
		return models.Order{}, models.NewErrProcessOrder("Unable to process order " + ord.OrderID)
	}

	log.Printf("[%s] - order status set to new", ord.OrderID)

	return ord, nil
}

func saveOrder(ctx context.Context, orders []models.Order, order models.Order) error {
	
	// Adding order to a json file
	ordersBytes, err  := json.MarshalIndent(append(orders, order), "", "    ")
  	if err != nil {
		log.Fatalf("Couldn't marshal orders, %v", err)
	}

	// Modify the JSON file that acts as the database
	err = ioutil.WriteFile("./orders.json", ordersBytes, 0644)
  	if err != nil {
		log.Fatalf("Couldn't write to json file, %v", err)
	}
	
	return nil
}