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
	if err != nil {
		_ = fmt.Errorf("Failed to create client: %w", err)
	}
	log.Fatal(c.StartReceiver(context.Background(), receive));
}

func receive( ctx context.Context, e cloudevents.Event ) error {	
	db, err := utils.ConnectDatabase()

	var allStoredOrders []models.StoredOrder

	err = json.Unmarshal(e.Data(), &allStoredOrders)
	if err != nil {
		return fmt.Errorf("Could not unmarshall e.Data() into type allStoredOrders: %w", err)
	}
	
	for i := range allStoredOrders {
		_, err = handler(ctx, allStoredOrders[i], db)
		if err != nil {
			return fmt.Errorf("Error connecting to database: %w", err)
		}
	}

	return nil
}

func handler(ctx context.Context, storedOrder models.StoredOrder, db *sql.DB) (models.StoredOrder, error) {

	log.Printf("[%s] - received new order", storedOrder.OrderID)

	// persist the order data. Set order status to new
	storedOrder.Order.OrderStatus = "New"

	err := saveOrder(ctx, storedOrder, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return models.StoredOrder{}, models.NewErrProcessOrder(err.Error())
	}

	log.Printf("[%s] - order status set to new", storedOrder.OrderID)

	fmt.Println("\nUpdated stored orders:")
	utils.ViewDatabase(db)

	// Only for restoring database for testing reasons
	// utils.ResetOrderStatus(db, storedOrder.OrderID)
	// fmt.Println("\nStored orders after reset:")
	// utils.ViewDatabase(db)

	// close database
	defer db.Close()
	return storedOrder, nil
}

func saveOrder(ctx context.Context, updatedOrder models.StoredOrder, db *sql.DB) error {
	
	// Converting the new order status into a byte slice
	newStatus := updatedOrder.Order.OrderStatus
	orderStatusBytes, err := json.Marshal(newStatus)
	if err != nil {
		return fmt.Errorf("Could not marshall order status: %w", err)
	}

	// Updating the order status in the database
	updateString := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{order_status}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
	_, err = db.Exec(updateString, orderStatusBytes, updatedOrder.OrderID)
	if err != nil {
		return fmt.Errorf("Could not update order status to new: %w", err)
	}

	return nil
}