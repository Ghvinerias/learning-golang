package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/cavaliercoder/go-zabbix"
)

func main() {
	// Zabbix API credentials
	apiURL := "https://your-zabbix-server.com/api_jsonrpc.php"
	user := "your_username"
	password := "your_password"

	// Initialize Zabbix API client
	api := zabbix.NewAPI(apiURL)
	api.Login(user, password)

	// Load host export JSON (you mentioned you have this file)
	hostData, err := ioutil.ReadFile("host_export.json")
	if err != nil {
		log.Fatalf("Error reading host export file: %v", err)
	}

	var host map[string]interface{}
	if err := json.Unmarshal(hostData, &host); err != nil {
		log.Fatalf("Error parsing host export JSON: %v", err)
	}

	// Assume hostID is available in the export JSON
	hostID := host["hostid"].(string)

	// Create a new item
	item := zabbix.Item{
		HostId:      hostID,
		Key:         "custom.item",
		Name:        "Custom Item",
		Type:        0, // Zabbix agent
		ValueType:   zabbix.ZabbixFloat,
		Delay:       "60s",
		Description: "This is a custom item",
	}

	err = api.ItemsCreate(zabbix.Items{item})
	if err != nil {
		log.Fatalf("Error creating item: %v", err)
	}

	fmt.Println("Item created successfully")

	// Create a new trigger based on the item
	trigger := zabbix.Trigger{
		Description: "Trigger for Custom Item",
		Expression:  fmt.Sprintf("{%s:custom.item.last()} > 100", hostID),
		Priority:    4, // High
	}

	err = api.TriggersCreate(zabbix.Triggers{trigger})
	if err != nil {
		log.Fatalf("Error creating trigger: %v", err)
	}

	fmt.Println("Trigger created successfully")
}
