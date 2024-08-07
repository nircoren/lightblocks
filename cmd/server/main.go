package main

import (
	"fmt"
	"main/messaging"
	"main/server"
)

func main() {
	orderedMap := server.NewOrderedMap()
	logger, err := server.SetupLogger()

	if err != nil {
		fmt.Printf("Failed to set up logger: %v\n", err)
		return
	}

	messaging.ReceiveMessages(orderedMap, logger, true)

}
