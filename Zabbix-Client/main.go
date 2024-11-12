package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const (
	apiURL = "https://zabbix.slick.ge/api_jsonrpc.php"
	apiKey = "xxx"
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

func main() {
	hostname := "Web Monitoring"

	// Step 1: Get Host ID using Hostname
	hostID, err := getHostID(hostname)
	if err != nil {
		log.Fatalf("Failed to get host ID: %v", err)
	}

	fmt.Printf("Host ID for '%s' is: %s\n", hostname, hostID)

	// Step 2: Create an item using the retrieved hostID
	itemID, err := createItem(apiKey, hostID, "slick.ge/api/example (HealthCheck)", "slick.ge_api/example_HealthCheck", "https://slick.ge/api/api/example/HealthCheck", "1m", 4)
	if err != nil {
		log.Fatalf("Failed to create item: %v", err)
	}

	fmt.Printf("Item created with ID: %s\n", itemID)

	// Step 3: Create a trigger using the created item key
	triggerID, err := createTrigger(apiKey, "slick.ge/api/example (HealthCheck)", `last(/Web Monitoring/slick.ge_api/example_HealthCheck)<>"Healthy"`, "https://slick.ge/api/api/example/HealthCheck")
	if err != nil {
		log.Fatalf("Failed to create trigger: %v", err)
	}

	fmt.Printf("Trigger created with ID: %s\n", triggerID)
}

func getHostID(hostname string) (string, error) {
	params := map[string]interface{}{
		"output": []string{"hostid"},
		"filter": map[string]interface{}{
			"host": []string{hostname},
		},
	}

	request := ZabbixRequest{
		Jsonrpc: "2.0",
		Method:  "host.get",
		Params:  params,
		Auth:    apiKey,
		ID:      1,
	}

	var response ZabbixResponse
	err := zabbixAPICall(request, &response)
	if err != nil {
		return "", err
	}

	result, ok := response.Result.([]interface{})
	if !ok || len(result) == 0 {
		return "", fmt.Errorf("host not found")
	}

	host := result[0].(map[string]interface{})
	hostID := host["hostid"].(string)

	return hostID, nil
}

func createItem(apiKey, hostID, name, key_, url, delay string, valueType int) (string, error) {
	itemParams := map[string]interface{}{
		"hostid":     hostID,
		"name":       name,
		"key_":       key_,
		"type":       19, // HTTP Agent
		"value_type": valueType,
		"url":        url,
		"delay":      delay,
		"timeout":    "30",
		"status_codes": "",
		"preprocessing": []map[string]interface{}{
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
return "Healthy";
				`,
				"error_handler":        0,
				"error_handler_params": nil,
			},
		},
	}

	itemRequest := ZabbixRequest{
		Jsonrpc: "2.0",
		Method:  "item.create",
		Params:  itemParams,
		Auth:    apiKey,
		ID:      2,
	}

	var itemResponse ZabbixResponse
	err := zabbixAPICall(itemRequest, &itemResponse)
	if err != nil {
		return "", err
	}

	result, ok := itemResponse.Result.(map[string]interface{})
	if !ok || len(result) == 0 {
		return "", fmt.Errorf("failed to create item")
	}

	itemID := result["itemids"].([]interface{})[0].(string)
	return itemID, nil
}

func createTrigger(apiKey, description, expression, url string) (string, error) {
	triggerParams := map[string]interface{}{
		"description": description,
		"expression":  expression,
		"url":    url,
		"priority":    4,
	}

	triggerRequest := ZabbixRequest{
		Jsonrpc: "2.0",
		Method:  "trigger.create",
		Params:  triggerParams,
		Auth:    apiKey,
		ID:      2,
	}

	var triggerResponse ZabbixResponse
	err := zabbixAPICall(triggerRequest, &triggerResponse)
	if err != nil {
		return "", err
	}

	result, ok := triggerResponse.Result.(map[string]interface{})
	if !ok || len(result) == 0 {
		return "", fmt.Errorf("failed to create trigger")
	}

	triggerID := result["triggerids"].([]interface{})[0].(string)
	return triggerID, nil
}

func zabbixAPICall(request ZabbixRequest, response *ZabbixResponse) error {
	reqBody, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	fmt.Printf("Request JSON: %s\n", string(reqBody)) // Print the JSON request for debugging

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to make API call: %v", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if response.Error != nil {
		return fmt.Errorf("API error: %v", response.Error)
	}

	return nil
}
