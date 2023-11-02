package models

type NamespaceNotExistError struct {
	Err error
}

func (err NamespaceNotExistError) Error() string {
	return err.Err.Error()
}

func (err *NamespaceNotExistError) Unwrap() error {
	return err.Err
}
