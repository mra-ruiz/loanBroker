package main

import (
	"encoding/json"
	"os"
	"testing"

	"e-commerce-app/models"

	"github.com/stretchr/testify/assert"
)

// Test Orders
var scenarioErrInventoryUpdate = "../test/order5.json"
var scenarioSuccessfulOrder = "../test/order7.json"

func TestHandler(t *testing.T) {
	assert := assert.New(t)

	o := parseOrder(scenarioSuccessfulOrder)
	orderSlice := []models.Order{o}

	order, err := handler(nil, orderSlice, o)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotEmpty(order.Inventory.TransactionID, "Inventory TransactionID must not be empty")

}

// func TestErrorIsOfTypeErrInventoryUpdate(t *testing.T) {
// 	assert := assert.New(t)

// 	input := parseOrder(scenarioErrInventoryUpdate)
// 	inputSlice := []models.Order{input}

// 	order, err := handler(nil, inputSlice, input)
// 	if err != nil {
// 		fmt.Print(err)
// 	}

// 	if assert.Error(err) {
// 		errorType := reflect.TypeOf(err)
// 		assert.Equal(errorType.String(), "*models.ErrReserveInventory", "Type does not match *models.ErrReserveInventory")
// 		assert.Empty(order.OrderID)
// 	}
// }

func parseOrder(filename string) models.Order {
	inputFile, err := os.Open(filename)
	if err != nil {
		println("opening input file", err.Error())
	}

	defer inputFile.Close()

	jsonParser := json.NewDecoder(inputFile)

	order := models.Order{}
	if err = jsonParser.Decode(&order); err != nil {
		println("parsing input file", err.Error())
	}

	return order
}