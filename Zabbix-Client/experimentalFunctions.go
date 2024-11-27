package main

import (
	"strings"
)

type ItemAndTriggerParameters1 struct {
	monitoringHostID   string
	itemAndTriggerName string
	itemKey            string
	itemValueType      string
	delay              string
	itemURL            string
	itemPreProcessing  interface{}
	triggerExpression  string
	triggerURL         string
	hostID             string
}

/* 	itemParams := map[string]interface{}{
	"hostid":        hostID,
	"name":          name,
	"key_":          key_,
	"type":          19, // HTTP Agent
	"value_type":    valueType,
	"url":           url,
	"delay":         delay,
	"timeout":       "30",
	"status_codes":  "",
	"preprocessing": preprocessing,
} */
/* 	triggerParams := map[string]interface{}{
	"description": description,
	"expression":  expression,
	"url":         url,
	"priority":    4,
	"opdata":      "{ITEM.LASTVALUE1}",
} */

func processWithMinimalParameters(inputValues ApplicationParameters) (outputParameters interface{}) {

	// Split the itemServer to get the number part (e.g., IISAppC-01 => 01)
	serverNumber := strings.Split(inputValues.itemServer, "-")[1]

	// Combine itemName and itemServer with an underscore
	itemName := inputValues.itemName + "_" + inputValues.itemServer

	// Generate itemKey
	itemKey := inputValues.itemName

	// Define URLs based on itemServer
	itemURL := "https://" + strings.Split(inputValues.itemName, ".")[0] + "." + serverNumber + ".API.slick.ge/" + getCheckTypeURL(inputValues.itemCheckType)
	triggerURL := itemURL
	itemPreProcessing = inputValues.preProcessingType

	if inputValues.itemCheckType == "Swagger" {
		triggerExpression = "last(/Web Monitoring/" + itemKey + ")<>200"
		triggerName = itemName + " (Swagger)"
	} else if inputValues.itemCheckType == "HealthCheck" {
		triggerExpression = "last(/Web Monitoring/" + itemKey + ")<>\"Healthy\""
		triggerName = itemName + " (SlickHealthCheck)"
	}
	return outputParameters
}
