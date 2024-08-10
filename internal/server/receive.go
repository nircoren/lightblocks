package server

import (
	"fmt"
	"log"
	"main/pkg/queue"
	"sync"
	"time"
)

type MessageReceiver interface {
	ReceiveMessages() ([]queue.Message, error)
}

// We have a map of channels that are used to send messages to workers
// Each worker is responsible for processing messages of a specific group (user)
func ReceiveMessages(orderedMap *OrderedMap, logger *log.Logger, deleteMessages bool) error {
	sqsClient := &queue.SQSService{}
	_, err := sqsClient.NewSQSClient()
	if err != nil {
		fmt.Println("Error creating session: ", err)
		return err
	}

	groupChanMap := make(map[string]chan queue.Message)
	channelCloser := make(chan string)
	var workersWg sync.WaitGroup

	// Start a goroutine for listening for channel closure requests
	workersWg.Add(1)
	go handleChannelClosures(&workersWg, groupChanMap, channelCloser)

	const fetchWorkers = 10
	fetchChan := make(chan struct{}, fetchWorkers)

	// Start message fetching and processing
	startMessageFetcher(orderedMap, sqsClient, logger, deleteMessages, &workersWg, fetchChan, groupChanMap, channelCloser)

	// Wait for all workers to finish
	workersWg.Wait()
	cleanupResources(fetchChan, channelCloser, groupChanMap)

	return nil
}

func handleChannelClosures(workersWg *sync.WaitGroup, groupChanMap map[string]chan queue.Message, channelCloser chan string) {
	defer workersWg.Done()
	for groupID := range channelCloser {
		close(groupChanMap[groupID])
		delete(groupChanMap, groupID)
	}
}

func startMessageFetcher(orderedMap *OrderedMap, sqsClient *queue.SQSService, logger *log.Logger, deleteMessages bool, workersWg *sync.WaitGroup, fetchChan chan struct{}, groupChanMap map[string]chan queue.Message, channelCloser chan string) {
	for {
		fetchChan <- struct{}{}
		go fetchAndProcessMessages(fetchChan, orderedMap, sqsClient, logger, deleteMessages, workersWg, groupChanMap, channelCloser)
	}
}

func fetchAndProcessMessages(fetchChan chan struct{}, orderedMap *OrderedMap, sqsClient *queue.SQSService, logger *log.Logger, deleteMessages bool, workersWg *sync.WaitGroup, groupChanMap map[string]chan queue.Message, channelCloser chan string) {
	defer func() { <-fetchChan }()

	messages, err := sqsClient.ReceiveMessages()
	if err != nil {
		logger.Printf("Error receiving messages: %v", err)
		return
	}

	for _, message := range messages {
		sendMessageToWorker(message, groupChanMap, orderedMap, sqsClient, logger, deleteMessages, workersWg, channelCloser)
		fmt.Println("Sending message to channel: ", message)
	}
}

func sendMessageToWorker(message queue.Message, groupChanMap map[string]chan queue.Message, orderedMap *OrderedMap, sqsClient *queue.SQSService, logger *log.Logger, deleteMessages bool, workersWg *sync.WaitGroup, channelCloser chan string) {
	if _, exists := groupChanMap[message.GroupID]; !exists {
		groupChanMap[message.GroupID] = make(chan queue.Message)
		workersWg.Add(1)
		go processMessages(orderedMap, sqsClient, message.GroupID, groupChanMap[message.GroupID], logger, deleteMessages, channelCloser, workersWg)
	}
	groupChanMap[message.GroupID] <- message
}

func processMessages(orderedMap *OrderedMap, sqsClient *queue.SQSService, workerID string, messageChan <-chan queue.Message, logger *log.Logger, deleteMessages bool, channelCloser chan string, workersWg *sync.WaitGroup) {
	defer workersWg.Done()

	timeoutDuration := 5 * time.Second

	for {
		select {
		case message, ok := <-messageChan:
			if !ok {
				return
			}
			processSingleMessage(orderedMap, message, logger, deleteMessages, sqsClient, workersWg)
		case <-time.After(timeoutDuration):
			channelCloser <- workerID
			fmt.Println("No messages received for 5 seconds")
		}
	}
}

func processSingleMessage(orderedMap *OrderedMap, message queue.Message, logger *log.Logger, deleteMessages bool, sqsClient *queue.SQSService, workersWg *sync.WaitGroup) {
	orderedMap.HandleCommand(message, logger, workersWg)

	if deleteMessages {
		workersWg.Add(1)
		go func(receiptHandle *string) {
			defer workersWg.Done()
			err := sqsClient.DeleteMessage(receiptHandle)
			if err != nil {
				logger.Printf("Failed to delete message: %v", err)
			}
		}(message.ReceiptHandle)
	}
}

func cleanupResources(fetchChan chan struct{}, channelCloser chan string, groupChanMap map[string]chan queue.Message) {
	close(fetchChan)
	close(channelCloser)
	for _, ch := range groupChanMap {
		close(ch)
	}
}
