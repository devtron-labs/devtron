// Copyright (c) 2020, Peter Ohler, All rights reserved.

package oj

import (
	"io"
	"sync"

	"github.com/ohler55/ojg"
	"github.com/ohler55/ojg/alt"
)

// Options is an alias for ojg.Options
type Options = ojg.Options

// Builder is an aliase for alt.Builder.
type Builder = alt.Builder

var (
	// DefaultOptions are the default options for the this package.
	DefaultOptions = ojg.DefaultOptions
	// BrightOptions are the bright color options.
	BrightOptions = ojg.BrightOptions

	// HTMLOptions are the options that can be used to encode as HTML JSON.
	HTMLOptions = ojg.HTMLOptions

	goOptions  = ojg.GoOptions
	writerPool = sync.Pool{
		New: func() any {
			return &Writer{Options: DefaultOptions, buf: make([]byte, 0, 1024)}
		},
	}
	marshalPool = sync.Pool{
		New: func() any {
			return &Writer{Options: goOptions, buf: make([]byte, 0, 1024), strict: true}
		},
	}
	parserPool = sync.Pool{
		New: func() any {
			return &Parser{}
		},
	}
)

// Parse JSON into a simple type. Arguments are optional and can be a bool,
// func(any) bool for callbacks, or a chan any for chan based
// result delivery.
//
// A bool indicates the NoComment parser attribute should be set to the bool
// value.
//
// A func argument is the callback for the parser if processing multiple
// JSONs. If no callback function is provided the processing is limited to
// only one JSON.
//
// A chan argument will be used to deliver parse results.
func Parse(b []byte, args ...any) (n any, err error) {
	p := parserPool.Get().(*Parser)
	defer parserPool.Put(p)
	return p.Parse(b, args...)
}

// MustParse JSON into a simple type. Arguments are optional and can be a bool,
// func(any) bool for callbacks, or a chan any for chan based
// result delivery. Panics on error
//
// A bool indicates the NoComment parser attribute should be set to the bool
// value.
//
// A func argument is the callback for the parser if processing multiple
// JSONs. If no callback function is provided the processing is limited to
// only one JSON.
//
// A chan argument will be used to deliver parse results.
func MustParse(b []byte, args ...any) (n any) {
	p := parserPool.Get().(*Parser)
	defer parserPool.Put(p)
	var err error
	if n, err = p.Parse(b, args...); err != nil {
		panic(err)
	}
	return
}

// ParseString is similar to Parse except it takes a string
// argument to be parsed instead of a []byte.
func ParseString(s string, args ...any) (n any, err error) {
	p := parserPool.Get().(*Parser)
	defer parserPool.Put(p)
	return p.Parse([]byte(s), args...)
}

// MustParseString is similar to MustParse except it takes a string
// argument to be parsed instead of a []byte.
func MustParseString(s string, args ...any) (n any) {
	p := parserPool.Get().(*Parser)
	defer parserPool.Put(p)
	var err error
	if n, err = p.Parse([]byte(s), args...); err != nil {
		panic(err)
	}
	return
}

// Load a JSON from a io.Reader into a simple type. An error is returned
// if not valid JSON.
func Load(r io.Reader, args ...any) (any, error) {
	p := parserPool.Get().(*Parser)
	defer parserPool.Put(p)
	return p.ParseReader(r, args...)
}

// MustLoad a JSON from a io.Reader into a simple type. Panics on error.
func MustLoad(r io.Reader, args ...any) (n any) {
	p := parserPool.Get().(*Parser)
	defer parserPool.Put(p)
	var err error
	if n, err = p.ParseReader(r, args...); err != nil {
		panic(err)
	}
	return
}

// Validate a JSON string. An error is returned if not valid JSON.
func Validate(b []byte) error {
	v := Validator{}
	return v.Validate(b)
}

// ValidateString a JSON string. An error is returned if not valid JSON.
func ValidateString(s string) error {
	v := Validator{}
	return v.Validate([]byte(s))
}

// ValidateReader a JSON stream. An error is returned if not valid JSON.
func ValidateReader(r io.Reader) error {
	v := Validator{}
	return v.ValidateReader(r)
}

// Unmarshal parses the provided JSON and stores the result in the value
// pointed to by vp.
func Unmarshal(data []byte, vp any, recomposer ...*alt.Recomposer) (err error) {
	p := Parser{}
	p.num.ForceFloat = true
	var v any
	if v, err = p.Parse(data); err == nil {
		if 0 < len(recomposer) {
			_, err = recomposer[0].Recompose(v, vp)
		} else {
			_, err = alt.Recompose(v, vp)
		}
	}
	return
}

// JSON returns a JSON string for the data provided. The data can be a
// simple type of nil, bool, int, floats, time.Time, []any, or
// map[string]any or a Node type, The args, if supplied can be an
// int as an indent or a *Options.
func JSON(data any, args ...any) string {
	var wr *Writer
	if 0 < len(args) {
		wr = pickWriter(args[0], false)
	}
	if wr == nil {
		wr, _ = writerPool.Get().(*Writer)
		defer writerPool.Put(wr)
	}
	return wr.JSON(data)
}

// Marshal returns a JSON string for the data provided. The data can be a
// simple type of nil, bool, int, floats, time.Time, []any, or
// map[string]any or a gen.Node type, The args, if supplied can be an
// int as an indent, *ojg.Options, or a *Writer. An error will be returned if
// the Option.Strict flag is true and a value is encountered that can not be
// encoded other than by using the %v format of the fmt package.
func Marshal(data any, args ...any) (out []byte, err error) {
	var wr *Writer
	if 0 < len(args) {
		wr = pickWriter(args[0], true)
	}
	if wr == nil {
		wr, _ = marshalPool.Get().(*Writer)
		defer marshalPool.Put(wr)
	} else {
		wr.strict = true
	}
	defer func() {
		if r := recover(); r != nil {
			wr.buf = wr.buf[:0]
			err = ojg.NewError(r)
		}
	}()
	wr.MustJSON(data)
	out = make([]byte, len(wr.buf))
	copy(out, wr.buf)

	return
}

// Write a JSON string for the data provided. The data can be a simple type of
// nil, bool, int, floats, time.Time, []any, or map[string]any
// or a Node type, The args, if supplied can be an int as an indent or a
// *Options.
func Write(w io.Writer, data any, args ...any) (err error) {
	var wr *Writer
	if 0 < len(args) {
		wr = pickWriter(args[0], false)
	}
	if wr == nil {
		wr, _ = writerPool.Get().(*Writer)
		defer writerPool.Put(wr)
	}
	return wr.Write(w, data)
}

func pickWriter(arg any, strict bool) (wr *Writer) {
	switch ta := arg.(type) {
	case int:
		wr = &Writer{
			Options: ojg.GoOptions,
			buf:     make([]byte, 0, 1024),
			strict:  strict,
		}
		wr.Indent = ta
	case *ojg.Options:
		wr = &Writer{
			Options: *ta,
			buf:     make([]byte, 0, 1024),
			strict:  strict,
		}
	case *Writer:
		wr = ta
	}
	return
}
