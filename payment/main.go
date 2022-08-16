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
	fmt.Println("Starting payment processing ...")
	// c, err := cloudevents.NewClientHTTP()
	// if err != nil {
	// 	fmt.Printf("Failed to create client: %v", err)
	// 	os.Exit(1)
	// }
	// log.Fatal(c.StartReceiver(context.Background(), receive));

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

	log.Printf("[%s] - processing payment", storedOrder.OrderID)

	var payment = models.Payment{
		OrderID:       storedOrder.OrderID,
		MerchantID:    "merch1",
		PaymentAmount: storedOrder.Order.Total(),
	}

	// Process payment
	payment.Pay()

	// Save payment
	err := savePayment(payment, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return storedOrder, models.NewErrProcessPayment(err.Error())
	}

	// Save state
	storedOrder.Order.Payment = payment

	log.Printf("[%s] - payment processed", storedOrder.OrderID)
	resultStoredOrder := storedOrder

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
	return resultStoredOrder, nil
}

func savePayment(payment models.Payment, db *sql.DB) error {
	// converting payment into a byte slice
	paymentBytes, err := json.Marshal(payment)
	if err != nil {
		fmt.Printf("Error with Marshall() in savePayment(): Could not marshall payment: %v", err)
		return fmt.Errorf("Error with Marshall() in savePayment(): Could not marshall payment: %w", err)
	}

	// Updating payment of specific order
	updatePaymentCommand := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{payment}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
	_, err = db.Exec(updatePaymentCommand, paymentBytes, payment.OrderID)
	if err != nil {
		fmt.Printf("Error with Exec() in saveOrder(): Could not update inventory: %v", err)
		return fmt.Errorf("Error with Exec() in saveOrder(): Could not update inventory: %w", err)
	}

	return nil
}