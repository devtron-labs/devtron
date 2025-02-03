package types

import (
	bean2 "github.com/devtron-labs/common-lib/imageScan/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
)

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
