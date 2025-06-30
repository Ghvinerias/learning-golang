package main

import (
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/mock"
)

// MockChannel mocks the amqp.Channel interface
type MockChannel struct {
	mock.Mock
}

// Implement the Publish method for MockChannel
func (m *MockChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	args := m.Called(exchange, key, mandatory, immediate, msg)
	return args.Error(0)
}

// QueueDeclare mock implementation
func (m *MockChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	a := m.Called(name, durable, autoDelete, exclusive, noWait, args)
	return a.Get(0).(amqp.Queue), a.Error(1)
}

// Qos mock implementation
func (m *MockChannel) Qos(prefetchCount, prefetchSize int, global bool) error {
	a := m.Called(prefetchCount, prefetchSize, global)
	return a.Error(0)
}

// Consume mock implementation
func (m *MockChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	a := m.Called(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	return a.Get(0).(<-chan amqp.Delivery), a.Error(1)
}

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
	ackError  error
	nackError error
	acked     bool
	nacked    bool
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

// ExchangeDeclare mock implementation
func (m *MockChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	a := m.Called(name, kind, durable, autoDelete, internal, noWait, args)
	return a.Error(0)
}

// QueueDelete mock implementation
func (m *MockChannel) QueueDelete(name string, ifUnused, ifEmpty, noWait bool) (int, error) {
	a := m.Called(name, ifUnused, ifEmpty, noWait)
	return a.Int(0), a.Error(1)
}

// QueueBind mock implementation
func (m *MockChannel) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	a := m.Called(name, key, exchange, noWait, args)
	return a.Error(0)
}
