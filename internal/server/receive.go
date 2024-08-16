package server

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nircoren/lightblocks/queue/models"
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

func ReceiveMessages(queueProvider *MessagingService, orderedMap *OrderedMap, logger *log.Logger) error {
	groupChanMap := make(map[string]chan models.Command)
	channelCloser := make(chan string)
	var workersWg sync.WaitGroup

	// Start a goroutine for listening for channel closure requests
	workersWg.Add(1)
	go handleChannelClosures(&workersWg, groupChanMap, channelCloser)

	const fetchWorkers = 10
	fetchChan := make(chan struct{}, fetchWorkers)

	// Start fetch workers
	for i := 0; i < fetchWorkers; i++ {
		workersWg.Add(1)
		go fetchWorker(fetchChan, queueProvider, orderedMap, groupChanMap, logger, &workersWg, channelCloser)
	}

	// Continuously feed fetch requests
	for {
		fetchChan <- struct{}{}
	}

	workersWg.Wait()
	cleanupResources(fetchChan, channelCloser, groupChanMap)

	return nil
}

func fetchWorker(fetchChan chan struct{}, queueProvider *MessagingService, orderedMap *OrderedMap, groupChanMap map[string]chan models.Command, logger *log.Logger, workersWg *sync.WaitGroup, channelCloser chan string) {
	defer workersWg.Done()

	for range fetchChan {
		messages, err := queueProvider.actions.ReceiveMessages()
		if err != nil {
			logger.Printf("Error receiving messages: %v", err)
			continue
		}
		processMessages(messages, orderedMap, queueProvider, logger, workersWg, groupChanMap, channelCloser)
	}
}

func handleChannelClosures(workersWg *sync.WaitGroup, groupChanMap map[string]chan models.Command, channelCloser chan string) {
	defer workersWg.Done()
	for groupID := range channelCloser {
		close(groupChanMap[groupID])
		delete(groupChanMap, groupID)
	}
}

func processMessages(messages []models.Command, orderedMap *OrderedMap, queueProvider *MessagingService, logger *log.Logger, workersWg *sync.WaitGroup, groupChanMap map[string]chan models.Command, channelCloser chan string) {
	for _, message := range messages {
		if _, exists := groupChanMap[message.GroupID]; !exists {
			groupChanMap[message.GroupID] = make(chan models.Command)
			workersWg.Add(1)
			go initWorker(orderedMap, queueProvider, message.GroupID, groupChanMap[message.GroupID], logger, channelCloser, workersWg)
		}

		groupChanMap[message.GroupID] <- message
	}
}

func initWorker(orderedMap *OrderedMap, queueProvider *MessagingService, workerID string, messageChan <-chan models.Command, logger *log.Logger, channelCloser chan string, workersWg *sync.WaitGroup) {
	defer workersWg.Done()

	timeoutDuration := 5 * time.Second

	for {
		select {
		case message, ok := <-messageChan:
			if !ok {
				continue
			}
			processSingleMessage(orderedMap, message, logger, queueProvider, workersWg)
		case <-time.After(timeoutDuration):
			channelCloser <- workerID
			fmt.Println("No messages received for 5 seconds")
			return
		}
	}
}

func processSingleMessage(orderedMap *OrderedMap, message models.Command, logger *log.Logger, queueProvider *MessagingService, workersWg *sync.WaitGroup) {
	orderedMap.HandleCommand(message, logger, workersWg)

	workersWg.Add(1)
	go func(receiptHandle *string) {
		defer workersWg.Done()
		err := queueProvider.actions.DeleteMessage(receiptHandle)
		if err != nil {
			logger.Printf("Failed to delete message: %v", err)
		}
	}(message.ReceiptHandle)
}

func cleanupResources(fetchChan chan struct{}, channelCloser chan string, groupChanMap map[string]chan models.Command) {
	close(fetchChan)
	close(channelCloser)
	for _, ch := range groupChanMap {
		close(ch)
	}
}
