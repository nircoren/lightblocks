package main

import (
	"os"
	"testing"
	"time"

	"github.com/nircoren/lightblocks/internal/client"
	"github.com/nircoren/lightblocks/internal/server"
	"github.com/nircoren/lightblocks/pkg/sqs"
	"github.com/nircoren/lightblocks/queue/models"
	"github.com/nircoren/lightblocks/util"
)

// Test that messages are sent and received correctly
// Need to add a test for read, as its async
// Test on empty queue that is not production
func TestMain(t *testing.T) {

	messages := &[]models.Command{
		{CommandBase: models.CommandBase{Action: "addItem", Key: "1", Value: "v1"}},
		{CommandBase: models.CommandBase{Action: "addItem", Key: "2", Value: "v2"}},
		{CommandBase: models.CommandBase{Action: "addItem", Key: "3", Value: "v3"}},
		{CommandBase: models.CommandBase{Action: "deleteItem", Key: "1"}},
		{CommandBase: models.CommandBase{Action: "addItem", Key: "4", Value: "v4"}},
	}

	expected := map[string]string{
		"1": "",
		"2": "v2",
		"3": "v3",
		"4": "val4",
	}

	// expected := {}interface{}{

	config := map[string]string{
		"region":                os.Getenv("AWS_REGION"),
		"aws_access_key_id":     os.Getenv("AWS_ACCESS_KEY_ID"),
		"aws_secret_access_key": os.Getenv("AWS_SECRET_ACCESS_KEY"),
		"queueURL":              os.Getenv("QUEUE_URL"),
	}

	// Init sending messages
	SQSService, err := sqs.New(config)
	if err != nil {
		t.Fatalf("Error creating SQS service: %s", err)
		return
	}

	queueProviderSend := client.NewMessagingService(SQSService)

	err = client.SendMessages(queueProviderSend, *messages, "test")
	if err != nil {
		t.Fatalf("Error sending messages: %s", err)
	}

	// Init receiving messages
	logger, err := util.SetupLogger("logs/sqs_messages.log")
	if err != nil {
		t.Fatalf("Failed to set up logger: %v", err)
	}

	queueProviderReceive := server.NewMessagingService(SQSService)

	orderedMap := server.NewOrderedMap()
	err = server.ReceiveMessages(queueProviderReceive, orderedMap, logger)
	if err != nil {
		t.Fatalf("Error receiving messages: %s", err)
	}

	time.Sleep(5 * time.Second)
	allItems := orderedMap.GetAllItems()

	for _, item := range allItems {
		if item.Value != expected[item.Key] {
			t.Fatalf("Expected: %s, got: %s", expected[item.Key], item.Value)
		}
	}

}
