package sqlbuilder

import (
	"context"
	"database/sql"
	"errors"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/internal/immutable"
	"github.com/upper/db/v4/internal/sqladapter/exql"
)

type inserterQuery struct {
	table          string
	enqueuedValues [][]interface{}
	returning      []exql.Fragment
	columns        []exql.Fragment
	values         []*exql.Values
	arguments      []interface{}
	amendFn        func(string) string
}

func (iq *inserterQuery) processValues() ([]*exql.Values, []interface{}, error) {
	var values []*exql.Values
	var arguments []interface{}

	var mapOptions *MapOptions
	if len(iq.enqueuedValues) > 1 {
		mapOptions = &MapOptions{IncludeZeroed: true, IncludeNil: true}
	}

	for _, enqueuedValue := range iq.enqueuedValues {
		if len(enqueuedValue) == 1 {
			// If and only if we passed one argument to Values.
			ff, vv, err := Map(enqueuedValue[0], mapOptions)

			if err == nil {
				// If we didn't have any problem with mapping we can convert it into
				// columns and values.
				columns, vals, args, _ := toColumnsValuesAndArguments(ff, vv)

				values, arguments = append(values, vals), append(arguments, args...)

				if len(iq.columns) == 0 {
					iq.columns = append(iq.columns, columns.Columns...)
				}
				continue
			}

			// The only error we can expect without exiting is this argument not
			// being a map or struct, in which case we can continue.
			if !errors.Is(err, ErrExpectingPointerToEitherMapOrStruct) {
				return nil, nil, err
			}
		}

		if len(iq.columns) == 0 || len(enqueuedValue) == len(iq.columns) {
			arguments = append(arguments, enqueuedValue...)

			l := len(enqueuedValue)
			placeholders := make([]exql.Fragment, l)
			for i := 0; i < l; i++ {
				placeholders[i] = sqlPlaceholder
			}
			values = append(values, exql.NewValueGroup(placeholders...))
		}
	}

	return values, arguments, nil
}

func (iq *inserterQuery) statement() *exql.Statement {
	stmt := &exql.Statement{
		Type:  exql.Insert,
		Table: exql.TableWithName(iq.table),
	}

	if len(iq.values) > 0 {
		stmt.Values = exql.JoinValueGroups(iq.values...)
	}

	if len(iq.columns) > 0 {
		stmt.Columns = exql.JoinColumns(iq.columns...)
	}

	if len(iq.returning) > 0 {
		stmt.Returning = exql.ReturningColumns(iq.returning...)
	}

	stmt.SetAmendment(iq.amendFn)

	return stmt
}

type inserter struct {
	builder *sqlBuilder

	fn   func(*inserterQuery) error
	prev *inserter
}

var _ = immutable.Immutable(&inserter{})

func (ins *inserter) SQL() *sqlBuilder {
	if ins.prev == nil {
		return ins.builder
	}
	return ins.prev.SQL()
}

func (ins *inserter) template() *exql.Template {
	return ins.SQL().t.Template
}

func (ins *inserter) String() string {
	s, err := ins.Compile()
	if err != nil {
		panic(err.Error())
	}
	return prepareQueryForDisplay(s)
}

func (ins *inserter) frame(fn func(*inserterQuery) error) *inserter {
	return &inserter{prev: ins, fn: fn}
}

func (ins *inserter) Batch(n int) db.BatchInserter {
	return newBatchInserter(ins, n)
}

func (ins *inserter) Amend(fn func(string) string) db.Inserter {
	return ins.frame(func(iq *inserterQuery) error {
		iq.amendFn = fn
		return nil
	})
}

func (ins *inserter) Arguments() []interface{} {
	iq, err := ins.build()
	if err != nil {
		return nil
	}
	return iq.arguments
}

func (ins *inserter) Returning(columns ...string) db.Inserter {
	return ins.frame(func(iq *inserterQuery) error {
		columnsToFragments(&iq.returning, columns)
		return nil
	})
}

func (ins *inserter) Exec() (sql.Result, error) {
	return ins.ExecContext(ins.SQL().sess.Context())
}

func (ins *inserter) ExecContext(ctx context.Context) (sql.Result, error) {
	iq, err := ins.build()
	if err != nil {
		return nil, err
	}
	return ins.SQL().sess.StatementExec(ctx, iq.statement(), iq.arguments...)
}

func (ins *inserter) Prepare() (*sql.Stmt, error) {
	return ins.PrepareContext(ins.SQL().sess.Context())
}

func (ins *inserter) PrepareContext(ctx context.Context) (*sql.Stmt, error) {
	iq, err := ins.build()
	if err != nil {
		return nil, err
	}
	return ins.SQL().sess.StatementPrepare(ctx, iq.statement())
}

func (ins *inserter) Query() (*sql.Rows, error) {
	return ins.QueryContext(ins.SQL().sess.Context())
}

func (ins *inserter) QueryContext(ctx context.Context) (*sql.Rows, error) {
	iq, err := ins.build()
	if err != nil {
		return nil, err
	}
	return ins.SQL().sess.StatementQuery(ctx, iq.statement(), iq.arguments...)
}

func (ins *inserter) QueryRow() (*sql.Row, error) {
	return ins.QueryRowContext(ins.SQL().sess.Context())
}

func (ins *inserter) QueryRowContext(ctx context.Context) (*sql.Row, error) {
	iq, err := ins.build()
	if err != nil {
		return nil, err
	}
	return ins.SQL().sess.StatementQueryRow(ctx, iq.statement(), iq.arguments...)
}

func (ins *inserter) Iterator() db.Iterator {
	return ins.IteratorContext(ins.SQL().sess.Context())
}

func (ins *inserter) IteratorContext(ctx context.Context) db.Iterator {
	rows, err := ins.QueryContext(ctx)
	return &iterator{ins.SQL().sess, rows, err}
}

func (ins *inserter) Into(table string) db.Inserter {
	return ins.frame(func(iq *inserterQuery) error {
		iq.table = table
		return nil
	})
}

func (ins *inserter) Columns(columns ...string) db.Inserter {
	return ins.frame(func(iq *inserterQuery) error {
		columnsToFragments(&iq.columns, columns)
		return nil
	})
}

func (ins *inserter) Values(values ...interface{}) db.Inserter {
	return ins.frame(func(iq *inserterQuery) error {
		iq.enqueuedValues = append(iq.enqueuedValues, values)
		return nil
	})
}

func (ins *inserter) statement() (*exql.Statement, error) {
	iq, err := ins.build()
	if err != nil {
		return nil, err
	}
	return iq.statement(), nil
}

func (ins *inserter) build() (*inserterQuery, error) {
	iq, err := immutable.FastForward(ins)
	if err != nil {
		return nil, err
	}
	ret := iq.(*inserterQuery)
	ret.values, ret.arguments, err = ret.processValues()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (ins *inserter) Compile() (string, error) {
	s, err := ins.statement()
	if err != nil {
		return "", err
	}
	return s.Compile(ins.template())
}

func (ins *inserter) Prev() immutable.Immutable {
	if ins == nil {
		return nil
	}
	return ins.prev
}

func (ins *inserter) Fn(in interface{}) error {
	if ins.fn == nil {
		return nil
	}
	return ins.fn(in.(*inserterQuery))
}

func (ins *inserter) Base() interface{} {
	return &inserterQuery{}
}

func columnsToFragments(dst *[]exql.Fragment, columns []string) {
	l := len(columns)
	f := make([]exql.Fragment, l)
	for i := 0; i < l; i++ {
		f[i] = exql.ColumnWithName(columns[i])
	}
	*dst = append(*dst, f...)
}
