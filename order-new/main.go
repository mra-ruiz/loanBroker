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
	fmt.Println("Received a new order ...")
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
		handler(ctx, allStoredOrders, allStoredOrders[i], db)
	}
}

func handler(ctx context.Context, allStoredOrders []models.StoredOrder, storedOrder models.StoredOrder, db *sql.DB) (models.StoredOrder, error) {

	log.Printf("[%s] - received new order", storedOrder.OrderID)

	// persist the order data. Set order status to new
	storedOrder.Order.OrderStatus = "New"

	err := saveOrder(ctx, allStoredOrders, storedOrder, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return models.StoredOrder{}, models.NewErrProcessOrder(err.Error())
	}

	log.Printf("[%s] - order status set to new", storedOrder.OrderID)

	fmt.Println("\nUpdated stored orders:")
	utils.ViewDatabase(db)

	// Only for restoring database for testing reasons
	// utils.ResetDatabase(db, "order-new")
	// fmt.Println("\nStored orders after reset:")
	// utils.ViewDatabase(db)

	// close database
	defer db.Close()
	return storedOrder, nil
}

func saveOrder(ctx context.Context, allStoredOrders []models.StoredOrder, updatedOrder models.StoredOrder, db *sql.DB) error {
	
	// Converting the new order status into a byte slice
	newStatus := updatedOrder.Order.OrderStatus
	orderStatusBytes, err := json.Marshal(newStatus)
	utils.CheckForErrors(err, "Could not marshall order status")

	// Updating the order status in the database
	updateString := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{order_status}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
	_, err = db.Exec(updateString, orderStatusBytes, updatedOrder.OrderID)
	utils.CheckForErrors(err, "Could not update order status to new")

	return nil
}