package gocd

import (
	"context"
	"fmt"
)

// PipelineTemplatesService describes the HAL _link resource for the api response object for a pipeline configuration objects.
type PipelineTemplatesService service

// PipelineTemplateRequest describes a PipelineTemplate
type PipelineTemplateRequest struct {
	Name    string   `json:"name"`
	Stages  []*Stage `json:"stages"`
	Version string   `json:"version"`
}

// PipelineTemplateResponse describes an api response for a single pipeline templates
type PipelineTemplateResponse struct {
	Name     string `json:"name"`
	Embedded *struct {
		Pipelines []*struct {
			Name string `json:"name"`
		}
	} `json:"_embedded,omitempty"`
}

// PipelineTemplatesResponse describes an api response for multiple pipeline templates
type PipelineTemplatesResponse struct {
	Links    *HALLinks `json:"_links,omitempty"`
	Embedded *struct {
		Templates []*PipelineTemplate `json:"templates"`
	} `json:"_embedded,omitempty"`
}

type embeddedPipelineTemplate struct {
	Pipelines []*Pipeline `json:"pipelines,omitempty"`
}

// PipelineTemplate describes a response from the API for a pipeline template object.
type PipelineTemplate struct {
	Links    *HALLinks                 `json:"_links,omitempty"`
	Name     string                    `json:"name"`
	Embedded *embeddedPipelineTemplate `json:"_embedded,omitempty"`
	Version  string                    `json:"template_version"`
	Stages   []*Stage                  `json:"stages,omitempty"`
}

// Get a single PipelineTemplate object in the GoCD API.
func (pts *PipelineTemplatesService) Get(ctx context.Context, name string) (pt *PipelineTemplate, resp *APIResponse, err error) {
	apiVersion, err := pts.client.getAPIVersion(ctx, "admin/templates/:template_name")
	if err != nil {
		return nil, nil, err
	}

	pt = &PipelineTemplate{}
	_, resp, err = pts.client.getAction(ctx, &APIClientRequest{
		Path:         "admin/templates/" + name,
		APIVersion:   apiVersion,
		ResponseBody: pt,
	})

	return
}

// List all PipelineTemplate objects in the GoCD API.
func (pts *PipelineTemplatesService) List(ctx context.Context) (pt []*PipelineTemplate, resp *APIResponse, err error) {
	apiVersion, err := pts.client.getAPIVersion(ctx, "admin/templates")
	if err != nil {
		return nil, nil, err
	}

	ptr := PipelineTemplatesResponse{}
	_, resp, err = pts.client.getAction(ctx, &APIClientRequest{
		Path:         "admin/templates",
		APIVersion:   apiVersion,
		ResponseBody: &ptr,
	})
	pt = ptr.Embedded.Templates

	return
}

// Create a new PipelineTemplate object in the GoCD API.
func (pts *PipelineTemplatesService) Create(ctx context.Context, name string, st []*Stage) (ptr *PipelineTemplate, resp *APIResponse, err error) {
	apiVersion, err := pts.client.getAPIVersion(ctx, "admin/templates")
	if err != nil {
		return nil, nil, err
	}

	pt := PipelineTemplateRequest{
		Name:   name,
		Stages: st,
	}
	ptr = &PipelineTemplate{}

	_, resp, err = pts.client.postAction(ctx, &APIClientRequest{
		Path:         "admin/templates",
		APIVersion:   apiVersion,
		RequestBody:  pt,
		ResponseBody: ptr,
	})

	return

}

// Update an PipelineTemplate object in the GoCD API.
func (pts *PipelineTemplatesService) Update(ctx context.Context, name string, template *PipelineTemplate) (ptr *PipelineTemplate, resp *APIResponse, err error) {
	apiVersion, err := pts.client.getAPIVersion(ctx, "admin/templates/:template_name")
	if err != nil {
		return nil, nil, err
	}

	ptr = &PipelineTemplate{}
	_, resp, err = pts.client.putAction(ctx, &APIClientRequest{
		Path:       "admin/templates/" + name,
		APIVersion: apiVersion,
		RequestBody: &PipelineTemplateRequest{
			Name:    name,
			Stages:  template.Stages,
			Version: template.Version,
		},
		ResponseBody: ptr,
	})

	return

}

// Delete a PipelineTemplate from the GoCD API.
func (pts *PipelineTemplatesService) Delete(ctx context.Context, name string) (string, *APIResponse, error) {
	apiVersion, err := pts.client.getAPIVersion(ctx, "admin/templates/:template_name")
	if err != nil {
		return "", nil, err
	}

	return pts.client.deleteAction(ctx, fmt.Sprintf("admin/templates/%s", name), apiVersion)
}
