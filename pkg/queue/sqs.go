package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/google/uuid"
)

type SQSClient interface {
	SendMessages(commands []Message, userName string) error
	ReceiveMessages() (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(msg *sqs.Message) error
}

type SQSService struct {
	client   *sqs.SQS
	queueURL string
}

// Not used currently
type Command struct {
	Command string `json:"command"`
	Key     string `json:"key,omitempty"`
	Value   string `json:"value,omitempty"`
}

type Message struct {
	Command       string `json:"command"`
	Key           string `json:"key,omitempty"`
	Value         string `json:"value,omitempty"`
	GroupID       string
	ReceiptHandle *string
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

func (s *SQSService) SendMessages(commands []Message, userName string) error {
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
	_, err := s.client.SendMessageBatch(&sqs.SendMessageBatchInput{
		QueueUrl: aws.String(s.queueURL),
		Entries:  entries,
	})

	if err != nil {
		return err
	}

	return nil

}

func (s *SQSService) ReceiveMessages() ([]Message, error) {
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
	messages := make([]Message, len(msgResult.Messages))
	for i, msg := range msgResult.Messages {
		messageModel, err := makeMessageModel(msg)
		if err != nil {
			s.DeleteMessage(msg.ReceiptHandle)
			continue
		}
		messages[i] = *messageModel
	}
	return messages, nil

}

func makeMessageModel(message *sqs.Message) (*Message, error) {
	messageModel := &Message{}

	// Unmarshal the message body into the Message struct
	err := json.Unmarshal([]byte(*message.Body), messageModel)
	if err != nil {
		log.Printf("Error unmarshalling message body: %v", err)
		return nil, err
	}

	// Validate the command
	if _, allowed := AllowedCommandsMap[messageModel.Command]; !allowed {
		log.Printf("Invalid command: %s", messageModel.Command)
		return nil, fmt.Errorf("invalid command: %s", messageModel.Command)
	}

	if messageModel.Command == "addItem" && (messageModel.Key == "" || messageModel.Value == "") {
		log.Printf("Missing key or value for addItem command: %s", *message.Body)
		return nil, fmt.Errorf("missing key or value for addItem command")
	}

	messageModel.GroupID = *message.Attributes["MessageGroupId"]
	messageModel.ReceiptHandle = message.ReceiptHandle

	return messageModel, nil
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
