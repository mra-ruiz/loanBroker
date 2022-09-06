package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"e-commerce-app/models"
	"e-commerce-app/utils"
)

var (
    db *sql.DB
)
 
func main() {
    connectDb()

    http.HandleFunc("/", handler)
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}

func connectDb() {
    var err error
    db, err = utils.ConnectDatabase()
    if err != nil {
        fmt.Printf("Could not connect to database: %v", err)
        log.Fatal(err)
    }
}

func handler(w http.ResponseWriter, req *http.Request) {
    body, err := io.ReadAll(req.Body)
    if err != nil {
        msg := fmt.Sprintf("Failed to read the request body: %v", err)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }
    defer req.Body.Close()

    // Receive order to refund
    var	order models.StoredOrder
    err = json.Unmarshal(body, &order)
    if err != nil {
        msg := fmt.Sprintf("handler(): Failed to unmarshal body: %v", err)
        log.Println(msg)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }

	log.Printf("[%s] - processing inventory release", order.OrderID)
	
	// Find inventory transaction in database
	fetchedInventory, err := getTransaction(order.OrderID)
    if err != nil {
        msg := fmt.Sprintf("handler(): Could not get inventory: %v", err)
        log.Println(msg)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }

	// Releasing items from inventory to make it available
	fetchedInventory.Release()

	// Saves transaction and updates inventory TransactionType to 'Release' 
	err = saveTransaction(fetchedInventory)
    if err != nil {
        msg := fmt.Sprintf("handler(): Could not save inventory: %v", err)
        log.Println(msg)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }

	log.Printf("[%s] - reservation processed", order.OrderID)
}

func getTransaction(orderID string) (models.Inventory, error) {
	// Searching for order
	resultingOrder, err := db.Query(`select order_info from stored_orders where order_id = $1;`, orderID)
    if err != nil {
        msg := fmt.Sprintf("Error with Query() in getTransaction(): Could not get order: %v", err)
        log.Println(msg)
        return models.Inventory{}, errors.New(msg)
    }

    // Convert order of type JSONB to type models.StoredOrder.Order
    var order models.Order
    for resultingOrder.Next() {
        resultingOrder.Scan(&order)
    }

    return order.Inventory, nil
}

func saveTransaction(inventory models.Inventory) error {
	// converting Inventory into a byte slice
	inventoryBytes, err := json.Marshal(inventory)
    if err != nil {
        msg := fmt.Sprintf("Error with Marshall() in saveTransaction(): Could not marshall inventory: %v", err)
        log.Println(msg)
        return errors.New(msg)
    }
	
	// Updating inventory of specific order in database
	updateString := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{inventory}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
	_, err = db.Exec(updateString, inventoryBytes, inventory.OrderID)
    if err != nil {
        msg := fmt.Sprintf("Error with Exec() in saveTransaction(): Could not update inventory: %v", err)
        log.Println(msg)
        return errors.New(msg)
    }

	return nil
}