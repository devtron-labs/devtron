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
	r.Steps = steps
	return r
}

func (r *ImageScanningSteps) WithScanToolId(scanToolId int) *ImageScanningSteps {
	r.ScanToolId = scanToolId
	return r
}

func (workflowRequest *WorkflowRequest) SetExecuteImageScanningVia(scanVia bean2.ExecutionMedium) {
	return
}

func (workflowRequest *WorkflowRequest) SetImageScanningSteps(imageScanningSteps []*ImageScanningSteps) {
	return
}
