package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"e-commerce-app/models" //local

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	fmt.Println("Starting inventory reserve ...")

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

	log.Printf("[%s] - processing inventory reservation", ord.OrderID)

	var newInvTrans = models.Inventory{
		OrderID:    ord.OrderID,
		OrderItems: ord.ItemIds(),
	}

	// reserve the items in the inventory
	newInvTrans.Reserve()

	// Annotate saga with inventory transaction id
	ord.Inventory = newInvTrans

	// Save the reservation
	err := saveInventory(ctx, orders, newInvTrans)
	if err != nil {
		log.Printf("[%s] - error! %s", ord.OrderID, err.Error())
		return models.Order{}, models.NewErrReserveInventory(err.Error())
	}

	log.Printf("[%s] - reservation processed", ord.OrderID)

	return ord, nil
}

func saveInventory(ctx context.Context, orders []models.Order, inventory models.Inventory) error {
	// Updating inventory of specific order
	for i:= 0; i < len(orders); i++ {
		if orders[i].OrderID == inventory.OrderID {
			orders[i].Inventory = inventory
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