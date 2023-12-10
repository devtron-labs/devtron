// Copyright (c) 2021, Peter Ohler, All rights reserved.

package oj

// TokenHandler describes an interface for handling tokens when using the
// Tokenizer for parsing JSON or SEN documents.
type TokenHandler interface {
	// Null is called when a JSON null is encountered.
	Null()

	// Bool is called when a JSON true or false is encountered.
	Bool(bool)

	// Int is called when a JSON integer is encountered.
	Int(int64)

	// Float is called when a JSON decimal is encountered that fits into a
	// float64.
	Float(float64)

	// Number is called when a JSON number is encountered that does not fit
	// into an int64 or float64.
	Number(string)

	// String is called when a JSON string is encountered.
	String(string)

	// ObjectStart is called when a JSON object start '{' is encountered.
	ObjectStart()

	// ObjectEnd is called when a JSON object end '}' is encountered.
	ObjectEnd()

	// Key is called when a JSON object key is encountered.
	Key(string)

	// ArrayStart is called when a JSON array start '[' is encountered.
	ArrayStart()

	// ArrayEnd is called when a JSON array end ']' is encountered.
	ArrayEnd()
}
