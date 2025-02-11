package exql

import (
	"github.com/upper/db/v4/internal/cache"
)

const (
	defaultColumnSeparator     = `.`
	defaultIdentifierSeparator = `, `
	defaultIdentifierQuote     = `"{{.Value}}"`
	defaultValueSeparator      = `, `
	defaultValueQuote          = `'{{.}}'`
	defaultAndKeyword          = `AND`
	defaultOrKeyword           = `OR`
	defaultDescKeyword         = `DESC`
	defaultAscKeyword          = `ASC`
	defaultAssignmentOperator  = `=`
	defaultClauseGroup         = `({{.}})`
	defaultClauseOperator      = ` {{.}} `
	defaultColumnValue         = `{{.Column}} {{.Operator}} {{.Value}}`
	defaultTableAliasLayout    = `{{.Name}}{{if .Alias}} AS {{.Alias}}{{end}}`
	defaultColumnAliasLayout   = `{{.Name}}{{if .Alias}} AS {{.Alias}}{{end}}`
	defaultSortByColumnLayout  = `{{.Column}} {{.Order}}`

	defaultOrderByLayout = `
    {{if .SortColumns}}
      ORDER BY {{.SortColumns}}
    {{end}}
  `

	defaultWhereLayout = `
    {{if .Conds}}
      WHERE {{.Conds}}
    {{end}}
  `

	defaultUsingLayout = `
    {{if .Columns}}
      USING ({{.Columns}})
    {{end}}
  `

	defaultJoinLayout = `
    {{if .Table}}
      {{ if .On }}
        {{.Type}} JOIN {{.Table}}
        {{.On}}
      {{ else if .Using }}
        {{.Type}} JOIN {{.Table}}
        {{.Using}}
      {{ else if .Type | eq "CROSS" }}
        {{.Type}} JOIN {{.Table}}
      {{else}}
        NATURAL {{.Type}} JOIN {{.Table}}
      {{end}}
    {{end}}
  `

	defaultOnLayout = `
    {{if .Conds}}
      ON {{.Conds}}
    {{end}}
  `

	defaultSelectLayout = `
    SELECT
      {{if .Distinct}}
        DISTINCT
      {{end}}

      {{if .Columns}}
        {{.Columns | compile}}
      {{else}}
        *
      {{end}}

      {{if defined .Table}}
        FROM {{.Table | compile}}
      {{end}}

      {{.Joins | compile}}

      {{.Where | compile}}

      {{.GroupBy | compile}}

      {{.OrderBy | compile}}

      {{if .Limit}}
        LIMIT {{.Limit}}
      {{end}}

      {{if .Offset}}
        OFFSET {{.Offset}}
      {{end}}
  `
	defaultDeleteLayout = `
    DELETE
      FROM {{.Table | compile}}
      {{.Where | compile}}
    {{if .Limit}}
      LIMIT {{.Limit}}
    {{end}}
    {{if .Offset}}
      OFFSET {{.Offset}}
    {{end}}
  `
	defaultUpdateLayout = `
    UPDATE
      {{.Table | compile}}
    SET {{.ColumnValues | compile}}
      {{.Where | compile}}
  `

	defaultCountLayout = `
    SELECT
      COUNT(1) AS _t
    FROM {{.Table | compile}}
      {{.Where | compile}}

      {{if .Limit}}
        LIMIT {{.Limit | compile}}
      {{end}}

      {{if .Offset}}
        OFFSET {{.Offset}}
      {{end}}
  `

	defaultInsertLayout = `
    INSERT INTO {{.Table | compile}}
      {{if .Columns }}({{.Columns | compile}}){{end}}
    VALUES
      {{.Values | compile}}
    {{if .Returning}}
      RETURNING {{.Returning | compile}}
    {{end}}
  `

	defaultTruncateLayout = `
    TRUNCATE TABLE {{.Table | compile}}
  `

	defaultDropDatabaseLayout = `
    DROP DATABASE {{.Database | compile}}
  `

	defaultDropTableLayout = `
    DROP TABLE {{.Table | compile}}
  `

	defaultGroupByLayout = `
    {{if .GroupColumns}}
      GROUP BY {{.GroupColumns}}
    {{end}}
  `
)

var defaultTemplate = &Template{
	AndKeyword:          defaultAndKeyword,
	AscKeyword:          defaultAscKeyword,
	AssignmentOperator:  defaultAssignmentOperator,
	ClauseGroup:         defaultClauseGroup,
	ClauseOperator:      defaultClauseOperator,
	ColumnAliasLayout:   defaultColumnAliasLayout,
	ColumnSeparator:     defaultColumnSeparator,
	ColumnValue:         defaultColumnValue,
	CountLayout:         defaultCountLayout,
	DeleteLayout:        defaultDeleteLayout,
	DescKeyword:         defaultDescKeyword,
	DropDatabaseLayout:  defaultDropDatabaseLayout,
	DropTableLayout:     defaultDropTableLayout,
	GroupByLayout:       defaultGroupByLayout,
	IdentifierQuote:     defaultIdentifierQuote,
	IdentifierSeparator: defaultIdentifierSeparator,
	InsertLayout:        defaultInsertLayout,
	JoinLayout:          defaultJoinLayout,
	OnLayout:            defaultOnLayout,
	OrKeyword:           defaultOrKeyword,
	OrderByLayout:       defaultOrderByLayout,
	SelectLayout:        defaultSelectLayout,
	SortByColumnLayout:  defaultSortByColumnLayout,
	TableAliasLayout:    defaultTableAliasLayout,
	TruncateLayout:      defaultTruncateLayout,
	UpdateLayout:        defaultUpdateLayout,
	UsingLayout:         defaultUsingLayout,
	ValueQuote:          defaultValueQuote,
	ValueSeparator:      defaultValueSeparator,
	WhereLayout:         defaultWhereLayout,

	Cache: cache.NewCache(),
}
