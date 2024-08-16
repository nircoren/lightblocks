package sqs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/nircoren/lightblocks/queue/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/google/uuid"
)

type SQSService struct {
	client   *sqs.SQS
	queueURL string
}

// TODO: Add the necessary structs and methods to implement the send and receive messages
// Should include the Command struct in it, with sqs additions.

// Inits connection to SQS
func New() (*SQSService, error) {

	config := map[string]string{
		"region":                os.Getenv("AWS_REGION"),
		"aws_access_key_id":     os.Getenv("AWS_ACCESS_KEY_ID"),
		"aws_secret_access_key": os.Getenv("AWS_SECRET_ACCESS_KEY"),
		"queueURL":              os.Getenv("QUEUE_URL"),
	}
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(config["region"]),
		Credentials: credentials.NewStaticCredentials(config["aws_access_key_id"], config["aws_secret_access_key"], ""),
	})

	if err != nil {
		log.Fatalf("failed to create session, %v", err)
		return nil, err
	}

	return &SQSService{
		queueURL: config["queueURL"],
		client:   sqs.New(sess),
	}, nil
}

func (s *SQSService) SendMessages(messages []models.Command, userName string) error {
	const SqsMaxBatchSize int = 10

	for i := 0; i < len(messages); i += SqsMaxBatchSize {
		end := i + SqsMaxBatchSize
		if end > len(messages) {
			end = len(messages)
		}
		batch := messages[i:end]

		// We don't use goroutine to maintain the order of the messages of user.
		err := s.sendBatch(batch, userName)
		if err != nil {
			log.Printf("Error sending batch: %v\n", err)
			return err
		}
	}
	return nil
}

func (s *SQSService) ReceiveMessages() ([]models.Command, error) {
	msgResult, err := s.client.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            aws.String(s.queueURL),
		MaxNumberOfMessages: aws.Int64(10),
		WaitTimeSeconds:     aws.Int64(10),
	})

	if err != nil {
		log.Printf("failed to fetch sqs message: %v", err)
		return nil, err
	}

	if len(msgResult.Messages) == 0 {
		log.Println("No messages received")
		return nil, nil
	}

	// Convert the messages to our Message struct
	messages := make([]models.Command, len(msgResult.Messages))
	for i, msg := range msgResult.Messages {
		messageModel, err := formatMessageModel(msg)
		if err != nil {
			s.DeleteMessage(msg.ReceiptHandle)
			continue
		}
		messages[i] = *messageModel
	}
	return messages, nil
}

func (s *SQSService) DeleteMessage(receiptHandle *string) error {
	_, err := s.client.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &s.queueURL,
		ReceiptHandle: receiptHandle,
	})

	if err != nil {
		log.Printf("Failed to delete message: %v", err)
		return err
	}
	return nil

}

func (s *SQSService) sendBatch(Commands []models.Command, userName string) error {
	entries := make([]*sqs.SendMessageBatchRequestEntry, len(Commands))
	for i, Command := range Commands {
		// Convert command struct to JSON string for the message body
		actionBody, err := json.Marshal(Command)
		if err != nil {
			log.Printf("Error marshalling command: %v\n", err)
		}
		msgID := uuid.New().String()
		entries[i] = &sqs.SendMessageBatchRequestEntry{
			Id:          aws.String(msgID),
			MessageBody: aws.String(string(actionBody)),
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"CommandType": {
					DataType:    aws.String("String"),
					StringValue: aws.String(Command.Action),
				},
			},
			MessageGroupId:         aws.String(userName),
			MessageDeduplicationId: aws.String(msgID),
		}
	}

	// Send the batch
	_, err := s.client.SendMessageBatch(&sqs.SendMessageBatchInput{
		QueueUrl: aws.String(s.queueURL),
		Entries:  entries,
	})

	if err != nil {
		return err
	}

	return nil
}

func formatMessageModel(message *sqs.Message) (*models.Command, error) {
	messageModel := &models.Command{}

	// Unmarshal the message body into the Message struct
	err := json.Unmarshal([]byte(*message.Body), messageModel)
	if err != nil {
		log.Printf("Error unmarshalling message body: %v", err)
		return nil, err
	}

	// Validate the command
	if err := messageModel.Validate(); err != nil {
		log.Printf("Invalid command: %v\n", err)
		return nil, fmt.Errorf("invalid command: %v", err)
	}
	messageModel.GroupID = *message.Attributes["MessageGroupId"]
	messageModel.ReceiptHandle = message.ReceiptHandle

	return messageModel, nil
}
