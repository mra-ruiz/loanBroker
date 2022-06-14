package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
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
		prepareTestData(db, sto_ord)

		stored_order, err := handler(nil, sto_ord, db)
		if err != nil {
			t.Fatal("Error failed to trigger with an invalid request")
		}

		assert.NotEmpty(stored_order.Order.Payment.TransactionID, "PaymentTransactionID must not be empty")

	})
}

func TestErrorIsOfTypeErrProcessPayment(t *testing.T) {
	assert := assert.New(t)
	t.Run("ProcessPaymentErr", func(t *testing.T) {

		sto_ord := parseOrder(scenarioErrProcessPayment)
		db, err := utils.ConnectDatabase()
		prepareTestData(db, sto_ord)

		stored_order, err := handler(nil, sto_ord, db)
		if err != nil {
			fmt.Print(err)
		}

		if assert.Error(err) {
			errorType := reflect.TypeOf(err)
			assert.Equal(errorType.String(), "*models.ErrProcessPayment", "Type does not match *models.ErrProcessPayment")
			assert.Empty(stored_order.OrderID)
		}
	})
}

func parseOrder(filename string) models.StoredOrder {
	inputFile, err := os.Open(filename)
	if err != nil {
		println("opening input file", err.Error())
	}

	defer inputFile.Close()

	jsonParser := json.NewDecoder(inputFile)

	stored_order := models.StoredOrder{}
	if err = jsonParser.Decode(&stored_order); err != nil {
		println("parsing input file", err.Error())
	}

	return stored_order
}

func prepareTestData(db *sql.DB, sto_ord models.StoredOrder) {
	order_id := sto_ord.OrderID
	order_info := sto_ord.Order
	command := `UPDATE stored_orders SET order_id = $1, order_info = $2;`
	_, err := db.Exec(command, order_id, order_info)
	utils.CheckForErrors(err, "Could not set up database for test")
}