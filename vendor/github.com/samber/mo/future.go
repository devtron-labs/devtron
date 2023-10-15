package mo

import (
	"sync"
)

// NewFuture instanciate a new future.
func NewFuture[T any](cb func(resolve func(T), reject func(error))) *Future[T] {
	future := Future[T]{
		cb:       cb,
		cancelCb: func() {},
		done:     make(chan struct{}),
	}

	future.active()

	return &future
}

// Future represents a value which may or may not currently be available, but will be
// available at some point, or an exception if that value could not be made available.
type Future[T any] struct {
	mu sync.Mutex

	cb       func(func(T), func(error))
	cancelCb func()
	next     *Future[T]
	done     chan struct{}
	result   Result[T]
}

func (f *Future[T]) active() {
	go f.cb(f.resolve, f.reject)
}

func (f *Future[T]) activeSync() {
	f.cb(f.resolve, f.reject)
}

func (f *Future[T]) resolve(value T) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.result = Ok(value)
	if f.next != nil {
		f.next.activeSync()
	}
	close(f.done)
}

func (f *Future[T]) reject(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.result = Err[T](err)
	if f.next != nil {
		f.next.activeSync()
	}
	close(f.done)
}

// Then is called when Future is resolved. It returns a new Future.
func (f *Future[T]) Then(cb func(T) (T, error)) *Future[T] {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.next = &Future[T]{
		cb: func(resolve func(T), reject func(error)) {
			if f.result.IsError() {
				reject(f.result.Error())
				return
			}
			newValue, err := cb(f.result.MustGet())
			if err != nil {
				reject(err)
				return
			}
			resolve(newValue)
		},
		cancelCb: func() {
			f.Cancel()
		},
		done: make(chan struct{}),
	}

	select {
	case <-f.done:
		f.next.active()
	default:
	}
	return f.next
}

// Catch is called when Future is rejected. It returns a new Future.
func (f *Future[T]) Catch(cb func(error) (T, error)) *Future[T] {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.next = &Future[T]{
		cb: func(resolve func(T), reject func(error)) {
			if f.result.IsOk() {
				resolve(f.result.MustGet())
				return
			}
			newValue, err := cb(f.result.Error())
			if err != nil {
				reject(err)
				return
			}
			resolve(newValue)
		},
		cancelCb: func() {
			f.Cancel()
		},
		done: make(chan struct{}),
	}

	select {
	case <-f.done:
		f.next.active()
	default:
	}
	return f.next
}

// Finally is called when Future is processed either resolved or rejected. It returns a new Future.
func (f *Future[T]) Finally(cb func(T, error) (T, error)) *Future[T] {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.next = &Future[T]{
		cb: func(resolve func(T), reject func(error)) {
			newValue, err := cb(f.result.Get())
			if err != nil {
				reject(err)
				return
			}
			resolve(newValue)
		},
		cancelCb: func() {
			f.Cancel()
		},
		done: make(chan struct{}),
	}

	select {
	case <-f.done:
		f.next.active()
	default:
	}
	return f.next
}

// Cancel cancels the Future chain.
func (f *Future[T]) Cancel() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.next = nil
	if f.cancelCb != nil {
		f.cancelCb()
	}
}

// Collect awaits and return result of the Future.
func (f *Future[T]) Collect() (T, error) {
	<-f.done
	return f.result.Get()
}

// Result wraps Collect and returns a Result.
func (f *Future[T]) Result() Result[T] {
	return TupleToResult(f.Collect())
}

// Either wraps Collect and returns a Either.
func (f *Future[T]) Either() Either[error, T] {
	v, err := f.Collect()
	if err != nil {
		return Left[error, T](err)
	}
	return Right[error, T](v)
}
