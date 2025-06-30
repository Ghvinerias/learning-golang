package main

import (
	"github.com/streadway/amqp"
)

// ChannelAdapter adapts MockChannel to be used where *amqp.Channel is expected
type ChannelAdapter struct {
	mockChan *MockChannel
}

// QueueDeclare delegates to the mock
func (a *ChannelAdapter) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return a.mockChan.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
}

// Publish delegates to the mock
func (a *ChannelAdapter) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return a.mockChan.Publish(exchange, key, mandatory, immediate, msg)
}

// Adapter function that creates a channel adapter
func NewChannelAdapter(mock *MockChannel) *ChannelAdapter {
	return &ChannelAdapter{mockChan: mock}
}
