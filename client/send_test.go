package client

import (
	"sync"
	"testing"

	"github.com/joho/godotenv"
	"github.com/nircoren/lightblocks/pkg/sqs"
	"github.com/nircoren/lightblocks/queue/models"
)

func TestSendMessages(t *testing.T) {

	messages := &[]models.CommandBase{
		{Action: "addItem", Key: "key1", Value: "test"},
		{Action: "addItem", Key: "key2", Value: "test"},
		{Action: "addItem", Key: "key3", Value: "test"},
		{Action: "addItem", Key: "key4", Value: "test"},
		{Action: "getAllItems"},
	}

	config, err := godotenv.Read()
	if err != nil {
		t.Errorf("Error reading .env file: %s", err)
		return
	}
	// Dependency Injection of SQS
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

			err = SendMessages(queueProviderSend, *messages, users[i])
			if err != nil {
				t.Errorf("Error sending messages: %s", err)
			}
		}(t)
	}

	wg.Wait()
}
