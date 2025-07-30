package config

import (
	"os"
	"testing"
)

func TestConnectionString(t *testing.T) {
	// Test default connection string
	cfg := &Config{
		RabbitMQ: RabbitMQConfig{
			Host:     "localhost",
			Port:     "5672",
			Username: "guest",
			Password: "guest",
			Vhost:    "/",
		},
	}

	expected := "amqp://guest:guest@localhost:5672//"
	if cs := cfg.ConnectionString(); cs != expected {
		t.Errorf("ConnectionString() = %v, want %v", cs, expected)
	}

	// Test custom connection string
	cfg = &Config{
		RabbitMQ: RabbitMQConfig{
			Host:     "rabbitmq",
			Port:     "5673",
			Username: "user",
			Password: "pass",
			Vhost:    "vhost",
		},
	}

	expected = "amqp://user:pass@rabbitmq:5673/vhost"
	if cs := cfg.ConnectionString(); cs != expected {
		t.Errorf("ConnectionString() = %v, want %v", cs, expected)
	}
}

func TestLoad(t *testing.T) {
	// Test loading with environment variables
	os.Setenv("RABBITMQ_HOST", "test-host")
	os.Setenv("RABBITMQ_PORT", "1234")
	os.Setenv("RABBITMQ_USERNAME", "test-user")
	os.Setenv("RABBITMQ_PASSWORD", "test-pass")
	os.Setenv("RABBITMQ_VHOST", "test-vhost")
	os.Setenv("RABBITMQ_QUEUE_DONE", "test-queue")
	os.Setenv("RABBITMQ_QUEUE_DLQ", "test-dlq")
	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	os.Setenv("TELEGRAM_CHAT_ID", "12345")

	defer func() {
		os.Unsetenv("RABBITMQ_HOST")
		os.Unsetenv("RABBITMQ_PORT")
		os.Unsetenv("RABBITMQ_USERNAME")
		os.Unsetenv("RABBITMQ_PASSWORD")
		os.Unsetenv("RABBITMQ_VHOST")
		os.Unsetenv("RABBITMQ_QUEUE_DONE")
		os.Unsetenv("RABBITMQ_QUEUE_DLQ")
		os.Unsetenv("TELEGRAM_BOT_TOKEN")
		os.Unsetenv("TELEGRAM_CHAT_ID")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify that environment variables were loaded correctly
	if cfg.RabbitMQ.Host != "test-host" {
		t.Errorf("cfg.RabbitMQ.Host = %v, want %v", cfg.RabbitMQ.Host, "test-host")
	}
	if cfg.RabbitMQ.Port != "1234" {
		t.Errorf("cfg.RabbitMQ.Port = %v, want %v", cfg.RabbitMQ.Port, "1234")
	}
	if cfg.RabbitMQ.Username != "test-user" {
		t.Errorf("cfg.RabbitMQ.Username = %v, want %v", cfg.RabbitMQ.Username, "test-user")
	}
	if cfg.RabbitMQ.Password != "test-pass" {
		t.Errorf("cfg.RabbitMQ.Password = %v, want %v", cfg.RabbitMQ.Password, "test-pass")
	}
	if cfg.RabbitMQ.Vhost != "test-vhost" {
		t.Errorf("cfg.RabbitMQ.Vhost = %v, want %v", cfg.RabbitMQ.Vhost, "test-vhost")
	}
	if cfg.RabbitMQ.Queue.Done != "test-queue" {
		t.Errorf("cfg.RabbitMQ.Queue.Done = %v, want %v", cfg.RabbitMQ.Queue.Done, "test-queue")
	}
	if cfg.RabbitMQ.Queue.DLQ != "test-dlq" {
		t.Errorf("cfg.RabbitMQ.Queue.DLQ = %v, want %v", cfg.RabbitMQ.Queue.DLQ, "test-dlq")
	}
	if cfg.Telegram.BotToken != "test-token" {
		t.Errorf("cfg.Telegram.BotToken = %v, want %v", cfg.Telegram.BotToken, "test-token")
	}
	if cfg.Telegram.ChatID != 12345 {
		t.Errorf("cfg.Telegram.ChatID = %v, want %v", cfg.Telegram.ChatID, 12345)
	}
}
