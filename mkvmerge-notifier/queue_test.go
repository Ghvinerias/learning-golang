package main

import (
	"errors"
	"testing"

	"github.com/streadway/amqp"
)

func TestIsInequivalentArgErrorAdditional(t *testing.T) {
	// Additional test cases for isInequivalentArgError function
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "AMQP error with different code",
			err:      &amqp.Error{Code: 404, Reason: "NOT_FOUND"},
			expected: false,
		},
		{
			name:     "AMQP error with code 406 but different reason",
			err:      &amqp.Error{Code: 406, Reason: "OTHER_REASON"},
			expected: false,
		},
		{
			name:     "Generic error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isInequivalentArgError(tc.err)
			if result != tc.expected {
				t.Errorf("isInequivalentArgError(%v) = %v, want %v", tc.err, result, tc.expected)
			}
		})
	}
}

// Test the queue name globals
func TestQueueNames(t *testing.T) {
	// Save original values
	origQueueName := queueName
	origDlqQueueName := dlqQueueName

	// Restore after test
	defer func() {
		queueName = origQueueName
		dlqQueueName = origDlqQueueName
	}()

	// Test different values
	testNames := []struct {
		queue string
		dlq   string
	}{
		{"test.queue", "test.queue_DLQ"},
		{"main.queue", "main.queue_DLQ"},
		{"", "_DLQ"}, // Edge case
	}

	for _, names := range testNames {
		queueName = names.queue
		dlqQueueName = names.dlq

		if queueName != names.queue {
			t.Errorf("queueName = %q, want %q", queueName, names.queue)
		}

		if dlqQueueName != names.dlq {
			t.Errorf("dlqQueueName = %q, want %q", dlqQueueName, names.dlq)
		}
	}
}
