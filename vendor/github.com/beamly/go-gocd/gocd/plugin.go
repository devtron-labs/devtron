package gocd

import (
	"context"
	"fmt"
)

// PluginsService exposes calls for interacting with Plugin objects in the GoCD API.
type PluginsService service

// PluginsResponse describes the response obejct for a plugin API call.
type PluginsResponse struct {
	Links    *HALLinks `json:"_links"`
	Embedded struct {
		PluginInfo []*Plugin `json:"plugin_info"`
	} `json:"_embedded"`
}

// Plugin describes a single plugin resource.
// codebeat:disable[TOO_MANY_IVARS]
type Plugin struct {
	Links                     *HALLinks                 `json:"_links"`
	ID                        string                    `json:"id"`
	Name                      string                    `json:"name,omitempty"`                        // Name is available for the plugin API v1 and v2 only (GoCD >= 16.7.0 to < 17.9.0).
	DisplayName               string                    `json:"display_name,omitempty"`                // DisplayName is available for the plugin API v1 and v2 only (GoCD >= 16.7.0 to < 17.9.0).
	Version                   string                    `json:"version,omitempty"`                     // Version is available for the plugin API v1 and v2 only (GoCD >= 16.7.0 to < 17.9.0).
	Type                      string                    `json:"type,omitempty"`                        // Type is available for the plugin API v1, v2 and v3 only (GoCD >= 16.7.0 to < 18.3.0). Can be one of `authentication`, `notification`, `package-repository`, `task`, `scm`.
	PluggableInstanceSettings PluggableInstanceSettings `json:"pluggable_instance_settings,omitempty"` // PluggableInstanceSettings is available for the plugin API v1 and v2 only (GoCD >= 16.7.0 to < 17.9.0).
	Image                     PluginIcon                `json:"image,omitempty"`                       // Image is available for the plugin API v2 only (GoCD >= 16.12.0 to < 17.9.0).
	Status                    PluginStatus              `json:"status,omitempty"`                      // Status is available for the plugin API v3 and v4 (GoCD >= 17.9.0).
	PluginFileLocation        string                    `json:"plugin_file_location,omitempty"`        // PluginFileLocation is available for the plugin API v3 and v4 (GoCD >= 17.9.0).
	BundledPlugin             bool                      `json:"bundled_plugin,omitempty"`              // BundledPlugin is available for the plugin API v3 and v4 (GoCD >= 17.9.0).
	About                     PluginAbout               `json:"about,omitempty"`                       // About is available for the plugin API v3 and v4 (GoCD >= 17.9.0).
	ExtensionInfo             *PluginExtensionInfo      `json:"extension_info,omitempty"`              // ExtensionInfo is available for the plugin API v3 only (GoCD >= 17.9.0 to < 18.3.0).
	Extensions                []*PluginExtension        `json:"extensions,omitempty"`                  //Extensions is available for the plugin API v4 (GoCD >= 18.3.0).
}

// codebeat:enable[TOO_MANY_IVARS]

// PluginIcon describes the content type of the plugin icon and the base-64 encoded byte array of the byte-sequence that
// composes the image. It is used for the plugin API v2 only (GoCD >= 16.12.0 to < 17.9.0).
type PluginIcon struct {
	ContentType string `json:"content_type"`
	Data        string `json:"data"`
}

// PluggableInstanceSettings describes plugin configuration.
type PluggableInstanceSettings struct {
	Configurations []PluginConfiguration `json:"configurations"`
	View           PluginView            `json:"view"`
}

// PluginStatus describes the status of a plugin.
type PluginStatus struct {
	// State can be one of `active`, `invalid`.
	State    string   `json:"state"`
	Messages []string `json:"messages,omitempty"`
}

// PluginAbout provides additional details about the plugin.
type PluginAbout struct {
	Name                   string       `json:"name"`
	Version                string       `json:"version,omitempty"`
	TargetGoVersion        string       `json:"target_go_version,omitempty"`
	Description            string       `json:"description,omitempty"`
	TargetOperatingSystems []string     `json:"target_operating_systems,omitempty"`
	Vendor                 PluginVendor `json:"vendor,omitempty"`
}

// PluginVendor describes the author of a plugin.
type PluginVendor struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// PluginExtensionInfo describes the extension info for the plugin API v3 only (GoCD >= 17.9.0 to < 18.3.0).
// codebeat:disable[TOO_MANY_IVARS]
type PluginExtensionInfo struct {
	PluginSettings                     PluggableInstanceSettings `json:"plugin_settings,omitempty"`
	ProfileSettings                    PluggableInstanceSettings `json:"profile_settings,omitempty"`
	Capabilities                       ExtensionCapabilities     `json:"capabilities,omitempty"`
	AuthConfigSettings                 PluggableInstanceSettings `json:"auth_config_settings,omitempty"`
	RoleSettings                       PluggableInstanceSettings `json:"role_settings,omitempty"`
	DisplayName                        string                    `json:"display_name,omitempty"`
	ScmSettings                        PluggableInstanceSettings `json:"scm_settings,omitempty"`
	TaskSettings                       PluggableInstanceSettings `json:"task_settings,omitempty"`
	PackageSettings                    PluggableInstanceSettings `json:"package_settings,omitempty"`
	RepositorySettings                 PluggableInstanceSettings `json:"repository_settings,omitempty"`
	DisplayImageURL                    string                    `json:"display_image_url,omitempty"`
	SupportPasswordBasedAuthentication bool                      `json:"supports_password_based_authentication"`
	SupportWebBasedAuthentication      bool                      `json:"supports_web_based_authentication"`
}

// codebeat:enable[TOO_MANY_IVARS]

// PluginExtension describes the different extensions available for a plugin. It is used for the plugin API v4 (GoCD >= 18.3.0).
// codebeat:disable[TOO_MANY_IVARS]
type PluginExtension struct {
	Type               string                `json:"type,omitempty"`
	PluginSettings     ExtensionSettings     `json:"plugin_settings,omitempty"`
	ProfileSettings    ExtensionSettings     `json:"profile_settings,omitempty"`
	Capabilities       ExtensionCapabilities `json:"capabilities,omitempty"`
	AuthConfigSettings ExtensionSettings     `json:"auth_config_settings,omitempty"`
	RoleSettings       ExtensionSettings     `json:"role_settings,omitempty"`
	DisplayName        string                `json:"display_name,omitempty"`
	ScmSettings        ExtensionSettings     `json:"scm_settings,omitempty"`
	TaskSettings       ExtensionSettings     `json:"task_settings,omitempty"`
	PackageSettings    ExtensionSettings     `json:"package_settings,omitempty"`
	RepositorySettings ExtensionSettings     `json:"repository_settings,omitempty"`
}

// codebeat:enable[TOO_MANY_IVARS]

// ExtensionCapabilities describes the enhancements that the plugin provides.
// codebeat:disable[TOO_MANY_IVARS]
type ExtensionCapabilities struct {
	SupportStatusReport      bool   `json:"supports_status_report,omitempty"`
	SupportAgentStatusReport bool   `json:"supports_agent_status_report,omitempty"`
	CanSearch                bool   `json:"can_search,omitempty"`
	SupportedAuthType        string `json:"supported_auth_type,omitempty"`
	CanAuthorize             bool   `json:"can_authorize,omitempty"`
	ID                       string `json:"id,omitempty"`
	Title                    string `json:"title,omitempty"`
	Type                     string `json:"type,omitempty"`
}

// codebeat:enable[TOO_MANY_IVARS]

// ExtensionSettings describes the html view for the plugin and the list of properties required to be configured on a plugin.
type ExtensionSettings struct {
	View           PluginView             `json:"view,omitempty"`
	Configurations []*PluginConfiguration `json:"configurations,omitempty"`
}

// PluginConfiguration describes the configuration related to a plugin extension.
type PluginConfiguration struct {
	Key      string                      `json:"key"`
	Metadata PluginConfigurationMetadata `json:"metadata"`
}

type PluginConfigurationMetadata struct {
	Secure         bool   `json:"secure"`
	Required       bool   `json:"required"`
	PartOfIdentity bool   `json:"part_of_identity,omitempty"`
	DisplayOrder   int    `json:"display_order,omitempty"`
	DisplayName    string `json:"display_name,omitempty"`
}

// PluginView describes any view attached to a plugin.
type PluginView struct {
	Template string `json:"template"`
}

// List retrieves all plugins
func (ps *PluginsService) List(ctx context.Context) (*PluginsResponse, *APIResponse, error) {
	apiVersion, err := ps.client.getAPIVersion(ctx, "admin/plugin_info")
	if err != nil {
		return nil, nil, err
	}
	pr := PluginsResponse{}
	_, resp, err := ps.client.getAction(ctx, &APIClientRequest{
		Path:         "admin/plugin_info",
		ResponseBody: &pr,
		APIVersion:   apiVersion,
	})

	return &pr, resp, err
}

// Get retrieves information about a specific plugin.
func (ps *PluginsService) Get(ctx context.Context, name string) (p *Plugin, resp *APIResponse, err error) {
	apiVersion, err := ps.client.getAPIVersion(ctx, "admin/plugin_info")
	if err != nil {
		return nil, nil, err
	}
	p = &Plugin{}
	_, resp, err = ps.client.getAction(ctx, &APIClientRequest{
		Path:         fmt.Sprintf("admin/plugin_info/%s", name),
		ResponseBody: p,
		APIVersion:   apiVersion,
	})

	return
}
