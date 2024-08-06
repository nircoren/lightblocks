package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"sync"

	messaging "main/message"
)

func main() {
	username := flag.String("username", "guest", "")
	msgs := flag.String("msgs", "", "")
	flag.Parse()
	// Loop to prevent program from exiting
	for {
		if *username != "" && *msgs != "" {
			var Messages []messaging.Message
			err := json.Unmarshal([]byte(*msgs), &Messages)
			if err != nil {
				fmt.Println("Error unmarshaling JSON:", err)
				return
			}
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				err = messaging.SendMessages(Messages, *username)
				if err != nil {
					fmt.Println("Error sending messages:", err)
				}
			}()

			wg.Wait()
			return

		}
	}
}
