package server

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/nircoren/lightblocks/internal/server/util"
)

func TestOrdered_map(t *testing.T) {

	logger, err := util.SetupLogger("logs/sqs_messages.log")
	if err != nil {
		t.Fatalf("Failed to set up logger: %v", err)
	}
	OrderMap := NewOrderedMap()
	for i := 0; i < 10; i++ {
		OrderMap.AddItem(fmt.Sprintf("key%d", i), strconv.Itoa(i))
	}

	done := make(chan string)
	OrderMap.GetItem("key1", logger, done)
	val := <-done
	if val == "" {
		t.Fatalf("Error With Adding or getting items.")
	} else if val != "1" {
		t.Fatalf("Value not saved properly.")
	}

	OrderMap.DeleteItem("key9")
	OrderMap.GetItem("key9", logger, done)
	val = <-done
	if val == "" {
		t.Fatalf("Deleting keys doesn't work.")
	}

	itemsChan := make(chan []Pair)
	OrderMap.GetAllItems(logger, itemsChan)
	items := <-itemsChan
	for i, item := range items {
		if item.Value != strconv.Itoa(i) {
			println(items)
			println(strconv.Itoa(i), item.Value)
			t.Fatalf("Map is not ordered.")
		}
	}
}
