package units

import (
	"fmt"
	"testing"
)

// todo: add more test cases
func TestParseQuantityString(t *testing.T) {
	memLimit := "01.400Gi"
	pos, val, num, denom, suf, err := ParseQuantityString(memLimit)
	fmt.Println("pos: ", pos)
	fmt.Println("val: ", val)
	fmt.Println("num: ", num)
	fmt.Println("denom: ", denom)
	fmt.Println("suf: ", suf)
	fmt.Println("err: ", err)

}
