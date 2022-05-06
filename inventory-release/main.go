package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"e-commerce-app/models"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	fmt.Println("Starting inventory release ...")

	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Fatal(c.StartReceiver(context.Background(), receive));
}

func receive( ctx context.Context, e cloudevents.Event ) {	
	var orders []models.Order

	err := json.Unmarshal(e.Data(), &orders)
	if err != nil {
		log.Fatalf("Couldn't unmarshal e.Data() into orders, %v", err)
	}

	for i := range orders {
		handler(ctx, orders, orders[i])
	}
}

func handler(ctx context.Context, orders []models.Order, ord models.Order) (models.Order, error) {
	fmt.Println()
	log.Printf("[%s] - processing inventory release", ord.OrderID)
	
	// Find inventory transaction
	inventory, err := getTransaction(ctx, orders, ord.OrderID)
	if err != nil {
		log.Printf("[%s] - error! %s", ord.OrderID, err.Error())
		return models.Order{}, models.NewErrReleaseInventory(err.Error())
	}

	fmt.Println("\nInventory after getTransaction(): \n", inventory)

	// Releasing items from inventory to make it available
	inventory.Release()

	fmt.Println("\nInventory after Release(): \n", inventory)

	// Saves transaction and updates inventory TransactionType to 'Release' 
	err = saveTransaction(ctx, orders, inventory)
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

func getTransaction(ctx context.Context, orders []models.Order, orderID string) (models.Inventory, error) {

	inventory := models.Inventory{}

	for _,curOrder := range orders {
		if curOrder.OrderID == orderID {
			inventory = curOrder.Inventory
			break
		}
	}

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
	
	return inventory, nil
}

func saveTransaction(ctx context.Context, orders []models.Order, inventory models.Inventory) error {
	for i,curOrder := range orders {
		if curOrder.OrderID == inventory.OrderID {
			curOrder.Inventory = inventory

			orders[i] = curOrder
			break
		}
	}

	// MarshalIndent is just for debugging. Change back to Marshal()
	ordersBytes, err  := json.Marshal(orders)
	// ordersBytes, err  := json.MarshalIndent(orders, "", "    ")
  	if err != nil {
		log.Fatalf("Couldn't marshal orders, %v", err)
	}

	err = ioutil.WriteFile("./orders.json", ordersBytes, 0644)
  	if err != nil {
		log.Fatalf("Couldn't write to json file, %v", err)
	}

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