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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mkvmerge-consumer .

# Create a minimal image to run the binary
FROM alpine:latest  

# Install certificates and mkvmerge
RUN apk --no-cache add ca-certificates mkvtoolnix

# Set up configuration directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/mkvmerge-consumer .

# Set the entrypoint
CMD ["./mkvmerge-consumer"]
