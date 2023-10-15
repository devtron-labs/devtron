package mo

func NewState[S any, A any](f func(state S) (A, S)) State[S, A] {
	return State[S, A]{
		run: f,
	}
}

func ReturnState[S any, A any](x A) State[S, A] {
	return State[S, A]{
		run: func(state S) (A, S) {
			return x, state
		},
	}
}

// State represents a function `(S) -> (A, S)`, where `S` is state, `A` is result.
type State[S any, A any] struct {
	run func(state S) (A, S)
}

// Run executes a computation in the State monad.
func (s State[S, A]) Run(state S) (A, S) {
	return s.run(state)
}

// Get returns the current state.
func (s State[S, A]) Get() State[S, S] {
	return State[S, S]{
		run: func(state S) (S, S) {
			return state, state
		},
	}
}

// Modify the state by applying a function to the current state.
func (s State[S, A]) Modify(f func(state S) S) State[S, A] {
	return State[S, A]{
		run: func(state S) (A, S) {
			return empty[A](), f(state)
		},
	}
}

// Put set the state.
func (s State[S, A]) Put(state S) State[S, A] {
	return State[S, A]{
		run: func(state S) (A, S) {
			return empty[A](), state
		},
	}
}
