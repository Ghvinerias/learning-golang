package main

import (
	"fmt"
	"log"
)

func main() {
	// Example input
	appInput := ApplicationParameters{
		itemName:          "Example.API",
		itemServer:        "IISAppC-01",
		itemCheckType:     "Swagger",
		preProcessingType: "Regexv1",
	}

	// Generate Item and Trigger Parameters
	result := generateItemAndTriggerParams(appInput)

	//Declare Host name in Zabbix
	hostname := "Web Monitoring"

	// Get Host ID by Host Name
	hostID, err := getHostID(hostname)
	if err != nil {
		log.Fatalf("Failed to get host ID: %v", err)
	}
	// fmt.Printf("Host ID for '%s' is: %s\n", hostname, hostID)

	if result.itemPreProcessing == "Regexv1" {
		// Perform type assertion for Regexv1
		preProcessing = itemProcessing.Regexv1.([]map[string]interface{})
	} else if result.itemPreProcessing == "Regexv2" {
		preProcessing = itemProcessing.Regexv2.([]map[string]interface{})
	} else if result.itemPreProcessing == "JsonParsing" {
		preProcessing = itemProcessing.JsonParsing.([]map[string]interface{})
	}

	// // Pass the asserted value to the function
	// item, trigger, err := createItemAndTrigger(apiKey, hostID, result.itemName, result.itemKey, result.itemURL, result.triggerURL, result.triggerExpression, preProcessing, "2")
	// if err != nil {
	// 	log.Fatalf("Failed to create item and trigger: %v", err)
	// }
	// fmt.Printf("item '%s\n' trigger %s\n", item, trigger)
	// Pass the asserted value to the function

	_, _, err = createItemAndTrigger(apiKey, hostID, result.itemName, result.itemKey, result.itemURL, result.triggerURL, result.triggerExpression, preProcessing, "2")
	if err != nil {
		log.Fatalf("Failed to create item and trigger: %v", err)
	}

}

func createItemAndTrigger(apiKey string, hostID, name, key_, url, description string, expression string, preprocessing interface{}, delay string) (string, string, error) {
	// Create an Item using parameters provided

	itemID, err := createItem(apiKey, hostID, name, key_, url, preprocessing, delay)
	if err != nil {
		return "", "", err
	}
	fmt.Printf("Item created with ID: %s\n", itemID)

	// Create a trigger using the created item key
	triggerID, err := createTrigger(apiKey, description, expression, url)
	if err != nil {
		return "", "", err
	}
	fmt.Printf("Trigger created with ID: %s\n", triggerID)
	return itemID, triggerID, nil
}
