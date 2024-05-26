package main

import (
    "bytes"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "path/filepath"
    "regexp"
)

// RemoveComments removes comments starting with "//" from JSON data
func RemoveComments(data []byte) []byte {
    re := regexp.MustCompile(`(?m)^\s*//.*$`)
    return re.ReplaceAll(data, nil)
}

// ValidateJSON takes a byte slice of JSON data and returns an error if invalid, including the line number
func ValidateJSON(data []byte) error {
    data = RemoveComments(data) // Remove comments before validating
    dec := json.NewDecoder(bytes.NewReader(data))
    dec.DisallowUnknownFields()
    for {
        if _, err := dec.Token(); err != nil {
            if serr, ok := err.(*json.SyntaxError); ok {
                line, col := findLineAndCol(data, serr.Offset)
                return fmt.Errorf("syntax error at line %d, column %d: %v", line, col, err)
            } else if err == io.EOF {
                break
            } else {
                return err
            }
        }
    }
    return nil
}

// findLineAndCol finds the line and column of the byte offset
func findLineAndCol(data []byte, offset int64) (line, col int) {
    line = 1
    col = 1
    for i, b := range data {
        if i == int(offset) {
            break
        }
        col++
        if b == '\n' {
            line++
            col = 1
        }
    }
    return
}

// FormatJSON takes a byte slice of JSON data and returns a pretty-printed version or an error if invalid
func FormatJSON(data []byte) ([]byte, error) {
    data = RemoveComments(data) // Remove comments before formatting
    var js map[string]interface{}
    if err := json.Unmarshal(data, &js); err != nil {
        return nil, err
    }
    return json.MarshalIndent(js, "", "  ")
}

// ProcessFile processes a single file for validation or formatting
func ProcessFile(action, filePath string) {
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        fmt.Printf("Failed to read file: %s\n", err)
        return
    }

    switch action {
    case "validate":
        err := ValidateJSON(data)
        if err != nil {
            fmt.Printf("Invalid JSON in file %s: %s\n", filePath, err)
        } else {
            fmt.Printf("Valid JSON in file %s\n", filePath)
        }
    case "format":
        formattedJSON, err := FormatJSON(data)
        if err != nil {
            fmt.Printf("Invalid JSON in file %s: %s\n", filePath, err)
        } else {
            fmt.Printf("Formatted JSON for file %s:\n%s\n", filePath, string(formattedJSON))
        }
    default:
        fmt.Println("Invalid action. Use 'validate' or 'format'.")
    }
}

// ProcessDirectory recursively processes all files in a directory
func ProcessDirectory(action, dirPath string) {
    err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() && filepath.Ext(path) == ".json" {
            ProcessFile(action, path)
        }
        return nil
    })
    if err != nil {
        fmt.Printf("Failed to walk directory: %s\n", err)
    }
}

func printHelp() {
    fmt.Println("Usage: jsonvalidator <validate|format> <-d directory|-f file>")
    fmt.Println("Options:")
    fmt.Println("  -h, --help    Show this help message")
    fmt.Println("  -d directory  Specify the directory to process recursively")
    fmt.Println("  -f file       Specify the file to process")
    fmt.Println("Commands:")
    fmt.Println("  validate      Validate JSON file or all JSON files in a directory recursively")
    fmt.Println("  format        Format JSON file or all JSON files in a directory recursively")
}

func main() {
    // Define flags
    dirPath := flag.String("d", "", "Directory to process recursively")
    filePath := flag.String("f", "", "File to process")
    helpFlag := flag.Bool("help", false, "Show help message")
    flag.BoolVar(helpFlag, "h", false, "Show help message")

    // Parse flags after the action argument
    if len(os.Args) < 2 {
        printHelp()
        os.Exit(1)
    }

    action := os.Args[1]
    flag.CommandLine.Parse(os.Args[2:])

    // Check for help flag
    if *helpFlag {
        printHelp()
        os.Exit(0)
    }

    // Ensure either directory or file flag is set
    if *dirPath == "" && *filePath == "" {
        printHelp()
        os.Exit(1)
    }

    // Ensure both directory and file flags are not set simultaneously
    if *dirPath != "" && *filePath != "" {
        fmt.Println("Please specify either a directory or a file, not both.")
        printHelp()
        os.Exit(1)
    }

    // Process directory or file based on the flags set
    if *dirPath != "" {
        ProcessDirectory(action, *dirPath)
    } else if *filePath != "" {
        ProcessFile(action, *filePath)
    } else {
        printHelp()
        os.Exit(1)
    }
}
