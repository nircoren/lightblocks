package main

import (
	"fmt"
	"main/messaging"
	"main/server"
	"main/util"
)

func main() {
	orderedMap := server.NewOrderedMap()
	logger, err := util.SetupLogger("logs/sqs_messages.log")

	if err != nil {
		fmt.Printf("Failed to set up logger: %v\n", err)
		return
	}

	messaging.ReceiveMessages(orderedMap, logger, true)

}
