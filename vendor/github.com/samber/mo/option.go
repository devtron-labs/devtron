package mo

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

var optionNoSuchElement = fmt.Errorf("no such element")

// Some builds an Option when value is present.
// Play: https://go.dev/play/p/iqz2n9n0tDM
func Some[T any](value T) Option[T] {
	return Option[T]{
		isPresent: true,
		value:     value,
	}
}

// None builds an Option when value is absent.
// Play: https://go.dev/play/p/yYQPsYCSYlD
func None[T any]() Option[T] {
	return Option[T]{
		isPresent: false,
	}
}

// TupleToOption builds a Some Option when second argument is true, or None.
// Play: https://go.dev/play/p/gkrg2pZwOty
func TupleToOption[T any](value T, ok bool) Option[T] {
	if ok {
		return Some(value)
	}
	return None[T]()
}

// EmptyableToOption builds a Some Option when value is not empty, or None.
// Play: https://go.dev/play/p/GSpQQ-q-UES
func EmptyableToOption[T any](value T) Option[T] {
	// 🤮
	isZero := reflect.ValueOf(&value).Elem().IsZero()
	if isZero {
		return None[T]()
	}

	return Some(value)
}

// PointerToOption builds a Some Option when value is not nil, or None.
// Play: https://go.dev/play/p/yPVMj4DUb-I
func PointerToOption[T any](value *T) Option[T] {
	if value == nil {
		return None[T]()
	}

	return Some(*value)
}

// Option is a container for an optional value of type T. If value exists, Option is
// of type Some. If the value is absent, Option is of type None.
type Option[T any] struct {
	isPresent bool
	value     T
}

// IsPresent returns false when value is absent.
// Play: https://go.dev/play/p/nDqIaiihyCA
func (o Option[T]) IsPresent() bool {
	return o.isPresent
}

// IsAbsent returns false when value is present.
// Play: https://go.dev/play/p/23e2zqyVOQm
func (o Option[T]) IsAbsent() bool {
	return !o.isPresent
}

// Size returns 1 when value is present or 0 instead.
// Play: https://go.dev/play/p/7ixCNG1E9l7
func (o Option[T]) Size() int {
	if o.isPresent {
		return 1
	}

	return 0
}

// Get returns value and presence.
// Play: https://go.dev/play/p/0-JBa1usZRT
func (o Option[T]) Get() (T, bool) {
	if !o.isPresent {
		return empty[T](), false
	}

	return o.value, true
}

// MustGet returns value if present or panics instead.
// Play: https://go.dev/play/p/RVBckjdi5WR
func (o Option[T]) MustGet() T {
	if !o.isPresent {
		panic(optionNoSuchElement)
	}

	return o.value
}

// OrElse returns value if present or default value.
// Play: https://go.dev/play/p/TrGByFWCzXS
func (o Option[T]) OrElse(fallback T) T {
	if !o.isPresent {
		return fallback
	}

	return o.value
}

// OrEmpty returns value if present or empty value.
// Play: https://go.dev/play/p/SpSUJcE-tQm
func (o Option[T]) OrEmpty() T {
	return o.value
}

// ForEach executes the given side-effecting function of value is present.
func (o Option[T]) ForEach(onValue func(value T)) {
	if o.isPresent {
		onValue(o.value)
	}
}

// Match executes the first function if value is present and second function if absent.
// It returns a new Option.
// Play: https://go.dev/play/p/1V6st3LDJsM
func (o Option[T]) Match(onValue func(value T) (T, bool), onNone func() (T, bool)) Option[T] {
	if o.isPresent {
		return TupleToOption(onValue(o.value))
	}
	return TupleToOption(onNone())
}

// Map executes the mapper function if value is present or returns None if absent.
// Play: https://go.dev/play/p/mvfP3pcP_eJ
func (o Option[T]) Map(mapper func(value T) (T, bool)) Option[T] {
	if o.isPresent {
		return TupleToOption(mapper(o.value))
	}

	return None[T]()
}

// MapNone executes the mapper function if value is absent or returns Option.
// Play: https://go.dev/play/p/_KaHWZ6Q17b
func (o Option[T]) MapNone(mapper func() (T, bool)) Option[T] {
	if o.isPresent {
		return Some(o.value)
	}

	return TupleToOption(mapper())
}

// FlatMap executes the mapper function if value is present or returns None if absent.
// Play: https://go.dev/play/p/OXO-zJx6n5r
func (o Option[T]) FlatMap(mapper func(value T) Option[T]) Option[T] {
	if o.isPresent {
		return mapper(o.value)
	}

	return None[T]()
}

// ToPointer returns value if present or a nil pointer.
// Play: https://go.dev/play/p/N43w92SM-Bs
func (o Option[T]) ToPointer() *T {
	if !o.isPresent {
		return nil
	}

	return &o.value
}

// MarshalJSON encodes Option into json.
func (o Option[T]) MarshalJSON() ([]byte, error) {
	if o.isPresent {
		return json.Marshal(o.value)
	}

	// if anybody find a way to support `omitempty` param, please contribute!
	return json.Marshal(nil)
}

// UnmarshalJSON decodes Option from json.
func (o *Option[T]) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		o.isPresent = false
		return nil
	}

	err := json.Unmarshal(b, &o.value)
	if err != nil {
		return err
	}

	o.isPresent = true
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface.
func (o Option[T]) MarshalText() ([]byte, error) {
	return json.Marshal(o)
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (o *Option[T]) UnmarshalText(data []byte) error {
	return json.Unmarshal(data, o)
}

// MarshalBinary is the interface implemented by an object that can marshal itself into a binary form.
func (o Option[T]) MarshalBinary() ([]byte, error) {
	if !o.isPresent {
		return []byte{0}, nil
	}

	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(o.value); err != nil {
		return []byte{}, err
	}

	return append([]byte{1}, buf.Bytes()...), nil
}

// UnmarshalBinary is the interface implemented by an object that can unmarshal a binary representation of itself.
func (o *Option[T]) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return errors.New("Option[T].UnmarshalBinary: no data")
	}

	if data[0] == 0 {
		o.isPresent = false
		o.value = empty[T]()
		return nil
	}

	buf := bytes.NewBuffer(data[1:])
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&o.value)
	if err != nil {
		return err
	}

	o.isPresent = true
	return nil
}

// GobEncode implements the gob.GobEncoder interface.
func (o Option[T]) GobEncode() ([]byte, error) {
	return o.MarshalBinary()
}

// GobDecode implements the gob.GobDecoder interface.
func (o *Option[T]) GobDecode(data []byte) error {
	return o.UnmarshalBinary(data)
}

// Scan implements the SQL sql.Scanner interface.
func (o *Option[T]) Scan(src any) error {
	if src == nil {
		o.isPresent = false
		o.value = empty[T]()
		return nil
	}

	// is is only possible to assert interfaces, so convert first
	// https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#why-not-permit-type-assertions-on-values-whose-type-is-a-type-parameter
	var t T
	if tScanner, ok := interface{}(&t).(sql.Scanner); ok {
		if err := tScanner.Scan(src); err != nil {
			return fmt.Errorf("failed to scan: %w", err)
		}

		o.isPresent = true
		o.value = t
		return nil
	}

	if av, err := driver.DefaultParameterConverter.ConvertValue(src); err == nil {
		if v, ok := av.(T); ok {
			o.isPresent = true
			o.value = v
			return nil
		}
	}

	return fmt.Errorf("failed to scan Option[T]")
}

// Value implements the driver Valuer interface.
func (o Option[T]) Value() (driver.Value, error) {
	if !o.isPresent {
		return nil, nil
	}

	return o.value, nil
}
