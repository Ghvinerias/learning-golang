# MKV Merge Notifier

A Go application that consumes messages from RabbitMQ when MKV processing is complete and sends Telegram notifications.

## Features

- Consumes messages from RabbitMQ queue when MKV processing is complete
- Sends formatted Telegram notifications
- Dead Letter Queue (DLQ) support for failed message processing
- Configuration via YAML file, environment variables, or .env file

## Configuration

Configuration can be provided in multiple ways (in order of precedence):
1. Environment variables
2. `.env` file in the root directory
3. `config.yml` file in the root directory or `/etc/mkvmerge-notifier/`
4. Default values

### Example Configuration

#### Environment Variables
See `.example.env` for the environment variable format.

#### Config File (YAML)
See `config.example.yml` for the YAML configuration format.

## Running the Application

### Local Development

1. Copy `.example.env` to `.env` and update values
2. Run `go run main.go`

### Docker

1. Build the Docker image:
   ```
   docker build -t mkvmerge-notifier .
   ```

2. Run the container:
   ```
   docker run -d --name mkvmerge-notifier \
     -e RABBITMQ_HOST=your-rabbitmq-host \
     -e RABBITMQ_USERNAME=your-username \
     -e RABBITMQ_PASSWORD=your-password \
     -e TELEGRAM_BOT_TOKEN=your-token \
     -e TELEGRAM_CHAT_ID=your-chat-id \
     mkvmerge-notifier
   ```

### Docker Compose

1. Copy `docker-compose.yml` and update environment variables
2. Run:
   ```
   docker-compose up -d
   ```

## Message Format

The application expects messages in the following JSON format:

```json
{
  "filename": "/path/to/your/file.mkv",
  "status": "Complete",
  "time": "2023-07-29T15:04:05Z"
}
```
