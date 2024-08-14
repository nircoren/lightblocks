package main

import (
	"os"
	"testing"

	"github.com/nircoren/lightblocks/internal/client"
	"github.com/nircoren/lightblocks/internal/server"
	"github.com/nircoren/lightblocks/pkg/sqs"
	"github.com/nircoren/lightblocks/queue/models"
	"github.com/nircoren/lightblocks/util"
)

// Need to test on empty queue.
func TestMain(t *testing.T) {

	messages := []models.Command{
		{Action: "addItem", Key: "1", Value: "v1"},
		{Action: "addItem", Key: "2", Value: "v2"},
		{Action: "addItem", Key: "4", Value: "val4"},
		{Action: "getItem", Key: "1"},
		{Action: "deleteItem", Key: "1"},
		{Action: "getItem", Key: "1"},
		{Action: "getItem", Key: "2"},
		{Action: "getItem", Key: "3"},
		{Action: "addItem", Key: "3", Value: "v3"},
		{Action: "getItem", Key: "4"},
		{Action: "getItem"},
	}

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

	err = client.SendMessages(queueProviderSend, messages, "test")
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
	err = server.ReceiveMessages(queueProviderReceive, orderedMap, logger, true)
	if err != nil {
		t.Fatalf("Error receiving messages: %s", err)
	}

	// expected := []interface{key,value}{
	// 	{"key": "1", "val": ""},
	// 	{"key": "2", "val": "v2"},
	// 	{"key": "3", "val": "v3"},
	// 	{"key": "4", "val": "v4"},
	// }

	// for _, res := range expected {
	// 	orderedMap[res]
	// }

}
