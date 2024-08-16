package client

import (
	"fmt"
	"log"

	"github.com/nircoren/lightblocks/queue/models"
)

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
	fmt.Printf("Sending messages: %v\n", messages)
	filteredMessages := []models.Command{}
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
