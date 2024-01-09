// Copyright (c) 2020, Peter Ohler, All rights reserved.

package jp

// Bracket is used as a flag to indicate the path should be displayed in a
// bracketed representation.
type Bracket byte

// Append a fragment string representation of the fragment to the buffer
// then returning the expanded buffer.
func (f Bracket) Append(buf []byte, bracket, first bool) []byte {
	return buf
}
