package main

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMockMkvmerge tests the mock structure for mkvmerge commands
func TestMockMkvmerge(t *testing.T) {
	// Create a mock command executor
	
	// Save original command executor
	originalExecutor := currentExecutor
	defer func() { currentExecutor = originalExecutor }()
	
	// Set mock executor
	SetCommandExecutor(&MockCommandExecutor{
		MockResponses: map[string]MockResponse{
			"mkvmerge -J test.mkv": {
				Output: []byte(`{"tracks":[{"id":0,"type":"video","properties":{"language":"und"}}]}`),
				Err: nil,
			},
		},
	})
	
	// Execute the command through our mock executor
	output, err := currentExecutor.ExecuteCommand("mkvmerge", "-J", "test.mkv")
	
	// Verify no error occurred
	assert.NoError(t, err)
	
	// Verify output is as expected
	assert.Contains(t, string(output), "tracks")
}
