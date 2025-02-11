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

// ConstraintValuer allows constraints to use specific values of their own.
type ConstraintValuer interface {
	ConstraintValue() interface{}
}

// Constraint interface represents a single condition, like "a = 1".  where `a`
// is the key and `1` is the value. This is an exported interface but it's
// rarely used directly, you may want to use the `db.Cond{}` map instead.
type Constraint interface {
	// Key is the leftmost part of the constraint and usually contains a column
	// name.
	Key() interface{}

	// Value if the rightmost part of the constraint and usually contains a
	// column value.
	Value() interface{}
}

// Constraints interface represents an array of constraints, like "a = 1, b =
// 2, c = 3".
type Constraints interface {
	// Constraints returns an array of constraints.
	Constraints() []Constraint
}

type constraint struct {
	k interface{}
	v interface{}
}

func (c constraint) Key() interface{} {
	return c.k
}

func (c constraint) Value() interface{} {
	if constraintValuer, ok := c.v.(ConstraintValuer); ok {
		return constraintValuer.ConstraintValue()
	}
	return c.v
}

// NewConstraint creates a constraint.
func NewConstraint(key interface{}, value interface{}) Constraint {
	return &constraint{k: key, v: value}
}

var (
	_ = Constraint(&constraint{})
)
