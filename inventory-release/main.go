package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"e-commerce-app/models" // local

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

// Hard coding Order
var myItem1 = models.Item{ItemID: "itemID456", Qty: 1, Description: "Pencil", UnitPrice: 2.50}
var myItem2 = models.Item{ItemID: "itemID789", Qty: 1, Description: "Paper", UnitPrice: 4.00}


var myOrder = models.Order{
	OrderID: "orderID123456", 
	OrderDate: time.Date(2022, time.January, 1, 2, 30, 50, 12, time.Now().Location()), 
	CustomerID: "id001", 
	OrderStatus: "processing", 
	Items: []models.Item{ myItem1, myItem2 }, 
	Payment: models.Payment{
		MerchantID: "merchantID1234", 
		PaymentAmount: 6.50, 
		TransactionID: "transactionID7845764", 
		TransactionDate: "01-1-2022", 
		OrderID: "orderID123456", 
		PaymentType: "creditcard"}, 
	Inventory: models.Inventory{
		TransactionID: "transactionID7845764", 
		TransactionDate: "01-1-2022", 
		OrderID: "orderID123456", 
		OrderItems: []string{"Pencil", "Paper"}, 
		TransactionType: "online"},
}

func handler(ctx context.Context, ord models.Order) (models.Order, error) {
	fmt.Println()
	log.Printf("[%s] - processing inventory release", ord.OrderID)

	// Find inventory transaction
	inventory, err := getTransaction(ctx, ord.OrderID)
	if err != nil {
		log.Printf("[%s] - error! %s", ord.OrderID, err.Error())
		return models.Order{}, models.NewErrReleaseInventory(err.Error())
	}

	fmt.Println("\nInventory after getTransaction(): \n", inventory)

	// release the items to the inventory
	inventory.Release()

	fmt.Println("\nInventory after Release(): \n", inventory)

	// // save the inventory transaction
	err = saveTransaction(ctx, inventory)
	if err != nil {
		log.Printf("[%s] - error! %s", ord.OrderID, err.Error())
		return ord, models.NewErrReleaseInventory(err.Error())
	}

	ord.Inventory = inventory

	// testing scenario
	if ord.OrderID[0:2] == "33" {
		return ord, models.NewErrReleaseInventory("Unable to release inventory for order " + ord.OrderID)
	}

	fmt.Println()
	log.Printf("[%s] - reservation processed", ord.OrderID)

	return ord, nil
}

func getTransaction(ctx context.Context, orderID string) (models.Inventory, error) {

	inventory := models.Inventory{}

	// // defining a query input type
	// input := &dynamodb.QueryInput{
	// 	ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
	// 		":v1": {
	// 			S: aws.String(orderID),
	// 		},
	// 		":v2": {
	// 			S: aws.String("Reserve"),
	// 		},
	// 	},
	// 	KeyConditionExpression: aws.String("order_id = :v1 AND transaction_type = :v2"),
	// 	TableName:              aws.String(os.Getenv("TABLE_NAME")),
	// 	IndexName:              aws.String("orderIDIndex"),
	// }

	// // Get payment transaction from database
	// result, err := dynamoDB.QueryWithContext(ctx, input)
	// if err != nil {
	// 	return inventory, err
	// }

	// err = dynamodbattribute.UnmarshalMap(result.Items[0], &inventory)
	// if err != nil {
	// 	return inventory, fmt.Errorf("failed to DynamoDB unmarshal Record, %v", err.(awserr.Error))
	// }
	
	// fake query for now :)
	if myOrder.Inventory.OrderID == orderID {
		inventory = myOrder.Inventory
	}
	
	return inventory, nil
}

func saveTransaction(ctx context.Context, inventory models.Inventory) error {

	// marshalledInventory, err := dynamodbattribute.MarshalMap(inventory)
	// if err != nil {
	// 	return fmt.Errorf("failed to DynamoDB marshal Inventory, %v", err)
	// }

	// _, err = dynamoDB.PutItemWithContext(ctx, &dynamodb.PutItemInput{
	// 	TableName: aws.String(os.Getenv("TABLE_NAME")),
	// 	Item:      marshalledInventory,
	// })

	// if err != nil {
	// 	return fmt.Errorf("failed to put record to DynamoDB, %v", err)
	// }
	return nil
}

func main() {

	fmt.Println("Order: \n", myOrder)

	ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8080/")

	result, err := handler(ctx, myOrder)
	if err != nil {
		log.Fatalf("some error, %v", err)
	}

	fmt.Println("\n", result)
}