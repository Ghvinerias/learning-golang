package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Create a mock for exec.Command
func mockExecCommand(mockCmd *MockCmd, command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{
		"GO_WANT_HELPER_PROCESS=1",
	}
	mockCmd.Commands = append(mockCmd.Commands, command)
	mockCmd.Args = append(mockCmd.Args, args)
	return cmd
}

// MockCmd stores the commands and arguments executed
type MockCmd struct {
	Commands []string
	Args     [][]string
}

// Test helper function for mocking exec.Command
// You would implement this in a separate test file (e.g., exec_test.go)
// For completeness, I'm including it here
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	// Parse the command being "executed"
	args := os.Args
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	switch cmd {
	case "mkvmerge":
		if args[0] == "-J" {
			// Return mock track info JSON
			fmt.Println(`{
				"tracks": [
					{"id": 0, "type": "video", "properties": {"language": "eng"}},
					{"id": 1, "type": "audio", "properties": {"language": "eng"}},
					{"id": 2, "type": "audio", "properties": {"language": "spa"}},
					{"id": 3, "type": "subtitles", "properties": {"language": "eng"}}
				]
			}`)
			os.Exit(0)
		} else if args[0] == "-o" {
			// Simulate successful mkvmerge execution
			// Create an empty output file to simulate success
			file, err := os.Create(args[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
				os.Exit(1)
			}
			file.Close()
			fmt.Println("Created mock output file")
			os.Exit(0)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unrecognized command: %s\n", cmd)
		os.Exit(2)
	}
}

// MockFileSystem provides mock file system operations
type MockFileSystem struct {
	mock.Mock
	TempFiles []string
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	args := m.Called(name)
	return args.Get(0).(os.FileInfo), args.Error(1)
}

func (m *MockFileSystem) Walk(root string, fn filepath.WalkFunc) error {
	args := m.Called(root, fn)
	return args.Error(0)
}

func (m *MockFileSystem) Rename(oldpath, newpath string) error {
	args := m.Called(oldpath, newpath)
	return args.Error(0)
}

func (m *MockFileSystem) Remove(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// MockFileInfo implements os.FileInfo for testing
type MockFileInfo struct {
	FileName    string
	FileSize    int64
	FileMode    os.FileMode
	FileModTime time.Time
	FileIsDir   bool
	FileSys     interface{}
}

func (m MockFileInfo) Name() string       { return m.FileName }
func (m MockFileInfo) Size() int64        { return m.FileSize }
func (m MockFileInfo) Mode() os.FileMode  { return m.FileMode }
func (m MockFileInfo) ModTime() time.Time { return m.FileModTime }
func (m MockFileInfo) IsDir() bool        { return m.FileIsDir }
func (m MockFileInfo) Sys() interface{}   { return m.FileSys }

// Test suite for processMessage
type ProcessMessageTestSuite struct {
	suite.Suite
	mockChannel *MockChannelInterface
	mockFS      *MockFileSystem
	mockCmd     *MockCmd
	origExec    func(string, ...string) *exec.Cmd
}

func (suite *ProcessMessageTestSuite) SetupTest() {
	suite.mockChannel = new(MockChannelInterface)
	suite.mockFS = new(MockFileSystem)
	suite.mockCmd = new(MockCmd)

	// Set up global variables for testing
	CategoryPathMap = map[string]string{
		"test-category": "/test/path",
	}
	queueName = "test-queue"
	doneQueueName = "test-done"
	dlqQueueName = "test-dlq"

	// Mock os.Stat
	statFunc = suite.mockFS.Stat

	// Mock filepath.Walk
	walkFunc = suite.mockFS.Walk

	// Mock os.Rename
	renameFunc = suite.mockFS.Rename

	// Mock os.Remove
	removeFunc = suite.mockFS.Remove

	// Store original exec.Command and replace with mock
	suite.origExec = execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		return mockExecCommand(suite.mockCmd, command, args...)
	}
}

func (suite *ProcessMessageTestSuite) TearDownTest() {
	// Restore original functions
	execCommand = suite.origExec
	statFunc = os.Stat
	walkFunc = filepath.Walk
	renameFunc = os.Rename
	removeFunc = os.Remove
}

// Test processing a valid message
func (suite *ProcessMessageTestSuite) TestProcessMessageSuccess() {
	// Create test message
	message := Message{
		TorrentName: "test-movie",
		Category:    "test-category",
	}

	// Marshal message to JSON
	body, err := json.Marshal(message)
	assert.NoError(suite.T(), err)

	// Note: We'll use mockAcker for the delivery

	// Mock os.Stat to return success
	mockFileInfo := MockFileInfo{
		FileName:  "test-movie",
		FileIsDir: true,
	}
	suite.mockFS.On("Stat", "/test/path/test-movie").Return(mockFileInfo, nil)

	// Mock filepath.Walk to simulate finding MKV files
	suite.mockFS.On("Walk", "/test/path/test-movie", mock.AnythingOfType("filepath.WalkFunc")).Return(nil).Run(func(args mock.Arguments) {
		fn := args.Get(1).(filepath.WalkFunc)
		// Simulate finding an MKV file
		mkvInfo := MockFileInfo{
			FileName:  "movie.mkv",
			FileIsDir: false,
		}
		fn("/test/path/test-movie/movie.mkv", mkvInfo, nil)
	})

	// Mock os.Rename to simulate successful file replacement
	suite.mockFS.On("Rename", "/test/path/test-movie/.movie.mkv.tmp.mkv", "/test/path/test-movie/movie.mkv").Return(nil)

	// Mock publishDoneMessage
	suite.mockChannel.On("Publish",
		"",
		"test-done",
		false,
		false,
		mock.MatchedBy(func(msg amqp.Publishing) bool {
			var message map[string]string
			err := json.Unmarshal(msg.Body, &message)
			return err == nil && message["filename"] == "test-movie"
		})).Return(nil)

	// Set up the mock delivery to expect Ack
	mockAcker := new(MockDelivery)
	mockAcker.On("Ack", false).Return(nil)

	// Process the message
	processMessage(suite.mockChannel, mockAcker, body)

	// Verify expectations
	suite.mockFS.AssertExpectations(suite.T())
	suite.mockChannel.AssertExpectations(suite.T())
	mockAcker.AssertExpectations(suite.T())
}

// Test processing a message with invalid JSON
func (suite *ProcessMessageTestSuite) TestProcessMessageInvalidJSON() {
	// Create invalid JSON
	body := []byte(`{"torrentName": "test-movie", "category": "test-category"`) // Missing closing brace

	// Set up the mock delivery to expect Reject
	mockAcker := new(MockDelivery)
	mockAcker.On("Reject", false).Return(nil)

	// Process the message
	processMessage(suite.mockChannel, mockAcker, body)

	// Verify expectations
	mockAcker.AssertExpectations(suite.T())
}

// Test processing a message with unknown category
func (suite *ProcessMessageTestSuite) TestProcessMessageUnknownCategory() {
	// Create test message with unknown category
	message := Message{
		TorrentName: "test-movie",
		Category:    "unknown-category",
	}

	// Marshal message to JSON
	body, err := json.Marshal(message)
	assert.NoError(suite.T(), err)

	// Set up the mock delivery to expect Reject
	mockAcker := new(MockDelivery)
	mockAcker.On("Reject", false).Return(nil)

	// Process the message
	processMessage(suite.mockChannel, mockAcker, body)

	// Verify expectations
	mockAcker.AssertExpectations(suite.T())
}

// Test processing a message with non-existent folder
func (suite *ProcessMessageTestSuite) TestProcessMessageNonExistentFolder() {
	// Create test message
	message := Message{
		TorrentName: "test-movie",
		Category:    "test-category",
	}

	// Marshal message to JSON
	body, err := json.Marshal(message)
	assert.NoError(suite.T(), err)

	// Mock os.Stat to return error
	suite.mockFS.On("Stat", "/test/path/test-movie").Return(MockFileInfo{}, fs.ErrNotExist)

	// Set up the mock delivery to expect Reject
	mockAcker := new(MockDelivery)
	mockAcker.On("Reject", false).Return(nil)

	// Process the message
	processMessage(suite.mockChannel, mockAcker, body)

	// Verify expectations
	suite.mockFS.AssertExpectations(suite.T())
	mockAcker.AssertExpectations(suite.T())
}

// Test processing a message with no MKV files
func (suite *ProcessMessageTestSuite) TestProcessMessageNoMKVFiles() {
	// Create test message
	message := Message{
		TorrentName: "test-movie",
		Category:    "test-category",
	}

	// Marshal message to JSON
	body, err := json.Marshal(message)
	assert.NoError(suite.T(), err)

	// Mock os.Stat to return success
	mockFileInfo := MockFileInfo{
		FileName:  "test-movie",
		FileIsDir: true,
	}
	suite.mockFS.On("Stat", "/test/path/test-movie").Return(mockFileInfo, nil)

	// Mock filepath.Walk to simulate finding no MKV files
	suite.mockFS.On("Walk", "/test/path/test-movie", mock.AnythingOfType("filepath.WalkFunc")).Return(nil)

	// Set up the mock delivery to expect Ack
	mockAcker := new(MockDelivery)
	mockAcker.On("Ack", false).Return(nil)

	// Process the message
	processMessage(suite.mockChannel, mockAcker, body)

	// Verify expectations
	suite.mockFS.AssertExpectations(suite.T())
	mockAcker.AssertExpectations(suite.T())
}

func TestProcessMessageSuite(t *testing.T) {
	suite.Run(t, new(ProcessMessageTestSuite))
}

// Using the variables declared in main.go
