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

// Database is an interface that defines methods that must be satisfied by
// all database adapters.
type Database interface {
	// Driver returns the underlying driver the wrapper uses as an interface{}.
	//
	// In order to actually use the driver, the `interface{}` value needs to be
	// casted into the appropriate type.
	//
	// Example:
	//  internalSQLDriver := sess.Driver().(*sql.DB)
	Driver() interface{}

	// Open attempts to establish a connection with a DBMS.
	Open(ConnectionURL) error

	// Clone duplicates the current database session. Returns an error if the
	// clone did not succeed.
	// Clone() (Database, error)

	// Ping returns an error if the database manager could not be reached.
	Ping() error

	// Close closes the currently active connection to the database and clears
	// caches.
	Close() error

	// Collection returns a collection reference given a table name.
	Collection(string) Collection

	// Collections returns the names of all non-system tables on the database.
	Collections() ([]string, error)

	// Name returns the name of the active database.
	Name() string

	// ConnectionURL returns the data used to set up the adapter.
	ConnectionURL() ConnectionURL

	// ClearCache clears all the cache mechanisms the adapter is using.
	ClearCache()

	Settings
}
