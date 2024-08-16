package client

import (
	"sync"
	"testing"

	"github.com/nircoren/lightblocks/pkg/sqs"
	"github.com/nircoren/lightblocks/queue/models"
)

func TestSendMessages(t *testing.T) {

	messages := &[]models.Command{
		{CommandBase: models.CommandBase{Action: "addItem", Key: "key1", Value: "test"}},
		{CommandBase: models.CommandBase{Action: "addItem", Key: "key2", Value: "test"}},
		{CommandBase: models.CommandBase{Action: "addItem", Key: "key3", Value: "test"}},
		{CommandBase: models.CommandBase{Action: "addItem", Key: "key4", Value: "test"}},
		{CommandBase: models.CommandBase{Action: "getAllItems"}},
	}

	SQSService, err := sqs.New()
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
