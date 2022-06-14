package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"e-commerce-app/models"
	"e-commerce-app/utils"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	fmt.Println("Starting inventory release ...")
	c, err := cloudevents.NewClientHTTP()
	utils.CheckForErrors(err, "Failed to create client")
	log.Fatal(c.StartReceiver(context.Background(), receive));
}

func receive( ctx context.Context, e cloudevents.Event ) {	
	db, err := utils.ConnectDatabase()

	var allStoredOrders []models.StoredOrder

	err = json.Unmarshal(e.Data(), &allStoredOrders)
	utils.CheckForErrors(err, "Could not unmarshall e.Data() into type allStoredOrders")
	
	for i := range allStoredOrders {
		handler(ctx, allStoredOrders[i], db)
	}
}

func handler(ctx context.Context, storedOrder models.StoredOrder, db *sql.DB) (models.StoredOrder, error) {
	
	log.Printf("[%s] - processing inventory release", storedOrder.OrderID)
	
	// Find inventory transaction
	inventory, err := getTransaction(ctx, storedOrder.OrderID, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return storedOrder, models.NewErrReleaseInventory(err.Error())
	}

	// Releasing items from inventory to make it available
	inventory.Release()

	// Saves transaction and updates inventory TransactionType to 'Release' 
	err = saveTransaction(ctx, inventory, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return storedOrder, models.NewErrReleaseInventory(err.Error())
	}

	storedOrder.Order.Inventory = inventory

	fmt.Println()
	log.Printf("[%s] - reservation processed", storedOrder.OrderID)

	fmt.Println("\nUpdated stored orders:")
	utils.ViewDatabase(db)

	// Only for restoring database for testing reasons
	// utils.ResetOrderInventory(db, storedOrder.OrderID)
	// fmt.Println("\nStored orders after reset:")
	// utils.ViewDatabase(db)

	// close database
	defer db.Close()
	return storedOrder, nil
}

func getTransaction(ctx context.Context, orderID string, db *sql.DB) (models.Inventory, error) {
	// Searching for inventory
	resultingInventory, err := db.Query(`select order_info -> 'inventory' from stored_orders where order_id = $1;`, orderID)
	utils.CheckForErrors(err, "Could not search for inventory")

	// Convert inventory of type JSONB to type models.Inventory
	var inventory models.Inventory
	var result []uint8

	for resultingInventory.Next() {
		resultingInventory.Scan(&result)
		json.Unmarshal(result, &inventory)		
	}
	
	return inventory, nil
}

func saveTransaction(ctx context.Context, inventory models.Inventory, db *sql.DB) error {
	// converting Inventory into a byte slice
	inventoryBytes, err := json.Marshal(inventory)
	utils.CheckForErrors(err, "Could not marshall inventory")

	// Updating inventory of specific order
	updateString := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{inventory}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
	_, err = db.Exec(updateString, inventoryBytes, inventory.OrderID)
	utils.CheckForErrors(err, "Could not update inventory")

	return nil
}