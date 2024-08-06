package server

import (
	"fmt"
	"strconv"
	"testing"
)

func TestOrderedMap(t *testing.T) {

	OrderMap := NewOrderedMap()
	for i := 0; i < 10; i++ {
		OrderMap.AddItem(fmt.Sprintf("key%d", i), strconv.Itoa(i))
	}

	val, exists := OrderMap.GetItem("key1")
	if !exists {
		t.Fatalf("Error With Adding or getting items.")
	} else if val != "1" {
		t.Fatalf("Value not saved properly.")
	}

	OrderMap.DeleteItem("key9")
	_, exists = OrderMap.GetItem("key9")
	if exists {
		fmt.Printf("Deleting keys doesn't work.")
	}

	items := OrderMap.GetAllItems()
	for i, item := range items {
		if item.Value != strconv.Itoa(i) {
			println(items)
			println(strconv.Itoa(i), item.Value)
			t.Fatalf("Map is not ordered.")
		}
	}
}
