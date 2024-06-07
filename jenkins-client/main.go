package main

import (
	"context"
	"fmt"
	"log"

	"github.com/bndr/gojenkins"
)

type JenkinsConfig struct {
	URL      string
	Username string
	Password string
}

func main() {
	ctx := context.Background()

	// Array with one set of Jenkins configurations
	jenkinsConfigs := []JenkinsConfig{
		{"https://Jenkins", "Test", "Test"},
	}

	// Use the first configuration
	jenkinsTestConfig := jenkinsConfigs[0]
	jenkinsClient := gojenkins.CreateJenkins(nil, jenkinsTestConfig.URL, jenkinsTestConfig.Username, jenkinsTestConfig.Password)
	_, err := jenkinsClient.Init(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Jenkins client for %s: %v", jenkinsTestConfig.URL, err)
	}

	// Get Jenkins info
	info, err := jenkinsClient.Info(ctx)
	if err != nil {
		log.Fatalf("Failed to get Jenkins info for %s: %v", jenkinsTestConfig.URL, err)
	}

	// Access Jenkins version
	fmt.Printf("Jenkins version at %s: %s\n", jenkinsTestConfig.URL, info.Version)
}
