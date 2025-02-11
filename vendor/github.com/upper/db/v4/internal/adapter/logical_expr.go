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

package adapter

import (
	"github.com/upper/db/v4/internal/immutable"
)

// LogicalExpr represents a group formed by one or more sentences joined by
// an Operator like "AND" or "OR".
type LogicalExpr interface {
	// Expressions returns child sentences.
	Expressions() []LogicalExpr

	// Operator returns the Operator that joins all the sentences in the group.
	Operator() LogicalOperator

	// Empty returns true if the compound has zero children, false otherwise.
	Empty() bool
}

// LogicalOperator represents the operation on a compound statement.
type LogicalOperator uint

// LogicalExpr Operators.
const (
	LogicalOperatorNone LogicalOperator = iota
	LogicalOperatorAnd
	LogicalOperatorOr
)

const DefaultLogicalOperator = LogicalOperatorAnd

type LogicalExprGroup struct {
	op LogicalOperator

	prev *LogicalExprGroup
	fn   func(*[]LogicalExpr) error
}

func NewLogicalExprGroup(op LogicalOperator, conds ...LogicalExpr) *LogicalExprGroup {
	group := &LogicalExprGroup{op: op}
	if len(conds) == 0 {
		return group
	}
	return group.Frame(func(in *[]LogicalExpr) error {
		*in = append(*in, conds...)
		return nil
	})
}

// Expressions returns each one of the conditions as a compound.
func (g *LogicalExprGroup) Expressions() []LogicalExpr {
	conds, err := immutable.FastForward(g)
	if err == nil {
		return *(conds.(*[]LogicalExpr))
	}
	return nil
}

// Operator is undefined for a logical group.
func (g *LogicalExprGroup) Operator() LogicalOperator {
	if g.op == LogicalOperatorNone {
		panic("operator is not defined")
	}
	return g.op
}

// Empty returns true if this condition has no elements. False otherwise.
func (g *LogicalExprGroup) Empty() bool {
	if g.fn != nil {
		return false
	}
	if g.prev != nil {
		return g.prev.Empty()
	}
	return true
}

func (g *LogicalExprGroup) Frame(fn func(*[]LogicalExpr) error) *LogicalExprGroup {
	return &LogicalExprGroup{prev: g, op: g.op, fn: fn}
}

func (g *LogicalExprGroup) Prev() immutable.Immutable {
	if g == nil {
		return nil
	}
	return g.prev
}

func (g *LogicalExprGroup) Fn(in interface{}) error {
	if g.fn == nil {
		return nil
	}
	return g.fn(in.(*[]LogicalExpr))
}

func (g *LogicalExprGroup) Base() interface{} {
	return &[]LogicalExpr{}
}

var (
	_ = immutable.Immutable(&LogicalExprGroup{})
)
