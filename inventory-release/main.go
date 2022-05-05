package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"e-commerce-app/models" // local

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

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
	// if myOrder.Inventory.OrderID == orderID {
	// 	inventory = myOrder.Inventory
	// }
	
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

func receive( ctx context.Context, e cloudevents.Event ) {	
	var newOrders []models.Order

	err := json.Unmarshal(e.Data(), &newOrders)
	if err != nil {
		log.Fatalf("Couldn't unmarshal e.Data() into newOrders, %v", err)
	}

	for i := range newOrders {
		handler(ctx, newOrders[i])
	}
}

func main() {

	// fmt.Println("Order: \n", myOrder)
	fmt.Println("In the main function of main.go")

	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Fatal(c.StartReceiver(context.Background(), receive));

	// Calling receive() function directly for debugging
	// ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:8080/")
	// receive(ctx, cloudevents.NewEvent())
}