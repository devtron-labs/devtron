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

package db

// Collection is an interface that defines methods useful for handling tables.
type Collection interface {
	// Insert inserts a new item into the collection, it accepts one argument
	// that can be either a map or a struct. If the call succeeds, it returns the
	// ID of the newly added element as an `interface{}` (the actual type of this
	// ID depends on both the database adapter and the column that stores this
	// ID). The ID returned by Insert() could be passed directly to Find() to
	// retrieve the newly added element.
	Insert(interface{}) (interface{}, error)

	// InsertReturning is like Insert() but it updates the passed map or struct
	// with the newly inserted element (and with automatic fields, like IDs,
	// timestamps, etc). This is all done atomically within a transaction.  If
	// the database does not support transactions this method returns
	// db.ErrUnsupported.
	InsertReturning(interface{}) error

	// UpdateReturning takes a pointer to map or struct and tries to update the
	// given item on the collection based on the item's primary keys. Once the
	// element is updated, UpdateReturning will query the element that was just
	// updated. If the database does not support transactions this method returns
	// db.ErrUnsupported
	UpdateReturning(interface{}) error

	// Exists returns true if the collection exists, false otherwise.
	Exists() bool

	// Find defines a new result set with elements from the collection.
	Find(...interface{}) Result

	// Truncate removes all elements on the collection and resets the
	// collection's IDs.
	Truncate() error

	// Name returns the name of the collection.
	Name() string
}
