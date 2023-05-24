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

// Union represents a compound joined by OR.
type Union struct {
	*compound
}

// Or adds more terms to the compound.
func (o *Union) Or(orConds ...Compound) *Union {
	var fn func(*[]Compound) error
	if len(orConds) > 0 {
		fn = func(in *[]Compound) error {
			*in = append(*in, orConds...)
			return nil
		}
	}
	return &Union{o.compound.frame(fn)}
}

// Operator returns the OR operator.
func (o *Union) Operator() CompoundOperator {
	return OperatorOr
}

// Empty returns false if this struct holds no conditions.
func (o *Union) Empty() bool {
	return o.compound.Empty()
}

// Or joins conditions under logical disjunction. Conditions can be represented
// by db.Cond{}, db.Or() or db.And().
//
// Example:
//
//	// year = 2012 OR year = 1987
//	db.Or(
//		db.Cond{"year": 2012},
//		db.Cond{"year": 1987},
//	)
func Or(conds ...Compound) *Union {
	return &Union{newCompound(defaultJoin(conds...)...)}
}
