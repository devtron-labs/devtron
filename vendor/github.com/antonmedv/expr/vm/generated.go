// Code generated by vm/func_types/main.go. DO NOT EDIT.

package vm

import (
	"fmt"
	"time"
)

var FuncTypes = []interface{}{
	1:  new(func() time.Duration),
	2:  new(func() time.Month),
	3:  new(func() time.Time),
	4:  new(func() time.Weekday),
	5:  new(func() []uint8),
	6:  new(func() []interface{}),
	7:  new(func() bool),
	8:  new(func() uint8),
	9:  new(func() float64),
	10: new(func() int),
	11: new(func() int64),
	12: new(func() interface{}),
	13: new(func() map[string]interface{}),
	14: new(func() int32),
	15: new(func() string),
	16: new(func() uint),
	17: new(func() uint64),
	18: new(func(time.Duration) time.Duration),
	19: new(func(time.Duration) time.Time),
	20: new(func(time.Time) time.Duration),
	21: new(func(time.Time) bool),
	22: new(func([]interface{}, string) string),
	23: new(func([]string, string) string),
	24: new(func(bool) bool),
	25: new(func(bool) float64),
	26: new(func(bool) int),
	27: new(func(bool) string),
	28: new(func(float64) bool),
	29: new(func(float64) float64),
	30: new(func(float64) int),
	31: new(func(float64) string),
	32: new(func(int) bool),
	33: new(func(int) float64),
	34: new(func(int) int),
	35: new(func(int) string),
	36: new(func(int, int) int),
	37: new(func(int, int) string),
	38: new(func(int64) time.Time),
	39: new(func(string) []string),
	40: new(func(string) bool),
	41: new(func(string) float64),
	42: new(func(string) int),
	43: new(func(string) string),
	44: new(func(string, uint8) int),
	45: new(func(string, int) int),
	46: new(func(string, int32) int),
	47: new(func(string, string) bool),
	48: new(func(string, string) string),
	49: new(func(interface{}) bool),
	50: new(func(interface{}) float64),
	51: new(func(interface{}) int),
	52: new(func(interface{}) string),
	53: new(func(interface{}) interface{}),
	54: new(func(interface{}) []interface{}),
	55: new(func(interface{}) map[string]interface{}),
	56: new(func([]interface{}) interface{}),
	57: new(func([]interface{}) []interface{}),
	58: new(func([]interface{}) map[string]interface{}),
	59: new(func(interface{}, interface{}) bool),
	60: new(func(interface{}, interface{}) string),
	61: new(func(interface{}, interface{}) interface{}),
	62: new(func(interface{}, interface{}) []interface{}),
}

func (vm *VM) call(fn interface{}, kind int) interface{} {
	switch kind {
	case 1:
		return fn.(func() time.Duration)()
	case 2:
		return fn.(func() time.Month)()
	case 3:
		return fn.(func() time.Time)()
	case 4:
		return fn.(func() time.Weekday)()
	case 5:
		return fn.(func() []uint8)()
	case 6:
		return fn.(func() []interface{})()
	case 7:
		return fn.(func() bool)()
	case 8:
		return fn.(func() uint8)()
	case 9:
		return fn.(func() float64)()
	case 10:
		return fn.(func() int)()
	case 11:
		return fn.(func() int64)()
	case 12:
		return fn.(func() interface{})()
	case 13:
		return fn.(func() map[string]interface{})()
	case 14:
		return fn.(func() int32)()
	case 15:
		return fn.(func() string)()
	case 16:
		return fn.(func() uint)()
	case 17:
		return fn.(func() uint64)()
	case 18:
		arg1 := vm.pop().(time.Duration)
		return fn.(func(time.Duration) time.Duration)(arg1)
	case 19:
		arg1 := vm.pop().(time.Duration)
		return fn.(func(time.Duration) time.Time)(arg1)
	case 20:
		arg1 := vm.pop().(time.Time)
		return fn.(func(time.Time) time.Duration)(arg1)
	case 21:
		arg1 := vm.pop().(time.Time)
		return fn.(func(time.Time) bool)(arg1)
	case 22:
		arg2 := vm.pop().(string)
		arg1 := vm.pop().([]interface{})
		return fn.(func([]interface{}, string) string)(arg1, arg2)
	case 23:
		arg2 := vm.pop().(string)
		arg1 := vm.pop().([]string)
		return fn.(func([]string, string) string)(arg1, arg2)
	case 24:
		arg1 := vm.pop().(bool)
		return fn.(func(bool) bool)(arg1)
	case 25:
		arg1 := vm.pop().(bool)
		return fn.(func(bool) float64)(arg1)
	case 26:
		arg1 := vm.pop().(bool)
		return fn.(func(bool) int)(arg1)
	case 27:
		arg1 := vm.pop().(bool)
		return fn.(func(bool) string)(arg1)
	case 28:
		arg1 := vm.pop().(float64)
		return fn.(func(float64) bool)(arg1)
	case 29:
		arg1 := vm.pop().(float64)
		return fn.(func(float64) float64)(arg1)
	case 30:
		arg1 := vm.pop().(float64)
		return fn.(func(float64) int)(arg1)
	case 31:
		arg1 := vm.pop().(float64)
		return fn.(func(float64) string)(arg1)
	case 32:
		arg1 := vm.pop().(int)
		return fn.(func(int) bool)(arg1)
	case 33:
		arg1 := vm.pop().(int)
		return fn.(func(int) float64)(arg1)
	case 34:
		arg1 := vm.pop().(int)
		return fn.(func(int) int)(arg1)
	case 35:
		arg1 := vm.pop().(int)
		return fn.(func(int) string)(arg1)
	case 36:
		arg2 := vm.pop().(int)
		arg1 := vm.pop().(int)
		return fn.(func(int, int) int)(arg1, arg2)
	case 37:
		arg2 := vm.pop().(int)
		arg1 := vm.pop().(int)
		return fn.(func(int, int) string)(arg1, arg2)
	case 38:
		arg1 := vm.pop().(int64)
		return fn.(func(int64) time.Time)(arg1)
	case 39:
		arg1 := vm.pop().(string)
		return fn.(func(string) []string)(arg1)
	case 40:
		arg1 := vm.pop().(string)
		return fn.(func(string) bool)(arg1)
	case 41:
		arg1 := vm.pop().(string)
		return fn.(func(string) float64)(arg1)
	case 42:
		arg1 := vm.pop().(string)
		return fn.(func(string) int)(arg1)
	case 43:
		arg1 := vm.pop().(string)
		return fn.(func(string) string)(arg1)
	case 44:
		arg2 := vm.pop().(uint8)
		arg1 := vm.pop().(string)
		return fn.(func(string, uint8) int)(arg1, arg2)
	case 45:
		arg2 := vm.pop().(int)
		arg1 := vm.pop().(string)
		return fn.(func(string, int) int)(arg1, arg2)
	case 46:
		arg2 := vm.pop().(int32)
		arg1 := vm.pop().(string)
		return fn.(func(string, int32) int)(arg1, arg2)
	case 47:
		arg2 := vm.pop().(string)
		arg1 := vm.pop().(string)
		return fn.(func(string, string) bool)(arg1, arg2)
	case 48:
		arg2 := vm.pop().(string)
		arg1 := vm.pop().(string)
		return fn.(func(string, string) string)(arg1, arg2)
	case 49:
		arg1 := vm.pop()
		return fn.(func(interface{}) bool)(arg1)
	case 50:
		arg1 := vm.pop()
		return fn.(func(interface{}) float64)(arg1)
	case 51:
		arg1 := vm.pop()
		return fn.(func(interface{}) int)(arg1)
	case 52:
		arg1 := vm.pop()
		return fn.(func(interface{}) string)(arg1)
	case 53:
		arg1 := vm.pop()
		return fn.(func(interface{}) interface{})(arg1)
	case 54:
		arg1 := vm.pop()
		return fn.(func(interface{}) []interface{})(arg1)
	case 55:
		arg1 := vm.pop()
		return fn.(func(interface{}) map[string]interface{})(arg1)
	case 56:
		arg1 := vm.pop().([]interface{})
		return fn.(func([]interface{}) interface{})(arg1)
	case 57:
		arg1 := vm.pop().([]interface{})
		return fn.(func([]interface{}) []interface{})(arg1)
	case 58:
		arg1 := vm.pop().([]interface{})
		return fn.(func([]interface{}) map[string]interface{})(arg1)
	case 59:
		arg2 := vm.pop()
		arg1 := vm.pop()
		return fn.(func(interface{}, interface{}) bool)(arg1, arg2)
	case 60:
		arg2 := vm.pop()
		arg1 := vm.pop()
		return fn.(func(interface{}, interface{}) string)(arg1, arg2)
	case 61:
		arg2 := vm.pop()
		arg1 := vm.pop()
		return fn.(func(interface{}, interface{}) interface{})(arg1, arg2)
	case 62:
		arg2 := vm.pop()
		arg1 := vm.pop()
		return fn.(func(interface{}, interface{}) []interface{})(arg1, arg2)

	}
	panic(fmt.Sprintf("unknown function kind (%v)", kind))
}
