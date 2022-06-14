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

var scenarioErrProcessOrder = "../test/order1.json"
var scenarioSuccessfulOrder = "../test/order7.json"

// testing scenario
// if storedOrder.OrderID[0:1] == "1" {
// 	return models.StoredOrder{}, models.NewErrProcessOrder("Unable to process order " + storedOrder.OrderID)
// }

func TestHandler(t *testing.T) {
	assert := assert.New(t)

	t.Run("ProcessOrder", func(t *testing.T) {

		sto_ord := parseOrder(scenarioSuccessfulOrder)
		db, err := utils.ConnectDatabase()
		prepareTestData(db, sto_ord)

		stored_order, err := handler(nil, sto_ord, db)
		if err != nil {
			t.Fatal("Error failed to trigger with an invalid request")
		}

		assert.NotEmpty(stored_order.OrderID, "OrderID must be empty")
		assert.NotEmpty(stored_order.Order.CustomerID, "CustomerID must not be empty")
		assert.True(stored_order.Order.Total() == 56.97, "OrderTotal does not equal expected value")
		assert.True(len(stored_order.Order.Items) == 3, "OrderItems should be contain 3 items ids")

	})
}

func TestErrorIsOfTypeErrProcessOrder(t *testing.T) {
	assert := assert.New(t)

	t.Run("OrderProcessErr", func(t *testing.T) {

		sto_ord := parseOrder(scenarioErrProcessOrder)
		db, err := utils.ConnectDatabase()
		prepareTestData(db, sto_ord)

		stored_order, err := handler(nil, sto_ord, db)
		if err != nil {
			fmt.Print(err)
		}

		assert.NotEmpty(stored_order)

		if assert.Error(err) {
			errorType := reflect.TypeOf(err)
			assert.Equal(errorType.String(), "*models.ErrProcessOrder", "Type does not match *models.ErrProcessOrder")
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