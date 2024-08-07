package server

import (
	"encoding/json"
	"fmt"
	"log"
	"main/pkg/queue"
	"sync"

	"github.com/aws/aws-sdk-go/service/sqs"
)

type MessageReceiver interface {
	ReceiveMessages() ([]queue.Message, error)
}

// The main function reads messages from the SQS queue and sends them to a channel.
// The channel is read by worker goroutines that process the messages.
// The main function waits for all workers to finish processing the messages before exiting.
func ReceiveMessages(orderedMap *OrderedMap, logger *log.Logger, deleteMessages bool) error {
	numWorkers := 3

	// Channel to send messages to workers
	messageChan := make(chan sqs.Message, numWorkers)

	// WaitGroup to wait for all workers to finish
	var workersWg sync.WaitGroup

	SqsClient := &queue.SQSService{}
	_, err := SqsClient.NewSQSClient()
	if err != nil {
		fmt.Println("Error creating session: ", err)
		return err
	}

	for i := 0; i < numWorkers; i++ {
		workersWg.Add(1)
		go func(workerID int) {
			defer workersWg.Done()
			if deleteMessages {
				processMessages(SqsClient, workerID, messageChan)
			}
		}(i)
	}

	for {
		msgResult, err := SqsClient.ReceiveMessages()

		if err != nil {
			log.Printf("Error receiving messages: %v", err)
			continue
		}

		if msgResult == nil {
			continue
		}

		for _, message := range msgResult.Messages {
			messageChan <- *message
		}

		for _, message := range msgResult.Messages {
			messageModel := &queue.Message{}
			err := json.Unmarshal([]byte(*message.Body), messageModel)

			if err != nil {
				// We don't stop processing the rest of the messages if one fails.
				log.Printf("Error unmarshalling message body: %v", err)
				return err
			}

			if _, allowed := queue.AllowedCommandsMap[messageModel.Command]; !allowed {
				log.Printf("invalid command: %s", messageModel.Command)
				return err
			}
			orderedMap.HandleCommand(messageModel, logger, &workersWg)
			messageChan <- *message
		}

	}
	workersWg.Wait()

	return nil
}

// The processMessages function reads messages from the messageChan channel and processes them.
func processMessages(SqsClient *queue.SQSService, workerID int, messageChan <-chan sqs.Message) {
	for message := range messageChan {
		err := SqsClient.DeleteMessage(&message)
		if err != nil {
			log.Printf("failed to delete message: %v", err)
			continue
		}

	}
}
