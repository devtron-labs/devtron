package sliceUtil

// GetUniqueElements returns a new slice containing only the unique elements of the specified sliceList.
// for example, GetUniqueElements([1, 2, 3, 2, 1]) returns [1, 2, 3]
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

// GetMapOf returns a map with the specified sliceList as keys and defaultValue as values.
// for example, GetMapOf([1, 2, 3], "default") returns {1: "default", 2: "default", 3: "default"}
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

// GetSliceOf returns a slice containing the specified element.
// for example, GetSliceOf(1) returns [1]
func GetSliceOf[T any](element T) []T {
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

// Difference returns the elements in the slice `a` that aren't in slice `b`.
// Examples:
//   - Difference([1, 2, 3, 4], [1, 2, 5]) => [3, 4]
//   - Difference([1, 2, 3, 4], [1, 2, 3, 4]) => []
//   - Difference([1, 2, 3, 4], [5, 6, 7]) => [1, 2, 3, 4]
//   - Difference([1, 2, 3, 4], []) => [1, 2, 3, 4]
//   - Difference([], [1, 2, 3, 4]) => []
func Difference[T comparable](a, b []T) []T {
	mb := GetMapOf(b, true)
	diff := make([]T, 0, len(a))
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

// -----------------------------
// TODO: Add unit tests for the below functions
// -----------------------------

// GetDeReferencedSlice converts an array of pointers to an array of values
func GetDeReferencedSlice[T any](ptrObjects []*T) []T {
	deReferencedArray := make([]T, 0)
	for _, item := range ptrObjects {
		deReferencedArray = append(deReferencedArray, *item)
	}
	return deReferencedArray
}

// GetReferencedSlice converts an array of values to an array of pointers
func GetReferencedSlice[T any](objects []T) []*T {
	deReferencedArray := make([]*T, 0)
	for i, _ := range objects {
		deReferencedArray = append(deReferencedArray, &objects[i])
	}
	return deReferencedArray
}

// GetBeansPtr returns a slice of pointers to the beans passed as arguments
// GetBeansPtr(&bean1, &bean2, &bean3) => []*Bean{&bean1, &bean2, &bean3}
// ...*T is a variadic parameter that accepts a variable number of pointers to objects of type T.
// But the cost of using variadic parameters is that the compiler has to create a new slice and copy all the arguments into it.
// Note: Make sure to use this function only when the number of arguments is lower.
func GetBeansPtr[T any](beans ...*T) []*T {
	finalBeans := make([]*T, 0)
	for _, bean := range beans {
		if bean != nil {
			finalBeans = append(finalBeans, bean)
		}
	}
	return finalBeans
}

// GetBeans returns a slice of beans passed as arguments
// GetBeans(bean1, bean2, bean3) => []Bean{bean1, bean2, bean3}
// ...T is a variadic parameter that accepts a variable number of objects of type T.
// But the cost of using variadic parameters is that the compiler has to create a new slice and copy all the arguments into it.
// Note: Make sure to use this function only when the number of arguments is lower.
func GetBeans[T any](beans ...T) []T {
	return beans
}

// NewSliceFromFuncExec applies the given function to each element of the input slice
// And returns a new slice with the transformed elements
// NewSliceFromFuncExec([1, 2, 3], func(x int) int { return x * 2 }) => [2, 4, 6]
func NewSliceFromFuncExec[T any, K any](input []T, transform func(inp T) K) []K {
	res := make([]K, len(input))
	for i, _ := range input {
		res[i] = transform(input[i])
	}
	return res
}

// NewMapFromFuncExec applies the given function to each element of the input slice
// And returns a new map with the transformed elements as keys and the original elements as values
// NewMapFromFuncExec([1, 2, 3], func(x int) string { return strconv.Itoa(x) }) => {"1": 1, "2": 2, "3": 3}
// NewMapFromFuncExec([{Name: "John", Age: 25}, {Name: "Doe", Age: 30}], func(x Person) string { return x.Name }) => {"John": {Name: "John", Age: 25}, "Doe": {Name: "Doe", Age: 30}}
// Note: The keys of the map should be unique, otherwise the last element with the same key will be stored in the map
func NewMapFromFuncExec[T any, K comparable](input []T, transform func(inp T) K) map[K]T {
	res := make(map[K]T)
	for i, _ := range input {
		res[transform(input[i])] = input[i]
	}
	return res
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

// GetMapValuesPtr returns a slice of pointers to the values of the map passed as an argument
func GetMapValuesPtr[T any](valueMap map[string]*T) []*T {
	values := make([]*T, 0)
	for key := range valueMap {
		values = append(values, valueMap[key])
	}
	return values
}
