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

// AndExpr represents an expression joined by a logical conjuction (AND).
type AndExpr struct {
	*adapter.LogicalExprGroup
}

// And adds more expressions to the group.
func (a *AndExpr) And(andConds ...LogicalExpr) *AndExpr {
	var fn func(*[]LogicalExpr) error
	if len(andConds) > 0 {
		fn = func(in *[]LogicalExpr) error {
			*in = append(*in, andConds...)
			return nil
		}
	}
	return &AndExpr{a.LogicalExprGroup.Frame(fn)}
}

// Empty returns false if the expressions has zero conditions.
func (a *AndExpr) Empty() bool {
	return a.LogicalExprGroup.Empty()
}

// And joins conditions under logical conjunction. Conditions can be
// represented by `db.Cond{}`, `db.Or()` or `db.And()`.
//
// Examples:
//
//	// name = "Peter" AND last_name = "Parker"
//	db.And(
//		db.Cond{"name": "Peter"},
//		db.Cond{"last_name": "Parker "},
//	)
//
//	// (name = "Peter" OR name = "Mickey") AND last_name = "Mouse"
//	db.And(
//		db.Or(
//			db.Cond{"name": "Peter"},
//			db.Cond{"name": "Mickey"},
//		),
//		db.Cond{"last_name": "Mouse"},
//	)
func And(conds ...LogicalExpr) *AndExpr {
	return &AndExpr{adapter.NewLogicalExprGroup(adapter.LogicalOperatorAnd, conds...)}
}

var _ = adapter.LogicalExpr(&AndExpr{})
