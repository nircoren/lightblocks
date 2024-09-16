package sqs

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nircoren/lightblocks/pkg/queue/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/google/uuid"
)

var SqsMaxBatchSize = 10

type SQSService struct {
	client   *sqs.SQS
	queueURL string
}

// Inits connection to SQS
func New(config map[string]string) (*SQSService, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(config["AWS_REGION"]),
		Credentials: credentials.NewStaticCredentials(config["AWS_ACCESS_KEY_ID"], config["AWS_SECRET_ACCESS_KEY"], ""),
	})

	if err != nil {
		log.Fatalf("failed to create session, %v", err)
		return nil, err
	}

	return &SQSService{
		queueURL: config["QUEUE_URL"],
		client:   sqs.New(sess),
	}, nil
}

func (s *SQSService) SendMessages(messages []models.CommandBase, userName string) error {
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
		MaxNumberOfMessages: aws.Int64(int64(SqsMaxBatchSize)),
		WaitTimeSeconds:     aws.Int64(10),
	})

	if err != nil {
		log.Printf("failed to fetch sqs message: %v", err)
		return nil, err
	}

	if len(msgResult.Messages) == 0 {
		return nil, nil
	}

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

func (s *SQSService) sendBatch(Commands []models.CommandBase, userName string) error {
	entries := make([]*sqs.SendMessageBatchRequestEntry, len(Commands))
	for i, Command := range Commands {
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

	err := json.Unmarshal([]byte(*message.Body), messageModel)
	if err != nil {
		log.Printf("Error unmarshalling message body: %v", err)
		return nil, err
	}

	err = messageModel.Validate()
	if err != nil {
		log.Printf("Invalid command: %v\n", err)
		return nil, fmt.Errorf("invalid command: %v", err)
	}
	messageModel.GroupID = *message.Attributes["MessageGroupId"]
	messageModel.ReceiptHandle = message.ReceiptHandle

	return messageModel, nil
}
