package exql

import (
	"bytes"
	"reflect"
	"sync"
	"text/template"

	"github.com/upper/db/v4/internal/adapter"
	"github.com/upper/db/v4/internal/cache"
)

// Type is the type of SQL query the statement represents.
type Type uint8

// Values for Type.
const (
	NoOp Type = iota

	Truncate
	DropTable
	DropDatabase
	Count
	Insert
	Select
	Update
	Delete

	SQL
)

func (t Type) Hash() uint64 {
	return cache.NewHash(FragmentType_StatementType, uint8(t))
}

type (
	// Limit represents the SQL limit in a query.
	Limit int64
	// Offset represents the SQL offset in a query.
	Offset int64
)

func (t Limit) Hash() uint64 {
	return cache.NewHash(FragmentType_Limit, uint64(t))
}

func (t Offset) Hash() uint64 {
	return cache.NewHash(FragmentType_Offset, uint64(t))
}

// Template is an SQL template.
type Template struct {
	AndKeyword          string
	AscKeyword          string
	AssignmentOperator  string
	ClauseGroup         string
	ClauseOperator      string
	ColumnAliasLayout   string
	ColumnSeparator     string
	ColumnValue         string
	CountLayout         string
	DeleteLayout        string
	DescKeyword         string
	DropDatabaseLayout  string
	DropTableLayout     string
	GroupByLayout       string
	IdentifierQuote     string
	IdentifierSeparator string
	InsertLayout        string
	JoinLayout          string
	OnLayout            string
	OrKeyword           string
	OrderByLayout       string
	SelectLayout        string
	SortByColumnLayout  string
	TableAliasLayout    string
	TruncateLayout      string
	UpdateLayout        string
	UsingLayout         string
	ValueQuote          string
	ValueSeparator      string
	WhereLayout         string

	ComparisonOperator map[adapter.ComparisonOperator]string

	templateMutex sync.RWMutex
	templateMap   map[string]*template.Template

	*cache.Cache
}

func (layout *Template) MustCompile(templateText string, data interface{}) string {
	var b bytes.Buffer

	v, ok := layout.getTemplate(templateText)
	if !ok {
		v = template.
			Must(template.New("").
				Funcs(map[string]interface{}{
					"defined": func(in Fragment) bool {
						if in == nil || reflect.ValueOf(in).IsNil() {
							return false
						}
						if check, ok := in.(hasIsEmpty); ok {
							if check.IsEmpty() {
								return false
							}
						}
						return true
					},
					"compile": func(in Fragment) (string, error) {
						s, err := layout.doCompile(in)
						if err != nil {
							return "", err
						}
						return s, nil
					},
				}).
				Parse(templateText))

		layout.setTemplate(templateText, v)
	}

	if err := v.Execute(&b, data); err != nil {
		panic("There was an error compiling the following template:\n" + templateText + "\nError was: " + err.Error())
	}

	return b.String()
}

func (t *Template) getTemplate(k string) (*template.Template, bool) {
	t.templateMutex.RLock()
	defer t.templateMutex.RUnlock()

	if t.templateMap == nil {
		t.templateMap = make(map[string]*template.Template)
	}

	v, ok := t.templateMap[k]
	return v, ok
}

func (t *Template) setTemplate(k string, v *template.Template) {
	t.templateMutex.Lock()
	defer t.templateMutex.Unlock()

	t.templateMap[k] = v
}
