package main

import (
	"fmt"
	messaging "main/message"
	"main/server"
)

func main() {
	orderedMap := server.NewOrderedMap()
	logger, err := server.SetupLogger()
	if err != nil {
		fmt.Printf("Failed to set up logger: %v\n", err)
		return
	}

	// numGoroutines := 5
	// waitgroup...
	// Launch multiple goroutines to receive messages (unorderd)
	// for i := 0; i < numGoroutines; i++ {
	// 	go messaging.ReceiveMessages(orderedMap, logger)
	// }
	messaging.ReceiveMessages(orderedMap, logger)
}
