package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken   string
	ADServer        string
	ADPort          string
	BindDN          string
	BindPassword    string
	BaseDN          string
	DefaultPassword string
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Username  string `json:"username,omitempty"`
	Error     string `json:"error,omitempty"`
}

func logToFile(level, message, username, errorMsg string) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Username:  username,
		Error:     errorMsg,
	}

	// Log to console
	if errorMsg != "" {
		log.Printf("[%s] %s - %s: %s", level, username, message, errorMsg)
	} else {
		log.Printf("[%s] %s - %s", level, username, message)
	}

	// Log to file
	file, err := os.OpenFile("bot.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}
	defer file.Close()

	jsonData, _ := json.Marshal(entry)
	file.WriteString(string(jsonData) + "\n")
}

func loadConfig() *Config {
	return &Config{
		TelegramToken:   os.Getenv("TELEGRAM_TOKEN"),
		ADServer:        os.Getenv("AD_SERVER"),
		ADPort:          getEnvOrDefault("AD_PORT", "389"),
		BindDN:          os.Getenv("BIND_DN"),
		BindPassword:    os.Getenv("BIND_PASSWORD"),
		BaseDN:          os.Getenv("BASE_DN"),
		DefaultPassword: os.Getenv("DEFAULT_PASSWORD"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func connectToAD(config *Config) (*ldap.Conn, error) {
	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%s", config.ADServer, config.ADPort))
	if err != nil {
		return nil, err
	}

	err = conn.Bind(config.BindDN, config.BindPassword)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func resetUserPassword(config *Config, username string) error {
	logToFile("INFO", "Starting password reset", username, "")

	conn, err := connectToAD(config)
	if err != nil {
		logToFile("ERROR", "Failed to connect to AD", username, err.Error())
		return fmt.Errorf("failed to connect to AD: %v", err)
	}
	defer conn.Close()

	// Handle both formats: john.doe@example.com and john.doe
	searchUsername := username
	if strings.Contains(username, "@") {
		searchUsername = strings.Split(username, "@")[0]
	}

	logToFile("INFO", "Searching for user", searchUsername, "")

	// Search for user
	searchRequest := ldap.NewSearchRequest(
		config.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(sAMAccountName=%s)", searchUsername),
		[]string{"dn"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		logToFile("ERROR", "User search failed", searchUsername, err.Error())
		return fmt.Errorf("search failed: %v", err)
	}

	if len(sr.Entries) == 0 {
		logToFile("ERROR", "User not found", searchUsername, "")
		return fmt.Errorf("user not found: %s", username)
	}

	userDN := sr.Entries[0].DN
	logToFile("INFO", "User found", searchUsername, "DN: "+userDN)

	// Reset password using userPassword attribute (for plain LDAP)
	modifyRequest := ldap.NewModifyRequest(userDN, nil)
	modifyRequest.Replace("userPassword", []string{getEnvOrDefault("DEFAULT_PASSWORD", "Aa12345@")})
	err = conn.Modify(modifyRequest)
	if err != nil {
		logToFile("ERROR", "Password reset failed", searchUsername, err.Error())
		return fmt.Errorf("password reset failed: %v", err)
	}

	logToFile("INFO", "Password reset successful", searchUsername, "")

	// Clear "user must change password at next login" flag
	modifyRequest = ldap.NewModifyRequest(userDN, nil)
	modifyRequest.Replace("pwdLastSet", []string{"-1"})
	err = conn.Modify(modifyRequest)
	if err != nil {
		logToFile("ERROR", "Failed to clear password change flag", searchUsername, err.Error())
		return fmt.Errorf("failed to clear password change flag: %v", err)
	}

	logToFile("INFO", "Password change flag cleared", searchUsername, "")
	return nil
}

func main() {
	godotenv.Load()

	config := loadConfig()

	if config.TelegramToken == "" {
		log.Fatal("TELEGRAM_TOKEN environment variable is required")
	}

	bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Bot started: %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		switch update.Message.Command() {
		case "reset":
			args := update.Message.CommandArguments()
			if args == "" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Usage: /reset <username>")
				bot.Send(msg)
				continue
			}

			username := strings.TrimSpace(args)
			err := resetUserPassword(config, username)

			var responseText string
			if err != nil {
				responseText = fmt.Sprintf("Failed to reset password for %s: %v", username, err)
			} else {
				responseText = fmt.Sprintf("Password reset successfully for %s", username)
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, responseText)
			bot.Send(msg)
		}
	}
}
