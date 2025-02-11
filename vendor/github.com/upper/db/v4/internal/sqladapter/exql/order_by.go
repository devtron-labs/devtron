package exql

import (
	"strings"

	"github.com/upper/db/v4/internal/cache"
)

// Order represents the order in which SQL results are sorted.
type Order uint8

// Possible values for Order
const (
	Order_Default Order = iota

	Order_Ascendent
	Order_Descendent
)

func (o Order) Hash() uint64 {
	return cache.NewHash(FragmentType_Order, uint8(o))
}

// SortColumn represents the column-order relation in an ORDER BY clause.
type SortColumn struct {
	Column Fragment
	Order
}

var _ = Fragment(&SortColumn{})

type sortColumnT struct {
	Column string
	Order  string
}

var _ = Fragment(&SortColumn{})

// SortColumns represents the columns in an ORDER BY clause.
type SortColumns struct {
	Columns []Fragment
}

var _ = Fragment(&SortColumns{})

// OrderBy represents an ORDER BY clause.
type OrderBy struct {
	SortColumns Fragment
}

var _ = Fragment(&OrderBy{})

type orderByT struct {
	SortColumns string
}

// JoinSortColumns creates and returns an array of column-order relations.
func JoinSortColumns(values ...Fragment) *SortColumns {
	return &SortColumns{Columns: values}
}

// JoinWithOrderBy creates an returns an OrderBy using the given SortColumns.
func JoinWithOrderBy(sc *SortColumns) *OrderBy {
	return &OrderBy{SortColumns: sc}
}

// Hash returns a unique identifier for the struct.
func (s *SortColumn) Hash() uint64 {
	if s == nil {
		return cache.NewHash(FragmentType_SortColumn, nil)
	}
	return cache.NewHash(FragmentType_SortColumn, s.Column, s.Order)
}

// Compile transforms the SortColumn into an equivalent SQL representation.
func (s *SortColumn) Compile(layout *Template) (compiled string, err error) {

	if c, ok := layout.Read(s); ok {
		return c, nil
	}

	column, err := s.Column.Compile(layout)
	if err != nil {
		return "", err
	}

	orderBy, err := s.Order.Compile(layout)
	if err != nil {
		return "", err
	}

	data := sortColumnT{Column: column, Order: orderBy}

	compiled = layout.MustCompile(layout.SortByColumnLayout, data)

	layout.Write(s, compiled)

	return
}

// Hash returns a unique identifier for the struct.
func (s *SortColumns) Hash() uint64 {
	if s == nil {
		return cache.NewHash(FragmentType_SortColumns, nil)
	}
	h := cache.InitHash(FragmentType_SortColumns)
	for i := range s.Columns {
		h = cache.AddToHash(h, s.Columns[i])
	}
	return h
}

// Compile transforms the SortColumns into an equivalent SQL representation.
func (s *SortColumns) Compile(layout *Template) (compiled string, err error) {
	if z, ok := layout.Read(s); ok {
		return z, nil
	}

	z := make([]string, len(s.Columns))

	for i := range s.Columns {
		z[i], err = s.Columns[i].Compile(layout)
		if err != nil {
			return "", err
		}
	}

	compiled = strings.Join(z, layout.IdentifierSeparator)

	layout.Write(s, compiled)

	return
}

// Hash returns a unique identifier for the struct.
func (s *OrderBy) Hash() uint64 {
	if s == nil {
		return cache.NewHash(FragmentType_OrderBy, nil)
	}
	return cache.NewHash(FragmentType_OrderBy, s.SortColumns)
}

// Compile transforms the SortColumn into an equivalent SQL representation.
func (s *OrderBy) Compile(layout *Template) (compiled string, err error) {
	if z, ok := layout.Read(s); ok {
		return z, nil
	}

	if s.SortColumns != nil {
		sortColumns, err := s.SortColumns.Compile(layout)
		if err != nil {
			return "", err
		}

		data := orderByT{
			SortColumns: sortColumns,
		}
		compiled = layout.MustCompile(layout.OrderByLayout, data)
	}

	layout.Write(s, compiled)

	return
}

// Compile transforms the SortColumn into an equivalent SQL representation.
func (s Order) Compile(layout *Template) (string, error) {
	switch s {
	case Order_Ascendent:
		return layout.AscKeyword, nil
	case Order_Descendent:
		return layout.DescKeyword, nil
	}
	return "", nil
}
