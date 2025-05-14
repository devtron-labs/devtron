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

package util

import (
	"math"
)

// XORBool returns the XOR of two boolean values
// XORBool(true, true) => false
// XORBool(true, false) => true
// XORBool(false, true) => true
// XORBool(false, false) => false
func XORBool(a, b bool) bool {
	return (a || b) && !(a && b)
}

// TruncateFloat truncates a float64 value to n decimal points using the math package.
func TruncateFloat(value float64, decimals int) float64 {
	pow10 := math.Pow10(decimals)
	return math.Trunc(value*pow10) / pow10
}

func GetDeReferencedBean[T any](ptrObject *T) T {
	var deReferencedObj T
	if ptrObject != nil {
		deReferencedObj = *ptrObject
	}
	return deReferencedObj
}
