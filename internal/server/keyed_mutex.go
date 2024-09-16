package server

import (
	"sync"
)

type KeyedMutex struct {
	mutexes       sync.Map
	getAllItemsWg sync.WaitGroup
}

func NewKeyedMutex() *KeyedMutex {
	m := &KeyedMutex{}
	return m
}

func (m *KeyedMutex) Lock(key string) func() {
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

	return func() {
		mtx.RUnlock()
	}
}
