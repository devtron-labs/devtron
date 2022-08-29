package devtron_integration_manager

type TaskExecutorService interface {
	executeTask(taskName string) error
}
