package errors

import "google.golang.org/grpc/codes"

type ClientStatusCode struct {
	Code codes.Code
}

func (r *ClientStatusCode) IsInvalidArgumentCode() bool {
	return r.Code == codes.InvalidArgument
}

func (r *ClientStatusCode) IsNotFoundCode() bool {
	return r.Code == codes.NotFound
}

func (r *ClientStatusCode) IsFailedPreconditionCode() bool {
	return r.Code == codes.FailedPrecondition
}
