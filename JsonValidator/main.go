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

// FormatJSON takes a byte slice of JSON data and returns a pretty-printed version or an error if invalid
func FormatJSON(data []byte) ([]byte, error) {
    var js map[string]interface{}
    if err := json.Unmarshal(data, &js); err != nil {
        return nil, err
    }
    return json.MarshalIndent(js, "", "  ")
}

func main() {
    if len(os.Args) < 3 {
        fmt.Println("Usage: jsonvalidator <validate|format> <path to JSON file>")
        os.Exit(1)
    }

    action := os.Args[1]
    filePath := os.Args[2]

    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        fmt.Printf("Failed to read file: %s\n", err)
        os.Exit(1)
    }

    switch action {
    case "validate":
        err := ValidateJSON(data)
        if err != nil {
            fmt.Printf("Invalid JSON: %s\n", err)
            os.Exit(1)
        }
        fmt.Println("Valid JSON")
    case "format":
        formattedJSON, err := FormatJSON(data)
        if err != nil {
            fmt.Printf("Invalid JSON: %s\n", err)
            os.Exit(1)
        }
        fmt.Println(string(formattedJSON))
    default:
        fmt.Println("Invalid action. Use 'validate' or 'format'.")
        os.Exit(1)
    }
}

