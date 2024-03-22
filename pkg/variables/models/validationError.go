package models

type ValidationError struct {
	Err error
}

func (err ValidationError) Error() string {
	return err.Err.Error()
}

func (err *ValidationError) Unwrap() error {
	return err.Err
}
