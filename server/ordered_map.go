package server

import (
	"container/list"
	"fmt"
	"sync"
)

// OrderedMap structure
type OrderedMap struct {
	data  map[string]*list.Element
	order *list.List
	mutex *sync.RWMutex
}

type Pair struct {
	Key   string
	Value string
}

func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		data:  make(map[string]*list.Element),
		order: list.New(),
		mutex: &sync.RWMutex{},
	}
}

func (om *OrderedMap) AddItem(key string, value string) {
	om.mutex.RLock()
	defer om.mutex.RUnlock()
	element, exists := om.data[key]
	if exists {
		element.Value.(*Pair).Value = value
	} else {
		// Insert new element
		pair := &Pair{key, value}
		element := om.order.PushBack(pair)
		om.data[key] = element
	}
}

func (om *OrderedMap) DeleteItem(key string) {
	om.mutex.RLock()
	defer om.mutex.RUnlock()
	element, exists := om.data[key]
	if exists {
		om.order.Remove(element)
		delete(om.data, key)
	} else {
		fmt.Printf("Item: %s not found\n", key)
	}
}

func (om *OrderedMap) GetItem(key string) (string, bool) {
	om.mutex.RLock()
	defer om.mutex.RUnlock()
	element, exists := om.data[key]
	if exists {
		return element.Value.(*Pair).Value, true
	}
	return "", false
}

func (om *OrderedMap) GetAllItems() []Pair {
	om.mutex.RLock()
	defer om.mutex.RUnlock()

	items := make([]Pair, 0, om.order.Len())
	for e := om.order.Front(); e != nil; e = e.Next() {
		pair := e.Value.(*Pair)
		items = append(items, *pair)
	}
	return items
}
