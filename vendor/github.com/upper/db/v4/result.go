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

import (
	"database/sql/driver"
)

// Result is an interface that defines methods for result sets.
type Result interface {

	// String returns the SQL statement to be used in the query.
	String() string

	// Limit defines the maximum number of results for this set. It only has
	// effect on `One()`, `All()` and `Next()`. A negative limit cancels any
	// previous limit settings.
	Limit(int) Result

	// Offset ignores the first n results. It only has effect on `One()`, `All()`
	// and `Next()`. A negative offset cancels any previous offset settings.
	Offset(int) Result

	// OrderBy receives one or more field names that define the order in which
	// elements will be returned in a query, field names may be prefixed with a
	// minus sign (-) indicating descending order, ascending order will be used
	// otherwise.
	OrderBy(...interface{}) Result

	// Select defines specific columns to be fetched on every column in the
	// result set.
	Select(...interface{}) Result

	// And adds more filtering conditions on top of the existing constraints.
	//
	//   res := col.Find(...).And(...)
	And(...interface{}) Result

	// GroupBy is used to group results that have the same value in the same column
	// or columns.
	GroupBy(...interface{}) Result

	// Delete deletes all items within the result set. `Offset()` and `Limit()`
	// are not honoured by `Delete()`.
	Delete() error

	// Update modifies all items within the result set. `Offset()` and `Limit()`
	// are not honoured by `Update()`.
	Update(interface{}) error

	// Count returns the number of items that match the set conditions.
	// `Offset()` and `Limit()` are not honoured by `Count()`
	Count() (uint64, error)

	// Exists returns true if at least one item on the collection exists. False
	// otherwise.
	Exists() (bool, error)

	// Next fetches the next result within the result set and dumps it into the
	// given pointer to struct or pointer to map. You must call
	// `Close()` after finishing using `Next()`.
	Next(ptrToStruct interface{}) bool

	// Err returns the last error that has happened with the result set, nil
	// otherwise.
	Err() error

	// One fetches the first result within the result set and dumps it into the
	// given pointer to struct or pointer to map. The result set is automatically
	// closed after picking the element, so there is no need to call Close()
	// after using One().
	One(ptrToStruct interface{}) error

	// All fetches all results within the result set and dumps them into the
	// given pointer to slice of maps or structs.  The result set is
	// automatically closed, so there is no need to call Close() after
	// using All().
	All(sliceOfStructs interface{}) error

	// Paginate splits the results of the query into pages containing pageSize
	// items. When using pagination previous settings for `Limit()` and
	// `Offset()` are ignored. Page numbering starts at 1.
	//
	// Use `Page()` to define the specific page to get results from.
	//
	// Example:
	//
	//   r = q.Paginate(12)
	//
	// You can provide constraints an order settings when using pagination:
	//
	// Example:
	//
	//   res := q.Where(conds).OrderBy("-id").Paginate(12)
	//   err := res.Page(4).All(&items)
	Paginate(pageSize uint) Result

	// Page makes the result set return results only from the page identified by
	// pageNumber. Page numbering starts from 1.
	//
	// Example:
	//
	//   r = q.Paginate(12).Page(4)
	Page(pageNumber uint) Result

	// Cursor defines the column that is going to be taken as basis for
	// cursor-based pagination.
	//
	// Example:
	//
	//   a = q.Paginate(10).Cursor("id")
	//   b = q.Paginate(12).Cursor("-id")
	//
	// You can set "" as cursorColumn to disable cursors.
	Cursor(cursorColumn string) Result

	// NextPage returns the next results page according to the cursor. It expects
	// a cursorValue, which is the value the cursor column had on the last item
	// of the current result set (lower bound).
	//
	// Example:
	//
	//   cursor = q.Paginate(12).Cursor("id")
	//   res = cursor.NextPage(items[len(items)-1].ID)
	//
	// Note that `NextPage()` requires a cursor, any column with an absolute
	// order (given two values one always precedes the other) can be a cursor.
	//
	// You can define the pagination order and add constraints to your result:
	//
	//	 cursor = q.Where(...).OrderBy("id").Paginate(10).Cursor("id")
	//   res = cursor.NextPage(lowerBound)
	NextPage(cursorValue interface{}) Result

	// PrevPage returns the previous results page according to the cursor. It
	// expects a cursorValue, which is the value the cursor column had on the
	// fist item of the current result set.
	//
	// Example:
	//
	//   current = current.PrevPage(items[0].ID)
	//
	// Note that PrevPage requires a cursor, any column with an absolute order
	// (given two values one always precedes the other) can be a cursor.
	//
	// You can define the pagination order and add constraints to your result:
	//
	//   cursor = q.Where(...).OrderBy("id").Paginate(10).Cursor("id")
	//   res = cursor.PrevPage(upperBound)
	PrevPage(cursorValue interface{}) Result

	// TotalPages returns the total number of pages the result set could produce.
	// If no pagination parameters have been set this value equals 1.
	TotalPages() (uint, error)

	// TotalEntries returns the total number of matching items in the result set.
	TotalEntries() (uint64, error)

	// Close closes the result set and frees all locked resources.
	Close() error
}

// InsertResult provides infomation about an insert operation.
type InsertResult interface {
	// ID returns the ID of the newly inserted record.
	ID() ID
}

type insertResult struct {
	id interface{}
}

func (r *insertResult) ID() ID {
	return r.id
}

// ConstraintValue satisfies adapter.ConstraintValuer
func (r *insertResult) ConstraintValue() interface{} {
	return r.id
}

// Value satisfies driver.Valuer
func (r *insertResult) Value() (driver.Value, error) {
	return r.id, nil
}

// NewInsertResult creates an InsertResult
func NewInsertResult(id interface{}) InsertResult {
	return &insertResult{id: id}
}

// ID represents a record ID
type ID interface{}

var _ = driver.Valuer(&insertResult{})
