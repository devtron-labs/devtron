// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file was split out from Go's crypto/cipher/xor.go

// +build !386,!amd64,!ppc64,!ppc64le,!s390x,!appengine

package sasl

func xorBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}
	return n
}

func xorWords(dst, a, b []byte) {
	xorBytes(dst, a, b)
}
