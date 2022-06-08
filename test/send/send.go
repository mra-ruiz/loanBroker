package send

import (
	"context"
	"database/sql"
	"e-commerce-app/models"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	fmt.Println("Hi! I am going to send a CloudEvent :)")

	db, err := ConnectDatabase()

	var allStoredOrders = importDbData(db)

	// Create client
	c, err := cloudevents.NewClientHTTP()
	CheckForErrors(err, "Failed to create client")

	// Create an Event.
	event :=  cloudevents.NewEvent()
	event.SetSource("example/uri")
	event.SetType("example.type")
	event.SetData(cloudevents.ApplicationJSON, &allStoredOrders)

	// Set a target.
	ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8080/")

	// Send that Event.
	if result := c.Send(ctx, event); cloudevents.IsUndelivered(result) {
		log.Fatalf("failed to send, %v", result)
	}
}

func ConnectDatabase() (*sql.DB, error) {
	// connection string
	host := "localhost"
    port := 5432
    user := "mruizcardenas"
    password := "K67u5ye"
    dbname := "postgres"

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	
	// open database
	db, err := sql.Open("postgres", psqlconn)
	CheckForErrors(err, "Could not open database")

	// check db
    err = db.Ping()
	CheckForErrors(err, "Could not ping database")
	fmt.Println("Connected to databse!")
	return db, err
}

func importDbData(db *sql.DB) []models.StoredOrder {
	var allStoredOrders []models.StoredOrder
	var storedOrder models.StoredOrder
	rows, err := db.Query(`SELECT * FROM stored_orders`)

	CheckForErrors(err, "send: Could not query select * from stored_orders")

	for rows.Next() {
		if err = rows.Scan(&storedOrder.OrderID, &storedOrder.Order); err != nil {
			CheckForErrors(err, "Error with scan")
		} else {
			fmt.Println("Here's where scan has no error")
		}
		allStoredOrders = append(allStoredOrders, storedOrder)
		fmt.Println(storedOrder)
		fmt.Println(allStoredOrders)
	}

	// Close database
	defer rows.Close()
	return allStoredOrders
}

func CheckForErrors(err error, s string) {
	if err != nil {
		fmt.Printf("%v\n", err)
		log.Fatalf(s)
	}
}