package server

import (
	"fmt"
	"log"

	"github.com/nircoren/lightblocks/pkg/queue/models"
)

type receiveActions interface {
	ReceiveMessages() ([]models.Command, error)
	DeleteMessage(receiptHandle *string) error
}

type MessagingService struct {
	actions receiveActions
}

func NewMessagingService(a receiveActions) *MessagingService {
	return &MessagingService{actions: a}
}

func ReceiveMessages(queueProvider *MessagingService, orderedMap *OrderedMap, logger *log.Logger) {

	const fetchWorkers int = 30
	const processWorkers int = 30
	fetchChan := make(chan struct{}, fetchWorkers)
	messagesChan := make(chan []models.Command, processWorkers)

	// Init fetch workers
	for i := 0; i < fetchWorkers; i++ {
		go fetchWorker(fetchChan, messagesChan, queueProvider)
	}

	// Init process workers
	for i := 0; i < processWorkers; i++ {
		go processWorker(messagesChan, queueProvider, orderedMap, logger)
	}

	// Continuously feed fetch requests
	for {
		fetchChan <- struct{}{}
	}

	// Add for gracful shut down?
	// sigit := make(chan os.Signal, 1)
	// <-sigit
	// fmt.Println("Shutting down...")
	// close(fetchChan)
	// close(messagesChan)
	// // will send a signal to all workers to stop
	// closeWorkers(fetchWorkers, processWorkers)
}

func fetchWorker(fetchChan chan struct{}, messagesChan chan []models.Command, queueProvider *MessagingService) {

	for range fetchChan {
		messages, err := queueProvider.actions.ReceiveMessages()
		if err != nil {
			fmt.Printf("Error receiving messages: %v", err)
			continue
		}
		if len(messages) == 0 {
			continue
		}
		messagesChan <- messages
	}
}

func processWorker(messagesChan chan []models.Command, queueProvider *MessagingService, orderedMap *OrderedMap, logger *log.Logger) {

	for messages := range messagesChan {
		for _, message := range messages {
			processMessage(orderedMap, message, logger, queueProvider)
		}
	}
}

func processMessage(orderedMap *OrderedMap, message models.Command, logger *log.Logger, queueProvider *MessagingService) {
	orderedMap.HandleCommand(message, logger)
	go func(receiptHandle *string) {
		err := queueProvider.actions.DeleteMessage(receiptHandle)
		if err != nil {
			logger.Printf("Failed to delete message: %v", err)
		}
	}(message.ReceiptHandle)
}
