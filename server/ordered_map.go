package server

import (
	"container/list"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nircoren/lightblocks/pkg/queue/models"
)

// Lock keys when writing, rlock when reading
type KeyedMutex struct {
	mutexes       sync.Map
	getAllItemsWg sync.WaitGroup
	// Locked blocks all write operations until it is false
}

func NewKeyedMutex() *KeyedMutex {
	m := &KeyedMutex{}
	return m
}

func (m *KeyedMutex) Lock(key string) func() {
	// If getAllItems is running, block all write operations by waiting for it to finish
	m.getAllItemsWg.Wait()
	value, _ := m.mutexes.LoadOrStore(key, &sync.RWMutex{})
	mtx := value.(*sync.RWMutex)
	mtx.Lock()

	return func() {
		mtx.Unlock()
	}
}

func (m *KeyedMutex) RLock(key string) func() {
	value, _ := m.mutexes.LoadOrStore(key, &sync.RWMutex{})
	mtx := value.(*sync.RWMutex)
	mtx.RLock()

	return func() { mtx.RUnlock() }
}

// The logic is having a map store key that is pointing to a doubly linked list element.
// This way we can have O(1) delete, insert and update operations.
type OrderedMap struct {
	data       map[string]*list.Element
	order      *list.List
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

func (om *OrderedMap) AddItem(key string, value string) string {
	element, exists := om.data[key]
	if exists {
		element.Value.(*Pair).Value = value
	} else {
		// Insert new element
		pair := &Pair{key, value}
		element := om.order.PushBack(pair)
		om.data[key] = element
	}
	return om.data[key].Value.(*Pair).Value
}

func (om *OrderedMap) DeleteItem(key string) bool {
	element, exists := om.data[key]
	var isDeleted bool = false
	if exists {
		om.order.Remove(element)
		delete(om.data, key)
		isDeleted = true
	}
	return isDeleted
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
func (om *OrderedMap) HandleCommand(message models.Command, logger *log.Logger) {
	switch message.Action {
	case "addItem":
		unlock := om.KeyedMutex.Lock(message.Key)
		val := om.AddItem(message.Key, message.Value)
		fmt.Printf("Added: %s -> %s\n", message.Key, val)
		unlock()
	case "deleteItem":
		unlock := om.KeyedMutex.Lock(message.Key)
		isDeleted := om.DeleteItem(message.Key)
		if isDeleted {
			fmt.Printf("Deleted: %s\n", message.Key)
		} else {
			fmt.Printf("Item: %s not found\n", message.Key)
		}
		unlock()
	case "getItem":
		rUnlock := om.KeyedMutex.RLock(message.Key)
		go func(rUnlock func()) {
			defer rUnlock()
			value, exists := om.GetItem(message.Key)
			if exists {
				logger.Printf("Get Item: %s -> %s\n", message.Key, value)
				fmt.Printf("Get Item: %s -> %s\n", message.Key, value)
			} else {
				logger.Printf("Item: %s not found\n", message.Key)
				fmt.Printf("Item: %s not found\n", message.Key)
			}
		}(rUnlock)
	case "getAllItems":
		om.KeyedMutex.getAllItemsWg.Add(1)
		go func() {
			defer om.KeyedMutex.getAllItemsWg.Done()
			items := om.GetAllItems()

			fmt.Printf("All items:\n")
			logger.Printf("All items:\n")
			for _, item := range items {
				fmt.Printf("	%s -> %s\n", item.Key, item.Value)
				logger.Printf("	%s -> %s\n", item.Key, item.Value)
			}
			time.Sleep(4 * time.Second)
		}()
	default:
		fmt.Printf("Unknown command: %s\n", message.Action)
	}
}
