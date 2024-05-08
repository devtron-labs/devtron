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
	"encoding/json"
	"errors"
	"reflect"

	"github.com/lib/pq"
	"upper.io/db.v3/lib/sqlbuilder"
)

// Array returns a sqlbuilder.ScannerValuer for any given slice. Slice elements
// may require their own sqlbuilder.ScannerValuer.
func Array(in interface{}) sqlbuilder.ScannerValuer {
	return pq.Array(in)
}

// JSONB represents a PostgreSQL's JSONB value:
// https://www.postgresql.org/docs/9.6/static/datatype-json.html. JSONB
// satisfies sqlbuilder.ScannerValuer.
type JSONB struct {
	V interface{}
}

// MarshalJSON encodes the wrapper value as JSON.
func (j JSONB) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.V)
}

// UnmarshalJSON decodes the given JSON into the wrapped value.
func (j *JSONB) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	j.V = v
	return nil
}

// Scan satisfies the sql.Scanner interface.
func (j *JSONB) Scan(src interface{}) error {
	if src == nil {
		j.V = nil
		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return errors.New("Scan source was not []bytes")
	}

	if err := json.Unmarshal(b, &j.V); err != nil {
		return err
	}
	return nil
}

// Value satisfies the driver.Valuer interface.
func (j JSONB) Value() (driver.Value, error) {
	// See https://github.com/lib/pq/issues/528#issuecomment-257197239 on why are
	// we returning string instead of []byte.
	if j.V == nil {
		return nil, nil
	}
	if v, ok := j.V.(json.RawMessage); ok {
		return string(v), nil
	}
	b, err := json.Marshal(j.V)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// StringArray represents a one-dimensional array of strings (`[]string{}`)
// that is compatible with PostgreSQL's text array (`text[]`). StringArray
// satisfies sqlbuilder.ScannerValuer.
type StringArray pq.StringArray

// Value satisfies the driver.Valuer interface.
func (a StringArray) Value() (driver.Value, error) {
	return pq.StringArray(a).Value()
}

// Scan satisfies the sql.Scanner interface.
func (a *StringArray) Scan(src interface{}) error {
	s := pq.StringArray(*a)
	if err := s.Scan(src); err != nil {
		return err
	}
	*a = StringArray(s)
	return nil
}

// Int64Array represents a one-dimensional array of int64s (`[]int64{}`) that
// is compatible with PostgreSQL's integer array (`integer[]`). Int64Array
// satisfies sqlbuilder.ScannerValuer.
type Int64Array pq.Int64Array

// Value satisfies the driver.Valuer interface.
func (i Int64Array) Value() (driver.Value, error) {
	return pq.Int64Array(i).Value()
}

// Scan satisfies the sql.Scanner interface.
func (i *Int64Array) Scan(src interface{}) error {
	s := pq.Int64Array(*i)
	if err := s.Scan(src); err != nil {
		return err
	}
	*i = Int64Array(s)
	return nil
}

// Float64Array represents a one-dimensional array of float64s (`[]float64{}`)
// that is compatible with PostgreSQL's double precision array (`double
// precision[]`). Float64Array satisfies sqlbuilder.ScannerValuer.
type Float64Array pq.Float64Array

// Value satisfies the driver.Valuer interface.
func (f Float64Array) Value() (driver.Value, error) {
	return pq.Float64Array(f).Value()
}

// Scan satisfies the sql.Scanner interface.
func (f *Float64Array) Scan(src interface{}) error {
	s := pq.Float64Array(*f)
	if err := s.Scan(src); err != nil {
		return err
	}
	*f = Float64Array(s)
	return nil
}

// BoolArray represents a one-dimensional array of int64s (`[]bool{}`) that
// is compatible with PostgreSQL's boolean type (`boolean[]`). BoolArray
// satisfies sqlbuilder.ScannerValuer.
type BoolArray pq.BoolArray

// Value satisfies the driver.Valuer interface.
func (b BoolArray) Value() (driver.Value, error) {
	return pq.BoolArray(b).Value()
}

// Scan satisfies the sql.Scanner interface.
func (b *BoolArray) Scan(src interface{}) error {
	s := pq.BoolArray(*b)
	if err := s.Scan(src); err != nil {
		return err
	}
	*b = BoolArray(s)
	return nil
}

// GenericArray represents a one-dimensional array of any type
// (`[]interface{}`) that is compatible with PostgreSQL's array type.
// GenericArray satisfies sqlbuilder.ScannerValuer and its elements may need to
// satisfy sqlbuilder.ScannerValuer too.
type GenericArray pq.GenericArray

// Value satisfies the driver.Valuer interface.
func (g GenericArray) Value() (driver.Value, error) {
	return pq.GenericArray(g).Value()
}

// Scan satisfies the sql.Scanner interface.
func (g *GenericArray) Scan(src interface{}) error {
	s := pq.GenericArray(*g)
	if err := s.Scan(src); err != nil {
		return err
	}
	*g = GenericArray(s)
	return nil
}

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

// EncodeJSONB is deprecated and going to be removed. Use ScanJSONB instead.
func EncodeJSONB(i interface{}) (driver.Value, error) {
	return JSONBValue(i)
}

// DecodeJSONB is deprecated and going to be removed. Use JSONBValue instead.
func DecodeJSONB(dst interface{}, src interface{}) error {
	return ScanJSONB(dst, src)
}

// JSONBConverter provides a helper method WrapValue that satisfies
// sqlbuilder.ValueWrapper, can be used to encode Go structs into JSONB
// PostgreSQL types and vice versa.
//
// Example:
//
//   type MyCustomStruct struct {
//     ID int64 `db:"id" json:"id"`
//     Name string `db:"name" json:"name"`
//     ...
//     postgresql.JSONBConverter
//   }
type JSONBConverter struct {
}

// WrapValue satisfies sqlbuilder.ValueWrapper
func (obj *JSONBConverter) WrapValue(src interface{}) interface{} {
	return &JSONB{src}
}

func autoWrap(elem reflect.Value, v interface{}) interface{} {
	kind := elem.Kind()

	if kind == reflect.Invalid {
		return v
	}

	if elem.Type().Implements(sqlbuilder.ScannerType) {
		return v
	}

	if elem.Type().Implements(sqlbuilder.ValuerType) {
		return v
	}

	if elem.Type().Implements(sqlbuilder.ValueWrapperType) {
		if elem.Type().Kind() == reflect.Ptr {
			w := reflect.ValueOf(v)
			if w.Kind() == reflect.Ptr {
				z := reflect.Zero(w.Elem().Type())
				w.Elem().Set(z)
				return &JSONB{v}
			}
		}
		vw := elem.Interface().(sqlbuilder.ValueWrapper)
		return vw.WrapValue(elem.Interface())
	}

	switch kind {
	case reflect.Ptr:
		return autoWrap(elem.Elem(), v)
	case reflect.Slice:
		return &JSONB{v}
	case reflect.Map:
		if reflect.TypeOf(v).Kind() == reflect.Ptr {
			w := reflect.ValueOf(v)
			z := reflect.New(w.Elem().Type())
			w.Elem().Set(z.Elem())
		}
		return &JSONB{v}
	}

	return v
}

// Type checks.
var (
	_ sqlbuilder.ValueWrapper = &JSONBConverter{}

	_ sqlbuilder.ScannerValuer = &StringArray{}
	_ sqlbuilder.ScannerValuer = &Int64Array{}
	_ sqlbuilder.ScannerValuer = &Float64Array{}
	_ sqlbuilder.ScannerValuer = &BoolArray{}
	_ sqlbuilder.ScannerValuer = &GenericArray{}
	_ sqlbuilder.ScannerValuer = &JSONBMap{}
	_ sqlbuilder.ScannerValuer = &JSONBArray{}
)
