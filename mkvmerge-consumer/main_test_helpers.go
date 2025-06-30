package main

import "github.com/streadway/amqp"

// Add this function to make the code more testable
// It allows injecting mock amqp.Dial for testing
var amqpDialFunc = amqp.Dial
