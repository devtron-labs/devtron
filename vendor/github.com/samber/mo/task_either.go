package mo

// NewTaskEither instanciates a new TaskEither.
func NewTaskEither[R any](f ff0[R]) TaskEither[R] {
	return TaskEither[R]{NewTask[R](f)}
}

// NewTaskEitherFromIO instanciates a new TaskEither from an existing IO.
func NewTaskEitherFromIO[R any](io IO[R]) TaskEither[R] {
	return TaskEither[R]{NewTaskFromIO[R](io)}
}

// TaskEither represents a non-deterministic asynchronous computation that
// can cause side effects, yields a value of type `R` and can fail.
type TaskEither[R any] struct {
	Task[R]
}

// OrElse returns value if task succeeded or default value.
func (t TaskEither[R]) OrElse(fallback R) R {
	either := t.Run().Either()

	right, isRight := either.Right()
	if !isRight {
		return fallback
	}

	return right
}

// Match executes the first function if task succeeded and second function if task failed.
// It returns a new Option.
func (t TaskEither[R]) Match(onLeft func(error) Either[error, R], onRight func(R) Either[error, R]) Either[error, R] {
	either := t.Run().Either()
	return either.Match(onLeft, onRight)
}

// TryCatch is an alias to Match
func (t TaskEither[R]) TryCatch(onLeft func(error) Either[error, R], onRight func(R) Either[error, R]) Either[error, R] {
	return t.Match(onLeft, onRight)
}

// ToTask converts TaskEither to Task
func (t TaskEither[R]) ToTask(fallback R) Task[R] {
	return NewTask(func() *Future[R] {
		return t.Run().
			Catch(func(err error) (R, error) {
				return fallback, nil
			})
	})
}

// ToEither converts TaskEither to Either.
func (t TaskEither[R]) ToEither() Either[error, R] {
	return t.Run().Either()
}
