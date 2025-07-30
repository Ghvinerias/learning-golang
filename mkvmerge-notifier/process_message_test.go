package main

import (
	"errors"
	"testing"

	"github.com/streadway/amqp"
)

// Test publishToDLQ function
func TestPublishToDLQ(t *testing.T) {
	// Save original function and restore it after test
	originalPublish := amqpPublish
	defer func() { amqpPublish = originalPublish }()

	// Test successful publish
	publishCalled := false
	amqpPublish = func(ch *amqp.Channel, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
		publishCalled = true
		if exchange != "" {
			t.Errorf("publish exchange = %v, want empty string", exchange)
		}
		if key != dlqQueueName {
			t.Errorf("publish routing key = %v, want %v", key, dlqQueueName)
		}
		return nil
	}

	err := publishToDLQ(nil, []byte("test message"), "test reason")
	if err != nil {
		t.Errorf("publishToDLQ() error = %v, want nil", err)
	}
	if !publishCalled {
		t.Error("publishToDLQ() did not call Publish")
	}

	// Test publish error
	publishError := errors.New("publish error")
	amqpPublish = func(ch *amqp.Channel, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
		return publishError
	}

	err = publishToDLQ(nil, []byte("test message"), "test reason")
	if err == nil {
		t.Error("publishToDLQ() error = nil, want error")
	}
}
