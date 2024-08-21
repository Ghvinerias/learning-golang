package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)


var outputFile *os.File

func main() {
	root := "./json-change" // Replace with your root directory path
	outputFilePath := "output.txt"        // Replace with your desired output file path

	var err error
	outputFile, err = os.Create(outputFilePath)
	if err != nil {
		log.Fatalf("Error creating output file %s: %v\n", outputFilePath, err)
	}
	defer outputFile.Close()

	// Walk through all directories and subdirectories
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the file has a .json extension
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			// Process the JSON file
			processJSONFile(path)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Error walking the path %q: %v\n", root, err)
	}
}

// Define a function to process each JSON file
func processJSONFile(path string) {
	// Read the JSON file
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Error reading file %s: %v\n", path, err)
		return
	}

	// Check and remove BOM if present
	content = removeBOM(content)

	// Create a map to hold the JSON data
	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		log.Printf("Skipping file %s due to JSON parsing error: %v\n", path, err)
		return
	}

	// Extract and save the required strings
	extractStrings(data)
}

// Function to remove BOM if present
func removeBOM(data []byte) []byte {
	bom := []byte{0xEF, 0xBB, 0xBF}
	if len(data) >= 3 && strings.HasPrefix(string(data[:3]), string(bom)) {
		return data[3:]
	}
	return data
}

// Define a function to extract strings based on the required patterns
func extractStrings(data interface{}) {
	// Define regular expressions for matching the required strings
	reURL := regexp.MustCompile(`https?://[^\s]+`)
	rePort := regexp.MustCompile(`\bport\b`)
	reSlickGE := regexp.MustCompile(`\blb\.ge\b`)
	reDataSource := regexp.MustCompile(`\bData Source=.*?\b`)

	// Walk through the JSON data recursively and extract strings
	extractRecursive(data, reURL, rePort, reSlickGE, reDataSource)
}

// Define a recursive function to walk through the JSON data
func extractRecursive(data interface{}, reURL, rePort, reSlickGE, reDataSource *regexp.Regexp) {
	switch v := data.(type) {
	case map[string]interface{}:
		for _, value := range v {
			extractRecursive(value, reURL, rePort, reSlickGE, reDataSource)
		}
	case []interface{}:
		for _, value := range v {
			extractRecursive(value, reURL, rePort, reSlickGE, reDataSource)
		}
	case string:
		if reURL.MatchString(v) || rePort.MatchString(v) || reSlickGE.MatchString(v) || reDataSource.MatchString(v) {
			// Validate UTF-8 encoding before writing to file
			if utf8.ValidString(v) {
				fmt.Fprintln(outputFile, v)
			} else {
				log.Printf("Skipping invalid UTF-8 string in file: %s\n", v)
			}
		}
	default:
		// Other types (e.g., float64, bool, etc.) are ignored
	}
}