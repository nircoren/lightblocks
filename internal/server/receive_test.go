package server

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/nircoren/lightblocks/pkg/sqs"
	"github.com/nircoren/lightblocks/util"
)

// This is a wrapper function for testing
func receiveMessagesWithTimeout(orderMap *OrderedMap, logger *log.Logger, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	errChan := make(chan error, 1)

	config := map[string]string{
		"region":                os.Getenv("AWS_REGION"),
		"aws_access_key_id":     os.Getenv("AWS_ACCESS_KEY_ID"),
		"aws_secret_access_key": os.Getenv("AWS_SECRET_ACCESS_KEY"),
		"queueURL":              os.Getenv("QUEUE_URL"),
	}

	SQSService, err := sqs.NewMessagingService(config)
	if err != nil {
		t.Fatalf("Error creating SQS service: %s", err)
		return err
	}

	queueProvider := NewMessagingService(SQSService)
	if err != nil {
		t.Fatalf("Error creating session: %s", err)
		return err
	}

	go func() {
		// Don't want to delete messages in the test as i can't control the messages that will return
		errChan <- ReceiveMessages(queueProvider, orderMap, logger, false)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		// Timeout occurred
		return nil
	}
}

func TestReceive(t *testing.T) {
	OrderMap := NewOrderedMap()
	logger, _ := util.SetupLogger("logs/test_sqs_messages.log")

	// Run ReceiveMessages for 5 seconds
	err := receiveMessagesWithTimeout(OrderMap, logger, 5*time.Second)
	if err != nil {
		t.Fatalf("Error receiving messages: %s", err)
	}
}
