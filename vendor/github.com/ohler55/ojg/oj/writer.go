// Copyright (c) 2020, Peter Ohler, All rights reserved.

package oj

import (
	"encoding"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/ohler55/ojg"
	"github.com/ohler55/ojg/alt"
)

const (
	spaces = "\n                                                                " +
		"                                                                "
	tabs = "\n\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t"
)

// Writer is a JSON writer that includes a reused buffer for reduced
// allocations for repeated encoding calls.
type Writer struct {
	ojg.Options
	buf           []byte
	w             io.Writer
	findex        byte
	strict        bool
	appendArray   func(wr *Writer, data []any, depth int)
	appendObject  func(wr *Writer, data map[string]any, depth int)
	appendDefault func(wr *Writer, data any, depth int)
	appendString  func(buf []byte, s string, htmlSafe bool) []byte
}

// JSON writes data, JSON encoded. On error, an empty string is returned.
func (wr *Writer) JSON(data any) string {
	defer func() {
		if r := recover(); r != nil {
			wr.buf = wr.buf[:0]
		}
	}()
	return string(wr.MustJSON(data))
}

// MustJSON writes data, JSON encoded as a []byte and not a string like the
// JSON() function. On error a panic is called with the error. The returned
// buffer is the Writer buffer and is reused on the next call to write. If
// returned value is to be preserved past a second invocation then the buffer
// should be copied.
func (wr *Writer) MustJSON(data any) []byte {
	wr.w = nil
	if wr.InitSize <= 0 {
		wr.InitSize = 256
	}
	if cap(wr.buf) < wr.InitSize {
		wr.buf = make([]byte, 0, wr.InitSize)
	} else {
		wr.buf = wr.buf[:0]
	}
	wr.calcFieldsIndex()
	if wr.Color {
		wr.colorJSON(data, 0)
	} else {
		wr.appendString = ojg.AppendJSONString
		if wr.Tab || 0 < wr.Indent {
			wr.appendArray = appendArray
			if wr.Sort {
				wr.appendObject = appendSortObject
			} else {
				wr.appendObject = appendObject
			}
			wr.appendDefault = appendDefault
		} else {
			wr.appendArray = tightArray
			if wr.Sort {
				wr.appendObject = tightSortObject
			} else {
				wr.appendObject = tightObject
			}
			wr.appendDefault = tightDefault
		}
		wr.appendJSON(data, 0)
	}
	return wr.buf
}

// Write a JSON string for the data provided.
func (wr *Writer) Write(w io.Writer, data any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			wr.buf = wr.buf[:0]
			err = ojg.NewError(r)
		}
	}()
	wr.MustWrite(w, data)
	return
}

// MustWrite a JSON string for the data provided. If an error occurs panic is
// called with the error.
func (wr *Writer) MustWrite(w io.Writer, data any) {
	wr.w = w
	if wr.InitSize <= 0 {
		wr.InitSize = 256
	}
	if wr.WriteLimit <= 0 {
		wr.WriteLimit = 1024
	}
	if cap(wr.buf) < wr.InitSize {
		wr.buf = make([]byte, 0, wr.InitSize)
	} else {
		wr.buf = wr.buf[:0]
	}
	wr.calcFieldsIndex()
	if wr.Color {
		wr.colorJSON(data, 0)
	} else {
		wr.appendString = ojg.AppendJSONString
		if wr.Tab || 0 < wr.Indent {
			wr.appendArray = appendArray
			if wr.Sort {
				wr.appendObject = appendSortObject
			} else {
				wr.appendObject = appendObject
			}
			wr.appendDefault = appendDefault
		} else {
			wr.appendArray = tightArray
			if wr.Sort {
				wr.appendObject = tightSortObject
			} else {
				wr.appendObject = tightObject
			}
			wr.appendDefault = tightDefault
		}
		wr.appendJSON(data, 0)
	}
	if 0 < len(wr.buf) {
		if _, err := wr.w.Write(wr.buf); err != nil {
			panic(err)
		}
	}
}

func (wr *Writer) calcFieldsIndex() {
	wr.findex = 0
	if wr.NestEmbed {
		wr.findex |= maskNested
	}
	if 0 < wr.Indent {
		wr.findex |= maskPretty
	}
	if wr.UseTags {
		wr.findex |= maskByTag
	} else if wr.KeyExact {
		wr.findex |= maskExact
	}
}

func (wr *Writer) appendJSON(data any, depth int) {
	switch td := data.(type) {
	case nil:
		wr.buf = append(wr.buf, "null"...)

	case bool:
		if td {
			wr.buf = append(wr.buf, "true"...)
		} else {
			wr.buf = append(wr.buf, "false"...)
		}

	case int:
		wr.buf = strconv.AppendInt(wr.buf, int64(td), 10)
	case int8:
		wr.buf = strconv.AppendInt(wr.buf, int64(td), 10)
	case int16:
		wr.buf = strconv.AppendInt(wr.buf, int64(td), 10)
	case int32:
		wr.buf = strconv.AppendInt(wr.buf, int64(td), 10)
	case int64:
		wr.buf = strconv.AppendInt(wr.buf, td, 10)
	case uint:
		wr.buf = strconv.AppendUint(wr.buf, uint64(td), 10)
	case uint8:
		wr.buf = strconv.AppendUint(wr.buf, uint64(td), 10)
	case uint16:
		wr.buf = strconv.AppendUint(wr.buf, uint64(td), 10)
	case uint32:
		wr.buf = strconv.AppendUint(wr.buf, uint64(td), 10)
	case uint64:
		wr.buf = strconv.AppendUint(wr.buf, td, 10)

	case float32:
		if 0 < len(wr.FloatFormat) {
			wr.buf = fmt.Appendf(wr.buf, wr.FloatFormat, float64(td))
		} else {
			wr.buf = strconv.AppendFloat(wr.buf, float64(td), 'g', -1, 32)
		}
	case float64:
		if 0 < len(wr.FloatFormat) {
			wr.buf = fmt.Appendf(wr.buf, wr.FloatFormat, td)
		} else {
			wr.buf = strconv.AppendFloat(wr.buf, td, 'g', -1, 64)
		}

	case string:
		wr.buf = wr.appendString(wr.buf, td, !wr.HTMLUnsafe)

	case []byte:
		switch wr.BytesAs {
		case ojg.BytesAsBase64:
			wr.buf = wr.appendString(wr.buf, base64.StdEncoding.EncodeToString(td), !wr.HTMLUnsafe)
		case ojg.BytesAsArray:
			a := make([]any, len(td))
			for i, m := range td {
				a[i] = int64(m)
			}
			wr.appendArray(wr, a, depth)
		default:
			wr.buf = wr.appendString(wr.buf, string(td), !wr.HTMLUnsafe)
		}

	case time.Time:
		wr.buf = wr.AppendTime(wr.buf, td, false)

	case []any:
		// go marshal treats a nil slice as a special case different from an
		// empty slice. Seems kind of odd but here is the check.
		if wr.strict && td == nil {
			wr.buf = append(wr.buf, "null"...)
			break
		}
		wr.appendArray(wr, td, depth)

	case map[string]any:
		wr.appendObject(wr, td, depth)

	case alt.Simplifier:
		wr.appendJSON(td.Simplify(), depth)
	case alt.Genericer:
		wr.appendJSON(td.Generic().Simplify(), depth)
	case json.Marshaler:
		out, err := td.MarshalJSON()
		if err != nil {
			panic(err)
		}
		wr.buf = append(wr.buf, out...)
	case encoding.TextMarshaler:
		out, err := td.MarshalText()
		if err != nil {
			panic(err)
		}
		wr.buf = wr.appendString(wr.buf, string(out), !wr.HTMLUnsafe)

	default:
		wr.appendDefault(wr, data, depth)
	}
	if wr.w != nil && wr.WriteLimit < len(wr.buf) {
		if _, err := wr.w.Write(wr.buf); err != nil {
			panic(err)
		}
		wr.buf = wr.buf[:0]
	}
}

func appendDefault(wr *Writer, data any, depth int) {
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
			wr.appendStruct(rv, depth, nil)
		case reflect.Slice, reflect.Array:
			wr.appendSlice(rv, depth, nil)
		case reflect.Map:
			wr.appendMap(rv, depth, nil)
		case reflect.Chan, reflect.Func, reflect.UnsafePointer:
			if wr.strict {
				panic(fmt.Errorf("%T can not be encoded as a JSON element", data))
			}
			wr.buf = append(wr.buf, "null"...)
		default:
			dec := alt.Decompose(data, &wr.Options)
			wr.appendJSON(dec, depth)
		}
	case wr.strict:
		panic(fmt.Errorf("%T can not be encoded as a JSON element", data))
	default:
		wr.buf = wr.appendString(wr.buf, fmt.Sprintf("%v", data), !wr.HTMLUnsafe)
	}
}

func appendArray(wr *Writer, n []any, depth int) {
	var is string
	var cs string
	d2 := depth + 1
	if wr.Tab {
		x := depth + 1
		if len(tabs) < x {
			x = len(tabs)
		}
		is = tabs[1:x]
		x = d2 + 1
		if len(tabs) < x {
			x = len(tabs)
		}
		cs = tabs[0:x]
	} else {
		x := depth*wr.Indent + 1
		if len(spaces) < x {
			x = len(spaces)
		}
		is = spaces[1:x]
		x = d2*wr.Indent + 1
		if len(spaces) < x {
			x = len(spaces)
		}
		cs = spaces[0:x]
	}
	if 0 < len(n) {
		wr.buf = append(wr.buf, '[')
		for _, m := range n {
			wr.buf = append(wr.buf, cs...)
			wr.appendJSON(m, d2)
			wr.buf = append(wr.buf, ',')
		}
		wr.buf[len(wr.buf)-1] = '\n'
		wr.buf = append(wr.buf, is...)
		wr.buf = append(wr.buf, ']')
	} else {
		wr.buf = append(wr.buf, "[]"...)
	}
}

func appendObject(wr *Writer, n map[string]any, depth int) {
	d2 := depth + 1
	var is string
	var cs string
	if wr.Tab {
		x := depth + 1
		if len(tabs) < x {
			x = len(tabs)
		}
		is = tabs[1:x]
		x = d2 + 1
		if len(tabs) < x {
			x = len(tabs)
		}
		cs = tabs[0:x]
	} else {
		x := depth*wr.Indent + 1
		if len(spaces) < x {
			x = len(spaces)
		}
		is = spaces[1:x]
		x = d2*wr.Indent + 1
		if len(spaces) < x {
			x = len(spaces)
		}
		cs = spaces[0:x]
	}
	empty := true
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
		empty = false
		wr.buf = append(wr.buf, cs...)
		wr.buf = wr.appendString(wr.buf, k, !wr.HTMLUnsafe)
		wr.buf = append(wr.buf, ':')
		wr.buf = append(wr.buf, ' ')
		wr.appendJSON(m, d2)
		wr.buf = append(wr.buf, ',')
	}
	if !empty {
		wr.buf[len(wr.buf)-1] = '\n'
		wr.buf = append(wr.buf, is...)
	}
	wr.buf = append(wr.buf, '}')
}

func appendSortObject(wr *Writer, n map[string]any, depth int) {
	d2 := depth + 1
	var is string
	var cs string
	if wr.Tab {
		x := depth + 1
		if len(tabs) < x {
			x = len(tabs)
		}
		is = tabs[1:x]
		x = d2 + 1
		if len(tabs) < x {
			x = len(tabs)
		}
		cs = tabs[0:x]
	} else {
		x := depth*wr.Indent + 1
		if len(spaces) < x {
			x = len(spaces)
		}
		is = spaces[1:x]
		x = d2*wr.Indent + 1
		if len(spaces) < x {
			x = len(spaces)
		}
		cs = spaces[0:x]
	}
	keys := make([]string, 0, len(n))
	for k := range n {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	empty := true
	wr.buf = append(wr.buf, '{')
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
		empty = false
		wr.buf = append(wr.buf, cs...)
		wr.buf = wr.appendString(wr.buf, k, !wr.HTMLUnsafe)
		wr.buf = append(wr.buf, ':')
		wr.buf = append(wr.buf, ' ')
		wr.appendJSON(m, d2)
		wr.buf = append(wr.buf, ',')
	}
	if !empty {
		wr.buf[len(wr.buf)-1] = '\n'
		wr.buf = append(wr.buf, is...)
	}
	wr.buf = append(wr.buf, '}')
}

func (wr *Writer) appendStruct(rv reflect.Value, depth int, si *sinfo) {
	if si == nil {
		si = getSinfo(rv.Interface(), wr.OmitEmpty)
	}
	d2 := depth + 1
	fields := si.fields[wr.findex]
	wr.buf = append(wr.buf, '{')
	empty := true
	var v any
	indented := false
	var is string
	var cs string
	if wr.Tab {
		x := depth + 1
		if len(tabs) < x {
			x = len(tabs)
		}
		is = tabs[1:x]
		x = d2 + 1
		if len(tabs) < x {
			x = len(tabs)
		}
		cs = tabs[0:x]
	} else {
		x := depth*wr.Indent + 1
		if len(spaces) < x {
			x = len(spaces)
		}
		is = spaces[1:x]
		x = d2*wr.Indent + 1
		if len(spaces) < x {
			x = len(spaces)
		}
		cs = spaces[0:x]
	}
	if 0 < len(wr.CreateKey) {
		wr.buf = append(wr.buf, cs...)
		wr.buf = append(wr.buf, '"')
		wr.buf = append(wr.buf, wr.CreateKey...)
		wr.buf = append(wr.buf, `": "`...)
		if wr.FullTypePath {
			wr.buf = append(wr.buf, si.rt.PkgPath()...)
			wr.buf = append(wr.buf, '/')
			wr.buf = append(wr.buf, si.rt.Name()...)
		} else {
			wr.buf = append(wr.buf, si.rt.Name()...)
		}
		wr.buf = append(wr.buf, `",`...)
		empty = false
	}
	var addr uintptr
	if rv.CanAddr() {
		addr = rv.UnsafeAddr()
	}
	var stat appendStatus
	for _, fi := range fields {
		if !indented {
			wr.buf = append(wr.buf, cs...)
			indented = true
		}
		if 0 < addr {
			wr.buf, v, stat = fi.Append(fi, wr.buf, rv, addr, !wr.HTMLUnsafe)
		} else {
			wr.buf, v, stat = fi.iAppend(fi, wr.buf, rv, addr, !wr.HTMLUnsafe)
		}
		switch stat {
		case aWrote:
			wr.buf = append(wr.buf, ',')
			empty = false
			indented = false
			continue
		case aSkip:
			continue
		case aChanged:
			if wr.OmitNil && (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0 {
				wr.buf = wr.buf[:len(wr.buf)-fi.keyLen()]
				continue
			}
			wr.appendJSON(v, d2)
			wr.buf = append(wr.buf, ',')
			indented = false
			empty = false
			continue
		}
		indented = false
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
				indented = true
				continue
			}
			wr.buf = append(wr.buf, "null"...)
		case reflect.Interface:
			if wr.OmitNil && (*[2]uintptr)(unsafe.Pointer(&v))[1] == 0 {
				wr.buf = wr.buf[:len(wr.buf)-fi.keyLen()]
				indented = true
				continue
			}
			wr.appendJSON(v, 0)
		case reflect.Struct:
			if !fv.IsValid() {
				fv = reflect.ValueOf(v)
			}
			wr.appendStruct(fv, d2, fi.elem)
		case reflect.Slice, reflect.Array:
			if !fv.IsValid() {
				fv = reflect.ValueOf(v)
			}
			wr.appendSlice(fv, d2, fi.elem)
		case reflect.Map:
			if !fv.IsValid() {
				fv = reflect.ValueOf(v)
			}
			wr.appendMap(fv, d2, fi.elem)
		default:
			wr.appendJSON(v, d2)
		}
		wr.buf = append(wr.buf, ',')
		empty = false
	}
	if indented {
		wr.buf = wr.buf[:len(wr.buf)-len(cs)]
	}
	if !empty {
		wr.buf[len(wr.buf)-1] = '\n'
		wr.buf = append(wr.buf, is...)
	}
	wr.buf = append(wr.buf, '}')
}

func (wr *Writer) appendSlice(rv reflect.Value, depth int, si *sinfo) {
	end := rv.Len()
	if end == 0 {
		wr.buf = append(wr.buf, "[]"...)
		return
	}
	d2 := depth + 1
	var is string
	var cs string
	if wr.Tab {
		x := depth + 1
		if len(tabs) < x {
			x = len(tabs)
		}
		is = tabs[1:x]
		x = d2 + 1
		if len(tabs) < x {
			x = len(tabs)
		}
		cs = tabs[0:x]
	} else {
		x := depth*wr.Indent + 1
		if len(spaces) < x {
			x = len(spaces)
		}
		is = spaces[1:x]
		x = d2*wr.Indent + 1
		if len(spaces) < x {
			x = len(spaces)
		}
		cs = spaces[0:x]
	}
	wr.buf = append(wr.buf, '[')
	for j := 0; j < end; j++ {
		wr.buf = append(wr.buf, cs...)
		rm := rv.Index(j)
		switch rm.Kind() {
		case reflect.Struct:
			wr.appendStruct(rm, d2, si)
		case reflect.Slice, reflect.Array:
			wr.appendSlice(rm, d2, si)
		case reflect.Map:
			wr.appendMap(rm, d2, si)
		default:
			wr.appendJSON(rm.Interface(), d2)
		}
		wr.buf = append(wr.buf, ',')
	}
	wr.buf[len(wr.buf)-1] = '\n'
	wr.buf = append(wr.buf, is...)
	wr.buf = append(wr.buf, ']')
}

func (wr *Writer) appendMap(rv reflect.Value, depth int, si *sinfo) {
	keys := rv.MapKeys()
	if wr.Sort {
		sort.Slice(keys, func(i, j int) bool { return 0 > strings.Compare(keys[i].String(), keys[j].String()) })
	}
	d2 := depth + 1
	var is string
	var cs string
	if wr.Tab {
		x := depth + 1
		if len(tabs) < x {
			x = len(tabs)
		}
		is = tabs[1:x]
		x = d2 + 1
		if len(tabs) < x {
			x = len(tabs)
		}
		cs = tabs[0:x]
	} else {
		x := depth*wr.Indent + 1
		if len(spaces) < x {
			x = len(spaces)
		}
		is = spaces[1:x]
		x = d2*wr.Indent + 1
		if len(spaces) < x {
			x = len(spaces)
		}
		cs = spaces[0:x]
	}
	empty := true
	wr.buf = append(wr.buf, '{')
	for _, kv := range keys {
		rm := rv.MapIndex(kv)
		if rm.Kind() == reflect.Ptr {
			if rm.IsNil() {
				if wr.OmitNil {
					continue
				}
			} else {
				rm = rm.Elem()
			}
		}
		switch rm.Kind() {
		case reflect.Struct:
			wr.buf = append(wr.buf, cs...)
			wr.buf = wr.appendString(wr.buf, kv.String(), !wr.HTMLUnsafe)
			wr.buf = append(wr.buf, ": "...)
			wr.appendStruct(rm, d2, si)
		case reflect.Slice, reflect.Array:
			if (wr.OmitNil || wr.OmitEmpty) && rm.Len() == 0 {
				continue
			}
			wr.buf = append(wr.buf, cs...)
			wr.buf = wr.appendString(wr.buf, kv.String(), !wr.HTMLUnsafe)
			wr.buf = append(wr.buf, ": "...)
			wr.appendSlice(rm, d2, si)
		case reflect.Map:
			if (wr.OmitNil || wr.OmitEmpty) && rm.Len() == 0 {
				continue
			}
			wr.buf = append(wr.buf, cs...)
			wr.buf = wr.appendString(wr.buf, kv.String(), !wr.HTMLUnsafe)
			wr.buf = append(wr.buf, ": "...)
			wr.appendMap(rm, d2, si)
		case reflect.String:
			if (wr.OmitEmpty) && rm.Len() == 0 {
				continue
			}
			wr.buf = append(wr.buf, cs...)
			wr.buf = wr.appendString(wr.buf, kv.String(), !wr.HTMLUnsafe)
			wr.buf = append(wr.buf, ": "...)
			wr.appendJSON(rm.Interface(), d2)
		default:
			wr.buf = append(wr.buf, cs...)
			wr.buf = wr.appendString(wr.buf, kv.String(), !wr.HTMLUnsafe)
			wr.buf = append(wr.buf, ": "...)
			wr.appendJSON(rm.Interface(), d2)
		}
		wr.buf = append(wr.buf, ',')
		empty = false
	}
	if !empty {
		wr.buf[len(wr.buf)-1] = '\n'
		wr.buf = append(wr.buf, is...)
	}
	wr.buf = append(wr.buf, '}')
}
