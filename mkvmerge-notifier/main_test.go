package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"mkvmerge-notifier/config"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/streadway/amqp"
)

// Mock for Telegram bot
type MockTelegramBot struct {
	messages     []tgbotapi.Chattable
	shouldFail   bool
	errorMessage string
}

func (m *MockTelegramBot) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	if m.shouldFail {
		return tgbotapi.Message{}, fmt.Errorf("%s", m.errorMessage)
	}
	m.messages = append(m.messages, c)
	return tgbotapi.Message{}, nil
}

// mockAcknowledger implements the Acknowledger interface
type mockAcknowledger struct {
	ackCallback    func(multiple bool) error
	rejectCallback func(requeue bool) error
}

func (m *mockAcknowledger) Ack(tag uint64, multiple bool) error {
	if m.ackCallback != nil {
		return m.ackCallback(multiple)
	}
	return nil
}

func (m *mockAcknowledger) Nack(tag uint64, multiple bool, requeue bool) error {
	return nil
}

func (m *mockAcknowledger) Reject(tag uint64, requeue bool) error {
	if m.rejectCallback != nil {
		return m.rejectCallback(requeue)
	}
	return nil
}

func TestFormatNotificationMessage(t *testing.T) {
	testCases := []struct {
		name     string
		message  Message
		expected string
	}{
		{
			name: "Basic message",
			message: Message{
				Filename: "/path/to/movie.mkv",
				Status:   "success",
				Time:     time.Now().Format(time.RFC3339),
			},
			expected: "🎬 *MKV Processing Complete*",
		},
		{
			name: "Message with special characters",
			message: Message{
				Filename: "/path/to/movie with spaces & special chars.mkv",
				Status:   "success",
				Time:     time.Now().Format(time.RFC3339),
			},
			expected: "movie with spaces & special chars.mkv",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatNotificationMessage(tc.message)

			// Check that the result contains the expected filename
			if !strings.Contains(result, filepath.Base(tc.message.Filename)) {
				t.Errorf("formatNotificationMessage() does not contain filename, got: %v", result)
			}

			// Check that the result contains the status
			if !strings.Contains(result, tc.message.Status) {
				t.Errorf("formatNotificationMessage() does not contain status, got: %v", result)
			}

			// Check for the expected text pattern
			if !strings.Contains(result, tc.expected) {
				t.Errorf("formatNotificationMessage() does not contain %v, got: %v", tc.expected, result)
			}
		})
	}
}

func TestSendTelegramNotification(t *testing.T) {
	// Setup
	cfg = &config.Config{
		Telegram: config.TelegramConfig{
			ChatID: 12345,
		},
	}

	testCases := []struct {
		name        string
		bot         *MockTelegramBot
		message     string
		shouldError bool
	}{
		{
			name:        "Successful notification",
			bot:         &MockTelegramBot{shouldFail: false},
			message:     "Test notification",
			shouldError: false,
		},
		{
			name:        "Failed notification",
			bot:         &MockTelegramBot{shouldFail: true, errorMessage: "API error"},
			message:     "Test notification",
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := sendTelegramNotification(tc.bot, tc.message)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error but got: %v", err)
				}

				// Verify the message was sent
				if len(tc.bot.messages) != 1 {
					t.Errorf("Expected 1 message to be sent, got %d", len(tc.bot.messages))
				}
			}
		})
	}
}

func TestIsInequivalentArgError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "PRECONDITION_FAILED error",
			err:      &amqp.Error{Code: 406, Reason: "PRECONDITION_FAILED"},
			expected: true,
		},
		{
			name:     "inequivalent arg error",
			err:      &amqp.Error{Code: 406, Reason: "inequivalent arg 'x-dead-letter-exchange'"},
			expected: true,
		},
		{
			name:     "other amqp error",
			err:      &amqp.Error{Code: 404, Reason: "NOT_FOUND"},
			expected: false,
		},
		{
			name:     "non-amqp error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isInequivalentArgError(tc.err)
			if result != tc.expected {
				t.Errorf("isInequivalentArgError() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestParseMessage(t *testing.T) {
	testCases := []struct {
		name        string
		message     []byte
		expectError bool
		expected    Message
	}{
		{
			name:        "Valid message",
			message:     []byte(`{"filename": "/path/to/file.mkv", "status": "success", "time": "2023-07-30T12:00:00Z"}`),
			expectError: false,
			expected: Message{
				Filename: "/path/to/file.mkv",
				Status:   "success",
				Time:     "2023-07-30T12:00:00Z",
			},
		},
		{
			name:        "Invalid JSON",
			message:     []byte(`{invalid json}`),
			expectError: true,
		},
		{
			name:        "Missing required field",
			message:     []byte(`{"status": "success", "time": "2023-07-30T12:00:00Z"}`),
			expectError: false, // JSON parsing succeeds but validation would fail
			expected: Message{
				Status: "success",
				Time:   "2023-07-30T12:00:00Z",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var msg Message
			err := json.Unmarshal(tc.message, &msg)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error but got: %v", err)
				}

				// Verify parsed message
				if msg.Filename != tc.expected.Filename {
					t.Errorf("Expected Filename=%q, got %q", tc.expected.Filename, msg.Filename)
				}

				if msg.Status != tc.expected.Status {
					t.Errorf("Expected Status=%q, got %q", tc.expected.Status, msg.Status)
				}

				if msg.Time != tc.expected.Time {
					t.Errorf("Expected Time=%q, got %q", tc.expected.Time, msg.Time)
				}

				// Validate required fields would succeed/fail as expected
				if tc.expected.Filename == "" && msg.Filename == "" {
					// This message would fail validation in the main function
					t.Logf("Note: This message would fail validation due to missing required field 'filename'")
				}
			}
		})
	}
}

func TestMessageIntegration(t *testing.T) {
	// Setup test config
	cfg = &config.Config{
		Telegram: config.TelegramConfig{
			ChatID: 12345,
		},
	}

	// Test with a valid message
	validMessage := Message{
		Filename: "/path/to/movie.mkv",
		Status:   "success",
		Time:     time.Now().Format(time.RFC3339),
	}

	// Convert to JSON
	messageJSON, _ := json.Marshal(validMessage)

	// Create a mock bot
	mockBot := &MockTelegramBot{}

	// Test formatting and notification logic
	notificationText := formatNotificationMessage(validMessage)
	err := sendTelegramNotification(mockBot, notificationText)

	if err != nil {
		t.Errorf("sendTelegramNotification() failed: %v", err)
	}

	// Verify notification format contains expected elements
	if !strings.Contains(notificationText, "MKV Processing Complete") {
		t.Errorf("Notification doesn't contain expected title")
	}

	if !strings.Contains(notificationText, filepath.Base(validMessage.Filename)) {
		t.Errorf("Notification doesn't contain filename")
	}

	// Verify bot received the message
	if len(mockBot.messages) != 1 {
		t.Errorf("Expected 1 message to be sent, got %d", len(mockBot.messages))
	}

	// Test message unmarshaling
	var parsedMsg Message
	err = json.Unmarshal(messageJSON, &parsedMsg)
	if err != nil {
		t.Errorf("Failed to unmarshal message: %v", err)
	}

	// Verify unmarshaled message matches original
	if parsedMsg.Filename != validMessage.Filename {
		t.Errorf("Unmarshaled filename doesn't match: got %s, want %s",
			parsedMsg.Filename, validMessage.Filename)
	}

	if parsedMsg.Status != validMessage.Status {
		t.Errorf("Unmarshaled status doesn't match: got %s, want %s",
			parsedMsg.Status, validMessage.Status)
	}

	if parsedMsg.Time != validMessage.Time {
		t.Errorf("Unmarshaled time doesn't match: got %s, want %s",
			parsedMsg.Time, validMessage.Time)
	}
}
