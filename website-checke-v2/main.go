package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
)

type RequestInput struct {
	HostHeader string `json:"Host_header"`
	Server     string `json:"Server"`
	Endpoint   string `json:"Endpoint"`
}

type ResponseOutput struct {
	Host   string `json:"Host"`
	Server string `json:"server"`
	HTTP   int    `json:"http"`
	HTTPS  int    `json:"https"`
}

func main() {
	// Define flags
	hostFlag := flag.String("HostName", "", "Host header")
	serverFlag := flag.String("Server", "", "Server")
	endpointFlag := flag.String("Endpoint", "", "Endpoint")
	fileFlag := flag.String("File", "", "Input JSON file")

	// Parse flags
	flag.Parse()

	var inputs []RequestInput

	if *fileFlag != "" {
		// Read input JSON file from flag
		inputData, err := ioutil.ReadFile(*fileFlag)
		if err != nil {
			fmt.Println("Error reading input file:", err)
			return
		}

		// Parse the input JSON
		err = json.Unmarshal(inputData, &inputs)
		if err != nil {
			fmt.Println("Error parsing input JSON:", err)
			return
		}
	} else if *hostFlag != "" && *serverFlag != "" && *endpointFlag != "" {
		// Read inputs from command line flags
		input := RequestInput{
			HostHeader: *hostFlag,
			Server:     *serverFlag,
			Endpoint:   *endpointFlag,
		}
		inputs = append(inputs, input)
	} else {
		fmt.Println("Usage: go run main.go -File <input_json_file> or go run main.go -HostName <Host_header> -Server <Server> -Endpoint <Endpoint>")
		return
	}

	var results []ResponseOutput

	// Process each input
	for _, input := range inputs {
		httpStatus := makeRequest("http", input.HostHeader, input.Server, input.Endpoint, false)
		httpsStatus := makeRequest("https", input.HostHeader, input.Server, input.Endpoint, true)

		// Collect the response
		result := ResponseOutput{
			Host:   input.HostHeader,
			Server: input.Server,
			HTTP:   httpStatus,
			HTTPS:  httpsStatus,
		}
		results = append(results, result)
	}

	// Convert results to JSON
	outputJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Println("Error generating output JSON:", err)
		return
	}

	// Print the response JSON
	fmt.Println(string(outputJSON))
}

func makeRequest(protocol, hostHeader, server, endpoint string, skipTLS bool) int {
	url := fmt.Sprintf("%s://%s%s", protocol, server, endpoint)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return 0
	}

	req.Host = hostHeader

	// Configure HTTP client with optional TLS skip
	client := &http.Client{}
	if skipTLS {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return 0
	}
	defer resp.Body.Close()

	return resp.StatusCode
}
