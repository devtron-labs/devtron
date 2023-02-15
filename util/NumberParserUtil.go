package util

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"strconv"
)

const EMPTY_VAL_ERR = "empty-val-err"

func ParseFloatNumber(inputMap map[string]interface{}, key string, merged []byte) (float64, error) {
	if _, ok := inputMap[key]; !ok {
		return 0, errors.New(EMPTY_VAL_ERR)
	}
	floatNumVal := fmt.Sprintf("%v", gjson.Get(string(merged), inputMap[key].(string)))
	floatNumber, err := strconv.ParseFloat(floatNumVal, 64)
	if err != nil {
		return 0, err
	}
	return floatNumber, nil
}
