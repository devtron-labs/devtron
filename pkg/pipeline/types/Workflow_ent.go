package types

import (
	bean2 "github.com/devtron-labs/common-lib/imageScan/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
)

type ImageScanningSteps struct {
	Steps      []*bean.StepObject `json:"steps"`
	ScanToolId int                `json:"scanToolId"`
}

func (r *ImageScanningSteps) WithSteps(steps []*bean.StepObject) {
	r.Steps = steps
}

func (r *ImageScanningSteps) WithScanToolId(scanToolId int) {
	r.ScanToolId = scanToolId
}

func (workflowRequest *WorkflowRequest) SetExecuteImageScanningVia(scanVia bean2.ExecutionMedium) {
	return
}

func (workflowRequest *WorkflowRequest) SetImageScanningSteps(imageScanningSteps []*ImageScanningSteps) {
	return
}
