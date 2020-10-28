package gocd

// StageInstance represents the stage from the result from a pipeline run
// codebeat:disable[TOO_MANY_IVARS]
type StageInstance struct {
	Name              string `json:"name"`
	ID                int    `json:"id"`
	Jobs              []*Job `json:"jobs,omitempty"`
	CanRun            bool   `json:"can_run"`
	Scheduled         bool   `json:"scheduled"`
	ApprovalType      string `json:"approval_type,omitempty"`
	ApprovedBy        string `json:"approved_by,omitempty"`
	Counter           string `json:"counter,omitempty"`
	OperatePermission bool   `json:"operate_permission,omitempty"`
	Result            string `json:"result,omitempty"`
	RerunOfCounter    *int   `json:"rerun_of_counter,omitempty"`
}

// codebeat:enable[TOO_MANY_IVARS]
