package messaging

import (
	"main/server"
	"testing"
)

func TestReceiveMessages(t *testing.T) {
	OrderMap := server.NewOrderedMap()
	logger, _ := server.SetupLogger()
	svc, err := NewSQSClient()
	err = receiveAndProcessMessages(svc, OrderMap, logger)
	if err != nil {
		t.Fatalf("Error receiving messages: %s", err)
	}
}
