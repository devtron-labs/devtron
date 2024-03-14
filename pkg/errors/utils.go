package errors

import (
	util2 "github.com/devtron-labs/devtron/internal/util"
	"net/http"
	"strconv"
	"strings"
)

// extract out this custom error messages in kubelink and send custom error messages from kubelink
const (
	ClusterUnreachableErrorMsg  = "cluster unreachable"
	CrdPreconditionErrorMsg     = "ensure CRDs are installed first"
	ArrayStringMismatchErrorMsg = "got array expected string"
	NamespaceNotFoundErrorMsg   = "namespace not found"
	InvalidValueErrorMsg        = "invalid value in manifest"
	OperationInProgressErrorMsg = "another operation (install/upgrade/rollback) is in progress"
)

func ConvertToApiError(err error) *util2.ApiError {
	errorHttpStatusCodeMap := map[string]int{
		ClusterUnreachableErrorMsg:  http.StatusUnprocessableEntity,
		CrdPreconditionErrorMsg:     http.StatusPreconditionFailed,
		NamespaceNotFoundErrorMsg:   http.StatusConflict,
		ArrayStringMismatchErrorMsg: http.StatusFailedDependency,
		InvalidValueErrorMsg:        http.StatusFailedDependency,
		OperationInProgressErrorMsg: http.StatusConflict,
	}
	for errMsg, statusCode := range errorHttpStatusCodeMap {
		if strings.Contains(err.Error(), errMsg) {
			return &util2.ApiError{
				InternalMessage: err.Error(),
				UserMessage:     err.Error(),
				HttpStatusCode:  statusCode,
				Code:            strconv.Itoa(statusCode),
			}
		}
	}
	return nil
}
