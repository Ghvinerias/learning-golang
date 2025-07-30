package main

import (
	"encoding/json"
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test for publishToDLQ function
func TestPublishToDLQ(t *testing.T) {
	// Setup global variables for test
	dlqQueueName = "test-dlq"

	// Create a mock channel
	mockChannel := new(MockChannelInterface)

	// Set up expectations for DLQ declaration
	mockChannel.On("QueueDeclare",
		dlqQueueName,
		true,
		false,
		false,
		false,
		mock.Anything).Return(amqp.Queue{Name: dlqQueueName}, nil)

	// Set up expectations for publishing to DLQ
	mockChannel.On("Publish",
		"",
		dlqQueueName,
		false,
		false,
		mock.MatchedBy(func(msg amqp.Publishing) bool {
			var dlqMessage map[string]interface{}
			err := json.Unmarshal(msg.Body, &dlqMessage)
			if err != nil {
				return false
			}

			// Check if dlqMessage has required fields
			originalMsg, hasOriginalMsg := dlqMessage["originalMessage"].(string)
			reason, hasReason := dlqMessage["errorReason"].(string)
			_, hasTimestamp := dlqMessage["timestamp"].(string)

			return hasOriginalMsg && originalMsg == "test message" &&
				hasReason && reason == "test reason" &&
				hasTimestamp
		})).Return(nil)

	// Call the function being tested
	err := publishToDLQ(mockChannel, []byte("test message"), "test reason")

	// Verify results
	assert.NoError(t, err)
	mockChannel.AssertExpectations(t)
}

// Test for publishToDLQ with error during publishing
func TestPublishToDLQError(t *testing.T) {
	// Setup global variables for test
	dlqQueueName = "test-dlq"

	// Create a mock channel
	mockChannel := new(MockChannelInterface)

	// Set up expectations for DLQ declaration
	mockChannel.On("QueueDeclare",
		dlqQueueName,
		true,
		false,
		false,
		false,
		mock.Anything).Return(amqp.Queue{Name: dlqQueueName}, nil)

	// Set up expectations for publishing to DLQ with an error
	publishErr := &amqp.Error{
		Code:   500,
		Reason: "Server internal error",
	}
	mockChannel.On("Publish",
		"",
		dlqQueueName,
		false,
		false,
		mock.Anything).Return(publishErr)

	// Call the function being tested
	err := publishToDLQ(mockChannel, []byte("test message"), "test reason")

	// Verify results
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to publish to DLQ")
	mockChannel.AssertExpectations(t)
}
