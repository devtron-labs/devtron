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
	"upper.io/db.v3/internal/immutable"
)

// Compound represents an statement that has one or many sentences joined by by
// an operator like "AND" or "OR". This is an exported interface but it was
// designed for internal usage, you may want to use the `db.And()` or `db.Or()`
// functions instead.
type Compound interface {
	// Sentences returns child sentences.
	Sentences() []Compound

	// Operator returns the operator that joins the compound's child sentences.
	Operator() CompoundOperator

	// Empty returns true if the compound has zero children, false otherwise.
	Empty() bool
}

// CompoundOperator represents the operation on a compound statement.
type CompoundOperator uint

// Compound operators.
const (
	OperatorNone CompoundOperator = iota
	OperatorAnd
	OperatorOr
)

type compound struct {
	prev *compound
	fn   func(*[]Compound) error
}

func newCompound(conds ...Compound) *compound {
	c := &compound{}
	if len(conds) == 0 {
		return c
	}
	return c.frame(func(in *[]Compound) error {
		*in = append(*in, conds...)
		return nil
	})
}

// Sentences returns each one of the conditions as a compound.
func (c *compound) Sentences() []Compound {
	conds, err := immutable.FastForward(c)
	if err == nil {
		return *(conds.(*[]Compound))
	}
	return nil
}

// Operator returns no operator.
func (c *compound) Operator() CompoundOperator {
	return OperatorNone
}

// Empty returns true if this condition has no elements. False otherwise.
func (c *compound) Empty() bool {
	if c.fn != nil {
		return false
	}
	if c.prev != nil {
		return c.prev.Empty()
	}
	return true
}

func (c *compound) frame(fn func(*[]Compound) error) *compound {
	return &compound{prev: c, fn: fn}
}

// Prev is for internal usage.
func (c *compound) Prev() immutable.Immutable {
	if c == nil {
		return nil
	}
	return c.prev
}

// Fn is for internal usage.
func (c *compound) Fn(in interface{}) error {
	if c.fn == nil {
		return nil
	}
	return c.fn(in.(*[]Compound))
}

// Base is for internal usage.
func (c *compound) Base() interface{} {
	return &[]Compound{}
}

func defaultJoin(in ...Compound) []Compound {
	for i := range in {
		if cond, ok := in[i].(Cond); ok && len(cond) > 1 {
			in[i] = And(cond)
		}
	}
	return in
}

var (
	_ = immutable.Immutable(&compound{})
	_ = Compound(Cond{})
)
