package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestStruct struct {
	Name        string `json:"name" isSearchField:"true"`
	Description string `json:"description" isSearchField:"true"`
	Count       int    `json:"count" isSearchField:"true"`
}

func TestGetSearchableFields(t *testing.T) {
	test := TestStruct{
		Name:        "test",
		Description: "testing",
		Count:       10,
	}
	fields := GetSearchableFields(test)
	assert.NotNil(t, fields)
}
