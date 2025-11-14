# AD Password Reset Telegram Bot

A minimal Golang Telegram bot that connects to Windows Active Directory to reset user passwords.

## Features

- `/reset <username>` - Resets user password to "Aa12345@"
- Supports usernames in formats: `john.doe` or `john.doe@example.com`

## Setup

1. Create a Telegram bot via [@BotFather](https://t.me/botfather) and get the token

2. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Install dependencies:
```bash
go mod tidy
```

4. Run the bot:
```bash
go run main.go
```

## Environment Variables

- `TELEGRAM_TOKEN` - Your Telegram bot token
- `AD_SERVER` - Active Directory server hostname
- `AD_PORT` - LDAP port (default: 389)
- `BIND_DN` - Service account DN for AD binding
- `BIND_PASSWORD` - Service account password
- `BASE_DN` - Base DN for user searches

## Usage

Send `/reset john.doe` or `/reset john.doe@example.com` to reset the user's password.

## Security Notes

- Use a dedicated service account with minimal required permissions
- Consider using LDAPS (port 636) for encrypted connections
- Restrict bot access to authorized users only
