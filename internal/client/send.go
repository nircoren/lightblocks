package client

import (
	"encoding/json"
	"fmt"
	"log"
	"main/pkg/queue"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/google/uuid"
)

const SqsMaxBatchSize int = 10

func sendBatch(svc *sqs.SQS, queueURL string, commands []queue.Message, userName string) error {
	fmt.Println(commands)
	entries := make([]*sqs.SendMessageBatchRequestEntry, len(commands))
	for i, cmd := range commands {
		// Convert command struct to JSON string for the message body
		cmdBody, err := json.Marshal(cmd)
		if err != nil {
			log.Printf("Error marshalling command: %v\n", err)
		}
		msgID := uuid.New().String()

		entries[i] = &sqs.SendMessageBatchRequestEntry{
			Id:          aws.String(msgID),
			MessageBody: aws.String(string(cmdBody)),
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"CommandType": {
					DataType:    aws.String("String"),
					StringValue: aws.String(cmd.Command),
				},
			},
			MessageGroupId:         aws.String(userName + "_" + msgID),
			MessageDeduplicationId: aws.String(msgID),
		}
	}

	// Send the batch
	_, err := svc.SendMessageBatch(&sqs.SendMessageBatchInput{
		QueueUrl: aws.String(queueURL),
		Entries:  entries,
	})

	if err != nil {
		return err
	}

	return nil

}

func SendMessages(messages []queue.Message, userName string) error {
	queueURL := os.Getenv("QUEUE_URL")

	svc, err := queue.NewSQSClient()
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
		sendBatch(svc, queueURL, batch, userName)
	}
	return nil
}
