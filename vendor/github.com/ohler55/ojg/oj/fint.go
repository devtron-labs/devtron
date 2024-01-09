// Copyright (c) 2021, Peter Ohler, All rights reserved.

package oj

import (
	"reflect"
	"strconv"
	"unsafe"
)

var intAppendFuncs = [8]appendFunc{
	appendInt,
	appendIntAsString,
	appendIntNotEmpty,
	appendIntNotEmptyAsString,
	iappendInt,
	iappendIntAsString,
	iappendIntNotEmpty,
	iappendIntNotEmptyAsString,
}

func appendInt(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendInt(buf, int64(*(*int)(unsafe.Pointer(addr + fi.offset))), 10)

	return buf, nil, aWrote
}

func appendIntAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendInt(buf, int64(*(*int)(unsafe.Pointer(addr + fi.offset))), 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}

func appendIntNotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := *(*int)(unsafe.Pointer(addr + fi.offset))
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendInt(buf, int64(v), 10)

	return buf, nil, aWrote
}

func appendIntNotEmptyAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := *(*int)(unsafe.Pointer(addr + fi.offset))
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendInt(buf, int64(v), 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}

func iappendInt(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendInt(buf, int64(rv.FieldByIndex(fi.index).Interface().(int)), 10)

	return buf, nil, aWrote
}

func iappendIntAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendInt(buf, int64(rv.FieldByIndex(fi.index).Interface().(int)), 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}

func iappendIntNotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface().(int)
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendInt(buf, int64(v), 10)

	return buf, nil, aWrote
}

func iappendIntNotEmptyAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface().(int)
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendInt(buf, int64(v), 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}
