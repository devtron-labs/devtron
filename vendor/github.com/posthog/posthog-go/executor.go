package posthog

type executor struct {
	queue chan func()
}

func newExecutor(cap int) *executor {
	e := &executor{
		queue: make(chan func(), cap),
	}
	go e.loop()

	return e
}

func (e *executor) do(task func()) bool {
	select {
	case e.queue <- task:
		// task is enqueued successfully
		return true

	default:
		// buffer was full; inform the caller rather than blocking
	}

	return false
}

func (e *executor) close() {
	close(e.queue)
}

func (e *executor) loop() {
	for task := range e.queue {
		capturedTask := task
		go capturedTask()
	}
}
