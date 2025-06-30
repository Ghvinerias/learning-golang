package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestJoinFunction(t *testing.T) {
	tests := []struct {
		name      string
		elements  []string
		separator string
		expected  string
	}{
		{"Empty slice", []string{}, ",", ""},
		{"Single element", []string{"one"}, ",", "one"},
		{"Multiple elements", []string{"one", "two", "three"}, ",", "one,two,three"},
		{"Different separator", []string{"one", "two", "three"}, ":", "one:two:three"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := join(tc.elements, tc.separator)
			assert.Equal(t, tc.expected, result, "Join function should correctly join elements with the separator")
		})
	}
}

func TestCategoryPathMapping(t *testing.T) {
	// Test that all expected categories are in the map
	expectedCategories := []string{"local-movies", "local-tvshows", "nas-tvshows", "nas-movies"}

	for _, category := range expectedCategories {
		_, exists := CategoryPathMap[category]
		assert.True(t, exists, fmt.Sprintf("Category %s should exist in CategoryPathMap", category))
	}
}

func TestMessageDeserialization(t *testing.T) {
	// Test JSON message parsing
	jsonData := `{"torrentName": "test-movie", "category": "local-movies"}`

	var msg Message
	err := json.Unmarshal([]byte(jsonData), &msg)

	assert.NoError(t, err, "Should parse valid JSON message")
	assert.Equal(t, "test-movie", msg.TorrentName)
	assert.Equal(t, "local-movies", msg.Category)
}

func TestPublishDoneMessageFormat(t *testing.T) {
	// Create a temporary file to simulate publishing a message
	tempFile := filepath.Join(os.TempDir(), "test-file.mkv")
	_, err := os.Create(tempFile)
	assert.NoError(t, err)
	defer os.Remove(tempFile)

	// We can test the message format by mocking the channel and capturing the published message
	mockChan := new(MockChannel)

	// Setup the expectation - capture the message for inspection
	var capturedMsg amqp.Publishing
	mockChan.On("Publish", "", doneQueueName, false, false, mock.AnythingOfType("amqp.Publishing")).
		Run(func(args mock.Arguments) {
			// Capture the message
			capturedMsg = args.Get(4).(amqp.Publishing)
		}).Return(nil)

	// Call the function under test
	err = publishDoneMessage(mockChan, tempFile)
	assert.NoError(t, err)

	// Verify the message format
	assert.Equal(t, "application/json", capturedMsg.ContentType)
	assert.True(t, len(capturedMsg.Body) > 0)

	// Parse the message body
	var msgData map[string]string
	err = json.Unmarshal(capturedMsg.Body, &msgData)
	assert.NoError(t, err)

	// Check message fields
	assert.Equal(t, tempFile, msgData["filename"])
	assert.Equal(t, "processed", msgData["status"])
	assert.NotEmpty(t, msgData["time"])
}
