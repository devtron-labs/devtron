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

// Collection defines methods to work with database tables or collections.
type Collection interface {

	// Name returns the name of the collection.
	Name() string

	// Session returns the Session that was used to create the collection
	// reference.
	Session() Session

	// Find defines a new result set.
	Find(...interface{}) Result

	Count() (uint64, error)

	// Insert inserts a new item into the collection, the type of this item could
	// be a map, a struct or pointer to either of them. If the call succeeds and
	// if the collection has a primary key, Insert returns the ID of the newly
	// added element as an `interface{}`. The underlying type of this ID depends
	// on both the database adapter and the column storing the ID.  The ID
	// returned by Insert() could be passed directly to Find() to retrieve the
	// newly added element.
	Insert(interface{}) (InsertResult, error)

	// InsertReturning is like Insert() but it takes a pointer to map or struct
	// and, if the operation succeeds, updates it with data from the newly
	// inserted row. If the database does not support transactions this method
	// returns db.ErrUnsupported.
	InsertReturning(interface{}) error

	// UpdateReturning takes a pointer to a map or struct and tries to update the
	// row the item is refering to. If the element is updated sucessfully,
	// UpdateReturning will fetch the row and update the fields of the passed
	// item.  If the database does not support transactions this method returns
	// db.ErrUnsupported
	UpdateReturning(interface{}) error

	// Exists returns true if the collection exists, false otherwise.
	Exists() (bool, error)

	// Truncate removes all elements on the collection.
	Truncate() error
}
