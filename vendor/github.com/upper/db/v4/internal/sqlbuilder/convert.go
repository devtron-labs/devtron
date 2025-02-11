package sqlbuilder

import (
	"bytes"
	"database/sql/driver"
	"reflect"

	"github.com/upper/db/v4/internal/adapter"
	"github.com/upper/db/v4/internal/sqladapter/exql"
)

var (
	sqlDefault = &exql.Raw{Value: "DEFAULT"}
)

func expandQuery(in []byte, inArgs []interface{}) ([]byte, []interface{}) {
	out := make([]byte, 0, len(in))
	outArgs := make([]interface{}, 0, len(inArgs))

	i := 0
	for i < len(in) && len(inArgs) > 0 {
		if in[i] == '?' {
			out = append(out, in[:i]...)
			in = in[i+1:]
			i = 0

			replace, replaceArgs := expandArgument(inArgs[0])
			inArgs = inArgs[1:]

			if len(replace) > 0 {
				replace, replaceArgs = expandQuery(replace, replaceArgs)
				out = append(out, replace...)
			} else {
				out = append(out, '?')
			}

			outArgs = append(outArgs, replaceArgs...)
			continue
		}
		i = i + 1
	}

	if len(out) < 1 {
		return in, inArgs
	}

	out = append(out, in[:len(in)]...)
	in = nil

	outArgs = append(outArgs, inArgs[:len(inArgs)]...)
	inArgs = nil

	return out, outArgs
}

func expandArgument(arg interface{}) ([]byte, []interface{}) {
	values, isSlice := toInterfaceArguments(arg)

	if isSlice {
		if len(values) == 0 {
			return []byte("(NULL)"), nil
		}
		buf := bytes.Repeat([]byte(" ?,"), len(values))
		buf[0] = '('
		buf[len(buf)-1] = ')'
		return buf, values
	}

	if len(values) == 1 {
		switch t := arg.(type) {
		case *adapter.RawExpr:
			return expandQuery([]byte(t.Raw()), t.Arguments())
		case hasPaginator:
			p, err := t.Paginator()
			if err == nil {
				return append([]byte{'('}, append([]byte(p.String()), ')')...), p.Arguments()
			}
			panic(err.Error())
		case isCompilable:
			s, err := t.Compile()
			if err == nil {
				return append([]byte{'('}, append([]byte(s), ')')...), t.Arguments()
			}
			panic(err.Error())
		}
	} else if len(values) == 0 {
		return []byte("NULL"), nil
	}

	return nil, []interface{}{arg}
}

// toInterfaceArguments converts the given value into an array of interfaces.
func toInterfaceArguments(value interface{}) (args []interface{}, isSlice bool) {
	if value == nil {
		return nil, false
	}

	switch t := value.(type) {
	case driver.Valuer:
		return []interface{}{t}, false
	}

	v := reflect.ValueOf(value)
	if v.Type().Kind() == reflect.Slice {
		var i, total int

		// Byte slice gets transformed into a string.
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return []interface{}{string(v.Bytes())}, false
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

// Preprocess expands arguments that needs to be expanded and compiles a query
// into a single string.
func Preprocess(in string, args []interface{}) (string, []interface{}) {
	b, args := expandQuery([]byte(in), args)
	return string(b), args
}
