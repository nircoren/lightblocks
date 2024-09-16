package server

import (
	"container/list"
	"fmt"
	"log"
)

// The logic is having a map store key that is pointing to a doubly linked list element.
// This way we can have O(1) delete, insert and update operations.
type OrderedMap struct {
	data  map[string]*list.Element
	order *list.List
	// Lock keys when writing, rlock when reading
	KeyedMutex *KeyedMutex
}

type Pair struct {
	Key   string
	Value string
}

func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		data:       make(map[string]*list.Element),
		order:      list.New(),
		KeyedMutex: NewKeyedMutex(),
	}
}

func (om *OrderedMap) AddItem(key string, value string) {
	unlock := om.KeyedMutex.Lock(key)
	element, exists := om.data[key]
	if exists {
		element.Value.(*Pair).Value = value
	} else {
		pair := &Pair{key, value}
		element := om.order.PushBack(pair)
		om.data[key] = element
	}

	fmt.Printf("Added: %s -> %s\n", key, value)
	unlock()
}

func (om *OrderedMap) DeleteItem(key string) {
	unlock := om.KeyedMutex.Lock(key)
	element, exists := om.data[key]
	if exists {
		om.order.Remove(element)
		delete(om.data, key)
	}
	unlock()
}

func (om *OrderedMap) GetItem(key string, logger *log.Logger, done chan string) {
	unlock := om.KeyedMutex.RLock(key)
	go func() {
		element, exists := om.data[key]
		var returnValue string
		if exists {
			logger.Printf("Get Item: %s -> %s\n", key, element.Value.(*Pair).Value)
			fmt.Printf("Get Item: %s -> %s\n", key, element.Value.(*Pair).Value)
			returnValue = element.Value.(*Pair).Value
		} else {
			logger.Printf("Item: %s not found\n", key)
			fmt.Printf("Item: %s not found\n", key)
			returnValue = ""
		}
		unlock()
		done <- returnValue
	}()
}

func (om *OrderedMap) GetAllItems(logger *log.Logger, done chan []Pair) {
	om.KeyedMutex.getAllItemsWg.Add(1)
	go func() {
		items := make([]Pair, 0, om.order.Len())
		for e := om.order.Front(); e != nil; e = e.Next() {
			pair := e.Value.(*Pair)
			items = append(items, *pair)
		}

		fmt.Printf("All items:\n")
		logger.Printf("All items:\n")

		for _, item := range items {
			fmt.Printf("	%s -> %s\n", item.Key, item.Value)
			logger.Printf("	%s -> %s\n", item.Key, item.Value)
		}
		om.KeyedMutex.getAllItemsWg.Done()
		done <- items
	}()
}
