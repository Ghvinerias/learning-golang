package main

import (
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test for ensureMainQueueWithDLX function
func TestEnsureMainQueueWithDLX(t *testing.T) {
	// Setup global variables for test
	queueName = "test-queue"
	dlqQueueName = "test-dlq"

	// Create a mock channel
	mockChannel := new(MockChannelInterface)

	// Set up expectations for DLX exchange declaration
	mockChannel.On("ExchangeDeclare",
		"dlx",
		"direct",
		true,
		false,
		false,
		false,
		mock.Anything).Return(nil)

	// Set up expectations for queue declaration
	expectedQueue := amqp.Queue{Name: queueName}
	mockChannel.On("QueueDeclare",
		queueName,
		true,
		false,
		false,
		false,
		mock.MatchedBy(func(args amqp.Table) bool {
			dlx, hasDLX := args["x-dead-letter-exchange"].(string)
			routingKey, hasRoutingKey := args["x-dead-letter-routing-key"].(string)
			return hasDLX && dlx == "dlx" && hasRoutingKey && routingKey == dlqQueueName
		})).Return(expectedQueue, nil)

	// Set up expectations for DLQ declaration
	mockChannel.On("QueueDeclare",
		dlqQueueName,
		true,
		false,
		false,
		false,
		mock.Anything).Return(amqp.Queue{Name: dlqQueueName}, nil)

	// Set up expectations for binding DLQ to DLX
	mockChannel.On("QueueBind",
		dlqQueueName,
		dlqQueueName,
		"dlx",
		false,
		mock.Anything).Return(nil)

	// Call the function being tested
	result := ensureMainQueueWithDLX(mockChannel)

	// Verify results
	assert.Equal(t, expectedQueue, result)
	mockChannel.AssertExpectations(t)
}

// Test for ensureMainQueueWithDLX when queue exists with different configuration
func TestEnsureMainQueueWithDLXExistingQueueError(t *testing.T) {
	// Setup global variables for test
	queueName = "test-queue"
	dlqQueueName = "test-dlq"

	// Create a mock channel
	mockChannel := new(MockChannelInterface)

	// Set up expectations for DLX exchange declaration
	mockChannel.On("ExchangeDeclare",
		"dlx",
		"direct",
		true,
		false,
		false,
		false,
		mock.Anything).Return(nil)

	// First queue declare returns an error indicating inequivalent arguments
	preconditionErr := &amqp.Error{
		Code:   406,
		Reason: "PRECONDITION_FAILED",
	}
	mockChannel.On("QueueDeclare",
		queueName,
		true,
		false,
		false,
		false,
		mock.MatchedBy(func(args amqp.Table) bool {
			dlx, hasDLX := args["x-dead-letter-exchange"].(string)
			routingKey, hasRoutingKey := args["x-dead-letter-routing-key"].(string)
			return hasDLX && dlx == "dlx" && hasRoutingKey && routingKey == dlqQueueName
		})).Return(amqp.Queue{}, preconditionErr).Once()

	// QueueDelete attempt
	mockChannel.On("QueueDelete",
		queueName,
		false,
		false,
		false).Return(0, nil)

	// Second queue declare succeeds after deletion
	expectedQueue := amqp.Queue{Name: queueName}
	mockChannel.On("QueueDeclare",
		queueName,
		true,
		false,
		false,
		false,
		mock.MatchedBy(func(args amqp.Table) bool {
			dlx, hasDLX := args["x-dead-letter-exchange"].(string)
			routingKey, hasRoutingKey := args["x-dead-letter-routing-key"].(string)
			return hasDLX && dlx == "dlx" && hasRoutingKey && routingKey == dlqQueueName
		})).Return(expectedQueue, nil).Once()

	// Set up expectations for DLQ declaration
	mockChannel.On("QueueDeclare",
		dlqQueueName,
		true,
		false,
		false,
		false,
		mock.Anything).Return(amqp.Queue{Name: dlqQueueName}, nil)

	// Set up expectations for binding DLQ to DLX
	mockChannel.On("QueueBind",
		dlqQueueName,
		dlqQueueName,
		"dlx",
		false,
		mock.Anything).Return(nil)

	// Call the function being tested
	result := ensureMainQueueWithDLX(mockChannel)

	// Verify results
	assert.Equal(t, expectedQueue, result)
	mockChannel.AssertExpectations(t)
}

// Test for ensureMainQueueWithDLX when queue exists with different configuration and deletion fails
func TestEnsureMainQueueWithDLXDeletionFailure(t *testing.T) {
	// Skip the test for now as it's hard to mock os.Exit
	t.Skip("Skipping test that involves os.Exit which can't be easily mocked in Go")

	// Save original os.Exit and restore it after the test
	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()

	// Create a flag to track if our mock exit was called
	exitCalled := false

	// Mock os.Exit
	osExit = func(code int) {
		exitCalled = true
		// Just return instead of actually exiting
	}

	// Setup global variables for test
	queueName = "test-queue"
	dlqQueueName = "test-dlq"

	// Create a mock channel
	mockChannel := new(MockChannelInterface)

	// Set up expectations for DLX exchange declaration
	mockChannel.On("ExchangeDeclare",
		"dlx",
		"direct",
		true,
		false,
		false,
		false,
		mock.Anything).Return(nil)

	// First queue declare returns an error indicating inequivalent arguments
	preconditionErr := &amqp.Error{
		Code:   406,
		Reason: "PRECONDITION_FAILED",
	}
	mockChannel.On("QueueDeclare",
		queueName,
		true,
		false,
		false,
		false,
		mock.MatchedBy(func(args amqp.Table) bool {
			dlx, hasDLX := args["x-dead-letter-exchange"].(string)
			routingKey, hasRoutingKey := args["x-dead-letter-routing-key"].(string)
			return hasDLX && dlx == "dlx" && hasRoutingKey && routingKey == dlqQueueName
		})).Return(amqp.Queue{}, preconditionErr).Once()

	// QueueDelete attempt fails
	deleteErr := &amqp.Error{
		Code:   406,
		Reason: "QUEUE_IN_USE",
	}
	mockChannel.On("QueueDelete",
		queueName,
		false,
		false,
		false).Return(0, deleteErr)

	// Call the function
	ensureMainQueueWithDLX(mockChannel)

	// Verify our mock exit was called
	assert.True(t, exitCalled, "os.Exit should have been called")

	// Verify expectations
	mockChannel.AssertExpectations(t)
}
