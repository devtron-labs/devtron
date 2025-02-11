// +build pq

package postgresql

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

import (
	"bytes"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/lib/pq"
)

// JSONB represents a PostgreSQL's JSONB value:
// https://www.postgresql.org/docs/9.6/static/datatype-json.html. JSONB
// satisfies sqlbuilder.ScannerValuer.
type JSONB struct {
	Data interface{}
}

// MarshalJSON encodes the wrapper value as JSON.
func (j JSONB) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Data)
}

// UnmarshalJSON decodes the given JSON into the wrapped value.
func (j *JSONB) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	j.Data = v
	return nil
}

// Scan satisfies the sql.Scanner interface.
func (j *JSONB) Scan(src interface{}) error {
	if j.Data == nil {
		return nil
	}
	if src == nil {
		dv := reflect.Indirect(reflect.ValueOf(j.Data))
		dv.Set(reflect.Zero(dv.Type()))
		return nil
	}

	b, ok := src.([]byte)
	if !ok {
		return errors.New("Scan source was not []bytes")
	}

	if err := json.Unmarshal(b, j.Data); err != nil {
		return err
	}
	return nil
}

// Value satisfies the driver.Valuer interface.
func (j JSONB) Value() (driver.Value, error) {
	// See https://github.com/lib/pq/issues/528#issuecomment-257197239 on why are
	// we returning string instead of []byte.
	if j.Data == nil {
		return nil, nil
	}
	if v, ok := j.Data.(json.RawMessage); ok {
		return string(v), nil
	}
	b, err := json.Marshal(j.Data)
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

// Float32Array represents a one-dimensional array of float32s (`[]float32{}`)
// that is compatible with PostgreSQL's double precision array (`double
// precision[]`). Float32Array satisfies sqlbuilder.ScannerValuer.
type Float32Array pq.Float32Array

// Value satisfies the driver.Valuer interface.
func (f Float32Array) Value() (driver.Value, error) {
	return pq.Float32Array(f).Value()
}

// Scan satisfies the sql.Scanner interface.
func (f *Float32Array) Scan(src interface{}) error {
	s := pq.Float32Array(*f)
	if err := s.Scan(src); err != nil {
		return err
	}
	*f = Float32Array(s)
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

type Bytea []byte

// Scan satisfies the sql.Scanner interface.
func (b *Bytea) Scan(src interface{}) error {
	decoded, err := parseBytea(src.([]byte))
	if err != nil {
		return err
	}
	if len(decoded) < 1 {
		*b = nil
		return nil
	}
	(*b) = make(Bytea, len(decoded))
	for i := range decoded {
		(*b)[i] = decoded[i]
	}
	return nil
}

type Time time.Time

// Parse a bytea value received from the server.  Both "hex" and the legacy
// "escape" format are supported.
func parseBytea(s []byte) (result []byte, err error) {
	if len(s) >= 2 && bytes.Equal(s[:2], []byte("\\x")) {
		// bytea_output = hex
		s = s[2:] // trim off leading "\\x"
		result = make([]byte, hex.DecodedLen(len(s)))
		_, err := hex.Decode(result, s)
		if err != nil {
			return nil, err
		}
	} else {
		// bytea_output = escape
		for len(s) > 0 {
			if s[0] == '\\' {
				// escaped '\\'
				if len(s) >= 2 && s[1] == '\\' {
					result = append(result, '\\')
					s = s[2:]
					continue
				}

				// '\\' followed by an octal number
				if len(s) < 4 {
					return nil, fmt.Errorf("invalid bytea sequence %v", s)
				}
				r, err := strconv.ParseInt(string(s[1:4]), 8, 9)
				if err != nil {
					return nil, fmt.Errorf("could not parse bytea value: %s", err.Error())
				}
				result = append(result, byte(r))
				s = s[4:]
			} else {
				// We hit an unescaped, raw byte.  Try to read in as many as
				// possible in one go.
				i := bytes.IndexByte(s, '\\')
				if i == -1 {
					result = append(result, s...)
					break
				}
				result = append(result, s[:i]...)
				s = s[i:]
			}
		}
	}

	return result, nil
}
