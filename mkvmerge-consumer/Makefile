# Golang Project Makefile
# -----------------
# Project: mkvmerge-consumer
# Date: June 30, 2025

# Binary name
BINARY_NAME := mkvmerge-consumer
WINDOWS_BINARY := $(BINARY_NAME).exe
LINUX_BINARY := $(BINARY_NAME)

# Build directories
BUILD_DIR := build
WINDOWS_DIR := $(BUILD_DIR)/windows
LINUX_DIR := $(BUILD_DIR)/linux

# Environment settings
GO := go
GOOS_WINDOWS := GOOS=windows
GOOS_LINUX := GOOS=linux
GOARCH := GOARCH=amd64

# Silence command echoing
.SILENT:

# Declare phony targets
.PHONY: all clean build build-win build-linux test test-coverage test-coverage-html tidy help

# Default target
all: clean build

# Build for all platforms
build: build-win build-linux
	@echo "Build completed for all platforms"

# Build for Windows
build-win:
	@echo "Building for Windows..."
	mkdir -p $(WINDOWS_DIR)
	$(GOOS_WINDOWS) $(GOARCH) $(GO) build -o $(WINDOWS_DIR)/$(WINDOWS_BINARY) main.go
	@echo "Windows build complete: $(WINDOWS_DIR)/$(WINDOWS_BINARY)"

# Build for Linux
build-linux:
	@echo "Building for Linux..."
	mkdir -p $(LINUX_DIR)
	$(GOOS_LINUX) $(GOARCH) $(GO) build -o $(LINUX_DIR)/$(LINUX_BINARY) main.go
	@echo "Linux build complete: $(LINUX_DIR)/$(LINUX_BINARY)"

# Clean build artifacts
clean:
	@echo "Cleaning build directories..."
	rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...
	@echo "Tests complete"

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -cover -v ./...
	@echo "Coverage tests complete"

# Run tests with coverage and generate HTML report
test-coverage-html:
	@echo "Generating coverage report..."
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	rm coverage.out
	@echo "Coverage report generated: coverage.html"

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GO) mod tidy
	@echo "Dependencies updated"

# Help documentation
help:
	@echo "Available targets:"
	@echo "  all        - Clean and build for all platforms (default)"
	@echo "  build      - Build for all platforms"
	@echo "  build-win  - Build for Windows"
	@echo "  build-linux- Build for Linux"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  tidy       - Tidy Go module dependencies"
	@echo "  help       - Display this help message"
	@echo "  test-coverage - Run tests with coverage summary"
	@echo "  test-coverage-html - Generate HTML coverage report"