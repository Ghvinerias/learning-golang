package main

import (
	"fmt"
	"os"
	convertapi "github.com/ConvertAPI/convertapi-go/pkg"
	"github.com/ConvertAPI/convertapi-go/pkg/config"
	"github.com/ConvertAPI/convertapi-go/pkg/param"
)

func main() {
	// Check if the URL is provided as a command-line argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <URL>")
		os.Exit(1)
	}
	url := os.Args[1]

	// Set up ConvertAPI configuration
	config.Default = config.NewDefault(os.Getenv("CONVERTAPI_SECRET")) // Get your secret at https://www.convertapi.com/a

	fmt.Println("Converting remote PPTX to PDF from URL:", url)
	
	// Convert the specified URL to PDF
	pptxRes := convertapi.ConvDef("web", "pdf",
		param.NewString("url", url),
		param.NewString("filename", "web-example"),
	)

	// Save the result to the specified directory
	if files, errs := pptxRes.ToPath("/workspaces/learning-golang/temp"); errs == nil {
		fmt.Println("PDF file saved to:", files[0].Name())
	} else {
		fmt.Println("Error:", errs)
	}
}
