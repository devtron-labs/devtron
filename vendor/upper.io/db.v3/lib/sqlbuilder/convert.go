package sqlbuilder

import (
	"database/sql/driver"
	"reflect"
	"strings"

	db "upper.io/db.v3"
	"upper.io/db.v3/internal/sqladapter/exql"
)

var (
	sqlDefault = exql.RawValue(`DEFAULT`)
)

func expandQuery(in string, args []interface{}, fn func(interface{}) (string, []interface{})) (string, []interface{}) {
	argn := 0
	argx := make([]interface{}, 0, len(args))
	for i := 0; i < len(in); i++ {
		if in[i] != '?' {
			continue
		}
		if len(args) > argn {
			k, values := fn(args[argn])
			k, values = expandQuery(k, values, fn)

			if k != "" {
				in = in[:i] + k + in[i+1:]
				i += len(k) - 1
			}
			if len(values) > 0 {
				argx = append(argx, values...)
			}
			argn++
		}
	}
	if len(argx) < len(args) {
		argx = append(argx, args[argn:]...)
	}
	return in, argx
}

// toInterfaceArguments converts the given value into an array of interfaces.
func toInterfaceArguments(value interface{}) (args []interface{}, isSlice bool) {
	v := reflect.ValueOf(value)

	if value == nil {
		return nil, false
	}

	switch t := value.(type) {
	case driver.Valuer:
		return []interface{}{t}, false
	}

	if v.Type().Kind() == reflect.Slice {
		var i, total int

		// Byte slice gets transformed into a string.
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return []interface{}{string(value.([]byte))}, false
		}

		total = v.Len()
		args = make([]interface{}, total)
		for i = 0; i < total; i++ {
			args[i] = v.Index(i).Interface()
		}
		return args, true
	}

	return []interface{}{value}, false
}

// toColumnsValuesAndArguments maps the given columnNames and columnValues into
// expr's Columns and Values, it also extracts and returns query arguments.
func toColumnsValuesAndArguments(columnNames []string, columnValues []interface{}) (*exql.Columns, *exql.Values, []interface{}, error) {
	var arguments []interface{}

	columns := new(exql.Columns)

	columns.Columns = make([]exql.Fragment, 0, len(columnNames))
	for i := range columnNames {
		columns.Columns = append(columns.Columns, exql.ColumnWithName(columnNames[i]))
	}

	values := new(exql.Values)

	arguments = make([]interface{}, 0, len(columnValues))
	values.Values = make([]exql.Fragment, 0, len(columnValues))

	for i := range columnValues {
		switch v := columnValues[i].(type) {
		case *exql.Raw, exql.Raw:
			values.Values = append(values.Values, sqlDefault)
		case *exql.Value:
			// Adding value.
			values.Values = append(values.Values, v)
		case exql.Value:
			// Adding value.
			values.Values = append(values.Values, &v)
		default:
			// Adding both value and placeholder.
			values.Values = append(values.Values, sqlPlaceholder)
			arguments = append(arguments, v)
		}
	}

	return columns, values, arguments, nil
}

func preprocessFn(arg interface{}) (string, []interface{}) {
	values, isSlice := toInterfaceArguments(arg)

	if isSlice {
		if len(values) == 0 {
			return `(NULL)`, nil
		}
		return `(?` + strings.Repeat(`, ?`, len(values)-1) + `)`, values
	}

	if len(values) == 1 {
		switch t := arg.(type) {
		case db.RawValue:
			return Preprocess(t.Raw(), t.Arguments())
		case compilable:
			c, err := t.Compile()
			if err == nil {
				return `(` + c + `)`, t.Arguments()
			}
			panic(err.Error())
		}
	} else if len(values) == 0 {
		return `NULL`, nil
	}

	return "", []interface{}{arg}
}

// Preprocess expands arguments that needs to be expanded and compiles a query
// into a single string.
func Preprocess(in string, args []interface{}) (string, []interface{}) {
	return expandQuery(in, args, preprocessFn)
}
