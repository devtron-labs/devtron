package util

import "testing"

func TestValidateGlobalEntityName(t *testing.T) {
	validator, err := IntValidator()
	if err != nil {
		t.Error("error in creating validator")
		t.Fail()
	}

	type TestStruct struct {
		Name string `validate:"global-entity-name"`
	}

	validCases := []TestStruct{
		{
			Name: "test",
		},
		{
			Name: "test-1",
		},
		{
			Name: "1test",
		},
		{
			Name: "test1",
		},
		{
			Name: "test__2",
		},
		{
			Name: "test--2",
		},
		{
			Name: "test_-.2",
		},
	}

	inValidCases := []TestStruct{
		{
			Name: "Test",
		},
		{
			Name: "test1.",
		},
		{
			Name: "-test",
		},
		{
			Name: "_test",
		},
		{
			Name: "test_",
		},
		{
			Name: "test-",
		},
	}

	for _, test := range validCases {
		if err := validator.Struct(test); err != nil {
			t.Error("error in validating valid names", "name", test.Name)
			t.Fail()
		}
	}
	for _, test := range inValidCases {
		if err := validator.Struct(test); err == nil {
			t.Error("error in validating invalid names", "name", test.Name)
			t.Fail()
		}
	}

}
