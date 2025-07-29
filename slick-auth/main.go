package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"

	"slick-auth/config"
)

func main() {
	// Create a new config instance with prefix "ZABBIX"
	cfg, err := config.New(
		config.WithEnvPrefix("ZABBIX"),
		// Optionally, you could also specify a config file:
		// config.WithConfigFile("config.yaml"),
	)
	if err != nil {
		log.Fatalf("Error initializing config: %v", err)
	}

	// Get all environment variables with prefix "ZABBIX"
	matchedVars := cfg.GetAllWithPrefix("ZABBIX")

	// Convert to JSON and print
	jsonData, err := json.MarshalIndent(matchedVars, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling to JSON: %v", err)
	}

	// Print JSON output
	fmt.Println(string(jsonData))

	// Alternative approach using Viper's functionality directly
	// jsonStr, err := cfg.GetAllAsJSON()
	// if err != nil {
	//     log.Fatalf("Error getting JSON: %v", err)
	// }
	// fmt.Println(jsonStr)
}

// Config wraps viper instance with additional functionality
type Config struct {
	viper      *viper.Viper
	envPrefix  string
	configFile string
}

// New creates a new Config instance with the given options
func New(options ...Option) (*Config, error) {
	c := &Config{
		viper:     viper.New(),
		envPrefix: "",
	}

	// Apply all options
	for _, option := range options {
		option(c)
	}

	// Load .env file if it exists (silently continue if it doesn't)
	_ = godotenv.Load()

	// Set automatic environment variable binding
	c.viper.SetEnvPrefix(c.envPrefix)
	c.viper.AutomaticEnv()
	c.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Load config file if specified
	if c.configFile != "" {
		c.viper.SetConfigFile(c.configFile)
		if err := c.viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	return c, nil
}

// Option defines a configuration option
type Option func(*Config)

// WithEnvPrefix sets the prefix for environment variables
func WithEnvPrefix(prefix string) Option {
	return func(c *Config) {
		c.envPrefix = prefix
	}
}

// WithConfigFile sets the configuration file to use
func WithConfigFile(file string) Option {
	return func(c *Config) {
		c.configFile = file
	}
}

// GetString retrieves a string value from environment or config file
func (c *Config) GetString(key string) string {
	return c.viper.GetString(key)
}

// GetInt retrieves an integer value from environment or config file
func (c *Config) GetInt(key string) int {
	return c.viper.GetInt(key)
}

// GetBool retrieves a boolean value from environment or config file
func (c *Config) GetBool(key string) bool {
	return c.viper.GetBool(key)
}

// GetStringMap retrieves a map of values with the given prefix
func (c *Config) GetStringMap(key string) map[string]interface{} {
	return c.viper.GetStringMap(key)
}

// GetAllWithPrefix returns all environment variables with the specified prefix
// This is similar to the original functionality
func (c *Config) GetAllWithPrefix(prefix string) map[string]string {
	result := make(map[string]string)

	// Get all environment variables
	envVars := os.Environ()

	// Filter by prefix
	for _, env := range envVars {
		if strings.HasPrefix(env, prefix) {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]
				result[key] = value
			}
		}
	}

	return result
}

// GetAllAsJSON returns all config values as a JSON string
func (c *Config) GetAllAsJSON() (string, error) {
	allSettings := c.viper.AllSettings()
	jsonData, err := json.Marshal(allSettings)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to JSON: %w", err)
	}
	return string(jsonData), nil
}

// Set sets a configuration value
func (c *Config) Set(key string, value interface{}) {
	c.viper.Set(key, value)
}

// Unmarshal binds the configuration values to a struct
func (c *Config) Unmarshal(rawVal interface{}) error {
	return c.viper.Unmarshal(rawVal)
}

// UnmarshalKey binds a specific key to a struct
func (c *Config) UnmarshalKey(key string, rawVal interface{}) error {
	return c.viper.UnmarshalKey(key, rawVal)
}
