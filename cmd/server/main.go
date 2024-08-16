package main

import (
	"fmt"
	"os"

	"github.com/nircoren/lightblocks/internal/server"
	"github.com/nircoren/lightblocks/pkg/sqs"
	"github.com/nircoren/lightblocks/util"
)

func main() {
	orderedMap := server.NewOrderedMap()
	logger, err := util.SetupLogger("logs/sqs_messages.log")

	if err != nil {
		fmt.Printf("Failed to set up logger: %v\n", err)
		return
	}

	config := map[string]string{
		"region":                os.Getenv("AWS_REGION"),
		"aws_access_key_id":     os.Getenv("AWS_ACCESS_KEY_ID"),
		"aws_secret_access_key": os.Getenv("AWS_SECRET_ACCESS_KEY"),
		"queueURL":              os.Getenv("QUEUE_URL"),
	}

	// Dependency Injection of sqs
	SQSService, err := sqs.New(config)
	if err != nil {
		fmt.Println("Error creating SQS service: ", err)
		return
	}

	queueProvider := server.NewMessagingService(SQSService)

	server.ReceiveMessages(queueProvider, orderedMap, logger)

}
