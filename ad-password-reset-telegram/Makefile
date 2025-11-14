APP_NAME=ad-password-reset-bot
VERSION=latest

.PHONY: build-linux build-windows build-docker clean

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/$(APP_NAME)-linux main.go

# Build for Windows
build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/$(APP_NAME)-windows.exe main.go

# Build Docker image
build-docker:
	docker build -t $(APP_NAME):$(VERSION) .

# Build all targets
build-all: build-linux build-windows build-docker

# Clean build artifacts
clean:
	rm -rf bin/
	docker rmi $(APP_NAME):$(VERSION) 2>/dev/null || true

# Run locally
run:
	go run main.go

# Install dependencies
deps:
	go mod tidy
