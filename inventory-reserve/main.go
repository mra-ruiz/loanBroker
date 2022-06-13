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
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("Starting inventory reserve ...")
	c, err := cloudevents.NewClientHTTP()
	utils.CheckForErrors(err, "Failed to create client")
	log.Fatal(c.StartReceiver(context.Background(), receive));
}

func receive(ctx context.Context, e cloudevents.Event) {
	db, err := utils.ConnectDatabase()

	var allStoredOrders []models.StoredOrder

	err = json.Unmarshal(e.Data(), &allStoredOrders)
	utils.CheckForErrors(err, "Could not unmarshall e.Data() into type allStoredOrders")
	
	for i := range allStoredOrders {
		handler(ctx, allStoredOrders[i], db)
	}
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
	utils.ResetOrderInventory(db, storedOrder.OrderID)
	fmt.Println("\nStored orders after reset:")
	utils.ViewDatabase(db)

	// close database
	defer db.Close()
	return storedOrder, nil
}

func saveInventory(ctx context.Context, inventory models.Inventory, db *sql.DB) error {
	// converting Inventory into a byte slice
	inventoryBytes, err := json.Marshal(inventory)
	utils.CheckForErrors(err, "Could not marshall inventory")

	// Updating inventory of specific order
	updateString := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{inventory}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
	_, err = db.Exec(updateString, inventoryBytes, inventory.OrderID)
	utils.CheckForErrors(err, "Could not update inventory")

	return nil
}