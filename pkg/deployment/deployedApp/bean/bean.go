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

package bean

import util5 "github.com/devtron-labs/common-lib/utils/k8s"

type PodRotateRequest struct {
	AppId               int                        `json:"appId" validate:"required"`
	EnvironmentId       int                        `json:"environmentId" validate:"required"`
	ResourceIdentifiers []util5.ResourceIdentifier `json:"resources" validate:"required"`
	UserId              int32                      `json:"-"`
}

type RequestType string

const (
	START RequestType = "START"
	STOP  RequestType = "STOP"
)

type StopAppRequest struct {
	AppId         int         `json:"appId" validate:"required"`
	EnvironmentId int         `json:"environmentId" validate:"required"`
	UserId        int32       `json:"userId"`
	RequestType   RequestType `json:"requestType" validate:"oneof=START STOP"`
	// ReferenceId is a unique identifier for the workflow runner
	// refer pipelineConfig.CdWorkflowRunner
	ReferenceId *string
}
