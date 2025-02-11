package sqlbuilder

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strings"

	db "github.com/upper/db/v4"
	"github.com/upper/db/v4/internal/immutable"
)

var (
	errMissingCursorColumn = errors.New("Missing cursor column")
)

type paginatorQuery struct {
	sel db.Selector

	cursorColumn       string
	cursorValue        interface{}
	cursorCond         db.Cond
	cursorReverseOrder bool

	pageSize   uint
	pageNumber uint
}

func newPaginator(sel db.Selector, pageSize uint) db.Paginator {
	pag := &paginator{}
	return pag.frame(func(pq *paginatorQuery) error {
		pq.pageSize = pageSize
		pq.sel = sel
		return nil
	}).Page(1)
}

func (pq *paginatorQuery) count() (uint64, error) {
	var count uint64

	row, err := pq.sel.(*selector).setColumns(db.Raw("count(1) AS _t")).
		Limit(0).
		Offset(0).
		OrderBy(nil).
		QueryRow()
	if err != nil {
		return 0, err
	}

	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

type paginator struct {
	fn   func(*paginatorQuery) error
	prev *paginator
}

var _ = immutable.Immutable(&paginator{})

func (pag *paginator) frame(fn func(*paginatorQuery) error) *paginator {
	return &paginator{prev: pag, fn: fn}
}

func (pag *paginator) Page(pageNumber uint) db.Paginator {
	return pag.frame(func(pq *paginatorQuery) error {
		if pageNumber < 1 {
			pageNumber = 1
		}
		pq.pageNumber = pageNumber
		return nil
	})
}

func (pag *paginator) Cursor(column string) db.Paginator {
	return pag.frame(func(pq *paginatorQuery) error {
		pq.cursorColumn = column
		pq.cursorValue = nil
		return nil
	})
}

func (pag *paginator) NextPage(cursorValue interface{}) db.Paginator {
	return pag.frame(func(pq *paginatorQuery) error {
		if pq.cursorValue != nil && pq.cursorColumn == "" {
			return errMissingCursorColumn
		}
		pq.cursorValue = cursorValue
		pq.cursorReverseOrder = false
		if strings.HasPrefix(pq.cursorColumn, "-") {
			pq.cursorCond = db.Cond{
				pq.cursorColumn[1:]: db.Lt(cursorValue),
			}
		} else {
			pq.cursorCond = db.Cond{
				pq.cursorColumn: db.Gt(cursorValue),
			}
		}
		return nil
	})
}

func (pag *paginator) PrevPage(cursorValue interface{}) db.Paginator {
	return pag.frame(func(pq *paginatorQuery) error {
		if pq.cursorValue != nil && pq.cursorColumn == "" {
			return errMissingCursorColumn
		}
		pq.cursorValue = cursorValue
		pq.cursorReverseOrder = true
		if strings.HasPrefix(pq.cursorColumn, "-") {
			pq.cursorCond = db.Cond{
				pq.cursorColumn[1:]: db.Gt(cursorValue),
			}
		} else {
			pq.cursorCond = db.Cond{
				pq.cursorColumn: db.Lt(cursorValue),
			}
		}
		return nil
	})
}

func (pag *paginator) TotalPages() (uint, error) {
	pq, err := pag.build()
	if err != nil {
		return 0, err
	}

	count, err := pq.count()
	if err != nil {
		return 0, err
	}
	if count < 1 {
		return 0, nil
	}

	if pq.pageSize < 1 {
		return 1, nil
	}

	pages := uint(math.Ceil(float64(count) / float64(pq.pageSize)))
	return pages, nil
}

func (pag *paginator) All(dest interface{}) error {
	pq, err := pag.buildWithCursor()
	if err != nil {
		return err
	}
	err = pq.sel.All(dest)
	if err != nil {
		return err
	}
	return nil
}

func (pag *paginator) One(dest interface{}) error {
	pq, err := pag.buildWithCursor()
	if err != nil {
		return err
	}
	return pq.sel.One(dest)
}

func (pag *paginator) Iterator() db.Iterator {
	pq, err := pag.buildWithCursor()
	if err != nil {
		sess := pq.sel.(*selector).SQL().sess
		return &iterator{sess, nil, err}
	}
	return pq.sel.Iterator()
}

func (pag *paginator) IteratorContext(ctx context.Context) db.Iterator {
	pq, err := pag.buildWithCursor()
	if err != nil {
		sess := pq.sel.(*selector).SQL().sess
		return &iterator{sess, nil, err}
	}
	return pq.sel.IteratorContext(ctx)
}

func (pag *paginator) String() string {
	pq, err := pag.buildWithCursor()
	if err != nil {
		panic(err.Error())
	}
	return pq.sel.String()
}

func (pag *paginator) Arguments() []interface{} {
	pq, err := pag.buildWithCursor()
	if err != nil {
		return nil
	}
	return pq.sel.Arguments()
}

func (pag *paginator) Compile() (string, error) {
	pq, err := pag.buildWithCursor()
	if err != nil {
		return "", err
	}
	return pq.sel.(*selector).Compile()
}

func (pag *paginator) Query() (*sql.Rows, error) {
	pq, err := pag.buildWithCursor()
	if err != nil {
		return nil, err
	}
	return pq.sel.Query()
}

func (pag *paginator) QueryContext(ctx context.Context) (*sql.Rows, error) {
	pq, err := pag.buildWithCursor()
	if err != nil {
		return nil, err
	}
	return pq.sel.QueryContext(ctx)
}

func (pag *paginator) QueryRow() (*sql.Row, error) {
	pq, err := pag.buildWithCursor()
	if err != nil {
		return nil, err
	}
	return pq.sel.QueryRow()
}

func (pag *paginator) QueryRowContext(ctx context.Context) (*sql.Row, error) {
	pq, err := pag.buildWithCursor()
	if err != nil {
		return nil, err
	}
	return pq.sel.QueryRowContext(ctx)
}

func (pag *paginator) Prepare() (*sql.Stmt, error) {
	pq, err := pag.buildWithCursor()
	if err != nil {
		return nil, err
	}
	return pq.sel.Prepare()
}

func (pag *paginator) PrepareContext(ctx context.Context) (*sql.Stmt, error) {
	pq, err := pag.buildWithCursor()
	if err != nil {
		return nil, err
	}
	return pq.sel.PrepareContext(ctx)
}

func (pag *paginator) TotalEntries() (uint64, error) {
	pq, err := pag.build()
	if err != nil {
		return 0, err
	}
	return pq.count()
}

func (pag *paginator) build() (*paginatorQuery, error) {
	pq, err := immutable.FastForward(pag)
	if err != nil {
		return nil, err
	}
	return pq.(*paginatorQuery), nil
}

func (pag *paginator) buildWithCursor() (*paginatorQuery, error) {
	pq, err := immutable.FastForward(pag)
	if err != nil {
		return nil, err
	}

	pqq := pq.(*paginatorQuery)

	if pqq.cursorReverseOrder {
		orderBy := pqq.cursorColumn

		if orderBy == "" {
			return nil, errMissingCursorColumn
		}

		if strings.HasPrefix(orderBy, "-") {
			orderBy = orderBy[1:]
		} else {
			orderBy = "-" + orderBy
		}

		pqq.sel = pqq.sel.OrderBy(orderBy)
	}

	if pqq.pageSize > 0 {
		pqq.sel = pqq.sel.Limit(int(pqq.pageSize))
		if pqq.pageNumber > 1 {
			pqq.sel = pqq.sel.Offset(int(pqq.pageSize * (pqq.pageNumber - 1)))
		}
	}

	if pqq.cursorCond != nil {
		pqq.sel = pqq.sel.Where(pqq.cursorCond).Offset(0)
	}

	if pqq.cursorColumn != "" {
		if pqq.cursorReverseOrder {
			pqq.sel = pqq.sel.(*selector).SQL().
				SelectFrom(db.Raw("? AS p0", pqq.sel)).
				OrderBy(pqq.cursorColumn)
		} else {
			pqq.sel = pqq.sel.OrderBy(pqq.cursorColumn)
		}
	}

	return pqq, nil
}

func (pag *paginator) Prev() immutable.Immutable {
	if pag == nil {
		return nil
	}
	return pag.prev
}

func (pag *paginator) Fn(in interface{}) error {
	if pag.fn == nil {
		return nil
	}
	return pag.fn(in.(*paginatorQuery))
}

func (pag *paginator) Base() interface{} {
	return &paginatorQuery{}
}
