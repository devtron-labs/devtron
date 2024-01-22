package drafts

type DraftApprovalValidationError struct {
	Err        error
	DraftState DraftState
}

func (err DraftApprovalValidationError) Error() string {
	return err.Err.Error()
}

func (err *DraftApprovalValidationError) Unwrap() error {
	return err.Err
}
