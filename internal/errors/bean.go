package errors

import (
	"github.com/devtron-labs/devtron/internal/constants"
	"google.golang.org/grpc/codes"
	"net/http"
)

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

func (r *ClientStatusCode) IsDeadlineExceededCode() bool {
	return r.Code == codes.DeadlineExceeded
}

func (r *ClientStatusCode) IsCanceledCode() bool {
	return r.Code == codes.Canceled
}

func (r *ClientStatusCode) GetHttpStatusCodeForGivenGrpcCode() int {
	switch r.Code {
	case codes.InvalidArgument:
		return http.StatusConflict
	case codes.NotFound:
		return http.StatusNotFound
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed
	case codes.DeadlineExceeded:
		return http.StatusRequestTimeout
	case codes.Canceled:
		return constants.HttpClientSideTimeout
	default:
		return http.StatusInternalServerError
	}
}
