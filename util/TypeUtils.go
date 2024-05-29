/*
 * Copyright (c) 2024. Devtron Inc.
 */

package util

import (
	"golang.org/x/exp/maps"
	"math"
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

func XORBool(a, b bool) bool {
	return (a || b) && !(a && b)
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

func Map[T any, K any](input []T, transform func(inp T) K) []K {

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

// ContainsStringAlias reports whether v is present in s.
func ContainsStringAlias[S ~[]E, E ~string](s S, v E) bool {
	for i := range s {
		if v == s[i] {
			return true
		}
	}
	return false
}

// TruncateFloat truncates a float64 value to n decimal points using the math package.
func TruncateFloat(value float64, decimals int) float64 {
	pow10 := math.Pow10(decimals)
	return math.Trunc(value*pow10) / pow10
}

func GetUniqueKeys[T string | int](keys []T) []T {
	set := make(map[T]bool)
	for _, key := range keys {
		set[key] = true
	}
	return maps.Keys(set)
}
