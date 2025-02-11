package exql

import (
	"strings"

	"github.com/upper/db/v4/internal/cache"
)

// Columns represents an array of Column.
type Columns struct {
	Columns []Fragment
}

var _ = Fragment(&Columns{})

// Hash returns a unique identifier.
func (c *Columns) Hash() uint64 {
	if c == nil {
		return cache.NewHash(FragmentType_Columns, nil)
	}
	h := cache.InitHash(FragmentType_Columns)
	for i := range c.Columns {
		h = cache.AddToHash(h, c.Columns[i])
	}
	return h
}

// JoinColumns creates and returns an array of Column.
func JoinColumns(columns ...Fragment) *Columns {
	return &Columns{Columns: columns}
}

// OnConditions creates and retuens a new On.
func OnConditions(conditions ...Fragment) *On {
	return &On{Conditions: conditions}
}

// UsingColumns builds a Using from the given columns.
func UsingColumns(columns ...Fragment) *Using {
	return &Using{Columns: columns}
}

// Append
func (c *Columns) Append(a *Columns) *Columns {
	c.Columns = append(c.Columns, a.Columns...)
	return c
}

// IsEmpty
func (c *Columns) IsEmpty() bool {
	if c == nil || len(c.Columns) < 1 {
		return true
	}
	return false
}

// Compile transforms the Columns into an equivalent SQL representation.
func (c *Columns) Compile(layout *Template) (compiled string, err error) {
	if z, ok := layout.Read(c); ok {
		return z, nil
	}

	l := len(c.Columns)

	if l > 0 {
		out := make([]string, l)

		for i := 0; i < l; i++ {
			out[i], err = c.Columns[i].Compile(layout)
			if err != nil {
				return "", err
			}
		}

		compiled = strings.Join(out, layout.IdentifierSeparator)
	} else {
		compiled = "*"
	}

	layout.Write(c, compiled)

	return
}
