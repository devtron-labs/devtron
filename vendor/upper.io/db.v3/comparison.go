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

package db

import (
	"reflect"
	"time"
)

// Comparison defines methods for representing comparison operators in a
// portable way across databases.
type Comparison interface {
	Operator() ComparisonOperator

	Value() interface{}
}

// ComparisonOperator is a type we use to label comparison operators.
type ComparisonOperator uint8

// Comparison operators
const (
	ComparisonOperatorNone ComparisonOperator = iota

	ComparisonOperatorEqual
	ComparisonOperatorNotEqual

	ComparisonOperatorLessThan
	ComparisonOperatorGreaterThan

	ComparisonOperatorLessThanOrEqualTo
	ComparisonOperatorGreaterThanOrEqualTo

	ComparisonOperatorBetween
	ComparisonOperatorNotBetween

	ComparisonOperatorIn
	ComparisonOperatorNotIn

	ComparisonOperatorIs
	ComparisonOperatorIsNot

	ComparisonOperatorLike
	ComparisonOperatorNotLike

	ComparisonOperatorRegExp
	ComparisonOperatorNotRegExp

	ComparisonOperatorAfter
	ComparisonOperatorBefore

	ComparisonOperatorOnOrAfter
	ComparisonOperatorOnOrBefore
)

type dbComparisonOperator struct {
	t  ComparisonOperator
	op string
	v  interface{}
}

func (c *dbComparisonOperator) CustomOperator() string {
	return c.op
}

func (c *dbComparisonOperator) Operator() ComparisonOperator {
	return c.t
}

func (c *dbComparisonOperator) Value() interface{} {
	return c.v
}

// Gte indicates whether the reference is greater than or equal to the given
// argument.
func Gte(v interface{}) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorGreaterThanOrEqualTo,
		v: v,
	}
}

// Lte indicates whether the reference is less than or equal to the given
// argument.
func Lte(v interface{}) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorLessThanOrEqualTo,
		v: v,
	}
}

// Eq indicates whether the constraint is equal to the given argument.
func Eq(v interface{}) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorEqual,
		v: v,
	}
}

// NotEq indicates whether the constraint is not equal to the given argument.
func NotEq(v interface{}) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorNotEqual,
		v: v,
	}
}

// Gt indicates whether the constraint is greater than the given argument.
func Gt(v interface{}) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorGreaterThan,
		v: v,
	}
}

// Lt indicates whether the constraint is less than the given argument.
func Lt(v interface{}) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorLessThan,
		v: v,
	}
}

// In indicates whether the argument is part of the reference.
func In(v interface{}) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorIn,
		v: toInterfaceArray(v),
	}
}

// NotIn indicates whether the argument is not part of the reference.
func NotIn(v interface{}) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorNotIn,
		v: toInterfaceArray(v),
	}
}

// After indicates whether the reference is after the given time.
func After(t time.Time) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorGreaterThan,
		v: t,
	}
}

// Before indicates whether the reference is before the given time.
func Before(t time.Time) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorLessThan,
		v: t,
	}
}

// OnOrAfter indicater whether the reference is after or equal to the given
// time value.
func OnOrAfter(t time.Time) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorGreaterThanOrEqualTo,
		v: t,
	}
}

// OnOrBefore indicates whether the reference is before or equal to the given
// time value.
func OnOrBefore(t time.Time) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorLessThanOrEqualTo,
		v: t,
	}
}

// Between indicates whether the reference is contained between the two given
// values.
func Between(a interface{}, b interface{}) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorBetween,
		v: []interface{}{a, b},
	}
}

// NotBetween indicates whether the reference is not contained between the two
// given values.
func NotBetween(a interface{}, b interface{}) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorNotBetween,
		v: []interface{}{a, b},
	}
}

// Is indicates whether the reference is nil, true or false.
func Is(v interface{}) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorIs,
		v: v,
	}
}

// IsNot indicates whether the reference is not nil, true nor false.
func IsNot(v interface{}) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorIsNot,
		v: v,
	}
}

// IsNull indicates whether the reference is a NULL value.
func IsNull() Comparison {
	return Is(nil)
}

// IsNotNull indicates whether the reference is a NULL value.
func IsNotNull() Comparison {
	return IsNot(nil)
}

/*
// IsDistinctFrom indicates whether the reference is different from
// the given value, including NULL values.
func IsDistinctFrom(v interface{}) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorIsDistinctFrom,
		v: v,
	}
}

// IsNotDistinctFrom indicates whether the reference is not different from the
// given value, including NULL values.
func IsNotDistinctFrom(v interface{}) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorIsNotDistinctFrom,
		v: v,
	}
}
*/

// Like indicates whether the reference matches the wildcard value.
func Like(v string) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorLike,
		v: v,
	}
}

// NotLike indicates whether the reference does not match the wildcard value.
func NotLike(v string) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorNotLike,
		v: v,
	}
}

/*
// ILike indicates whether the reference matches the wildcard value (case
// insensitive).
func ILike(v string) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorILike,
		v: v,
	}
}

// NotILike indicates whether the reference does not match the wildcard value
// (case insensitive).
func NotILike(v string) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorNotILike,
		v: v,
	}
}
*/

// RegExp indicates whether the reference matches the regexp pattern.
func RegExp(v string) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorRegExp,
		v: v,
	}
}

// NotRegExp indicates whether the reference does not match the regexp pattern.
func NotRegExp(v string) Comparison {
	return &dbComparisonOperator{
		t: ComparisonOperatorNotRegExp,
		v: v,
	}
}

// Op represents a custom comparison operator against the reference.
func Op(customOperator string, v interface{}) Comparison {
	return &dbComparisonOperator{
		op: customOperator,
		t:  ComparisonOperatorNone,
		v:  v,
	}
}

func toInterfaceArray(v interface{}) []interface{} {
	rv := reflect.ValueOf(v)
	switch rv.Type().Kind() {
	case reflect.Ptr:
		return toInterfaceArray(rv.Elem().Interface())
	case reflect.Slice:
		elems := rv.Len()
		args := make([]interface{}, elems)
		for i := 0; i < elems; i++ {
			args[i] = rv.Index(i).Interface()
		}
		return args
	}
	return []interface{}{v}
}

var _ = Comparison(&dbComparisonOperator{})
