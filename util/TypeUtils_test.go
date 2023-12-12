package util

import (
	"fmt"
	"testing"
)

func TestTransform(t *testing.T) {
	t.Run("primitive transform", func(tt *testing.T) {
		input := []int{1, 2, 3, 4}
		expectedOutput := []string{"1", "2", "3", "4"}
		transformer := func(a int) string {
			return fmt.Sprintf("%d", a)
		}
		output := Transform(input, transformer)
		for i, out := range output {
			if expectedOutput[i] != out {
				tt.Fail()
			}
		}
	})

	t.Run("structs transform", func(tt *testing.T) {
		type A struct {
			Name *string
		}
		input := []int{1, 2, 3, 4}
		expectedOutput := []string{"1", "2", "3", "4"}
		transformer := func(a int) string {
			return fmt.Sprintf("%d", a)
		}
		output := Transform(input, transformer)
		for i, out := range output {
			if expectedOutput[i] != out {
				tt.Fail()
			}
		}
	})

}
