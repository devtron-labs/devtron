// Copyright (c) 2021, Peter Ohler, All rights reserved.

package oj

import (
	"reflect"
	"strconv"
	"unsafe"
)

var uintAppendFuncs = [8]appendFunc{
	appendUint,
	appendUintAsString,
	appendUintNotEmpty,
	appendUintNotEmptyAsString,
	iappendUint,
	iappendUintAsString,
	iappendUintNotEmpty,
	iappendUintNotEmptyAsString,
}

func appendUint(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendUint(buf, uint64(*(*uint)(unsafe.Pointer(addr + fi.offset))), 10)

	return buf, nil, aWrote
}

func appendUintAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendUint(buf, uint64(*(*uint)(unsafe.Pointer(addr + fi.offset))), 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}

func appendUintNotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := *(*uint)(unsafe.Pointer(addr + fi.offset))
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendUint(buf, uint64(v), 10)

	return buf, nil, aWrote
}

func appendUintNotEmptyAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := *(*uint)(unsafe.Pointer(addr + fi.offset))
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendUint(buf, uint64(v), 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}

func iappendUint(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendUint(buf, uint64(rv.FieldByIndex(fi.index).Interface().(uint)), 10)

	return buf, nil, aWrote
}

func iappendUintAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendUint(buf, uint64(rv.FieldByIndex(fi.index).Interface().(uint)), 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}

func iappendUintNotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface().(uint)
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendUint(buf, uint64(v), 10)

	return buf, nil, aWrote
}

func iappendUintNotEmptyAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface().(uint)
	if v == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendUint(buf, uint64(v), 10)
	buf = append(buf, '"')

	return buf, nil, aWrote
}
