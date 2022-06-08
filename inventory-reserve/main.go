package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"e-commerce-app/models"
	"e-commerce-app/test/send"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("Starting inventory reserve ...")
	c, err := cloudevents.NewClientHTTP()
	send.CheckForErrors(err, "Failed to create client")
	log.Fatal(c.StartReceiver(context.Background(), receive));
}

func receive(ctx context.Context, e cloudevents.Event) {
	db, err := send.ConnectDatabase()

	var allStoredOrders []models.StoredOrder

	err = json.Unmarshal(e.Data(), &allStoredOrders)
	send.CheckForErrors(err, "Could not unmarshall e.Data() into type allStoredOrders")
	
	for i := range allStoredOrders {
		handler(ctx, allStoredOrders, allStoredOrders[i], db)
	}
}

func handler(ctx context.Context, allStoredOrders []models.StoredOrder, storedOrder models.StoredOrder, db *sql.DB) (models.StoredOrder, error) {
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
	err := saveInventory(ctx, allStoredOrders, newInvTrans, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return models.StoredOrder{}, models.NewErrReserveInventory(err.Error())
	}

	log.Printf("[%s] - reservation processed", storedOrder.OrderID)

	return storedOrder, nil
}

func saveInventory(ctx context.Context, allStoredOrders []models.StoredOrder, inventory models.Inventory, db *sql.DB) error {
	fmt.Println("in saveInventory() function")

	// converting Inventory into a byte slice
	inventoryBytes, err := json.Marshal(inventory)

	// Updating inventory of specific order
	updateString := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{inventory}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
	_, err = db.Exec(updateString, inventoryBytes, inventory.OrderID)
	send.CheckForErrors(err, "Could not update inventory")

	// Only for restoring database for testing reasons
	// resetDatabase(db)

	// close database
    defer db.Close()
	return nil
}

func resetDatabase(db *sql.DB) {
	// Resetting after inventory-reserve
	originalInventory := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{inventory}', '{
		"transaction_id": "transactionID7845764", 
		"transaction_date": "01-1-2022", 
		"order_id": "orderID123456", 
		"items": [
			"Pencil", 
			"Paper"
		], 
		"transaction_type": "online"
	}', true) WHERE order_id = 'orderID123456';`

	_, err := db.Exec(originalInventory)
	send.CheckForErrors(err, "Could not reset database")
}