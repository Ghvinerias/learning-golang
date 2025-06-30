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
	rabbitmqHost     = "10.10.40.19"
	rabbitmqPort     = "5672"
	rabbitmqUsername = "mkvmerge-consumer"
	rabbitmqPassword = "mkvmerge-consumer"
	rabbitmqVhost    = "media-automation"
	queueName        = "mkvmerge.tasks"     // Queue to consume from
	doneQueueName    = "mkvmerge.done"      // Queue to publish to when done
	dlqQueueName     = "mkvmerge.tasks_DLQ" // Dead Letter Queue for failed messages
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
		osExit(1)
	}
}

// ensureQueueExists creates a queue if it does not already exist
func ensureQueueExists(ch *amqp.Channel, qName string) amqp.Queue {
	q, err := ch.QueueDeclare(
		qName, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, fmt.Sprintf("Failed to declare queue '%s'", qName))
	log.Printf("Queue '%s' declared", qName)
	return q
}

// ensureMainQueueWithDLX creates the main processing queue with Dead Letter Exchange configuration
func ensureMainQueueWithDLX(ch *amqp.Channel) amqp.Queue {
	// First, declare the DLX exchange
	err := ch.ExchangeDeclare(
		"dlx",    // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare DLX exchange")

	// Try to declare the main queue with DLX configuration
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    "dlx",
			"x-dead-letter-routing-key": dlqQueueName,
		}, // arguments
	)

	// If we get a precondition failed error, try to delete and recreate the queue
	if err != nil && isInequivalentArgError(err) {
		log.Printf("Queue '%s' exists with different configuration, attempting to delete and recreate...", queueName)

		// Try to delete the existing queue (this will fail if it has messages)
		_, deleteErr := ch.QueueDelete(queueName, false, false, false)
		if deleteErr != nil {
			log.Printf("Warning: Could not delete existing queue '%s': %v", queueName, deleteErr)
			log.Printf("You may need to manually delete the queue '%s' from RabbitMQ management interface", queueName)
			failOnError(err, fmt.Sprintf("Failed to declare main queue '%s' with DLX (queue exists with different config)", queueName))
		}

		// Try to declare again after deletion
		q, err = ch.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			amqp.Table{
				"x-dead-letter-exchange":    "dlx",
				"x-dead-letter-routing-key": dlqQueueName,
			}, // arguments
		)
		if err != nil {
			failOnError(err, fmt.Sprintf("Failed to declare main queue '%s' with DLX after deletion", queueName))
		}
		log.Printf("Successfully recreated queue '%s' with DLX configuration", queueName)
	} else if err != nil {
		failOnError(err, fmt.Sprintf("Failed to declare main queue '%s' with DLX", queueName))
	} else {
		log.Printf("Main queue '%s' declared with DLX configuration", queueName)
	}

	// Ensure DLQ exists and bind to the DLX exchange
	_ = ensureQueueExists(ch, dlqQueueName)

	err = ch.QueueBind(
		dlqQueueName, // queue name
		dlqQueueName, // routing key
		"dlx",        // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind DLQ to DLX")
	log.Printf("DLQ '%s' bound to DLX exchange", dlqQueueName)

	return q
}

// isInequivalentArgError checks if the error is due to inequivalent arguments
func isInequivalentArgError(err error) bool {
	if amqpErr, ok := err.(*amqp.Error); ok {
		return amqpErr.Code == 406 &&
			(amqpErr.Reason == "PRECONDITION_FAILED" ||
				amqpErr.Reason != "" &&
					(amqpErr.Reason[:len("PRECONDITION_FAILED")] == "PRECONDITION_FAILED" ||
						amqpErr.Reason[:len("inequivalent arg")] == "inequivalent arg"))
	}
	return false
}

// publishDoneMessage publishes a message to the done queue
func publishDoneMessage(ch ChannelInterface, filename string) error {
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
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // make message persistent
		})

	if err != nil {
		return fmt.Errorf("failed to publish done message: %v", err)
	}

	log.Printf("Published completion message for file %s to %s queue", filename, doneQueueName)
	return nil
}

// publishToDLQ publishes a message to the Dead Letter Queue with an error reason
func publishToDLQ(ch *amqp.Channel, body []byte, reason string) error {
	// Create a wrapper message with the original message and error reason
	dlqMessage := map[string]interface{}{
		"originalMessage": string(body),
		"errorReason":     reason,
		"timestamp":       time.Now().Format(time.RFC3339),
	}

	// Convert to JSON
	dlqBody, err := json.Marshal(dlqMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal DLQ message: %v", err)
	}

	// Ensure DLQ exists
	_ = ensureQueueExists(ch, dlqQueueName)

	// Publish to the DLQ
	err = ch.Publish(
		"",           // exchange
		dlqQueueName, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         dlqBody,
			DeliveryMode: amqp.Persistent, // make message persistent
		})

	if err != nil {
		return fmt.Errorf("failed to publish to DLQ: %v", err)
	}

	log.Printf("Published message to DLQ %s with reason: %s", dlqQueueName, reason)
	return nil
}

// rejectMessageToDLQ rejects a message and routes it to the DLQ automatically
func rejectMessageToDLQ(d amqp.Delivery, reason string) error {
	log.Printf("Rejecting message to DLQ with reason: %s", reason)

	// Reject the message with requeue=false, which will send it to DLX
	err := d.Reject(false)
	if err != nil {
		return fmt.Errorf("failed to reject message: %v", err)
	}

	log.Printf("Message rejected and routed to DLQ: %s", reason)
	return nil
}

// ChannelInterface defines the interface for RabbitMQ channel operations needed by our application
type ChannelInterface interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
	QueueDelete(name string, ifUnused, ifEmpty, noWait bool) (int, error)
	QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error
	Qos(prefetchCount, prefetchSize int, global bool) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Close() error
}

// Ensure amqp.Channel implements our interface at compile time
var _ ChannelInterface = (*amqp.Channel)(nil)

// Helper variable to make failOnError testable
var osExit = os.Exit

func main() {
	// Set up logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("Starting RabbitMQ consumer...")

	// Connect to RabbitMQ
	connectionString := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		rabbitmqUsername, rabbitmqPassword, rabbitmqHost, rabbitmqPort, rabbitmqVhost)
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
	processingQueue := ensureMainQueueWithDLX(ch) // Use DLX-configured queue
	_ = ensureQueueExists(ch, doneQueueName)      // Ensure done queue exists
	_ = ensureQueueExists(ch, dlqQueueName)       // Ensure dead letter queue exists

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
		// Reject message to DLQ for parsing errors
		if err := rejectMessageToDLQ(d, fmt.Sprintf("JSON parsing error: %v", err)); err != nil {
			log.Printf("Error rejecting message to DLQ: %v", err)
		}
		return
	}

	// Map category to base directory path
	basePath, exists := CategoryPathMap[msg.Category]
	if !exists {
		log.Printf("Unknown category: %s", msg.Category)
		// Reject message to DLQ for unknown category
		reason := fmt.Sprintf("Unknown category: %s", msg.Category)
		if err := rejectMessageToDLQ(d, reason); err != nil {
			log.Printf("Error rejecting message to DLQ: %v", err)
		}
		return
	}

	// Construct full folder path
	folderPath := filepath.Join(basePath, msg.TorrentName)
	log.Printf("Looking for MKV files in: %s", folderPath)

	// Check if the folder exists
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		log.Printf("Folder does not exist: %s", folderPath)
		// Reject message to DLQ for non-existent folder
		reason := fmt.Sprintf("Folder does not exist: %s", folderPath)
		if err := rejectMessageToDLQ(d, reason); err != nil {
			log.Printf("Error rejecting message to DLQ: %v", err)
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
