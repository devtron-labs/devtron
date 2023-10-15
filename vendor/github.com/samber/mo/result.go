package mo

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Ok builds a Result when value is valid.
// Play: https://go.dev/play/p/PDwADdzNoyZ
func Ok[T any](value T) Result[T] {
	return Result[T]{
		value: value,
		isErr: false,
	}
}

// Err builds a Result when value is invalid.
// Play: https://go.dev/play/p/PDwADdzNoyZ
func Err[T any](err error) Result[T] {
	return Result[T]{
		err:   err,
		isErr: true,
	}
}

// Errf builds a Result when value is invalid.
// Errf formats according to a format specifier and returns the error as a value that satisfies Result[T].
// Play: https://go.dev/play/p/N43w92SM-Bs
func Errf[T any](format string, a ...any) Result[T] {
	return Err[T](fmt.Errorf(format, a...))
}

// TupleToResult convert a pair of T and error into a Result.
// Play: https://go.dev/play/p/KWjfqQDHQwa
func TupleToResult[T any](value T, err error) Result[T] {
	if err != nil {
		return Err[T](err)
	}
	return Ok(value)
}

// Try returns either a Ok or Err object.
// Play: https://go.dev/play/p/ilOlQx-Mx42
func Try[T any](f func() (T, error)) Result[T] {
	return TupleToResult(f())
}

// Result represents a result of an action having one
// of the following output: success or failure.
// An instance of Result is an instance of either Ok or Err.
// It could be compared to `Either[error, T]`.
type Result[T any] struct {
	isErr bool
	value T
	err   error
}

// IsOk returns true when value is valid.
// Play: https://go.dev/play/p/sfNvBQyZfgU
func (r Result[T]) IsOk() bool {
	return !r.isErr
}

// IsError returns true when value is invalid.
// Play: https://go.dev/play/p/xkV9d464scV
func (r Result[T]) IsError() bool {
	return r.isErr
}

// Error returns error when value is invalid or nil.
// Play: https://go.dev/play/p/CSkHGTyiXJ5
func (r Result[T]) Error() error {
	return r.err
}

// Get returns value and error.
// Play: https://go.dev/play/p/8KyX3z6TuNo
func (r Result[T]) Get() (T, error) {
	if r.isErr {
		return empty[T](), r.err
	}

	return r.value, nil
}

// MustGet returns value when Result is valid or panics.
// Play: https://go.dev/play/p/8LSlndHoTAE
func (r Result[T]) MustGet() T {
	if r.isErr {
		panic(r.err)
	}

	return r.value
}

// OrElse returns value when Result is valid or default value.
// Play: https://go.dev/play/p/MN_ULx0soi6
func (r Result[T]) OrElse(fallback T) T {
	if r.isErr {
		return fallback
	}

	return r.value
}

// OrEmpty returns value when Result is valid or empty value.
// Play: https://go.dev/play/p/rdKtBmOcMLh
func (r Result[T]) OrEmpty() T {
	return r.value
}

// ToEither transforms a Result into an Either type.
// Play: https://go.dev/play/p/Uw1Zz6b952q
func (r Result[T]) ToEither() Either[error, T] {
	if r.isErr {
		return Left[error, T](r.err)
	}

	return Right[error, T](r.value)
}

// ForEach executes the given side-effecting function if Result is valid.
func (r Result[T]) ForEach(mapper func(value T)) {
	if !r.isErr {
		mapper(r.value)
	}
}

// Match executes the first function if Result is valid and second function if invalid.
// It returns a new Result.
// Play: https://go.dev/play/p/-_eFaLJ31co
func (r Result[T]) Match(onSuccess func(value T) (T, error), onError func(err error) (T, error)) Result[T] {
	if r.isErr {
		return TupleToResult(onError(r.err))
	}
	return TupleToResult(onSuccess(r.value))
}

// Map executes the mapper function if Result is valid. It returns a new Result.
// Play: https://go.dev/play/p/-ndpN_b_OSc
func (r Result[T]) Map(mapper func(value T) (T, error)) Result[T] {
	if !r.isErr {
		return TupleToResult(mapper(r.value))
	}

	return Err[T](r.err)
}

// MapErr executes the mapper function if Result is invalid. It returns a new Result.
// Play: https://go.dev/play/p/WraZixg9GGf
func (r Result[T]) MapErr(mapper func(error) (T, error)) Result[T] {
	if r.isErr {
		return TupleToResult(mapper(r.err))
	}

	return Ok(r.value)
}

// FlatMap executes the mapper function if Result is valid. It returns a new Result.
// Play: https://go.dev/play/p/Ud5QjZOqg-7
func (r Result[T]) FlatMap(mapper func(value T) Result[T]) Result[T] {
	if !r.isErr {
		return mapper(r.value)
	}

	return Err[T](r.err)
}

// MarshalJSON encodes Result into json, following the JSON-RPC specification for results,
// with one exception: when the result is an error, the "code" field is not included.
// Reference: https://www.jsonrpc.org/specification
func (o Result[T]) MarshalJSON() ([]byte, error) {
	if o.isErr {
		return json.Marshal(map[string]any{
			"error": map[string]any{
				"message": o.err.Error(),
			},
		})
	}

	return json.Marshal(map[string]any{
		"result": o.value,
	})
}

// UnmarshalJSON decodes json into Result. If "error" is set, the result is an
// Err containing the error message as a generic error object. Otherwise, the
// result is an Ok containing the result. If the JSON object contains netiher
// an error nor a result, the result is an Ok containing an empty value. If the
// JSON object contains both an error and a result, the result is an Err. Finally,
// if the JSON object contains an error but is not structured correctly (no message
// field), the unmarshaling fails.
func (o *Result[T]) UnmarshalJSON(data []byte) error {
	var result struct {
		Result T `json:"result"`
		Error  struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}

	if result.Error.Message != "" {
		o.err = errors.New(result.Error.Message)
		o.isErr = true
		return nil
	}

	o.value = result.Result
	o.isErr = false
	return nil
}
