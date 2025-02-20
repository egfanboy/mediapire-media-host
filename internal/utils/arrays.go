package utils

func Filter[T any](src []T, predicate func(arg T) bool) []T {
	result := make([]T, 0)

	for _, v := range src {
		if predicate(v) {
			result = append(result, v)
		}
	}

	return result
}

func Contains[T comparable](src []T, value T) bool {
	for _, v := range src {
		if v == value {
			return true
		}
	}

	return false
}
