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

package util

import "strings"

const (
	ApiTokenPrefix = "API-TOKEN:"
)

func CheckValidationForRoleGroupCreation(name string) bool {
	if strings.Contains(name, ",") {
		return false
	}
	return true
}

func CheckIfAdminOrApiToken(email string) bool {
	if email == "admin" || CheckIfApiToken(email) {
		return true
	}
	return false
}

func CheckIfApiToken(email string) bool {
	return strings.HasPrefix(email, ApiTokenPrefix)
}

// GetCrossProductMappingForSlices generates a cross product mapping for passed on slice, order of function
// params matter where slice1 values will be the key and slice2 values will be corresponding map's values.
// eg. params:- slice1:= [a,b], slice2:= [p,q] ;returns:- a:p, a:q, b:p, b:q
func GetCrossProductMappingForSlices[T comparable](slice1 []T, slice2 []T) map[T]T {
	entityMapping := make(map[T]T, len(slice1)*len(slice2))
	for _, item1 := range slice1 {
		for _, item2 := range slice2 {
			entityMapping[item1] = item2
		}
	}
	return entityMapping
}
