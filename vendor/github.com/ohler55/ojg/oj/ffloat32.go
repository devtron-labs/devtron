// Copyright (c) 2021, Peter Ohler, All rights reserved.

package oj

import (
	"reflect"
	"strconv"
	"unsafe"
)

var float32AppendFuncs = [8]appendFunc{
	appendFloat32,
	appendFloat32AsString,
	appendFloat32NotEmpty,
	appendFloat32NotEmptyAsString,
	iappendFloat32,
	iappendFloat32AsString,
	iappendFloat32NotEmpty,
	iappendFloat32NotEmptyAsString,
}

func appendFloat32(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendFloat(buf, float64(*(*float32)(unsafe.Pointer(addr + fi.offset))), 'g', -1, 32)

	return buf, nil, aWrote
}

func appendFloat32AsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendFloat(buf, float64(*(*float32)(unsafe.Pointer(addr + fi.offset))), 'g', -1, 32)
	buf = append(buf, '"')

	return buf, nil, aWrote
}

func appendFloat32NotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := *(*float32)(unsafe.Pointer(addr + fi.offset))
	if v == 0.0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendFloat(buf, float64(v), 'g', -1, 32)

	return buf, nil, aWrote
}

func appendFloat32NotEmptyAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := *(*float32)(unsafe.Pointer(addr + fi.offset))
	if v == 0.0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendFloat(buf, float64(v), 'g', -1, 32)
	buf = append(buf, '"')

	return buf, nil, aWrote
}

func iappendFloat32(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendFloat(buf, float64(rv.FieldByIndex(fi.index).Interface().(float32)), 'g', -1, 32)

	return buf, nil, aWrote
}

func iappendFloat32AsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendFloat(buf, float64(rv.FieldByIndex(fi.index).Interface().(float32)), 'g', -1, 32)
	buf = append(buf, '"')

	return buf, nil, aWrote
}

func iappendFloat32NotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface().(float32)
	if v == 0.0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = strconv.AppendFloat(buf, float64(v), 'g', -1, 32)

	return buf, nil, aWrote
}

func iappendFloat32NotEmptyAsString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface().(float32)
	if v == 0.0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = append(buf, '"')
	buf = strconv.AppendFloat(buf, float64(v), 'g', -1, 32)
	buf = append(buf, '"')

	return buf, nil, aWrote
}
