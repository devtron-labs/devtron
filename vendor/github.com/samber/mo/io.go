package mo

// NewIO instanciates a new IO.
func NewIO[R any](f f0[R]) IO[R] {
	return IO[R]{
		unsafePerform: f,
	}
}

// IO represents a non-deterministic synchronous computation that
// can cause side effects, yields a value of type `R` and never fails.
type IO[R any] struct {
	unsafePerform f0[R]
}

// Run execute the non-deterministic synchronous computation, with side effect.
func (io IO[R]) Run() R {
	return io.unsafePerform()
}

// NewIO1 instanciates a new IO1.
func NewIO1[R any, A any](f f1[R, A]) IO1[R, A] {
	return IO1[R, A]{
		unsafePerform: f,
	}
}

// IO1 represents a non-deterministic synchronous computation that
// can cause side effects, yields a value of type `R` and never fails.
type IO1[R any, A any] struct {
	unsafePerform f1[R, A]
}

// Run execute the non-deterministic synchronous computation, with side effect.
func (io IO1[R, A]) Run(a A) R {
	return io.unsafePerform(a)
}

// NewIO2 instanciates a new IO2.
func NewIO2[R any, A any, B any](f f2[R, A, B]) IO2[R, A, B] {
	return IO2[R, A, B]{
		unsafePerform: f,
	}
}

// IO2 represents a non-deterministic synchronous computation that
// can cause side effects, yields a value of type `R` and never fails.
type IO2[R any, A any, B any] struct {
	unsafePerform f2[R, A, B]
}

// Run execute the non-deterministic synchronous computation, with side effect.
func (io IO2[R, A, B]) Run(a A, b B) R {
	return io.unsafePerform(a, b)
}

// NewIO3 instanciates a new IO3.
func NewIO3[R any, A any, B any, C any](f f3[R, A, B, C]) IO3[R, A, B, C] {
	return IO3[R, A, B, C]{
		unsafePerform: f,
	}
}

// IO3 represents a non-deterministic synchronous computation that
// can cause side effects, yields a value of type `R` and never fails.
type IO3[R any, A any, B any, C any] struct {
	unsafePerform f3[R, A, B, C]
}

// Run execute the non-deterministic synchronous computation, with side effect.
func (io IO3[R, A, B, C]) Run(a A, b B, c C) R {
	return io.unsafePerform(a, b, c)
}

// NewIO4 instanciates a new IO4.
func NewIO4[R any, A any, B any, C any, D any](f f4[R, A, B, C, D]) IO4[R, A, B, C, D] {
	return IO4[R, A, B, C, D]{
		unsafePerform: f,
	}
}

// IO4 represents a non-deterministic synchronous computation that
// can cause side effects, yields a value of type `R` and never fails.
type IO4[R any, A any, B any, C any, D any] struct {
	unsafePerform f4[R, A, B, C, D]
}

// Run execute the non-deterministic synchronous computation, with side effect.
func (io IO4[R, A, B, C, D]) Run(a A, b B, c C, d D) R {
	return io.unsafePerform(a, b, c, d)
}

// NewIO5 instanciates a new IO5.
func NewIO5[R any, A any, B any, C any, D any, E any](f f5[R, A, B, C, D, E]) IO5[R, A, B, C, D, E] {
	return IO5[R, A, B, C, D, E]{
		unsafePerform: f,
	}
}

// IO5 represents a non-deterministic synchronous computation that
// can cause side effects, yields a value of type `R` and never fails.
type IO5[R any, A any, B any, C any, D any, E any] struct {
	unsafePerform f5[R, A, B, C, D, E]
}

// Run execute the non-deterministic synchronous computation, with side effect.
func (io IO5[R, A, B, C, D, E]) Run(a A, b B, c C, d D, e E) R {
	return io.unsafePerform(a, b, c, d, e)
}
