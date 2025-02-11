package exql

import (
	"github.com/upper/db/v4/internal/cache"
)

// Database represents a SQL database.
type Database struct {
	Name string
}

var _ = Fragment(&Database{})

// DatabaseWithName returns a Database with the given name.
func DatabaseWithName(name string) *Database {
	return &Database{Name: name}
}

// Hash returns a unique identifier for the struct.
func (d *Database) Hash() uint64 {
	if d == nil {
		return cache.NewHash(FragmentType_Database, nil)
	}
	return cache.NewHash(FragmentType_Database, d.Name)
}

// Compile transforms the Database into an equivalent SQL representation.
func (d *Database) Compile(layout *Template) (compiled string, err error) {
	if c, ok := layout.Read(d); ok {
		return c, nil
	}

	compiled = layout.MustCompile(layout.IdentifierQuote, Raw{Value: d.Name})

	layout.Write(d, compiled)
	return
}
