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

	_ "github.com/lib/pq"
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

	// time.Sleep(10*time.Second)

    body, err := io.ReadAll(req.Body)
    if err != nil {
        msg := fmt.Sprintf("Failed to read the request body: %v", err)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }
    defer req.Body.Close()

    // Receive order with inventory reservation info
    var	order models.StoredOrder
    err = json.Unmarshal(body, &order)
    if err != nil {
        msg := fmt.Sprintf("Failed to unmarshal body: %v", err)
        log.Println(msg)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }

    log.Printf("[%s] - processing inventory reservation", order.OrderID)

    var newInvTrans = models.Inventory{
        OrderID:    order.OrderID,
        OrderItems: order.Order.ItemIds(),
    }

    // Reserve the items in the inventory
    newInvTrans.Reserve()

    // Annotate saga with inventory transaction id
    order.Order.Inventory = newInvTrans

    // Save the reservation
    err = saveInventory(newInvTrans)
    if err != nil {
        msg := fmt.Sprintf("Could not save order with inventory reservation details: %v", err)
        log.Println(msg)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }

    log.Printf("[%s] - reservation processed", order.OrderID)
}

func saveInventory(inventory models.Inventory) error {
    // converting Inventory into a byte slice
    inventoryBytes, err := json.Marshal(inventory)
    if err != nil {
        msg := fmt.Sprintf("Error with Marshall() in saveInventory(): Could not marshall inventory: %v", err)
        log.Println(msg)
        return errors.New(msg)
    }

    // Updating inventory of specific order
    updateString := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{inventory}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
    _, err = db.Exec(updateString, inventoryBytes, inventory.OrderID)
    if err != nil {
        msg := fmt.Sprintf("Error with Exec() in saveInventory(): Could not marshall inventory: %v", err)
        log.Println(msg)
        return errors.New(msg)
    }

    return nil
}