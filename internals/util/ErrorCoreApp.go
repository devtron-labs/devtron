package util

type InternalServerError struct {
	E error
}

func (err *InternalServerError) Error() string {
	return err.Error()
}

type BadRequestError struct {
	E error
}

func (err *BadRequestError) Error() string {
	return err.Error()
}

type ForbiddenError struct {
	E error
}

func (err *ForbiddenError) Error() string {
	return err.Error()
}
