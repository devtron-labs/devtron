// Copyright (c) 2021, Peter Ohler, All rights reserved.

package oj

import (
	"encoding/json"
	"reflect"
	"unsafe"
)

func appendJSONMarshaler(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface()
	buf = append(buf, fi.jkey...)
	if (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0 { // real nil check
		return buf, nil, aJustKey
	}
	return appendJSONMarshalerVal(buf, v)
}

func appendJSONMarshalerAddr(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Addr().Interface()
	buf = append(buf, fi.jkey...)
	return appendJSONMarshalerVal(buf, v)
}

func appendJSONMarshalerNotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface()
	if (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0 { // real nil check
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	return appendJSONMarshalerVal(buf, v)
}

func appendJSONMarshalerVal(buf []byte, v any) ([]byte, any, appendStatus) {
	m := v.(json.Marshaler)
	j, err := m.MarshalJSON()
	if err != nil {
		panic(err)
	}
	buf = append(buf, j...)

	return buf, nil, aWrote
}
