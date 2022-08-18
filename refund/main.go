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
	fmt.Println("Starting credit card refund ...")
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

	log.Printf("[%s] - processing refund", storedOrder.OrderID)

	// find Payment transaction for this order
	payment, err := getTransaction(ctx, storedOrder.OrderID, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return storedOrder, models.NewErrProcessRefund(err.Error())
	}

	// process the refund for the order
	payment.Refund()

	// write to database.
	err = saveTransaction(ctx, payment, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return storedOrder, models.NewErrProcessRefund(err.Error())
	}

	// save state
	storedOrder.Order.Payment = payment

	log.Printf("[%s] - refund processed", storedOrder.OrderID)

	fmt.Println("\nUpdated stored orders:")
	utils.ViewDatabase(db)
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

// returns a specified payment transaction from the database
func getTransaction(ctx context.Context, orderID string, db *sql.DB) (models.Payment, error) {
	// Searching for payment
	resultingPayment, err := db.Query(`select order_info -> 'payment' from stored_orders where order_id = $1;`, orderID)
	if err != nil {
		fmt.Printf("getTransaction(): Could not search for payment %v", err)
		return models.Payment{}, fmt.Errorf("getTransaction(): Could not search for payment %w", err)
	}

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
	if err != nil {
		fmt.Printf("saveTransaction(): Could not marshall payment: %v", err)
		return fmt.Errorf("saveTransaction(): Could not marshall payment: %w", err)
	}

	// Updating payment of specific order
	updateString := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{payment}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
	_, err = db.Exec(updateString, paymentBytes, payment.OrderID)
	if err != nil {
		fmt.Printf("saveTransaction(): Could not update payment: %v", err)
		return fmt.Errorf("saveTransaction(): Could not update payment: %w", err)
	}

	return nil
}