package main

import (
//	"os"
	"fmt"
	"flag"
)

func main() {
	// Define command-line flags
	name := flag.String("name", "", "The name option")
	response := flag.String("response", "", "The response message")

	// Parse the flags
	flag.Parse()

	// Check if required flags are provided
	if *name == "" || *response == "" {
		fmt.Println("Usage: go run main.go --name <name> --response <response>")
		return
	}

	// Print the formatted output
	fmt.Printf("You Have Chosen option %s, message was %s\n", *name, *response)
}