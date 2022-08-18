package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"e-commerce-app/models"
	"e-commerce-app/utils"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	fmt.Println("Starting inventory release ...")
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		fmt.Printf("Failed to create client: %v", err)
		os.Exit(1)
	}
	log.Fatal(c.StartReceiver(context.Background(), receive));
}

func receive( ctx context.Context, e cloudevents.Event ) error {	
	db, err := utils.ConnectDatabase()
	if err != nil {
		fmt.Printf("receive(): Could not connect to database: %v", err)
		return fmt.Errorf("receive(): Could not connect to database: %w", err)
	}

	var allStoredOrders []models.StoredOrder

	err = json.Unmarshal(e.Data(), &allStoredOrders)
	if err != nil {
		fmt.Printf("receive(): Could not unmarshall e.Data() into type allStoredOrders: %v", err)
		return fmt.Errorf("receive(): Could not unmarshall e.Data() into type allStoredOrders: %w", err)
	}

	for i := range allStoredOrders {
		_, err = handler(ctx, allStoredOrders[i], db)
		if err != nil {
			return fmt.Errorf("Error in handler(): %w", err)
		}
	}

	return nil
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
	err = utils.ViewDatabase(db)
	if err != nil {
		fmt.Printf("Error with ViewDatabase() in handler(): %v", err)
		return models.StoredOrder{}, fmt.Errorf("Error with ViewDatabase() in handler(): %w", err)
	}

	// Only for restoring database for testing reasons
	// err = utils.ResetOrderInventory(db, storedOrder.OrderID)
	// if err != nil {
	// 	fmt.Printf("Error with ResetOrderInventory() in handler(): %v", err)
	// 	return models.StoredOrder{}, fmt.Errorf("Error with ResetOrderInventory() in handler(): %w", err)
	// }
	// fmt.Println("\nStored orders after reset:")
	// err = utils.ViewDatabase(db)
	// if err != nil {
	// 	fmt.Printf("Error with ViewDatabase() in handler(): %v", err)
	// 	return models.StoredOrder{}, fmt.Errorf("Error with ViewDatabase() in handler(): %w", err)
	// }

	// close database
	defer db.Close()
	return storedOrder, nil
}

func getTransaction(ctx context.Context, orderID string, db *sql.DB) (models.Inventory, error) {
	// Searching for inventory
	resultingInventory, err := db.Query(`select order_info -> 'inventory' from stored_orders where order_id = $1;`, orderID)
	if err != nil {
		fmt.Printf("getTransaction(): Could not search for inventory %v", err)
		return models.Inventory{}, fmt.Errorf("getTransaction(): Could not search for inventory %w", err)
	}

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
	if err != nil {
		fmt.Printf("saveTransaction(): Could not marshall inventory: %v", err)
		return fmt.Errorf("saveTransaction(): Could not marshall inventory: %w", err)
	}
	
	// Updating inventory of specific order in database
	updateString := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{inventory}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
	_, err = db.Exec(updateString, inventoryBytes, inventory.OrderID)
	if err != nil {
		fmt.Printf("saveTransaction(): Could not update inventory: %v", err)
		return fmt.Errorf("saveTransaction(): Could not update inventory: %w", err)
	}

	return nil
}