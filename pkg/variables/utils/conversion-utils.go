package utils

import mapset "github.com/deckarep/golang-set"

// ToInterfaceArray converts an array of string to an array of interface{}
func ToInterfaceArray(arr []string) []interface{} {
	interfaceArr := make([]interface{}, len(arr))
	for i, v := range arr {
		interfaceArr[i] = v
	}
	return interfaceArr
}

// ToStringArray converts an array of interface{} back to an array of string
func ToStringArray(interfaceArr []interface{}) []string {
	stringArr := make([]string, len(interfaceArr))
	for i, v := range interfaceArr {
		stringArr[i] = v.(string)
	}
	return stringArr
}

// ToIntArray converts an array of interface{} back to an array of int
func ToIntArray(interfaceArr []interface{}) []int {
	intArr := make([]int, len(interfaceArr))
	for i, v := range interfaceArr {
		intArr[i] = v.(int)
	}
	return intArr
}

func FilterDuplicatesInStringArray(items []string) []string {
	itemsSet := mapset.NewSetFromSlice(ToInterfaceArray(items))
	uniqueItems := ToStringArray(itemsSet.ToSlice())
	return uniqueItems
}
