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
	fmt.Println("Starting credit card refund ...")

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

	log.Printf("[%s] - processing refund", ord.OrderID)

	// find Payment transaction for this order
	payment, err := getTransaction(ctx, orders, ord.OrderID)
	if err != nil {
		log.Printf("[%s] - error! %s", ord.OrderID, err.Error())
		return ord, models.NewErrProcessRefund(err.Error())
	}

	// process the refund for the order
	payment.Refund()

	// write to database.
	err = saveTransaction(ctx, orders, payment)
	if err != nil {
		log.Printf("[%s] - error! %s", ord.OrderID, err.Error())
		return ord, models.NewErrProcessRefund(err.Error())
	}

	// save state
	ord.Payment = payment

	log.Printf("[%s] - refund processed", ord.OrderID)

	return ord, nil
}

// returns a specified payment transaction from the database
func getTransaction(ctx context.Context, orders []models.Order, orderID string) (models.Payment, error) {

	payment := models.Payment{}

	for _,curOrder := range orders {
		if curOrder.OrderID == orderID {
			payment = curOrder.Payment
			break
		}
	}

	return payment, nil
}

// saves refund transaction to the database
func saveTransaction(ctx context.Context, orders []models.Order, payment models.Payment) error {
	// Updating inventory of specific order
	for i:= 0; i < len(orders); i++ {
		if orders[i].OrderID == payment.OrderID {
			orders[i].Payment = payment
			break
		}
	}
	
	ordersBytes, err  := json.Marshal(orders)
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