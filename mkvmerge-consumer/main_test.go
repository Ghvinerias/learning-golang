package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockChannel mocks the amqp.Channel for testing
type MockChannel struct {
	mock.Mock
}

// Implementation of the amqp.Channel interface methods needed for tests
func (m *MockChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	callArgs := m.Called(name, durable, autoDelete, exclusive, noWait, args)
	return callArgs.Get(0).(amqp.Queue), callArgs.Error(1)
}

func (m *MockChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	callArgs := m.Called(exchange, key, mandatory, immediate, msg)
	return callArgs.Error(0)
}

func (m *MockChannel) Qos(prefetchCount, prefetchSize int, global bool) error {
	callArgs := m.Called(prefetchCount, prefetchSize, global)
	return callArgs.Error(0)
}

func (m *MockChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	callArgs := m.Called(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	return callArgs.Get(0).(<-chan amqp.Delivery), callArgs.Error(1)
}

// The MockChannel tests have been moved to helper_test.go

// For testing failOnError function

// TestFailOnError tests the failOnError function
func TestFailOnError(t *testing.T) {
	// Create a function that captures os.Exit calls
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	
	var exitCode int
	osExit = func(code int) {
		exitCode = code
		panic("os.Exit called")
	}
	
	// Redirect log output
	oldLogOutput := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)
	defer log.SetOutput(oldLogOutput)
	
	// Test with no error
	failOnError(nil, "This should not fail")
	
	// Test with error
	assert.Panics(t, func() {
		failOnError(fmt.Errorf("test error"), "Test error message")
	})
	
	// Check that exit code was 1
	assert.Equal(t, 1, exitCode)
	
	// Close the writer and read the output
	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()
	
	// Check log output contains our error message
	assert.Contains(t, output, "Test error message: test error")
}

// TestJoin tests the join function
func TestJoin(t *testing.T) {
	tests := []struct {
		elements  []string
		separator string
		expected  string
	}{
		{[]string{}, ",", ""},
		{[]string{"a"}, ",", "a"},
		{[]string{"a", "b", "c"}, ",", "a,b,c"},
		{[]string{"1", "2", "3"}, ":", "1:2:3"},
	}

	for _, tt := range tests {
		result := join(tt.elements, tt.separator)
		assert.Equal(t, tt.expected, result)
	}
}

// ProcessMessage tests moved to process_message_test.go

// Helper variable to make failOnError testable
var osExit = os.Exit
