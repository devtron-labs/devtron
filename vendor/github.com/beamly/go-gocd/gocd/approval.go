package gocd

// Approval represents a request/response object describing the approval configuration for a GoCD Job
type Approval struct {
	Type          string         `json:"type"`
	Authorization *Authorization `json:"authorization"`
}

// Authorization describes the access control for a "manual" approval type. Specifies who (role or users) can approve
// the job to move to the next stage in the pipeline.
type Authorization struct {
	Users []string `json:"users"`
	Roles []string `json:"roles"`
}
