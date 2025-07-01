package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/streadway/amqp"
)

// RabbitMQ connection details (you may want to load these from environment variables)
const (
	rabbitmqHost     = "10.10.40.19"
	rabbitmqPort     = "5672"
	rabbitmqUsername = "mkvmerge-notifier"
	rabbitmqPassword = "mkvmerge-notifier"
	rabbitmqVhost    = "media-automation"
	queueName        = "mkvmerge.done"     // Queue to consume from
	dlqQueueName     = "mkvmerge.done_DLQ" // Dead Letter Queue for failed messages
)

// Telegram bot configuration (you should load these from environment variables)
const (
	telegramBotToken = ""   // Replace with your actual bot token
	telegramChatID   = 1234 // Replace with your actual chat ID
)

// Message represents the structure of incoming RabbitMQ messages from mkvmerge.done queue
type Message struct {
	Filename string `json:"filename"`
	Status   string `json:"status"`
	Time     string `json:"time"`
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
func ensureMainQueueWithDLX(conn *amqp.Connection) (amqp.Queue, *amqp.Channel) {
	// Create a new channel for queue operations
	ch, err := conn.Channel()
	failOnError(err, "Failed to open channel for queue operations")

	// First, declare the DLX exchange
	err = ch.ExchangeDeclare(
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

	// If we get a precondition failed error, the queue exists with different configuration
	if err != nil && isInequivalentArgError(err) {
		log.Printf("Queue '%s' exists with different configuration", queueName)

		// Close the current channel and create a new one since RabbitMQ may have closed it
		ch.Close()
		ch, err = conn.Channel()
		failOnError(err, "Failed to reopen channel after queue configuration mismatch")

		// Try to use the existing queue by declaring it passively (just check if it exists)
		log.Printf("Attempting to use existing queue '%s' as-is...", queueName)
		q, err = ch.QueueDeclarePassive(
			queueName, // name
			true,      // durable (we'll accept whatever it is)
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)

		if err != nil {
			log.Printf("Failed to access existing queue '%s': %v", queueName, err)
			log.Printf("Warning: Queue '%s' exists but cannot be accessed with current configuration", queueName)
			log.Printf("Consider manually deleting the queue '%s' from RabbitMQ management interface if you want DLX functionality", queueName)
			failOnError(err, fmt.Sprintf("Cannot access existing queue '%s'", queueName))
		} else {
			log.Printf("Successfully using existing queue '%s' (DLX functionality may not be available)", queueName)
		}
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
	if err != nil {
		log.Printf("Warning: Failed to bind DLQ to DLX: %v (DLX functionality may not work)", err)
	} else {
		log.Printf("DLQ '%s' bound to DLX exchange", dlqQueueName)
	}

	return q, ch
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
	QueueDeclarePassive(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
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

// TelegramBotInterface defines the interface for Telegram bot operations
type TelegramBotInterface interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

// Ensure tgbotapi.BotAPI implements our interface at compile time
var _ TelegramBotInterface = (*tgbotapi.BotAPI)(nil)

// Helper variable to make failOnError testable
var osExit = os.Exit

// sendTelegramNotification sends a notification message via Telegram bot
func sendTelegramNotification(bot TelegramBotInterface, message string) error {
	msg := tgbotapi.NewMessage(telegramChatID, message)
	msg.ParseMode = "Markdown"

	_, err := bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %v", err)
	}

	log.Printf("Successfully sent Telegram notification: %s", message)
	return nil
}

// formatNotificationMessage formats the notification message for Telegram
func formatNotificationMessage(msg Message) string {
	filename := filepath.Base(msg.Filename)

	// Parse the time to make it more readable
	parsedTime, err := time.Parse(time.RFC3339, msg.Time)
	if err != nil {
		log.Printf("Warning: Could not parse time %s: %v", msg.Time, err)
		parsedTime = time.Now()
	}

	formattedTime := parsedTime.Format("2006-01-02 15:04:05")

	return fmt.Sprintf(
		"🎬 *MKV Processing Complete*\n\n"+
			"📁 *File:* `%s`\n"+
			"✅ *Status:* %s\n"+
			"🕒 *Completed:* %s\n"+
			"📂 *Full Path:* `%s`",
		filename,
		msg.Status,
		formattedTime,
		msg.Filename,
	)
}

func main() {
	// Set up logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("Starting MKV Notifier RabbitMQ consumer...")

	// Initialize Telegram bot
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		failOnError(err, "Failed to initialize Telegram bot")
	}
	log.Printf("Telegram bot initialized: @%s", bot.Self.UserName)

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

	// Declare queues (ensures they exist) - this may create a new channel if needed
	processingQueue, mainCh := ensureMainQueueWithDLX(conn) // Use DLX-configured queue
	defer mainCh.Close()
	_ = ensureQueueExists(mainCh, dlqQueueName) // Ensure dead letter queue exists

	// Set QoS (prefetch count)
	err = mainCh.Qos(
		1,     // prefetch count (process one message at a time)
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	// Register consumer
	msgs, err := mainCh.Consume(
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

			// Process the message and acknowledge only after successful notification
			processMessage(mainCh, bot, d, d.Body)
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

	log.Println("MKV Notifier is now running. Press CTRL+C to exit")
	<-forever
	log.Println("MKV Notifier shutdown complete")
}

// processMessage handles the received message by sending a Telegram notification
func processMessage(ch *amqp.Channel, bot TelegramBotInterface, d amqp.Delivery, body []byte) {
	log.Printf("Processing notification message: %s", body)

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

	// Validate required fields
	if msg.Filename == "" {
		log.Printf("Message missing required field 'filename'")
		if err := rejectMessageToDLQ(d, "Missing required field: filename"); err != nil {
			log.Printf("Error rejecting message to DLQ: %v", err)
		}
		return
	}

	// Format and send the Telegram notification
	notificationText := formatNotificationMessage(msg)

	// Attempt to send the notification
	if err := sendTelegramNotification(bot, notificationText); err != nil {
		log.Printf("Failed to send Telegram notification: %v", err)

		// Move message to DLQ since notification failed
		if err := publishToDLQ(ch, body, fmt.Sprintf("Failed to send Telegram notification: %v", err)); err != nil {
			log.Printf("Error publishing to DLQ: %v", err)
		}

		// Acknowledge the original message to remove it from the main queue
		if err := d.Ack(false); err != nil {
			log.Printf("Error acknowledging failed message: %v", err)
		} else {
			log.Println("Message acknowledged after moving to DLQ due to notification failure")
		}
		return
	}

	// If notification was sent successfully, acknowledge the message
	if err := d.Ack(false); err != nil {
		log.Printf("Error acknowledging message after successful notification: %v", err)
	} else {
		log.Println("Message acknowledged after successful Telegram notification")
	}

	log.Println("Notification message processing completed")
}
