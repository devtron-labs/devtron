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

package types

import (
	bean2 "github.com/devtron-labs/common-lib/imageScan/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
)

type WorkflowRequestEnt struct {
}

type ConfigMapSecretEntDto struct {
}

type ImageScanningSteps struct {
	Steps      []*bean.StepObject `json:"steps"`
	ScanToolId int                `json:"scanToolId"`
}

func NewImageScanningStepsDto() *ImageScanningSteps {
	return &ImageScanningSteps{}
}

func (r *ImageScanningSteps) WithSteps(steps []*bean.StepObject) *ImageScanningSteps {
	return r
}

func (r *ImageScanningSteps) WithScanToolId(scanToolId int) *ImageScanningSteps {
	return r
}

func (workflowRequest *WorkflowRequest) SetExecuteImageScanningVia(scanVia bean2.ScanExecutionMedium) {
	return
}

func (workflowRequest *WorkflowRequest) SetImageScanningSteps(imageScanningSteps []*ImageScanningSteps) {
	return
}

func (workflowRequest *WorkflowRequest) SetAwsInspectorConfig(awsInspectorConfig string) {
	return
}
