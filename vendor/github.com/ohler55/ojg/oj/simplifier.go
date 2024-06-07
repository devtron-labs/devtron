// Copyright (c) 2021, Peter Ohler, All rights reserved.

package oj

import (
	"reflect"
	"unsafe"

	"github.com/ohler55/ojg/alt"
)

func appendSimplifier(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface()
	buf = append(buf, fi.jkey...)
	if (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0 {
		return buf, nil, aChanged
	}
	return buf, v.(alt.Simplifier).Simplify(), aChanged
}

func appendSimplifierNotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface()
	if (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0 { // real nil check
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	if s, ok := v.(alt.Simplifier); ok {
		v = s.Simplify()
	}
	return buf, v, aChanged
}

func appendSimplifierAddr(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Addr().Interface()
	buf = append(buf, fi.jkey...)

	return buf, v.(alt.Simplifier).Simplify(), aChanged
}
