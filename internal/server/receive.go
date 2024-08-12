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

// We have a map of channels that are used to send messages to workers
// Each worker is responsible for processing messages of a specific group (user)
func ReceiveMessages(queueProvider *MessagingService, orderedMap *OrderedMap, logger *log.Logger, deleteMessages bool) error {
	groupChanMap := make(map[string]chan models.Command)
	channelCloser := make(chan string)
	var workersWg sync.WaitGroup

	// Start a goroutine for listening for channel closure requests
	workersWg.Add(1)
	go handleChannelClosures(&workersWg, groupChanMap, channelCloser)

	const fetchWorkers = 10
	fetchChan := make(chan struct{}, fetchWorkers)

	// Start message fetching and processing
	startMessageFetcher(orderedMap, queueProvider, logger, deleteMessages, &workersWg, fetchChan, groupChanMap, channelCloser)

	// Wait for all workers to finish
	workersWg.Wait()
	cleanupResources(fetchChan, channelCloser, groupChanMap)

	return nil
}

func handleChannelClosures(workersWg *sync.WaitGroup, groupChanMap map[string]chan models.Command, channelCloser chan string) {
	defer workersWg.Done()
	for groupID := range channelCloser {
		close(groupChanMap[groupID])
		delete(groupChanMap, groupID)
	}
}

func startMessageFetcher(orderedMap *OrderedMap, queueProvider *MessagingService, logger *log.Logger, deleteMessages bool, workersWg *sync.WaitGroup, fetchChan chan struct{}, groupChanMap map[string]chan models.Command, channelCloser chan string) {
	for {
		fetchChan <- struct{}{}
		go fetchAndProcessMessages(fetchChan, orderedMap, queueProvider, logger, deleteMessages, workersWg, groupChanMap, channelCloser)
	}
}

func fetchAndProcessMessages(fetchChan chan struct{}, orderedMap *OrderedMap, queueProvider *MessagingService, logger *log.Logger, deleteMessages bool, workersWg *sync.WaitGroup, groupChanMap map[string]chan models.Command, channelCloser chan string) {
	defer func() { <-fetchChan }()

	messages, err := queueProvider.actions.ReceiveMessages()
	if err != nil {
		logger.Printf("Error receiving messages: %v", err)
		return
	}

	for _, message := range messages {
		sendMessageToWorker(message, groupChanMap, orderedMap, queueProvider, logger, deleteMessages, workersWg, channelCloser)
		fmt.Println("Sending message to channel: ", message)
	}
}

func sendMessageToWorker(message models.Command, groupChanMap map[string]chan models.Command, orderedMap *OrderedMap, queueProvider *MessagingService, logger *log.Logger, deleteMessages bool, workersWg *sync.WaitGroup, channelCloser chan string) {
	if _, exists := groupChanMap[message.GroupID]; !exists {
		groupChanMap[message.GroupID] = make(chan models.Command)
		workersWg.Add(1)
		go initWorker(orderedMap, queueProvider, message.GroupID, groupChanMap[message.GroupID], logger, deleteMessages, channelCloser, workersWg)
	}
	groupChanMap[message.GroupID] <- message
}

func initWorker(orderedMap *OrderedMap, queueProvider *MessagingService, workerID string, messageChan <-chan models.Command, logger *log.Logger, deleteMessages bool, channelCloser chan string, workersWg *sync.WaitGroup) {
	defer workersWg.Done()

	timeoutDuration := 5 * time.Second

	for {
		select {
		case message, ok := <-messageChan:
			if !ok {
				return
			}
			processSingleMessage(orderedMap, message, logger, deleteMessages, queueProvider, workersWg)
		case <-time.After(timeoutDuration):
			channelCloser <- workerID
			fmt.Println("No messages received for 5 seconds")
			return
		}
	}
}

func processSingleMessage(orderedMap *OrderedMap, message models.Command, logger *log.Logger, deleteMessages bool, queueProvider *MessagingService, workersWg *sync.WaitGroup) {
	orderedMap.HandleCommand(message, logger, workersWg)

	if deleteMessages {
		workersWg.Add(1)
		go func(receiptHandle *string) {
			defer workersWg.Done()
			err := queueProvider.actions.DeleteMessage(receiptHandle)
			if err != nil {
				logger.Printf("Failed to delete message: %v", err)
			}
		}(message.ReceiptHandle)
	}
}

func cleanupResources(fetchChan chan struct{}, channelCloser chan string, groupChanMap map[string]chan models.Command) {
	close(fetchChan)
	close(channelCloser)
	for _, ch := range groupChanMap {
		close(ch)
	}
}
