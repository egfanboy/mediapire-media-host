package utils

import (
	"sync"
)

type concurrentMap[K comparable, T any] struct {
	mu   sync.Mutex
	data map[K]T
}

func (m *concurrentMap[K, T]) Get() map[K]T {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.data
}

func (m *concurrentMap[K, T]) GetKey(k K) (T, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	v, ok := m.data[k]
	return v, ok
}

func (m *concurrentMap[K, T]) Add(k K, v T) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[k] = v
}

func (m *concurrentMap[K, T]) Delete(k K) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, k)
}

func (m *concurrentMap[K, T]) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.data)
}

func NewConcurrentMap[K comparable, T any]() *concurrentMap[K, T] {
	m := &concurrentMap[K, T]{
		mu:   sync.Mutex{},
		data: map[K]T{},
	}
	return m
}
