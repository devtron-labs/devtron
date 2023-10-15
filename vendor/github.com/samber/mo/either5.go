package mo

import "fmt"

const (
	either5ArgId1 = iota
	either5ArgId2
	either5ArgId3
	either5ArgId4
	either5ArgId5
)

var (
	either5InvalidArgumentId = fmt.Errorf("either5 argument should be between 1 and 5")
	either5MissingArg1       = fmt.Errorf("either5 doesn't contain expected argument 1")
	either5MissingArg2       = fmt.Errorf("either5 doesn't contain expected argument 2")
	either5MissingArg3       = fmt.Errorf("either5 doesn't contain expected argument 3")
	either5MissingArg4       = fmt.Errorf("either5 doesn't contain expected argument 4")
	either5MissingArg5       = fmt.Errorf("either5 doesn't contain expected argument 5")
)

// NewEither5Arg1 builds the first argument of the Either5 struct.
func NewEither5Arg1[T1 any, T2 any, T3 any, T4 any, T5 any](value T1) Either5[T1, T2, T3, T4, T5] {
	return Either5[T1, T2, T3, T4, T5]{
		argId: either5ArgId1,
		arg1:  value,
	}
}

// NewEither5Arg2 builds the second argument of the Either5 struct.
func NewEither5Arg2[T1 any, T2 any, T3 any, T4 any, T5 any](value T2) Either5[T1, T2, T3, T4, T5] {
	return Either5[T1, T2, T3, T4, T5]{
		argId: either5ArgId2,
		arg2:  value,
	}
}

// NewEither5Arg3 builds the third argument of the Either5 struct.
func NewEither5Arg3[T1 any, T2 any, T3 any, T4 any, T5 any](value T3) Either5[T1, T2, T3, T4, T5] {
	return Either5[T1, T2, T3, T4, T5]{
		argId: either5ArgId3,
		arg3:  value,
	}
}

// NewEither5Arg4 builds the fourth argument of the Either5 struct.
func NewEither5Arg4[T1 any, T2 any, T3 any, T4 any, T5 any](value T4) Either5[T1, T2, T3, T4, T5] {
	return Either5[T1, T2, T3, T4, T5]{
		argId: either5ArgId4,
		arg4:  value,
	}
}

// NewEither5Arg5 builds the fith argument of the Either5 struct.
func NewEither5Arg5[T1 any, T2 any, T3 any, T4 any, T5 any](value T5) Either5[T1, T2, T3, T4, T5] {
	return Either5[T1, T2, T3, T4, T5]{
		argId: either5ArgId5,
		arg5:  value,
	}
}

// Either5 respresents a value of 5 possible types.
// An instance of Either5 is an instance of either T1, T2, T3, T4, or T5.
type Either5[T1 any, T2 any, T3 any, T4 any, T5 any] struct {
	argId int8

	arg1 T1
	arg2 T2
	arg3 T3
	arg4 T4
	arg5 T5
}

// IsArg1 returns true if Either5 uses the first argument.
func (e Either5[T1, T2, T3, T4, T5]) IsArg1() bool {
	return e.argId == either5ArgId1
}

// IsArg2 returns true if Either5 uses the second argument.
func (e Either5[T1, T2, T3, T4, T5]) IsArg2() bool {
	return e.argId == either5ArgId2
}

// IsArg3 returns true if Either5 uses the third argument.
func (e Either5[T1, T2, T3, T4, T5]) IsArg3() bool {
	return e.argId == either5ArgId3
}

// IsArg4 returns true if Either5 uses the fourth argument.
func (e Either5[T1, T2, T3, T4, T5]) IsArg4() bool {
	return e.argId == either5ArgId4
}

// IsArg5 returns true if Either5 uses the fith argument.
func (e Either5[T1, T2, T3, T4, T5]) IsArg5() bool {
	return e.argId == either5ArgId5
}

// Arg1 returns the first argument of a Either5 struct.
func (e Either5[T1, T2, T3, T4, T5]) Arg1() (T1, bool) {
	if e.IsArg1() {
		return e.arg1, true
	}
	return empty[T1](), false
}

// Arg2 returns the second argument of a Either5 struct.
func (e Either5[T1, T2, T3, T4, T5]) Arg2() (T2, bool) {
	if e.IsArg2() {
		return e.arg2, true
	}
	return empty[T2](), false
}

// Arg3 returns the third argument of a Either5 struct.
func (e Either5[T1, T2, T3, T4, T5]) Arg3() (T3, bool) {
	if e.IsArg3() {
		return e.arg3, true
	}
	return empty[T3](), false
}

// Arg4 returns the fourth argument of a Either5 struct.
func (e Either5[T1, T2, T3, T4, T5]) Arg4() (T4, bool) {
	if e.IsArg4() {
		return e.arg4, true
	}
	return empty[T4](), false
}

// Arg5 returns the fith argument of a Either5 struct.
func (e Either5[T1, T2, T3, T4, T5]) Arg5() (T5, bool) {
	if e.IsArg5() {
		return e.arg5, true
	}
	return empty[T5](), false
}

// MustArg1 returns the first argument of a Either5 struct or panics.
func (e Either5[T1, T2, T3, T4, T5]) MustArg1() T1 {
	if !e.IsArg1() {
		panic(either5MissingArg1)
	}
	return e.arg1
}

// MustArg2 returns the second argument of a Either5 struct or panics.
func (e Either5[T1, T2, T3, T4, T5]) MustArg2() T2 {
	if !e.IsArg2() {
		panic(either5MissingArg2)
	}
	return e.arg2
}

// MustArg3 returns the third argument of a Either5 struct or panics.
func (e Either5[T1, T2, T3, T4, T5]) MustArg3() T3 {
	if !e.IsArg3() {
		panic(either5MissingArg3)
	}
	return e.arg3
}

// MustArg4 returns the fourth argument of a Either5 struct or panics.
func (e Either5[T1, T2, T3, T4, T5]) MustArg4() T4 {
	if !e.IsArg4() {
		panic(either5MissingArg4)
	}
	return e.arg4
}

// MustArg5 returns the fith argument of a Either5 struct or panics.
func (e Either5[T1, T2, T3, T4, T5]) MustArg5() T5 {
	if !e.IsArg5() {
		panic(either5MissingArg5)
	}
	return e.arg5
}

// Unpack returns all values
func (e Either5[T1, T2, T3, T4, T5]) Unpack() (T1, T2, T3, T4, T5) {
	return e.arg1, e.arg2, e.arg3, e.arg4, e.arg5
}

// Arg1OrElse returns the first argument of a Either5 struct or fallback.
func (e Either5[T1, T2, T3, T4, T5]) Arg1OrElse(fallback T1) T1 {
	if e.IsArg1() {
		return e.arg1
	}
	return fallback
}

// Arg2OrElse returns the second argument of a Either5 struct or fallback.
func (e Either5[T1, T2, T3, T4, T5]) Arg2OrElse(fallback T2) T2 {
	if e.IsArg2() {
		return e.arg2
	}
	return fallback
}

// Arg3OrElse returns the third argument of a Either5 struct or fallback.
func (e Either5[T1, T2, T3, T4, T5]) Arg3OrElse(fallback T3) T3 {
	if e.IsArg3() {
		return e.arg3
	}
	return fallback
}

// Arg4OrElse returns the fourth argument of a Either5 struct or fallback.
func (e Either5[T1, T2, T3, T4, T5]) Arg4OrElse(fallback T4) T4 {
	if e.IsArg4() {
		return e.arg4
	}
	return fallback
}

// Arg5OrElse returns the fith argument of a Either5 struct or fallback.
func (e Either5[T1, T2, T3, T4, T5]) Arg5OrElse(fallback T5) T5 {
	if e.IsArg5() {
		return e.arg5
	}
	return fallback
}

// Arg1OrEmpty returns the first argument of a Either5 struct or empty value.
func (e Either5[T1, T2, T3, T4, T5]) Arg1OrEmpty() T1 {
	if e.IsArg1() {
		return e.arg1
	}
	return empty[T1]()
}

// Arg2OrEmpty returns the second argument of a Either5 struct or empty value.
func (e Either5[T1, T2, T3, T4, T5]) Arg2OrEmpty() T2 {
	if e.IsArg2() {
		return e.arg2
	}
	return empty[T2]()
}

// Arg3OrEmpty returns the third argument of a Either5 struct or empty value.
func (e Either5[T1, T2, T3, T4, T5]) Arg3OrEmpty() T3 {
	if e.IsArg3() {
		return e.arg3
	}
	return empty[T3]()
}

// Arg4OrEmpty returns the fourth argument of a Either5 struct or empty value.
func (e Either5[T1, T2, T3, T4, T5]) Arg4OrEmpty() T4 {
	if e.IsArg4() {
		return e.arg4
	}
	return empty[T4]()
}

// Arg5OrEmpty returns the fifth argument of a Either5 struct or empty value.
func (e Either5[T1, T2, T3, T4, T5]) Arg5OrEmpty() T5 {
	if e.IsArg5() {
		return e.arg5
	}
	return empty[T5]()
}

// ForEach executes the given side-effecting function, depending of the argument set.
func (e Either5[T1, T2, T3, T4, T5]) ForEach(arg1Cb func(T1), arg2Cb func(T2), arg3Cb func(T3), arg4Cb func(T4), arg5Cb func(T5)) {
	switch e.argId {
	case either5ArgId1:
		arg1Cb(e.arg1)
	case either5ArgId2:
		arg2Cb(e.arg2)
	case either5ArgId3:
		arg3Cb(e.arg3)
	case either5ArgId4:
		arg4Cb(e.arg4)
	case either5ArgId5:
		arg5Cb(e.arg5)
	}
}

// Match executes the given function, depending of the argument set, and returns result.
func (e Either5[T1, T2, T3, T4, T5]) Match(
	onArg1 func(T1) Either5[T1, T2, T3, T4, T5],
	onArg2 func(T2) Either5[T1, T2, T3, T4, T5],
	onArg3 func(T3) Either5[T1, T2, T3, T4, T5],
	onArg4 func(T4) Either5[T1, T2, T3, T4, T5],
	onArg5 func(T5) Either5[T1, T2, T3, T4, T5]) Either5[T1, T2, T3, T4, T5] {

	switch e.argId {
	case either5ArgId1:
		return onArg1(e.arg1)
	case either5ArgId2:
		return onArg2(e.arg2)
	case either5ArgId3:
		return onArg3(e.arg3)
	case either5ArgId4:
		return onArg4(e.arg4)
	case either5ArgId5:
		return onArg5(e.arg5)
	}

	panic(either5InvalidArgumentId)
}

// MapArg1 executes the given function, if Either5 use the first argument, and returns result.
func (e Either5[T1, T2, T3, T4, T5]) MapArg1(mapper func(T1) Either5[T1, T2, T3, T4, T5]) Either5[T1, T2, T3, T4, T5] {
	if e.IsArg1() {
		return mapper(e.arg1)
	}

	return e
}

// MapArg2 executes the given function, if Either5 use the second argument, and returns result.
func (e Either5[T1, T2, T3, T4, T5]) MapArg2(mapper func(T2) Either5[T1, T2, T3, T4, T5]) Either5[T1, T2, T3, T4, T5] {
	if e.IsArg2() {
		return mapper(e.arg2)
	}

	return e
}

// MapArg3 executes the given function, if Either5 use the third argument, and returns result.
func (e Either5[T1, T2, T3, T4, T5]) MapArg3(mapper func(T3) Either5[T1, T2, T3, T4, T5]) Either5[T1, T2, T3, T4, T5] {
	if e.IsArg3() {
		return mapper(e.arg3)
	}

	return e
}

// MapArg4 executes the given function, if Either5 use the fourth argument, and returns result.
func (e Either5[T1, T2, T3, T4, T5]) MapArg4(mapper func(T4) Either5[T1, T2, T3, T4, T5]) Either5[T1, T2, T3, T4, T5] {
	if e.IsArg4() {
		return mapper(e.arg4)
	}

	return e
}

// MapArg5 executes the given function, if Either5 use the fith argument, and returns result.
func (e Either5[T1, T2, T3, T4, T5]) MapArg5(mapper func(T5) Either5[T1, T2, T3, T4, T5]) Either5[T1, T2, T3, T4, T5] {
	if e.IsArg5() {
		return mapper(e.arg5)
	}

	return e
}
