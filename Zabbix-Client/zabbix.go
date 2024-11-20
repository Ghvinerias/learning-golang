package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func createItem(apiKey string, hostID, name, key_, url, preprocessing interface{}, delay string, valueType string) (string, error) {
	itemParams := map[string]interface{}{
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
		"url":         url,
		"priority":    4,
		"opdata":      "{ITEM.LASTVALUE1}",
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

	// Print the full JSON request being sent
	fmt.Println("Zabbix API Request:", string(reqBody))

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

	hostID := result[0].(map[string]interface{})["hostid"].(string)
	return hostID, nil
}
