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

// Iterator provides methods for iterating over query results.
type Iterator interface {
	// ResultMapper provides methods to retrieve and map results.
	ResultMapper

	// Scan dumps the current result into the given pointer variable pointers.
	Scan(dest ...interface{}) error

	// NextScan advances the iterator and performs Scan.
	NextScan(dest ...interface{}) error

	// ScanOne advances the iterator, performs Scan and closes the iterator.
	ScanOne(dest ...interface{}) error

	// Next dumps the current element into the given destination, which could be
	// a pointer to either a map or a struct.
	Next(dest ...interface{}) bool

	// Err returns the last error produced by the cursor.
	Err() error

	// Close closes the iterator and frees up the cursor.
	Close() error
}
