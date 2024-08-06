package messaging

import (
	"main/server"
	"testing"
)

func TestReceive(t *testing.T) {
	OrderMap := server.NewOrderedMap()
	logger, _ := server.SetupLogger()
	err := ReceiveMessages(OrderMap, logger)
	if err != nil {
		t.Fatalf("Error receiving messages: %s", err)
	}
}
