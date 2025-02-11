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

// RawExpr interface represents values that can bypass SQL filters. This is an
// exported interface but it's rarely used directly, you may want to use the
// `db.Raw()` function instead.
type RawExpr struct {
	value string
	args  *[]interface{}
}

func (r *RawExpr) Arguments() []interface{} {
	if r.args != nil {
		return *r.args
	}
	return nil
}

func (r RawExpr) Raw() string {
	return r.value
}

func (r RawExpr) String() string {
	return r.Raw()
}

// Expressions returns a logical expressio.n
func (r *RawExpr) Expressions() []LogicalExpr {
	return []LogicalExpr{r}
}

// Operator returns the default compound operator.
func (r RawExpr) Operator() LogicalOperator {
	return LogicalOperatorNone
}

// Empty return false if this struct has no value.
func (r *RawExpr) Empty() bool {
	return r.value == ""
}

func NewRawExpr(value string, args []interface{}) *RawExpr {
	r := &RawExpr{value: value, args: nil}
	if len(args) > 0 {
		r.args = &args
	}
	return r
}

var _ = LogicalExpr(&RawExpr{})
