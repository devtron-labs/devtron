// Copyright (c) 2021, Peter Ohler, All rights reserved.

package oj

import (
	"reflect"
	"unsafe"

	"github.com/ohler55/ojg/alt"
)

func appendGenericer(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface()
	buf = append(buf, fi.jkey...)
	if (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0 {
		return buf, nil, aChanged
	}
	if g, _ := v.(alt.Genericer); g != nil {
		if n := g.Generic(); n != nil {
			v = n.Simplify()
		}
	}
	return buf, v, aChanged
}

func appendGenericerNotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface()
	if (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0 { // real nil check
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	if g, ok := v.(alt.Genericer); ok {
		if n := g.Generic(); n != nil {
			v = n.Simplify()
		}
	}
	return buf, v, aChanged
}

func appendGenericerAddr(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Addr().Interface()
	buf = append(buf, fi.jkey...)
	if g, _ := v.(alt.Genericer); g != nil {
		if n := g.Generic(); n != nil {
			v = n.Simplify()
		}
	}
	return buf, v, aChanged
}
