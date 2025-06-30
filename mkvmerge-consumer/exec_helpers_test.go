package main

import (
	"os/exec"
)

// CommandExecutor defines an interface for executing commands
type CommandExecutor interface {
	ExecuteCommand(name string, args ...string) ([]byte, error)
}

// RealCommandExecutor executes actual commands using os/exec
type RealCommandExecutor struct{}

func (r *RealCommandExecutor) ExecuteCommand(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

// MockCommandExecutor mocks command execution for testing
type MockCommandExecutor struct {
	MockResponses map[string]MockResponse
}

type MockResponse struct {
	Output []byte
	Err    error
}

// ExecuteCommand returns mock responses based on the command name and arguments
func (m *MockCommandExecutor) ExecuteCommand(name string, args ...string) ([]byte, error) {
	// Create a key combining command name and arguments
	key := name
	for _, arg := range args {
		key += " " + arg
	}

	// Check if we have a mock response for this command
	if response, ok := m.MockResponses[key]; ok {
		return response.Output, response.Err
	}

	// Return empty response if command not mocked
	return []byte{}, nil
}

var currentExecutor CommandExecutor = &RealCommandExecutor{}

// SetCommandExecutor sets the executor for commands
func SetCommandExecutor(executor CommandExecutor) {
	currentExecutor = executor
}
