package main

import (
	"encoding/json"
	"os"
	"strings"
)

type ZabbixRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Auth    string      `json:"auth,omitempty"`
	ID      int         `json:"id"`
}

type ZabbixResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   interface{} `json:"error,omitempty"`
	ID      int         `json:"id"`
}

type PreprocessingVariants struct {
	Regexv1     interface{}
	Regexv2     interface{}
	JsonParsing interface{}
}

var (
	apiURL            string
	hostID            int
	apiKey            string
	baseURL           string
	valueType         string
	itemProcessing    PreprocessingVariants
	preProcessing     interface{}
	itemPreProcessing string
	triggerExpression string
	triggerName       string
)

type ApplicationParameters struct {
	itemName          string
	itemServer        string
	itemCheckType     string
	preProcessingType string
}

type ItemAndTriggerParameters struct {
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

type inputParams struct {
	monitoringHostID   string
	itemAndTriggerName string
	itemKey            string
	itemValueType      string
	itemAndTriggerURL  string
	itemDelay          string
	itepPreProcessing  interface{}
	triggerExpression  string
	description        string
}

func init() {
	// Initialize API URL and Key from environment variables
	baseURL = os.Getenv("ZABBIX_SLICK_GE_API_URL")
	apiKey = os.Getenv("ZABBIX_SLICK_GE_API_KEY")
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	apiURL = baseURL + "api_jsonrpc.php"

	// Initialize itemProcessing Preprocessing variants
	itemProcessing = PreprocessingVariants{
		Regexv1: []map[string]interface{}{
			{
				"type":                 "5",
				"params":               json.RawMessage(`"HTTP\/1.1 ([0-9]+)\n\\1"`),
				"error_handler":        0,
				"error_handler_params": nil,
			},
		},
		Regexv2: []map[string]interface{}{
			{
				"type":                 "5",
				"params":               json.RawMessage(`"HTTP\/2 ([0-9]+)\n\\1"`),
				"error_handler":        0,
				"error_handler_params": nil,
			},
		},
		JsonParsing: []map[string]interface{}{
			{
				"sortorder": 0,
				"type":      21, // JavaScript preprocessing
				"params": `
var response;
try {
	response = JSON.parse(value);
} catch (e) {
	return "Service Unavailable";
}
var entries = response.Entries;
var unhealthyEntries = [];

for (var entryName in entries) {
	if (entries.hasOwnProperty(entryName)) {
		if (entries[entryName].Status == "Unhealthy") {
			unhealthyEntries.push(entryName);
		}
	}
}

if (unhealthyEntries.length > 0) {
	return "Unhealthy entries: " + unhealthyEntries.join(", ");
}
return "Healthy";`,
				"error_handler":        0,
				"error_handler_params": nil,
			},
		},
	}
}
