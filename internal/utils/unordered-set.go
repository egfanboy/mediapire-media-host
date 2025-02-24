package utils

type UnorderedSet[T comparable] interface {
	Add(v T)
	Values() []T
}

type unorderedSet[T comparable] struct {
	values map[T]bool
}

func (s *unorderedSet[T]) Add(v T) {
	if s.values == nil {
		s.values = make(map[T]bool)
	}

	s.values[v] = true
}

func (s *unorderedSet[T]) Values() []T {
	vals := make([]T, 0)

	for v := range s.values {
		vals = append(vals, v)
	}

	return vals
}

func NewUnorderedSet[T comparable]() UnorderedSet[T] {
	return &unorderedSet[T]{}
}
