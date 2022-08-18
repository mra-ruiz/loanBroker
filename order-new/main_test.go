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
		if err != nil {
			_ = fmt.Errorf("TestHandler(): Error with ConnectDatabase() %w", err)
		}
		prepareTestData(db, sto_ord)

		stored_order, err := handler(sto_ord, db)
		if err != nil {
			t.Fatal("Error failed to trigger with an invalid request")
		}

		assert.NotEmpty(stored_order.OrderID, "OrderID must be empty")
		assert.NotEmpty(stored_order.Order, "Order must be empty")
		assert.True(stored_order.OrderID == sto_ord.OrderID, "OrderID was modified which should not have happened")
		assert.True(stored_order.Order.OrderStatus == "New", "OrderStatus not set to 'New'")
		assert.True(stored_order.Order.CustomerID == sto_ord.Order.CustomerID, "CustomerID was modified which should not have happened")
		assert.True(len(stored_order.Order.Items) == 3, "OrderItems should be contain 3 items ids")
		assert.True(stored_order.Order.Total() == 56.97, "OrderTotal does not equal expected value")

	})
}

func TestError(t *testing.T) {
	assert := assert.New(t)

	t.Run("OrderProcessErr", func(t *testing.T) {

		sto_ord := parseOrder(scenarioErrProcessOrder)
		db, err := utils.ConnectDatabase()
		if err != nil {
			_ = fmt.Errorf("TestError(): Error with ConnectDatabase() %w", err)
		}
		prepareTestData(db, sto_ord)

		stored_order, err := handler(sto_ord, db)
		if err != nil {
			fmt.Print(err)
		}

		assert.False(len(stored_order.Order.Items) > 0, "OrderItems has no items")
	})
}

func parseOrder(filename string) models.StoredOrder {
	inputFile, err := os.Open(filename)
	if err != nil {
		fmt.Println("parseOrder(): opening input file", err.Error())
	}

	defer inputFile.Close()

	jsonParser := json.NewDecoder(inputFile)

	stored_order := models.StoredOrder{}
	if err = jsonParser.Decode(&stored_order); err != nil {
		fmt.Println("parseOrder(): parsing input file", err.Error())
	}

	return stored_order
}

func prepareTestData(db *sql.DB, sto_ord models.StoredOrder) {
	order_id := sto_ord.OrderID
	order_info := sto_ord.Order
	command := `UPDATE stored_orders SET order_id = $1, order_info = $2;`
	_, err := db.Exec(command, order_id, order_info)
	if err != nil {
		_ = fmt.Errorf("prepareTestData(): Could not set up database for test: %w", err)
		os.Exit(1)
	}
}