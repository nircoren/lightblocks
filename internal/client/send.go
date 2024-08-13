package client

import (
	"log"

	"github.com/nircoren/lightblocks/queue/models"
)

const SqsMaxBatchSize int = 10

type sendActions interface {
	SendMessages(messages []models.Command, userName string) error
}

type MessagingService struct {
	actions sendActions
}

func NewMessagingService(a sendActions) *MessagingService {
	return &MessagingService{actions: a}
}

func SendMessages(queueProvider *MessagingService, messages []models.Command, userName string) error {
	// Client side filter questions with unknown command
	filteredMessages := []models.Command{}
	for _, msg := range messages {
		if _, ok := models.AllowedActionsMap[msg.Action]; ok {
			filteredMessages = append(filteredMessages, msg)
		} else {
			log.Printf("Unknown command: %s\n", msg.Action)
		}
	}

	// Split the messages into batches of 10 as its the max for sqs

	// We don't use goroutine to maintain the order of the messages of user.
	err := queueProvider.actions.SendMessages(filteredMessages, userName)
	if err != nil {
		log.Printf("Error sending batch: %v\n", err)
		return err
	}
	return nil
}
