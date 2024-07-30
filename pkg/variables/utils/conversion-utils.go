/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

func PartitionSlice[T any](array []T, chunkSize int) [][]T {
	partitionedArray := make([][]T, 0)
	for index := 0; index < len(array); {
		chunk := make([]T, 0)
		for i := 0; i < chunkSize; i++ {
			if index+i == len(array) {
				break
			}
			chunk = append(chunk, array[index+i])
		}
		partitionedArray = append(partitionedArray, chunk)
		index = index + chunkSize
	}
	return partitionedArray
}
