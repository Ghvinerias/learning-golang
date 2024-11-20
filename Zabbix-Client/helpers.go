package main

import (
	"strings"
)

type ApplicationParameters struct {
	itemName          string
	itemServer        string
	itemCheckType     string
	preProcessingType string
}
type ItemAndTriggerParameters struct {
	itemName          string
	itemKey           string
	itemURL           string
	itemPreProcessing string
	triggerName       string
	triggerExpression string
	triggerURL        string
}

var itemPreProcessing, triggerExpression, triggerName string

// Function to generate ItemAndTriggerParameters based on ApplicationParameters
func generateItemAndTriggerParams(input ApplicationParameters) ItemAndTriggerParameters {
	// Split the itemServer to get the number part (e.g., IISAppC-01 => 01)
	serverNumber := strings.Split(input.itemServer, "-")[1]
	//serviceName := strings.Split(input.itemName, ".")[0]

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
		itemName:          itemName,
		itemKey:           itemKey,
		itemURL:           itemURL,
		itemPreProcessing: itemPreProcessing,
		triggerName:       triggerName,
		triggerExpression: triggerExpression,
		triggerURL:        triggerURL,
	}
}

// Helper function to determine the type of URL for itemCheckType
func getCheckTypeURL(checkType string) string {
	if checkType == "Swagger" {
		return "swagger/index.html"
	}
	return "SlickHealthCheck"
}
