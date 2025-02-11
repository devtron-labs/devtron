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

// Package sqlbuilder provides tools for building custom SQL queries.
package sqlbuilder

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	db "github.com/upper/db/v4"
	"github.com/upper/db/v4/internal/adapter"
	"github.com/upper/db/v4/internal/reflectx"
	"github.com/upper/db/v4/internal/sqladapter/compat"
	"github.com/upper/db/v4/internal/sqladapter/exql"
)

// MapOptions represents options for the mapper.
type MapOptions struct {
	IncludeZeroed bool
	IncludeNil    bool
}

var defaultMapOptions = MapOptions{
	IncludeZeroed: false,
	IncludeNil:    false,
}

type hasPaginator interface {
	Paginator() (db.Paginator, error)
}

type isCompilable interface {
	Compile() (string, error)
	Arguments() []interface{}
}

type hasIsZero interface {
	IsZero() bool
}

type iterator struct {
	sess   exprDB
	cursor *sql.Rows // This is the main query cursor. It starts as a nil value.
	err    error
}

type fieldValue struct {
	fields []string
	values []interface{}
}

var (
	sqlPlaceholder = &exql.Raw{Value: `?`}
)

var (
	errDeprecatedJSONBTag = errors.New(`Tag "jsonb" is deprecated. See "PostgreSQL: jsonb tag" at https://github.com/upper/db/releases/tag/v3.4.0`)
)

type exprDB interface {
	StatementExec(ctx context.Context, stmt *exql.Statement, args ...interface{}) (sql.Result, error)
	StatementPrepare(ctx context.Context, stmt *exql.Statement) (*sql.Stmt, error)
	StatementQuery(ctx context.Context, stmt *exql.Statement, args ...interface{}) (*sql.Rows, error)
	StatementQueryRow(ctx context.Context, stmt *exql.Statement, args ...interface{}) (*sql.Row, error)

	Context() context.Context
}

type sqlBuilder struct {
	sess exprDB
	t    *templateWithUtils
}

// WithSession returns a query builder that is bound to the given database session.
func WithSession(sess interface{}, t *exql.Template) db.SQL {
	if sqlDB, ok := sess.(*sql.DB); ok {
		sess = sqlDB
	}
	return &sqlBuilder{
		sess: sess.(exprDB), // Let it panic, it will show the developer an informative error.
		t:    newTemplateWithUtils(t),
	}
}

// WithTemplate returns a builder that is based on the given template.
func WithTemplate(t *exql.Template) db.SQL {
	return &sqlBuilder{
		t: newTemplateWithUtils(t),
	}
}

func (b *sqlBuilder) NewIteratorContext(ctx context.Context, rows *sql.Rows) db.Iterator {
	return &iterator{b.sess, rows, nil}
}

func (b *sqlBuilder) NewIterator(rows *sql.Rows) db.Iterator {
	return b.NewIteratorContext(b.sess.Context(), rows)
}

func (b *sqlBuilder) Iterator(query interface{}, args ...interface{}) db.Iterator {
	return b.IteratorContext(b.sess.Context(), query, args...)
}

func (b *sqlBuilder) IteratorContext(ctx context.Context, query interface{}, args ...interface{}) db.Iterator {
	rows, err := b.QueryContext(ctx, query, args...)
	return &iterator{b.sess, rows, err}
}

func (b *sqlBuilder) Prepare(query interface{}) (*sql.Stmt, error) {
	return b.PrepareContext(b.sess.Context(), query)
}

func (b *sqlBuilder) PrepareContext(ctx context.Context, query interface{}) (*sql.Stmt, error) {
	switch q := query.(type) {
	case *exql.Statement:
		return b.sess.StatementPrepare(ctx, q)
	case string:
		return b.sess.StatementPrepare(ctx, exql.RawSQL(q))
	case *adapter.RawExpr:
		return b.PrepareContext(ctx, q.Raw())
	default:
		return nil, fmt.Errorf("unsupported query type %T", query)
	}
}

func (b *sqlBuilder) Exec(query interface{}, args ...interface{}) (sql.Result, error) {
	return b.ExecContext(b.sess.Context(), query, args...)
}

func (b *sqlBuilder) ExecContext(ctx context.Context, query interface{}, args ...interface{}) (sql.Result, error) {
	switch q := query.(type) {
	case *exql.Statement:
		return b.sess.StatementExec(ctx, q, args...)
	case string:
		return b.sess.StatementExec(ctx, exql.RawSQL(q), args...)
	case *adapter.RawExpr:
		return b.ExecContext(ctx, q.Raw(), q.Arguments()...)
	default:
		return nil, fmt.Errorf("unsupported query type %T", query)
	}
}

func (b *sqlBuilder) Query(query interface{}, args ...interface{}) (*sql.Rows, error) {
	return b.QueryContext(b.sess.Context(), query, args...)
}

func (b *sqlBuilder) QueryContext(ctx context.Context, query interface{}, args ...interface{}) (*sql.Rows, error) {
	switch q := query.(type) {
	case *exql.Statement:
		return b.sess.StatementQuery(ctx, q, args...)
	case string:
		return b.sess.StatementQuery(ctx, exql.RawSQL(q), args...)
	case *adapter.RawExpr:
		return b.QueryContext(ctx, q.Raw(), q.Arguments()...)
	default:
		return nil, fmt.Errorf("unsupported query type %T", query)
	}
}

func (b *sqlBuilder) QueryRow(query interface{}, args ...interface{}) (*sql.Row, error) {
	return b.QueryRowContext(b.sess.Context(), query, args...)
}

func (b *sqlBuilder) QueryRowContext(ctx context.Context, query interface{}, args ...interface{}) (*sql.Row, error) {
	switch q := query.(type) {
	case *exql.Statement:
		return b.sess.StatementQueryRow(ctx, q, args...)
	case string:
		return b.sess.StatementQueryRow(ctx, exql.RawSQL(q), args...)
	case *adapter.RawExpr:
		return b.QueryRowContext(ctx, q.Raw(), q.Arguments()...)
	default:
		return nil, fmt.Errorf("unsupported query type %T", query)
	}
}

func (b *sqlBuilder) SelectFrom(table ...interface{}) db.Selector {
	qs := &selector{
		builder: b,
	}
	return qs.From(table...)
}

func (b *sqlBuilder) Select(columns ...interface{}) db.Selector {
	qs := &selector{
		builder: b,
	}
	return qs.Columns(columns...)
}

func (b *sqlBuilder) InsertInto(table string) db.Inserter {
	qi := &inserter{
		builder: b,
	}
	return qi.Into(table)
}

func (b *sqlBuilder) DeleteFrom(table string) db.Deleter {
	qd := &deleter{
		builder: b,
	}
	return qd.setTable(table)
}

func (b *sqlBuilder) Update(table string) db.Updater {
	qu := &updater{
		builder: b,
	}
	return qu.setTable(table)
}

// Map receives a pointer to map or struct and maps it to columns and values.
func Map(item interface{}, options *MapOptions) ([]string, []interface{}, error) {
	var fv fieldValue
	if options == nil {
		options = &defaultMapOptions
	}

	itemV := reflect.ValueOf(item)
	if !itemV.IsValid() {
		return nil, nil, nil
	}

	itemT := itemV.Type()

	if itemT.Kind() == reflect.Ptr {
		// Single dereference. Just in case the user passes a pointer to struct
		// instead of a struct.
		item = itemV.Elem().Interface()
		itemV = reflect.ValueOf(item)
		itemT = itemV.Type()
	}

	switch itemT.Kind() {
	case reflect.Struct:
		fieldMap := Mapper.TypeMap(itemT).Names
		nfields := len(fieldMap)

		fv.values = make([]interface{}, 0, nfields)
		fv.fields = make([]string, 0, nfields)

		for _, fi := range fieldMap {

			// Check for deprecated JSONB tag
			if _, hasJSONBTag := fi.Options["jsonb"]; hasJSONBTag {
				return nil, nil, errDeprecatedJSONBTag
			}

			// Field options
			_, tagOmitEmpty := fi.Options["omitempty"]

			fld := reflectx.FieldByIndexesReadOnly(itemV, fi.Index)
			if fld.Kind() == reflect.Ptr && fld.IsNil() {
				if tagOmitEmpty && !options.IncludeNil {
					continue
				}
				fv.fields = append(fv.fields, fi.Name)
				if tagOmitEmpty {
					fv.values = append(fv.values, sqlDefault)
				} else {
					fv.values = append(fv.values, nil)
				}
				continue
			}

			value := fld.Interface()

			isZero := false
			if t, ok := fld.Interface().(hasIsZero); ok {
				if t.IsZero() {
					isZero = true
				}
			} else if fld.Kind() == reflect.Array || fld.Kind() == reflect.Slice {
				if fld.Len() == 0 {
					isZero = true
				}
			} else if reflect.DeepEqual(fi.Zero.Interface(), value) {
				isZero = true
			}

			if isZero && tagOmitEmpty && !options.IncludeZeroed {
				continue
			}

			fv.fields = append(fv.fields, fi.Name)
			v, err := marshal(value)
			if err != nil {
				return nil, nil, err
			}
			if isZero && tagOmitEmpty {
				v = sqlDefault
			}
			fv.values = append(fv.values, v)
		}

	case reflect.Map:
		nfields := itemV.Len()
		fv.values = make([]interface{}, nfields)
		fv.fields = make([]string, nfields)
		mkeys := itemV.MapKeys()

		for i, keyV := range mkeys {
			valv := itemV.MapIndex(keyV)
			fv.fields[i] = fmt.Sprintf("%v", keyV.Interface())

			v, err := marshal(valv.Interface())
			if err != nil {
				return nil, nil, err
			}

			fv.values[i] = v
		}
	default:
		return nil, nil, ErrExpectingPointerToEitherMapOrStruct
	}

	sort.Sort(&fv)

	return fv.fields, fv.values, nil
}

func columnFragments(columns []interface{}) ([]exql.Fragment, []interface{}, error) {
	f := make([]exql.Fragment, len(columns))
	args := []interface{}{}

	for i := range columns {
		switch v := columns[i].(type) {
		case hasPaginator:
			p, err := v.Paginator()
			if err != nil {
				return nil, nil, err
			}

			q, a := Preprocess(p.String(), p.Arguments())

			f[i] = &exql.Raw{Value: "(" + q + ")"}
			args = append(args, a...)
		case isCompilable:
			c, err := v.Compile()
			if err != nil {
				return nil, nil, err
			}
			q, a := Preprocess(c, v.Arguments())
			if _, ok := v.(db.Selector); ok {
				q = "(" + q + ")"
			}
			f[i] = &exql.Raw{Value: q}
			args = append(args, a...)
		case *adapter.FuncExpr:
			fnName, fnArgs := v.Name(), v.Arguments()
			if len(fnArgs) == 0 {
				fnName = fnName + "()"
			} else {
				fnName = fnName + "(?" + strings.Repeat(", ?", len(fnArgs)-1) + ")"
			}
			fnName, fnArgs = Preprocess(fnName, fnArgs)
			f[i] = &exql.Raw{Value: fnName}
			args = append(args, fnArgs...)
		case *adapter.RawExpr:
			q, a := Preprocess(v.Raw(), v.Arguments())
			f[i] = &exql.Raw{Value: q}
			args = append(args, a...)
		case exql.Fragment:
			f[i] = v
		case string:
			f[i] = exql.ColumnWithName(v)
		case fmt.Stringer:
			f[i] = exql.ColumnWithName(v.String())
		default:
			var err error
			f[i], err = exql.NewRawValue(columns[i])
			if err != nil {
				return nil, nil, fmt.Errorf("unexpected argument type %T for Select() argument: %w", v, err)
			}
		}
	}
	return f, args, nil
}

func prepareQueryForDisplay(in string) string {
	out := make([]byte, 0, len(in))

	offset := 0
	whitespace := true
	placeholders := 1

	for i := 0; i < len(in); i++ {
		if in[i] == ' ' || in[i] == '\r' || in[i] == '\n' || in[i] == '\t' {
			if whitespace {
				offset = i
			} else {
				whitespace = true
				out = append(out, in[offset:i]...)
				offset = i
			}
			continue
		}
		if whitespace {
			whitespace = false
			if len(out) > 0 {
				out = append(out, ' ')
			}
			offset = i
		}
		if in[i] == '?' {
			out = append(out, in[offset:i]...)
			offset = i + 1

			out = append(out, '$')
			out = append(out, strconv.Itoa(placeholders)...)
			placeholders++
		}
	}
	if !whitespace {
		out = append(out, in[offset:len(in)]...)
	}
	return string(out)
}

func (iter *iterator) NextScan(dst ...interface{}) error {
	if ok := iter.Next(); ok {
		return iter.Scan(dst...)
	}
	if err := iter.Err(); err != nil {
		return err
	}
	return db.ErrNoMoreRows
}

func (iter *iterator) ScanOne(dst ...interface{}) error {
	defer iter.Close()
	return iter.NextScan(dst...)
}

func (iter *iterator) Scan(dst ...interface{}) error {
	if err := iter.Err(); err != nil {
		return err
	}
	return iter.cursor.Scan(dst...)
}

func (iter *iterator) setErr(err error) error {
	iter.err = err
	return iter.err
}

func (iter *iterator) One(dst interface{}) error {
	if err := iter.Err(); err != nil {
		return err
	}
	defer iter.Close()
	return iter.setErr(iter.next(dst))
}

func (iter *iterator) All(dst interface{}) error {
	if err := iter.Err(); err != nil {
		return err
	}
	defer iter.Close()

	// Fetching all results within the cursor.
	if err := fetchRows(iter, dst); err != nil {
		return iter.setErr(err)
	}

	return nil
}

func (iter *iterator) Err() (err error) {
	return iter.err
}

func (iter *iterator) Next(dst ...interface{}) bool {
	if err := iter.Err(); err != nil {
		return false
	}

	if err := iter.next(dst...); err != nil {
		// ignore db.ErrNoMoreRows, just break.
		if !errors.Is(err, db.ErrNoMoreRows) {
			_ = iter.setErr(err)
		}
		return false
	}

	return true
}

func (iter *iterator) next(dst ...interface{}) error {
	if iter.cursor == nil {
		return iter.setErr(db.ErrNoMoreRows)
	}

	switch len(dst) {
	case 0:
		if ok := iter.cursor.Next(); !ok {
			defer iter.Close()
			err := iter.cursor.Err()
			if err == nil {
				err = db.ErrNoMoreRows
			}
			return err
		}
		return nil
	case 1:
		if err := fetchRow(iter, dst[0]); err != nil {
			defer iter.Close()
			return err
		}
		return nil
	}

	return errors.New("Next does not currently supports more than one parameters")
}

func (iter *iterator) Close() (err error) {
	if iter.cursor != nil {
		err = iter.cursor.Close()
		iter.cursor = nil
	}
	return err
}

func marshal(v interface{}) (interface{}, error) {
	if m, isMarshaler := v.(db.Marshaler); isMarshaler {
		var err error
		if v, err = m.MarshalDB(); err != nil {
			return nil, err
		}
	}
	return v, nil
}

func (fv *fieldValue) Len() int {
	return len(fv.fields)
}

func (fv *fieldValue) Swap(i, j int) {
	fv.fields[i], fv.fields[j] = fv.fields[j], fv.fields[i]
	fv.values[i], fv.values[j] = fv.values[j], fv.values[i]
}

func (fv *fieldValue) Less(i, j int) bool {
	return fv.fields[i] < fv.fields[j]
}

type exprProxy struct {
	db *sql.DB
	t  *exql.Template
}

func (p *exprProxy) Context() context.Context {
	return context.Background()
}

func (p *exprProxy) StatementExec(ctx context.Context, stmt *exql.Statement, args ...interface{}) (sql.Result, error) {
	s, err := stmt.Compile(p.t)
	if err != nil {
		return nil, err
	}
	return compat.ExecContext(p.db, ctx, s, args)
}

func (p *exprProxy) StatementPrepare(ctx context.Context, stmt *exql.Statement) (*sql.Stmt, error) {
	s, err := stmt.Compile(p.t)
	if err != nil {
		return nil, err
	}
	return compat.PrepareContext(p.db, ctx, s)
}

func (p *exprProxy) StatementQuery(ctx context.Context, stmt *exql.Statement, args ...interface{}) (*sql.Rows, error) {
	s, err := stmt.Compile(p.t)
	if err != nil {
		return nil, err
	}
	return compat.QueryContext(p.db, ctx, s, args)
}

func (p *exprProxy) StatementQueryRow(ctx context.Context, stmt *exql.Statement, args ...interface{}) (*sql.Row, error) {
	s, err := stmt.Compile(p.t)
	if err != nil {
		return nil, err
	}
	return compat.QueryRowContext(p.db, ctx, s, args), nil
}

var (
	_ = db.SQL(&sqlBuilder{})
	_ = exprDB(&exprProxy{})
)

func joinArguments(args ...[]interface{}) []interface{} {
	total := 0
	for i := range args {
		total += len(args[i])
	}
	if total == 0 {
		return nil
	}

	flatten := make([]interface{}, 0, total)
	for i := range args {
		flatten = append(flatten, args[i]...)
	}
	return flatten
}
