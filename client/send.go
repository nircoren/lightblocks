package client

import (
	"log"

	"github.com/nircoren/lightblocks/queue/models"
)

type sendActions interface {
	SendMessages(messages []models.CommandBase, userName string) error
}

type MessagingService struct {
	actions sendActions
}

func NewMessagingService(a sendActions) *MessagingService {
	return &MessagingService{actions: a}
}

func SendMessages(queueProvider *MessagingService, messages []models.CommandBase, userName string) error {
	// Max capacity of the slice is the same as the length of the messages slice.
	filteredMessages := make([]models.CommandBase, 0, len(messages))
	for _, message := range messages {
		// Validate the command
		if err := message.Validate(); err != nil {
			log.Printf("Invalid command: %v\n", err)
			continue
		}
		filteredMessages = append(filteredMessages, message)
	}
	// We don't use goroutine to maintain the order of the messages of user.
	err := queueProvider.actions.SendMessages(filteredMessages, userName)
	if err != nil {
		log.Printf("Error sending batch: %v\n", err)
		return err
	}
	return nil
}
