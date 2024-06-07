// Copyright (c) 2020, Peter Ohler, All rights reserved.

package jp

// At is the @ in a JSON path representation.
type At byte

// Append a fragment string representation of the fragment to the buffer
// then returning the expanded buffer.
func (f At) Append(buf []byte, bracket, first bool) []byte {
	buf = append(buf, '@')
	return buf
}
