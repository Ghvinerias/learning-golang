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
		{"http://Jenkins-Test.slick.ge:8080", "user", "0"},
		{"https://Jenkins.slick.ge", "user", "0"},
	}

	// Use the first configuration
	jenkinsTestConfig := jenkinsConfigs[0]
	jenkinsTest := gojenkins.CreateJenkins(nil, jenkinsTestConfig.URL, jenkinsTestConfig.Username, jenkinsTestConfig.Password)
	_, err := jenkinsTest.Init(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Jenkins client for %s: %v", jenkinsTestConfig.URL, err)
	}

	jenkinsProdConfig := jenkinsConfigs[1]
	jenkinsProd := gojenkins.CreateJenkins(nil, jenkinsProdConfig.URL, jenkinsProdConfig.Username, jenkinsProdConfig.Password)
	_, err = jenkinsProd.Init(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Jenkins client for %s: %v", jenkinsProdConfig.URL, err)
	}

	// Get Jenkins info
	info, err := jenkinsTest.Info(ctx)
	if err != nil {
		log.Fatalf("Failed to get Jenkins info for %s: %v", jenkinsTestConfig.URL, err)
	}
	info, err = jenkinsProd.Info(ctx)
	if err != nil {
		log.Fatalf("Failed to get Jenkins info for %s: %v", jenkinsProdConfig.URL, err)
	}

	// Access Jenkins version
	fmt.Printf("Jenkins version at %s: %s\n", jenkinsTestConfig.URL, info)
	fmt.Printf("Jenkins version at %s: %s\n", jenkinsProdConfig.URL, info)
}
