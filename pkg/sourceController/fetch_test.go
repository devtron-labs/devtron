package main

import "testing"

func Test_fetch(t *testing.T) {
	tests := []struct {
		name string
	}{{name: "hello"}} // TODO: Add test cases.

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetch()
		})
	}
}
