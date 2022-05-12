package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"e-commerce-app/models"

	"github.com/stretchr/testify/assert"
)

// Test Orders
var scenarioErrProcessRefund = "../test/order4.json"
var scenarioSuccessfulOrder = "../test/order7.json"

func TestHandler(t *testing.T) {
	assert := assert.New(t)

	t.Run("ProcessRefund", func(t *testing.T) {

		input := parseOrder(scenarioSuccessfulOrder)
		input.OrderID = "77063fe3-56d9-4c51-b91f-71929834ce03"
		inputSlice := []models.Order{input}

		order, err := handler(nil, inputSlice, input)
		if err != nil {
			t.Fatal("Error failed to trigger with an invalid request")
		}

		assert.NotEmpty(order.Inventory.TransactionID, "TransactionID must not be empty")
	})

}

func TestErrorIsOfTypeErrProcessRefund(t *testing.T) {
	assert := assert.New(t)
	t.Run("ErrProcessRefund", func(t *testing.T) {

		input := parseOrder(scenarioErrProcessRefund)
		inputSlice := []models.Order{input}

		order, err := handler(nil, inputSlice, input)
		if err != nil {
			fmt.Print(err)
		}

		if assert.Error(err) {
			errorType := reflect.TypeOf(err)
			assert.Equal(errorType.String(), "*models.ErrProcessRefund", "Type does not match *models.ErrProcessRefund")
			assert.Empty(order.OrderID)
		}
	})
}

func parseOrder(filename string) models.Order {
	inputFile, err := os.Open(filename)
	if err != nil {
		println("opening input file", err.Error())
	}

	defer inputFile.Close()

	jsonParser := json.NewDecoder(inputFile)

	o := models.Order{}
	if err = jsonParser.Decode(&o); err != nil {
		println("parsing input file", err.Error())
	}

	return o
}