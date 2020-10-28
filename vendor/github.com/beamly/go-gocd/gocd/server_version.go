package gocd

import (
	"context"
	"github.com/hashicorp/go-version"
)

// ServerVersionService exposes calls for interacting with ServerVersion objects in the GoCD API.
type ServerVersionService service

var cachedServerVersion *ServerVersion

// ServerVersion of the GoCD installation
type ServerVersion struct {
	Version      string `json:"version"`
	VersionParts *version.Version
	BuildNumber  string `json:"build_number"`
	GitSha       string `json:"git_sha"`
	FullVersion  string `json:"full_version"`
	CommitURL    string `json:"commit_url"`
}

// Get retrieves information about a specific plugin.
func (svs *ServerVersionService) Get(ctx context.Context) (v *ServerVersion, resp *APIResponse, err error) {
	if cachedServerVersion != nil {
		return cachedServerVersion, nil, nil
	}

	v = &ServerVersion{}
	if _, resp, err = svs.client.getAction(ctx, &APIClientRequest{
		Path:         "version",
		ResponseBody: v,
		APIVersion:   apiV1,
	}); err != nil {
		return
	}

	err = v.parseVersion()

	cachedServerVersion = v

	return
}
