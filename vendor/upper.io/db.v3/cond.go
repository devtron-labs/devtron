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
)

// Cond is a map that defines conditions for a query and satisfies the
// Constraints and Compound interfaces.
//
// Each entry of the map represents a condition (a column-value relation bound
// by a comparison operator). The comparison operator is optional and can be
// specified after the column name, if no comparison operator is provided the
// equality operator is used as default.
//
// Examples:
//
//  // Where age equals 18.
//  db.Cond{"age": 18}
//  //	// Where age is greater than or equal to 18.
//  db.Cond{"age >=": 18}
//
//  // Where id is in a list of ids.
//  db.Cond{"id IN": []{1, 2, 3}}
//
//  // Where age is lower than 18 (you could use this syntax when using
//  // mongodb).
//  db.Cond{"age $lt": 18}
//
//  // Where age > 32 and age < 35
//  db.Cond{"age >": 32, "age <": 35}
type Cond map[interface{}]interface{}

// Constraints returns each one of the Cond map records as a constraint.
func (c Cond) Constraints() []Constraint {
	z := make([]Constraint, 0, len(c))
	for _, k := range c.Keys() {
		z = append(z, NewConstraint(k, c[k]))
	}
	return z
}

// Keys returns the keys of this map sorted by name.
func (c Cond) Keys() []interface{} {
	keys := make(condKeys, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	if len(c) > 1 {
		sort.Sort(keys)
	}
	return keys
}

// Sentences return each one of the map records as a compound.
func (c Cond) Sentences() []Compound {
	z := make([]Compound, 0, len(c))
	for _, k := range c.Keys() {
		z = append(z, Cond{k: c[k]})
	}
	return z
}

// Operator returns the default compound operator.
func (c Cond) Operator() CompoundOperator {
	return OperatorNone
}

// Empty returns false if there are no conditions.
func (c Cond) Empty() bool {
	for range c {
		return false
	}
	return true
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
