package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"e-commerce-app/models"
	"e-commerce-app/utils"

	"github.com/stretchr/testify/assert"
)

// testing scenario
// if storedOrder.OrderID[0:1] == "2" {
// 	return models.Order{}, models.NewErrProcessPayment("Unable to process payment for order " + storedOrder.OrderID)
// }

// Test Orders
var scenarioErrProcessPayment = "../test/order3.json"
var scenarioSuccessfulOrder = "../test/order7.json"

func TestHandler(t *testing.T) {
	assert := assert.New(t)

	t.Run("ProcessPayment", func(t *testing.T) {

		sto_ord := parseOrder(scenarioSuccessfulOrder)
		db, err := utils.ConnectDatabase()
		if err != nil {
			fmt.Printf("TestHandler(): Could not connect to database: %v", err)
		}
		prepareTestData(db, sto_ord)

		stored_order, err := handler(sto_ord, db)
		if err != nil {
			t.Fatal("Error failed to trigger with an invalid request")
		}

		assert.NotEmpty(stored_order.Order.Payment, "Payment must not be empty")
		assert.NotEmpty(stored_order.Order.Payment.TransactionID, "PaymentTransactionID must not be empty")
		assert.True(stored_order.Order.Payment.TransactionDate != sto_ord.Order.Payment.TransactionDate, "TransactionDate should be modified")
		assert.True(stored_order.Order.Payment.TransactionID != sto_ord.Order.Payment.TransactionID, "TransactionID should be modified")
		assert.True(stored_order.Order.Payment.PaymentType != sto_ord.Order.Payment.PaymentType, "PaymentType should be modified")
	})
}

func TestError(t *testing.T) {
	assert := assert.New(t)
	t.Run("ProcessPaymentErr", func(t *testing.T) {

		sto_ord := parseOrder(scenarioErrProcessPayment)
		db, err := utils.ConnectDatabase()
		if err != nil {
			fmt.Printf("TestError(): Could not connect to database: %v", err)
		}
		prepareTestData(db, sto_ord)

		stored_order, err := handler(sto_ord, db)
		if err != nil {
			fmt.Print(err)
		}

		assert.NotEmpty(stored_order.Order)

	})
}

func parseOrder(filename string) models.StoredOrder {
	inputFile, err := os.Open(filename)
	if err != nil {
		fmt.Printf("parseOrder(): opening input file: %v", err.Error())
	}

	defer inputFile.Close()

	jsonParser := json.NewDecoder(inputFile)

	stored_order := models.StoredOrder{}
	if err = jsonParser.Decode(&stored_order); err != nil {
		fmt.Printf("parseOrder(): parsing input file: %v", err.Error())
	}

	return stored_order
}

func prepareTestData(db *sql.DB, sto_ord models.StoredOrder) {
	order_id := sto_ord.OrderID
	order_info := sto_ord.Order
	command := `UPDATE stored_orders SET order_id = $1, order_info = $2;`
	_, err := db.Exec(command, order_id, order_info)
	if err != nil {
		fmt.Printf("prepareTestData(): Could not set up database for test: %v", err)
	}
}