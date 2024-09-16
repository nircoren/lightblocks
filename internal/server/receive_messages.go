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

type ReceiveMessagesService struct {
	actions receiveActions
}

func NewReceiveMessagesService(a receiveActions) *ReceiveMessagesService {
	return &ReceiveMessagesService{actions: a}
}

func ReceiveMessages(queueService *ReceiveMessagesService, orderedMap *OrderedMap, logger *log.Logger) {

	const fetchWorkers int = 30
	const processWorkers int = 30
	fetchChan := make(chan struct{}, fetchWorkers)
	messagesChan := make(chan []models.Command, processWorkers)

	for i := 0; i < fetchWorkers; i++ {
		go fetchWorker(fetchChan, messagesChan, queueService)
	}

	for i := 0; i < processWorkers; i++ {
		go processWorker(messagesChan, queueService, orderedMap, logger)
	}

	// Continuously feed fetch requests
	for {
		fetchChan <- struct{}{}
	}

	// shutDown()
}

// func shutDown() {
// Add for gracful shut down?
// sigit := make(chan os.Signal, 1)
// <-sigit
// fmt.Println("Shutting down...")
// close(fetchChan)
// close(messagesChan)
// // will send a signal to all workers to stop
// closeWorkers(fetchWorkers, processWorkers)
// }

func fetchWorker(fetchChan chan struct{}, messagesChan chan []models.Command, queueService *ReceiveMessagesService) {

	for range fetchChan {
		messages, err := queueService.actions.ReceiveMessages()
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

func processWorker(messagesChan chan []models.Command, queueService *ReceiveMessagesService, orderedMap *OrderedMap, logger *log.Logger) {

	for messages := range messagesChan {
		for _, message := range messages {
			processMessage(orderedMap, message, logger, queueService)
		}
	}
}

func processMessage(om *OrderedMap, message models.Command, logger *log.Logger, queueService *ReceiveMessagesService) {
	switch message.Action {
	case "addItem":
		processAddItem(om, message, queueService)
	case "deleteItem":
		processDeleteItem(om, message, queueService)
	case "getItem":
		processGetItem(om, message, logger, queueService)
	case "getAllItems":
		processGetAllItems(om, message, logger, queueService)
	default:
		fmt.Printf("Unknown command: %s\n", message.Action)
	}
}

func processAddItem(om *OrderedMap, message models.Command, queueService *ReceiveMessagesService) {
	defer queueService.actions.DeleteMessage(message.ReceiptHandle)
	om.AddItem(message.Key, message.Value)
}

func processDeleteItem(om *OrderedMap, message models.Command, queueService *ReceiveMessagesService) {
	defer queueService.actions.DeleteMessage(message.ReceiptHandle)
	om.DeleteItem(message.Key)
}

func processGetItem(om *OrderedMap, message models.Command, logger *log.Logger, queueService *ReceiveMessagesService) {
	done := make(chan string)
	om.GetItem(message.Key, logger, done)
	go waitForCompletion(done, message, queueService)
}

func processGetAllItems(om *OrderedMap, message models.Command, logger *log.Logger, queueService *ReceiveMessagesService) {
	done := make(chan []Pair)
	om.GetAllItems(logger, done)
	go waitForCompletion(done, message, queueService)
}

func waitForCompletion[T any](done chan T, message models.Command, queueService *ReceiveMessagesService) {
	defer func() {
		queueService.actions.DeleteMessage(message.ReceiptHandle)
		close(done)
	}()
	<-done
}
