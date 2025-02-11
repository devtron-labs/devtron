// Copyright (c) 2012-present The upper.io/db authors. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package postgresql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"time"

	"github.com/upper/db/v4/internal/sqlbuilder"
)

// JSONBMap represents a map of interfaces with string keys
// (`map[string]interface{}`) that is compatible with PostgreSQL's JSONB type.
// JSONBMap satisfies sqlbuilder.ScannerValuer.
type JSONBMap map[string]interface{}

// Value satisfies the driver.Valuer interface.
func (m JSONBMap) Value() (driver.Value, error) {
	return JSONBValue(m)
}

// Scan satisfies the sql.Scanner interface.
func (m *JSONBMap) Scan(src interface{}) error {
	*m = map[string]interface{}(nil)
	return ScanJSONB(m, src)
}

// JSONBArray represents an array of any type (`[]interface{}`) that is
// compatible with PostgreSQL's JSONB type. JSONBArray satisfies
// sqlbuilder.ScannerValuer.
type JSONBArray []interface{}

// Value satisfies the driver.Valuer interface.
func (a JSONBArray) Value() (driver.Value, error) {
	return JSONBValue(a)
}

// Scan satisfies the sql.Scanner interface.
func (a *JSONBArray) Scan(src interface{}) error {
	return ScanJSONB(a, src)
}

// JSONBValue takes an interface and provides a driver.Value that can be
// stored as a JSONB column.
func JSONBValue(i interface{}) (driver.Value, error) {
	v := JSONB{i}
	return v.Value()
}

// ScanJSONB decodes a JSON byte stream into the passed dst value.
func ScanJSONB(dst interface{}, src interface{}) error {
	v := JSONB{dst}
	return v.Scan(src)
}

type JSONBConverter struct {
}

func (*JSONBConverter) ConvertValue(in interface{}) interface {
	sql.Scanner
	driver.Valuer
} {
	return &JSONB{in}
}

type timeWrapper struct {
	v   **time.Time
	loc *time.Location
}

func (t timeWrapper) Value() (driver.Value, error) {
	if *t.v != nil {
		return **t.v, nil
	}
	return nil, nil
}

func (t *timeWrapper) Scan(src interface{}) error {
	if src == nil {
		nilTime := (*time.Time)(nil)
		if t.v == nil {
			t.v = &nilTime
		} else {
			*(t.v) = nilTime
		}
		return nil
	}
	tz := src.(time.Time)
	if t.loc != nil && (tz.Location() == time.Local) {
		tz = tz.In(t.loc)
	}
	if tz.Location().String() == "" {
		tz = tz.In(time.UTC)
	}
	if *(t.v) == nil {
		*(t.v) = &tz
	} else {
		**t.v = tz
	}
	return nil
}

func (d *database) ConvertValueContext(ctx context.Context, in interface{}) interface{} {
	tz, _ := ctx.Value("timezone").(*time.Location)

	switch v := in.(type) {
	case *time.Time:
		return &timeWrapper{&v, tz}
	case **time.Time:
		return &timeWrapper{v, tz}
	}

	return d.ConvertValue(in)
}

// Type checks.
var (
	_ sqlbuilder.ScannerValuer = &StringArray{}
	_ sqlbuilder.ScannerValuer = &Int64Array{}
	_ sqlbuilder.ScannerValuer = &Float64Array{}
	_ sqlbuilder.ScannerValuer = &Float32Array{}
	_ sqlbuilder.ScannerValuer = &BoolArray{}
	_ sqlbuilder.ScannerValuer = &JSONBMap{}
	_ sqlbuilder.ScannerValuer = &JSONBArray{}
	_ sqlbuilder.ScannerValuer = &JSONB{}
)
