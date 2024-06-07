package main

import (
    "bytes"
    "io/ioutil"
    "os"
    "path/filepath"
    "testing"
)

var validJSON = []byte(`
{
    "name": "John Doe",
    "age": 30,
    "email": "john.doe@example.com",
    "isActive": true,
    "address": {
        "street": "123 Main St",
        "city": "Anytown",
        "zipcode": "12345"
    },
    "phoneNumbers": [
        {
            "type": "home",
            "number": "555-555-5555"
        },
        {
            "type": "work",
            "number": "555-555-5556"
        }
    ],
    "skills": [
        "Go",
        "Python",
        "JavaScript"
    ]
}
`)

var invalidJSON = []byte(`
{
    "name": "Jane Doe",
    "age": "thirty",
    "email": "jane.doe@example.com",
    "isActive": "true",
    "address": {
        "street": "456 Elm St",
        "city": "Othertown",
        "zipcode": 54321
    }
    "phoneNumbers": [
        {
            "type": "home",
            "number": "555-555-5557"
        },
        {
            "type": "work",
            "number": "555-555-5558"
        }
    ],
    "skills": [
        "Java",
        "C++",
        "Ruby"
    ]
}
`)

var jsonWithComments = []byte(`
// This is a comment
{
    "name": "John Doe", // Inline comment
    "age": 30,
    "email": "john.doe@example.com",
    "isActive": true,
    // Another comment
    "address": {
        "street": "123 Main St",
        "city": "Anytown",
        "zipcode": "12345"
    }
}
`)

func createTempFile(t *testing.T, content []byte) string {
    tmpFile, err := ioutil.TempFile("", "*.json")
    if err != nil {
        t.Fatalf("Failed to create temp file: %s", err)
    }
    if _, err := tmpFile.Write(content); err != nil {
        t.Fatalf("Failed to write to temp file: %s", err)
    }
    if err := tmpFile.Close(); err != nil {
        t.Fatalf("Failed to close temp file: %s", err)
    }
    return tmpFile.Name()
}

func createTempDir(t *testing.T, files map[string][]byte) string {
    tmpDir, err := ioutil.TempDir("", "jsonfiles")
    if err != nil {
        t.Fatalf("Failed to create temp directory: %s", err)
    }
    for name, content := range files {
        filePath := filepath.Join(tmpDir, name)
        if err := ioutil.WriteFile(filePath, content, 0644); err != nil {
            t.Fatalf("Failed to write file %s: %s", name, err)
        }
    }
    return tmpDir
}

func TestValidateJSON(t *testing.T) {
    // Test valid JSON
    err := ValidateJSON(validJSON)
    if err != nil {
        t.Errorf("Expected valid JSON to pass validation, but got error: %s", err)
    }

    // Test invalid JSON
    err = ValidateJSON(invalidJSON)
    if err == nil {
        t.Errorf("Expected invalid JSON to fail validation, but got no error")
    }

    // Test JSON with comments
    err = ValidateJSON(jsonWithComments)
    if err != nil {
        t.Errorf("Expected JSON with comments to pass validation, but got error: %s", err)
    }
}

func TestFormatJSON(t *testing.T) {
    // Test valid JSON formatting
    formatted, err := FormatJSON(validJSON)
    if err != nil {
        t.Errorf("Expected valid JSON to be formatted, but got error: %s", err)
    }

    // Check if the formatted JSON is correctly pretty-printed
    expectedFormatted := []byte(`{
  "address": {
    "city": "Anytown",
    "street": "123 Main St",
    "zipcode": "12345"
  },
  "age": 30,
  "email": "john.doe@example.com",
  "isActive": true,
  "name": "John Doe",
  "phoneNumbers": [
    {
      "number": "555-555-5555",
      "type": "home"
    },
    {
      "number": "555-555-5556",
      "type": "work"
    }
  ],
  "skills": [
    "Go",
    "Python",
    "JavaScript"
  ]
}`)
    if !bytes.Equal(formatted, expectedFormatted) {
        t.Errorf("Expected formatted JSON to be:\n%s\nbut got:\n%s", expectedFormatted, formatted)
    }

    // Test invalid JSON formatting
    _, err = FormatJSON(invalidJSON)
    if err == nil {
        t.Errorf("Expected invalid JSON to fail formatting, but got no error")
    }

    // Test JSON with comments formatting
    formatted, err = FormatJSON(jsonWithComments)
    if err != nil {
        t.Errorf("Expected JSON with comments to be formatted, but got error: %s", err)
    }
}

func TestRemoveComments(t *testing.T) {
    input := []byte(`// This is a comment
    {
        "key": "value" // Inline comment
    }
    // Another comment`)
    expected := []byte(`
    {
        "key": "value" 
    }
    `)
    output := RemoveComments(input)
    if !bytes.Equal(output, expected) {
        t.Errorf("Expected comments to be removed, but got:\n%s", output)
    }
}

func TestProcessFile(t *testing.T) {
    // Create temp files for testing
    validFile := createTempFile(t, validJSON)
    invalidFile := createTempFile(t, invalidJSON)
    commentFile := createTempFile(t, jsonWithComments)
    defer os.Remove(validFile)
    defer os.Remove(invalidFile)
    defer os.Remove(commentFile)

    // Test valid JSON file validation
    ProcessFile("validate", validFile)

    // Test invalid JSON file validation
    ProcessFile("validate", invalidFile)

    // Test JSON file with comments validation
    ProcessFile("validate", commentFile)

    // Test valid JSON file formatting
    ProcessFile("format", validFile)

    // Test invalid JSON file formatting
    ProcessFile("format", invalidFile)

    // Test JSON file with comments formatting
    ProcessFile("format", commentFile)
}

func TestProcessDirectory(t *testing.T) {
    // Create a temp directory with valid, invalid, and commented JSON files
    files := map[string][]byte{
        "valid1.json":    validJSON,
        "invalid1.json":  invalidJSON,
        "comment1.json":  jsonWithComments,
    }
    tempDir := createTempDir(t, files)
    defer os.RemoveAll(tempDir)

    // Test validating all files in the directory
    ProcessDirectory("validate", tempDir)

    // Test formatting all files in the directory
    ProcessDirectory("format", tempDir)
}
