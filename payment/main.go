package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"e-commerce-app/models"
	"e-commerce-app/utils"
)

var (
    db *sql.DB
)
 
func main() {
    connectDb()

    http.HandleFunc("/", handler)
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}

func connectDb() {
    var err error
    db, err = utils.ConnectDatabase()
    if err != nil {
        fmt.Printf("Could not connect to database: %v", err)
        log.Fatal(err)
    }
}

func handler(w http.ResponseWriter, req *http.Request) {

	time.Sleep(10*time.Second)

    body, err := io.ReadAll(req.Body)
    if err != nil {
        msg := fmt.Sprintf("Failed to read the request body: %v", err)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }
    defer req.Body.Close()

    // Receive order with payment info
    var	order models.StoredOrder
    err = json.Unmarshal(body, &order)
    if err != nil {
        msg := fmt.Sprintf("Failed to unmarshal body: %v", err)
        log.Println(msg)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }

    log.Printf("[%s] - processing payment", order.OrderID)

    var payment = models.Payment{
        OrderID:       order.OrderID,
        MerchantID:    "merch1",
        PaymentAmount: order.Order.Total(),
    }

    // Process payment
    payment.Pay()

    // Save payment
    err = savePayment(payment)
    if err != nil {
        msg := fmt.Sprintf("Could not save order with payment details: %v", err)
        log.Println(msg)
        w.Write([]byte(msg))
        w.WriteHeader(500) 
        return
    }

    // Save state
    order.Order.Payment = payment

    log.Printf("[%s] - payment processed", order.OrderID)
}

func savePayment(payment models.Payment) error {
    // converting payment into a byte slice
    paymentBytes, err := json.Marshal(payment)
    if err != nil {
        msg := fmt.Sprintf("Error with Marshall() in savePayment(): Could not marshall payment: %v", err)
        log.Println(msg)
        return errors.New(msg)
    }

    // Updating payment of specific order
    updatePaymentCommand := `UPDATE stored_orders SET order_info = jsonb_set(order_info, '{payment}', to_jsonb($1::JSONB), true) WHERE order_id = $2;`
    _, err = db.Exec(updatePaymentCommand, paymentBytes, payment.OrderID)
    if err != nil {
        msg := fmt.Sprintf("Error with Exec() in savePayment(): Could not update payment: %v", err)
        log.Println(msg)
        return errors.New(msg)
    }

    return nil
}