package messaging

import (
	"testing"

	"sync"
)

func TestSendMessages(t *testing.T) {

	Messages := []Message{
		{
			Command: "addItem",
			Key:     "key1",
			Value:   "value1",
		},
		{
			Command: "addItem",
			Key:     "key2",
			Value:   "value2",
		},
		{
			Command: "addItem",
			Key:     "key3",
			Value:   "value3",
		},
		{
			Command: "addItem",
			Key:     "key4",
			Value:   "value4",
		},
		{
			Command: "getAllItems",
		},
	}

	var wg sync.WaitGroup
	var err error
	wg.Add(1)
	users := [2]string{"guest1", "guest2"}
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
