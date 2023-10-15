package mo

import "fmt"

const (
	either3ArgId1 = iota
	either3ArgId2
	either3ArgId3
)

var (
	either3InvalidArgumentId = fmt.Errorf("either3 argument should be between 1 and 3")
	either3MissingArg1       = fmt.Errorf("either3 doesn't contain expected argument 1")
	either3MissingArg2       = fmt.Errorf("either3 doesn't contain expected argument 2")
	either3MissingArg3       = fmt.Errorf("either3 doesn't contain expected argument 3")
)

// NewEither3Arg1 builds the first argument of the Either3 struct.
func NewEither3Arg1[T1 any, T2 any, T3 any](value T1) Either3[T1, T2, T3] {
	return Either3[T1, T2, T3]{
		argId: either3ArgId1,
		arg1:  value,
	}
}

// NewEither3Arg2 builds the second argument of the Either3 struct.
func NewEither3Arg2[T1 any, T2 any, T3 any](value T2) Either3[T1, T2, T3] {
	return Either3[T1, T2, T3]{
		argId: either3ArgId2,
		arg2:  value,
	}
}

// NewEither3Arg3 builds the third argument of the Either3 struct.
func NewEither3Arg3[T1 any, T2 any, T3 any](value T3) Either3[T1, T2, T3] {
	return Either3[T1, T2, T3]{
		argId: either3ArgId3,
		arg3:  value,
	}
}

// Either3 represents a value of 3 possible types.
// An instance of Either3 is an instance of either T1, T2 or T3.
type Either3[T1 any, T2 any, T3 any] struct {
	argId int8

	arg1 T1
	arg2 T2
	arg3 T3
}

// IsArg1 returns true if Either3 uses the first argument.
func (e Either3[T1, T2, T3]) IsArg1() bool {
	return e.argId == either3ArgId1
}

// IsArg2 returns true if Either3 uses the second argument.
func (e Either3[T1, T2, T3]) IsArg2() bool {
	return e.argId == either3ArgId2
}

// IsArg3 returns true if Either3 uses the third argument.
func (e Either3[T1, T2, T3]) IsArg3() bool {
	return e.argId == either3ArgId3
}

// Arg1 returns the first argument of a Either3 struct.
func (e Either3[T1, T2, T3]) Arg1() (T1, bool) {
	if e.IsArg1() {
		return e.arg1, true
	}
	return empty[T1](), false
}

// Arg2 returns the second argument of a Either3 struct.
func (e Either3[T1, T2, T3]) Arg2() (T2, bool) {
	if e.IsArg2() {
		return e.arg2, true
	}
	return empty[T2](), false
}

// Arg3 returns the third argument of a Either3 struct.
func (e Either3[T1, T2, T3]) Arg3() (T3, bool) {
	if e.IsArg3() {
		return e.arg3, true
	}
	return empty[T3](), false
}

// MustArg1 returns the first argument of a Either3 struct or panics.
func (e Either3[T1, T2, T3]) MustArg1() T1 {
	if !e.IsArg1() {
		panic(either3MissingArg1)
	}
	return e.arg1
}

// MustArg2 returns the second argument of a Either3 struct or panics.
func (e Either3[T1, T2, T3]) MustArg2() T2 {
	if !e.IsArg2() {
		panic(either3MissingArg2)
	}
	return e.arg2
}

// MustArg3 returns the third argument of a Either3 struct or panics.
func (e Either3[T1, T2, T3]) MustArg3() T3 {
	if !e.IsArg3() {
		panic(either3MissingArg3)
	}
	return e.arg3
}

// Unpack returns all values
func (e Either3[T1, T2, T3]) Unpack() (T1, T2, T3) {
	return e.arg1, e.arg2, e.arg3
}

// Arg1OrElse returns the first argument of a Either3 struct or fallback.
func (e Either3[T1, T2, T3]) Arg1OrElse(fallback T1) T1 {
	if e.IsArg1() {
		return e.arg1
	}
	return fallback
}

// Arg2OrElse returns the second argument of a Either3 struct or fallback.
func (e Either3[T1, T2, T3]) Arg2OrElse(fallback T2) T2 {
	if e.IsArg2() {
		return e.arg2
	}
	return fallback
}

// Arg3OrElse returns the third argument of a Either3 struct or fallback.
func (e Either3[T1, T2, T3]) Arg3OrElse(fallback T3) T3 {
	if e.IsArg3() {
		return e.arg3
	}
	return fallback
}

// Arg1OrEmpty returns the first argument of a Either3 struct or empty value.
func (e Either3[T1, T2, T3]) Arg1OrEmpty() T1 {
	if e.IsArg1() {
		return e.arg1
	}
	return empty[T1]()
}

// Arg2OrEmpty returns the second argument of a Either3 struct or empty value.
func (e Either3[T1, T2, T3]) Arg2OrEmpty() T2 {
	if e.IsArg2() {
		return e.arg2
	}
	return empty[T2]()
}

// Arg3OrEmpty returns the third argument of a Either3 struct or empty value.
func (e Either3[T1, T2, T3]) Arg3OrEmpty() T3 {
	if e.IsArg3() {
		return e.arg3
	}
	return empty[T3]()
}

// ForEach executes the given side-effecting function, depending of the argument set.
func (e Either3[T1, T2, T3]) ForEach(arg1Cb func(T1), arg2Cb func(T2), arg3Cb func(T3)) {
	switch e.argId {
	case either3ArgId1:
		arg1Cb(e.arg1)
	case either3ArgId2:
		arg2Cb(e.arg2)
	case either3ArgId3:
		arg3Cb(e.arg3)
	}
}

// Match executes the given function, depending of the argument set, and returns result.
func (e Either3[T1, T2, T3]) Match(
	onArg1 func(T1) Either3[T1, T2, T3],
	onArg2 func(T2) Either3[T1, T2, T3],
	onArg3 func(T3) Either3[T1, T2, T3]) Either3[T1, T2, T3] {

	switch e.argId {
	case either3ArgId1:
		return onArg1(e.arg1)
	case either3ArgId2:
		return onArg2(e.arg2)
	case either3ArgId3:
		return onArg3(e.arg3)
	}

	panic(either3InvalidArgumentId)
}

// MapArg1 executes the given function, if Either3 use the first argument, and returns result.
func (e Either3[T1, T2, T3]) MapArg1(mapper func(T1) Either3[T1, T2, T3]) Either3[T1, T2, T3] {
	if e.IsArg1() {
		return mapper(e.arg1)
	}

	return e
}

// MapArg2 executes the given function, if Either3 use the second argument, and returns result.
func (e Either3[T1, T2, T3]) MapArg2(mapper func(T2) Either3[T1, T2, T3]) Either3[T1, T2, T3] {
	if e.IsArg2() {
		return mapper(e.arg2)
	}

	return e
}

// MapArg3 executes the given function, if Either3 use the third argument, and returns result.
func (e Either3[T1, T2, T3]) MapArg3(mapper func(T3) Either3[T1, T2, T3]) Either3[T1, T2, T3] {
	if e.IsArg3() {
		return mapper(e.arg3)
	}

	return e
}
