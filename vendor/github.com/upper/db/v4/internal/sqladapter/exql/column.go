package exql

import (
	"fmt"
	"strings"

	"github.com/upper/db/v4/internal/cache"
)

type columnWithAlias struct {
	Name  string
	Alias string
}

// Column represents a SQL column.
type Column struct {
	Name interface{}
}

var _ = Fragment(&Column{})

// ColumnWithName creates and returns a Column with the given name.
func ColumnWithName(name string) *Column {
	return &Column{Name: name}
}

// Hash returns a unique identifier for the struct.
func (c *Column) Hash() uint64 {
	if c == nil {
		return cache.NewHash(FragmentType_Column, nil)
	}
	return cache.NewHash(FragmentType_Column, c.Name)
}

// Compile transforms the ColumnValue into an equivalent SQL representation.
func (c *Column) Compile(layout *Template) (compiled string, err error) {
	if z, ok := layout.Read(c); ok {
		return z, nil
	}

	var alias string
	switch value := c.Name.(type) {
	case string:
		value = trimString(value)

		chunks := separateByAS(value)
		if len(chunks) == 1 {
			chunks = separateBySpace(value)
		}

		name := chunks[0]
		nameChunks := strings.SplitN(name, layout.ColumnSeparator, 2)

		for i := range nameChunks {
			nameChunks[i] = trimString(nameChunks[i])
			if nameChunks[i] == "*" {
				continue
			}
			nameChunks[i] = layout.MustCompile(layout.IdentifierQuote, Raw{Value: nameChunks[i]})
		}

		compiled = strings.Join(nameChunks, layout.ColumnSeparator)

		if len(chunks) > 1 {
			alias = trimString(chunks[1])
			alias = layout.MustCompile(layout.IdentifierQuote, Raw{Value: alias})
		}
	case compilable:
		compiled, err = value.Compile(layout)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf(errExpectingHashableFmt, c.Name)
	}

	if alias != "" {
		compiled = layout.MustCompile(layout.ColumnAliasLayout, columnWithAlias{compiled, alias})
	}

	layout.Write(c, compiled)
	return
}
