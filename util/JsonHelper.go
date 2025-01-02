package util

import (
	"encoding/json"
	"fmt"
)

// IsEmptyJSON checks if the given json is empty or not and only throws error if there is some issue in unmarshalling
func IsEmptyJSON(s []byte) (bool, error) {
	var result map[string]interface{}
	err := json.Unmarshal(s, &result)
	if err != nil {
		fmt.Println("Error unmarshalling:", err)
		return false, err
	}
	if len(result) == 0 {
		return true, nil
	}
	return false, nil
}
