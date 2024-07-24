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

func (r *ClientStatusCode) IsUnavailableCode() bool {
	return r.Code == codes.Unavailable
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
	case codes.PermissionDenied:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}
