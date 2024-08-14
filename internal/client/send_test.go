package client

import (
	"os"
	"sync"
	"testing"

	"github.com/nircoren/lightblocks/pkg/sqs"
	"github.com/nircoren/lightblocks/queue/models"
)

func TestSendMessages(t *testing.T) {

	Messages := []models.Command{
		{
			Action: "addItem",
			Key:    "key1",
			Value:  "test",
		},
		{
			Action: "addItem",
			Key:    "key2",
			Value:  "test",
		},
		{
			Action: "addItem",
			Key:    "key3",
			Value:  "test",
		},
		{
			Action: "addItem",
			Key:    "key4",
			Value:  "test",
		},
		{
			Action: "getAllItems",
		},
	}

	config := map[string]string{
		"region":                os.Getenv("AWS_REGION"),
		"aws_access_key_id":     os.Getenv("AWS_ACCESS_KEY_ID"),
		"aws_secret_access_key": os.Getenv("AWS_SECRET_ACCESS_KEY"),
		"queueURL":              os.Getenv("QUEUE_URL"),
	}
	SQSService, err := sqs.New(config)
	if err != nil {
		t.Fatalf("Error creating SQS service: %s", err)
		return
	}

	queueProviderSend := NewMessagingService(SQSService)
	var wg sync.WaitGroup
	wg.Add(1)
	users := [2]string{"test1", "test2"}
	for i := 0; i < 2; i++ {
		go func(t *testing.T) {
			defer wg.Done()

			err = SendMessages(queueProviderSend, Messages, users[i])
			if err != nil {
				t.Errorf("Error sending messages: %s", err)
			}
		}(t)
	}

	wg.Wait()
}
