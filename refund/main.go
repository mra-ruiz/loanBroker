package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"e-commerce-app/models"
	"e-commerce-app/utils"
)

func main() {
	fmt.Println("Starting credit card refund ...")
	http.HandleFunc("/", prepareData)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func prepareData(w http.ResponseWriter, req *http.Request) {
	// reading headers
	for k, v := range req.Header {
		fmt.Printf("%s=%s\n", k, v[0])
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Printf("Error with io.ReadAll() in prepareData(): : %v", err)
		log.Println(err)
		return
	}
	defer req.Body.Close()
	fmt.Println("#####################################")
	fmt.Println(string(body))

	db, err := utils.ConnectDatabase()
	if err != nil {
		fmt.Printf("In prepareData(): Could not connect to database: %v", err)
		return
	}

	allStoredOrders := utils.ImportDbData(db)
	var returned_stored_order models.StoredOrder
	var bytesRetSto []byte

	for i := range allStoredOrders {
		returned_stored_order, err = handler(allStoredOrders[i], db)
		if err != nil {
			fmt.Printf("Error with handler() in prepareData(): : %v", err)
			return
		}
		// converting returned stored order into a byte slice
		bytesRetSto, err = json.Marshal(returned_stored_order)
		fmt.Printf("returned stored order:\n%v", returned_stored_order)
		w.Write(bytesRetSto)
	}
}

func handler(storedOrder models.StoredOrder, db *sql.DB) (models.StoredOrder, error) {

	fmt.Println("#####################################")

	log.Printf("[%s] - processing refund", storedOrder.OrderID)

	// Find Payment transaction for this order
	payment, err := getTransaction(storedOrder.OrderID, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return storedOrder, models.NewErrProcessRefund(err.Error())
	}

	// Process the refund for the order
	payment.Refund()

	// Saves refunded transaction to the database
	err = saveTransaction(payment, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return storedOrder, models.NewErrProcessRefund(err.Error())
	}

	// save state
	storedOrder.Order.Payment = payment

	fmt.Println()
	log.Printf("[%s] - refund processed", storedOrder.OrderID)

	fmt.Println("\nUpdated stored orders:")
	err = utils.ViewDatabase(db)
	if err != nil {
		fmt.Printf("Error with ViewDatabase() in handler(): %v", err)
		return models.StoredOrder{}, fmt.Errorf("Error with ViewDatabase() in handler(): %w", err)
	}

	// Only for restoring database for testing reasons
	err = utils.ResetOrderPayment(db, storedOrder.OrderID)
	if err != nil {
		fmt.Printf("Error with ResetOrderPayment() in handler(): %v", err)
		return models.StoredOrder{}, fmt.Errorf("Error with ResetOrderPayment() in handler(): %w", err)
	}
	fmt.Println("\nStored orders after reset:")
	err = utils.ViewDatabase(db)
	if err != nil {
		fmt.Printf("Error with ViewDatabase() in handler(): %v", err)
		return models.StoredOrder{}, fmt.Errorf("Error with ViewDatabase() in handler(): %w", err)
	}

	// close database
	defer db.Close()
	return storedOrder, nil
}

func getTransaction(orderID string, db *sql.DB) (models.Payment, error) {
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

func saveTransaction(payment models.Payment, db *sql.DB) error {
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
		fmt.Printf("Error with Exec() in saveTransaction(): Could not update payment: %v", err)
		return fmt.Errorf("Error with Exec() in saveTransaction(): Could not update payment: %w", err)
	}

	return nil
}