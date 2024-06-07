// Copyright (c) 2021, Peter Ohler, All rights reserved.

package oj

// ZeroHandler is a TokenHandler whose functions do nothing. It is used as an
// embedded member for TokenHandlers that don't care about all of the
// TokenHandler functions.
type ZeroHandler struct {
}

// Null is called when a JSON null is encountered.
func (z *ZeroHandler) Null() {
}

// Bool is called when a JSON true or false is encountered.
func (z *ZeroHandler) Bool(bool) {
}

// Int is called when a JSON integer is encountered.
func (z *ZeroHandler) Int(int64) {
}

// Float is called when a JSON decimal is encountered that fits into a
// float64.
func (z *ZeroHandler) Float(float64) {
}

// Number is called when a JSON number is encountered that does not fit
// into an int64 or float64.
func (z *ZeroHandler) Number(string) {
}

// String is called when a JSON string is encountered.
func (z *ZeroHandler) String(string) {
}

// ObjectStart is called when a JSON object start '{' is encountered.
func (z *ZeroHandler) ObjectStart() {
}

// ObjectEnd is called when a JSON object end '}' is encountered.
func (z *ZeroHandler) ObjectEnd() {
}

// Key is called when a JSON object key is encountered.
func (z *ZeroHandler) Key(string) {
}

// ArrayStart is called when a JSON array start '[' is encountered.
func (z *ZeroHandler) ArrayStart() {
}

// ArrayEnd is called when a JSON array end ']' is encountered.
func (z *ZeroHandler) ArrayEnd() {
}
