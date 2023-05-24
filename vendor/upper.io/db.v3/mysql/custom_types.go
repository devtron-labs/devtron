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

package mysql

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"reflect"

	"upper.io/db.v3/lib/sqlbuilder"
)

// JSON represents a MySQL's JSON value:
// https://www.mysql.org/docs/9.6/static/datatype-json.html. JSON
// satisfies sqlbuilder.ScannerValuer.
type JSON struct {
	V interface{}
}

// MarshalJSON encodes the wrapper value as JSON.
func (j JSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.V)
}

// UnmarshalJSON decodes the given JSON into the wrapped value.
func (j *JSON) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	j.V = v
	return nil
}

// Scan satisfies the sql.Scanner interface.
func (j *JSON) Scan(src interface{}) error {
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
func (j JSON) Value() (driver.Value, error) {
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

// JSONMap represents a map of interfaces with string keys
// (`map[string]interface{}`) that is compatible with MySQL's JSON type.
// JSONMap satisfies sqlbuilder.ScannerValuer.
type JSONMap map[string]interface{}

// Value satisfies the driver.Valuer interface.
func (m JSONMap) Value() (driver.Value, error) {
	return JSONValue(m)
}

// Scan satisfies the sql.Scanner interface.
func (m *JSONMap) Scan(src interface{}) error {
	*m = map[string]interface{}(nil)
	return ScanJSON(m, src)
}

// JSONArray represents an array of any type (`[]interface{}`) that is
// compatible with MySQL's JSON type. JSONArray satisfies
// sqlbuilder.ScannerValuer.
type JSONArray []interface{}

// Value satisfies the driver.Valuer interface.
func (a JSONArray) Value() (driver.Value, error) {
	return JSONValue(a)
}

// Scan satisfies the sql.Scanner interface.
func (a *JSONArray) Scan(src interface{}) error {
	return ScanJSON(a, src)
}

// JSONValue takes an interface and provides a driver.Value that can be
// stored as a JSON column.
func JSONValue(i interface{}) (driver.Value, error) {
	v := JSON{i}
	return v.Value()
}

// ScanJSON decodes a JSON byte stream into the passed dst value.
func ScanJSON(dst interface{}, src interface{}) error {
	v := JSON{dst}
	return v.Scan(src)
}

// EncodeJSON is deprecated and going to be removed. Use ScanJSON instead.
func EncodeJSON(i interface{}) (driver.Value, error) {
	return JSONValue(i)
}

// DecodeJSON is deprecated and going to be removed. Use JSONValue instead.
func DecodeJSON(dst interface{}, src interface{}) error {
	return ScanJSON(dst, src)
}

// JSONConverter provides a helper method WrapValue that satisfies
// sqlbuilder.ValueWrapper, can be used to encode Go structs into JSON
// MySQL types and vice versa.
//
// Example:
//
//   type MyCustomStruct struct {
//     ID int64 `db:"id" json:"id"`
//     Name string `db:"name" json:"name"`
//     ...
//     mysql.JSONConverter
//   }
type JSONConverter struct {
}

// WrapValue satisfies sqlbuilder.ValueWrapper
func (obj *JSONConverter) WrapValue(src interface{}) interface{} {
	return &JSON{src}
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
				return &JSON{v}
			}
		}
		vw := elem.Interface().(sqlbuilder.ValueWrapper)
		return vw.WrapValue(elem.Interface())
	}

	switch kind {
	case reflect.Ptr:
		return autoWrap(elem.Elem(), v)
	case reflect.Slice:
		return &JSON{v}
	case reflect.Map:
		if reflect.TypeOf(v).Kind() == reflect.Ptr {
			w := reflect.ValueOf(v)
			z := reflect.New(w.Elem().Type())
			w.Elem().Set(z.Elem())
		}
		return &JSON{v}
	}

	return v
}

// Type checks.
var (
	_ sqlbuilder.ValueWrapper  = &JSONConverter{}
	_ sqlbuilder.ScannerValuer = &JSONMap{}
	_ sqlbuilder.ScannerValuer = &JSONArray{}
)
