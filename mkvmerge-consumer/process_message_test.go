package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDelivery provides a mock of amqp.Delivery for testing
type MockDelivery struct {
	body    []byte
	ackErr  error
	nackErr error
	acked   bool
	nacked  bool
}

func (m *MockDelivery) Ack(multiple bool) error {
	m.acked = true
	return m.ackErr
}

func (m *MockDelivery) Nack(multiple, requeue bool) error {
	m.nacked = true
	return m.nackErr
}

func (m *MockDelivery) Body() []byte {
	return m.body
}

// TestMessageProcessing tests the processMessage function with various scenarios
func TestMessageProcessing(t *testing.T) {
	// Setup test environment
	tempDir, err := os.MkdirTemp("", "mkv-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// Create test torrent directory structure
	torrentName := "test-torrent"
	torrentDir := filepath.Join(tempDir, torrentName)
	err = os.Mkdir(torrentDir, 0755)
	assert.NoError(t, err)

	// Override category map for testing
	originalMap := CategoryPathMap
	CategoryPathMap = map[string]string{
		"test-category": tempDir,
	}
	defer func() { CategoryPathMap = originalMap }() // Restore original after test

	// Create a mock MKV file with some content
	mkvFilePath := filepath.Join(torrentDir, "test.mkv")
	err = os.WriteFile(mkvFilePath, []byte("mock mkv content"), 0644)
	assert.NoError(t, err)

	// Create a test message
	msg := Message{
		TorrentName: torrentName,
		Category:    "test-category",
	}
	msgBytes, err := json.Marshal(msg)
	assert.NoError(t, err)

	// Mock RabbitMQ channel
	mockChannel := new(MockChannel)
	mockChannel.On("Publish", "", doneQueueName, false, false, mock.AnythingOfType("amqp.Publishing")).Return(nil)

	// We would use this delivery in a full test of processMessage
	_ = amqp.Delivery{
		Body: msgBytes,
	}

	// Setup mock command executor
	mockExec := &MockCommandExecutor{
		MockResponses: map[string]MockResponse{
			// Mock mkvmerge -J response
			"mkvmerge -J " + mkvFilePath: {
				Output: []byte(`{
					"tracks": [
						{"id": 0, "type": "video", "properties": {"language": "und"}},
						{"id": 1, "type": "audio", "properties": {"language": "eng"}},
						{"id": 2, "type": "audio", "properties": {"language": "spa"}},
						{"id": 3, "type": "subtitles", "properties": {"language": "eng"}}
					]
				}`),
				Err: nil,
			},
		},
	}
	
	// Save original exec.Command and restore after test
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()
	
	// Override exec.Command to use our mock
	execCommand = func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestExecCommandHelper", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		
		// Setup environment to indicate this is a test
		cmd.Env = []string{
			"GO_TEST_HELPER_PROCESS=1",
			"STDOUT=" + string(mockExec.MockResponses["mkvmerge -J "+mkvFilePath].Output),
		}
		return cmd
	}
	
	// TODO: Complete this test by mocking the mkvmerge command execution
	// This requires more setup for the exec.Command mock
}

// TestExecCommandHelper is a helper function that's called by the mocked exec.Command
// It outputs the predefined response based on environment variables
func TestExecCommandHelper(t *testing.T) {
	if os.Getenv("GO_TEST_HELPER_PROCESS") != "1" {
		return
	}
	
	// Get the stdout content from env
	stdout := os.Getenv("STDOUT")
	if stdout != "" {
		os.Stdout.WriteString(stdout)
	}
	
	// Get desired exit code from env
	os.Exit(0)
}

// Override exec.Command for testing
var execCommand = exec.Command
