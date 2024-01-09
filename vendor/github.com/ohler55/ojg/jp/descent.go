// Copyright (c) 2020, Peter Ohler, All rights reserved.

package jp

// Descent is used as a flag to indicate the path should be displayed in a
// recursive descent representation.
type Descent byte

// Append a fragment string representation of the fragment to the buffer
// then returning the expanded buffer.
func (f Descent) Append(buf []byte, bracket, first bool) []byte {
	if bracket {
		buf = append(buf, "[..]"...)
	} else {
		buf = append(buf, '.')
	}
	return buf
}
