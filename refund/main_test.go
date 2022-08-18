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

// Test Orders
var scenarioErrProcessRefund = "../test/order4.json"
var scenarioSuccessfulOrder = "../test/order7.json"

func TestHandler(t *testing.T) {
	assert := assert.New(t)

	t.Run("ProcessRefund", func(t *testing.T) {

		sto_ord := parseOrder(scenarioSuccessfulOrder)
		db, err := utils.ConnectDatabase()
		if err != nil {
			fmt.Printf("TestHandler(): Error with ConnectDatabase(): %v", err)
		}
		prepareTestData(db, sto_ord)

		stored_order, err := handler(sto_ord, db)
		if err != nil {
			t.Fatal("Error failed to trigger with an invalid request")
		}

		assert.NotEmpty(stored_order.Order.Payment.TransactionID, "OrderID must be empty")
		assert.True(stored_order.Order.Payment.PaymentType == "Credit", "PaymentType must not be 'Credit'")
	})

}

func TestError(t *testing.T) {
	assert := assert.New(t)
	t.Run("ErrProcessRefund", func(t *testing.T) {

		sto_ord := parseOrder(scenarioErrProcessRefund)
		db, err := utils.ConnectDatabase()
		if err != nil {
			fmt.Printf("TestError(): Error with ConnectDatabase(): %v", err)
		}
		prepareTestData(db, sto_ord)

		stored_order, err := handler(sto_ord, db)
		if err != nil {
			fmt.Print(err)
		}

		assert.NotEmpty(stored_order.OrderID)
	})
}

func parseOrder(filename string) models.StoredOrder {
	inputFile, err := os.Open(filename)
	if err != nil {
		println("parseOrder(): opening input file", err.Error())
	}

	defer inputFile.Close()

	jsonParser := json.NewDecoder(inputFile)

	stored_order := models.StoredOrder{}
	if err = jsonParser.Decode(&stored_order); err != nil {
		println("parseOrder(): parsing input file", err.Error())
	}

	return stored_order
}

func prepareTestData(db *sql.DB, sto_ord models.StoredOrder) {
	order_id := sto_ord.OrderID
	order_info := sto_ord.Order
	command := `UPDATE stored_orders SET order_id = $1, order_info = $2;`
	_, err := db.Exec(command, order_id, order_info)
	if err != nil {
		fmt.Printf("prepareTestData(): Error updating database: %v", err)
	}
}