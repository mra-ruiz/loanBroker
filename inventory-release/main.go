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
	fmt.Println("Starting inventory release ...")

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
	fmt.Println()
	log.Printf("[%s] - processing inventory release", ord.OrderID)
	
	// Find inventory transaction
	inventory, err := getTransaction(ctx, orders, ord.OrderID)
	if err != nil {
		log.Printf("[%s] - error! %s", ord.OrderID, err.Error())
		return ord, models.NewErrReleaseInventory(err.Error())
	}

	fmt.Println("\nInventory after getTransaction(): \n", inventory)

	// Releasing items from inventory to make it available
	inventory.Release()

	fmt.Println("\nInventory after Release(): \n", inventory)

	// Saves transaction and updates inventory TransactionType to 'Release' 
	err = saveTransaction(ctx, orders, inventory)
	if err != nil {
		log.Printf("[%s] - error! %s", ord.OrderID, err.Error())
		return ord, models.NewErrReleaseInventory(err.Error())
	}

	ord.Inventory = inventory

	fmt.Println()
	log.Printf("[%s] - reservation processed", ord.OrderID)

	return ord, nil
}

func getTransaction(ctx context.Context, orders []models.Order, orderID string) (models.Inventory, error) {

	inventory := models.Inventory{}

	for _,curOrder := range orders {
		if curOrder.OrderID == orderID {
			inventory = curOrder.Inventory
			break
		}
	}
	
	return inventory, nil
}

func saveTransaction(ctx context.Context, orders []models.Order, inventory models.Inventory) error {
	// Updating inventory of specific orderw
	for i:= 0; i < len(orders); i++ {
		if orders[i].OrderID == inventory.OrderID {
			orders[i].Inventory = inventory
			break
		}
	}
	// MarshalIndent is just for debugging. Change back to Marshal()
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