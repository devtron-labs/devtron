/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package util

import (
	"github.com/google/go-cmp/cmp"
	"sort"
)

func CompareUnOrdered(a, b []int) bool {
	sort.Ints(a)
	sort.Ints(b)
	return cmp.Equal(a, b)
}

func GetTruncatedMessage(message string, maxLength int) string {
	_length := len(message)
	if _length == 0 {
		return message
	}
	_truncatedLength := maxLength
	if _length < _truncatedLength {
		return message
	} else {
		if _truncatedLength > 3 {
			return message[:(_truncatedLength-3)] + "..."
		}
		return message[:_truncatedLength]
	}
}
