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

// Package sqladapter provides common logic for SQL adapters.
package sqladapter

import (
	"database/sql"
	"database/sql/driver"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/internal/sqlbuilder"
)

// IsKeyValue reports whether v is a valid value for a primary key that can be
// used with Find(pKey).
func IsKeyValue(v interface{}) bool {
	if v == nil {
		return true
	}
	switch v.(type) {
	case int64, int, uint, uint64,
		[]int64, []int, []uint, []uint64,
		[]byte, []string,
		[]interface{},
		driver.Valuer:
		return true
	}
	return false
}

type sqlAdapterWrapper struct {
	adapter AdapterSession
}

func (w *sqlAdapterWrapper) OpenDSN(dsn db.ConnectionURL) (db.Session, error) {
	sess := NewSession(dsn, w.adapter)
	if err := sess.Open(); err != nil {
		return nil, err
	}
	return sess, nil
}

func (w *sqlAdapterWrapper) NewTx(sqlTx *sql.Tx) (sqlbuilder.Tx, error) {
	tx, err := NewTx(w.adapter, sqlTx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (w *sqlAdapterWrapper) New(sqlDB *sql.DB) (db.Session, error) {
	sess := NewSession(nil, w.adapter)
	if err := sess.BindDB(sqlDB); err != nil {
		return nil, err
	}
	return sess, nil
}

// RegisterAdapter registers a new SQL adapter.
func RegisterAdapter(name string, adapter AdapterSession) sqlbuilder.Adapter {
	z := &sqlAdapterWrapper{adapter}
	db.RegisterAdapter(name, sqlbuilder.NewCompatAdapter(z))
	return z
}
