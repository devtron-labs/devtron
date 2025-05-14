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

import "encoding/json"

// IsEmptyJSONForJsonString checks if a given string represents an empty JSON object
func IsEmptyJSONForJsonString(jsonStr string) (bool, error) {
	var data interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		// If unmarshaling fails, it's not valid JSON
		return false, err
	}

	// Check if data is an empty map
	if obj, ok := data.(map[string]interface{}); ok {
		return len(obj) == 0, nil
	}

	// If not a JSON object, return false
	return false, nil
}

func GetEmptyJSON() []byte {
	return []byte("{}")
}
