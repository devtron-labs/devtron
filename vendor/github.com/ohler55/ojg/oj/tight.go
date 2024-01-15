// Copyright (c) 2020, Peter Ohler, All rights reserved.

package oj

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unsafe"

	"github.com/ohler55/ojg"
	"github.com/ohler55/ojg/alt"
)

func tightDefault(wr *Writer, data any, _ int) {
	switch {
	case !wr.NoReflect:
		rv := reflect.ValueOf(data)
		kind := rv.Kind()
		if kind == reflect.Ptr {
			rv = rv.Elem()
			kind = rv.Kind()
		}
		switch kind {
		case reflect.Struct:
			wr.tightStruct(rv, nil)
		case reflect.Slice, reflect.Array:
			wr.tightSlice(rv, nil)
		case reflect.Map:
			wr.tightMap(rv, nil)
		case reflect.Chan, reflect.Func, reflect.UnsafePointer:
			if wr.strict {
				panic(fmt.Errorf("%T can not be encoded as a JSON element", data))
			}
			wr.buf = append(wr.buf, "null"...)
		default:
			dec := alt.Decompose(data, &wr.Options)
			wr.appendJSON(dec, 0)
		}
	case wr.strict:
		panic(fmt.Errorf("%T can not be encoded as a JSON element", data))
	default:
		wr.buf = ojg.AppendJSONString(wr.buf, fmt.Sprintf("%v", data), !wr.HTMLUnsafe)
	}
}

func tightArray(wr *Writer, n []any, _ int) {
	if 0 < len(n) {
		wr.buf = append(wr.buf, '[')
		for _, m := range n {
			wr.appendJSON(m, 0)
			wr.buf = append(wr.buf, ',')
		}
		wr.buf[len(wr.buf)-1] = ']'
	} else {
		wr.buf = append(wr.buf, "[]"...)
	}
}

func tightObject(wr *Writer, n map[string]any, _ int) {
	comma := false
	wr.buf = append(wr.buf, '{')
	for k, m := range n {
		switch tm := m.(type) {
		case nil:
			if wr.OmitNil {
				continue
			}
		case string:
			if wr.OmitEmpty && len(tm) == 0 {
				continue
			}
		case map[string]any:
			if wr.OmitEmpty && len(tm) == 0 {
				continue
			}
		case []any:
			if wr.OmitEmpty && len(tm) == 0 {
				continue
			}
		}
		wr.buf = ojg.AppendJSONString(wr.buf, k, !wr.HTMLUnsafe)
		wr.buf = append(wr.buf, ':')
		wr.appendJSON(m, 0)
		wr.buf = append(wr.buf, ',')
		comma = true
	}
	if comma {
		wr.buf[len(wr.buf)-1] = '}'
	} else {
		wr.buf = append(wr.buf, '}')
	}
}

func tightSortObject(wr *Writer, n map[string]any, _ int) {
	comma := false
	wr.buf = append(wr.buf, '{')
	keys := make([]string, 0, len(n))
	for k := range n {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		m := n[k]
		switch tm := m.(type) {
		case nil:
			if wr.OmitNil {
				continue
			}
		case string:
			if wr.OmitEmpty && len(tm) == 0 {
				continue
			}
		case map[string]any:
			if wr.OmitEmpty && len(tm) == 0 {
				continue
			}
		case []any:
			if wr.OmitEmpty && len(tm) == 0 {
				continue
			}
		}
		wr.buf = ojg.AppendJSONString(wr.buf, k, !wr.HTMLUnsafe)
		wr.buf = append(wr.buf, ':')
		wr.appendJSON(m, 0)
		wr.buf = append(wr.buf, ',')
		comma = true
	}
	if comma {
		wr.buf[len(wr.buf)-1] = '}'
	} else {
		wr.buf = append(wr.buf, '}')
	}
}

func (wr *Writer) tightStruct(rv reflect.Value, si *sinfo) {
	if si == nil {
		si = getSinfo(rv.Interface(), wr.OmitEmpty)
	}
	fields := si.fields[wr.findex]
	wr.buf = append(wr.buf, '{')
	var v any
	comma := false
	if 0 < len(wr.CreateKey) {
		wr.buf = wr.appendString(wr.buf, wr.CreateKey, !wr.HTMLUnsafe)
		wr.buf = append(wr.buf, `:"`...)
		if wr.FullTypePath {
			wr.buf = append(wr.buf, si.rt.PkgPath()...)
			wr.buf = append(wr.buf, '/')
			wr.buf = append(wr.buf, si.rt.Name()...)
		} else {
			wr.buf = append(wr.buf, si.rt.Name()...)
		}
		wr.buf = append(wr.buf, `",`...)
		comma = true
	}
	var addr uintptr
	if rv.CanAddr() {
		addr = rv.UnsafeAddr()
	}
	var stat appendStatus
	for _, fi := range fields {
		if 0 < addr {
			wr.buf, v, stat = fi.Append(fi, wr.buf, rv, addr, !wr.HTMLUnsafe)
		} else {
			wr.buf, v, stat = fi.iAppend(fi, wr.buf, rv, addr, !wr.HTMLUnsafe)
		}
		switch stat {
		case aWrote:
			wr.buf = append(wr.buf, ',')
			comma = true
			continue
		case aSkip:
			continue
		case aChanged:
			if wr.OmitNil && (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0 {
				wr.buf = wr.buf[:len(wr.buf)-fi.keyLen()]
				continue
			}
			wr.appendJSON(v, 0)
			wr.buf = append(wr.buf, ',')
			comma = true
			continue
		}
		var fv reflect.Value
		kind := fi.kind
	Retry:
		switch kind {
		case reflect.Ptr:
			if (*[2]uintptr)(unsafe.Pointer(&v))[1] != 0 { // Check for nil of any type
				fv = reflect.ValueOf(v).Elem()
				kind = fv.Kind()
				v = fv.Interface()
				goto Retry
			}
			if wr.OmitNil {
				wr.buf = wr.buf[:len(wr.buf)-fi.keyLen()]
				continue
			}
			wr.buf = append(wr.buf, "null"...)
		case reflect.Interface:
			if wr.OmitNil && (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0 {
				wr.buf = wr.buf[:len(wr.buf)-fi.keyLen()]
				continue
			}
			wr.appendJSON(v, 0)
		case reflect.Struct:
			if !fv.IsValid() {
				fv = reflect.ValueOf(v)
			}
			wr.tightStruct(fv, fi.elem)
		case reflect.Slice, reflect.Array:
			if !fv.IsValid() {
				fv = reflect.ValueOf(v)
			}
			wr.tightSlice(fv, fi.elem)
		case reflect.Map:
			if !fv.IsValid() {
				fv = reflect.ValueOf(v)
			}
			wr.tightMap(fv, fi.elem)
		default:
			wr.appendJSON(v, 0)
		}
		wr.buf = append(wr.buf, ',')
		comma = true
	}
	if comma {
		wr.buf[len(wr.buf)-1] = '}'
	} else {
		wr.buf = append(wr.buf, '}')
	}
}

func (wr *Writer) tightSlice(rv reflect.Value, si *sinfo) {
	end := rv.Len()
	comma := false
	wr.buf = append(wr.buf, '[')
	for j := 0; j < end; j++ {
		rm := rv.Index(j)
		if rm.Kind() == reflect.Ptr {
			rm = rm.Elem()
		}
		switch rm.Kind() {
		case reflect.Struct:
			wr.tightStruct(rm, si)
		case reflect.Slice, reflect.Array:
			wr.tightSlice(rm, si)
		case reflect.Map:
			wr.tightMap(rm, si)
		default:
			wr.appendJSON(rm.Interface(), 0)
		}
		wr.buf = append(wr.buf, ',')
		comma = true
	}
	if comma {
		wr.buf[len(wr.buf)-1] = ']'
	} else {
		wr.buf = append(wr.buf, ']')
	}
}

func (wr *Writer) tightMap(rv reflect.Value, si *sinfo) {
	wr.buf = append(wr.buf, '{')
	keys := rv.MapKeys()
	if wr.Sort {
		sort.Slice(keys, func(i, j int) bool { return 0 > strings.Compare(keys[i].String(), keys[j].String()) })
	}
	comma := false
	for _, kv := range keys {
		rm := rv.MapIndex(kv)
		if rm.Kind() == reflect.Ptr {
			if wr.OmitNil && rm.IsNil() {
				continue
			}
			rm = rm.Elem()
		}
		switch rm.Kind() {
		case reflect.Struct:
			wr.buf = ojg.AppendJSONString(wr.buf, kv.String(), !wr.HTMLUnsafe)
			wr.buf = append(wr.buf, ':')
			wr.tightStruct(rm, si)
		case reflect.Slice, reflect.Array:
			if (wr.OmitNil || wr.OmitEmpty) && rm.Len() == 0 {
				continue
			}
			wr.buf = ojg.AppendJSONString(wr.buf, kv.String(), !wr.HTMLUnsafe)
			wr.buf = append(wr.buf, ':')
			wr.tightSlice(rm, si)
		case reflect.Map:
			if (wr.OmitNil || wr.OmitEmpty) && rm.Len() == 0 {
				continue
			}
			wr.buf = ojg.AppendJSONString(wr.buf, kv.String(), !wr.HTMLUnsafe)
			wr.buf = append(wr.buf, ':')
			wr.tightMap(rm, si)
		case reflect.String:
			if (wr.OmitNil || wr.OmitEmpty) && rm.Len() == 0 {
				continue
			}
			wr.buf = ojg.AppendJSONString(wr.buf, kv.String(), !wr.HTMLUnsafe)
			wr.buf = append(wr.buf, ':')
			wr.appendJSON(rm.Interface(), 0)
		default:
			wr.buf = ojg.AppendJSONString(wr.buf, kv.String(), !wr.HTMLUnsafe)
			wr.buf = append(wr.buf, ':')
			wr.appendJSON(rm.Interface(), 0)
		}
		wr.buf = append(wr.buf, ',')
		comma = true
	}
	if comma {
		wr.buf[len(wr.buf)-1] = '}'
	} else {
		wr.buf = append(wr.buf, '}')
	}
}
