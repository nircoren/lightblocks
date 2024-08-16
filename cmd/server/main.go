package main

import (
	"fmt"

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

	// Dependency Injection of sqs
	SQSService, err := sqs.New()
	if err != nil {
		fmt.Println("Error creating SQS service: ", err)
		return
	}

	queueProvider := server.NewMessagingService(SQSService)

	server.ReceiveMessages(queueProvider, orderedMap, logger)

}
