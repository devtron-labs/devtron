package main

import "fmt"

type Response struct {
	Field1 *string
	Field2 *int
}

func Provider() *Response {
	// Simulate provider function returning a response object
	return &Response{
		Field1: nil,
		Field2: nil,
	}
}

func InitializeAppTest() {
	response := Provider()
	fmt.Println(response)
}
