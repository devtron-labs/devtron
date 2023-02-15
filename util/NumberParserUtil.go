package util

import (
	"errors"
	"fmt"
	"strconv"
)

const EMPTY_VAL_ERR = "empty-val-err"

func ParseFloatNumber(inputMap map[string]interface{}, key string) (float64, error) {
	if _, ok := inputMap[key]; !ok {
		return 0, errors.New(EMPTY_VAL_ERR)
	}
	floatNumVal := fmt.Sprintf("%v", inputMap[key])
	floatNumber, err := strconv.ParseFloat(floatNumVal, 64)
	if err != nil {
		return 0, err
	}
	return floatNumber, nil
}
