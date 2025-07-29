# Golang Project Makefile
# -----------------
# Project: mkvmerge-notifier
# Date: July 29, 2025

# Binary name
BINARY_NAME := mkvmerge-notifier
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

# Docker settings
DOCKER_IMAGE_NAME := slickg/mkvmerge-notifier
DOCKER_TAG := latest

# Silence command echoing
.SILENT:

# Declare phony targets
.PHONY: all clean build build-win build-linux test tidy help docker docker-build docker-push run

# Default target
all: clean build

# Build for all platforms
build: tidy build-win build-linux
	@echo "Build completed for all platforms"

# Build for Windows
build-win:
	@echo "Building for Windows..."
	mkdir -p $(WINDOWS_DIR)
	$(GOOS_WINDOWS) $(GOARCH) $(GO) build -o $(WINDOWS_DIR)/$(WINDOWS_BINARY)
	@echo "Windows build complete: $(WINDOWS_DIR)/$(WINDOWS_BINARY)"

# Build for Linux
build-linux:
	@echo "Building for Linux..."
	mkdir -p $(LINUX_DIR)
	$(GOOS_LINUX) $(GOARCH) $(GO) build -o $(LINUX_DIR)/$(LINUX_BINARY)
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

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GO) mod tidy
	@echo "Dependencies updated"

# Build docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)"

# Push docker image
docker-push:
	@echo "Pushing Docker image..."
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)
	@echo "Docker image pushed: $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)"

# Docker build and push
docker: docker-build docker-push

# Run local binary
run:
	@echo "Running application..."
	$(GO) run main.go
	
# Help documentation
help:
	@echo "Available targets:"
	@echo "  all          - Clean and build for all platforms (default)"
	@echo "  build        - Build for all platforms"
	@echo "  build-win    - Build for Windows"
	@echo "  build-linux  - Build for Linux"
	@echo "  clean        - Clean build directories"
	@echo "  test         - Run tests"
	@echo "  tidy         - Tidy dependencies"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-push  - Push Docker image"
	@echo "  docker       - Build and push Docker image"
	@echo "  run          - Run application locally"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  tidy       - Tidy Go module dependencies"
	@echo "  help       - Display this help message"