package util

func GetArrayObject[T any, R any](entities []T, getValFunc func(entity T) R) []R {
	objArr := make([]R, len(entities))
	for i, _ := range entities {
		objArr[i] = getValFunc(entities[i])
	}
	return objArr
}

func GetIdToObjectMap[T any](entities []T, getIdFunc func(entity T) int) map[int]T {
	idToEntityMap := make(map[int]T)
	for _, entity := range entities {
		idToEntityMap[getIdFunc(entity)] = entity
	}
	return idToEntityMap
}

func GetIdToIdMapping[T any](entities []T, getKeyValueFunc func(entity T) (key int, value int)) map[int]int {
	idToEntityMap := make(map[int]int)
	for _, entity := range entities {
		key, val := getKeyValueFunc(entity)
		idToEntityMap[key] = val
	}
	return idToEntityMap
}
