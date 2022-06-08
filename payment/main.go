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
	fmt.Println("Starting payment processing ...")
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
	utils.ViewDatabase(db)

	// Only for restoring database for testing reasons
	// utils.ResetDatabase(db, "payment")
	// fmt.Println("\nStored orders after reset:")
	// utils.ViewDatabase(db)

	// close database
	defer db.Close()
	return storedOrder, nil
}

func savePayment(ctx context.Context, payment models.Payment, db *sql.DB) error {
	// converting payment into a byte slice
	paymentBytes, err := json.Marshal(payment)
	utils.CheckForErrors(err, "Could not marshall payment")

	// Updating payment of specific order
	updatePaymentCommand := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{payment}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
	_, err = db.Exec(updatePaymentCommand, paymentBytes, payment.OrderID)
	utils.CheckForErrors(err, "Could not update inventory")

	return nil
}