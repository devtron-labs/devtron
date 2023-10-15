package mo

import "fmt"

const (
	either4ArgId1 = iota
	either4ArgId2
	either4ArgId3
	either4ArgId4
)

var (
	either4InvalidArgumentId = fmt.Errorf("either4 argument should be between 1 and 4")
	either4MissingArg1       = fmt.Errorf("either4 doesn't contain expected argument 1")
	either4MissingArg2       = fmt.Errorf("either4 doesn't contain expected argument 2")
	either4MissingArg3       = fmt.Errorf("either4 doesn't contain expected argument 3")
	either4MissingArg4       = fmt.Errorf("either4 doesn't contain expected argument 4")
)

// NewEither4Arg1 builds the first argument of the Either4 struct.
func NewEither4Arg1[T1 any, T2 any, T3 any, T4 any](value T1) Either4[T1, T2, T3, T4] {
	return Either4[T1, T2, T3, T4]{
		argId: either4ArgId1,
		arg1:  value,
	}
}

// NewEither4Arg2 builds the second argument of the Either4 struct.
func NewEither4Arg2[T1 any, T2 any, T3 any, T4 any](value T2) Either4[T1, T2, T3, T4] {
	return Either4[T1, T2, T3, T4]{
		argId: either4ArgId2,
		arg2:  value,
	}
}

// NewEither4Arg3 builds the third argument of the Either4 struct.
func NewEither4Arg3[T1 any, T2 any, T3 any, T4 any](value T3) Either4[T1, T2, T3, T4] {
	return Either4[T1, T2, T3, T4]{
		argId: either4ArgId3,
		arg3:  value,
	}
}

// NewEither4Arg4 builds the fourth argument of the Either4 struct.
func NewEither4Arg4[T1 any, T2 any, T3 any, T4 any](value T4) Either4[T1, T2, T3, T4] {
	return Either4[T1, T2, T3, T4]{
		argId: either4ArgId4,
		arg4:  value,
	}
}

// Either4 respresents a value of 4 possible types.
// An instance of Either4 is an instance of either T1, T2, T3 or T4.
type Either4[T1 any, T2 any, T3 any, T4 any] struct {
	argId int8

	arg1 T1
	arg2 T2
	arg3 T3
	arg4 T4
}

// IsArg1 returns true if Either4 uses the first argument.
func (e Either4[T1, T2, T3, T4]) IsArg1() bool {
	return e.argId == either4ArgId1
}

// IsArg2 returns true if Either4 uses the second argument.
func (e Either4[T1, T2, T3, T4]) IsArg2() bool {
	return e.argId == either4ArgId2
}

// IsArg3 returns true if Either4 uses the third argument.
func (e Either4[T1, T2, T3, T4]) IsArg3() bool {
	return e.argId == either4ArgId3
}

// IsArg4 returns true if Either4 uses the fourth argument.
func (e Either4[T1, T2, T3, T4]) IsArg4() bool {
	return e.argId == either4ArgId4
}

// Arg1 returns the first argument of a Either4 struct.
func (e Either4[T1, T2, T3, T4]) Arg1() (T1, bool) {
	if e.IsArg1() {
		return e.arg1, true
	}
	return empty[T1](), false
}

// Arg2 returns the second argument of a Either4 struct.
func (e Either4[T1, T2, T3, T4]) Arg2() (T2, bool) {
	if e.IsArg2() {
		return e.arg2, true
	}
	return empty[T2](), false
}

// Arg3 returns the third argument of a Either4 struct.
func (e Either4[T1, T2, T3, T4]) Arg3() (T3, bool) {
	if e.IsArg3() {
		return e.arg3, true
	}
	return empty[T3](), false
}

// Arg4 returns the fourth argument of a Either4 struct.
func (e Either4[T1, T2, T3, T4]) Arg4() (T4, bool) {
	if e.IsArg4() {
		return e.arg4, true
	}
	return empty[T4](), false
}

// MustArg1 returns the first argument of a Either4 struct or panics.
func (e Either4[T1, T2, T3, T4]) MustArg1() T1 {
	if !e.IsArg1() {
		panic(either4MissingArg1)
	}
	return e.arg1
}

// MustArg2 returns the second argument of a Either4 struct or panics.
func (e Either4[T1, T2, T3, T4]) MustArg2() T2 {
	if !e.IsArg2() {
		panic(either4MissingArg2)
	}
	return e.arg2
}

// MustArg3 returns the third argument of a Either4 struct or panics.
func (e Either4[T1, T2, T3, T4]) MustArg3() T3 {
	if !e.IsArg3() {
		panic(either4MissingArg3)
	}
	return e.arg3
}

// MustArg4 returns the fourth argument of a Either4 struct or panics.
func (e Either4[T1, T2, T3, T4]) MustArg4() T4 {
	if !e.IsArg4() {
		panic(either4MissingArg4)
	}
	return e.arg4
}

// Unpack returns all values
func (e Either4[T1, T2, T3, T4]) Unpack() (T1, T2, T3, T4) {
	return e.arg1, e.arg2, e.arg3, e.arg4
}

// Arg1OrElse returns the first argument of a Either4 struct or fallback.
func (e Either4[T1, T2, T3, T4]) Arg1OrElse(fallback T1) T1 {
	if e.IsArg1() {
		return e.arg1
	}
	return fallback
}

// Arg2OrElse returns the second argument of a Either4 struct or fallback.
func (e Either4[T1, T2, T3, T4]) Arg2OrElse(fallback T2) T2 {
	if e.IsArg2() {
		return e.arg2
	}
	return fallback
}

// Arg3OrElse returns the third argument of a Either4 struct or fallback.
func (e Either4[T1, T2, T3, T4]) Arg3OrElse(fallback T3) T3 {
	if e.IsArg3() {
		return e.arg3
	}
	return fallback
}

// Arg4OrElse returns the fourth argument of a Either4 struct or fallback.
func (e Either4[T1, T2, T3, T4]) Arg4OrElse(fallback T4) T4 {
	if e.IsArg4() {
		return e.arg4
	}
	return fallback
}

// Arg1OrEmpty returns the first argument of a Either4 struct or empty value.
func (e Either4[T1, T2, T3, T4]) Arg1OrEmpty() T1 {
	if e.IsArg1() {
		return e.arg1
	}
	return empty[T1]()
}

// Arg2OrEmpty returns the second argument of a Either4 struct or empty value.
func (e Either4[T1, T2, T3, T4]) Arg2OrEmpty() T2 {
	if e.IsArg2() {
		return e.arg2
	}
	return empty[T2]()
}

// Arg3OrEmpty returns the third argument of a Either4 struct or empty value.
func (e Either4[T1, T2, T3, T4]) Arg3OrEmpty() T3 {
	if e.IsArg3() {
		return e.arg3
	}
	return empty[T3]()
}

// Arg4OrEmpty returns the fourth argument of a Either4 struct or empty value.
func (e Either4[T1, T2, T3, T4]) Arg4OrEmpty() T4 {
	if e.IsArg4() {
		return e.arg4
	}
	return empty[T4]()
}

// ForEach executes the given side-effecting function, depending of the argument set.
func (e Either4[T1, T2, T3, T4]) ForEach(arg1Cb func(T1), arg2Cb func(T2), arg3Cb func(T3), arg4Cb func(T4)) {
	switch e.argId {
	case either4ArgId1:
		arg1Cb(e.arg1)
	case either4ArgId2:
		arg2Cb(e.arg2)
	case either4ArgId3:
		arg3Cb(e.arg3)
	case either4ArgId4:
		arg4Cb(e.arg4)
	}
}

// Match executes the given function, depending of the argument set, and returns result.
func (e Either4[T1, T2, T3, T4]) Match(
	onArg1 func(T1) Either4[T1, T2, T3, T4],
	onArg2 func(T2) Either4[T1, T2, T3, T4],
	onArg3 func(T3) Either4[T1, T2, T3, T4],
	onArg4 func(T4) Either4[T1, T2, T3, T4]) Either4[T1, T2, T3, T4] {

	switch e.argId {
	case either4ArgId1:
		return onArg1(e.arg1)
	case either4ArgId2:
		return onArg2(e.arg2)
	case either4ArgId3:
		return onArg3(e.arg3)
	case either4ArgId4:
		return onArg4(e.arg4)
	}

	panic(either4InvalidArgumentId)
}

// MapArg1 executes the given function, if Either4 use the first argument, and returns result.
func (e Either4[T1, T2, T3, T4]) MapArg1(mapper func(T1) Either4[T1, T2, T3, T4]) Either4[T1, T2, T3, T4] {
	if e.IsArg1() {
		return mapper(e.arg1)
	}

	return e
}

// MapArg2 executes the given function, if Either4 use the second argument, and returns result.
func (e Either4[T1, T2, T3, T4]) MapArg2(mapper func(T2) Either4[T1, T2, T3, T4]) Either4[T1, T2, T3, T4] {
	if e.IsArg2() {
		return mapper(e.arg2)
	}

	return e
}

// MapArg3 executes the given function, if Either4 use the third argument, and returns result.
func (e Either4[T1, T2, T3, T4]) MapArg3(mapper func(T3) Either4[T1, T2, T3, T4]) Either4[T1, T2, T3, T4] {
	if e.IsArg3() {
		return mapper(e.arg3)
	}

	return e
}

// MapArg4 executes the given function, if Either4 use the fourth argument, and returns result.
func (e Either4[T1, T2, T3, T4]) MapArg4(mapper func(T4) Either4[T1, T2, T3, T4]) Either4[T1, T2, T3, T4] {
	if e.IsArg4() {
		return mapper(e.arg4)
	}

	return e
}
