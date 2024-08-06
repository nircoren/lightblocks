package messaging

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

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
func NewSQSClient() (*sqs.SQS, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	})

	if err != nil {
		log.Fatalf("failed to create session, %v", err)
		return nil, err
	}

	return sqs.New(sess), nil

}
