package mo

// NewIOEither instanciates a new IO.
func NewIOEither[R any](f fe0[R]) IOEither[R] {
	return IOEither[R]{
		unsafePerform: f,
	}
}

// IOEither represents a non-deterministic synchronous computation that
// can cause side effects, yields a value of type `R` and can fail.
type IOEither[R any] struct {
	unsafePerform fe0[R]
}

// Run execute the non-deterministic synchronous computation, with side effect.
func (io IOEither[R]) Run() Either[error, R] {
	v, err := io.unsafePerform()
	if err != nil {
		return Left[error, R](err)
	}

	return Right[error, R](v)
}

// NewIOEither1 instanciates a new IO1.
func NewIOEither1[R any, A any](f fe1[R, A]) IOEither1[R, A] {
	return IOEither1[R, A]{
		unsafePerform: f,
	}
}

// IOEither1 represents a non-deterministic synchronous computation that
// can cause side effects, yields a value of type `R` and can fail.
type IOEither1[R any, A any] struct {
	unsafePerform fe1[R, A]
}

// Run execute the non-deterministic synchronous computation, with side effect.
func (io IOEither1[R, A]) Run(a A) Either[error, R] {
	v, err := io.unsafePerform(a)
	if err != nil {
		return Left[error, R](err)
	}

	return Right[error, R](v)
}

// NewIOEither2 instanciates a new IO2.
func NewIOEither2[R any, A any, B any](f fe2[R, A, B]) IOEither2[R, A, B] {
	return IOEither2[R, A, B]{
		unsafePerform: f,
	}
}

// IOEither2 represents a non-deterministic synchronous computation that
// can cause side effects, yields a value of type `R` and can fail.
type IOEither2[R any, A any, B any] struct {
	unsafePerform fe2[R, A, B]
}

// Run execute the non-deterministic synchronous computation, with side effect.
func (io IOEither2[R, A, B]) Run(a A, b B) Either[error, R] {
	v, err := io.unsafePerform(a, b)
	if err != nil {
		return Left[error, R](err)
	}

	return Right[error, R](v)
}

// NewIOEither3 instanciates a new IO3.
func NewIOEither3[R any, A any, B any, C any](f fe3[R, A, B, C]) IOEither3[R, A, B, C] {
	return IOEither3[R, A, B, C]{
		unsafePerform: f,
	}
}

// IOEither3 represents a non-deterministic synchronous computation that
// can cause side effects, yields a value of type `R` and can fail.
type IOEither3[R any, A any, B any, C any] struct {
	unsafePerform fe3[R, A, B, C]
}

// Run execute the non-deterministic synchronous computation, with side effect.
func (io IOEither3[R, A, B, C]) Run(a A, b B, c C) Either[error, R] {
	v, err := io.unsafePerform(a, b, c)
	if err != nil {
		return Left[error, R](err)
	}

	return Right[error, R](v)
}

// NewIOEither4 instanciates a new IO4.
func NewIOEither4[R any, A any, B any, C any, D any](f fe4[R, A, B, C, D]) IOEither4[R, A, B, C, D] {
	return IOEither4[R, A, B, C, D]{
		unsafePerform: f,
	}
}

// IOEither4 represents a non-deterministic synchronous computation that
// can cause side effects, yields a value of type `R` and can fail.
type IOEither4[R any, A any, B any, C any, D any] struct {
	unsafePerform fe4[R, A, B, C, D]
}

// Run execute the non-deterministic synchronous computation, with side effect.
func (io IOEither4[R, A, B, C, D]) Run(a A, b B, c C, d D) Either[error, R] {
	v, err := io.unsafePerform(a, b, c, d)
	if err != nil {
		return Left[error, R](err)
	}

	return Right[error, R](v)
}

// NewIOEither5 instanciates a new IO5.
func NewIOEither5[R any, A any, B any, C any, D any, E any](f fe5[R, A, B, C, D, E]) IOEither5[R, A, B, C, D, E] {
	return IOEither5[R, A, B, C, D, E]{
		unsafePerform: f,
	}
}

// IOEither5 represents a non-deterministic synchronous computation that
// can cause side effects, yields a value of type `R` and can fail.
type IOEither5[R any, A any, B any, C any, D any, E any] struct {
	unsafePerform fe5[R, A, B, C, D, E]
}

// Run execute the non-deterministic synchronous computation, with side effect.
func (io IOEither5[R, A, B, C, D, E]) Run(a A, b B, c C, d D, e E) Either[error, R] {
	v, err := io.unsafePerform(a, b, c, d, e)
	if err != nil {
		return Left[error, R](err)
	}

	return Right[error, R](v)
}
