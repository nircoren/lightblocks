package client

import (
	"fmt"
	"log"
	"main/pkg/queue"
)

const SqsMaxBatchSize int = 10

func SendMessages(messages []queue.Message, userName string) error {
	SqsClient := &queue.SQSService{}
	_, err := SqsClient.NewSQSClient()
	if err != nil {
		fmt.Println("Error creating session: ", err)
		return err
	}
	// Client side filter questions with unknown command
	filteredMessages := []queue.Message{}
	for _, msg := range messages {
		if _, ok := queue.AllowedCommandsMap[msg.Command]; ok {
			filteredMessages = append(filteredMessages, msg)
		} else {
			log.Printf("Unknown command: %s\n", msg.Command)
		}
	}

	// Split the messages into batches of 10 as its the max for sqs
	for i := 0; i < len(filteredMessages); i += SqsMaxBatchSize {
		end := i + SqsMaxBatchSize
		if end > len(filteredMessages) {
			end = len(filteredMessages)
		}
		batch := filteredMessages[i:end]

		// We don't send goroutine to maintain the order of the messages of user.
		err := SqsClient.SendBatch(batch, userName)
		if err != nil {
			log.Printf("Error sending batch: %v\n", err)
			return err
		}
	}
	return nil
}
