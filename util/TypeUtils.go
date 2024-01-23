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

func GetBeansPtr[T any](beans ...*T) []*T {

	finalBeans := make([]*T, 0)
	for _, bean := range beans {
		if bean != nil {
			finalBeans = append(finalBeans, bean)
		}
	}
	return finalBeans
}

func GetBeans[T any](beans ...T) []T {
	return beans
}

func GetMapValuesPtr[T any](valueMap map[string]*T) []*T {
	values := make([]*T, 0)
	for key := range valueMap {
		values = append(values, valueMap[key])
	}
	return values
}

func Transform[T any, K any](input []T, transform func(inp T) K) []K {

	res := make([]K, len(input))
	for i, _ := range input {
		res[i] = transform(input[i])
	}
	return res

}

func Contains[T any](input []T, check func(inp T) bool) bool {
	for i, _ := range input {
		if check(input[i]) {
			return true
		}
	}
	return false
}
