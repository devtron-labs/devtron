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
	"fmt"
	"sort"

	"github.com/upper/db/v4/internal/adapter"
)

// LogicalExpr represents an expression to be used in logical statements.
type LogicalExpr = adapter.LogicalExpr

// LogicalOperator represents a logical operation.
type LogicalOperator = adapter.LogicalOperator

// Cond is a map that defines conditions for a query.
//
// Each entry of the map represents a condition (a column-value relation bound
// by a comparison Operator). The comparison can be specified after the column
// name, if no comparison operator is provided the equality operator is used as
// default.
//
// Examples:
//
//  // Age equals 18.
//  db.Cond{"age": 18}
//
//  // Age is greater than or equal to 18.
//  db.Cond{"age >=": 18}
//
//  // id is any of the values 1, 2 or 3.
//  db.Cond{"id IN": []{1, 2, 3}}
//
//  // Age is lower than 18 (MongoDB syntax)
//  db.Cond{"age $lt": 18}
//
//  // age > 32 and age < 35
//  db.Cond{"age >": 32, "age <": 35}
type Cond map[interface{}]interface{}

// Empty returns false if there are no conditions.
func (c Cond) Empty() bool {
	for range c {
		return false
	}
	return true
}

// Constraints returns each one of the Cond map entires as a constraint.
func (c Cond) Constraints() []adapter.Constraint {
	z := make([]adapter.Constraint, 0, len(c))
	for _, k := range c.keys() {
		z = append(z, adapter.NewConstraint(k, c[k]))
	}
	return z
}

// Operator returns the equality operator.
func (c Cond) Operator() LogicalOperator {
	return adapter.DefaultLogicalOperator
}

func (c Cond) keys() []interface{} {
	keys := make(condKeys, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	if len(c) > 1 {
		sort.Sort(keys)
	}
	return keys
}

// Expressions returns all the expressions contained in the condition.
func (c Cond) Expressions() []LogicalExpr {
	z := make([]LogicalExpr, 0, len(c))
	for _, k := range c.keys() {
		z = append(z, Cond{k: c[k]})
	}
	return z
}

type condKeys []interface{}

func (ck condKeys) Len() int {
	return len(ck)
}

func (ck condKeys) Less(i, j int) bool {
	return fmt.Sprintf("%v", ck[i]) < fmt.Sprintf("%v", ck[j])
}

func (ck condKeys) Swap(i, j int) {
	ck[i], ck[j] = ck[j], ck[i]
}

func defaultJoin(in ...adapter.LogicalExpr) []adapter.LogicalExpr {
	for i := range in {
		cond, ok := in[i].(Cond)
		if ok && !cond.Empty() {
			in[i] = And(cond)
		}
	}
	return in
}

var (
	_ = LogicalExpr(Cond{})
)
