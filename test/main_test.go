package main

import (
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/nircoren/lightblocks/internal/client"
	"github.com/nircoren/lightblocks/internal/server"
	"github.com/nircoren/lightblocks/internal/server/util"
	"github.com/nircoren/lightblocks/pkg/queue/models"
	"github.com/nircoren/lightblocks/pkg/queue/sqs"
)

// Test that messages are sent and received correctly
// Need to add a test for read, as its async
// Test on empty queue that is not production
func TestMain(t *testing.T) {

	messages := &[]models.CommandBase{
		{Action: "addItem", Key: "1", Value: "ה1"},
		{Action: "addItem", Key: "2", Value: "v2"},
		{Action: "addItem", Key: "3", Value: "v3"},
		{Action: "deleteItem", Key: "1"},
		{Action: "addItem", Key: "4", Value: "v4"},
	}

	expected := map[string]string{
		"1": "",
		"2": "v2",
		"3": "v3",
		"4": "val4",
	}

	// Init sending messages
	config, err := godotenv.Read()
	if err != nil {
		t.Fatalf("Error reading .env file: %s", err)
	}
	// Dependency Injection of SQS
	SQSService, err := sqs.New(config)
	if err != nil {
		t.Fatalf("Error creating SQS service: %s", err)
		return
	}

	queueProviderSend := client.NewSendMessagesService(SQSService)

	err = client.SendMessages(queueProviderSend, *messages, "test")
	if err != nil {
		t.Fatalf("Error sending messages: %s", err)
	}

	// Init receiving messages
	logger, err := util.SetupLogger("logs/sqs_messages.log")
	if err != nil {
		t.Fatalf("Failed to set up logger: %v", err)
	}

	queueProviderReceive := server.NewReceiveMessagesService(SQSService)

	orderedMap := server.NewOrderedMap()
	server.ReceiveMessages(queueProviderReceive, orderedMap, logger)

	time.Sleep(5 * time.Second)
	done := make(chan []server.Pair)
	orderedMap.GetAllItems(logger, done)
	allItems := <-done
	for _, item := range allItems {
		if item.Value != expected[item.Key] {
			t.Fatalf("Expected: %s, got: %s", expected[item.Key], item.Value)
		}
	}

}
