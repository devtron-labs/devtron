package bean

type ExecutionMedium string

const (
	Rest  ExecutionMedium = "rest"
	Steps ExecutionMedium = "steps"
)

func (e ExecutionMedium) IsExecutionMediumRest() bool {
	return e == Rest
}

func (e ExecutionMedium) IsExecutionMediumSteps() bool {
	return e == Steps
}
