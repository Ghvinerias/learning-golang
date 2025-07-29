# Telegram Torrent Bot

A Telegram bot that downloads torrent files from Kinozal and adds them to qBittorrent with specified categories.

## Features
 
- Download torrents from Kinozal with proper authentication
- Add torrents to qBittorrent with specific categories
- Support for movies and TV shows categories
- Docker support for easy deployment

## Commands

- `/help` - Show help message
- `/movies [URL]` - Download torrent and add to qBittorrent with "local-movies" category
- `/tvshows [URL]` - Download torrent and add to qBittorrent with "local-tvshows" category

## Required Environment Variables

| Variable | Description |
|----------|-------------|
| `TELEGRAM_BOT_TOKEN` | Your Telegram Bot token from BotFather |
| `KINOZAL_LOGIN_URL` | URL for Kinozal login page (https://kinozal.guru/login.php) |
| `QBITTORRENT_URL` | URL for qBittorrent WebUI (including protocol and port) |
| `QBITTORRENT_USER` | qBittorrent WebUI username |
| `QBITTORRENT_PASS` | qBittorrent WebUI password |
| `KINOZAL_USER` | Kinozal username |
| `KINOZAL_PASS` | Kinozal password |
| `TORRENT_TEMP_FILE` | Path where temporary torrent files will be stored |

## Setup Options

### Docker Setup (Recommended)

1. Edit the `docker-compose.yml` file to set your environment variables:
   ```yaml
   environment:
     - TELEGRAM_BOT_TOKEN=your_telegram_bot_token
     - KINOZAL_USER=your_kinozal_username
     - KINOZAL_PASS=your_kinozal_password
   ```

2. Create required directories:
   ```bash
   mkdir -p tmp qbittorrent/config qbittorrent/downloads
   ```

3. Run with docker-compose:
   ```bash
   docker-compose up -d
   ```

4. Access qBittorrent WebUI at http://localhost:8080 (default credentials: admin/adminadmin)

### Manual Setup

1. Clone this repository
2. Install dependencies: `go mod tidy`
3. Set required environment variables:
   ```bash
   export TELEGRAM_BOT_TOKEN=your_bot_token_here
   export KINOZAL_LOGIN_URL=https://kinozal.guru/login.php
   export QBITTORRENT_URL=http://localhost:8080
   export QBITTORRENT_USER=admin
   export QBITTORRENT_PASS=adminadmin
   export KINOZAL_USER=your_kinozal_username
   export KINOZAL_PASS=your_kinozal_password
   export TORRENT_TEMP_FILE=/tmp/temp.torrent
   ```
4. Build and run the bot:
   ```bash
   make run
   ```

## Requirements

- Go 1.16 or higher
- A valid Telegram bot token (get one from @BotFather)
- Access to Kinozal and qBittorrent
