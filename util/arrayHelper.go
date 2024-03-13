package util

func GetIds[T any](entities []T, getIdFunc func(entity T) int) []int {
	var Ids []int
	for _, entity := range entities {
		Ids = append(Ids, getIdFunc(entity))
	}
	return Ids
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
