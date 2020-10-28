package gocd

import (
	"context"
	"fmt"
)

// ConfigRepoService allows admin users to define and manage config repos using
// which pipelines defined in external repositories can be included in GoCD,
// thereby allowing users to have their Pipeline as code.
type ConfigRepoService service

// ConfigReposListResponse describes the structure of the API response when listing collections of ConfigRepo objects
type ConfigReposListResponse struct {
	Links    *HALLinks `json:"_links,omitempty"`
	Embedded *struct {
		Repos []*ConfigRepo `json:"config_repos"`
	} `json:"_embedded,omitempty"`
}

// ConfigRepo represents a config repo object in GoCD
type ConfigRepo struct {
	ID            string                `json:"id"`
	PluginID      string                `json:"plugin_id"`
	Material      Material              `json:"material"`
	Configuration []*ConfigRepoProperty `json:"configuration,omitempty"`
	Links         *HALLinks             `json:"_links,omitempty,omitempty"`
	Version       string                `json:"version,omitempty"`
	client        *Client
}

// ConfigRepoProperty represents a configuration related to a ConfigRepo
type ConfigRepoProperty struct {
	Key            string `json:"key"`
	Value          string `json:"value,omitempty"`
	EncryptedValue string `json:"encrypted_value,omitempty"`
}

// List returns all available config repos, these are config repositories that
// are present in the in `cruise-config.xml`
func (crs *ConfigRepoService) List(ctx context.Context) (repos []*ConfigRepo, resp *APIResponse, err error) {
	r := &ConfigReposListResponse{}
	_, resp, err = crs.client.getAction(ctx, &APIClientRequest{
		Path:         "admin/config_repos",
		ResponseBody: r,
		APIVersion:   apiV1,
	})

	for _, repos := range r.Embedded.Repos {
		repos.client = crs.client
	}
	repos = r.Embedded.Repos

	return
}

// Get fetches the config repo object for a specified id
func (crs *ConfigRepoService) Get(ctx context.Context, id string) (out *ConfigRepo, resp *APIResponse, err error) {
	out = &ConfigRepo{}
	_, resp, err = crs.client.getAction(ctx, &APIClientRequest{
		Path:         fmt.Sprintf("admin/config_repos/%s", id),
		ResponseBody: out,
		APIVersion:   apiV1,
	})

	out.client = crs.client
	return
}

// Create a config repo
func (crs *ConfigRepoService) Create(ctx context.Context, cr *ConfigRepo) (out *ConfigRepo, resp *APIResponse, err error) {
	out = &ConfigRepo{}
	_, resp, err = crs.client.postAction(ctx, &APIClientRequest{
		Path:         "admin/config_repos",
		RequestBody:  cr,
		ResponseBody: out,
		APIVersion:   apiV1,
	})

	out.client = crs.client
	return
}

// Update config repos for specified config repo id
func (crs *ConfigRepoService) Update(ctx context.Context, id string, cr *ConfigRepo) (out *ConfigRepo, resp *APIResponse, err error) {
	out = &ConfigRepo{}
	_, resp, err = crs.client.putAction(ctx, &APIClientRequest{
		Path:         fmt.Sprintf("admin/config_repos/%s", id),
		RequestBody:  cr,
		ResponseBody: out,
		APIVersion:   apiV1,
	})

	out.client = crs.client
	return
}

// Delete the specified config repo
func (crs *ConfigRepoService) Delete(ctx context.Context, id string) (string, *APIResponse, error) {
	return crs.client.deleteAction(ctx, fmt.Sprintf("admin/config_repos/%s", id), apiV1)
}
