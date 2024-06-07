// Copyright (c) 2021, Peter Ohler, All rights reserved.

package oj

import (
	"reflect"
	"strconv"
	"unsafe"
)

var uint8AppendFuncs = [8]appendFunc{
	appendUint8,
	appendUint8AsString,
	appendUint8NotEmpty,
	appendUint8NotEmptyAsString,
	iappendUint8,
	iappendUint8AsString,
	iappendUint8NotEmpty,
	iappendUint8NotEmptyAsString,
}

func appendUint8(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendUint(buf, uint64(*(*uint8)(unsafe.Pointer(addr + fi.offset))), 10)

	return buf, nil, aWrote
}

func appendUint8AsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendUint(buf, uint64(*(*uint8)(unsafe.Pointer(addr + fi.offset))), 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}

func appendUint8NotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := *(*uint8)(unsafe.Pointer(addr + fi.offset))
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendUint(buf, uint64(v), 10)

	return buf, nil, aWrote
}

func appendUint8NotEmptyAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := *(*uint8)(unsafe.Pointer(addr + fi.offset))
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendUint(buf, uint64(v), 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}

func iappendUint8(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendUint(buf, uint64(rv.FieldByIndex(fi.index).Interface().(uint8)), 10)

	return buf, nil, aWrote
}

func iappendUint8AsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendUint(buf, uint64(rv.FieldByIndex(fi.index).Interface().(uint8)), 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}

func iappendUint8NotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface().(uint8)
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendUint(buf, uint64(v), 10)

	return buf, nil, aWrote
}

func iappendUint8NotEmptyAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface().(uint8)
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendUint(buf, uint64(v), 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}
