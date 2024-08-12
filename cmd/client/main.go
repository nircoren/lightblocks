package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/nircoren/lightblocks/internal/client"
	"github.com/nircoren/lightblocks/pkg/sqs"
	"github.com/nircoren/lightblocks/queue/models"
)

func main() {
	username := flag.String("username", "guest", "")
	rawMessages := flag.String("msgs", "", "")
	flag.Parse()
	if *username != "" && *rawMessages != "" {
		config := map[string]string{
			"region":                os.Getenv("AWS_REGION"),
			"aws_access_key_id":     os.Getenv("AWS_ACCESS_KEY_ID"),
			"aws_secret_access_key": os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"queueURL":              os.Getenv("QUEUE_URL"),
		}
		// Create a new SQS client
		SQSService, err := sqs.NewMessagingService(config)
		if err != nil {
			fmt.Println("Error creating SQS service: ", err)
			return
		}

		queueProvider := client.NewMessagingService(SQSService)
		if err != nil {
			fmt.Println("Error creating session: ", err)
			return
		}

		var messages []models.Command
		err = json.Unmarshal([]byte(*rawMessages), &messages)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return
		}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := client.SendMessages(queueProvider, messages, *username)
			if err != nil {
				log.Println("Error sending messages:", err)
			}
		}()
		wg.Wait()
		return

	}
}
