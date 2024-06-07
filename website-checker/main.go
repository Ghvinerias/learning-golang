package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Website struct {
	URL      interface{} `json:"url"`
	Name     string      `json:"name"`
	Insecure bool        `json:"insecure"`
}

type Response struct {
	URL           string `json:"url"`
	Name          string `json:"name"`
	StatusCode    int    `json:"status_code"`
	SSLCertExpiry string `json:"ssl_cert_expiry,omitempty"`
}

func main() {
	// Define flags
	urlPtr := flag.String("url", "", "URL to check")
	namePtr := flag.String("name", "", "Name of the Service")
	insecurePtr := flag.Bool("insecure", false, "Skip SSL certificate verification")
	filePtr := flag.String("file", "", "Path to JSON formatted file")

	// Parse flags
	flag.Parse()

	var websites []Website

	if *filePtr != "" {
		// Load data from file
		fileData, err := ioutil.ReadFile(*filePtr)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		var request struct {
			Websites []Website `json:"websites"`
		}
		err = json.Unmarshal(fileData, &request)
		if err != nil {
			fmt.Printf("Error parsing JSON file: %v\n", err)
			os.Exit(1)
		}

		websites = request.Websites
	} else {
		// Use flags for a single website
		if *urlPtr == "" || *namePtr == "" {
			fmt.Println("Please provide a URL and Name using the -url and -name flags, or use the -file flag to provide a JSON file.")
			os.Exit(1)
		}

		websites = []Website{
			{
				URL:      *urlPtr,
				Name:     *namePtr,
				Insecure: *insecurePtr,
			},
		}
	}

	var responses []Response

	for _, website := range websites {
		var urls []string
		switch v := website.URL.(type) {
		case string:
			urls = []string{v}
		case []interface{}:
			for _, u := range v {
				if urlStr, ok := u.(string); ok {
					urls = append(urls, urlStr)
				}
			}
		default:
			fmt.Printf("Unsupported URL format for website: %s\n", website.Name)
			continue
		}

		for _, url := range urls {
			// Check the status code of the URL
			client := &http.Client{}
			if website.Insecure {
				tr := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
				client.Transport = tr
			}

			resp, err := client.Get(url)
			if err != nil {
				fmt.Printf("Error fetching URL %s: %v\n", url, err)
				continue
			}
			defer resp.Body.Close()

			response := Response{
				URL:        url,
				Name:       website.Name,
				StatusCode: resp.StatusCode,
			}

			// If the URL is HTTPS, check SSL server certificate expiration date
			if resp.TLS != nil {
				serverCert := resp.TLS.PeerCertificates[0]
				response.SSLCertExpiry = serverCert.NotAfter.Format(time.RFC3339)
			} else {
				response.SSLCertExpiry = "No_SSL_WEBSITE"
			}

			responses = append(responses, response)
		}
	}

	// Convert responses to JSON
	jsonResponse, err := json.MarshalIndent(responses, "", "  ")
	if err != nil {
		fmt.Printf("Error creating JSON response: %v\n", err)
		os.Exit(1)
	}

	// Print JSON response
	fmt.Println(string(jsonResponse))
}
