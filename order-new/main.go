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
	fmt.Println("Received a new order ...")
	// c, err := cloudevents.NewClientHTTP()
	// if err != nil {
	// 	fmt.Printf("Failed to create client: %v", err)
	// 	os.Exit(1)
	// }

	// logger, err := zap.NewDevelopment()
	// if err != nil {
	// 	fmt.Printf("Logger error: %v", err)
	// 	os.Exit(1)
	// }

	// fmt.Println("end of main .. almost")
	// ctx := cecontext.WithLogger(context.Background(), logger.Sugar())
	// log.Fatal(c.StartReceiver(ctx, receive));

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

	log.Printf("[%s] - received new order", storedOrder.OrderID)

	// persist the order data. Set order status to new
	storedOrder.Order.OrderStatus = "New"

	err := saveOrder(storedOrder, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return models.StoredOrder{}, models.NewErrProcessOrder(err.Error())
	}

	fmt.Printf("[%s] - order status set to new", storedOrder.OrderID)
	resultStoredOrder := storedOrder

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
	return resultStoredOrder, nil
}

func saveOrder(updatedOrder models.StoredOrder, db *sql.DB) error {	
	// Converting the new order status into a byte slice
	newStatus := updatedOrder.Order.OrderStatus
	orderStatusBytes, err := json.Marshal(newStatus)
	if err != nil {
		fmt.Printf("Error with Marshall() in saveOrder(): Could not marshall order status: %v", err)
		return fmt.Errorf("Error with Marshall() in saveOrder(): Could not marshall order status: %w", err)
	}

	// Updating the order status in the database
	updateString := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{order_status}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
	_, err = db.Exec(updateString, orderStatusBytes, updatedOrder.OrderID)
	if err != nil {
		fmt.Printf("Error with Exec() in saveOrder(): Could not update order status to new: %v", err)
		return fmt.Errorf("Error with Exec() in saveOrder(): Could not update order status to new: %w", err)
	}

	return nil
}

// func receive( ctx context.Context, e cloudevents.Event ) error {
// 	db, err := utils.ConnectDatabase()
// 	if err != nil {
// 		fmt.Printf("Could not connect to database: %v", err)
// 		return fmt.Errorf("Could not connect to database: %w", err)
// 	}

// 	var allStoredOrders []models.StoredOrder

// 	err = json.Unmarshal(e.Data(), &allStoredOrders)
// 	if err != nil {
// 		fmt.Printf("Could not unmarshall e.Data() into type allStoredOrders: %v", err)
// 		return fmt.Errorf("Could not unmarshall e.Data() into type allStoredOrders: %w", err)
// 	}
	
// 	for i := range allStoredOrders {
// 		_, err = handler(ctx, allStoredOrders[i], db)
// 		if err != nil {
// 			fmt.Printf("Error in handler(): %v", err)
// 			return fmt.Errorf("Error in handler(): %w", err)
// 		}
// 	}
// 	return nil
// }