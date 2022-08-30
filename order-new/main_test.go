package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"e-commerce-app/utils"

	"github.com/lib/pq"
)

var scenarioErrProcessOrder = "../test/order1.json"
var scenarioSuccessfulOrder = "../test/order7.json"

const (
	testOrder = `
	{
    "order_id": "orderID123456",
    "order_info": {
      "order_date": "2022-01-01T02:30:50Z",
      "customer_id": "id001",
      "order_status": "processing",
      "items": [
        {
          "item_id": "itemID456",
          "qty": 1,
          "description": "Pencil",
          "unit_price": 2.5
        },
        {
          "item_id": "itemID789",
          "qty": 1,
          "description": "Paper",
          "unit_price": 4.0
        }
      ],
      "payment": {
        "merchant_id": "merchantID1234",
        "payment_amount": 6.5,
        "transaction_id": "transactionID7845764",
        "transaction_date": "01-1-2022",
        "order_id": "orderID123456",
        "payment_type": "creditcard"
      },
      "inventory": {
        "transaction_id": "transactionID7845764",
        "transaction_date": "01-1-2022",
        "order_id": "orderID123456",
        "items": ["Pencil", "Paper"],
        "transaction_type": "online"
      }
    }
  }`
)

// testing scenario
// if storedOrder.OrderID[0:1] == "1" {
// 	return models.StoredOrder{}, models.NewErrProcessOrder("Unable to process order " + storedOrder.OrderID)
// }

func TestHandler(t *testing.T) {
	// assert := assert.New(t)

	t.Run("ProcessOrder", func(t *testing.T) {

		prepareDb(t)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testOrder))
		resp := httptest.NewRecorder()

		handler(resp, req)

		// assert.NotEmpty(stored_order.OrderID, "OrderID must be empty")
		// assert.NotEmpty(stored_order.Order, "Order must be empty")
		// assert.True(stored_order.OrderID == sto_ord.OrderID, "OrderID was modified which should not have happened")
		// assert.True(stored_order.Order.OrderStatus == "New", "OrderStatus not set to 'New'")
		// assert.True(stored_order.Order.CustomerID == sto_ord.Order.CustomerID, "CustomerID was modified which should not have happened")
		// assert.True(len(stored_order.Order.Items) == 3, "OrderItems should be contain 3 items ids")
		// assert.True(stored_order.Order.Total() == 56.97, "OrderTotal does not equal expected value")

	})
}

func prepareDb(t *testing.T) {
	utils.CredsLocation = "../test/postgres-creds.json"
	utils.SSLMode = "disable"

	var err error
	db, err = utils.ConnectDatabase()
	if err != nil {
		t.Fatalf("error connecting to the db: %v", err)
	}

	createTable(t)
}

func createTable(t *testing.T) {
	_, err := db.Exec(`CREATE TABLE stored_orders (order_id text, order_info JSONB);`)

	if err, ok := err.(*pq.Error); ok && err.Code.Name() != "duplicate_table" {
		t.Fatalf("createTable(): Error creating table %v", err)
	}

	// Cleanup table
	_, err = db.Exec(`DELETE FROM stored_orders;`)
	if err != nil {
		t.Fatalf("createTable(): Error deleting daa in table %v", err)
	}
}

func TestError(t *testing.T) {
	t.Skip()
	// assert := assert.New(t)

	t.Run("OrderProcessErr", func(t *testing.T) {

		// sto_ord := parseOrder(scenarioErrProcessOrder)
		// db, err := utils.ConnectDatabase()
		// if err != nil {
		// 	_ = fmt.Errorf("TestError(): Error with ConnectDatabase() %w", err)
		// }
		// prepareTestData(db, sto_ord)

		// stored_order, err := handler(sto_ord, db)
		// if err != nil {
		// 	fmt.Print(err)
		// }

		// assert.False(len(stored_order.Order.Items) > 0, "OrderItems has no items")
	})
}
