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

// Package postgresql provides an adapter for PostgreSQL.
// See https://github.com/upper/db/adapter/postgresql for documentation,
// particularities and usage examples.
package postgresql

import (
	"fmt"
	"strings"

	db "github.com/upper/db/v4"
	"github.com/upper/db/v4/internal/sqladapter"
	"github.com/upper/db/v4/internal/sqladapter/exql"
	"github.com/upper/db/v4/internal/sqlbuilder"
)

type database struct {
}

func (*database) Template() *exql.Template {
	return template
}

func (*database) Collections(sess sqladapter.Session) (collections []string, err error) {
	q := sess.SQL().
		Select("table_name").
		From("information_schema.tables").
		Where("table_schema = ?", "public")

	iter := q.Iterator()
	defer iter.Close()

	for iter.Next() {
		var name string
		if err := iter.Scan(&name); err != nil {
			return nil, err
		}
		collections = append(collections, name)
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	return collections, nil
}

func (*database) ConvertValue(in interface{}) interface{} {
	switch v := in.(type) {
	case *[]int64:
		return (*Int64Array)(v)
	case *[]string:
		return (*StringArray)(v)
	case *[]float64:
		return (*Float64Array)(v)
	case *[]bool:
		return (*BoolArray)(v)
	case *map[string]interface{}:
		return (*JSONBMap)(v)

	case []int64:
		return (*Int64Array)(&v)
	case []string:
		return (*StringArray)(&v)
	case []float64:
		return (*Float64Array)(&v)
	case []bool:
		return (*BoolArray)(&v)
	case map[string]interface{}:
		return (*JSONBMap)(&v)

	}
	return in
}

func (*database) CompileStatement(sess sqladapter.Session, stmt *exql.Statement, args []interface{}) (string, []interface{}, error) {
	compiled, err := stmt.Compile(template)
	if err != nil {
		return "", nil, err
	}

	query, args := sqlbuilder.Preprocess(compiled, args)
	query = string(sqladapter.ReplaceWithDollarSign([]byte(query)))
	return query, args, nil
}

func (*database) Err(err error) error {
	if err != nil {
		s := err.Error()
		// These errors are not exported so we have to check them by they string value.
		if strings.Contains(s, `too many clients`) || strings.Contains(s, `remaining connection slots are reserved`) || strings.Contains(s, `too many open`) {
			return db.ErrTooManyClients
		}
	}
	return err
}

func (*database) NewCollection() sqladapter.CollectionAdapter {
	return &collectionAdapter{}
}

func (*database) LookupName(sess sqladapter.Session) (string, error) {
	q := sess.SQL().
		Select(db.Raw("CURRENT_DATABASE() AS name"))

	iter := q.Iterator()
	defer iter.Close()

	if iter.Next() {
		var name string
		if err := iter.Scan(&name); err != nil {
			return "", err
		}
		return name, nil
	}

	return "", iter.Err()
}

func (*database) TableExists(sess sqladapter.Session, name string) error {
	q := sess.SQL().
		Select("table_name").
		From("information_schema.tables").
		Where("table_catalog = ? AND table_name = ?", sess.Name(), name)

	iter := q.Iterator()
	defer iter.Close()

	if iter.Next() {
		var name string
		if err := iter.Scan(&name); err != nil {
			return err
		}
		return nil
	}
	if err := iter.Err(); err != nil {
		return err
	}

	return db.ErrCollectionDoesNotExist
}

func (*database) PrimaryKeys(sess sqladapter.Session, tableName string) ([]string, error) {
	q := sess.SQL().
		Select("pg_attribute.attname AS pkey").
		From("pg_index", "pg_class", "pg_attribute").
		Where(`
			pg_class.oid = '` + quotedTableName(tableName) + `'::regclass
			AND indrelid = pg_class.oid
			AND pg_attribute.attrelid = pg_class.oid
			AND pg_attribute.attnum = ANY(pg_index.indkey)
			AND indisprimary
		`).OrderBy("pkey")

	iter := q.Iterator()
	defer iter.Close()

	pk := []string{}

	for iter.Next() {
		var k string
		if err := iter.Scan(&k); err != nil {
			return nil, err
		}
		pk = append(pk, k)
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	return pk, nil
}

// quotedTableName returns a valid regclass name for both regular tables and
// for schemas.
func quotedTableName(s string) string {
	chunks := strings.Split(s, ".")
	for i := range chunks {
		chunks[i] = fmt.Sprintf("%q", chunks[i])
	}
	return strings.Join(chunks, ".")
}
