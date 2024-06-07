// Copyright (c) 2021, Peter Ohler, All rights reserved.

package oj

import (
	"reflect"
	"unsafe"
)

var boolAppendFuncs = [8]appendFunc{
	appendBool,
	appendBoolAsString,
	appendBoolNotEmpty,
	appendBoolNotEmptyAsString,
	iappendBool,
	iappendBoolAsString,
	iappendBoolNotEmpty,
	iappendBoolNotEmptyAsString,
}

func appendBool(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	if *(*bool)(unsafe.Pointer(addr + fi.offset)) {
		buf = append(buf, "true"...)
	} else {
		buf = append(buf, "false"...)
	}
	return buf, nil, aWrote
}

func appendBoolAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	if *(*bool)(unsafe.Pointer(addr + fi.offset)) {
		buf = append(buf, `"true"`...)
	} else {
		buf = append(buf, `"false"`...)
	}
	return buf, nil, aWrote
}

func appendBoolNotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	if *(*bool)(unsafe.Pointer(addr + fi.offset)) {
		buf = append(buf, fi.jkey...)
		buf = append(buf, "true"...)
		return buf, nil, aWrote
	}
	return buf, nil, aSkip
}

func appendBoolNotEmptyAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	if *(*bool)(unsafe.Pointer(addr + fi.offset)) {
		buf = append(buf, fi.jkey...)
		buf = append(buf, `"true"`...)
		return buf, nil, aWrote
	}
	return buf, nil, aSkip
}

func iappendBool(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	if rv.FieldByIndex(fi.index).Interface().(bool) {
		buf = append(buf, "true"...)
	} else {
		buf = append(buf, "false"...)
	}
	return buf, nil, aWrote
}

func iappendBoolAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	if rv.FieldByIndex(fi.index).Interface().(bool) {
		buf = append(buf, `"true"`...)
	} else {
		buf = append(buf, `"false"`...)
	}
	return buf, nil, aWrote
}

func iappendBoolNotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	if rv.FieldByIndex(fi.index).Interface().(bool) {
		buf = append(buf, fi.jkey...)
		buf = append(buf, "true"...)
		return buf, nil, aWrote
	}
	return buf, nil, aSkip
}

func iappendBoolNotEmptyAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	if rv.FieldByIndex(fi.index).Interface().(bool) {
		buf = append(buf, fi.jkey...)
		buf = append(buf, `"true"`...)
		return buf, nil, aWrote
	}
	return buf, nil, aSkip
}
