package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/nircoren/lightblocks/internal/client"
	"github.com/nircoren/lightblocks/pkg/queue/models"
	"github.com/nircoren/lightblocks/pkg/queue/sqs"
)

func main() {
	config, err := loadEnv()
	if err != nil {
		log.Println("Error loading environment:", err)
		return
	}

	queueProvider, err := initializeSQSService(config)
	if err != nil {
		log.Println("Error initializing SQS service:", err)
		return
	}

	username, err := promptUsername()
	if err != nil {
		log.Println("Error reading username:", err)
		return
	}

	runMessageLoop(queueProvider, username)
}

func loadEnv() (map[string]string, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file:", err)
	}
	config, err := godotenv.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading environment variables: %w", err)
	}
	return config, nil
}

func initializeSQSService(config map[string]string) (*client.SendMessagesService, error) {
	SQSService, err := sqs.New(config)
	if err != nil {
		return nil, fmt.Errorf("error creating SQS service: %w", err)
	}
	queueProvider := client.NewSendMessagesService(SQSService)
	return queueProvider, nil
}

func promptUsername() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	username = strings.TrimSpace(username)
	if username == "" {
		return "", fmt.Errorf("username cannot be empty")
	}
	return username, nil
}

func runMessageLoop(queueProvider *client.SendMessagesService, username string) {
	reader := bufio.NewReader(os.Stdin)

	for {
		rawMessages, err := promptMessages(reader)
		if err != nil {
			log.Println("Error reading messages:", err)
			continue
		}

		err = sendMessages(queueProvider, rawMessages, username)
		if err != nil {
			log.Println("Error sending messages:", err)
		} else {
			fmt.Println("Messages sent successfully.")
		}

		if !askToContinue(reader) {
			break
		}
	}

	fmt.Println("Exiting...")
}

func promptMessages(reader *bufio.Reader) (string, error) {
	fmt.Println("Enter messages (in JSON format), followed by an empty line to finish:")
	var rawMessagesBuilder strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		rawMessagesBuilder.WriteString(line)
	}

	rawMessages := rawMessagesBuilder.String()
	if rawMessages == "" {
		return "", fmt.Errorf("no messages provided")
	}

	return rawMessages, nil
}

func sendMessages(queueProvider *client.SendMessagesService, rawMessages string, username string) error {
	var messages []models.CommandBase
	err := json.Unmarshal([]byte(rawMessages), &messages)
	if err != nil {
		return fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	err = client.SendMessages(queueProvider, messages, username)
	if err != nil {
		return fmt.Errorf("error sending messages: %w", err)
	}

	return nil
}

func askToContinue(reader *bufio.Reader) bool {
	fmt.Print("Do you want to send more messages? (y/n): ")
	cont, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading input:", err)
		return false
	}
	cont = strings.TrimSpace(cont)
	return strings.ToLower(cont) == "y"
}
