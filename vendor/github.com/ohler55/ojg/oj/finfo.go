// Copyright (c) 2021, Peter Ohler, All rights reserved.

package oj

import (
	"encoding"
	"encoding/json"
	"reflect"
	"unsafe"

	"github.com/ohler55/ojg"
	"github.com/ohler55/ojg/alt"
)

const (
	strMask   = byte(0x01)
	omitMask  = byte(0x02)
	embedMask = byte(0x04)

	aJustKey appendStatus = iota
	aWrote
	aSkip
	aChanged
)

type appendStatus byte

type appendFunc func(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus)

// Field hold information about a struct field.
type finfo struct {
	rt      reflect.Type
	key     string
	kind    reflect.Kind
	elem    *sinfo
	Append  appendFunc
	iAppend appendFunc
	jkey    []byte
	index   []int
	offset  uintptr
}

func (f *finfo) keyLen() int {
	return len(f.jkey)
}

func appendString(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).String()
	buf = append(buf, fi.jkey...)
	buf = ojg.AppendJSONString(buf, v, safe)

	return buf, nil, aWrote
}

func appendStringNotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	s := rv.FieldByIndex(fi.index).String()
	if len(s) == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	buf = ojg.AppendJSONString(buf, s, safe)

	return buf, nil, aWrote
}

func appendJustKey(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface()
	buf = append(buf, fi.jkey...)
	return buf, v, aJustKey
}

func appendPtrNotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	v := rv.FieldByIndex(fi.index).Interface()
	if (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0 { // real nil check
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	return buf, v, aJustKey
}

func appendSliceNotEmpty(fi *finfo, buf []byte, rv reflect.Value, addr uintptr, safe bool) ([]byte, any, appendStatus) {
	fv := rv.FieldByIndex(fi.index)
	if fv.Len() == 0 {
		return buf, nil, aSkip
	}
	buf = append(buf, fi.jkey...)
	return buf, fv.Interface(), aJustKey
}

func whichAppend(rt reflect.Type, omitEmpty bool) (f appendFunc, af appendFunc) {
	v := reflect.New(rt).Elem().Interface()
	switch v.(type) {
	case json.Marshaler:
		if omitEmpty {
			f = appendJSONMarshalerNotEmpty
		} else {
			f = appendJSONMarshaler
		}
	case encoding.TextMarshaler:
		if omitEmpty {
			f = appendTextMarshalerNotEmpty
		} else {
			f = appendTextMarshaler
		}
	case alt.Simplifier:
		if omitEmpty {
			f = appendSimplifierNotEmpty
		} else {
			f = appendSimplifier
		}
	case alt.Genericer:
		if omitEmpty {
			f = appendGenericerNotEmpty
		} else {
			f = appendGenericer
		}
	}
	vp := reflect.New(rt).Interface()
	switch vp.(type) {
	case json.Marshaler:
		af = appendJSONMarshalerAddr
	case encoding.TextMarshaler:
		af = appendTextMarshalerAddr
	case alt.Simplifier:
		af = appendSimplifierAddr
	case alt.Genericer:
		af = appendGenericerAddr
	}
	return
}

func newFinfo(f *reflect.StructField, key string, omitEmpty, asString, pretty, embedded bool) *finfo {
	fi := finfo{
		rt:     f.Type,
		key:    key,
		kind:   f.Type.Kind(),
		index:  f.Index,
		offset: f.Offset,
	}
	var fx byte
	// Check for interfaces first since almost any type can implement one of
	// the supported interfaces.
	ff, af := whichAppend(fi.rt, omitEmpty)
	if ff != nil && af != nil {
		fi.Append = ff
		fi.iAppend = ff
		goto Key
	}
	if omitEmpty {
		fx |= omitMask
	}
	if asString {
		fx |= strMask
	}
	if embedded {
		fx |= embedMask
	}
	switch fi.kind {
	case reflect.Bool:
		fi.Append = boolAppendFuncs[fx]
		fi.iAppend = boolAppendFuncs[fx|embedMask]

	case reflect.Int:
		fi.Append = intAppendFuncs[fx]
		fi.iAppend = intAppendFuncs[fx|embedMask]
	case reflect.Int8:
		fi.Append = int8AppendFuncs[fx]
		fi.iAppend = int8AppendFuncs[fx|embedMask]
	case reflect.Int16:
		fi.Append = int16AppendFuncs[fx]
		fi.iAppend = int16AppendFuncs[fx|embedMask]
	case reflect.Int32:
		fi.Append = int32AppendFuncs[fx]
		fi.iAppend = int32AppendFuncs[fx|embedMask]
	case reflect.Int64:
		fi.Append = int64AppendFuncs[fx]
		fi.iAppend = int64AppendFuncs[fx|embedMask]

	case reflect.Uint:
		fi.Append = uintAppendFuncs[fx]
		fi.iAppend = uintAppendFuncs[fx|embedMask]
	case reflect.Uint8:
		fi.Append = uint8AppendFuncs[fx]
		fi.iAppend = uint8AppendFuncs[fx|embedMask]
	case reflect.Uint16:
		fi.Append = uint16AppendFuncs[fx]
		fi.iAppend = uint16AppendFuncs[fx|embedMask]
	case reflect.Uint32:
		fi.Append = uint32AppendFuncs[fx]
		fi.iAppend = uint32AppendFuncs[fx|embedMask]
	case reflect.Uint64:
		fi.Append = uint64AppendFuncs[fx]
		fi.iAppend = uint64AppendFuncs[fx|embedMask]

	case reflect.Float32:
		fi.Append = float32AppendFuncs[fx]
		fi.iAppend = float32AppendFuncs[fx|embedMask]
	case reflect.Float64:
		fi.Append = float64AppendFuncs[fx]
		fi.iAppend = float64AppendFuncs[fx|embedMask]

	case reflect.String:
		if omitEmpty {
			fi.Append = appendStringNotEmpty
			fi.iAppend = appendStringNotEmpty
		} else {
			fi.Append = appendString
			fi.iAppend = appendString
		}
	case reflect.Struct:
		fi.elem = getTypeStruct(fi.rt, true, omitEmpty)
		fi.Append = appendJustKey
		fi.iAppend = appendJustKey
	case reflect.Ptr:
		et := fi.rt.Elem()
		if et.Kind() == reflect.Ptr {
			et = et.Elem()
		}
		if et.Kind() == reflect.Struct {
			fi.elem = getTypeStruct(et, false, omitEmpty)
		}
		if omitEmpty {
			fi.Append = appendPtrNotEmpty
			fi.iAppend = appendPtrNotEmpty
		} else {
			fi.Append = appendJustKey
			fi.iAppend = appendJustKey
		}
	case reflect.Interface:
		if omitEmpty {
			fi.Append = appendPtrNotEmpty
			fi.iAppend = appendPtrNotEmpty
		} else {
			fi.Append = appendJustKey
			fi.iAppend = appendJustKey
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		et := fi.rt.Elem()
		embedded := true
		if et.Kind() == reflect.Ptr {
			embedded = false
			et = et.Elem()
		}
		if et.Kind() == reflect.Struct {
			fi.elem = getTypeStruct(et, embedded, omitEmpty)
		}
		if omitEmpty {
			fi.Append = appendSliceNotEmpty
			fi.iAppend = appendSliceNotEmpty
		} else {
			fi.Append = appendJustKey
			fi.iAppend = appendJustKey
		}
	}
	if ff != nil { // override
		fi.iAppend = ff
		fi.Append = ff
	}
	if af != nil { // override
		fi.Append = af
	}
Key:
	fi.jkey = ojg.AppendJSONString(fi.jkey, fi.key, false)
	fi.jkey = append(fi.jkey, ':')
	if pretty {
		fi.jkey = append(fi.jkey, ' ')
	}
	return &fi
}
