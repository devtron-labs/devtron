/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package errors

import (
	util2 "github.com/devtron-labs/devtron/internal/util"
	"google.golang.org/grpc/status"
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
	ForbiddenErrMsg             = "forbidden"
)

var errorHttpStatusCodeMap = map[string]int{
	ClusterUnreachableErrorMsg:  http.StatusUnprocessableEntity,
	CrdPreconditionErrorMsg:     http.StatusPreconditionFailed,
	NamespaceNotFoundErrorMsg:   http.StatusConflict,
	ArrayStringMismatchErrorMsg: http.StatusFailedDependency,
	InvalidValueErrorMsg:        http.StatusFailedDependency,
	OperationInProgressErrorMsg: http.StatusConflict,
	//forbidden error from kubernetes would not be a parameter for us to mark a user forbidden to that resource or not,
	//since this is not rbac from devtron, hence map it to StatusUnprocessableEntity
	ForbiddenErrMsg: http.StatusUnprocessableEntity,
}

func ConvertToApiError(err error) *util2.ApiError {
	var apiError *util2.ApiError
	if _, ok := status.FromError(err); ok {
		clientCode, errMsg := util2.GetClientDetailedError(err)
		httpStatusCode := clientCode.GetHttpStatusCodeForGivenGrpcCode()
		apiError = &util2.ApiError{
			HttpStatusCode:  httpStatusCode,
			Code:            strconv.Itoa(httpStatusCode),
			InternalMessage: errMsg,
			UserMessage:     errMsg,
		}
	} else {
		for errMsg, statusCode := range errorHttpStatusCodeMap {
			if strings.Contains(err.Error(), errMsg) {
				apiError = &util2.ApiError{
					InternalMessage: err.Error(),
					UserMessage:     err.Error(),
					HttpStatusCode:  statusCode,
					Code:            strconv.Itoa(statusCode),
				}
			}
		}
	}

	return apiError
}
