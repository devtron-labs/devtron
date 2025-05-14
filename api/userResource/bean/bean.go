/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bean

import (
	bean2 "github.com/devtron-labs/devtron/pkg/appWorkflow/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/k8s/bean"
)

const (
	PathParamKind    = "kind"
	PathParamVersion = "version"
)

type PathParams struct {
	Kind    string
	Version string
}

type ResourceOptionsReqDto struct {
	EntityAccessType
	AppAndJobReqDto
	ClusterReqDto
	JobWorkflowReqDto
}
type EntityAccessType struct {
	Entity     string `json:"entity"`
	AccessType string `json:"accessType"`
}

type AppAndJobReqDto struct {
	TeamIds []int `json:"teamIds"`
}
type ClusterReqDto struct {
	*bean3.ResourceRequestBean
}
type JobWorkflowReqDto struct {
	*bean2.WorkflowNamesRequest
}
