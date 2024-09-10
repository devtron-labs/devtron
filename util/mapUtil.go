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
