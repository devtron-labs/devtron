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

package common

const (
	UnAuthenticated     = "E100"
	UnAuthorized        = "E101"
	BadRequest          = "E102"
	InternalServerError = "E103"
	ResourceNotFound    = "E104"
	UnknownError        = "E105"
	CONTENT_DISPOSITION = "Content-Disposition"
	CONTENT_TYPE        = "Content-Type"
	CONTENT_LENGTH      = "Content-Length"
	APPLICATION_JSON    = "application/json"
)

var errorMessage = map[string]string{
	UnAuthenticated: "User is not authenticated",
	UnAuthorized:    "User is not authorized to perform this action",
}

func ErrorMessage(code string) string {
	return errorMessage[code]
}
