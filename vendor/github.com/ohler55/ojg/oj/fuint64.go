// Copyright (c) 2021, Peter Ohler, All rights reserved.

package oj

import (
	"reflect"
	"strconv"
	"unsafe"
)

var uint64AppendFuncs = [8]appendFunc{
	appendUint64,
	appendUint64AsString,
	appendUint64NotEmpty,
	appendUint64NotEmptyAsString,
	iappendUint64,
	iappendUint64AsString,
	iappendUint64NotEmpty,
	iappendUint64NotEmptyAsString,
}

func appendUint64(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendUint(buf, *(*uint64)(unsafe.Pointer(addr + fi.offset)), 10)

	return buf, nil, aWrote
}

func appendUint64AsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendUint(buf, *(*uint64)(unsafe.Pointer(addr + fi.offset)), 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}

func appendUint64NotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := *(*uint64)(unsafe.Pointer(addr + fi.offset))
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendUint(buf, v, 10)

	return buf, nil, aWrote
}

func appendUint64NotEmptyAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := *(*uint64)(unsafe.Pointer(addr + fi.offset))
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendUint(buf, v, 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}

func iappendUint64(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendUint(buf, rv.FieldByIndex(fi.index).Interface().(uint64), 10)

	return buf, nil, aWrote
}

func iappendUint64AsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendUint(buf, rv.FieldByIndex(fi.index).Interface().(uint64), 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}

func iappendUint64NotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface().(uint64)
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendUint(buf, v, 10)

	return buf, nil, aWrote
}

func iappendUint64NotEmptyAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface().(uint64)
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendUint(buf, v, 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}
