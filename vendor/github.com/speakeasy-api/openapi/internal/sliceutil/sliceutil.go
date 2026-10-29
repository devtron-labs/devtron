package sliceutil

func Map[T any, U any](slice []T, fn func(T) U) []U {
	mapped := make([]U, len(slice))
	for i, elem := range slice {
		mapped[i] = fn(elem)
	}
	return mapped
}
