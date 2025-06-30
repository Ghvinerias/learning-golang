package main

import (
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/mock"
)

// MockExecCommand is a utility function to mock exec.Command
// It returns a function that replaces exec.Command and a channel to receive the executed commands
func MockExecCommand(t *testing.T, mockOutput string, mockError error) (func(string, ...string) *exec.Cmd, chan []string) {
	cmdChan := make(chan []string, 1)
	
	mockExecCommand := func(command string, args ...string) *exec.Cmd {
		cmdArgs := append([]string{command}, args...)
		cmdChan <- cmdArgs
		
		// Create a fake command that returns our mock output
		cmd := exec.Command("echo", mockOutput)
		return cmd
	}
	
	return mockExecCommand, cmdChan
}

// Helper for testing message processing with various scenarios
func setupMessageTest(t *testing.T, msg Message, mockDeliveryAckError error) (*MockChannel, amqp.Delivery) {
	// Mock channel
	mockChannel := new(MockChannel)
	mockChannel.On("Publish", "", doneQueueName, false, false, mock.AnythingOfType("amqp.Publishing")).Return(nil)
	
	// Create delivery
	msgBytes, _ := json.Marshal(msg)
	mockDelivery := amqp.Delivery{
		Body: msgBytes,
		Acknowledger: &mockAcknowledger{
			ackError: mockDeliveryAckError,
		},
	}
	
	return mockChannel, mockDelivery
}

// mockAcknowledger implements the amqp.Acknowledger interface for testing
type mockAcknowledger struct {
	ackError error
	nackError error
	acked bool
	nacked bool
}

func (m *mockAcknowledger) Ack(tag uint64, multiple bool) error {
	m.acked = true
	return m.ackError
}

func (m *mockAcknowledger) Nack(tag uint64, multiple bool, requeue bool) error {
	m.nacked = true
	return m.nackError
}

func (m *mockAcknowledger) Reject(tag uint64, requeue bool) error {
	m.nacked = true
	return m.nackError
}

// Add mock implementations of interface methods needed for testing
func (m *MockChannel) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockChannel) NotifyClose(c chan *amqp.Error) chan *amqp.Error {
	args := m.Called(c)
	return args.Get(0).(chan *amqp.Error)
}

// MockRabbitMQConnection mocks the amqp.Connection
type MockRabbitMQConnection struct {
	mock.Mock
}

func (m *MockRabbitMQConnection) Channel() (*amqp.Channel, error) {
	args := m.Called()
	return args.Get(0).(*amqp.Channel), args.Error(1)
}

func (m *MockRabbitMQConnection) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRabbitMQConnection) NotifyClose(c chan *amqp.Error) chan *amqp.Error {
	args := m.Called(c)
	return args.Get(0).(chan *amqp.Error)
}
