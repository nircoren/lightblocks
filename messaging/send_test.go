package messaging

import (
	"main/models"
	"sync"
	"testing"
)

func TestSendMessages(t *testing.T) {

	Messages := []models.Message{
		{
			Command: "addItem",
			Key:     "key1",
			Value:   "test",
		},
		{
			Command: "addItem",
			Key:     "key2",
			Value:   "test",
		},
		{
			Command: "addItem",
			Key:     "key3",
			Value:   "test",
		},
		{
			Command: "addItem",
			Key:     "key4",
			Value:   "test",
		},
		{
			Command: "getAllItems",
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
