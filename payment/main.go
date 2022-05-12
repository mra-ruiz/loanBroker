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
	fmt.Println("Starting payment processing ...")

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

func handler(ctx context.Context, orders []models.Order, ord models.Order) (models.Order, error) {

	log.Printf("[%s] - processing payment", ord.OrderID)

	var payment = models.Payment{
		OrderID:       ord.OrderID,
		MerchantID:    "merch1",
		PaymentAmount: ord.Total(),
	}

	// Process payment
	payment.Pay()

	// Save payment
	err := savePayment(ctx, orders, payment)
	if err != nil {
		log.Printf("[%s] - error! %s", ord.OrderID, err.Error())
		return ord, models.NewErrProcessPayment(err.Error())
	}

	// Save state
	ord.Payment = payment

	// testing scenario
	if ord.OrderID[0:1] == "2" {
		return models.Order{}, models.NewErrProcessPayment("Unable to process payment for order " + ord.OrderID)
	}

	log.Printf("[%s] - payment processed", ord.OrderID)

	return ord, nil
}

func savePayment(ctx context.Context, orders []models.Order, payment models.Payment) error {
	// Updating inventory of specific order
	for i:= 0; i < len(orders); i++ {
		if orders[i].OrderID == payment.OrderID {
			orders[i].Payment = payment
			break
		}
	}
	
	ordersBytes, err  := json.MarshalIndent(orders, "", "    ")
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