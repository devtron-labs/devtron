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
	"errors"
)

// Error messages
var (
	ErrMissingAdapter           = errors.New(`upper: missing adapter`)
	ErrAlreadyWithinTransaction = errors.New(`upper: already within a transaction`)
	ErrCollectionDoesNotExist   = errors.New(`upper: collection does not exist`)
	ErrExpectingNonNilModel     = errors.New(`upper: expecting non nil model`)
	ErrExpectingPointerToStruct = errors.New(`upper: expecting pointer to struct`)
	ErrGivingUpTryingToConnect  = errors.New(`upper: giving up trying to connect: too many clients`)
	ErrInvalidCollection        = errors.New(`upper: invalid collection`)
	ErrMissingCollectionName    = errors.New(`upper: missing collection name`)
	ErrMissingConditions        = errors.New(`upper: missing selector conditions`)
	ErrMissingConnURL           = errors.New(`upper: missing DSN`)
	ErrMissingDatabaseName      = errors.New(`upper: missing database name`)
	ErrNoMoreRows               = errors.New(`upper: no more rows in this result set`)
	ErrNotConnected             = errors.New(`upper: not connected to a database`)
	ErrNotImplemented           = errors.New(`upper: call not implemented`)
	ErrQueryIsPending           = errors.New(`upper: can't execute this instruction while the result set is still open`)
	ErrQueryLimitParam          = errors.New(`upper: a query can accept only one limit parameter`)
	ErrQueryOffsetParam         = errors.New(`upper: a query can accept only one offset parameter`)
	ErrQuerySortParam           = errors.New(`upper: a query can accept only one order-by parameter`)
	ErrSockerOrHost             = errors.New(`upper: you may connect either to a UNIX socket or a TCP address, but not both`)
	ErrTooManyClients           = errors.New(`upper: can't connect to database server: too many clients`)
	ErrUndefined                = errors.New(`upper: value is undefined`)
	ErrUnknownConditionType     = errors.New(`upper: arguments of type %T can't be used as constraints`)
	ErrUnsupported              = errors.New(`upper: action is not supported by the DBMS`)
	ErrUnsupportedDestination   = errors.New(`upper: unsupported destination type`)
	ErrUnsupportedType          = errors.New(`upper: type does not support marshaling`)
	ErrUnsupportedValue         = errors.New(`upper: value does not support unmarshaling`)
	ErrNilRecord                = errors.New(`upper: invalid item (nil)`)
	ErrRecordIDIsZero           = errors.New(`upper: item ID is not defined`)
	ErrMissingPrimaryKeys       = errors.New(`upper: collection %q has no primary keys`)
	ErrWarnSlowQuery            = errors.New(`upper: slow query`)
	ErrTransactionAborted       = errors.New(`upper: transaction was aborted`)
	ErrNotWithinTransaction     = errors.New(`upper: not within transaction`)
	ErrNotSupportedByAdapter    = errors.New(`upper: not supported by adapter`)
)
