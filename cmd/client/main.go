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
	"github.com/nircoren/lightblocks/pkg/queue/models"
	"github.com/nircoren/lightblocks/pkg/queue/sqs"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file:", err)
	}

	// Read all the environment variables.
	config, err := godotenv.Read()
	if err != nil {
		log.Println("Error reading .env file:", err)
		return
	}

	// Dependency Injection of SQS
	SQSService, err := sqs.New(config)
	if err != nil {
		log.Println("Error creating SQS service:", err)
		return
	}

	queueProvider := client.NewMessagingService(SQSService)

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading username:", err)
		return
	}
	username = strings.TrimSpace(username)

	if username == "" {
		fmt.Println("Username cannot be empty.")
		return
	}

	// Prompt user to enter messages in a loop
	for {
		fmt.Println("Enter messages (in JSON format), followed by an empty line to finish:")

		// Accept multiple lines of input until an empty line is entered
		var rawMessagesBuilder strings.Builder
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				log.Println("Error reading input:", err)
				return
			}
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}
			rawMessagesBuilder.WriteString(line)
		}

		rawMessages := rawMessagesBuilder.String()

		// If no messages provided, prompt again or allow to exit
		if rawMessages == "" {
			fmt.Println("No messages provided.")
			continue
		}

		// Parse messages
		var messages []models.CommandBase
		err = json.Unmarshal([]byte(rawMessages), &messages)
		if err != nil {
			log.Println("Error unmarshaling JSON:", err)
			continue
		}

		err = client.SendMessages(queueProvider, messages, username)
		if err != nil {
			log.Println("Error sending messages:", err)
		} else {
			fmt.Println("Messages sent successfully.")
		}

		fmt.Print("Do you want to send more messages? (y/n): ")
		cont, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading input:", err)
			return
		}
		cont = strings.TrimSpace(cont)
		if strings.ToLower(cont) != "y" {
			break
		}
	}
	fmt.Println("Exiting...")
}
