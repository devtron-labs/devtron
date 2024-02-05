package util

import "fmt"

func GetCommaSeparatedStringsFromIntArray(vals []int) string {
	res := ""
	for i, val := range vals {
		if i == 0 {
			res = fmt.Sprintf("%d", val)
		} else {
			res = fmt.Sprintf("%s,%d", res, val)
		}
	}
	return res
}
