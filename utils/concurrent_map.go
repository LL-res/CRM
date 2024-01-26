package utils

import (
	"fmt"
	"sync"
)

type ConcurrentMap[K comparable, T any] struct {
	Data map[K]T
	sync.RWMutex
}

func (m *ConcurrentMap[K, T]) Load(key K) (T, error) {
	m.RLock()
	defer m.RUnlock()
	val, ok := m.Data[key]
	if !ok {
		return val, fmt.Errorf("value not found,key [%v]", key)
	}
	return val, nil
}
func (m *ConcurrentMap[K, T]) Range(do func(key K, val T, attr ...any) error, attr ...any) (errors []error) {
	m.Lock()
	defer m.Unlock()
	for k, v := range m.Data {
		if err := do(k, v, attr...); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}
func (m *ConcurrentMap[K, T]) Store(key K, val T) {
	m.Lock()
	defer m.Unlock()
	m.Data[key] = val
}
func (m *ConcurrentMap[K, T]) Delete(key K) {
	m.Lock()
	defer m.Unlock()
	delete(m.Data, key)
}
func (m *ConcurrentMap[K, T]) NewConcurrentMap() {
	if m.Data == nil {
		m.Data = make(map[K]T)
	}
}
func NewConcurrentMap[K comparable, T any]() *ConcurrentMap[K, T] {
	return &ConcurrentMap[K, T]{
		Data: make(map[K]T),
	}
}
