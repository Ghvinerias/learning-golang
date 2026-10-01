package main

import (
	"fmt"
	"log"
	"os"

	sdk "github.com/bitwarden/sdk-go"
)

func main() {
	accessToken := os.Getenv("BWS_ACCESS_TOKEN")
	secretID := os.Getenv("BWS_SECRET_ID")
	if accessToken == "" || secretID == "" {
		log.Fatal("BWS_ACCESS_TOKEN and BWS_SECRET_ID must be set")
	}

	client, err := sdk.NewBitwardenClient(nil, nil)
	if err != nil {
		log.Fatalf("create Bitwarden client: %v", err)
	}
	defer client.Close()

	if err := client.AccessTokenLogin(accessToken, nil); err != nil {
		log.Fatalf("authenticate with Bitwarden: %v", err)
	}

	secret, err := client.Secrets().Get(secretID)
	if err != nil {
		log.Fatalf("retrieve secret: %v", err)
	}

	// Deliberately avoid printing the secret value.
	fmt.Printf("Secret %q retrieved successfully.\n", secret.Key)
}
