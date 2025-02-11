package sqlbuilder

import (
	"context"
	"database/sql"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/internal/immutable"
	"github.com/upper/db/v4/internal/sqladapter/exql"
)

type updaterQuery struct {
	table string

	columnValues     *exql.ColumnValues
	columnValuesArgs []interface{}

	limit int

	where     *exql.Where
	whereArgs []interface{}

	amendFn func(string) string
}

func (uq *updaterQuery) and(b *sqlBuilder, terms ...interface{}) error {
	where, whereArgs := b.t.toWhereWithArguments(terms)

	if uq.where == nil {
		uq.where, uq.whereArgs = &exql.Where{}, []interface{}{}
	}
	uq.where.Append(&where)
	uq.whereArgs = append(uq.whereArgs, whereArgs...)

	return nil
}

func (uq *updaterQuery) statement() *exql.Statement {
	stmt := &exql.Statement{
		Type:         exql.Update,
		Table:        exql.TableWithName(uq.table),
		ColumnValues: uq.columnValues,
	}

	if uq.where != nil {
		stmt.Where = uq.where
	}

	if uq.limit != 0 {
		stmt.Limit = exql.Limit(uq.limit)
	}

	stmt.SetAmendment(uq.amendFn)

	return stmt
}

func (uq *updaterQuery) arguments() []interface{} {
	return joinArguments(
		uq.columnValuesArgs,
		uq.whereArgs,
	)
}

type updater struct {
	builder *sqlBuilder

	fn   func(*updaterQuery) error
	prev *updater
}

var _ = immutable.Immutable(&updater{})

func (upd *updater) SQL() *sqlBuilder {
	if upd.prev == nil {
		return upd.builder
	}
	return upd.prev.SQL()
}

func (upd *updater) template() *exql.Template {
	return upd.SQL().t.Template
}

func (upd *updater) String() string {
	s, err := upd.Compile()
	if err != nil {
		panic(err.Error())
	}
	return prepareQueryForDisplay(s)
}

func (upd *updater) setTable(table string) *updater {
	return upd.frame(func(uq *updaterQuery) error {
		uq.table = table
		return nil
	})
}

func (upd *updater) frame(fn func(*updaterQuery) error) *updater {
	return &updater{prev: upd, fn: fn}
}

func (upd *updater) Set(terms ...interface{}) db.Updater {
	return upd.frame(func(uq *updaterQuery) error {
		if uq.columnValues == nil {
			uq.columnValues = &exql.ColumnValues{}
		}

		if len(terms) == 1 {
			ff, vv, err := Map(terms[0], nil)
			if err == nil && len(ff) > 0 {
				cvs := make([]exql.Fragment, 0, len(ff))
				args := make([]interface{}, 0, len(vv))

				for i := range ff {
					cv := &exql.ColumnValue{
						Column:   exql.ColumnWithName(ff[i]),
						Operator: upd.SQL().t.AssignmentOperator,
					}

					var localArgs []interface{}
					cv.Value, localArgs = upd.SQL().t.PlaceholderValue(vv[i])

					args = append(args, localArgs...)
					cvs = append(cvs, cv)
				}

				uq.columnValues.Insert(cvs...)
				uq.columnValuesArgs = append(uq.columnValuesArgs, args...)

				return nil
			}
		}

		cv, arguments := upd.SQL().t.setColumnValues(terms)
		uq.columnValues.Insert(cv.ColumnValues...)
		uq.columnValuesArgs = append(uq.columnValuesArgs, arguments...)
		return nil
	})
}

func (upd *updater) Amend(fn func(string) string) db.Updater {
	return upd.frame(func(uq *updaterQuery) error {
		uq.amendFn = fn
		return nil
	})
}

func (upd *updater) Arguments() []interface{} {
	uq, err := upd.build()
	if err != nil {
		return nil
	}
	return uq.arguments()
}

func (upd *updater) Where(terms ...interface{}) db.Updater {
	return upd.frame(func(uq *updaterQuery) error {
		uq.where, uq.whereArgs = &exql.Where{}, []interface{}{}
		return uq.and(upd.SQL(), terms...)
	})
}

func (upd *updater) And(terms ...interface{}) db.Updater {
	return upd.frame(func(uq *updaterQuery) error {
		return uq.and(upd.SQL(), terms...)
	})
}

func (upd *updater) Prepare() (*sql.Stmt, error) {
	return upd.PrepareContext(upd.SQL().sess.Context())
}

func (upd *updater) PrepareContext(ctx context.Context) (*sql.Stmt, error) {
	uq, err := upd.build()
	if err != nil {
		return nil, err
	}
	return upd.SQL().sess.StatementPrepare(ctx, uq.statement())
}

func (upd *updater) Exec() (sql.Result, error) {
	return upd.ExecContext(upd.SQL().sess.Context())
}

func (upd *updater) ExecContext(ctx context.Context) (sql.Result, error) {
	uq, err := upd.build()
	if err != nil {
		return nil, err
	}
	return upd.SQL().sess.StatementExec(ctx, uq.statement(), uq.arguments()...)
}

func (upd *updater) Limit(limit int) db.Updater {
	return upd.frame(func(uq *updaterQuery) error {
		uq.limit = limit
		return nil
	})
}

func (upd *updater) statement() (*exql.Statement, error) {
	iq, err := upd.build()
	if err != nil {
		return nil, err
	}
	return iq.statement(), nil
}

func (upd *updater) build() (*updaterQuery, error) {
	uq, err := immutable.FastForward(upd)
	if err != nil {
		return nil, err
	}
	return uq.(*updaterQuery), nil
}

func (upd *updater) Compile() (string, error) {
	s, err := upd.statement()
	if err != nil {
		return "", err
	}
	return s.Compile(upd.template())
}

func (upd *updater) Prev() immutable.Immutable {
	if upd == nil {
		return nil
	}
	return upd.prev
}

func (upd *updater) Fn(in interface{}) error {
	if upd.fn == nil {
		return nil
	}
	return upd.fn(in.(*updaterQuery))
}

func (upd *updater) Base() interface{} {
	return &updaterQuery{}
}
