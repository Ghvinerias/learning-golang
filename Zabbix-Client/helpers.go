package main

import (
	"fmt"
	"strings"
)

func printItemsWithValueType(hostID string) error {
	// Get all items for the specified host
	items, err := getAllItems(hostID)
	if err != nil {
		return fmt.Errorf("failed to retrieve items: %v", err)
	}

	// Iterate over items and print names of items with value_type == 3
	fmt.Println("Items with value_type = 3:")
	for _, item := range items {
		if item["value_type"] == "3" {
			fmt.Printf(" - %s, ItemID: %v\n", item["name"], item["itemid"])
		}
	}
	return nil
}

// Helper function to determine the type of URL for itemCheckType
func getCheckTypeURL(checkType string) string {
	if checkType == "Swagger" {
		return "swagger/index.html"
	}
	return "SlickHealthCheck"
}

// Function to generate ItemAndTriggerParameters based on ApplicationParameters
func generateItemAndTriggerParams(input ApplicationParameters) ItemAndTriggerParameters {
	// Split the itemServer to get the number part (e.g., IISAppC-01 => 01)
	serverNumber := strings.Split(input.itemServer, "-")[1]

	// Combine itemName and itemServer with an underscore
	itemName := input.itemName + "_" + input.itemServer

	// Generate itemKey
	itemKey := input.itemName

	// Define URLs based on itemServer
	itemURL := "https://" + strings.Split(input.itemName, ".")[0] + "." + serverNumber + ".API.slick.ge/" + getCheckTypeURL(input.itemCheckType)
	triggerURL := itemURL
	itemPreProcessing = input.preProcessingType

	if input.itemCheckType == "Swagger" {
		triggerExpression = "last(/Web Monitoring/" + itemKey + ")<>200"
		triggerName = itemName + " (Swagger)"
	} else if input.itemCheckType == "HealthCheck" {
		triggerExpression = "last(/Web Monitoring/" + itemKey + ")<>\"Healthy\""
		triggerName = itemName + " (SlickHealthCheck)"
	}

	// Return the struct with generated fields
	return ItemAndTriggerParameters{
		itemAndTriggerName: itemName,
		itemKey:            itemKey,
		itemURL:            itemURL,
		itemPreProcessing:  itemPreProcessing,
		triggerExpression:  triggerExpression,
		triggerURL:         triggerURL,
	}
}

func createItemAndTrigger(hostID, name, key_, url, description string, expression string, preprocessing interface{}, delay string, valueType string) (string, string, error) {
	// Create an Item using parameters provided

	itemID, err := createItem(hostID, name, key_, url, preprocessing, delay, valueType)
	if err != nil {
		return "", "", err
	}
	fmt.Printf("Item created with ID: %s\n", itemID)

	// Create a trigger using the created item key
	triggerID, err := createTrigger(description, expression, url)
	if err != nil {
		return "", "", err
	}
	fmt.Printf("Trigger created with ID: %s\n", triggerID)
	return itemID, triggerID, nil
}

func updatePreprocessingForItems(itemIDs []string, preprocessing interface{}) error {
	for _, itemID := range itemIDs {
		// Prepare the update request
		params := map[string]interface{}{
			"itemid":        itemID,
			"preprocessing": preprocessing,
		}

		request := ZabbixRequest{
			Jsonrpc: "2.0",
			Method:  "item.update",
			Params:  params,
			Auth:    apiKey,
			ID:      1,
		}

		// Perform the API call
		var response ZabbixResponse
		err := zabbixAPICall(request, &response)
		if err != nil {
			return fmt.Errorf("failed to update item %s: %v", itemID, err)
		}

		fmt.Printf("Updated preprocessing for item ID: %s\n", itemID)
	}

	return nil
}
