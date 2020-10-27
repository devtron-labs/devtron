package gocd

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-version"
	"net/url"
)

// PipelinesService describes the HAL _link resource for the api response object for a pipelineconfig
type PipelinesService service

// PipelineRequest describes a pipeline request object
type PipelineRequest struct {
	Group    string    `json:"group"`
	Pipeline *Pipeline `json:"pipeline"`
}

// Pipeline describes a pipeline object
// codebeat:disable[TOO_MANY_IVARS]
type Pipeline struct {
	Links                 *HALLinks              `json:"_links,omitempty"`
	Group                 string                 `json:"group,omitempty"`                   // Group is only used/set when creating or editing a pipeline config
	LabelTemplate         string                 `json:"label_template,omitempty"`          // LabelTemplate is available for the pipeline config API since v1 (GoCD >= 15.3.0).
	EnablePipelineLocking bool                   `json:"enable_pipeline_locking,omitempty"` // EnablePipelineLocking is available for the pipeline config API v1 to v4 (GoCD >= 15.3.0 to GoCD < 17.12.0). Use LockBehavior after that.
	Name                  string                 `json:"name"`                              // Name is available for the pipeline config API since v1 (GoCD >= 15.3.0).
	LockBehavior          string                 `json:"lock_behavior,omitempty"`           // LockBehavior is available for the pipeline config API v5 and v6 only (GoCD >= 17.12.0).
	Template              string                 `json:"template,omitempty"`                // Template is available for the pipeline config API since v1 (GoCD >= 15.3.0).
	Origin                *PipelineConfigOrigin  `json:"origin,omitempty"`                  // Origin is available for the pipeline config API since v3 (GoCD >= 17.4.0).
	Parameters            []*Parameter           `json:"parameters,omitempty"`              // Parameters is available for the pipeline config API since v1 (GoCD >= 15.3.0).
	EnvironmentVariables  []*EnvironmentVariable `json:"environment_variables,omitempty"`   // EnvironmentVariables is available for the pipeline config API since v1 (GoCD >= 15.3.0).
	Materials             []Material             `json:"materials,omitempty"`               // Materials is available for the pipeline config API since v1 (GoCD >= 15.3.0).
	Label                 string                 `json:"label,omitempty"`                   // Label is only available for the pipeline instance
	Stages                []*Stage               `json:"stages,omitempty"`                  // Stages is available for the pipeline config API since v1 (GoCD >= 15.3.0).
	TrackingTool          *TrackingTool          `json:"tracking_tool"`                     // TrackingTool is available for the pipeline config API since v1 (GoCD >= 15.3.0).
	Timer                 *Timer                 `json:"timer"`                             // Timer is available for the pipeline config API since v1 (GoCD >= 15.3.0).
	Version               string                 `json:"version,omitempty"`                 // Version corresponds to the ETag header used when updating a pipeline config
}

// codebeat:enable[TOO_MANY_IVARS]

// Timer describes the cron-like schedule to build a pipeline
type Timer struct {
	Spec          string `json:"spec,omitempty"`
	OnlyOnChanges bool   `json:"only_on_changes,omitempty"`
}

// TrackingTool describes the type of a tracking tool and its attributes
type TrackingTool struct {
	Type       string                 `json:"type,omitempty"`
	Attributes TrackingToolAttributes `json:"attributes,omitempty"`
}

// TrackingToolAttributes describes the attributes of a tracking tool
type TrackingToolAttributes struct {
	URLPattern            string `json:"url_pattern,omitempty"`
	Regex                 string `json:"regex,omitempty"`
	BaseURL               string `json:"base_url,omitempty"`
	ProjectIdentifier     string `json:"project_identifier,omitempty"`
	MqlGroupingConditions string `json:"mql_grouping_conditions,omitempty"`
}

// Parameter represents a key/value
type Parameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// PipelineConfigOrigin describes where a pipeline config is being loaded from
type PipelineConfigOrigin struct {
	Type string `json:"type"`
	File string `json:"file"`
}

// Material describes an artifact dependency for a pipeline object.
type Material struct {
	Type        string            `json:"type"`
	Fingerprint string            `json:"fingerprint,omitempty"`
	Description string            `json:"description,omitempty"`
	Attributes  MaterialAttribute `json:"attributes"`
}

// MaterialFilter describes which globs to ignore
type MaterialFilter struct {
	Ignore []string `json:"ignore"`
}

// PipelineHistory describes the history of runs for a pipeline
type PipelineHistory struct {
	Pipelines []*PipelineInstance `json:"pipelines"`
}

// PipelineInstance describes a single pipeline run
// codebeat:disable[TOO_MANY_IVARS]
type PipelineInstance struct {
	BuildCause          BuildCause `json:"build_cause"`
	Label               string     `json:"label"`
	Counter             int        `json:"counter"`
	PreparingToSchedule bool       `json:"preparing_to_schedule"`
	CanRun              bool       `json:"can_run"`
	Name                string     `json:"name"`
	NaturalOrder        float32    `json:"natural_order"`
	Comment             string     `json:"comment"`
	Stages              []*Stage   `json:"stages"`
}

// codebeat:enable[TOO_MANY_IVARS]

// BuildCause describes the triggers which caused the build to start.
type BuildCause struct {
	Approver          string             `json:"approver,omitempty"`
	MaterialRevisions []MaterialRevision `json:"material_revisions"`
	TriggerForced     bool               `json:"trigger_forced"`
	TriggerMessage    string             `json:"trigger_message"`
}

// MaterialRevision describes the uniquely identifiable version for the material which was pulled for this build
type MaterialRevision struct {
	Modifications []Modification `json:"modifications"`
	Material      struct {
		Description string `json:"description"`
		Fingerprint string `json:"fingerprint"`
		Type        string `json:"type"`
		ID          int    `json:"id"`
	} `json:"material"`
	Changed bool `json:"changed"`
}

// Modification describes the commit/revision for the material which kicked off the build.
type Modification struct {
	EmailAddress string `json:"email_address"`
	ID           int    `json:"id"`
	ModifiedTime int    `json:"modified_time"`
	UserName     string `json:"user_name"`
	Comment      string `json:"comment"`
	Revision     string `json:"revision"`
}

// PipelineStatus describes whether a pipeline can be run or scheduled.
type PipelineStatus struct {
	Locked      bool `json:"locked"`
	Paused      bool `json:"paused"`
	Schedulable bool `json:"schedulable"`
}

// ScheduleMaterial describes a material that must be used to trigger a new instance of the pipeline.
type ScheduleMaterial struct {
	Name        string `json:"name,omitempty"` // Name is used to build a post query for GoCD version < 18.2.0
	Fingerprint string `json:"fingerprint"`
	Revision    string `json:"revision"`
}

// ScheduleRequestBody describes properties to trigger a new instance of the pipeline.
type ScheduleRequestBody struct {
	EnvironmentVariables            []*EnvironmentVariable `json:"environment_variables,omitempty"`
	Materials                       []*ScheduleMaterial    `json:"materials,omitempty"`
	UpdateMaterialsBeforeScheduling bool                   `json:"update_materials_before_scheduling"`
}

// pipelineActionRequest describes pipeline action details
type pipelineActionRequest struct {
	Action   string
	Pipeline string
	Body     interface{}
	RawQuery string
}

// GetStatus returns a list of pipeline instanves describing the pipeline history.
func (pgs *PipelinesService) GetStatus(ctx context.Context, name string, offset int) (ps *PipelineStatus, resp *APIResponse, err error) {
	ps = &PipelineStatus{}
	_, resp, err = pgs.client.getAction(ctx, &APIClientRequest{
		Path:         fmt.Sprintf("pipelines/%s/status", name),
		ResponseBody: ps,
	})

	return
}

// Pause allows a pipeline to handle new build events
func (pgs *PipelinesService) Pause(ctx context.Context, name string) (bool, *APIResponse, error) {
	return pgs.pipelineAction(ctx, &pipelineActionRequest{
		Action:   "pause",
		Pipeline: name,
	})
}

// Unpause allows a pipeline to handle new build events
func (pgs *PipelinesService) Unpause(ctx context.Context, name string) (bool, *APIResponse, error) {
	return pgs.pipelineAction(ctx, &pipelineActionRequest{
		Action:   "unpause",
		Pipeline: name,
	})
}

// ReleaseLock frees a pipeline to handle new build events
func (pgs *PipelinesService) ReleaseLock(ctx context.Context, name string) (bool, *APIResponse, error) {
	request := &pipelineActionRequest{
		Action:   "unlock",
		Pipeline: name,
	}

	// if version is >= 18.2.0 use "unlock"
	releaseLockBeforeVersion, _ := version.NewVersion("18.2.0")
	v, _, err := pgs.client.ServerVersion.Get(context.Background())
	if err != nil {
		return false, nil, err
	}

	if v.VersionParts.LessThan(releaseLockBeforeVersion) {
		request.Action = "releaseLock"
	}

	return pgs.pipelineAction(ctx, request)
}

// Schedule allows to trigger a specific pipeline.
func (pgs *PipelinesService) Schedule(ctx context.Context, name string, body *ScheduleRequestBody) (bool, *APIResponse, error) {

	request := pipelineActionRequest{
		Pipeline: name,
		Action:   "schedule",
		RawQuery: buildScheduleQuery(body),
	}

	if body != nil {
		request.Body = body
	}

	return pgs.pipelineAction(ctx, &request)
}

// GetInstance of a pipeline run.
func (pgs *PipelinesService) GetInstance(ctx context.Context, name string, counter int) (pt *PipelineInstance, resp *APIResponse, err error) {

	pt = &PipelineInstance{}
	_, resp, err = pgs.client.getAction(ctx, &APIClientRequest{
		Path:         fmt.Sprintf("pipelines/%s/instance/%d", name, counter),
		ResponseBody: &pt,
	})

	return
}

// GetHistory returns a list of pipeline instances describing the pipeline history.
func (pgs *PipelinesService) GetHistory(ctx context.Context, name string, offset int) (pt *PipelineHistory, resp *APIResponse, err error) {

	pt = &PipelineHistory{}
	_, resp, err = pgs.client.getAction(ctx, &APIClientRequest{
		Path:         pgs.buildPaginatedStub("pipelines/%s/history", name, offset),
		ResponseBody: &pt,
	})

	return
}

func (pgs *PipelinesService) pipelineAction(ctx context.Context, request *pipelineActionRequest) (bool, *APIResponse, error) {

	apiVersion, err := pgs.client.getAPIVersion(ctx, fmt.Sprintf("pipelines/:pipeline_name/%s", request.Action))
	if err != nil {
		return false, nil, err
	}

	apiRequest := &APIClientRequest{
		Path:        fmt.Sprintf("pipelines/%s/%s", request.Pipeline, request.Action),
		APIVersion:  apiVersion,
		RequestBody: request.Body,
	}

	setPostData(apiRequest, request.RawQuery, apiVersion)

	choosePipelineConfirmHeader(apiRequest, apiVersion)

	_, resp, err := pgs.client.postAction(ctx, apiRequest)
	if err != nil {
		return false, nil, err
	}

	return resp.HTTP.StatusCode == 200, resp, err
}

func (pgs *PipelinesService) buildPaginatedStub(format string, name string, offset int) (stub string) {
	stub = fmt.Sprintf(format, name)
	if offset > 0 {
		stub = fmt.Sprintf("%s/%d", stub, offset)
	}
	return
}

func choosePipelineConfirmHeader(request *APIClientRequest, apiVersion string) {
	request.Headers = map[string]string{"X-GoCD-Confirm": "true"}
	if apiVersion == apiV0 {
		request.Headers = map[string]string{"Confirm": "true"}
	} else {
		request.ResponseType = responseTypeJSON
		request.ResponseBody = &map[string]interface{}{}
	}

}

func setPostData(request *APIClientRequest, query string, apiVersion string) {
	if query != "" && apiVersion == apiV0 {
		request.Path = fmt.Sprintf("%s?%s", request.Path, query)
		request.RequestBody = nil
	}
}

func buildScheduleQuery(body *ScheduleRequestBody) string {
	if body == nil {
		return ""
	}

	params := url.Values{}

	for _, m := range body.Materials {
		params.Add(fmt.Sprintf("materials[%s]", m.Name), m.Revision)
	}

	for _, v := range body.EnvironmentVariables {
		if v.Secure {
			params.Add(fmt.Sprintf("secure_variables[%s]", v.Name), v.Value)
		} else {
			params.Add(fmt.Sprintf("variables[%s]", v.Name), v.Value)
		}
	}

	return params.Encode()
}
