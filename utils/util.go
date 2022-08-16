package utils

import (
	"database/sql"
	"e-commerce-app/models"
	"fmt"

	_ "github.com/lib/pq"
)

func ConnectDatabase() (*sql.DB, error) {
	// connection string
	host := "0b536a47-b602-4e97-bb22-2fb574ec2db6.6131b73286f34215871dfad7254b4f7d.databases.appdomain.cloud"
    port := 31466
    user := "ibm_cloud_529548e9_4806_4c7e_adcd_6289bfda05db"
    password := "fbbc5ede25b2a0a0f21273139a774e94ae76a624f4ecf9dfef412bd029268cb7"
    dbname := "ibmclouddb"

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=require", host, port, user, password, dbname)
	
	// open database
	db, err := sql.Open("postgres", psqlconn)
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
				return fmt.Errorf( "ViewDatabase(): Error with scan: %w", err)
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

func ResetDatabase(db *sql.DB) error {
	originalDatabase := `UPDATE stored_orders SET order_id = 'orderID123456', order_info = '{
        "order_date": "2022-01-01T02:30:50Z", 
        "customer_id": "id001",
        "order_status": "fillIn",
        "items": [
            {
                "item_id": "itemID456", 
                "qty": 1, 
                "description": "Pencil", 
                "unit_price": 2.50
            },
            {
                "item_id": "itemID789", 
                "qty": 1, 
                "description": "Paper", 
                "unit_price": 4.00
            }
        ],            
        "payment":
        {
            "merchant_id": "merchantID1234", 
            "payment_amount": 6.50, 
            "transaction_id": "transactionID7845764", 
            "transaction_date": "01-1-2022", 
            "order_id": "orderID123456", 
            "payment_type": "creditcard"
        },
        "inventory":
        {
            "transaction_id": "transactionID7845764", 
            "transaction_date": "01-1-2022", 
            "order_id": "orderID123456", 
            "items": [
                "Pencil", 
                "Paper"
            ], 
            "transaction_type": "online"
        }
    }';`

	_, err := db.Exec(originalDatabase)
	if err != nil {
		return fmt.Errorf("Could not reset database: %w", err)
	}
	return nil

}