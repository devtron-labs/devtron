// Copyright 2016 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

import (
	sql2 "database/sql"
	"fmt"
)

type optype byte

const (
	condType   optype = iota // only conditions
	selectType               // select
	insertType               // insert
	updateType               // update
	deleteType               // delete
	unionType                // union
)

// all databasees
const (
	POSTGRES = "postgres"
	SQLITE   = "sqlite3"
	MYSQL    = "mysql"
	MSSQL    = "mssql"
	ORACLE   = "oracle"
)

type join struct {
	joinType  string
	joinTable interface{}
	joinCond  Cond
}

type union struct {
	unionType string
	builder   *Builder
}

type limit struct {
	limitN int
	offset int
}

// Builder describes a SQL statement
type Builder struct {
	optype
	dialect    string
	isNested   bool
	into       string
	from       string
	subQuery   *Builder
	cond       Cond
	selects    []string
	joins      []join
	unions     []union
	limitation *limit
	insertCols []string
	insertVals []interface{}
	updates    []UpdateCond
	orderBy    string
	groupBy    string
	having     string
}

// Dialect sets the db dialect of Builder.
func Dialect(dialect string) *Builder {
	builder := &Builder{cond: NewCond(), dialect: dialect}
	return builder
}

// MySQL is shortcut of Dialect(MySQL)
func MySQL() *Builder {
	return Dialect(MYSQL)
}

// MsSQL is shortcut of Dialect(MsSQL)
func MsSQL() *Builder {
	return Dialect(MSSQL)
}

// Oracle is shortcut of Dialect(Oracle)
func Oracle() *Builder {
	return Dialect(ORACLE)
}

// Postgres is shortcut of Dialect(Postgres)
func Postgres() *Builder {
	return Dialect(POSTGRES)
}

// SQLite is shortcut of Dialect(SQLITE)
func SQLite() *Builder {
	return Dialect(SQLITE)
}

// Where sets where SQL
func (b *Builder) Where(cond Cond) *Builder {
	if b.cond.IsValid() {
		b.cond = b.cond.And(cond)
	} else {
		b.cond = cond
	}
	return b
}

// From sets from subject(can be a table name in string or a builder pointer) and its alias
func (b *Builder) From(subject interface{}, alias ...string) *Builder {
	switch subject.(type) {
	case *Builder:
		b.subQuery = subject.(*Builder)

		if len(alias) > 0 {
			b.from = alias[0]
		} else {
			b.isNested = true
		}
	case string:
		b.from = subject.(string)

		if len(alias) > 0 {
			b.from = b.from + " " + alias[0]
		}
	}

	return b
}

// TableName returns the table name
func (b *Builder) TableName() string {
	if b.optype == insertType {
		return b.into
	}
	return b.from
}

// Into sets insert table name
func (b *Builder) Into(tableName string) *Builder {
	b.into = tableName
	return b
}

// Union sets union conditions
func (b *Builder) Union(unionTp string, unionCond *Builder) *Builder {
	var builder *Builder
	if b.optype != unionType {
		builder = &Builder{cond: NewCond()}
		builder.optype = unionType
		builder.dialect = b.dialect
		builder.selects = b.selects

		currentUnions := b.unions
		// erase sub unions (actually append to new Builder.unions)
		b.unions = nil

		for e := range currentUnions {
			currentUnions[e].builder.dialect = b.dialect
		}

		builder.unions = append(append(builder.unions, union{"", b}), currentUnions...)
	} else {
		builder = b
	}

	if unionCond != nil {
		if unionCond.dialect == "" && builder.dialect != "" {
			unionCond.dialect = builder.dialect
		}

		builder.unions = append(builder.unions, union{unionTp, unionCond})
	}

	return builder
}

// Limit sets limitN condition
func (b *Builder) Limit(limitN int, offset ...int) *Builder {
	b.limitation = &limit{limitN: limitN}

	if len(offset) > 0 {
		b.limitation.offset = offset[0]
	}

	return b
}

// Select sets select SQL
func (b *Builder) Select(cols ...string) *Builder {
	b.selects = cols
	if b.optype == condType {
		b.optype = selectType
	}
	return b
}

// And sets AND condition
func (b *Builder) And(cond Cond) *Builder {
	b.cond = And(b.cond, cond)
	return b
}

// Or sets OR condition
func (b *Builder) Or(cond Cond) *Builder {
	b.cond = Or(b.cond, cond)
	return b
}

// Update sets update SQL
func (b *Builder) Update(updates ...Cond) *Builder {
	b.updates = make([]UpdateCond, 0, len(updates))
	for _, update := range updates {
		if u, ok := update.(UpdateCond); ok && u.IsValid() {
			b.updates = append(b.updates, u)
		}
	}
	b.optype = updateType
	return b
}

// Delete sets delete SQL
func (b *Builder) Delete(conds ...Cond) *Builder {
	b.cond = b.cond.And(conds...)
	b.optype = deleteType
	return b
}

// WriteTo implements Writer interface
func (b *Builder) WriteTo(w Writer) error {
	switch b.optype {
	/*case condType:
	return b.cond.WriteTo(w)*/
	case selectType:
		return b.selectWriteTo(w)
	case insertType:
		return b.insertWriteTo(w)
	case updateType:
		return b.updateWriteTo(w)
	case deleteType:
		return b.deleteWriteTo(w)
	case unionType:
		return b.unionWriteTo(w)
	}

	return ErrNotSupportType
}

// ToSQL convert a builder to SQL and args
func (b *Builder) ToSQL() (string, []interface{}, error) {
	w := NewWriter()
	if err := b.WriteTo(w); err != nil {
		return "", nil, err
	}

	// in case of sql.NamedArg in args
	for e := range w.args {
		if namedArg, ok := w.args[e].(sql2.NamedArg); ok {
			w.args[e] = namedArg.Value
		}
	}

	var sql = w.String()
	var err error

	switch b.dialect {
	case ORACLE, MSSQL:
		// This is for compatibility with different sql drivers
		for e := range w.args {
			w.args[e] = sql2.Named(fmt.Sprintf("p%d", e+1), w.args[e])
		}

		var prefix string
		if b.dialect == ORACLE {
			prefix = ":p"
		} else {
			prefix = "@p"
		}

		if sql, err = ConvertPlaceholder(sql, prefix); err != nil {
			return "", nil, err
		}
	case POSTGRES:
		if sql, err = ConvertPlaceholder(sql, "$"); err != nil {
			return "", nil, err
		}
	}

	return sql, w.args, nil
}

// ToBoundSQL generated a bound SQL string
func (b *Builder) ToBoundSQL() (string, error) {
	w := NewWriter()
	if err := b.WriteTo(w); err != nil {
		return "", err
	}

	return ConvertToBoundSQL(w.String(), w.args)
}
