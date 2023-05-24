package exql

import (
	"errors"
	"reflect"
	"strings"
)

var errUnknownTemplateType = errors.New("Unknown template type")

//  represents different kinds of SQL statements.
type Statement struct {
	Type
	Table        Fragment
	Database     Fragment
	Columns      Fragment
	Values       Fragment
	Distinct     bool
	ColumnValues Fragment
	OrderBy      Fragment
	GroupBy      Fragment
	Joins        Fragment
	Where        Fragment
	Returning    Fragment

	Limit
	Offset

	SQL string

	hash    hash
	amendFn func(string) string
}

func (layout *Template) doCompile(c Fragment) (string, error) {
	if c != nil && !reflect.ValueOf(c).IsNil() {
		return c.Compile(layout)
	}
	return "", nil
}

// Hash returns a unique identifier for the struct.
func (s *Statement) Hash() string {
	return s.hash.Hash(s)
}

func (s *Statement) SetAmendment(amendFn func(string) string) {
	s.amendFn = amendFn
}

func (s *Statement) Amend(in string) string {
	if s.amendFn == nil {
		return in
	}
	return s.amendFn(in)
}

func (s *Statement) template(layout *Template) (string, error) {
	switch s.Type {
	case Truncate:
		return layout.TruncateLayout, nil
	case DropTable:
		return layout.DropTableLayout, nil
	case DropDatabase:
		return layout.DropDatabaseLayout, nil
	case Count:
		return layout.CountLayout, nil
	case Select:
		return layout.SelectLayout, nil
	case Delete:
		return layout.DeleteLayout, nil
	case Update:
		return layout.UpdateLayout, nil
	case Insert:
		return layout.InsertLayout, nil
	default:
		return "", errUnknownTemplateType
	}
}

// Compile transforms the Statement into an equivalent SQL query.
func (s *Statement) Compile(layout *Template) (compiled string, err error) {
	if s.Type == SQL {
		// No need to hit the cache.
		return s.SQL, nil
	}

	if z, ok := layout.Read(s); ok {
		return s.Amend(z), nil
	}

	tpl, err := s.template(layout)
	if err != nil {
		return "", err
	}

	compiled = layout.MustCompile(tpl, s)

	compiled = strings.TrimSpace(compiled)
	layout.Write(s, compiled)

	return s.Amend(compiled), nil
}

// RawSQL represents a raw SQL statement.
func RawSQL(s string) *Statement {
	return &Statement{
		Type: SQL,
		SQL:  s,
	}
}
