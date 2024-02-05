// Copyright (c) 2021, Peter Ohler, All rights reserved.

package oj

import (
	"reflect"
	"strconv"
	"unsafe"
)

var int8AppendFuncs = [8]appendFunc{
	appendInt8,
	appendInt8AsString,
	appendInt8NotEmpty,
	appendInt8NotEmptyAsString,
	iappendInt8,
	iappendInt8AsString,
	iappendInt8NotEmpty,
	iappendInt8NotEmptyAsString,
}

func appendInt8(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendInt(buf, int64(*(*int8)(unsafe.Pointer(addr + fi.offset))), 10)

	return buf, nil, aWrote
}

func appendInt8AsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendInt(buf, int64(*(*int8)(unsafe.Pointer(addr + fi.offset))), 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}

func appendInt8NotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := *(*int8)(unsafe.Pointer(addr + fi.offset))
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendInt(buf, int64(v), 10)

	return buf, nil, aWrote
}

func appendInt8NotEmptyAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := *(*int8)(unsafe.Pointer(addr + fi.offset))
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendInt(buf, int64(v), 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}

func iappendInt8(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendInt(buf, int64(rv.FieldByIndex(fi.index).Interface().(int8)), 10)

	return buf, nil, aWrote
}

func iappendInt8AsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendInt(buf, int64(rv.FieldByIndex(fi.index).Interface().(int8)), 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}

func iappendInt8NotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface().(int8)
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendInt(buf, int64(v), 10)

	return buf, nil, aWrote
}

func iappendInt8NotEmptyAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface().(int8)
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendInt(buf, int64(v), 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}
