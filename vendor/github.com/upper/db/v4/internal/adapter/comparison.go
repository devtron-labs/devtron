package adapter

// ComparisonOperator is the base type for comparison operators.
type ComparisonOperator uint8

// Comparison operators
const (
	ComparisonOperatorNone ComparisonOperator = iota
	ComparisonOperatorCustom

	ComparisonOperatorEqual
	ComparisonOperatorNotEqual

	ComparisonOperatorLessThan
	ComparisonOperatorGreaterThan

	ComparisonOperatorLessThanOrEqualTo
	ComparisonOperatorGreaterThanOrEqualTo

	ComparisonOperatorBetween
	ComparisonOperatorNotBetween

	ComparisonOperatorIn
	ComparisonOperatorNotIn

	ComparisonOperatorIs
	ComparisonOperatorIsNot

	ComparisonOperatorLike
	ComparisonOperatorNotLike

	ComparisonOperatorRegExp
	ComparisonOperatorNotRegExp
)

type Comparison struct {
	t  ComparisonOperator
	op string
	v  interface{}
}

func (c *Comparison) CustomOperator() string {
	return c.op
}

func (c *Comparison) Operator() ComparisonOperator {
	return c.t
}

func (c *Comparison) Value() interface{} {
	return c.v
}

func NewComparisonOperator(t ComparisonOperator, v interface{}) *Comparison {
	return &Comparison{t: t, v: v}
}

func NewCustomComparisonOperator(op string, v interface{}) *Comparison {
	return &Comparison{t: ComparisonOperatorCustom, op: op, v: v}
}
