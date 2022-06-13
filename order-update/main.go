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
	fmt.Println("Received request to update order status ...")
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

func handler(ctx context.Context, storedOrder models.StoredOrder, db *sql.DB) (models.Order, error) {

	log.Printf("[%s] - received request to update order status", storedOrder.OrderID)

	order, err := getOrder(ctx, storedOrder.OrderID, db)

	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return storedOrder.Order, models.NewErrUpdateOrderStatus(err.Error())
	}

	// Set order to status to "pending"
	order.OrderStatus = "Pending"

	err = saveOrder(ctx, order, storedOrder.OrderID, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return storedOrder.Order, models.NewErrUpdateOrderStatus(err.Error())
	}

	log.Printf("[%s] - order status updated to pending", storedOrder.OrderID)

	fmt.Println("\nUpdated stored orders:")
	utils.ViewDatabase(db)

	// Only for restoring database for testing reasons
	// utils.ResetOrderStatus(db, storedOrder.OrderID)
	// fmt.Println("\nStored orders after reset:")
	// utils.ViewDatabase(db)

	// close database
	defer db.Close()
	return storedOrder.Order, nil
}

// getOrder retrieves a specified from DynamoDB and marshals it to a Order type
func getOrder(ctx context.Context, orderID string, db *sql.DB) (models.Order, error) {
	// Searching for order
	resultingOrder, err := db.Query(`select order_info from stored_orders where order_id = $1;`, orderID)
	utils.CheckForErrors(err, "Could not search for order")

	// Convert order of type JSONB to type models.StoredOrder.Order
	var order models.Order
	var result []uint8

	for resultingOrder.Next() {
		resultingOrder.Scan(&order)
		json.Unmarshal(result, &order)		
	}
	
	return order, nil
}

func saveOrder(ctx context.Context, order models.Order, orderId string, db *sql.DB) error {
	// converting order into a byte slice
	orderStatusBytes, err := json.Marshal(order.OrderStatus)
	utils.CheckForErrors(err, "Could not marshall order status")

	// Updating order of specific order
	updateString := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{order_status}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
	_, err = db.Exec(updateString, orderStatusBytes, orderId)
	utils.CheckForErrors(err, "Could not update order")

	return nil
} 