// Copyright (c) 2021, Peter Ohler, All rights reserved.

package oj

import (
	"encoding"
	"reflect"
	"unsafe"

	"github.com/ohler55/ojg"
)

func appendTextMarshaler(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface()
	buf = append(buf, fi.jkey...)
	if (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0 { // real nil check
		return buf, nil, aJustKey
	}
	return appendTextMarshalerVal(buf, v, safe)
}

func appendTextMarshalerAddr(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Addr().Interface()
	buf = append(buf, fi.jkey...)
	return appendTextMarshalerVal(buf, v, safe)
}

func appendTextMarshalerNotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface()
	if (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0 { // real nil check
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	return appendTextMarshalerVal(buf, v, safe)
}

func appendTextMarshalerVal(buf []byte, v any, safe bool) ([]byte, any, appendStatus) {
	m := v.(encoding.TextMarshaler)
	j, err := m.MarshalText()
	if err != nil {
		panic(err)
	}
	buf = ojg.AppendJSONString(buf, string(j), safe)

	return buf, nil, aWrote
}
