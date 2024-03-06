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

// Intersection represents a compound joined by AND.
type Intersection struct {
	*compound
}

// And adds more terms to the compound.
func (a *Intersection) And(andConds ...Compound) *Intersection {
	var fn func(*[]Compound) error
	if len(andConds) > 0 {
		fn = func(in *[]Compound) error {
			*in = append(*in, andConds...)
			return nil
		}
	}
	return &Intersection{a.compound.frame(fn)}
}

// Empty returns false if this struct holds no conditions.
func (a *Intersection) Empty() bool {
	return a.compound.Empty()
}

// Operator returns the AND operator.
func (a *Intersection) Operator() CompoundOperator {
	return OperatorAnd
}

// And joins conditions under logical conjunction. Conditions can be
// represented by db.Cond{}, db.Or() or db.And().
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
func And(conds ...Compound) *Intersection {
	return &Intersection{newCompound(conds...)}
}
