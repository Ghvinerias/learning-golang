package main

import (
	"fmt"
	"log"
)

func main() {
	// Declare Host name in Zabbix
	hostname := "Web Monitoring"
	// Get Host ID by Host Name
	hostID, err := getHostID(hostname)
	if err != nil {
		log.Fatalf("Failed to get host ID: %v", err)
	}
	fmt.Printf("Got HostID: %s\n", hostID)

	// printItemsWithValueType(hostID)
	// updatePreprocessingForItems()

	appInput2 := ItemAndTriggerParameters{
		monitoringHostID:   hostID,
		itemAndTriggerName: "Example1.API(Swagger)",
		itemKey:            "Example1.API",
		itemURL:            "https://Example1.01.API.slick.ge/swagger/index.html",
		itemPreProcessing:  itemProcessing.Regexv1,
		triggerExpression:  "last(/Web Monitoring/Example1.API)<>200",
		triggerURL:         "https://Example1.01.API.slick.ge/swagger/index.html",
		delay:              "2m",
		itemValueType:      "3",
	}

	// Pass the asserted value to the function
	item, trigger, err := createItemAndTrigger(appInput2)
	if err != nil {
		fmt.Printf("Failed to create item and trigger: %v", err)
		// log.Fatalf("Failed to create item and trigger: %v", err)
	}
	fmt.Printf("item '%s\n' trigger %s\n", item, trigger)

}
