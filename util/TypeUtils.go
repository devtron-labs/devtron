package util

import (
	"strconv"
	"strings"
)

func GetDeReferencedArray[T any](ptrObjects []*T) []T {
	deReferencedArray := make([]T, 0)
	for _, item := range ptrObjects {
		deReferencedArray = append(deReferencedArray, *item)
	}
	return deReferencedArray
}

func GetReferencedArray[T any](objects []T) []*T {
	deReferencedArray := make([]*T, 0)
	for i, _ := range objects {
		deReferencedArray = append(deReferencedArray, &objects[i])
	}
	return deReferencedArray
}

func SplitCommaSeparatedIntValues(input string) ([]int, error) {
	items := make([]int, 0)
	itemSlices := strings.Split(input, ",")
	for _, envId := range itemSlices {
		id, err := strconv.Atoi(envId)
		if err != nil {
			return items, err
		}
		items = append(items, id)
	}
	return items, nil
}
