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

package sqladapter

import (
	"errors"
	"sync"
	"sync/atomic"

	db "github.com/upper/db/v4"
	"github.com/upper/db/v4/internal/immutable"
)

type Result struct {
	builder db.SQL

	err atomic.Value

	iter   db.Iterator
	iterMu sync.Mutex

	prev *Result
	fn   func(*result) error
}

// result represents a delimited set of items bound by a condition.
type result struct {
	table  string
	limit  int
	offset int

	pageSize   uint
	pageNumber uint

	cursorColumn        string
	nextPageCursorValue interface{}
	prevPageCursorValue interface{}

	fields  []interface{}
	orderBy []interface{}
	groupBy []interface{}
	conds   [][]interface{}
}

func filter(conds []interface{}) []interface{} {
	return conds
}

// NewResult creates and Results a new Result set on the given table, this set
// is limited by the given exql.Where conditions.
func NewResult(builder db.SQL, table string, conds []interface{}) *Result {
	r := &Result{
		builder: builder,
	}
	return r.from(table).where(conds)
}

func (r *Result) frame(fn func(*result) error) *Result {
	return &Result{err: r.err, prev: r, fn: fn}
}

func (r *Result) SQL() db.SQL {
	if r.prev == nil {
		return r.builder
	}
	return r.prev.SQL()
}

func (r *Result) from(table string) *Result {
	return r.frame(func(res *result) error {
		res.table = table
		return nil
	})
}

func (r *Result) where(conds []interface{}) *Result {
	return r.frame(func(res *result) error {
		res.conds = [][]interface{}{conds}
		return nil
	})
}

func (r *Result) setErr(err error) {
	if err == nil {
		return
	}
	r.err.Store(err)
}

// Err returns the last error that has happened with the result set,
// nil otherwise
func (r *Result) Err() error {
	if errV := r.err.Load(); errV != nil {
		return errV.(error)
	}
	return nil
}

// Where sets conditions for the result set.
func (r *Result) Where(conds ...interface{}) db.Result {
	return r.where(conds)
}

// And adds more conditions on top of the existing ones.
func (r *Result) And(conds ...interface{}) db.Result {
	return r.frame(func(res *result) error {
		res.conds = append(res.conds, conds)
		return nil
	})
}

// Limit determines the maximum limit of Results to be returned.
func (r *Result) Limit(n int) db.Result {
	return r.frame(func(res *result) error {
		res.limit = n
		return nil
	})
}

func (r *Result) Paginate(pageSize uint) db.Result {
	return r.frame(func(res *result) error {
		res.pageSize = pageSize
		return nil
	})
}

func (r *Result) Page(pageNumber uint) db.Result {
	return r.frame(func(res *result) error {
		res.pageNumber = pageNumber
		res.nextPageCursorValue = nil
		res.prevPageCursorValue = nil
		return nil
	})
}

func (r *Result) Cursor(cursorColumn string) db.Result {
	return r.frame(func(res *result) error {
		res.cursorColumn = cursorColumn
		return nil
	})
}

func (r *Result) NextPage(cursorValue interface{}) db.Result {
	return r.frame(func(res *result) error {
		res.nextPageCursorValue = cursorValue
		res.prevPageCursorValue = nil
		return nil
	})
}

func (r *Result) PrevPage(cursorValue interface{}) db.Result {
	return r.frame(func(res *result) error {
		res.nextPageCursorValue = nil
		res.prevPageCursorValue = cursorValue
		return nil
	})
}

// Offset determines how many documents will be skipped before starting to grab
// Results.
func (r *Result) Offset(n int) db.Result {
	return r.frame(func(res *result) error {
		res.offset = n
		return nil
	})
}

// GroupBy is used to group Results that have the same value in the same column
// or columns.
func (r *Result) GroupBy(fields ...interface{}) db.Result {
	return r.frame(func(res *result) error {
		res.groupBy = fields
		return nil
	})
}

// OrderBy determines sorting of Results according to the provided names. Fields
// may be prefixed by - (minus) which means descending order, ascending order
// would be used otherwise.
func (r *Result) OrderBy(fields ...interface{}) db.Result {
	return r.frame(func(res *result) error {
		res.orderBy = fields
		return nil
	})
}

// Select determines which fields to return.
func (r *Result) Select(fields ...interface{}) db.Result {
	return r.frame(func(res *result) error {
		res.fields = fields
		return nil
	})
}

// String satisfies fmt.Stringer
func (r *Result) String() string {
	query, err := r.Paginator()
	if err != nil {
		panic(err.Error())
	}
	return query.String()
}

// All dumps all Results into a pointer to an slice of structs or maps.
func (r *Result) All(dst interface{}) error {
	query, err := r.Paginator()
	if err != nil {
		r.setErr(err)
		return err
	}
	err = query.Iterator().All(dst)
	r.setErr(err)
	return err
}

// One fetches only one Result from the set.
func (r *Result) One(dst interface{}) error {
	one := r.Limit(1).(*Result)
	query, err := one.Paginator()
	if err != nil {
		r.setErr(err)
		return err
	}

	err = query.Iterator().One(dst)
	r.setErr(err)
	return err
}

// Next fetches the next Result from the set.
func (r *Result) Next(dst interface{}) bool {
	r.iterMu.Lock()
	defer r.iterMu.Unlock()

	if r.iter == nil {
		query, err := r.Paginator()
		if err != nil {
			r.setErr(err)
			return false
		}
		r.iter = query.Iterator()
	}

	if r.iter.Next(dst) {
		return true
	}

	if err := r.iter.Err(); !errors.Is(err, db.ErrNoMoreRows) {
		r.setErr(err)
		return false
	}

	return false
}

// Delete deletes all matching items from the collection.
func (r *Result) Delete() error {
	query, err := r.buildDelete()
	if err != nil {
		r.setErr(err)
		return err
	}

	_, err = query.Exec()
	r.setErr(err)
	return err
}

// Close closes the Result set.
func (r *Result) Close() error {
	if r.iter != nil {
		err := r.iter.Close()
		r.setErr(err)
		return err
	}
	return nil
}

// Update updates matching items from the collection with values of the given
// map or struct.
func (r *Result) Update(values interface{}) error {
	query, err := r.buildUpdate(values)
	if err != nil {
		r.setErr(err)
		return err
	}

	_, err = query.Exec()
	r.setErr(err)
	return err
}

func (r *Result) TotalPages() (uint, error) {
	query, err := r.Paginator()
	if err != nil {
		r.setErr(err)
		return 0, err
	}

	total, err := query.TotalPages()
	if err != nil {
		r.setErr(err)
		return 0, err
	}

	return total, nil
}

func (r *Result) TotalEntries() (uint64, error) {
	query, err := r.Paginator()
	if err != nil {
		r.setErr(err)
		return 0, err
	}

	total, err := query.TotalEntries()
	if err != nil {
		r.setErr(err)
		return 0, err
	}

	return total, nil
}

// Exists returns true if at least one item on the collection exists.
func (r *Result) Exists() (bool, error) {
	query, err := r.buildCount()
	if err != nil {
		r.setErr(err)
		return false, err
	}

	query = query.Limit(1)

	value := struct {
		Exists uint64 `db:"_t"`
	}{}

	if err := query.One(&value); err != nil {
		if errors.Is(err, db.ErrNoMoreRows) {
			return false, nil
		}
		r.setErr(err)
		return false, err
	}

	if value.Exists > 0 {
		return true, nil
	}

	return false, nil
}

// Count counts the elements on the set.
func (r *Result) Count() (uint64, error) {
	query, err := r.buildCount()
	if err != nil {
		r.setErr(err)
		return 0, err
	}

	counter := struct {
		Count uint64 `db:"_t"`
	}{}
	if err := query.One(&counter); err != nil {
		if errors.Is(err, db.ErrNoMoreRows) {
			return 0, nil
		}
		r.setErr(err)
		return 0, err
	}

	return counter.Count, nil
}

func (r *Result) Paginator() (db.Paginator, error) {
	if err := r.Err(); err != nil {
		return nil, err
	}

	res, err := r.fastForward()
	if err != nil {
		return nil, err
	}

	sel := r.SQL().Select(res.fields...).
		From(res.table).
		Limit(res.limit).
		Offset(res.offset).
		GroupBy(res.groupBy...).
		OrderBy(res.orderBy...)

	for i := range res.conds {
		sel = sel.And(filter(res.conds[i])...)
	}

	pag := sel.Paginate(res.pageSize).
		Page(res.pageNumber).
		Cursor(res.cursorColumn)

	if res.nextPageCursorValue != nil {
		pag = pag.NextPage(res.nextPageCursorValue)
	}

	if res.prevPageCursorValue != nil {
		pag = pag.PrevPage(res.prevPageCursorValue)
	}

	return pag, nil
}

func (r *Result) buildDelete() (db.Deleter, error) {
	if err := r.Err(); err != nil {
		return nil, err
	}

	res, err := r.fastForward()
	if err != nil {
		return nil, err
	}

	del := r.SQL().DeleteFrom(res.table).
		Limit(res.limit)

	for i := range res.conds {
		del = del.And(filter(res.conds[i])...)
	}

	return del, nil
}

func (r *Result) buildUpdate(values interface{}) (db.Updater, error) {
	if err := r.Err(); err != nil {
		return nil, err
	}

	res, err := r.fastForward()
	if err != nil {
		return nil, err
	}

	upd := r.SQL().Update(res.table).
		Set(values).
		Limit(res.limit)

	for i := range res.conds {
		upd = upd.And(filter(res.conds[i])...)
	}

	return upd, nil
}

func (r *Result) buildCount() (db.Selector, error) {
	if err := r.Err(); err != nil {
		return nil, err
	}

	res, err := r.fastForward()
	if err != nil {
		return nil, err
	}

	sel := r.SQL().Select(db.Raw("count(1) AS _t")).
		From(res.table).
		GroupBy(res.groupBy...)

	for i := range res.conds {
		sel = sel.And(filter(res.conds[i])...)
	}

	return sel, nil
}

func (r *Result) Prev() immutable.Immutable {
	if r == nil {
		return nil
	}
	return r.prev
}

func (r *Result) Fn(in interface{}) error {
	if r.fn == nil {
		return nil
	}
	return r.fn(in.(*result))
}

func (r *Result) Base() interface{} {
	return &result{}
}

func (r *Result) fastForward() (*result, error) {
	ff, err := immutable.FastForward(r)
	if err != nil {
		return nil, err
	}
	return ff.(*result), nil
}

var _ = immutable.Immutable(&Result{})
