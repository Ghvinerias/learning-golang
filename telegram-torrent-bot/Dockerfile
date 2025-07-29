FROM golang:1.24-alpine AS builder

# Install git and build dependencies
RUN apk add --no-cache git build-base

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o telegram-torrent-bot .

# Create a minimal image to run the binary
FROM alpine:latest  

# Install certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/telegram-torrent-bot .

# Create a directory for temporary torrent files
RUN mkdir -p /tmp/torrents

# Set the entrypoint
CMD ["./telegram-torrent-bot"]
