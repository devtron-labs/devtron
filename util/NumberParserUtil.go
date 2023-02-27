package util

import (
	"fmt"
	"strconv"
)

func ParseFloatNumber(inputVal interface{}) (float64, error) {
	floatNumVal := fmt.Sprintf("%v", inputVal)
	floatNumber, err := strconv.ParseFloat(floatNumVal, 64)
	if err != nil {
		return 0, err
	}
	return floatNumber, nil
}
