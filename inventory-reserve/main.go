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
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("Starting inventory reserve ...")
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		fmt.Printf("Failed to create client: %v", err)
		os.Exit(1)
	}
	log.Fatal(c.StartReceiver(context.Background(), receive));
}

func receive(ctx context.Context, e cloudevents.Event) error {
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
			fmt.Printf("receive(): Error when handler() is called: %v", err)
			return fmt.Errorf("receive(): Error when handler() is called: %w", err)
		}
	}

	return nil
}

func handler(ctx context.Context, storedOrder models.StoredOrder, db *sql.DB) (models.StoredOrder, error) {
	log.Printf("[%s] - processing inventory reservation", storedOrder.OrderID)

	var newInvTrans = models.Inventory{
		OrderID:    storedOrder.OrderID,
		OrderItems: storedOrder.Order.ItemIds(),
	}

	// reserve the items in the inventory
	newInvTrans.Reserve()

	// Annotate saga with inventory transaction id
	storedOrder.Order.Inventory = newInvTrans

	// Save the reservation
	err := saveInventory(ctx, newInvTrans, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return models.StoredOrder{}, models.NewErrReserveInventory(err.Error())
	}

	log.Printf("[%s] - reservation processed", storedOrder.OrderID)

	fmt.Println("\nUpdated stored orders:")
	utils.ViewDatabase(db)

	// Only for restoring database for testing reasons
	// err = utils.ResetOrderInventory(db, storedOrder.OrderID)
	// if err != nil {
	// 	fmt.Printf("handler(): Error with ResetOrderInventory(): %v", err)
	// 	return models.StoredOrder{}, fmt.Errorf("handler(): Error with ResetOrderInventory(): %w", err)
	// }
	// fmt.Println("\nStored orders after reset:")
	// err = utils.ViewDatabase(db)
	// if err != nil {
	// 	fmt.Printf("handler(): Error with ViewDatabase(): %v", err)
	// 	return models.StoredOrder{}, fmt.Errorf("handler(): Error with ViewDatabase(): %w", err)
	// }

	// close database
	defer db.Close()
	return storedOrder, nil
}

func saveInventory(ctx context.Context, inventory models.Inventory, db *sql.DB) error {
	// converting Inventory into a byte slice
	inventoryBytes, err := json.Marshal(inventory)
	if err != nil {
		fmt.Printf("saveInventory(): Could not marshal inventory: %v", err)
		return fmt.Errorf("saveInventory(): Could not marshal inventory: %w", err)
	}

	// Updating inventory of specific order
	updateString := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{inventory}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
	_, err = db.Exec(updateString, inventoryBytes, inventory.OrderID)
	if err != nil {
		fmt.Printf("saveInventory(): Could not update inventory: %v", err)
		return fmt.Errorf("saveInventory(): Could not update inventory: %w", err)
	}

	return nil
}