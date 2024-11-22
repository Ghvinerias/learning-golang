package main

import (
	"fmt"
	"log"
)

func main() {
	//Declare Host name in Zabbix
	hostname := "Web Monitoring"
	// Get Host ID by Host Name
	hostID, err := getHostID(hostname)
	if err != nil {
		log.Fatalf("Failed to get host ID: %v", err)
	}
	fmt.Printf("Got HostID: %s\n", hostID)

	printItemsWithValueType(hostID)
	// updatePreprocessingForItems()

/* 	// Example input
	appInput := ApplicationParameters{
		itemName:          "Example.API",
		itemServer:        "IISAppC-01",
		itemCheckType:     "HealthCheck",
		preProcessingType: "Regexv1",
	}
	// Generate Item and Trigger Parameters
	result := generateItemAndTriggerParams(appInput)
	if result.itemPreProcessing == "Regexv1" {
		preProcessing = itemProcessing.Regexv1.([]map[string]interface{})
		valueType = "3"
	} else if result.itemPreProcessing == "Regexv2" {
		preProcessing = itemProcessing.Regexv2.([]map[string]interface{})
		valueType = "3"
	} else if result.itemPreProcessing == "JsonParsing" {
		preProcessing = itemProcessing.JsonParsing.([]map[string]interface{})
		valueType = "5"
	}
	// Pass the asserted value to the function
	item, trigger, err := createItemAndTrigger(hostID, result.itemAndTriggerName, result.itemKey, result.itemURL, result.triggerURL, result.triggerExpression, preProcessing, "2", valueType)
	if err != nil {
		fmt.Printf("Failed to create item and trigger: %v", err)
		// log.Fatalf("Failed to create item and trigger: %v", err)
	}
	fmt.Printf("item '%s\n' trigger %s\n", item, trigger) */

	/*
		 	appInput2 := ItemAndTriggerParameters{
				itemAndTriggerName: "Example.API",
				itemKey:            "",
				itemURL:            "",
				itemPreProcessing:  "",
				triggerExpression:  "",
				triggerURL:         "",
			}

			// Pass the asserted value to the function
			item, trigger, err := createItemAndTrigger(hostID, appInput2.itemAndTriggerName, appInput2.itemKey, appInput2.itemURL, appInput2.triggerURL, appInput2.triggerExpression, appInput2.itemPreProcessing, "2", valueType)
			if err != nil {
				fmt.Printf("Failed to create item and trigger: %v", err)
				// log.Fatalf("Failed to create item and trigger: %v", err)
			}
			fmt.Printf("item '%s\n' trigger %s\n", item, trigger)
	*/
}
