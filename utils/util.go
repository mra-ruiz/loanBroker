package utils

import (
	"database/sql"
	"e-commerce-app/models"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func CheckForErrors(err error, s string) {
	if err != nil {
		fmt.Printf("%v\n", err)
		log.Fatalf(s)
	}
}

func ConnectDatabase() (*sql.DB, error) {
	// connection string
	host := "localhost"
    port := 5432
    user := "mruizcardenas"
    password := "K67u5ye"
    dbname := "postgres"

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	
	// open database
	db, err := sql.Open("postgres", psqlconn)
	CheckForErrors(err, "Could not open database")

	// check db
    err = db.Ping()
	CheckForErrors(err, "Could not ping database")
	fmt.Println("Connected to databse!")
	return db, err
}

func ViewDatabase(db *sql.DB) {
	var allStoredOrders []models.StoredOrder
	var storedOrder models.StoredOrder
	rows, err := db.Query(`SELECT * FROM stored_orders`)
	CheckForErrors(err, "send: Could not query select * from stored_orders")

	for rows.Next() {
		if err = rows.Scan(&storedOrder.OrderID, &storedOrder.Order); err != nil {
			CheckForErrors(err, "ViewDatabase(): Error with scan")
		} else {
			// fmt.Println("Here's where scan has no error")
		}
		allStoredOrders = append(allStoredOrders, storedOrder)
	}

	fmt.Println(allStoredOrders)
}

func ResetDatabase(db *sql.DB, resetType string) {
	if resetType == "order-new" {
		// reset the order status
		originalOrderStatus := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{order_status}', '"fillIn"', true) WHERE order_id = 'orderID123456';`
		_, err := db.Exec(originalOrderStatus)
		CheckForErrors(err, "Could not reset database")
	} else if resetType == "payment" {
		originalPayment := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{payment}', '{
			"order_id": "orderID123456",
			"merchant_id": "merchantID1234",
			"payment_type": "creditcard",
			"payment_amount": 6.5,
			"transaction_id": "transactionID7845764",
			"transaction_date": "01-1-2022"
		}', true) WHERE order_id = 'orderID123456';`

		_, err := db.Exec(originalPayment)
		CheckForErrors(err, "Could not reset database")
	} else if resetType == "inventory-reserve" { 
		originalInventory := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{inventory}', '{
			"transaction_id": "transactionID7845764", 
			"transaction_date": "01-1-2022", 
			"order_id": "orderID123456", 
			"items": [
				"Pencil", 
				"Paper"
			], 
			"transaction_type": "online"
		}', true) WHERE order_id = 'orderID123456';`

		_, err := db.Exec(originalInventory)
		CheckForErrors(err, "Could not reset database")
	} else if resetType == "inventory-release" {
		originalInventory := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{inventory}', '{
			"transaction_id": "transactionID7845764", 
			"transaction_date": "01-1-2022", 
			"order_id": "orderID123456", 
			"items": [
				"Pencil", 
				"Paper"
			], 
			"transaction_type": "online"
		}', true) WHERE order_id = 'orderID123456';`

		_, err := db.Exec(originalInventory)
		CheckForErrors(err, "Could not reset database")
	} else if resetType == "refund" {
		originalPayment := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{payment}', '{
			"order_id": "orderID123456",
			"merchant_id": "merchantID1234",
			"payment_type": "creditcard",
			"payment_amount": 6.5,
			"transaction_id": "transactionID7845764",
			"transaction_date": "01-1-2022"
		}', true) WHERE order_id = 'orderID123456';`

		_, err := db.Exec(originalPayment)
		CheckForErrors(err, "Could not reset database")
	} else if resetType == "order-update" {
		// reset order status
		originalOrderStatus := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{order_status}', '"fillIn"', true) WHERE order_id = 'orderID123456';`
		_, err := db.Exec(originalOrderStatus)
		CheckForErrors(err, "Could not reset database")
	} else {
		fmt.Fprintf(os.Stderr, "[%s] - Invalid type of reset", resetType)
        os.Exit(1)
	}
}