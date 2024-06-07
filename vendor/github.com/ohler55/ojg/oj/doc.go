// // Copyright (c) 2020, Peter Ohler, All rights reserved.

// Package oj contains functions and types to support building simple types
// where simple types are:
//
//	nil
//	bool
//	int64
//	float64
//	string
//	time.Time
//	[]any
//	map[string]any
//
// # Parser
//
// Parse a JSON file or stream. The parser can be used on a single JSON document or a
// document with multiple JSON elements. An option also exists to allow comments
// in the JSON.
//
//	v, err := oj.ParseString("[true,[false,[null],123],456]")
//
// or for a performance gain on repeated parses:
//
//	var p oj.Parser
//	v, err := p.Parse([]byte("[true,[false,[null],123],456]"))
//
// # Validator
//
// Validates a JSON file or stream. It can be used on a single JSON document or a
// document with multiple JSON elements. An option also exists to allow comments
// in the JSON.
//
//	err := oj.ValidateString("[true,[false,[null],123],456]")
//
// or for a slight performance gain on repeated validations:
//
//	var v oj.Validator
//	err := v.Validate([]byte("[true,[false,[null],123],456]"))
//
// # Builder
//
// An example of building simple data is:
//
//	var b oj.Builder
//
//	b.Object()
//	b.Value(1, "a")
//	b.Array("b")
//	b.Value(2)
//	b.Pop()
//	b.Pop()
//	v := b.Result()
//
//	// v: map[string]any{"a": 1, "b": []any{2}}
//
// # Writer
//
// The writer function's output data values to JSON. The basic oj.JSON() attempts
// to build JSON from any data provided skipping types that can not be converted.
//
//	s := oj.JSON([]any{1, 2, "abc", true})
//
// Output can also be use with an io.Writer.
//
//	var b strings.Builder
//
//	err := oj.Write(&b, []any{1, 2, "abc", true})
package oj
