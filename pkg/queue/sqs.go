package queue

import (
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/google/uuid"
)

type SQSClient interface {
	SendMessage(input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error)
	ReceiveMessages(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error)
}

type SQSService struct {
	client   *sqs.SQS
	queueURL string
}

type Message struct {
	Command string `json:"command"`
	Key     string `json:"key,omitempty"`
	Value   string `json:"value,omitempty"`
}

var AllowedCommandsMap = map[string]bool{
	"addItem":     true,
	"deleteItem":  true,
	"getItem":     true,
	"getAllItems": true,
}

// Inits connection to SQS
func (s *SQSService) NewSQSClient() (*sqs.SQS, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	})

	if err != nil {
		log.Fatalf("failed to create session, %v", err)
		return nil, err
	}
	s.queueURL = os.Getenv("QUEUE_URL")
	s.client = sqs.New(sess)
	return s.client, nil

}

func (s *SQSService) SendBatch(commands []Message, userName string) error {
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
	_, err := s.client.SendMessageBatch(&sqs.SendMessageBatchInput{
		QueueUrl: aws.String(s.queueURL),
		Entries:  entries,
	})

	if err != nil {
		return err
	}

	return nil

}

func (s *SQSService) ReceiveMessages() (*sqs.ReceiveMessageOutput, error) {
	msgResult, err := s.client.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
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
	return msgResult, nil

}

func (s *SQSService) DeleteMessage(msg *sqs.Message) error {
	_, err := s.client.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &s.queueURL,
		ReceiptHandle: msg.ReceiptHandle,
	})

	if err != nil {
		log.Printf("Failed to delete message: %v", err)
		return err
	}
	return nil

}
