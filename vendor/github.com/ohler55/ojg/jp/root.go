// Copyright (c) 2020, Peter Ohler, All rights reserved.

package jp

// Root represents a root $ in a JSON path representation.
type Root byte

// Append a fragment string representation of the fragment to the buffer
// then returning the expanded buffer.
func (f Root) Append(buf []byte, bracket, first bool) []byte {
	buf = append(buf, '$')
	return buf
}
