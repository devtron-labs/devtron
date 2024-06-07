// Copyright (c) 2020, Peter Ohler, All rights reserved.

package jp

import (
	"reflect"
	"sort"
	"strings"

	"github.com/ohler55/ojg/gen"
)

// Wildcard is used as a flag to indicate the path should be displayed in a
// wildcarded representation.
type Wildcard byte

// Append a fragment string representation of the fragment to the buffer
// then returning the expanded buffer.
func (f Wildcard) Append(buf []byte, bracket, first bool) []byte {
	if bracket || f == '#' {
		buf = append(buf, "[*]"...)
	} else {
		if !first {
			buf = append(buf, '.')
		}
		buf = append(buf, '*')
	}
	return buf
}

func (f Wildcard) remove(value any) (out any, changed bool) {
	out = value
	switch tv := value.(type) {
	case []any:
		if 0 < len(tv) {
			changed = true
			out = []any{}
		}
	case map[string]any:
		if 0 < len(tv) {
			changed = true
			for k := range tv {
				delete(tv, k)
			}
		}
	case gen.Array:
		if 0 < len(tv) {
			changed = true
			out = gen.Array{}
		}
	case gen.Object:
		if 0 < len(tv) {
			changed = true
			for k := range tv {
				delete(tv, k)
			}
		}
	default:
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Slice:
			if 0 < rv.Len() {
				changed = true
				out = reflect.MakeSlice(rv.Type(), 0, 0).Interface()
			}
		case reflect.Map:
			if 0 < rv.Len() {
				changed = true
				out = reflect.MakeMap(rv.Type()).Interface()
			}
		}
	}
	return
}

func (f Wildcard) removeOne(value any) (out any, changed bool) {
	out = value
	switch tv := value.(type) {
	case []any:
		if 0 < len(tv) {
			changed = true
			out = tv[1:]
		}
	case map[string]any:
		if 0 < len(tv) {
			changed = true
			keys := make([]string, 0, len(tv))
			for k := range tv {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			delete(tv, keys[0])
		}
	case gen.Array:
		if 0 < len(tv) {
			changed = true
			out = tv[1:]
		}
	case gen.Object:
		if 0 < len(tv) {
			changed = true
			keys := make([]string, 0, len(tv))
			for k := range tv {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			delete(tv, keys[0])
		}
	default:
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Slice:
			if 0 < rv.Len() {
				changed = true
				out = rv.Slice(1, rv.Len()).Interface()
			}
		case reflect.Map:
			if 0 < rv.Len() {
				changed = true
				keys := rv.MapKeys()
				sort.Slice(keys, func(i, j int) bool {
					return strings.Compare(keys[i].String(), keys[j].String()) < 0
				})
				rv.SetMapIndex(keys[0], reflect.Value{})
			}
		}
	}
	return
}
