package exql

import (
	"fmt"

	"github.com/upper/db/v4/internal/cache"
)

var (
	_ = fmt.Stringer(&Raw{})
)

// Raw represents a value that is meant to be used in a query without escaping.
type Raw struct {
	Value string
}

func NewRawValue(v interface{}) (*Raw, error) {
	switch t := v.(type) {
	case string:
		return &Raw{Value: t}, nil
	case int, uint, int64, uint64, int32, uint32, int16, uint16:
		return &Raw{Value: fmt.Sprintf("%d", t)}, nil
	case fmt.Stringer:
		return &Raw{Value: t.String()}, nil
	}
	return nil, fmt.Errorf("unexpected type: %T", v)
}

// Hash returns a unique identifier for the struct.
func (r *Raw) Hash() uint64 {
	if r == nil {
		return cache.NewHash(FragmentType_Raw, nil)
	}
	return cache.NewHash(FragmentType_Raw, r.Value)
}

// Compile returns the raw value.
func (r *Raw) Compile(*Template) (string, error) {
	return r.Value, nil
}

// String returns the raw value.
func (r *Raw) String() string {
	return r.Value
}

var _ = Fragment(&Raw{})
