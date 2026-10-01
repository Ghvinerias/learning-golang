package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// Variables to allow mocking in tests

// Variables to allow mocking in tests
var (
	amqpPublish = func(ch *amqp.Channel, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
		return ch.Publish(exchange, key, mandatory, immediate, msg)
	}

	// Mock version of publishToDLQ
	mockPublishToDLQ func(ch *amqp.Channel, body []byte, reason string) error
)
