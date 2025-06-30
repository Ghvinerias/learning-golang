package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/streadway/amqp"
)

// RabbitMQ connection details (you may want to load these from environment variables)
const (
	rabbitmqHost     = "localhost"
	rabbitmqPort     = "5672"
	rabbitmqUsername = "guest"
	rabbitmqPassword = "guest"
	queueName        = "mkvmerge.tasks" // Queue to consume from
	doneQueueName    = "mkvmerge.done"  // Queue to publish to when done
)

// Message represents the structure of incoming RabbitMQ messages
type Message struct {
	TorrentName string `json:"torrentName"`
	Category    string `json:"category"`
}

// CategoryPathMap maps category names to their filesystem paths
var CategoryPathMap = map[string]string{
	"local-movies":  "/mnt/vault/media/jello/movies",
	"local-tvshows": "/mnt/vault/media/jello/tvshows",
	"nas-tvshows":   "/mnt/vault-media/tvshows",
	"nas-movies":    "/mnt/vault-media/movies",
}

// failOnError logs and exits on error
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// ensureQueueExists creates a queue if it does not already exist
func ensureQueueExists(ch *amqp.Channel, qName string) amqp.Queue {
	q, err := ch.QueueDeclare(
		qName,  // name
		true,   // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	failOnError(err, fmt.Sprintf("Failed to declare queue '%s'", qName))
	log.Printf("Queue '%s' declared", qName)
	return q
}

// publishDoneMessage publishes a message to the done queue
func publishDoneMessage(ch *amqp.Channel, filename string) error {
	// Create a simple message with the filename
	message := map[string]string{
		"filename": filename,
		"status":   "processed",
		"time":     time.Now().Format(time.RFC3339),
	}

	// Convert to JSON
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal done message: %v", err)
	}

	// Publish to the done queue
	err = ch.Publish(
		"",            // exchange
		doneQueueName, // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			DeliveryMode: amqp.Persistent, // make message persistent
		})

	if err != nil {
		return fmt.Errorf("failed to publish done message: %v", err)
	}

	log.Printf("Published completion message for file %s to %s queue", filename, doneQueueName)
	return nil
}

func main() {
	// Set up logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("Starting RabbitMQ consumer...")

	// Connect to RabbitMQ
	connectionString := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		rabbitmqUsername, rabbitmqPassword, rabbitmqHost, rabbitmqPort)
	conn, err := amqp.Dial(connectionString)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	log.Println("Successfully connected to RabbitMQ")

	// Create a channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	log.Println("Channel opened")

	// Declare queues (ensures they exist)
	processingQueue := ensureQueueExists(ch, queueName)
	_ = ensureQueueExists(ch, doneQueueName) // Ensure done queue exists

	// Set QoS (prefetch count)
	err = ch.Qos(
		1,     // prefetch count (process one message at a time)
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	// Register consumer
	msgs, err := ch.Consume(
		processingQueue.Name, // queue
		"",                   // consumer tag (empty means auto-generated)
		false,                // auto-ack (false means manual acknowledgment)
		false,                // exclusive
		false,                // no-local
		false,                // no-wait
		nil,                  // args
	)
	failOnError(err, "Failed to register a consumer")
	log.Println("Consumer registered, waiting for messages...")

	// Create a channel to handle shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Create a channel to process messages
	forever := make(chan bool)
	// Process messages in a goroutine
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			// Process the message and acknowledge only after successful processing
			processMessage(ch, d, d.Body)
		}
	}()

	// Wait for shutdown signal
	go func() {
		sig := <-shutdown
		log.Printf("Received shutdown signal: %v", sig)
		log.Println("Shutting down gracefully...")

		// Close RabbitMQ connection
		if err := conn.Close(); err != nil {
			log.Printf("Error closing RabbitMQ connection: %v", err)
		}

		close(forever)
	}()

	log.Println("Consumer is now running. Press CTRL+C to exit")
	<-forever
	log.Println("Consumer shutdown complete")
}

// processMessage handles the received message
func processMessage(ch *amqp.Channel, d amqp.Delivery, body []byte) {
	log.Printf("Processing message: %s", body)

	// Parse the JSON message
	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("Error parsing message JSON: %v", err)
		return
	}

	// Map category to base directory path
	basePath, exists := CategoryPathMap[msg.Category]
	if !exists {
		log.Printf("Unknown category: %s", msg.Category)
		// Acknowledge the message since the category is unknown and can't be processed
		if err := d.Ack(false); err != nil {
			log.Printf("Error acknowledging message with unknown category: %v", err)
		} else {
			log.Println("Message acknowledged (unknown category)")
		}
		return
	}

	// Construct full folder path
	folderPath := filepath.Join(basePath, msg.TorrentName)
	log.Printf("Looking for MKV files in: %s", folderPath)

	// Check if the folder exists
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		log.Printf("Folder does not exist: %s", folderPath)
		// Acknowledge the message since the folder doesn't exist and can't be processed
		if err := d.Ack(false); err != nil {
			log.Printf("Error acknowledging message with non-existent folder: %v", err)
		} else {
			log.Println("Message acknowledged (folder does not exist)")
		}
		return
	}

	// Find all .mkv files in the folder
	var mkvFiles []string
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".mkv" {
			mkvFiles = append(mkvFiles, path)
		}
		return nil
	})

	if err != nil {
		log.Printf("Error walking directory %s: %v", folderPath, err)
		return
	}

	log.Printf("Found %d MKV files to process", len(mkvFiles))
	if len(mkvFiles) == 0 {
		log.Println("No MKV files found, nothing to process")
		// Acknowledge the message since there are no files to process
		if err := d.Ack(false); err != nil {
			log.Printf("Error acknowledging message with no MKV files: %v", err)
		} else {
			log.Println("Message acknowledged (no MKV files to process)")
		}
		return
	}

	// Track successful file processing
	successfullyProcessed := false
	
	// Process each MKV file
	for _, file := range mkvFiles {
		log.Printf("Processing file: %s", file)

		// Get track information using mkvmerge
		jsonCmd := exec.Command("mkvmerge", "-J", file)
		jsonOutput, err := jsonCmd.Output()
		if err != nil {
			log.Printf("Error getting track info for %s: %v", file, err)
			continue
		}

		// Parse JSON output
		var trackInfo struct {
			Tracks []struct {
				ID         int    `json:"id"`
				Type       string `json:"type"`
				Properties struct {
					Language string `json:"language"`
				} `json:"properties"`
			} `json:"tracks"`
		}

		if err := json.Unmarshal(jsonOutput, &trackInfo); err != nil {
			log.Printf("Error parsing track info JSON for %s: %v", file, err)
			continue
		}

		// Check if file already has only English audio tracks
		hasNonEnglishAudio := false
		for _, track := range trackInfo.Tracks {
			if track.Type == "audio" && track.Properties.Language != "eng" {
				hasNonEnglishAudio = true
				break
			}
		}

		if !hasNonEnglishAudio {
			log.Printf("File %s already has only English audio tracks, skipping", file)
			continue
		}

		// Build track selection arguments
		var videoTracks, audioTracks, subtitleTracks []string

		for _, track := range trackInfo.Tracks {
			id := fmt.Sprintf("%d", track.ID)

			if track.Type == "video" {
				videoTracks = append(videoTracks, id)
			} else if track.Type == "audio" && track.Properties.Language == "eng" {
				audioTracks = append(audioTracks, id)
			} else if track.Type == "subtitles" && track.Properties.Language == "eng" {
				subtitleTracks = append(subtitleTracks, id)
			}
		}

		// If no English audio tracks found, keep all audio tracks
		if len(audioTracks) == 0 {
			log.Printf("No English audio tracks found in %s, keeping all audio tracks", file)
			audioTracks = nil
			for _, track := range trackInfo.Tracks {
				if track.Type == "audio" {
					audioTracks = append(audioTracks, fmt.Sprintf("%d", track.ID))
				}
			}
		}

		// Prepare output filename
		dir := filepath.Dir(file)
		basename := filepath.Base(file)
		tmpFile := filepath.Join(dir, "."+basename+".tmp.mkv")

		// Build mkvmerge command
		args := []string{"-o", tmpFile}

		if len(videoTracks) > 0 {
			args = append(args, "--video-tracks", join(videoTracks, ","))
		}

		if len(audioTracks) > 0 {
			args = append(args, "--audio-tracks", join(audioTracks, ","))
		}

		if len(subtitleTracks) > 0 {
			args = append(args, "--subtitle-tracks", join(subtitleTracks, ","))
		}

		args = append(args, file)

		// Run mkvmerge
		log.Printf("Running mkvmerge with args: %v", args)
		cmd := exec.Command("mkvmerge", args...)
		output, err := cmd.CombinedOutput()

		if err != nil {
			// mkvmerge returns 1 for warnings, but the file is still usable
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				log.Printf("mkvmerge completed with warnings for %s: %s", file, string(output))
				// Continue with the file move despite warnings
			} else {
				log.Printf("Error running mkvmerge for %s: %v", file, err)
				log.Printf("Output: %s", string(output))
				// Clean up temporary file
				os.Remove(tmpFile)
				continue
			}
		} else {
			log.Printf("mkvmerge completed successfully for %s", file)
		}

		// Replace original file with new file
		if err := os.Rename(tmpFile, file); err != nil {
			log.Printf("Error replacing original file %s: %v", file, err)
			os.Remove(tmpFile) // Clean up in case of error
		} else {
			log.Printf("Successfully processed %s", file)
			// Publish a message to the done queue
			if err := publishDoneMessage(ch, file); err != nil {
				log.Printf("Error publishing done message for %s: %v", file, err)
			} else {
				// Mark that at least one file was successfully processed
				successfullyProcessed = true
			}
		}
	}
	
	// If at least one file was processed successfully, acknowledge the original message
	if successfullyProcessed {
		if err := d.Ack(false); err != nil {
			log.Printf("Error acknowledging message: %v", err)
		} else {
			log.Println("Message acknowledged after successful processing")
		}
	} else {
		log.Println("No files were successfully processed, message remains in queue")
	}

	log.Println("Message processing completed")
}

// join joins string slice elements with a separator
func join(elements []string, separator string) string {
	if len(elements) == 0 {
		return ""
	}

	result := elements[0]
	for _, element := range elements[1:] {
		result += separator + element
	}

	return result
}