// Copyright (c) 2021, Peter Ohler, All rights reserved.

package oj

import (
	"bytes"
	"reflect"
	"sort"
	"strings"
	"sync"
	"unsafe"
)

const (
	maskByTag  = byte(0x01)
	maskExact  = byte(0x02) // exact key vs lowwer case first letter
	maskNested = byte(0x04)
	maskPretty = byte(0x08)
	maskMax    = byte(0x10)
)

type sinfo struct {
	rt     reflect.Type
	fields [16][]*finfo
}

var (
	structMut sync.Mutex
	// Keyed by the pointer to the type.
	structMap      = map[uintptr]*sinfo{}
	structEmptyMap = map[uintptr]*sinfo{}
)

// Non-locking version used in field creation.
func getTypeStruct(rt reflect.Type, embedded, omitEmpty bool) (st *sinfo) {
	x := (*[2]uintptr)(unsafe.Pointer(&rt))[1]
	if st = structMap[x]; st != nil {
		return
	}
	return buildStruct(rt, x, embedded, omitEmpty)
}

func getSinfo(v any, omitEmpty bool) (st *sinfo) {
	x := (*[2]uintptr)(unsafe.Pointer(&v))[0]
	sm := structMap
	if omitEmpty {
		sm = structEmptyMap
	}
	structMut.Lock()
	defer structMut.Unlock()
	if st = sm[x]; st != nil {
		return
	}
	return buildStruct(reflect.TypeOf(v), x, false, omitEmpty)
}

func buildStruct(rt reflect.Type, x uintptr, embedded, omitEmpty bool) (st *sinfo) {
	st = &sinfo{rt: rt}
	if omitEmpty {
		structEmptyMap[x] = st
	} else {
		structMap[x] = st
	}
	for u := byte(0); u < maskMax; u++ {
		if (maskByTag&u) != 0 && (maskExact&u) != 0 { // reuse previously built
			st.fields[u] = st.fields[u & ^maskExact]
			continue
		}
		st.fields[u] = buildFields(st.rt, u, embedded, omitEmpty)
	}
	return
}

func buildFields(rt reflect.Type, u byte, embedded, omitEmpty bool) (fa []*finfo) {
	switch {
	case (maskByTag & u) != 0:
		fa = buildTagFields(rt, (maskNested&u) != 0, (maskPretty&u) != 0, embedded, omitEmpty)
	case (maskExact & u) != 0:
		fa = buildExactFields(rt, (maskNested&u) != 0, (maskPretty&u) != 0, embedded, omitEmpty)
	default:
		fa = buildLowFields(rt, (maskNested&u) != 0, (maskPretty&u) != 0, embedded, omitEmpty)
	}
	sort.Slice(fa, func(i, j int) bool { return 0 > strings.Compare(fa[i].key, fa[j].key) })
	return
}

func buildTagFields(rt reflect.Type, out, pretty, embedded, omitEmpty bool) (fa []*finfo) {
	for i := rt.NumField() - 1; 0 <= i; i-- {
		f := rt.Field(i)
		name := []byte(f.Name)
		if len(name) == 0 || 'a' <= name[0] {
			continue
		}
		if f.Anonymous && !out {
			if f.Type.Kind() == reflect.Ptr {
				for _, fi := range buildTagFields(f.Type.Elem(), out, pretty, embedded, omitEmpty) {
					fi.index = append([]int{i}, fi.index...)
					fi.Append = fi.iAppend
					fa = append(fa, fi)
				}
			} else {
				for _, fi := range buildTagFields(f.Type, out, pretty, embedded, omitEmpty) {
					fi.index = append([]int{i}, fi.index...)
					fi.offset += f.Offset
					fa = append(fa, fi)
				}
			}
		} else {
			asString := false
			key := f.Name
			if tag, ok := f.Tag.Lookup("json"); ok && 0 < len(tag) {
				parts := strings.Split(tag, ",")
				switch parts[0] {
				case "":
					key = f.Name
				case "-":
					if 1 < len(parts) {
						key = "-"
					} else {
						continue
					}
				default:
					key = parts[0]
				}
				for _, p := range parts[1:] {
					switch p {
					case "omitempty":
						omitEmpty = true
					case "string":
						asString = true
					}
				}
			}
			fa = append(fa, newFinfo(&f, key, omitEmpty, asString, pretty, embedded))
		}
	}
	return
}

func buildExactFields(rt reflect.Type, out, pretty, embedded, omitEmpty bool) (fa []*finfo) {
	for i := rt.NumField() - 1; 0 <= i; i-- {
		f := rt.Field(i)
		name := []byte(f.Name)
		if len(name) == 0 || 'a' <= name[0] {
			continue
		}
		if f.Anonymous && !out {
			if f.Type.Kind() == reflect.Ptr {
				for _, fi := range buildExactFields(f.Type.Elem(), out, pretty, embedded, omitEmpty) {
					fi.index = append([]int{i}, fi.index...)
					fi.Append = fi.iAppend
					fa = append(fa, fi)
				}
			} else {
				for _, fi := range buildExactFields(f.Type, out, pretty, embedded, omitEmpty) {
					fi.index = append([]int{i}, fi.index...)
					fi.offset += f.Offset
					fa = append(fa, fi)
				}
			}
		} else {
			fa = append(fa, newFinfo(&f, f.Name, omitEmpty, false, pretty, embedded))
		}
	}
	return
}

func buildLowFields(rt reflect.Type, out, pretty, embedded, omitEmpty bool) (fa []*finfo) {
	for i := rt.NumField() - 1; 0 <= i; i-- {
		f := rt.Field(i)
		name := []byte(f.Name)
		if len(name) == 0 || 'a' <= name[0] {
			continue
		}
		if f.Anonymous && !out {
			if f.Type.Kind() == reflect.Ptr {
				for _, fi := range buildLowFields(f.Type.Elem(), out, pretty, embedded, omitEmpty) {
					fi.index = append([]int{i}, fi.index...)
					fi.Append = fi.iAppend
					fa = append(fa, fi)
				}
			} else {
				for _, fi := range buildLowFields(f.Type, out, pretty, embedded, omitEmpty) {
					fi.index = append([]int{i}, fi.index...)
					fi.offset += f.Offset
					fa = append(fa, fi)
				}
			}
		} else {
			if 3 < len(name) {
				if name[0] < 0x80 {
					name[0] |= 0x20
				}
			} else {
				name = bytes.ToLower(name)
			}
			fa = append(fa, newFinfo(&f, string(name), omitEmpty, false, pretty, embedded))
		}
	}
	return
}
