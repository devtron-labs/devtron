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
	"github.com/upper/db/v4/internal/adapter"
)

// OrExpr represents a logical expression joined by logical disjunction (OR).
type OrExpr struct {
	*adapter.LogicalExprGroup
}

// Or adds more expressions to the group.
func (o *OrExpr) Or(orConds ...LogicalExpr) *OrExpr {
	var fn func(*[]LogicalExpr) error
	if len(orConds) > 0 {
		fn = func(in *[]LogicalExpr) error {
			*in = append(*in, orConds...)
			return nil
		}
	}
	return &OrExpr{o.LogicalExprGroup.Frame(fn)}
}

// Empty returns false if the expressions has zero conditions.
func (o *OrExpr) Empty() bool {
	return o.LogicalExprGroup.Empty()
}

// Or joins conditions under logical disjunction. Conditions can be represented
// by `db.Cond{}`, `db.Or()` or `db.And()`.
//
// Example:
//
//	// year = 2012 OR year = 1987
//	db.Or(
//		db.Cond{"year": 2012},
//		db.Cond{"year": 1987},
//	)
func Or(conds ...LogicalExpr) *OrExpr {
	return &OrExpr{adapter.NewLogicalExprGroup(adapter.LogicalOperatorOr, defaultJoin(conds...)...)}
}

var _ = adapter.LogicalExpr(&OrExpr{})
