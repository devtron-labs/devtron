package sqlbuilder

import (
	"fmt"
	"strings"

	db "upper.io/db.v3"
	"upper.io/db.v3/internal/sqladapter/exql"
)

var comparisonOperators = map[db.ComparisonOperator]string{
	db.ComparisonOperatorEqual:    "=",
	db.ComparisonOperatorNotEqual: "!=",

	db.ComparisonOperatorLessThan:    "<",
	db.ComparisonOperatorGreaterThan: ">",

	db.ComparisonOperatorLessThanOrEqualTo:    "<=",
	db.ComparisonOperatorGreaterThanOrEqualTo: ">=",

	db.ComparisonOperatorBetween:    "BETWEEN",
	db.ComparisonOperatorNotBetween: "NOT BETWEEN",

	db.ComparisonOperatorIn:    "IN",
	db.ComparisonOperatorNotIn: "NOT IN",

	db.ComparisonOperatorIs:    "IS",
	db.ComparisonOperatorIsNot: "IS NOT",

	db.ComparisonOperatorLike:    "LIKE",
	db.ComparisonOperatorNotLike: "NOT LIKE",

	db.ComparisonOperatorRegExp:    "REGEXP",
	db.ComparisonOperatorNotRegExp: "NOT REGEXP",
}

type hasCustomOperator interface {
	CustomOperator() string
}

type operatorWrapper struct {
	tu *templateWithUtils
	cv *exql.ColumnValue

	op db.Comparison
	v  interface{}
}

func (ow *operatorWrapper) cmp() db.Comparison {
	if ow.op != nil {
		return ow.op
	}

	if ow.cv.Operator != "" {
		return db.Op(ow.cv.Operator, ow.v)
	}

	if ow.v == nil {
		return db.Is(nil)
	}

	args, isSlice := toInterfaceArguments(ow.v)
	if isSlice {
		return db.In(args)
	}

	return db.Eq(ow.v)
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
	case db.ComparisonOperatorNone:
		if c, ok := c.(hasCustomOperator); ok {
			op = c.CustomOperator()
		} else {
			panic("no operator given")
		}
	case db.ComparisonOperatorIn, db.ComparisonOperatorNotIn:
		values := c.Value().([]interface{})
		if len(values) < 1 {
			placeholder, args = "(NULL)", []interface{}{}
			break
		}
		placeholder, args = "(?"+strings.Repeat(", ?", len(values)-1)+")", values
	case db.ComparisonOperatorIs, db.ComparisonOperatorIsNot:
		switch c.Value() {
		case nil:
			placeholder, args = "NULL", []interface{}{}
		case false:
			placeholder, args = "FALSE", []interface{}{}
		case true:
			placeholder, args = "TRUE", []interface{}{}
		}
	case db.ComparisonOperatorBetween, db.ComparisonOperatorNotBetween:
		values := c.Value().([]interface{})
		placeholder, args = "? AND ?", []interface{}{values[0], values[1]}
	case db.ComparisonOperatorEqual:
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
