package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	kinozalLoginURL string
	qbittorrentURL  string
	qbittorrentUser string
	qbittorrentPass string
	kinozalUser     string
	kinozalPass     string
	torrentTempFile string
)

func main() {
	// Get Telegram bot token from environment variable
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}
	kinozalLoginURL = os.Getenv("KINOZAL_LOGIN_URL")
	if kinozalLoginURL == "" {
		log.Fatal("KINOZAL_LOGIN_URL environment variable is required")
	}
	qbittorrentURL = os.Getenv("QBITTORRENT_URL")
	if qbittorrentURL == "" {
		log.Fatal("QBITTORRENT_URL environment variable is required")
	}
	qbittorrentUser = os.Getenv("QBITTORRENT_USER")
	if qbittorrentUser == "" {
		log.Fatal("QBITTORRENT_USER environment variable is required")
	}
	qbittorrentPass = os.Getenv("QBITTORRENT_PASS")
	if qbittorrentPass == "" {
		log.Fatal("QBITTORRENT_PASS environment variable is required")
	}
	kinozalUser = os.Getenv("KINOZAL_USER")
	if kinozalUser == "" {
		log.Fatal("KINOZAL_USER environment variable is required")
	}
	kinozalPass = os.Getenv("KINOZAL_PASS")
	if kinozalPass == "" {
		log.Fatal("KINOZAL_PASS environment variable is required")
	}
	torrentTempFile = os.Getenv("TORRENT_TEMP_FILE")
	if torrentTempFile == "" {
		log.Fatal("TORRENT_TEMP_FILE environment variable is required")
	}
	// Create a new bot instance
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	// Set this to true for debugging
	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Create a new update configuration
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	// Start receiving updates
	updates := bot.GetUpdatesChan(updateConfig)

	// Process incoming updates
	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Check if the message has text
		if update.Message.Text == "" {
			continue
		}

		// Handle commands
		if update.Message.IsCommand() {
			command := update.Message.Command()
			args := update.Message.CommandArguments()

			switch command {
			case "start", "help":
				handleHelpCommand(bot, update.Message)
			case "movies":
				if args == "" {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Please provide a torrent URL. Usage: /movies URL"))
					continue
				}
				handleTorrentCommand(bot, update.Message, args, "local-movies")
			case "tvshows":
				if args == "" {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Please provide a torrent URL. Usage: /tvshows URL"))
					continue
				}
				handleTorrentCommand(bot, update.Message, args, "local-tvshows")
			default:
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command. Type /help for available commands."))
			}
		}
	}
}

func handleHelpCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	helpText := `
Available commands:
/help - Show this help message
/movies [URL] - Download torrent and add to qBittorrent with "local-movies" category
/tvshows [URL] - Download torrent and add to qBittorrent with "local-tvshows" category

Example:
/movies https://dl.kinozal.guru/download.php?id=2094298
`
	msg := tgbotapi.NewMessage(message.Chat.ID, helpText)
	bot.Send(msg)
}

func handleTorrentCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message, torrentURL string, category string) {
	// Inform user that processing has started
	statusMsg, _ := bot.Send(tgbotapi.NewMessage(message.Chat.ID, "⏳ Processing your request..."))

	// Clean up URL - remove any quotes that might have been included in the command
	torrentURL = strings.Trim(torrentURL, "\"'")

	// Create a cookie jar to handle cookies automatically
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	// Parse the base URL from login URL
	parsedURL, err := url.Parse(kinozalLoginURL)
	if err != nil {
		updateMessage(bot, message.Chat.ID, statusMsg.MessageID, "❌ Error parsing login URL: "+err.Error())
		return
	}
	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	// First visit the login page to get cookies
	log.Printf("Visiting login page: %s", kinozalLoginURL)
	_, err = client.Get(kinozalLoginURL)
	if err != nil {
		updateMessage(bot, message.Chat.ID, statusMsg.MessageID, "❌ Error accessing login page: "+err.Error())
		return
	}

	// Submit login form
	loginEndpoint := baseURL + "/takelogin.php"
	log.Printf("Submitting login form to: %s", loginEndpoint)

	// Create login form data
	formData := url.Values{
		"username": {kinozalUser},
		"password": {kinozalPass},
		"returnto": {"/"},
	}

	loginResp, err := client.PostForm(loginEndpoint, formData)
	if err != nil {
		updateMessage(bot, message.Chat.ID, statusMsg.MessageID, "❌ Error submitting login form: "+err.Error())
		return
	}
	defer loginResp.Body.Close()

	log.Printf("Login response status: %d", loginResp.StatusCode)

	// Now download the torrent file using the same client with cookies
	updateMessage(bot, message.Chat.ID, statusMsg.MessageID, "⏳ Downloading torrent file...")
	log.Printf("Downloading torrent from: %s", torrentURL)

	// Use the same client with cookies
	torrentResp, err := client.Get(torrentURL)
	if err != nil {
		updateMessage(bot, message.Chat.ID, statusMsg.MessageID, "❌ Error downloading torrent: "+err.Error())
		return
	}
	defer torrentResp.Body.Close()

	// Check if we got a valid response
	if torrentResp.StatusCode != http.StatusOK {
		updateMessage(bot, message.Chat.ID, statusMsg.MessageID, fmt.Sprintf("❌ Failed to download torrent. Status: %d", torrentResp.StatusCode))
		return
	}

	// Check content type to verify we got a torrent file
	contentType := torrentResp.Header.Get("Content-Type")
	log.Printf("Downloaded content type: %s", contentType)

	if strings.Contains(contentType, "text/html") {
		// Got HTML instead of a torrent file - save for debugging
		body, _ := io.ReadAll(torrentResp.Body)
		debugPath := torrentTempFile + ".html"
		os.WriteFile(debugPath, body, 0644)
		log.Printf("Received HTML instead of torrent file. Saved to: %s", debugPath)
		updateMessage(bot, message.Chat.ID, statusMsg.MessageID, "❌ Authentication failed - received HTML instead of torrent file")
		return
	}

	// Save the torrent file
	log.Printf("Writing torrent file to: %s", torrentTempFile)
	torrentFile, err := os.Create(torrentTempFile)
	if err != nil {
		updateMessage(bot, message.Chat.ID, statusMsg.MessageID, "❌ Error creating temporary file: "+err.Error())
		return
	}

	bytesWritten, err := io.Copy(torrentFile, torrentResp.Body)
	torrentFile.Close()
	if err != nil {
		updateMessage(bot, message.Chat.ID, statusMsg.MessageID, "❌ Error saving torrent data: "+err.Error())
		return
	}

	log.Printf("Downloaded %d bytes", bytesWritten)
	if bytesWritten < 100 {
		updateMessage(bot, message.Chat.ID, statusMsg.MessageID, "❌ Downloaded file is too small to be a valid torrent")
		return
	}

	// Upload to qBittorrent
	updateMessage(bot, message.Chat.ID, statusMsg.MessageID, "⏳ Adding torrent to qBittorrent...")

	// Create a new HTTP client for qBittorrent
	qbClient := &http.Client{}

	// Login to qBittorrent WebUI
	qbLoginData := url.Values{
		"username": {qbittorrentUser},
		"password": {qbittorrentPass},
	}

	qbLoginResp, err := qbClient.PostForm(qbittorrentURL+"/api/v2/auth/login", qbLoginData)
	if err != nil {
		updateMessage(bot, message.Chat.ID, statusMsg.MessageID, "❌ Error logging in to qBittorrent: "+err.Error())
		return
	}
	defer qbLoginResp.Body.Close()

	if qbLoginResp.StatusCode != http.StatusOK {
		updateMessage(bot, message.Chat.ID, statusMsg.MessageID, fmt.Sprintf("❌ Failed to login to qBittorrent. Status: %d", qbLoginResp.StatusCode))
		return
	}

	// Get cookies from login
	qbCookies := qbLoginResp.Cookies()

	// Create multipart form for uploading
	var requestBody strings.Builder
	multipartWriter := multipart.NewWriter(&requestBody)

	// Add torrent file to form
	fileWriter, err := multipartWriter.CreateFormFile("torrents", filepath.Base(torrentTempFile))
	if err != nil {
		updateMessage(bot, message.Chat.ID, statusMsg.MessageID, "❌ Error creating form file: "+err.Error())
		return
	}

	// Read torrent file
	torrentData, err := os.ReadFile(torrentTempFile)
	if err != nil {
		updateMessage(bot, message.Chat.ID, statusMsg.MessageID, "❌ Error reading torrent file: "+err.Error())
		return
	}

	// Write torrent data to form
	_, err = fileWriter.Write(torrentData)
	if err != nil {
		updateMessage(bot, message.Chat.ID, statusMsg.MessageID, "❌ Error writing torrent data to form: "+err.Error())
		return
	}

	// Add other form fields
	multipartWriter.WriteField("category", category)
	multipartWriter.WriteField("autoTMM", "true")

	// Close the multipart writer
	multipartWriter.Close()

	// Create upload request
	uploadReq, err := http.NewRequest("POST", qbittorrentURL+"/api/v2/torrents/add", strings.NewReader(requestBody.String()))
	if err != nil {
		updateMessage(bot, message.Chat.ID, statusMsg.MessageID, "❌ Error creating upload request: "+err.Error())
		return
	}

	// Set content type
	uploadReq.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	// Add cookies
	for _, cookie := range qbCookies {
		uploadReq.AddCookie(cookie)
	}

	// Execute upload request
	uploadResp, err := qbClient.Do(uploadReq)
	if err != nil {
		updateMessage(bot, message.Chat.ID, statusMsg.MessageID, "❌ Error uploading torrent: "+err.Error())
		return
	}
	defer uploadResp.Body.Close()

	// Check response
	respBody, _ := io.ReadAll(uploadResp.Body)
	if uploadResp.StatusCode != http.StatusOK {
		log.Printf("Upload failed with status %d: %s", uploadResp.StatusCode, string(respBody))
		updateMessage(bot, message.Chat.ID, statusMsg.MessageID, fmt.Sprintf("❌ Failed to add torrent. Status: %d, Response: %s", uploadResp.StatusCode, string(respBody)))
		return
	}

	log.Printf("Upload response: %s", string(respBody))

	// Clean up temporary file
	os.Remove(torrentTempFile)

	// Success message
	successMsg := fmt.Sprintf("✅ Torrent successfully added to qBittorrent with category: %s", category)
	updateMessage(bot, message.Chat.ID, statusMsg.MessageID, successMsg)
}

// Helper function to update an existing message
func updateMessage(bot *tgbotapi.BotAPI, chatID int64, messageID int, text string) {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	bot.Send(msg)
}
