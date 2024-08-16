package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/nircoren/lightblocks/client"
	"github.com/nircoren/lightblocks/pkg/sqs"
	"github.com/nircoren/lightblocks/queue/models"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file:", err)
	}

	config := map[string]string{
		"region":                os.Getenv("AWS_REGION"),
		"aws_access_key_id":     os.Getenv("AWS_ACCESS_KEY_ID"),
		"aws_secret_access_key": os.Getenv("AWS_SECRET_ACCESS_KEY"),
		"queueURL":              os.Getenv("QUEUE_URL"),
	}
	// Dependency Injection of SQS
	SQSService, err := sqs.New(config)
	if err != nil {
		log.Println("Error creating SQS service:", err)
		return
	}

	queueProvider := client.NewMessagingService(SQSService)

	// Read input from the command line
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Enter messages (in JSON format): ")
	rawMessages, _ := reader.ReadString('\n')
	rawMessages = strings.TrimSpace(rawMessages)

	// If no username or messages provided, exit
	if username == "" || rawMessages == "" {
		fmt.Println("Username or messages cannot be empty.")
		return
	}

	// Parse messages
	var messages []models.Command
	err = json.Unmarshal([]byte(rawMessages), &messages)
	if err != nil {
		log.Println("Error unmarshaling JSON:", err)
		return
	}

	err = client.SendMessages(queueProvider, messages, username)
	if err != nil {
		log.Println("Error sending messages:", err)
	} else {
		fmt.Println("Messages sent successfully.")
	}

}
