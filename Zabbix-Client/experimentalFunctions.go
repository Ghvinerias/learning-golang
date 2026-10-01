package main

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

	// Combine itemName and itemServer with an underscore
	itemName := inputValues.itemName + "_" + inputValues.itemServer

	// Generate itemKey
	itemKey := inputValues.itemName

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
