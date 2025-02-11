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

package postgresql

import (
	"database/sql"

	db "github.com/upper/db/v4"
	"github.com/upper/db/v4/internal/sqladapter"
	"github.com/upper/db/v4/internal/sqlbuilder"
)

// Adapter is the internal name of the adapter.
const Adapter = "postgresql"

var registeredAdapter = sqladapter.RegisterAdapter(Adapter, &database{})

// Open establishes a connection to the database server and returns a
// sqlbuilder.Session instance (which is compatible with db.Session).
func Open(connURL db.ConnectionURL) (db.Session, error) {
	return registeredAdapter.OpenDSN(connURL)
}

// NewTx creates a sqlbuilder.Tx instance by wrapping a *sql.Tx value.
func NewTx(sqlTx *sql.Tx) (sqlbuilder.Tx, error) {
	return registeredAdapter.NewTx(sqlTx)
}

// New creates a sqlbuilder.Sesion instance by wrapping a *sql.DB value.
func New(sqlDB *sql.DB) (db.Session, error) {
	return registeredAdapter.New(sqlDB)
}
