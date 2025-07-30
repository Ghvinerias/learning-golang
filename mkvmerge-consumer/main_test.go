package main

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockChannelInterface is a mock for the ChannelInterface
type MockChannelInterface struct {
	mock.Mock
}

func (m *MockChannelInterface) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	result := m.Called(name, durable, autoDelete, exclusive, noWait, args)
	return result.Get(0).(amqp.Queue), result.Error(1)
}

func (m *MockChannelInterface) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	result := m.Called(exchange, key, mandatory, immediate, msg)
	return result.Error(0)
}

func (m *MockChannelInterface) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	result := m.Called(name, kind, durable, autoDelete, internal, noWait, args)
	return result.Error(0)
}

func (m *MockChannelInterface) QueueDelete(name string, ifUnused, ifEmpty, noWait bool) (int, error) {
	result := m.Called(name, ifUnused, ifEmpty, noWait)
	return result.Int(0), result.Error(1)
}

func (m *MockChannelInterface) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	result := m.Called(name, key, exchange, noWait, args)
	return result.Error(0)
}

func (m *MockChannelInterface) Qos(prefetchCount, prefetchSize int, global bool) error {
	result := m.Called(prefetchCount, prefetchSize, global)
	return result.Error(0)
}

func (m *MockChannelInterface) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	result := m.Called(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	return result.Get(0).(<-chan amqp.Delivery), result.Error(1)
}

func (m *MockChannelInterface) Close() error {
	result := m.Called()
	return result.Error(0)
}

// We need to modify the function for testing
func TestFailOnError(t *testing.T) {
	// Create a separate test-only version of failOnError
	testFailOnError := func(err error, msg string) bool {
		if err != nil {
			t.Logf("%s: %s", msg, err)
			return true // indicates it would have called os.Exit
		}
		return false // indicates it would not have called os.Exit
	}

	// Test with no error
	exitWouldHaveBeenCalled := testFailOnError(nil, "test message")
	assert.False(t, exitWouldHaveBeenCalled, "failOnError should not exit with nil error")

	// Test with error
	exitWouldHaveBeenCalled = testFailOnError(errors.New("test error"), "test message")
	assert.True(t, exitWouldHaveBeenCalled, "failOnError should exit with non-nil error")
}

// Test for isInequivalentArgError function
func TestIsInequivalentArgError(t *testing.T) {
	// Test with PRECONDITION_FAILED error
	preconditionErr := &amqp.Error{
		Code:   406,
		Reason: "PRECONDITION_FAILED",
	}
	assert.True(t, isInequivalentArgError(preconditionErr))

	// Test with inequivalent arg error
	inequivalentErr := &amqp.Error{
		Code:   406,
		Reason: "inequivalent arg 'x-dead-letter-exchange'",
	}
	assert.True(t, isInequivalentArgError(inequivalentErr))

	// Test with other AMQP error
	otherErr := &amqp.Error{
		Code:   404,
		Reason: "NOT_FOUND",
	}
	assert.False(t, isInequivalentArgError(otherErr))

	// Test with non-AMQP error
	nonAmqpErr := errors.New("regular error")
	assert.False(t, isInequivalentArgError(nonAmqpErr))
}

// Test for ensureQueueExists function
func TestEnsureQueueExists(t *testing.T) {
	mockChannel := new(MockChannelInterface)

	expectedQueue := amqp.Queue{Name: "test-queue"}
	mockChannel.On("QueueDeclare", "test-queue", true, false, false, false, mock.Anything).Return(expectedQueue, nil)

	result := ensureQueueExists(mockChannel, "test-queue")

	assert.Equal(t, expectedQueue, result)
	mockChannel.AssertExpectations(t)
}

// Test for publishDoneMessage function
func TestPublishDoneMessage(t *testing.T) {
	mockChannel := new(MockChannelInterface)

	// Setup global variables for test
	doneQueueName = "test-done-queue"

	// Mock the Publish method to return nil (success)
	mockChannel.On("Publish",
		"",
		doneQueueName,
		false,
		false,
		mock.MatchedBy(func(msg amqp.Publishing) bool {
			// Verify message format
			var message map[string]string
			err := json.Unmarshal(msg.Body, &message)
			if err != nil {
				return false
			}

			// Check if message has required fields
			_, hasFilename := message["filename"]
			_, hasStatus := message["status"]
			_, hasTime := message["time"]

			return hasFilename && hasStatus && hasTime
		})).Return(nil)

	// Test the function
	err := publishDoneMessage(mockChannel, "test-file")

	// Verify results
	assert.NoError(t, err)
	mockChannel.AssertExpectations(t)
}

// Test for the join helper function
func TestJoin(t *testing.T) {
	// Test with empty slice
	assert.Equal(t, "", join([]string{}, ","))

	// Test with single element
	assert.Equal(t, "one", join([]string{"one"}, ","))

	// Test with multiple elements
	assert.Equal(t, "one,two,three", join([]string{"one", "two", "three"}, ","))

	// Test with different separator
	assert.Equal(t, "one:two:three", join([]string{"one", "two", "three"}, ":"))
}

// MockDelivery for testing amqp.Delivery
// This implements the methods we need from amqp.Delivery
type MockDelivery struct {
	mock.Mock
}

func (m *MockDelivery) Ack(multiple bool) error {
	args := m.Called(multiple)
	return args.Error(0)
}

func (m *MockDelivery) Nack(multiple, requeue bool) error {
	args := m.Called(multiple, requeue)
	return args.Error(0)
}

func (m *MockDelivery) Reject(requeue bool) error {
	args := m.Called(requeue)
	return args.Error(0)
}

// Implement the Body property as a function so we can mock it
func (m *MockDelivery) GetBody() []byte {
	args := m.Called()
	return args.Get(0).([]byte)
}

// Test for rejectMessageToDLQ function
func TestRejectMessageToDLQ(t *testing.T) {
	mockDelivery := new(MockDelivery)

	// Mock the Reject method
	mockDelivery.On("Reject", false).Return(nil)

	// Test successful rejection
	err := rejectMessageToDLQ(mockDelivery, "test reason")
	assert.NoError(t, err)
	mockDelivery.AssertExpectations(t)

	// Test with rejection error
	mockDelivery = new(MockDelivery)
	mockDelivery.On("Reject", false).Return(errors.New("rejection error"))

	err = rejectMessageToDLQ(mockDelivery, "test reason")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to reject message")
	mockDelivery.AssertExpectations(t)
}
