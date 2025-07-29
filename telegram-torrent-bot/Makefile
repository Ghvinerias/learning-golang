.PHONY: build run clean deps docker-build docker-run docker-compose-up docker-compose-down docker-logs

# Default make target
all: build

# Build the application
build:
	go build -o telegram-torrent-bot

# Run the application
run: build
	./telegram-torrent-bot

# Clean build artifacts
clean:
	rm -f telegram-torrent-bot
	rm -f movie.torrent
	rm -f temp.torrent

# Download dependencies
deps:
	go mod tidy

# Docker targets
docker-build:
	docker build -t telegram-torrent-bot .

# Run Docker container directly
docker-run: docker-build
	docker run -d \
	--name telegram-torrent-bot \
	--env-file .env \
	telegram-torrent-bot

# Docker Compose targets
docker-compose-up:
	mkdir -p tmp qbittorrent/config qbittorrent/downloads
	docker-compose up -d

docker-compose-down:
	docker-compose down

docker-logs:
	docker-compose logs -f
