package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Define the prefix to filter environment variables
	prefix := "ZABBIX"

	// Map to store matched environment variables
	matchedVars := make(map[string]string)

	// Retrieve all environment variables
	envVars := os.Environ()

	// Iterate and filter environment variables that start with "SLICK"
	for _, envVar := range envVars {
		if strings.HasPrefix(envVar, prefix) {
			// Split "KEY=value" into key and value
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) == 2 {
				matchedVars[parts[0]] = parts[1]
			}
		}
	}

	// Convert the matched variables map to JSON
	jsonData, err := json.MarshalIndent(matchedVars, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	// Print JSON output
	fmt.Println(string(jsonData))
}
