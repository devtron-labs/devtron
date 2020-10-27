package gocd

import (
	"context"
	//"fmt"
)

// PipelineConfigsService describes the HAL _link resource for the api response object for a pipelineconfig
type PipelineConfigsService service

// PipelineConfigRequest describes a request object for creating or updating pipelines
type PipelineConfigRequest struct {
	Group    string    `json:"group,omitempty"`
	Pipeline *Pipeline `json:"pipeline"`
}

// Get a single Pipeline object in the GoCD API.
func (pcs *PipelineConfigsService) Get(ctx context.Context, name string) (p *Pipeline, resp *APIResponse, err error) {

	apiVersion, err := pcs.client.getAPIVersion(ctx, "admin/pipelines/:pipeline_name")
	if err != nil {
		return nil, nil, err
	}

	p = &Pipeline{}
	_, resp, err = pcs.client.getAction(ctx, &APIClientRequest{
		Path:         "admin/pipelines/" + name,
		APIVersion:   apiVersion,
		ResponseBody: p,
	})

	return
}

// Update a pipeline configuration
func (pcs *PipelineConfigsService) Update(ctx context.Context, name string, p *Pipeline) (pr *Pipeline, resp *APIResponse, err error) {

	apiVersion, err := pcs.client.getAPIVersion(ctx, "admin/pipelines/:pipeline_name")
	if err != nil {
		return nil, nil, err
	}

	pr = &Pipeline{}
	_, resp, err = pcs.client.putAction(ctx, &APIClientRequest{
		Path:         "admin/pipelines/" + name,
		APIVersion:   apiVersion,
		RequestBody:  p,
		ResponseBody: pr,
	})

	pr.Group = p.Group

	return
}

// Create a pipeline configuration
func (pcs *PipelineConfigsService) Create(ctx context.Context, group string, p *Pipeline) (pr *Pipeline, resp *APIResponse, err error) {

	apiVersion, err := pcs.client.getAPIVersion(ctx, "admin/pipelines/:pipeline_name")
	if err != nil {
		return nil, nil, err
	}

	pr = &Pipeline{}
	_, resp, err = pcs.client.postAction(ctx, &APIClientRequest{
		Path:       "admin/pipelines",
		APIVersion: apiVersion,
		RequestBody: &PipelineConfigRequest{
			Group:    group,
			Pipeline: p,
		},
		ResponseBody: pr,
	})

	pr.Group = group

	return
}

// Delete a pipeline configuration
func (pcs *PipelineConfigsService) Delete(ctx context.Context, name string) (string, *APIResponse, error) {
	apiVersion, err := pcs.client.getAPIVersion(ctx, "admin/pipelines/:pipeline_name")
	if err != nil {
		return "", nil, err
	}

	return pcs.client.deleteAction(ctx, "admin/pipelines/"+name, apiVersion)
}
