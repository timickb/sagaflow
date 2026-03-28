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
