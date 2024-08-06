package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"main/server"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func ReceiveMessages(orderedMap *server.OrderedMap, logger *log.Logger) error {
	svc, err := NewSQSClient()
	if err != nil {
		fmt.Println("Error creating session: ", err)
		return err
	}

	// Loop endlessly to receive messages
	for {
		err = receiveAndProcessMessages(svc, orderedMap, logger)
		if err != nil {
			log.Printf("Error receiving messages: %v", err)
			return err
		}
	}
}

func receiveAndProcessMessages(svc *sqs.SQS, orderedMap *server.OrderedMap, logger *log.Logger) error {
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
		log.Printf("Failed to receive messages: %v", err)
		return err
	}

	if len(msgResult.Messages) == 0 {
		log.Printf("No messages received.")
	}

	var wg sync.WaitGroup
	for _, msg := range msgResult.Messages {
		processMessage(svc, msg, orderedMap, logger, &wg)
	}

	wg.Wait()
	return nil
}

func processMessage(svc *sqs.SQS, msg *sqs.Message, orderedMap *server.OrderedMap, logger *log.Logger, wg *sync.WaitGroup) {
	defer deleteMessage(svc, msg)

	message := &Message{}
	err := json.Unmarshal([]byte(*msg.Body), message)

	if err != nil {
		// We don't stop processing the rest of the messages if one fails.
		log.Printf("Error unmarshalling message body: %v", err)
		return
	}

	if _, allowed := AllowedCommandsMap[message.Command]; !allowed {
		log.Printf("invalid command: %s", message.Command)
		return
	}

	handleCommand(message, orderedMap, logger, wg)
}

func handleCommand(message *Message, orderedMap *server.OrderedMap, logger *log.Logger, wg *sync.WaitGroup) {
	// We use goroutines only on read, to prevent exec order issues.
	switch message.Command {
	case "addItem":
		orderedMap.AddItem(message.Key, message.Value)
		fmt.Printf("Added: %s -> %s\n", message.Key, message.Value)
	case "deleteItem":
		orderedMap.DeleteItem(message.Key)
		fmt.Printf("Deleted: %s\n", message.Key)
	case "getItem":
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, exists := orderedMap.GetItem(message.Key)
			if exists {
				logger.Printf("Get Item: %s -> %s\n", message.Key, value)
				fmt.Printf("Get Item: %s -> %s\n", message.Key, value)
			} else {
				fmt.Printf("Item: %s not found\n", message.Key)
			}
		}()
	case "getAllItems":
		wg.Add(1)
		go func() {
			defer wg.Done()
			items := orderedMap.GetAllItems()
			logger.Printf("All items:")
			for _, item := range items {
				fmt.Printf("%s -> %s\n", item.Key, item.Value)
				logger.Printf("	Got Item: %s -> %s\n", item.Key, item.Value)
			}
		}()
	default:
		fmt.Printf("Unknown command: %s\n", message.Command)
	}
}

func deleteMessage(svc *sqs.SQS, msg *sqs.Message) {
	queueURL := os.Getenv("QUEUE_URL")
	if queueURL == "" {
		log.Println("QUEUE_URL environment variable is not set.")
		return
	}

	_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &queueURL,
		ReceiptHandle: msg.ReceiptHandle,
	})

	if err != nil {
		log.Printf("Failed to delete message: %v", err)
	}
}
