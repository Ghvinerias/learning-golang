package main

import (
    "bytes"
    "testing"
)

// Sample JSON data for testing
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
}
