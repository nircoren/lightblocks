package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/google/uuid"
)

func sendBatch(svc *sqs.SQS, queueURL string, commands []Message, userName string) error {
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
			MessageGroupId:         aws.String(userName),
			MessageDeduplicationId: aws.String(msgID),
		}
	}

	// Send the batch
	_, err := svc.SendMessageBatch(&sqs.SendMessageBatchInput{
		QueueUrl: aws.String(queueURL),
		Entries:  entries,
	})

	// Send the result to the channel
	if err != nil {
		return err
	}

	return nil

}

func SendMessages(messages []Message, userName string) error {
	queueURL := os.Getenv("QUEUE_URL")

	svc, err := NewSQSClient()
	if err != nil {
		fmt.Println("Error creating session: ", err)
		return err
	}

	// Client side filter questions by unknown command
	filteredMessages := []Message{}
	for _, msg := range messages {
		if _, ok := AllowedCommandsMap[msg.Command]; ok {
			filteredMessages = append(filteredMessages, msg)
		} else {
			log.Printf("Unknown command: %s\n", msg.Command)
		}
	}

	// Split the messages into batches of 10
	for i := 0; i < len(filteredMessages); i += 10 {
		end := i + 10
		if end > len(filteredMessages) {
			end = len(filteredMessages)
		}
		batch := filteredMessages[i:end]

		// We don't send goroutine to maintain the order of the messages of user.
		sendBatch(svc, queueURL, batch, userName)
	}
	return nil
}
