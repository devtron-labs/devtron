package sliceUtil

func GetUniqueElements[T comparable](sliceList []T) []T {
	if len(sliceList) == 0 {
		return sliceList
	}
	allKeys := make(map[T]bool)
	var list []T
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func GetMapOf[K comparable, V comparable](sliceList []K, defaultValue V) map[K]V {
	if len(sliceList) == 0 {
		return make(map[K]V)
	}
	result := make(map[K]V)
	for _, item := range sliceList {
		result[item] = defaultValue
	}
	return result
}

func GetSliceOfElement[T any](element T) []T {
	return []T{element}
}

// CompareTwoSlices asserts that the specified listA(array, slice...) is equal to specified
// listB(array, slice...) ignoring the order of the elements. If there are duplicate elements,
// the number of appearances of each of them in both lists should match.
//
// CompareTwoSlices([1, 3, 2, 3], [1, 3, 3, 2])
func CompareTwoSlices[T comparable](listA, listB []T) bool {
	if len(listA) != len(listB) {
		return false
	}
	diff := make(map[T]int, len(listA))
	for _, a := range listA {
		diff[a]++
	}

	for _, b := range listB {
		if _, ok := diff[b]; !ok {
			return false
		}
		diff[b]--
		if diff[b] == 0 {
			delete(diff, b)
		}
	}
	return len(diff) == 0
}
