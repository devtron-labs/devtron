package util

import "encoding/json"

// IsEmptyJSON checks if a given string represents an empty JSON object
func IsEmptyJSONForJsonString(jsonStr string) bool {
	var data interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		// If unmarshaling fails, it's not valid JSON
		return false
	}

	// Check if data is an empty map
	if obj, ok := data.(map[string]interface{}); ok {
		return len(obj) == 0
	}

	// If not a JSON object, return false
	return false
}
