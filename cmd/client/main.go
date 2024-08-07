package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"main/messaging"
	"main/models"
	"sync"
)

func main() {
	username := flag.String("username", "guest", "")
	msgs := flag.String("msgs", "", "")
	flag.Parse()
	if *username != "" && *msgs != "" {
		var messages []models.Message
		err := json.Unmarshal([]byte(*msgs), &messages)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := messaging.SendMessages(messages, *username)
			if err != nil {
				log.Println("Error sending messages:", err)
			}
		}()
		wg.Wait()
		return

	}
}
