package main

import (
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
	apiURL         string
	apiKey         string
	baseURL        string
	itemProcessing PreprocessingVariants
)

var preProcessing interface{}

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
				"params":               `HTTP\/1.1 ([0-9]+)\n\\1`,
				"error_handler":        0,
				"error_handler_params": nil,
			},
		},
		Regexv2: []map[string]interface{}{
			{
				"type":                 "5",
				"params":               `HTTP\/2 ([0-9]+)\n\\1`,
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