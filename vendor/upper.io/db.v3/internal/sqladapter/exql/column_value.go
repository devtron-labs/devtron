package exql

import (
	"strings"
)

// ColumnValue represents a bundle between a column and a corresponding value.
type ColumnValue struct {
	Column   Fragment
	Operator string
	Value    Fragment
	hash     hash
}

var _ = Fragment(&ColumnValue{})

type columnValueT struct {
	Column   string
	Operator string
	Value    string
}

// Hash returns a unique identifier for the struct.
func (c *ColumnValue) Hash() string {
	return c.hash.Hash(c)
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
	hash         hash
}

var _ = Fragment(&ColumnValues{})

// JoinColumnValues returns an array of ColumnValue
func JoinColumnValues(values ...Fragment) *ColumnValues {
	return &ColumnValues{ColumnValues: values}
}

// Insert adds a column to the columns array.
func (c *ColumnValues) Insert(values ...Fragment) *ColumnValues {
	c.ColumnValues = append(c.ColumnValues, values...)
	c.hash.Reset()
	return c
}

// Hash returns a unique identifier for the struct.
func (c *ColumnValues) Hash() string {
	return c.hash.Hash(c)
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
