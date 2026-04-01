package utils

func SliceToMap[T any, K comparable](slice []T, fn func(T) (K, T)) map[K]T {
	result := make(map[K]T)
	for _, item := range slice {
		key, value := fn(item)
		result[key] = value
	}
	return result
}

func MapSlice[T any, V any](slice []T, fn func(T) V) []V {
	result := make([]V, len(slice))
	for i, item := range slice {
		result[i] = fn(item)
	}
	return result
}

func MapToKeysSlice[T any, K comparable](data map[K]T) []K {
	slice := make([]K, 0)
	for k, _ := range data {
		slice = append(slice, k)
	}
	return slice
}

func Contains[T comparable](slice []T, item T) bool {
	for _, i := range slice {
		if i == item {
			return true
		}
	}
	return false
}

func Find[T any](slice []*T, fn func(*T) bool) *T {
	for _, item := range slice {
		if fn(item) {
			return item
		}
	}
	return nil
}
