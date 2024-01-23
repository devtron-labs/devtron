package bean

type DeploymentApprovalValidationError struct {
	Err           error
	ApprovalState ApprovalState
}

func (err DeploymentApprovalValidationError) Error() string {
	return err.Err.Error()
}

func (err *DeploymentApprovalValidationError) Unwrap() error {
	return err.Err
}
