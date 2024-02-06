package util

import (
	"fmt"
	"k8s.io/utils/pointer"
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
		type B struct {
			Name string
		}

		input := []B{
			{"1"},
			{"2"},
			{"3"},
			{"4"},
		}
		expectedOutput := []A{
			{pointer.String("1")},
			{pointer.String("2")},
			{pointer.String("3")},
			{pointer.String("4")},
		}

		transformer := func(a B) A {
			return A{&a.Name}
		}
		output := Transform(input, transformer)
		for i, out := range output {
			if *expectedOutput[i].Name != *out.Name {
				tt.Fail()
			}
		}
	})

}
