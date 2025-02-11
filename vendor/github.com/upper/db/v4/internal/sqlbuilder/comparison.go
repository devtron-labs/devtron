package sqlbuilder

import (
	"fmt"
	"strings"

	db "github.com/upper/db/v4"
	"github.com/upper/db/v4/internal/adapter"
	"github.com/upper/db/v4/internal/sqladapter/exql"
)

var comparisonOperators = map[adapter.ComparisonOperator]string{
	adapter.ComparisonOperatorEqual:    "=",
	adapter.ComparisonOperatorNotEqual: "!=",

	adapter.ComparisonOperatorLessThan:    "<",
	adapter.ComparisonOperatorGreaterThan: ">",

	adapter.ComparisonOperatorLessThanOrEqualTo:    "<=",
	adapter.ComparisonOperatorGreaterThanOrEqualTo: ">=",

	adapter.ComparisonOperatorBetween:    "BETWEEN",
	adapter.ComparisonOperatorNotBetween: "NOT BETWEEN",

	adapter.ComparisonOperatorIn:    "IN",
	adapter.ComparisonOperatorNotIn: "NOT IN",

	adapter.ComparisonOperatorIs:    "IS",
	adapter.ComparisonOperatorIsNot: "IS NOT",

	adapter.ComparisonOperatorLike:    "LIKE",
	adapter.ComparisonOperatorNotLike: "NOT LIKE",

	adapter.ComparisonOperatorRegExp:    "REGEXP",
	adapter.ComparisonOperatorNotRegExp: "NOT REGEXP",
}

type operatorWrapper struct {
	tu *templateWithUtils
	cv *exql.ColumnValue

	op *adapter.Comparison
	v  interface{}
}

func (ow *operatorWrapper) cmp() *adapter.Comparison {
	if ow.op != nil {
		return ow.op
	}

	if ow.cv.Operator != "" {
		return db.Op(ow.cv.Operator, ow.v).Comparison
	}

	if ow.v == nil {
		return db.Is(nil).Comparison
	}

	args, isSlice := toInterfaceArguments(ow.v)
	if isSlice {
		return db.In(args...).Comparison
	}

	return db.Eq(ow.v).Comparison
}

func (ow *operatorWrapper) preprocess() (string, []interface{}) {
	placeholder := "?"

	column, err := ow.cv.Column.Compile(ow.tu.Template)
	if err != nil {
		panic(fmt.Sprintf("could not compile column: %v", err.Error()))
	}

	c := ow.cmp()

	op := ow.tu.comparisonOperatorMapper(c.Operator())

	var args []interface{}

	switch c.Operator() {
	case adapter.ComparisonOperatorNone:
		panic("no operator given")
	case adapter.ComparisonOperatorCustom:
		op = c.CustomOperator()
	case adapter.ComparisonOperatorIn, adapter.ComparisonOperatorNotIn:
		values := c.Value().([]interface{})
		if len(values) < 1 {
			placeholder, args = "(NULL)", []interface{}{}
			break
		}
		placeholder, args = "(?"+strings.Repeat(", ?", len(values)-1)+")", values
	case adapter.ComparisonOperatorIs, adapter.ComparisonOperatorIsNot:
		switch c.Value() {
		case nil:
			placeholder, args = "NULL", []interface{}{}
		case false:
			placeholder, args = "FALSE", []interface{}{}
		case true:
			placeholder, args = "TRUE", []interface{}{}
		}
	case adapter.ComparisonOperatorBetween, adapter.ComparisonOperatorNotBetween:
		values := c.Value().([]interface{})
		placeholder, args = "? AND ?", []interface{}{values[0], values[1]}
	case adapter.ComparisonOperatorEqual:
		v := c.Value()
		if b, ok := v.([]byte); ok {
			v = string(b)
		}
		args = []interface{}{v}
	}

	if args == nil {
		args = []interface{}{c.Value()}
	}

	if strings.Contains(op, ":column") {
		return strings.Replace(op, ":column", column, -1), args
	}

	return column + " " + op + " " + placeholder, args
}
