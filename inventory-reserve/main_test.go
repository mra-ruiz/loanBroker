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

// Test Orders
var scenarioErrInventoryUpdate = "../test/order5.json"
var scenarioSuccessfulOrder = "../test/order7.json"

func TestHandler(t *testing.T) {
	assert := assert.New(t)

	t.Run("ProcessPayment", func(t *testing.T) {

		sto_ord := parseOrder(scenarioSuccessfulOrder)
		db, err := utils.ConnectDatabase()
		if err != nil {
			fmt.Printf("TestHandler(): Error with ConnectDatabase(): %v", err)
		}
		prepareTestData(db, sto_ord)

		stored_order, err := handler(nil, sto_ord, db)
		if err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(stored_order.Order.Inventory.TransactionID, "Inventory TransactionID must not be empty")

	})
}

func TestErrorIsOfTypeErrInventoryUpdate(t *testing.T) {
	assert := assert.New(t)
	t.Run("ProcessPaymentErr", func(t *testing.T) {

		sto_ord := parseOrder(scenarioErrInventoryUpdate)
		db, err := utils.ConnectDatabase()
		if err != nil {
			fmt.Printf("TestErrorIsOfTypeErrInventoryUpdate(): Error with ConnectDatabase(): %v", err)
		}
		prepareTestData(db, sto_ord)

		stored_order, err := handler(nil, sto_ord, db)
		if err != nil {
			fmt.Print(err)
		}

		if assert.Error(err) {
			errorType := reflect.TypeOf(err)
			assert.Equal(errorType.String(), "*models.ErrReserveInventory", "Type does not match *models.ErrReserveInventory")
			assert.Empty(stored_order.OrderID)
		}
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
		fmt.Printf("prepareTestData(): Error with updating database: %v", err)
	}
}