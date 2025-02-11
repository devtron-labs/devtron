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

package db

// Record is the equivalence between concrete database schemas and Go values.
type Record interface {
	Store(sess Session) Store
}

// HasConstraints is an interface for records that defines a Constraints method
// that returns the record's own constraints.
type HasConstraints interface {
	Constraints() Cond
}

// Validator is an interface for records that defines an (optional) Validate
// method that is called before persisting a record (creating or updating).  If
// Validate returns an error the current operation is cancelled and rolled
// back.
type Validator interface {
	Validate() error
}

// BeforeCreateHook is an interface for records that defines an BeforeCreate
// method that is called before creating a record. If BeforeCreate returns an
// error the create process is cancelled and rolled back.
type BeforeCreateHook interface {
	BeforeCreate(Session) error
}

// AfterCreateHook is an interface for records that defines an AfterCreate
// method that is called after creating a record. If AfterCreate returns an
// error the create process is cancelled and rolled back.
type AfterCreateHook interface {
	AfterCreate(Session) error
}

// BeforeUpdateHook is an interface for records that defines a BeforeUpdate
// method that is called before updating a record. If BeforeUpdate returns an
// error the update process is cancelled and rolled back.
type BeforeUpdateHook interface {
	BeforeUpdate(Session) error
}

// AfterUpdateHook is an interface for records that defines an AfterUpdate
// method that is called after updating a record. If AfterUpdate returns an
// error the update process is cancelled and rolled back.
type AfterUpdateHook interface {
	AfterUpdate(Session) error
}

// BeforeDeleteHook is an interface for records that defines a BeforeDelete
// method that is called before removing a record. If BeforeDelete returns an
// error the delete process is cancelled and rolled back.
type BeforeDeleteHook interface {
	BeforeDelete(Session) error
}

// AfterDeleteHook is an interface for records that defines a AfterDelete
// method that is called after removing a record. If AfterDelete returns an
// error the delete process is cancelled and rolled back.
type AfterDeleteHook interface {
	AfterDelete(Session) error
}
