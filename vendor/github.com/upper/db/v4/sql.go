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
	"database/sql"
)

// SQL defines methods that can be used to build a SQL query with chainable
// method calls.
//
// Queries are immutable, so every call to any method will return a new
// pointer, if you want to build a query using variables you need to reassign
// them, like this:
//
//  a = builder.Select("name").From("foo") // "a" is created
//
//  a.Where(...) // No effect, the value returned from Where is ignored.
//
//  a = a.Where(...) // "a" is reassigned and points to a different address.
//
type SQL interface {

	// Select initializes and returns a Selector, it accepts column names as
	// parameters.
	//
	// The returned Selector does not initially point to any table, a call to
	// From() is required after Select() to complete a valid query.
	//
	// Example:
	//
	//  q := sqlbuilder.Select("first_name", "last_name").From("people").Where(...)
	Select(columns ...interface{}) Selector

	// SelectFrom creates a Selector that selects all columns (like SELECT *)
	// from the given table.
	//
	// Example:
	//
	//  q := sqlbuilder.SelectFrom("people").Where(...)
	SelectFrom(table ...interface{}) Selector

	// InsertInto prepares and returns an Inserter targeted at the given table.
	//
	// Example:
	//
	//   q := sqlbuilder.InsertInto("books").Columns(...).Values(...)
	InsertInto(table string) Inserter

	// DeleteFrom prepares a Deleter targeted at the given table.
	//
	// Example:
	//
	//  q := sqlbuilder.DeleteFrom("tasks").Where(...)
	DeleteFrom(table string) Deleter

	// Update prepares and returns an Updater targeted at the given table.
	//
	// Example:
	//
	//  q := sqlbuilder.Update("profile").Set(...).Where(...)
	Update(table string) Updater

	// Exec executes a SQL query that does not return any rows, like sql.Exec.
	// Queries can be either strings or upper-db statements.
	//
	// Example:
	//
	//  sqlbuilder.Exec(`INSERT INTO books (title) VALUES("La Ciudad y los Perros")`)
	Exec(query interface{}, args ...interface{}) (sql.Result, error)

	// ExecContext executes a SQL query that does not return any rows, like sql.ExecContext.
	// Queries can be either strings or upper-db statements.
	//
	// Example:
	//
	//  sqlbuilder.ExecContext(ctx, `INSERT INTO books (title) VALUES(?)`, "La Ciudad y los Perros")
	ExecContext(ctx context.Context, query interface{}, args ...interface{}) (sql.Result, error)

	// Prepare creates a prepared statement for later queries or executions. The
	// caller must call the statement's Close method when the statement is no
	// longer needed.
	Prepare(query interface{}) (*sql.Stmt, error)

	// Prepare creates a prepared statement on the guiven context for later
	// queries or executions. The caller must call the statement's Close method
	// when the statement is no longer needed.
	PrepareContext(ctx context.Context, query interface{}) (*sql.Stmt, error)

	// Query executes a SQL query that returns rows, like sql.Query.  Queries can
	// be either strings or upper-db statements.
	//
	// Example:
	//
	//  sqlbuilder.Query(`SELECT * FROM people WHERE name = "Mateo"`)
	Query(query interface{}, args ...interface{}) (*sql.Rows, error)

	// QueryContext executes a SQL query that returns rows, like
	// sql.QueryContext.  Queries can be either strings or upper-db statements.
	//
	// Example:
	//
	//  sqlbuilder.QueryContext(ctx, `SELECT * FROM people WHERE name = ?`, "Mateo")
	QueryContext(ctx context.Context, query interface{}, args ...interface{}) (*sql.Rows, error)

	// QueryRow executes a SQL query that returns one row, like sql.QueryRow.
	// Queries can be either strings or upper-db statements.
	//
	// Example:
	//
	//  sqlbuilder.QueryRow(`SELECT * FROM people WHERE name = "Haruki" AND last_name = "Murakami" LIMIT 1`)
	QueryRow(query interface{}, args ...interface{}) (*sql.Row, error)

	// QueryRowContext executes a SQL query that returns one row, like
	// sql.QueryRowContext.  Queries can be either strings or upper-db statements.
	//
	// Example:
	//
	//  sqlbuilder.QueryRowContext(ctx, `SELECT * FROM people WHERE name = "Haruki" AND last_name = "Murakami" LIMIT 1`)
	QueryRowContext(ctx context.Context, query interface{}, args ...interface{}) (*sql.Row, error)

	// Iterator executes a SQL query that returns rows and creates an Iterator
	// with it.
	//
	// Example:
	//
	//  sqlbuilder.Iterator(`SELECT * FROM people WHERE name LIKE "M%"`)
	Iterator(query interface{}, args ...interface{}) Iterator

	// IteratorContext executes a SQL query that returns rows and creates an Iterator
	// with it.
	//
	// Example:
	//
	//  sqlbuilder.IteratorContext(ctx, `SELECT * FROM people WHERE name LIKE "M%"`)
	IteratorContext(ctx context.Context, query interface{}, args ...interface{}) Iterator

	// NewIterator converts a *sql.Rows value into an Iterator.
	NewIterator(rows *sql.Rows) Iterator

	// NewIteratorContext converts a *sql.Rows value into an Iterator.
	NewIteratorContext(ctx context.Context, rows *sql.Rows) Iterator
}

// SQLExecer provides methods for executing statements that do not return
// results.
type SQLExecer interface {
	// Exec executes a statement and returns sql.Result.
	Exec() (sql.Result, error)

	// ExecContext executes a statement and returns sql.Result.
	ExecContext(context.Context) (sql.Result, error)
}

// SQLPreparer provides the Prepare and PrepareContext methods for creating
// prepared statements.
type SQLPreparer interface {
	// Prepare creates a prepared statement.
	Prepare() (*sql.Stmt, error)

	// PrepareContext creates a prepared statement.
	PrepareContext(context.Context) (*sql.Stmt, error)
}

// SQLGetter provides methods for executing statements that return results.
type SQLGetter interface {
	// Query returns *sql.Rows.
	Query() (*sql.Rows, error)

	// QueryContext returns *sql.Rows.
	QueryContext(context.Context) (*sql.Rows, error)

	// QueryRow returns only one row.
	QueryRow() (*sql.Row, error)

	// QueryRowContext returns only one row.
	QueryRowContext(ctx context.Context) (*sql.Row, error)
}

// SQLEngine represents a SQL engine that can execute SQL queries. This is
// compatible with *sql.DB.
type SQLEngine interface {
	Exec(string, ...interface{}) (sql.Result, error)
	Prepare(string) (*sql.Stmt, error)
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row

	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}
