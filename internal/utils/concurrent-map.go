package utils

import (
	"sync"
)

type ConcurrentMap[K comparable, T any] struct {
	mu   sync.Mutex
	data map[K]T
}

func (m *ConcurrentMap[K, T]) Get() map[K]T {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.data
}

func (m *ConcurrentMap[K, T]) GetKey(k K) (T, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	v, ok := m.data[k]
	return v, ok
}

func (m *ConcurrentMap[K, T]) Add(k K, v T) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[k] = v
}

func (m *ConcurrentMap[K, T]) Delete(k K) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, k)
}

func (m *ConcurrentMap[K, T]) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.data)
}

func (m *ConcurrentMap[K, T]) Values() map[K]T {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.data
}

func NewConcurrentMap[K comparable, T any]() *ConcurrentMap[K, T] {
	m := &ConcurrentMap[K, T]{
		mu:   sync.Mutex{},
		data: map[K]T{},
	}
	return m
}
