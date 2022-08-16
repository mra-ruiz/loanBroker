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
	fmt.Println("Starting inventory release ...")
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

	log.Printf("[%s] - processing inventory release", storedOrder.OrderID)
	
	// Find inventory transaction in database
	inventory, err := getTransaction(storedOrder.OrderID, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return storedOrder, models.NewErrReleaseInventory(err.Error())
	}

	// Releasing items from inventory to make it available
	inventory.Release()

	// Saves transaction and updates inventory TransactionType to 'Release' 
	err = saveTransaction(inventory, db)
	if err != nil {
		log.Printf("[%s] - error! %s", storedOrder.OrderID, err.Error())
		return storedOrder, models.NewErrReleaseInventory(err.Error())
	}

	storedOrder.Order.Inventory = inventory

	fmt.Println()
	log.Printf("[%s] - reservation processed", storedOrder.OrderID)

	fmt.Println("\nUpdated stored orders:")
	err = utils.ViewDatabase(db)
	if err != nil {
		fmt.Printf("Error with ViewDatabase() in handler(): %v", err)
		return models.StoredOrder{}, fmt.Errorf("Error with ViewDatabase() in handler(): %w", err)
	}

	// Only for restoring database for testing reasons
	err = utils.ResetOrderInventory(db, storedOrder.OrderID)
	if err != nil {
		fmt.Printf("Error with ResetOrderInventory() in handler(): %v", err)
		return models.StoredOrder{}, fmt.Errorf("Error with ResetOrderInventory() in handler(): %w", err)
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

func getTransaction(orderID string, db *sql.DB) (models.Inventory, error) {
	// Searching for inventory
	resultingInventory, err := db.Query(`select order_info -> 'inventory' from stored_orders where order_id = $1;`, orderID)
	if err != nil {
		fmt.Printf("getTransaction(): Could not search for inventory %v", err)
		return models.Inventory{}, fmt.Errorf("getTransaction(): Could not search for inventory %w", err)
	}

	// Convert inventory of type JSONB to type models.Inventory
	var inventory models.Inventory
	var result []uint8

	for resultingInventory.Next() {
		resultingInventory.Scan(&result)
		json.Unmarshal(result, &inventory)		
	}
	
	return inventory, nil
}

func saveTransaction(inventory models.Inventory, db *sql.DB) error {
	// converting Inventory into a byte slice
	inventoryBytes, err := json.Marshal(inventory)
	if err != nil {
		fmt.Printf("saveTransaction(): Could not marshall inventory: %v", err)
		return fmt.Errorf("saveTransaction(): Could not marshall inventory: %w", err)
	}
	
	// Updating inventory of specific order in database
	updateString := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{inventory}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
	_, err = db.Exec(updateString, inventoryBytes, inventory.OrderID)
	if err != nil {
		fmt.Printf("Error with Exec() in saveTransaction(): Could not update inventory: %v", err)
		return fmt.Errorf("Error with Exec() in saveTransaction(): Could not update inventory: %w", err)
	}

	return nil
}