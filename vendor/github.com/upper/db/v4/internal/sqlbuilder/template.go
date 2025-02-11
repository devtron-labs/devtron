package sqlbuilder

import (
	"database/sql/driver"
	"fmt"
	"strings"

	db "github.com/upper/db/v4"
	"github.com/upper/db/v4/internal/adapter"
	"github.com/upper/db/v4/internal/sqladapter/exql"
)

type templateWithUtils struct {
	*exql.Template
}

func newTemplateWithUtils(template *exql.Template) *templateWithUtils {
	return &templateWithUtils{template}
}

func (tu *templateWithUtils) PlaceholderValue(in interface{}) (exql.Fragment, []interface{}) {
	switch t := in.(type) {
	case *adapter.RawExpr:
		return &exql.Raw{Value: t.Raw()}, t.Arguments()
	case *adapter.FuncExpr:
		fnName := t.Name()
		fnArgs := []interface{}{}
		args, _ := toInterfaceArguments(t.Arguments())
		fragments := []string{}
		for i := range args {
			frag, args := tu.PlaceholderValue(args[i])
			fragment, err := frag.Compile(tu.Template)
			if err == nil {
				fragments = append(fragments, fragment)
				fnArgs = append(fnArgs, args...)
			}
		}
		return &exql.Raw{Value: fnName + `(` + strings.Join(fragments, `, `) + `)`}, fnArgs
	default:
		return sqlPlaceholder, []interface{}{in}
	}
}

// toWhereWithArguments converts the given parameters into a exql.Where value.
func (tu *templateWithUtils) toWhereWithArguments(term interface{}) (where exql.Where, args []interface{}) {
	args = []interface{}{}

	switch t := term.(type) {
	case []interface{}:
		if len(t) > 0 {
			if s, ok := t[0].(string); ok {
				if strings.ContainsAny(s, "?") || len(t) == 1 {
					s, args = Preprocess(s, t[1:])
					where.Conditions = []exql.Fragment{&exql.Raw{Value: s}}
				} else {
					var val interface{}
					key := s
					if len(t) > 2 {
						val = t[1:]
					} else {
						val = t[1]
					}
					cv, v := tu.toColumnValues(adapter.NewConstraint(key, val))
					args = append(args, v...)
					where.Conditions = append(where.Conditions, cv.ColumnValues...)
				}
				return
			}
		}
		for i := range t {
			w, v := tu.toWhereWithArguments(t[i])
			if len(w.Conditions) == 0 {
				continue
			}
			args = append(args, v...)
			where.Conditions = append(where.Conditions, w.Conditions...)
		}
		return
	case *adapter.RawExpr:
		r, v := Preprocess(t.Raw(), t.Arguments())
		where.Conditions = []exql.Fragment{&exql.Raw{Value: r}}
		args = append(args, v...)
		return
	case adapter.Constraints:
		for _, c := range t.Constraints() {
			w, v := tu.toWhereWithArguments(c)
			if len(w.Conditions) == 0 {
				continue
			}
			args = append(args, v...)
			where.Conditions = append(where.Conditions, w.Conditions...)
		}
		return
	case adapter.LogicalExpr:
		var cond exql.Where

		expressions := t.Expressions()
		for i := range expressions {
			w, v := tu.toWhereWithArguments(expressions[i])
			if len(w.Conditions) == 0 {
				continue
			}
			args = append(args, v...)
			cond.Conditions = append(cond.Conditions, w.Conditions...)
		}
		if len(cond.Conditions) < 1 {
			return
		}

		if len(cond.Conditions) <= 1 {
			where.Conditions = append(where.Conditions, cond.Conditions...)
			return where, args
		}

		var frag exql.Fragment
		switch t.Operator() {
		case adapter.LogicalOperatorNone, adapter.LogicalOperatorAnd:
			q := exql.And(cond)
			frag = &q
		case adapter.LogicalOperatorOr:
			q := exql.Or(cond)
			frag = &q
		default:
			panic(fmt.Sprintf("Unknown type %T", t))
		}
		where.Conditions = append(where.Conditions, frag)
		return

	case db.InsertResult:
		return tu.toWhereWithArguments(t.ID())

	case adapter.Constraint:
		cv, v := tu.toColumnValues(t)
		args = append(args, v...)
		where.Conditions = append(where.Conditions, cv.ColumnValues...)
		return where, args
	}

	panic(fmt.Sprintf("Unknown condition type %T", term))
}

func (tu *templateWithUtils) comparisonOperatorMapper(t adapter.ComparisonOperator) string {
	if t == adapter.ComparisonOperatorCustom {
		return ""
	}
	if tu.ComparisonOperator != nil {
		if op, ok := tu.ComparisonOperator[t]; ok {
			return op
		}
	}
	if op, ok := comparisonOperators[t]; ok {
		return op
	}
	panic(fmt.Sprintf("unsupported comparison operator %v", t))
}

func (tu *templateWithUtils) toColumnValues(term interface{}) (cv exql.ColumnValues, args []interface{}) {
	args = []interface{}{}

	switch t := term.(type) {
	case adapter.Constraint:
		columnValue := exql.ColumnValue{}

		// TODO: Key and Value are similar. Can we refactor this? Maybe think about
		// Left/Right rather than Key/Value.

		switch key := t.Key().(type) {
		case string:
			chunks := strings.SplitN(strings.TrimSpace(key), " ", 2)
			columnValue.Column = exql.ColumnWithName(chunks[0])
			if len(chunks) > 1 {
				columnValue.Operator = chunks[1]
			}
		case *adapter.RawExpr:
			columnValue.Column = &exql.Raw{Value: key.Raw()}
			args = append(args, key.Arguments()...)
		case *db.FuncExpr:
			fnName, fnArgs := key.Name(), key.Arguments()
			if len(fnArgs) == 0 {
				fnName = fnName + "()"
			} else {
				fnName = fnName + "(?" + strings.Repeat(", ?", len(fnArgs)-1) + ")"
			}
			fnName, fnArgs = Preprocess(fnName, fnArgs)
			columnValue.Column = &exql.Raw{Value: fnName}
			args = append(args, fnArgs...)
		default:
			columnValue.Column = &exql.Raw{Value: fmt.Sprintf("%v", key)}
		}

		switch value := t.Value().(type) {
		case *db.FuncExpr:
			fnName, fnArgs := value.Name(), value.Arguments()
			if len(fnArgs) == 0 {
				// A function with no arguments.
				fnName = fnName + "()"
			} else {
				// A function with one or more arguments.
				fnName = fnName + "(?" + strings.Repeat(", ?", len(fnArgs)-1) + ")"
			}
			fnName, fnArgs = Preprocess(fnName, fnArgs)
			columnValue.Value = &exql.Raw{Value: fnName}
			args = append(args, fnArgs...)
		case *db.RawExpr:
			q, a := Preprocess(value.Raw(), value.Arguments())
			columnValue.Value = &exql.Raw{Value: q}
			args = append(args, a...)
		case driver.Valuer:
			columnValue.Value = sqlPlaceholder
			args = append(args, value)
		case *db.Comparison:
			wrapper := &operatorWrapper{
				tu: tu,
				cv: &columnValue,
				op: value.Comparison,
			}

			q, a := wrapper.preprocess()
			q, a = Preprocess(q, a)

			columnValue = exql.ColumnValue{
				Column: &exql.Raw{Value: q},
			}
			if a != nil {
				args = append(args, a...)
			}

			cv.ColumnValues = append(cv.ColumnValues, &columnValue)
			return cv, args
		default:
			wrapper := &operatorWrapper{
				tu: tu,
				cv: &columnValue,
				v:  value,
			}

			q, a := wrapper.preprocess()
			q, a = Preprocess(q, a)

			columnValue = exql.ColumnValue{
				Column: &exql.Raw{Value: q},
			}
			if a != nil {
				args = append(args, a...)
			}

			cv.ColumnValues = append(cv.ColumnValues, &columnValue)
			return cv, args
		}

		if columnValue.Operator == "" {
			columnValue.Operator = tu.comparisonOperatorMapper(adapter.ComparisonOperatorEqual)
		}

		cv.ColumnValues = append(cv.ColumnValues, &columnValue)
		return cv, args

	case *adapter.RawExpr:
		columnValue := exql.ColumnValue{}
		p, q := Preprocess(t.Raw(), t.Arguments())
		columnValue.Column = &exql.Raw{Value: p}
		cv.ColumnValues = append(cv.ColumnValues, &columnValue)
		args = append(args, q...)
		return cv, args

	case adapter.Constraints:
		for _, constraint := range t.Constraints() {
			p, q := tu.toColumnValues(constraint)
			cv.ColumnValues = append(cv.ColumnValues, p.ColumnValues...)
			args = append(args, q...)
		}
		return cv, args
	}

	panic(fmt.Sprintf("Unknown term type %T.", term))
}

func (tu *templateWithUtils) setColumnValues(term interface{}) (cv exql.ColumnValues, args []interface{}) {
	args = []interface{}{}

	switch t := term.(type) {
	case []interface{}:
		l := len(t)
		for i := 0; i < l; i++ {
			column, isString := t[i].(string)

			if !isString {
				p, q := tu.setColumnValues(t[i])
				cv.ColumnValues = append(cv.ColumnValues, p.ColumnValues...)
				args = append(args, q...)
				continue
			}

			if !strings.ContainsAny(column, tu.AssignmentOperator) {
				column = column + " " + tu.AssignmentOperator + " ?"
			}

			chunks := strings.SplitN(column, tu.AssignmentOperator, 2)

			column = chunks[0]
			format := strings.TrimSpace(chunks[1])

			columnValue := exql.ColumnValue{
				Column:   exql.ColumnWithName(column),
				Operator: tu.AssignmentOperator,
				Value:    &exql.Raw{Value: format},
			}

			ps := strings.Count(format, "?")
			if i+ps < l {
				for j := 0; j < ps; j++ {
					args = append(args, t[i+j+1])
				}
				i = i + ps
			} else {
				panic(fmt.Sprintf("Format string %q has more placeholders than given arguments.", format))
			}

			cv.ColumnValues = append(cv.ColumnValues, &columnValue)
		}
		return cv, args
	case *adapter.RawExpr:
		columnValue := exql.ColumnValue{}
		p, q := Preprocess(t.Raw(), t.Arguments())
		columnValue.Column = &exql.Raw{Value: p}
		cv.ColumnValues = append(cv.ColumnValues, &columnValue)
		args = append(args, q...)
		return cv, args
	}

	panic(fmt.Sprintf("Unknown term type %T.", term))
}
