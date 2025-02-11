// +build !pq

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
	"database/sql/driver"

	"github.com/jackc/pgtype"
)

// JSONB represents a PostgreSQL's JSONB value:
// https://www.postgresql.org/docs/9.6/static/datatype-json.html. JSONB
// satisfies sqlbuilder.ScannerValuer.
type JSONB struct {
	Data interface{}
}

// MarshalJSON encodes the wrapper value as JSON.
func (j JSONB) MarshalJSON() ([]byte, error) {
	t := &pgtype.JSONB{}
	if err := t.Set(j.Data); err != nil {
		return nil, err
	}
	return t.MarshalJSON()
}

// UnmarshalJSON decodes the given JSON into the wrapped value.
func (j *JSONB) UnmarshalJSON(b []byte) error {
	t := &pgtype.JSONB{}
	if err := t.UnmarshalJSON(b); err != nil {
		return err
	}
	if j.Data == nil {
		j.Data = t.Get()
		return nil
	}
	if err := t.AssignTo(&j.Data); err != nil {
		return err
	}
	return nil
}

// Scan satisfies the sql.Scanner interface.
func (j *JSONB) Scan(src interface{}) error {
	t := &pgtype.JSONB{}
	if err := t.Scan(src); err != nil {
		return err
	}
	if j.Data == nil {
		j.Data = t.Get()
		return nil
	}
	if err := t.AssignTo(j.Data); err != nil {
		return err
	}
	return nil
}

// Value satisfies the driver.Valuer interface.
func (j JSONB) Value() (driver.Value, error) {
	t := &pgtype.JSONB{}
	if err := t.Set(j.Data); err != nil {
		return nil, err
	}
	return t.Value()
}

// StringArray represents a one-dimensional array of strings (`[]string{}`)
// that is compatible with PostgreSQL's text array (`text[]`). StringArray
// satisfies sqlbuilder.ScannerValuer.
type StringArray []string

// Value satisfies the driver.Valuer interface.
func (a StringArray) Value() (driver.Value, error) {
	t := pgtype.TextArray{}
	if err := t.Set(a); err != nil {
		return nil, err
	}
	return t.Value()
}

// Scan satisfies the sql.Scanner interface.
func (sa *StringArray) Scan(src interface{}) error {
	d := []string{}
	t := pgtype.TextArray{}
	if err := t.Scan(src); err != nil {
		return err
	}
	if err := t.AssignTo(&d); err != nil {
		return err
	}
	*sa = StringArray(d)
	return nil
}

type Bytea []byte

func (b Bytea) Value() (driver.Value, error) {
	t := pgtype.Bytea{Bytes: b}
	if err := t.Set(b); err != nil {
		return nil, err
	}
	return t.Value()
}

func (b *Bytea) Scan(src interface{}) error {
	d := []byte{}
	t := pgtype.Bytea{}
	if err := t.Scan(src); err != nil {
		return err
	}
	if err := t.AssignTo(&d); err != nil {
		return err
	}
	*b = Bytea(d)
	return nil
}

// ByteaArray represents a one-dimensional array of strings (`[]string{}`)
// that is compatible with PostgreSQL's text array (`text[]`). ByteaArray
// satisfies sqlbuilder.ScannerValuer.
type ByteaArray [][]byte

// Value satisfies the driver.Valuer interface.
func (a ByteaArray) Value() (driver.Value, error) {
	t := pgtype.ByteaArray{}
	if err := t.Set(a); err != nil {
		return nil, err
	}
	return t.Value()
}

// Scan satisfies the sql.Scanner interface.
func (ba *ByteaArray) Scan(src interface{}) error {
	d := [][]byte{}
	t := pgtype.ByteaArray{}
	if err := t.Scan(src); err != nil {
		return err
	}
	if err := t.AssignTo(&d); err != nil {
		return err
	}
	*ba = ByteaArray(d)
	return nil
}

// Int64Array represents a one-dimensional array of int64s (`[]int64{}`) that
// is compatible with PostgreSQL's integer array (`integer[]`). Int64Array
// satisfies sqlbuilder.ScannerValuer.
type Int64Array []int64

// Value satisfies the driver.Valuer interface.
func (i64a Int64Array) Value() (driver.Value, error) {
	t := pgtype.Int8Array{}
	if err := t.Set(i64a); err != nil {
		return nil, err
	}
	return t.Value()
}

// Scan satisfies the sql.Scanner interface.
func (i64a *Int64Array) Scan(src interface{}) error {
	d := []int64{}
	t := pgtype.Int8Array{}
	if err := t.Scan(src); err != nil {
		return err
	}
	if err := t.AssignTo(&d); err != nil {
		return err
	}
	*i64a = Int64Array(d)
	return nil
}

// Int32Array represents a one-dimensional array of int32s (`[]int32{}`) that
// is compatible with PostgreSQL's integer array (`integer[]`). Int32Array
// satisfies sqlbuilder.ScannerValuer.
type Int32Array []int32

// Value satisfies the driver.Valuer interface.
func (i32a Int32Array) Value() (driver.Value, error) {
	t := pgtype.Int4Array{}
	if err := t.Set(i32a); err != nil {
		return nil, err
	}
	return t.Value()
}

// Scan satisfies the sql.Scanner interface.
func (i32a *Int32Array) Scan(src interface{}) error {
	d := []int32{}
	t := pgtype.Int4Array{}
	if err := t.Scan(src); err != nil {
		return err
	}
	if err := t.AssignTo(&d); err != nil {
		return err
	}
	*i32a = Int32Array(d)
	return nil
}

// Float64Array represents a one-dimensional array of float64s (`[]float64{}`)
// that is compatible with PostgreSQL's double precision array (`double
// precision[]`). Float64Array satisfies sqlbuilder.ScannerValuer.
type Float64Array []float64

// Value satisfies the driver.Valuer interface.
func (f64a Float64Array) Value() (driver.Value, error) {
	t := pgtype.Float8Array{}
	if err := t.Set(f64a); err != nil {
		return nil, err
	}
	return t.Value()
}

// Scan satisfies the sql.Scanner interface.
func (f64a *Float64Array) Scan(src interface{}) error {
	d := []float64{}
	t := pgtype.Float8Array{}
	if err := t.Scan(src); err != nil {
		return err
	}
	if err := t.AssignTo(&d); err != nil {
		return err
	}
	*f64a = Float64Array(d)
	return nil
}

// Float32Array represents a one-dimensional array of float32s (`[]float32{}`)
// that is compatible with PostgreSQL's double precision array (`double
// precision[]`). Float32Array satisfies sqlbuilder.ScannerValuer.
type Float32Array []float32

// Value satisfies the driver.Valuer interface.
func (f32a Float32Array) Value() (driver.Value, error) {
	t := pgtype.Float8Array{}
	if err := t.Set(f32a); err != nil {
		return nil, err
	}
	return t.Value()
}

// Scan satisfies the sql.Scanner interface.
func (f32a *Float32Array) Scan(src interface{}) error {
	d := []float32{}
	t := pgtype.Float8Array{}
	if err := t.Scan(src); err != nil {
		return err
	}
	if err := t.AssignTo(&d); err != nil {
		return err
	}
	*f32a = Float32Array(d)
	return nil
}

// BoolArray represents a one-dimensional array of int64s (`[]bool{}`) that
// is compatible with PostgreSQL's boolean type (`boolean[]`). BoolArray
// satisfies sqlbuilder.ScannerValuer.
type BoolArray []bool

// Value satisfies the driver.Valuer interface.
func (ba BoolArray) Value() (driver.Value, error) {
	t := pgtype.BoolArray{}
	if err := t.Set(ba); err != nil {
		return nil, err
	}
	return t.Value()
}

// Scan satisfies the sql.Scanner interface.
func (ba *BoolArray) Scan(src interface{}) error {
	d := []bool{}
	t := pgtype.BoolArray{}
	if err := t.Scan(src); err != nil {
		return err
	}
	if err := t.AssignTo(&d); err != nil {
		return err
	}
	*ba = BoolArray(d)
	return nil
}
