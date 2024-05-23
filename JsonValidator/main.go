package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
)

// ValidateJSON takes a byte slice of JSON data and returns an error if invalid
func ValidateJSON(data []byte) error {
    var js map[string]interface{}
    return json.Unmarshal(data, &js)
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: jsonvalidator <path to JSON file>")
        os.Exit(1)
    }

    filePath := os.Args[1]
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        fmt.Printf("Failed to read file: %s\n", err)
        os.Exit(1)
    }

    err = ValidateJSON(data)
    if err != nil {
        fmt.Printf("Invalid JSON: %s\n", err)
        os.Exit(1)
    }

    fmt.Println("Valid JSON")
}
