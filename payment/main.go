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
	fmt.Println("Starting payment processing ...")
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		fmt.Printf("Failed to create client: %v", err)
		os.Exit(1)
	}
	log.Fatal(c.StartReceiver(context.Background(), receive));
}

func receive( ctx context.Context, e cloudevents.Event ) error{	
	db, err := utils.ConnectDatabase()
	if err != nil {
		fmt.Printf("receive() - Could not connect to database: %v", err)
		return fmt.Errorf("receive() - Could not connect to database: %w", err)
	}

	var allStoredOrders []models.StoredOrder

	err = json.Unmarshal(e.Data(), &allStoredOrders)
	if err != nil {
		fmt.Printf("Could not unmarshall e.Data() into type allStoredOrders: %v", err)
		return fmt.Errorf("Could not unmarshall e.Data() into type allStoredOrders: %w", err)
	}
	
	for i := range allStoredOrders {
		_, err = handler(ctx, allStoredOrders[i], db)
		if err != nil {
			fmt.Printf("Error in handler(): %v", err)
			return fmt.Errorf("Error in handler(): %w", err)
		}
	}
	return nil
}

func handler(ctx context.Context, storedOrder models.StoredOrder, db *sql.DB) (models.StoredOrder, error) {

	log.Printf("[%s] - processing payment", storedOrder.OrderID)

	var payment = models.Payment{
		OrderID:       storedOrder.OrderID,
		MerchantID:    "merch1",
		PaymentAmount: storedOrder.Order.Total(),
	}

	// Process payment
	payment.Pay()

	// Save payment
	err := savePayment(ctx, payment, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return storedOrder, models.NewErrProcessPayment(err.Error())
	}

	// Save state
	storedOrder.Order.Payment = payment

	log.Printf("[%s] - payment processed", storedOrder.OrderID)

	fmt.Println("\nUpdated stored orders:")
	err = utils.ViewDatabase(db)
	if err != nil {
		fmt.Printf("Error with ViewDatabase() in handler(): %v", err)
		return models.StoredOrder{}, fmt.Errorf("Error with ViewDatabase() in handler(): %w", err)
	}

	// Only for restoring database for testing reasons
	// err = utils.ResetOrderPayment(db, storedOrder.OrderID)
	// if err != nil {
	// 	fmt.Printf("Error with ResetOrderPayment() in handler(): %v", err)
	// 	return models.StoredOrder{}, fmt.Errorf("Error with ResetOrderPayment() in handler(): %w", err)
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

func savePayment(ctx context.Context, payment models.Payment, db *sql.DB) error {
	// converting payment into a byte slice
	paymentBytes, err := json.Marshal(payment)
	if err != nil {
		fmt.Printf("Could not marshall payment: %v", err)
		return fmt.Errorf("Could not marshall payment: %w", err)
	}

	// Updating payment of specific order
	updatePaymentCommand := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{payment}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
	_, err = db.Exec(updatePaymentCommand, paymentBytes, payment.OrderID)
	if err != nil {
		fmt.Printf("Could not update inventory: %v", err)
		return fmt.Errorf("Could not update inventory: %w", err)
	}

	return nil
}