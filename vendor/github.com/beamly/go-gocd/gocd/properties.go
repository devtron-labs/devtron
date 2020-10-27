package gocd

import (
	"bytes"
	"context"
	"fmt"
)

// PropertiesService describes Actions which can be performed on agents
type PropertiesService service

// PropertyRequest describes the parameters to be submitted when calling/creating properties.
// codebeat:disable[TOO_MANY_IVARS]
type PropertyRequest struct {
	Pipeline        string
	PipelineCounter int
	Stage           string
	StageCounter    int
	Job             string
	LimitPipeline   string
	Limit           int
	Single          bool
}

// codebeat:enable[TOO_MANY_IVARS]

// PropertyCreateResponse handles the parsing of the response when creating a property
type PropertyCreateResponse struct {
	Name  string
	Value string
}

// List the properties for the given job/pipeline/stage run.
func (ps *PropertiesService) List(ctx context.Context, pr *PropertyRequest) (*Properties, *APIResponse, error) {

	ps.log.WithField("endpoint", "PropertiesServices.List").Info("Calling endpoint")
	return ps.commonPropertiesAction(ctx, fmt.Sprintf("/properties/%s/%d/%s/%d/%s",
		pr.Pipeline, pr.PipelineCounter,
		pr.Stage, pr.StageCounter,
		pr.Job,
	), pr.Single)
}

// Get a specific property for the given job/pipeline/stage run.
func (ps *PropertiesService) Get(ctx context.Context, name string, pr *PropertyRequest) (*Properties, *APIResponse, error) {
	ps.log.WithField("endpoint", "PropertiesServices.Get").Info("Calling endpoint")
	return ps.commonPropertiesAction(ctx, fmt.Sprintf("/properties/%s/%d/%s/%d/%s/%s",
		pr.Pipeline, pr.PipelineCounter,
		pr.Stage, pr.StageCounter,
		pr.Job, name,
	), true)
}

// Create a specific property for the given job/pipeline/stage run.
func (ps *PropertiesService) Create(ctx context.Context, name string, value string, pr *PropertyRequest) (responseIsValid bool, resp *APIResponse, err error) {
	responseBuffer := bytes.NewBuffer([]byte(""))

	ps.log.WithField("endpoint", "PropertiesServices.Create").Info("Calling endpoint")
	_, resp, err = ps.client.postAction(ctx, &APIClientRequest{
		Path: fmt.Sprintf("/properties/%s/%d/%s/%d/%s/%s",
			pr.Pipeline, pr.PipelineCounter,
			pr.Stage, pr.StageCounter,
			pr.Job, name,
		),
		ResponseType: responseTypeText,
		ResponseBody: responseBuffer,
		RequestBody:  fmt.Sprintf("%s=%s", name, value),
		Headers: map[string]string{
			"Confirm": "true",
		},
	})
	resp.Body = responseBuffer.String()
	responseIsValid = resp.Body == fmt.Sprintf("Property '%s' created with value '%s'", name, value)

	return
}

// ListHistorical properties for a given pipeline, stage, job.
func (ps *PropertiesService) ListHistorical(ctx context.Context, pr *PropertyRequest) (*Properties, *APIResponse, error) {
	u := ps.client.BaseURL()
	q := u.Query()
	q.Set("pipelineName", pr.Pipeline)
	q.Set("stageName", pr.Stage)
	q.Set("jobName", pr.Job)
	if pr.Limit >= 0 && pr.LimitPipeline != "" {
		q.Set("limitCount", fmt.Sprintf("%d", pr.Limit))
		q.Set("limitPipeline", pr.LimitPipeline)
	}
	u.RawQuery = q.Encode()
	return ps.commonPropertiesAction(ctx, "/properties/search", false)
}

func (ps *PropertiesService) commonPropertiesAction(ctx context.Context, path string, isDatum bool) (p *Properties, resp *APIResponse, err error) {
	p = &Properties{
		UnmarshallWithHeader: true,
		IsDatum:              isDatum,
	}
	_, resp, err = ps.client.getAction(ctx, &APIClientRequest{
		Path:         path,
		ResponseBody: p,
	})

	return
}
