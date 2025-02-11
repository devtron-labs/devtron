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
	"context"
	"fmt"
)

// Selector represents a SELECT statement.
type Selector interface {
	// Columns defines which columns to retrive.
	//
	// You should call From() after Columns() if you want to query data from an
	// specific table.
	//
	//   s.Columns("name", "last_name").From(...)
	//
	// It is also possible to use an alias for the column, this could be handy if
	// you plan to use the alias later, use the "AS" keyword to denote an alias.
	//
	//   s.Columns("name AS n")
	//
	// or the shortcut:
	//
	//   s.Columns("name n")
	//
	// If you don't want the column to be escaped use the db.Raw
	// function.
	//
	//   s.Columns(db.Raw("MAX(id)"))
	//
	// The above statement is equivalent to:
	//
	//   s.Columns(db.Func("MAX", "id"))
	Columns(columns ...interface{}) Selector

	// From represents a FROM clause and is tipically used after Columns().
	//
	// FROM defines from which table data is going to be retrieved
	//
	//   s.Columns(...).From("people")
	//
	// It is also possible to use an alias for the table, this could be handy if
	// you plan to use the alias later:
	//
	//   s.Columns(...).From("people AS p").Where("p.name = ?", ...)
	//
	// Or with the shortcut:
	//
	//   s.Columns(...).From("people p").Where("p.name = ?", ...)
	From(tables ...interface{}) Selector

	// Distict represents a DISTINCT clause
	//
	// DISTINCT is used to ask the database to return only values that are
	// different.
	Distinct(columns ...interface{}) Selector

	// As defines an alias for a table.
	As(string) Selector

	// Where specifies the conditions that columns must match in order to be
	// retrieved.
	//
	// Where accepts raw strings and fmt.Stringer to define conditions and
	// interface{} to specify parameters. Be careful not to embed any parameters
	// within the SQL part as that could lead to security problems. You can use
	// que question mark (?) as placeholder for parameters.
	//
	//   s.Where("name = ?", "max")
	//
	//   s.Where("name = ? AND last_name = ?", "Mary", "Doe")
	//
	//   s.Where("last_name IS NULL")
	//
	// You can also use other types of parameters besides only strings, like:
	//
	//   s.Where("online = ? AND last_logged <= ?", true, time.Now())
	//
	// and Where() will transform them into strings before feeding them to the
	// database.
	//
	// When an unknown type is provided, Where() will first try to match it with
	// the Marshaler interface, then with fmt.Stringer and finally, if the
	// argument does not satisfy any of those interfaces Where() will use
	// fmt.Sprintf("%v", arg) to transform the type into a string.
	//
	// Subsequent calls to Where() will overwrite previously set conditions, if
	// you want these new conditions to be appended use And() instead.
	Where(conds ...interface{}) Selector

	// And appends more constraints to the WHERE clause without overwriting
	// conditions that have been already set.
	And(conds ...interface{}) Selector

	// GroupBy represents a GROUP BY statement.
	//
	// GROUP BY defines which columns should be used to aggregate and group
	// results.
	//
	//   s.GroupBy("country_id")
	//
	// GroupBy accepts more than one column:
	//
	//   s.GroupBy("country_id", "city_id")
	GroupBy(columns ...interface{}) Selector

	// Having(...interface{}) Selector

	// OrderBy represents a ORDER BY statement.
	//
	// ORDER BY is used to define which columns are going to be used to sort
	// results.
	//
	// Use the column name to sort results in ascendent order.
	//
	//   // "last_name" ASC
	//   s.OrderBy("last_name")
	//
	// Prefix the column name with the minus sign (-) to sort results in
	// descendent order.
	//
	//   // "last_name" DESC
	//   s.OrderBy("-last_name")
	//
	// If you would rather be very explicit, you can also use ASC and DESC.
	//
	//   s.OrderBy("last_name ASC")
	//
	//   s.OrderBy("last_name DESC", "name ASC")
	OrderBy(columns ...interface{}) Selector

	// Join represents a JOIN statement.
	//
	// JOIN statements are used to define external tables that the user wants to
	// include as part of the result.
	//
	// You can use the On() method after Join() to define the conditions of the
	// join.
	//
	//   s.Join("author").On("author.id = book.author_id")
	//
	// If you don't specify conditions for the join, a NATURAL JOIN will be used.
	//
	// On() accepts the same arguments as Where()
	//
	// You can also use Using() after Join().
	//
	//   s.Join("employee").Using("department_id")
	Join(table ...interface{}) Selector

	// FullJoin is like Join() but with FULL JOIN.
	FullJoin(...interface{}) Selector

	// CrossJoin is like Join() but with CROSS JOIN.
	CrossJoin(...interface{}) Selector

	// RightJoin is like Join() but with RIGHT JOIN.
	RightJoin(...interface{}) Selector

	// LeftJoin is like Join() but with LEFT JOIN.
	LeftJoin(...interface{}) Selector

	// Using represents the USING clause.
	//
	// USING is used to specifiy columns to join results.
	//
	//   s.LeftJoin(...).Using("country_id")
	Using(...interface{}) Selector

	// On represents the ON clause.
	//
	// ON is used to define conditions on a join.
	//
	//   s.Join(...).On("b.author_id = a.id")
	On(...interface{}) Selector

	// Limit represents the LIMIT parameter.
	//
	// LIMIT defines the maximum number of rows to return from the table.  A
	// negative limit cancels any previous limit settings.
	//
	//  s.Limit(42)
	Limit(int) Selector

	// Offset represents the OFFSET parameter.
	//
	// OFFSET defines how many results are going to be skipped before starting to
	// return results. A negative offset cancels any previous offset settings.
	//
	// s.Offset(56)
	Offset(int) Selector

	// Amend lets you alter the query's text just before sending it to the
	// database server.
	Amend(func(queryIn string) (queryOut string)) Selector

	// Paginate returns a paginator that can display a paginated lists of items.
	// Paginators ignore previous Offset and Limit settings. Page numbering
	// starts at 1.
	Paginate(uint) Paginator

	// Iterator provides methods to iterate over the results returned by the
	// Selector.
	Iterator() Iterator

	// IteratorContext provides methods to iterate over the results returned by
	// the Selector.
	IteratorContext(ctx context.Context) Iterator

	// SQLPreparer provides methods for creating prepared statements.
	SQLPreparer

	// SQLGetter provides methods to compile and execute a query that returns
	// results.
	SQLGetter

	// ResultMapper provides methods to retrieve and map results.
	ResultMapper

	// fmt.Stringer provides `String() string`, you can use `String()` to compile
	// the `Selector` into a string.
	fmt.Stringer

	// Arguments returns the arguments that are prepared for this query.
	Arguments() []interface{}
}

// Inserter represents an INSERT statement.
type Inserter interface {
	// Columns represents the COLUMNS clause.
	//
	// COLUMNS defines the columns that we are going to provide values for.
	//
	//   i.Columns("name", "last_name").Values(...)
	Columns(...string) Inserter

	// Values represents the VALUES clause.
	//
	// VALUES defines the values of the columns.
	//
	//   i.Columns(...).Values("María", "Méndez")
	//
	//   i.Values(map[string][string]{"name": "María"})
	Values(...interface{}) Inserter

	// Arguments returns the arguments that are prepared for this query.
	Arguments() []interface{}

	// Returning represents a RETURNING clause.
	//
	// RETURNING specifies which columns should be returned after INSERT.
	//
	// RETURNING may not be supported by all SQL databases.
	Returning(columns ...string) Inserter

	// Iterator provides methods to iterate over the results returned by the
	// Inserter. This is only possible when using Returning().
	Iterator() Iterator

	// IteratorContext provides methods to iterate over the results returned by
	// the Inserter. This is only possible when using Returning().
	IteratorContext(ctx context.Context) Iterator

	// Amend lets you alter the query's text just before sending it to the
	// database server.
	Amend(func(queryIn string) (queryOut string)) Inserter

	// Batch provies a BatchInserter that can be used to insert many elements at
	// once by issuing several calls to Values(). It accepts a size parameter
	// which defines the batch size. If size is < 1, the batch size is set to 1.
	Batch(size int) BatchInserter

	// SQLExecer provides the Exec method.
	SQLExecer

	// SQLPreparer provides methods for creating prepared statements.
	SQLPreparer

	// SQLGetter provides methods to return query results from INSERT statements
	// that support such feature (e.g.: queries with Returning).
	SQLGetter

	// fmt.Stringer provides `String() string`, you can use `String()` to compile
	// the `Inserter` into a string.
	fmt.Stringer
}

// Deleter represents a DELETE statement.
type Deleter interface {
	// Where represents the WHERE clause.
	//
	// See Selector.Where for documentation and usage examples.
	Where(...interface{}) Deleter

	// And appends more constraints to the WHERE clause without overwriting
	// conditions that have been already set.
	And(conds ...interface{}) Deleter

	// Limit represents the LIMIT clause.
	//
	// See Selector.Limit for documentation and usage examples.
	Limit(int) Deleter

	// Amend lets you alter the query's text just before sending it to the
	// database server.
	Amend(func(queryIn string) (queryOut string)) Deleter

	// SQLPreparer provides methods for creating prepared statements.
	SQLPreparer

	// SQLExecer provides the Exec method.
	SQLExecer

	// fmt.Stringer provides `String() string`, you can use `String()` to compile
	// the `Inserter` into a string.
	fmt.Stringer

	// Arguments returns the arguments that are prepared for this query.
	Arguments() []interface{}
}

// Updater represents an UPDATE statement.
type Updater interface {
	// Set represents the SET clause.
	Set(...interface{}) Updater

	// Where represents the WHERE clause.
	//
	// See Selector.Where for documentation and usage examples.
	Where(...interface{}) Updater

	// And appends more constraints to the WHERE clause without overwriting
	// conditions that have been already set.
	And(conds ...interface{}) Updater

	// Limit represents the LIMIT parameter.
	//
	// See Selector.Limit for documentation and usage examples.
	Limit(int) Updater

	// SQLPreparer provides methods for creating prepared statements.
	SQLPreparer

	// SQLExecer provides the Exec method.
	SQLExecer

	// fmt.Stringer provides `String() string`, you can use `String()` to compile
	// the `Inserter` into a string.
	fmt.Stringer

	// Arguments returns the arguments that are prepared for this query.
	Arguments() []interface{}

	// Amend lets you alter the query's text just before sending it to the
	// database server.
	Amend(func(queryIn string) (queryOut string)) Updater
}

// Paginator provides tools for splitting the results of a query into chunks
// containing a fixed number of items.
type Paginator interface {
	// Page sets the page number.
	Page(uint) Paginator

	// Cursor defines the column that is going to be taken as basis for
	// cursor-based pagination.
	//
	// Example:
	//
	//   a = q.Paginate(10).Cursor("id")
	//	 b = q.Paginate(12).Cursor("-id")
	//
	// You can set "" as cursorColumn to disable cursors.
	Cursor(cursorColumn string) Paginator

	// NextPage returns the next page according to the cursor. It expects a
	// cursorValue, which is the value the cursor column has on the last item of
	// the current result set (lower bound).
	//
	// Example:
	//
	//   p = q.NextPage(items[len(items)-1].ID)
	NextPage(cursorValue interface{}) Paginator

	// PrevPage returns the previous page according to the cursor. It expects a
	// cursorValue, which is the value the cursor column has on the fist item of
	// the current result set (upper bound).
	//
	// Example:
	//
	//   p = q.PrevPage(items[0].ID)
	PrevPage(cursorValue interface{}) Paginator

	// TotalPages returns the total number of pages in the query.
	TotalPages() (uint, error)

	// TotalEntries returns the total number of entries in the query.
	TotalEntries() (uint64, error)

	// SQLPreparer provides methods for creating prepared statements.
	SQLPreparer

	// SQLGetter provides methods to compile and execute a query that returns
	// results.
	SQLGetter

	// Iterator provides methods to iterate over the results returned by the
	// Selector.
	Iterator() Iterator

	// IteratorContext provides methods to iterate over the results returned by
	// the Selector.
	IteratorContext(ctx context.Context) Iterator

	// ResultMapper provides methods to retrieve and map results.
	ResultMapper

	// fmt.Stringer provides `String() string`, you can use `String()` to compile
	// the `Selector` into a string.
	fmt.Stringer

	// Arguments returns the arguments that are prepared for this query.
	Arguments() []interface{}
}

// ResultMapper defined methods for a result mapper.
type ResultMapper interface {
	// All dumps all the results into the given slice, All() expects a pointer to
	// slice of maps or structs.
	//
	// The behaviour of One() extends to each one of the results.
	All(destSlice interface{}) error

	// One maps the row that is in the current query cursor into the
	// given interface, which can be a pointer to either a map or a
	// struct.
	//
	// If dest is a pointer to map, each one of the columns will create a new map
	// key and the values of the result will be set as values for the keys.
	//
	// Depending on the type of map key and value, the results columns and values
	// may need to be transformed.
	//
	// If dest if a pointer to struct, each one of the fields will be tested for
	// a `db` tag which defines the column mapping. The value of the result will
	// be set as the value of the field.
	One(dest interface{}) error
}

// BatchInserter provides an interface to do massive insertions in batches.
type BatchInserter interface {
	// Values pushes column values to be inserted as part of the batch.
	Values(...interface{}) BatchInserter

	// NextResult dumps the next slice of results to dst, which can mean having
	// the IDs of all inserted elements in the batch.
	NextResult(dst interface{}) bool

	// Done signals that no more elements are going to be added.
	Done()

	// Wait blocks until the whole batch is executed.
	Wait() error

	// Err returns the last error that happened while executing the batch (or nil
	// if no error happened).
	Err() error
}
