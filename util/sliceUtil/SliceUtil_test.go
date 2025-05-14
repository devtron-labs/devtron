/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sliceUtil

import (
	"reflect"
	"testing"
)

func TestSliceUtil(t *testing.T) {
	t.Run("GetUniqueElementsWithDuplicates", TestGetUniqueElementsWithDuplicates)
	t.Run("GetUniqueElementsWithNoDuplicates", TestGetUniqueElementsWithNoDuplicates)
	t.Run("TestGetUniqueElementsWithNilSlice", TestGetUniqueElementsWithNilSlice)
	t.Run("GetUniqueElementsWithEmptySlice", TestGetUniqueElementsWithEmptySlice)
	t.Run("GetUniqueElementsWithAllDuplicates", TestGetUniqueElementsWithAllDuplicates)
	t.Run("GetUniqueElementsWithLargeInput", TestGetUniqueElementsWithLargeInput)
	t.Run("GetMapOfWithNonEmptySlice", TestGetMapOfWithNonEmptySlice)
	t.Run("GetMapOfWithEmptySlice", TestGetMapOfWithEmptySlice)
	t.Run("GetMapOfWithLargeInput", TestGetMapOfWithLargeInput)
	t.Run("GetSliceOf", TestGetSliceOfElement)
	t.Run("GetSliceOfElementWithZeroValue", TestGetSliceOfElementWithZeroValue)
	t.Run("CompareTwoSlicesEqualIgnoringOrder", TestCompareTwoSlicesEqualIgnoringOrder)
	t.Run("CompareTwoSlicesNotEqual", TestCompareTwoSlicesNotEqual)
	t.Run("CompareTwoSlicesDifferentLengths", TestCompareTwoSlicesDifferentLengths)
	t.Run("CompareTwoSlicesWithDifferentOrder", TestCompareTwoSlicesWithDifferentOrder)
	t.Run("CompareTwoSlicesWithDuplicatesInOne", TestCompareTwoSlicesWithDuplicatesInOne)
}

func TestGetUniqueElementsWithDuplicates(t *testing.T) {
	input := []int{1, 2, 3, 2, 1}
	expected := []int{1, 2, 3}
	result := GetUniqueElements(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestGetUniqueElementsWithNoDuplicates(t *testing.T) {
	input := []int{1, 2, 3}
	expected := []int{1, 2, 3}
	result := GetUniqueElements(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestGetUniqueElementsWithNilSlice(t *testing.T) {
	var input, expected []int
	result := GetUniqueElements(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestGetUniqueElementsWithEmptySlice(t *testing.T) {
	input := []int{}
	expected := []int{}
	result := GetUniqueElements(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestGetUniqueElementsWithAllDuplicates(t *testing.T) {
	input := []int{1, 1, 1, 1}
	expected := []int{1}
	result := GetUniqueElements(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestGetUniqueElementsWithLargeInput(t *testing.T) {
	input := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		input[i] = i % 10
	}
	expected := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	result := GetUniqueElements(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestGetMapOfWithNonEmptySlice(t *testing.T) {
	input := []int{1, 2, 3}
	defaultValue := "default"
	expected := map[int]string{1: "default", 2: "default", 3: "default"}
	result := GetMapOf(input, defaultValue)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestGetMapOfWithEmptySlice(t *testing.T) {
	input := []int{}
	defaultValue := "default"
	expected := map[int]string{}
	result := GetMapOf(input, defaultValue)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestGetMapOfWithLargeInput(t *testing.T) {
	input := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		input[i] = i
	}
	defaultValue := "default"
	expected := make(map[int]string, 1000)
	for i := 0; i < 1000; i++ {
		expected[i] = defaultValue
	}
	result := GetMapOf(input, defaultValue)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestGetSliceOfElement(t *testing.T) {
	element := 1
	expected := []int{1}
	result := GetSliceOf(element)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestGetSliceOfElementWithZeroValue(t *testing.T) {
	type test_element struct {
		key1 string
		key2 int
		key3 bool
		key4 float64
		key5 []int
		key6 map[string]string
	}
	element := &test_element{}
	expected := []*test_element{{}}
	result := GetSliceOf(element)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestCompareTwoSlicesEqualIgnoringOrder(t *testing.T) {
	listA := []int{1, 3, 2, 3}
	listB := []int{1, 3, 3, 2}
	if !CompareTwoSlices(listA, listB) {
		t.Errorf("expected true, got false")
	}
}

func TestCompareTwoSlicesNotEqual(t *testing.T) {
	listA := []int{1, 2, 3}
	listB := []int{1, 2, 2}
	if CompareTwoSlices(listA, listB) {
		t.Errorf("expected false, got true")
	}
}

func TestCompareTwoSlicesDifferentLengths(t *testing.T) {
	listA := []int{1, 2, 3}
	listB := []int{1, 2}
	if CompareTwoSlices(listA, listB) {
		t.Errorf("expected false, got true")
	}
}

func TestCompareTwoSlicesWithDifferentOrder(t *testing.T) {
	listA := []int{1, 2, 3, 4, 5}
	listB := []int{5, 4, 3, 2, 1}
	if !CompareTwoSlices(listA, listB) {
		t.Errorf("expected true, got false")
	}
}

func TestCompareTwoSlicesWithDuplicatesInOne(t *testing.T) {
	listA := []int{1, 2, 2, 3}
	listB := []int{1, 2, 3}
	if CompareTwoSlices(listA, listB) {
		t.Errorf("expected false, got true")
	}
}
