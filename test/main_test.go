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

func TestMain(t *testing.T) {

	messages := []models.Command{
		{Action: "addItem", Key: "1", Value: "nir1"},
		{Action: "getItem", Key: "1"},
		{Action: "getItem", Key: "1"},
		{Action: "addItem", Key: "2", Value: "nir112"},
		{Action: "getItem", Key: "1"},
		{Action: "getItem", Key: "1"},
		{Action: "deleteItem", Key: "1"},
		{Action: "getItem", Key: "1"},
		{Action: "getItem", Key: "1"},
		{Action: "getItem", Key: "2"},
		{Action: "getItem", Key: "2"},
		{Action: "getItem", Key: "1"},
		{Action: "getItem", Key: "2"},
		{Action: "getItem", Key: "1"},
		{Action: "deleteItem", Key: "1"},
		{Action: "getItem", Key: "1"},
	}

	config := map[string]string{
		"region":                os.Getenv("AWS_REGION"),
		"aws_access_key_id":     os.Getenv("AWS_ACCESS_KEY_ID"),
		"aws_secret_access_key": os.Getenv("AWS_SECRET_ACCESS_KEY"),
		"queueURL":              os.Getenv("QUEUE_URL"),
	}

	SQSService, err := sqs.NewMessagingService(config)
	if err != nil {
		t.Fatalf("Error creating SQS service: %s", err)
		return
	}

	queueProvider := client.NewMessagingService(SQSService)
	if err != nil {
		t.Fatalf("Error creating session: %s", err)
		return
	}

	err = client.SendMessages(queueProvider, messages, "test")
	if err != nil {
		t.Fatalf("Error sending messages: %s", err)
	}

	logger, err := util.SetupLogger("logs/sqs_messages.log")
	if err != nil {
		t.Fatalf("Failed to set up logger: %v", err)
	}

	SQSService, err = sqs.NewMessagingService(config)
	if err != nil {
		t.Fatalf("Error creating SQS service: %s", err)
		return
	}

	queueProvider = server.NewMessagingService(SQSService)
	if err != nil {
		t.Fatalf("Error creating session: %s", err)
		return
	}

	orderedMap := server.NewOrderedMap()
	err = server.ReceiveMessages(queueProvider, orderedMap, logger, true)
	if err != nil {
		t.Fatalf("Error receiving messages: %s", err)
	}

	for idx, msg := range messages {
		expectedMsg := messages[idx]

		if msg.Action != expectedMsg.Action {
			t.Fatalf("Error: received command %s != expected command %s", msg.Action, expectedMsg.Action)
		}

		if expectedMsg.Key != "" && msg.Key != expectedMsg.Key {
			t.Fatalf("Error: received key %s != expected key %s", msg.Key, expectedMsg.Key)
		}

		if expectedMsg.Value != "" && msg.Value != expectedMsg.Value {
			t.Fatalf("Error: received value %s != expected value %s", msg.Value, expectedMsg.Value)
		}
	}

}
