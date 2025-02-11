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

// Store represents a data store.
type Store interface {
	Collection
}

// StoreSaver is an interface for data stores that defines a Save method that
// has the task of persisting a record.
type StoreSaver interface {
	Save(record Record) error
}

// StoreCreator is an interface for data stores that defines a Create method
// that has the task of creating a new record.
type StoreCreator interface {
	Create(record Record) error
}

// StoreDeleter is an interface for data stores that defines a Delete method
// that has the task of removing a record.
type StoreDeleter interface {
	Delete(record Record) error
}

// StoreUpdater is an interface for data stores that defines a Update method
// that has the task of updating a record.
type StoreUpdater interface {
	Update(record Record) error
}

// StoreGetter is an interface for data stores that defines a Get method that
// has the task of retrieving a record.
type StoreGetter interface {
	Get(record Record, id interface{}) error
}
