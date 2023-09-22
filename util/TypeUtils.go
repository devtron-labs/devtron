package util

func GetDeReferencedArray[T any](t []*T) []T {
	deReferencedArray := make([]T, 0)
	for _, item := range t {
		deReferencedArray = append(deReferencedArray, *item)
	}
	return deReferencedArray
}

func GetReferencedArray[T any](t []T) []*T {
	deReferencedArray := make([]*T, 0)
	for i, _ := range t {
		deReferencedArray = append(deReferencedArray, &t[i])
	}
	return deReferencedArray
}
