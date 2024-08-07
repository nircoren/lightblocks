package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"main/models"
	"main/server"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type MessageReceiver interface {
	ReceiveMessages() ([]models.Message, error)
}

type SQSReceiver struct {
	Client *sqs.SQS
}

// The main function reads messages from the SQS queue and sends them to a channel.
// The channel is read by worker goroutines that process the messages.
// The main function waits for all workers to finish processing the messages before exiting.
func ReceiveMessages(orderedMap *server.OrderedMap, logger *log.Logger, deleteMessages bool) error {
	numWorkers := 3

	// Channel to send messages to workers
	messageChan := make(chan sqs.Message, numWorkers)
	queueURL := os.Getenv("QUEUE_URL")

	// WaitGroup to wait for all workers to finish
	var workersWg sync.WaitGroup

	svc, err := NewSQSClient()
	if err != nil {
		fmt.Println("Error creating session: ", err)
		return err
	}

	for i := 0; i < numWorkers; i++ {
		workersWg.Add(1)
		go func(workerID int) {
			defer workersWg.Done()
			if deleteMessages {
				processMessages(workerID, svc, queueURL, messageChan)
			}
		}(i)
	}

	for {
		msgResult, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			AttributeNames: []*string{
				aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
			},
			MessageAttributeNames: []*string{
				aws.String(sqs.QueueAttributeNameAll),
			},
			QueueUrl:            aws.String(os.Getenv("QUEUE_URL")),
			MaxNumberOfMessages: aws.Int64(10),
			WaitTimeSeconds:     aws.Int64(10),
		})

		if err != nil {
			log.Printf("failed to fetch sqs message: %v", err)
			continue
		}

		if len(msgResult.Messages) == 0 {
			log.Println("No messages received")
			continue
		}

		for _, message := range msgResult.Messages {
			messageModel := &models.Message{}
			err := json.Unmarshal([]byte(*message.Body), messageModel)

			if err != nil {
				// We don't stop processing the rest of the messages if one fails.
				log.Printf("Error unmarshalling message body: %v", err)
				return err
			}

			if _, allowed := AllowedCommandsMap[messageModel.Command]; !allowed {
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
func processMessages(workerID int, svc *sqs.SQS, queueURL string, messageChan <-chan sqs.Message) {
	for message := range messageChan {
		err := deleteMessage(svc, &message)
		if err != nil {
			log.Printf("failed to delete message: %v", err)
			continue
		}

	}
}

func deleteMessage(svc *sqs.SQS, msg *sqs.Message) error {
	queueURL := os.Getenv("QUEUE_URL")

	_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &queueURL,
		ReceiptHandle: msg.ReceiptHandle,
	})

	if err != nil {
		log.Printf("Failed to delete message: %v", err)
		return err
	}
	return nil

}
