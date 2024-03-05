package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestStruct struct {
	Name        string `json:"name" devtronSearchableField:"name"`
	Description string `json:"description" devtronSearchableField:"description"`
	Count       int    `json:"count" devtronSearchableField:"count"`
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
