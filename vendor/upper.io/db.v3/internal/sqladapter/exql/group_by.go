package exql

// GroupBy represents a SQL's "group by" statement.
type GroupBy struct {
	Columns Fragment
	hash    hash
}

var _ = Fragment(&GroupBy{})

type groupByT struct {
	GroupColumns string
}

// Hash returns a unique identifier.
func (g *GroupBy) Hash() string {
	return g.hash.Hash(g)
}

// GroupByColumns creates and returns a GroupBy with the given column.
func GroupByColumns(columns ...Fragment) *GroupBy {
	return &GroupBy{Columns: JoinColumns(columns...)}
}

func (g *GroupBy) IsEmpty() bool {
	if g == nil || g.Columns == nil {
		return true
	}
	return g.Columns.(hasIsEmpty).IsEmpty()
}

// Compile transforms the GroupBy into an equivalent SQL representation.
func (g *GroupBy) Compile(layout *Template) (compiled string, err error) {

	if c, ok := layout.Read(g); ok {
		return c, nil
	}

	if g.Columns != nil {
		columns, err := g.Columns.Compile(layout)
		if err != nil {
			return "", err
		}

		data := groupByT{
			GroupColumns: columns,
		}
		compiled = layout.MustCompile(layout.GroupByLayout, data)
	}

	layout.Write(g, compiled)

	return
}
