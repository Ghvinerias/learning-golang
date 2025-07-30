package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoad tests the Load function
func TestLoad(t *testing.T) {
	// Test with environment variables
	os.Setenv("RABBITMQ_HOST", "test-host")
	os.Setenv("RABBITMQ_PORT", "5673")
	os.Setenv("RABBITMQ_USERNAME", "test-user")
	os.Setenv("RABBITMQ_PASSWORD", "test-pass")
	os.Setenv("RABBITMQ_VHOST", "test-vhost")
	os.Setenv("RABBITMQ_QUEUE_TASKS", "test-tasks")
	os.Setenv("RABBITMQ_QUEUE_DONE", "test-done")
	os.Setenv("RABBITMQ_QUEUE_DLQ", "test-dlq")

	// Clean up environment variables after test
	defer func() {
		os.Unsetenv("RABBITMQ_HOST")
		os.Unsetenv("RABBITMQ_PORT")
		os.Unsetenv("RABBITMQ_USERNAME")
		os.Unsetenv("RABBITMQ_PASSWORD")
		os.Unsetenv("RABBITMQ_VHOST")
		os.Unsetenv("RABBITMQ_QUEUE_TASKS")
		os.Unsetenv("RABBITMQ_QUEUE_DONE")
		os.Unsetenv("RABBITMQ_QUEUE_DLQ")
	}()

	// Load configuration
	config, err := Load()

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Check that environment variables were loaded correctly
	assert.Equal(t, "test-host", config.RabbitMQ.Host)
	assert.Equal(t, "5673", config.RabbitMQ.Port)
	assert.Equal(t, "test-user", config.RabbitMQ.Username)
	assert.Equal(t, "test-pass", config.RabbitMQ.Password)
	assert.Equal(t, "test-vhost", config.RabbitMQ.Vhost)
	assert.Equal(t, "test-tasks", config.RabbitMQ.Queue.Tasks)
	assert.Equal(t, "test-done", config.RabbitMQ.Queue.Done)
	assert.Equal(t, "test-dlq", config.RabbitMQ.Queue.DLQ)
}

// TestConnectionString tests the ConnectionString method
func TestConnectionString(t *testing.T) {
	// Create a test config
	config := &Config{
		RabbitMQ: RabbitMQConfig{
			Host:     "localhost",
			Port:     "5672",
			Username: "guest",
			Password: "pass",
			Vhost:    "/test",
		},
	}

	// Test the ConnectionString method
	connString := config.ConnectionString()
	expected := "amqp://guest:pass@localhost:5672//test"
	assert.Equal(t, expected, connString)
}

// TestLoadWithConfigFile tests loading configuration from a file
func TestLoadWithConfigFile(t *testing.T) {
	// Create a temporary directory for the test config file
	tmpDir := t.TempDir()

	// Create a test config file
	configContent := `
rabbitmq:
  host: file-host
  port: 5674
  username: file-user
  password: file-pass
  vhost: file-vhost
  queue:
    tasks: file-tasks
    done: file-done
    dlq: file-dlq
paths:
  categories:
    test-category: /test/path
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	// Set the config directory
	oldwd, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(oldwd)

	err = os.Chdir(tmpDir)
	assert.NoError(t, err)

	// Load configuration
	config, err := Load()

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Check that file values were loaded correctly
	assert.Equal(t, "file-host", config.RabbitMQ.Host)
	assert.Equal(t, "5674", config.RabbitMQ.Port)
	assert.Equal(t, "file-user", config.RabbitMQ.Username)
	assert.Equal(t, "file-pass", config.RabbitMQ.Password)
	assert.Equal(t, "file-vhost", config.RabbitMQ.Vhost)
	assert.Equal(t, "file-tasks", config.RabbitMQ.Queue.Tasks)
	assert.Equal(t, "file-done", config.RabbitMQ.Queue.Done)
	assert.Equal(t, "file-dlq", config.RabbitMQ.Queue.DLQ)
	assert.Equal(t, "/test/path", config.Paths.Categories["test-category"])
}

// TestLoadDefaults tests that default values are correctly set
func TestLoadDefaults(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()

	// Change to the temp directory so we don't find any config files
	oldwd, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(oldwd)

	err = os.Chdir(tmpDir)
	assert.NoError(t, err)

	// Clear environment variables that might affect the test
	os.Unsetenv("RABBITMQ_HOST")
	os.Unsetenv("RABBITMQ_PORT")
	os.Unsetenv("RABBITMQ_USERNAME")
	os.Unsetenv("RABBITMQ_PASSWORD")
	os.Unsetenv("RABBITMQ_VHOST")
	os.Unsetenv("RABBITMQ_QUEUE_TASKS")
	os.Unsetenv("RABBITMQ_QUEUE_DONE")
	os.Unsetenv("RABBITMQ_QUEUE_DLQ")

	// Load configuration with defaults
	config, err := Load()

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Check that default values were set correctly
	assert.Equal(t, "localhost", config.RabbitMQ.Host)
	assert.Equal(t, "5672", config.RabbitMQ.Port)
	assert.Equal(t, "guest", config.RabbitMQ.Username)
	assert.Equal(t, "guest", config.RabbitMQ.Password)
	assert.Equal(t, "/", config.RabbitMQ.Vhost)
	assert.Equal(t, "mkvmerge.tasks", config.RabbitMQ.Queue.Tasks)
	assert.Equal(t, "mkvmerge.done", config.RabbitMQ.Queue.Done)
	assert.Equal(t, "mkvmerge.tasks_DLQ", config.RabbitMQ.Queue.DLQ)
	assert.Contains(t, config.Paths.Categories, "local-movies")
	assert.Contains(t, config.Paths.Categories, "local-tvshows")
}
