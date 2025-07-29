package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for our application
type Config struct {
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	Telegram TelegramConfig `mapstructure:"telegram"`
}

// RabbitMQConfig holds all RabbitMQ related configuration
type RabbitMQConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Vhost    string `mapstructure:"vhost"`
	Queue    struct {
		Done string `mapstructure:"done"`
		DLQ  string `mapstructure:"dlq"`
	} `mapstructure:"queue"`
}

// TelegramConfig holds Telegram bot related configuration
type TelegramConfig struct {
	BotToken string `mapstructure:"bot_token"`
	ChatID   int64  `mapstructure:"chat_id"`
}

// Load reads in config from files and environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Create a new viper instance
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Set config name and paths
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/mkvmerge-notifier/")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		log.Println("No config file found. Using defaults and environment variables.")
	}

	// Allow environment variables to override config files
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Unmarshal config into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults(v *viper.Viper) {
	// RabbitMQ defaults
	v.SetDefault("rabbitmq.host", "localhost")
	v.SetDefault("rabbitmq.port", "5672")
	v.SetDefault("rabbitmq.username", "guest")
	v.SetDefault("rabbitmq.password", "guest")
	v.SetDefault("rabbitmq.vhost", "/")
	v.SetDefault("rabbitmq.queue.done", "mkvmerge.done")
	v.SetDefault("rabbitmq.queue.dlq", "mkvmerge.done_DLQ")

	// Telegram defaults
	v.SetDefault("telegram.bot_token", "")
	v.SetDefault("telegram.chat_id", 0)
}

// ConnectionString returns the RabbitMQ connection string
func (c *Config) ConnectionString() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		c.RabbitMQ.Username,
		c.RabbitMQ.Password,
		c.RabbitMQ.Host,
		c.RabbitMQ.Port,
		c.RabbitMQ.Vhost)
}
