package exql

import (
	"fmt"
	"strings"
)

// ValueGroups represents an array of value groups.
type ValueGroups struct {
	Values []*Values
	hash   hash
}

func (vg *ValueGroups) IsEmpty() bool {
	if vg == nil || len(vg.Values) < 1 {
		return true
	}
	for i := range vg.Values {
		if !vg.Values[i].IsEmpty() {
			return false
		}
	}
	return true
}

var _ = Fragment(&ValueGroups{})

// Values represents an array of Value.
type Values struct {
	Values []Fragment
	hash   hash
}

func (vs *Values) IsEmpty() bool {
	if vs == nil || len(vs.Values) < 1 {
		return true
	}
	return false
}

var _ = Fragment(&Values{})

// Value represents an escaped SQL value.
type Value struct {
	V    interface{}
	hash hash
}

var _ = Fragment(&Value{})

// NewValue creates and returns a Value.
func NewValue(v interface{}) *Value {
	return &Value{V: v}
}

// NewValueGroup creates and returns an array of values.
func NewValueGroup(v ...Fragment) *Values {
	return &Values{Values: v}
}

// Hash returns a unique identifier for the struct.
func (v *Value) Hash() string {
	return v.hash.Hash(v)
}

func (v *Value) IsEmpty() bool {
	return false
}

// Compile transforms the Value into an equivalent SQL representation.
func (v *Value) Compile(layout *Template) (compiled string, err error) {

	if z, ok := layout.Read(v); ok {
		return z, nil
	}

	switch t := v.V.(type) {
	case Raw:
		compiled, err = t.Compile(layout)
		if err != nil {
			return "", err
		}
	case Fragment:
		compiled, err = t.Compile(layout)
		if err != nil {
			return "", err
		}
	default:
		compiled = layout.MustCompile(layout.ValueQuote, RawValue(fmt.Sprintf(`%v`, v.V)))
	}

	layout.Write(v, compiled)

	return
}

// Hash returns a unique identifier for the struct.
func (vs *Values) Hash() string {
	return vs.hash.Hash(vs)
}

// Compile transforms the Values into an equivalent SQL representation.
func (vs *Values) Compile(layout *Template) (compiled string, err error) {
	if c, ok := layout.Read(vs); ok {
		return c, nil
	}

	l := len(vs.Values)
	if l > 0 {
		chunks := make([]string, 0, l)
		for i := 0; i < l; i++ {
			chunk, err := vs.Values[i].Compile(layout)
			if err != nil {
				return "", err
			}
			chunks = append(chunks, chunk)
		}
		compiled = layout.MustCompile(layout.ClauseGroup, strings.Join(chunks, layout.ValueSeparator))
	}
	layout.Write(vs, compiled)
	return
}

// Hash returns a unique identifier for the struct.
func (vg *ValueGroups) Hash() string {
	return vg.hash.Hash(vg)
}

// Compile transforms the ValueGroups into an equivalent SQL representation.
func (vg *ValueGroups) Compile(layout *Template) (compiled string, err error) {
	if c, ok := layout.Read(vg); ok {
		return c, nil
	}

	l := len(vg.Values)
	if l > 0 {
		chunks := make([]string, 0, l)
		for i := 0; i < l; i++ {
			chunk, err := vg.Values[i].Compile(layout)
			if err != nil {
				return "", err
			}
			chunks = append(chunks, chunk)
		}
		compiled = strings.Join(chunks, layout.ValueSeparator)
	}

	layout.Write(vg, compiled)
	return
}

// JoinValueGroups creates a new *ValueGroups object.
func JoinValueGroups(values ...*Values) *ValueGroups {
	return &ValueGroups{Values: values}
}
