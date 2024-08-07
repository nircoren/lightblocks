package messaging

import (
	"context"
	"log"
	"main/server"
	"main/util"
	"testing"
	"time"
)

// This is a wrapper function for testing
func receiveMessagesWithTimeout(orderMap *server.OrderedMap, logger *log.Logger, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	errChan := make(chan error, 1)

	go func() {
		// Don't want to delete messages in the test as i can't control the messages that will return
		errChan <- ReceiveMessages(orderMap, logger, false)
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
	OrderMap := server.NewOrderedMap()
	logger, _ := util.SetupLogger("logs/test_sqs_messages.log")

	// Run ReceiveMessages for 5 seconds
	err := receiveMessagesWithTimeout(OrderMap, logger, 5*time.Second)
	if err != nil {
		t.Fatalf("Error receiving messages: %s", err)
	}
}
