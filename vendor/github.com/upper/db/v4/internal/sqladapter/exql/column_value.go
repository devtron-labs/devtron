package exql

import (
	"github.com/upper/db/v4/internal/cache"
	"strings"
)

// ColumnValue represents a bundle between a column and a corresponding value.
type ColumnValue struct {
	Column   Fragment
	Operator string
	Value    Fragment
}

var _ = Fragment(&ColumnValue{})

type columnValueT struct {
	Column   string
	Operator string
	Value    string
}

// Hash returns a unique identifier for the struct.
func (c *ColumnValue) Hash() uint64 {
	if c == nil {
		return cache.NewHash(FragmentType_ColumnValue, nil)
	}
	return cache.NewHash(FragmentType_ColumnValue, c.Column, c.Operator, c.Value)
}

// Compile transforms the ColumnValue into an equivalent SQL representation.
func (c *ColumnValue) Compile(layout *Template) (compiled string, err error) {
	if z, ok := layout.Read(c); ok {
		return z, nil
	}

	column, err := c.Column.Compile(layout)
	if err != nil {
		return "", err
	}

	data := columnValueT{
		Column:   column,
		Operator: c.Operator,
	}

	if c.Value != nil {
		data.Value, err = c.Value.Compile(layout)
		if err != nil {
			return "", err
		}
	}

	compiled = strings.TrimSpace(layout.MustCompile(layout.ColumnValue, data))

	layout.Write(c, compiled)

	return
}

// ColumnValues represents an array of ColumnValue
type ColumnValues struct {
	ColumnValues []Fragment
}

var _ = Fragment(&ColumnValues{})

// JoinColumnValues returns an array of ColumnValue
func JoinColumnValues(values ...Fragment) *ColumnValues {
	return &ColumnValues{ColumnValues: values}
}

// Insert adds a column to the columns array.
func (c *ColumnValues) Insert(values ...Fragment) *ColumnValues {
	c.ColumnValues = append(c.ColumnValues, values...)
	return c
}

// Hash returns a unique identifier for the struct.
func (c *ColumnValues) Hash() uint64 {
	h := cache.InitHash(FragmentType_ColumnValues)
	for i := range c.ColumnValues {
		h = cache.AddToHash(h, c.ColumnValues[i])
	}
	return h
}

// Compile transforms the ColumnValues into its SQL representation.
func (c *ColumnValues) Compile(layout *Template) (compiled string, err error) {

	if z, ok := layout.Read(c); ok {
		return z, nil
	}

	l := len(c.ColumnValues)

	out := make([]string, l)

	for i := range c.ColumnValues {
		out[i], err = c.ColumnValues[i].Compile(layout)
		if err != nil {
			return "", err
		}
	}

	compiled = strings.TrimSpace(strings.Join(out, layout.IdentifierSeparator))

	layout.Write(c, compiled)

	return
}
