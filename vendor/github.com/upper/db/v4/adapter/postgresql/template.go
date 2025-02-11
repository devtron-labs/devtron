// Copyright (c) 2012-present The upper.io/db authors. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package postgresql

import (
	"github.com/upper/db/v4/internal/adapter"
	"github.com/upper/db/v4/internal/cache"
	"github.com/upper/db/v4/internal/sqladapter/exql"
)

const (
	adapterColumnSeparator     = `.`
	adapterIdentifierSeparator = `, `
	adapterIdentifierQuote     = `"{{.Value}}"`
	adapterValueSeparator      = `, `
	adapterValueQuote          = `'{{.}}'`
	adapterAndKeyword          = `AND`
	adapterOrKeyword           = `OR`
	adapterDescKeyword         = `DESC`
	adapterAscKeyword          = `ASC`
	adapterAssignmentOperator  = `=`
	adapterClauseGroup         = `({{.}})`
	adapterClauseOperator      = ` {{.}} `
	adapterColumnValue         = `{{.Column}} {{.Operator}} {{.Value}}`
	adapterTableAliasLayout    = `{{.Name}}{{if .Alias}} AS {{.Alias}}{{end}}`
	adapterColumnAliasLayout   = `{{.Name}}{{if .Alias}} AS {{.Alias}}{{end}}`
	adapterSortByColumnLayout  = `{{.Column}} {{.Order}}`

	adapterOrderByLayout = `
    {{if .SortColumns}}
      ORDER BY {{.SortColumns}}
    {{end}}
  `

	adapterWhereLayout = `
    {{if .Conds}}
      WHERE {{.Conds}}
    {{end}}
  `

	adapterUsingLayout = `
    {{if .Columns}}
      USING ({{.Columns}})
    {{end}}
  `

	adapterJoinLayout = `
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

	adapterOnLayout = `
    {{if .Conds}}
      ON {{.Conds}}
    {{end}}
  `

	adapterSelectLayout = `
    SELECT
      {{if .Distinct}}
        DISTINCT
      {{end}}

      {{if defined .Columns}}
        {{.Columns | compile}}
      {{else}}
        *
      {{end}}

      {{if defined .Table}}
        FROM {{.Table | compile}}
      {{end}}

      {{.Joins | compile}}

      {{.Where | compile}}

      {{if defined .GroupBy}}
        {{.GroupBy | compile}}
      {{end}}

      {{.OrderBy | compile}}

      {{if .Limit}}
        LIMIT {{.Limit}}
      {{end}}

      {{if .Offset}}
        OFFSET {{.Offset}}
      {{end}}
  `
	adapterDeleteLayout = `
    DELETE
      FROM {{.Table | compile}}
      {{.Where | compile}}
  `
	adapterUpdateLayout = `
    UPDATE
      {{.Table | compile}}
    SET {{.ColumnValues | compile}}
      {{.Where | compile}}
  `

	adapterSelectCountLayout = `
    SELECT
      COUNT(1) AS _t
    FROM {{.Table | compile}}
      {{.Where | compile}}
  `

	adapterInsertLayout = `
    INSERT INTO {{.Table | compile}}
      {{if defined .Columns}}({{.Columns | compile}}){{end}}
    VALUES
    {{if defined .Values}}
      {{.Values | compile}}
    {{else}}
      (default)
    {{end}}
    {{if defined .Returning}}
      RETURNING {{.Returning | compile}}
    {{end}}
  `

	adapterTruncateLayout = `
    TRUNCATE TABLE {{.Table | compile}} RESTART IDENTITY
  `

	adapterDropDatabaseLayout = `
    DROP DATABASE {{.Database | compile}}
  `

	adapterDropTableLayout = `
    DROP TABLE {{.Table | compile}}
  `

	adapterGroupByLayout = `
    {{if .GroupColumns}}
      GROUP BY {{.GroupColumns}}
    {{end}}
  `
)

var template = &exql.Template{
	ColumnSeparator:     adapterColumnSeparator,
	IdentifierSeparator: adapterIdentifierSeparator,
	IdentifierQuote:     adapterIdentifierQuote,
	ValueSeparator:      adapterValueSeparator,
	ValueQuote:          adapterValueQuote,
	AndKeyword:          adapterAndKeyword,
	OrKeyword:           adapterOrKeyword,
	DescKeyword:         adapterDescKeyword,
	AscKeyword:          adapterAscKeyword,
	AssignmentOperator:  adapterAssignmentOperator,
	ClauseGroup:         adapterClauseGroup,
	ClauseOperator:      adapterClauseOperator,
	ColumnValue:         adapterColumnValue,
	TableAliasLayout:    adapterTableAliasLayout,
	ColumnAliasLayout:   adapterColumnAliasLayout,
	SortByColumnLayout:  adapterSortByColumnLayout,
	WhereLayout:         adapterWhereLayout,
	JoinLayout:          adapterJoinLayout,
	OnLayout:            adapterOnLayout,
	UsingLayout:         adapterUsingLayout,
	OrderByLayout:       adapterOrderByLayout,
	InsertLayout:        adapterInsertLayout,
	SelectLayout:        adapterSelectLayout,
	UpdateLayout:        adapterUpdateLayout,
	DeleteLayout:        adapterDeleteLayout,
	TruncateLayout:      adapterTruncateLayout,
	DropDatabaseLayout:  adapterDropDatabaseLayout,
	DropTableLayout:     adapterDropTableLayout,
	CountLayout:         adapterSelectCountLayout,
	GroupByLayout:       adapterGroupByLayout,
	Cache:               cache.NewCache(),
	ComparisonOperator: map[adapter.ComparisonOperator]string{
		adapter.ComparisonOperatorRegExp:    "~",
		adapter.ComparisonOperatorNotRegExp: "!~",
	},
}
