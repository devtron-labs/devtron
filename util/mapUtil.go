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

func MergeMaps(map1, map2 map[string][]string) {
	for key, values := range map2 {
		if existingValues, found := map1[key]; found {
			// Key exists in map1, append the values from map2
			map1[key] = append(existingValues, values...)
		} else {
			// Key does not exist in map1, add the new key-value pair
			map1[key] = values
		}
	}
}
