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
	fmt.Println("Starting credit card refund ...")
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

	log.Printf("[%s] - processing refund", storedOrder.OrderID)

	// find Payment transaction for this order
	payment, err := getTransaction(ctx, storedOrder.OrderID, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return storedOrder.Order, models.NewErrProcessRefund(err.Error())
	}

	// process the refund for the order
	payment.Refund()

	// write to database.
	err = saveTransaction(ctx, payment, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return storedOrder.Order, models.NewErrProcessRefund(err.Error())
	}

	// save state
	storedOrder.Order.Payment = payment

	log.Printf("[%s] - refund processed", storedOrder.OrderID)

	fmt.Println("\nUpdated stored orders:")
	utils.ViewDatabase(db)

	// Only for restoring database for testing reasons
	// utils.ResetOrderPayment(db, storedOrder.OrderID)
	// fmt.Println("\nStored orders after reset:")
	// utils.ViewDatabase(db)

	// close database
	defer db.Close()
	return storedOrder.Order, nil
}

// returns a specified payment transaction from the database
func getTransaction(ctx context.Context, orderID string, db *sql.DB) (models.Payment, error) {
	// Searching for payment
	resultingPayment, err := db.Query(`select order_info -> 'payment' from stored_orders where order_id = $1;`, orderID)
	utils.CheckForErrors(err, "Could not search for payment")

	// Convert payment of type JSONB to type models.Payment
	var payment models.Payment
	var result []uint8

	for resultingPayment.Next() {
		resultingPayment.Scan(&result)
		json.Unmarshal(result, &payment)		
	}
	
	return payment, nil
}

// saves refund transaction to the database
func saveTransaction(ctx context.Context, payment models.Payment, db *sql.DB) error {
	// converting payment into a byte slice
	paymentBytes, err := json.Marshal(payment)
	utils.CheckForErrors(err, "Could not marshall payment")

	// Updating payment of specific order
	updateString := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{payment}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
	_, err = db.Exec(updateString, paymentBytes, payment.OrderID)
	utils.CheckForErrors(err, "Could not update payment")

	return nil
}