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

package sqlbuilder

import (
	"database/sql"
	"database/sql/driver"

	"reflect"
)

var (
	// ValuerType is the reflection type for the driver.Valuer interface.
	ValuerType = reflect.TypeOf((*driver.Valuer)(nil)).Elem()

	// ScannerType is the reflection type for the sql.Scanner interface.
	ScannerType = reflect.TypeOf((*sql.Scanner)(nil)).Elem()

	// ValueWrapperType is the reflection type for the sql.ValueWrapper interface.
	ValueWrapperType = reflect.TypeOf((*ValueWrapper)(nil)).Elem()
)

// ValueWrapper defines a method WrapValue that query arguments can use to wrap
// themselves around helper types right before being used in a query.
//
// Example:
//
//   func (a MyCustomArray) WrapValue(value interface{}) interface{} {
//     // postgresql.Array adds a driver.Valuer and sql.Scanner around
//     // custom arrays.
//	   return postgresql.Array(values)
//   }
type ValueWrapper interface {
	WrapValue(value interface{}) interface{}
}

// ScannerValuer represents a value that satisfies both driver.Valuer and
// sql.Scanner interfaces.
type ScannerValuer interface {
	driver.Valuer
	sql.Scanner
}
