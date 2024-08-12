package client

import (
	"sync"
	"testing"

	"github.com/nircoren/lightblocks/queue/models"
)

func TestSendMessages(t *testing.T) {

	Messages := []models.Command{
		{
			Action: "addItem",
			Key:    "key1",
			Value:  "test",
		},
		{
			Action: "addItem",
			Key:    "key2",
			Value:  "test",
		},
		{
			Action: "addItem",
			Key:    "key3",
			Value:  "test",
		},
		{
			Action: "addItem",
			Key:    "key4",
			Value:  "test",
		},
		{
			Action: "getAllItems",
		},
	}

	var wg sync.WaitGroup
	var err error
	wg.Add(1)
	users := [2]string{"test1", "test2"}
	for i := 0; i < 2; i++ {
		go func(t *testing.T) {
			defer wg.Done()
			err = SendMessages(Messages, users[i])
			if err != nil {
				t.Fatalf("Error sending messages: %s", err)
			}
		}(t)
	}

	wg.Wait()
}
