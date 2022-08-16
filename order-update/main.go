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
	fmt.Println("Received request to update order status ...")
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

	log.Printf("[%s] - received request to update order status", storedOrder.OrderID)

	// Find order in database
	order, err := getOrder(storedOrder.OrderID, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return storedOrder, models.NewErrUpdateOrderStatus(err.Error())
	}

	// Set order to status to "pending"
	order.OrderStatus = "Pending"

	// Saves order and updates order status to 'Pending'
	err = saveOrder(order, storedOrder.OrderID, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return storedOrder, models.NewErrUpdateOrderStatus(err.Error())
	}

	fmt.Println()
	log.Printf("[%s] - order status updated to pending", storedOrder.OrderID)

	fmt.Println("\nUpdated stored orders:")
	err = utils.ViewDatabase(db)
	if err != nil {
		fmt.Printf("Error with ViewDatabase() in handler(): %v", err)
		return models.StoredOrder{}, fmt.Errorf("Error with ViewDatabase() in handler(): %w", err)
	}

	// Only for restoring database for testing reasons
	err = utils.ResetOrderStatus(db, storedOrder.OrderID)
	if err != nil {
		fmt.Printf("Error with ResetOrderStatus() in handler(): %v", err)
		return models.StoredOrder{}, fmt.Errorf("Error with ResetOrderStatus() in handler(): %w", err)
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

func getOrder(orderID string, db *sql.DB) (models.Order, error) {
	// Searching for order
	resultingOrder, err := db.Query(`select order_info from stored_orders where order_id = $1;`, orderID)
	if err != nil {
		fmt.Printf("getOrder(): Could not search for order %v", err)
		return models.Order{}, fmt.Errorf("getOrder(): Could not search for order %w", err)
	}

	// Convert order of type JSONB to type models.StoredOrder.Order
	var order models.Order
	var result []uint8

	for resultingOrder.Next() {
		resultingOrder.Scan(&order)
		json.Unmarshal(result, &order)		
	}
	
	return order, nil
}

func saveOrder(order models.Order, orderId string, db *sql.DB) error {
	// converting order into a byte slice
	orderStatusBytes, err := json.Marshal(order.OrderStatus)
	if err != nil {
		fmt.Printf("saveOrder(): Could not marshall order status: %v", err)
		return fmt.Errorf("saveOrder(): Could not marshall order status: %w", err)
	}

	// Updating order of specific order in database
	updateString := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{order_status}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
	_, err = db.Exec(updateString, orderStatusBytes, orderId)
	if err != nil {
		fmt.Printf("Error with Exec() in saveOrder(): Could not update order: %v", err)
		return fmt.Errorf("Error with Exec() in saveOrder(): Could not update order: %w", err)
	}

	return nil
} 