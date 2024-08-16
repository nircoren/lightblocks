package server

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/nircoren/lightblocks/pkg/sqs"
	"github.com/nircoren/lightblocks/util"
)

// Test if service reaches provider without errors.
// I assume this test is not on production
func TestReceive(t *testing.T) {
	OrderMap := NewOrderedMap()
	logger, _ := util.SetupLogger("logs/test_sqs_messages.log")

	// Run ReceiveMessages for 5 seconds
	err := receiveMessagesWithTimeout(OrderMap, logger, 5*time.Second, t)
	if err != nil {
		t.Fatalf("Error receiving messages: %s", err)
	}
}

// This is a wrapper function for testing
func receiveMessagesWithTimeout(orderMap *OrderedMap, logger *log.Logger, timeout time.Duration, t *testing.T) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	errChan := make(chan error, 1)

	SQSService, err := sqs.New()
	if err != nil {
		t.Fatalf("Error creating SQS service: %s", err)
		return err
	}
	queueProvider := NewMessagingService(SQSService)

	go func() {
		errChan <- ReceiveMessages(queueProvider, orderMap, logger)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		// Timeout occurred
		return nil
	}
}
