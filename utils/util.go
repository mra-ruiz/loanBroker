package utils

import (
	"database/sql"
	"e-commerce-app/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	_ "github.com/lib/pq"
)

func ConnectDatabase() (*sql.DB, error) {
    // open database
    db, err := sql.Open("postgres", dataSourceName())
    if err != nil {
        return nil, fmt.Errorf("Could not open databse: %w", err)
    }

    // check db
    err = db.Ping()
    if err != nil {
        return nil, fmt.Errorf("Could not open database: %w", err)
    }
    return db, nil
}

func ViewDatabase(db *sql.DB) error {
    var allStoredOrders []models.StoredOrder
    var storedOrder models.StoredOrder
    rows, err := db.Query(`SELECT * FROM stored_orders`)
    if err != nil {
        return fmt.Errorf("send: Could not query select * from stored_orders: %w", err)
    }

    for rows.Next() {
        if err = rows.Scan(&storedOrder.OrderID, &storedOrder.Order); err != nil {
            if err != nil {
                return fmt.Errorf("ViewDatabase(): Error with scan: %w", err)
            }
        } else {
            // fmt.Println("Here's where scan has no error")
        }
        allStoredOrders = append(allStoredOrders, storedOrder)
    }

    fmt.Println(allStoredOrders)
    return nil
}

func ImportDbData(db *sql.DB) []models.StoredOrder {
    var allStoredOrders []models.StoredOrder
    var storedOrder models.StoredOrder
    rows, err := db.Query(`SELECT * FROM stored_orders`)
    if err != nil {
        _ = fmt.Errorf("send: Could not query select * from stored_orders: %w", err)
        return nil
    }

    for rows.Next() {
        if err = rows.Scan(&storedOrder.OrderID, &storedOrder.Order); err != nil {
            if err != nil {
                _ = fmt.Errorf("ImportDBData(): Error with scan: %w", err)
                return nil
            }
        } else {
            // fmt.Println("Here's where scan has no error")
        }
        allStoredOrders = append(allStoredOrders, storedOrder)
        fmt.Println("Original stored orders:")
        fmt.Println(allStoredOrders)
    }

    return allStoredOrders
}

func ResetOrderStatus(db *sql.DB, orderID string) error {
    originalOrderStatus := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{order_status}', '"fillIn"', true) WHERE order_id = $1;`
    _, err := db.Exec(originalOrderStatus, orderID)
    if err != nil {
        return fmt.Errorf("Could not reset order status: %w", err)
    }
    return nil
}

func ResetOrderPayment(db *sql.DB, orderID string) error {
    originalPayment := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{payment}', '{
		"order_id": "orderID123456",
		"merchant_id": "merchantID1234",
		"payment_type": "creditcard",
		"payment_amount": 6.5,
		"transaction_id": "transactionID7845764",
		"transaction_date": "01-1-2022"
	}', true) WHERE order_id = $1;`

    _, err := db.Exec(originalPayment, orderID)
    if err != nil {
        return fmt.Errorf("Could not reset database: %w", err)
    }
    return nil
}

func ResetOrderInventory(db *sql.DB, orderID string) error {
    originalInventory := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{inventory}', '{
		"transaction_id": "transactionID7845764", 
		"transaction_date": "01-1-2022", 
		"order_id": "orderID123456", 
		"items": [
			"Pencil", 
			"Paper"
		], 
		"transaction_type": "online"
	}', true) WHERE order_id = $1;`

    _, err := db.Exec(originalInventory, orderID)
    if err != nil {
        return fmt.Errorf("Could not reset database: %w", err)
    }
    return nil
}

func dataSourceName() string {
    bytes, err := ioutil.ReadFile("/etc/postgresql/creds.json")
    if err != nil {
        log.Fatalf("failed to load postgreSQL credentials: %v", err)
    }

    var creds map[string]string
    err = json.Unmarshal(bytes, &creds)
    if err != nil {
        log.Fatalf("failed to load postgreSQL credentials as JSON: %v", err)
    }

    host, ok := creds["host"]
    if !ok {
        log.Fatal("failed to create postgreSQL connection: missing host")
    }

    port, ok := creds["port"]
    if !ok {
        log.Fatal("failed to create postgreSQL connection: missing port")
    }

    user, ok := creds["user"]
    if !ok {
        log.Fatal("failed to create postgreSQL connection: missing user")
    }

    password, ok := creds["password"]
    if !ok {
        log.Fatal("failed to create postgreSQL connection: missing password")
    }

    dbname, ok := creds["dbname"]
    if !ok {
        log.Fatal("failed to create postgreSQL connection: missing dbname")
    }

    return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require", host, port, user, password, dbname)
}
