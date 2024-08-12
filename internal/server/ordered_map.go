package server

import (
	"container/list"
	"fmt"
	"log"
	"sync"

	"github.com/nircoren/lightblocks/queue/models"
)

// The logic is having a map store key that is pointing to a doubly linked list element.
// This way we can have O(1) delete, insert and update operations.

type OrderedMap struct {
	data  map[string]*list.Element
	order *list.List
	mutex *sync.RWMutex
}

type Pair struct {
	Key   string
	Value string
	mu    *sync.RWMutex
}

func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		data:  make(map[string]*list.Element),
		order: list.New(),
		mutex: &sync.RWMutex{},
	}
}

func (om *OrderedMap) AddItem(key string, value string) {
	element, exists := om.data[key]
	if exists {
		element.Value.(*Pair).Value = value
	} else {
		// Insert new element
		pair := &Pair{key, value, &sync.RWMutex{}}
		element := om.order.PushBack(pair)
		om.data[key] = element
	}
}

func (om *OrderedMap) DeleteItem(key string) {
	element, exists := om.data[key]
	if exists {
		om.order.Remove(element)
		delete(om.data, key)
	} else {
		fmt.Printf("Item: %s not found\n", key)
	}
}

func (om *OrderedMap) GetItem(key string) (string, bool) {
	element, exists := om.data[key]
	if exists {
		return element.Value.(*Pair).Value, true
	}
	return "", false
}

func (om *OrderedMap) GetAllItems() []Pair {

	items := make([]Pair, 0, om.order.Len())
	for e := om.order.Front(); e != nil; e = e.Next() {
		pair := e.Value.(*Pair)
		items = append(items, *pair)
	}
	return items
}

// We use a mutex to ensure safe concurrent access to the map
func (om *OrderedMap) HandleCommand(message models.Command, logger *log.Logger, wg *sync.WaitGroup) {
	switch message.Action {
	case "addItem":
		om.AddItem(message.Key, message.Value)
		fmt.Printf("Added: %s -> %s\n", message.Key, message.Value)
	case "deleteItem":
		om.DeleteItem(message.Key)
		fmt.Printf("Deleted: %s\n", message.Key)
	case "getItem":
		om.mutex.RLock()
		wg.Add(1)
		go func() {
			defer om.mutex.RUnlock()
			defer wg.Done()
			value, exists := om.GetItem(message.Key)
			if exists {
				logger.Printf("Get Item: %s -> %s\n", message.Key, value)
				fmt.Printf("Get Item: %s -> %s\n", message.Key, value)
			} else {
				fmt.Printf("Item: %s not found\n", message.Key)
			}
		}()
	case "getAllItems":
		om.mutex.RLock()
		wg.Add(1)
		go func() {
			defer om.mutex.RUnlock()
			defer wg.Done()
			items := om.GetAllItems()
			logger.Printf("All items:")
			for _, item := range items {
				fmt.Printf("%s -> %s\n", item.Key, item.Value)
				logger.Printf("	Got Item: %s -> %s\n", item.Key, item.Value)
			}
		}()
	default:
		fmt.Printf("Unknown command: %s\n", message.Action)
	}
}
