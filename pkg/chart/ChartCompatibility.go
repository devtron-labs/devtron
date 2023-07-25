package chart

var chartCompatibilityMatrix = map[int]map[int]struct{}{
	32: {33: struct{}{}},
	33: {32: struct{}{}},
}

func CheckCompatibility(oldChartId, newChartId int) bool {
	compatibilityOfOld, found := chartCompatibilityMatrix[oldChartId]
	if !found {
		return false
	}
	_, found = compatibilityOfOld[newChartId]
	if !found {
		return false
	}
	return true
}
