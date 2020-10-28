package gocd

import (
	"context"
)

// ConfigurationService describes the HAL _link resource for the api response object for a pipelineconfig
type ConfigurationService service

// ConfigXML part of cruise-control.xml. @TODO better documentation
type ConfigXML struct {
	Repositories       []ConfigMaterialRepository `xml:"repositories>repository"`
	Server             ConfigServer               `xml:"server"`
	SCMS               []ConfigSCM                `xml:"scms>scm"`
	ConfigRepositories []ConfigRepository         `xml:"config-repos>config-repo"`
	PipelineGroups     []ConfigPipelineGroup      `xml:"pipelines"`
}

// ConfigPipelineGroup contains a single pipeline groups
type ConfigPipelineGroup struct {
	Name      string           `xml:"group,attr"`
	Pipelines []ConfigPipeline `xml:"pipeline"`
}

// ConfigPipeline part of cruise-control.xml. @TODO better documentation
// codebeat:disable[TOO_MANY_IVARS]
type ConfigPipeline struct {
	Name                 string                      `xml:"name,attr"`
	LabelTemplate        string                      `xml:"labeltemplate,attr"`
	Params               []ConfigParam               `xml:"params>param"`
	GitMaterials         []GitRepositoryMaterial     `xml:"materials>git,omitempty"`
	PipelineMaterials    []PipelineMaterial          `xml:"materials>pipeline,omitempty"`
	Timer                string                      `xml:"timer"`
	EnvironmentVariables []ConfigEnvironmentVariable `xml:"environmentvariables>variable"`
	Stages               []ConfigStage               `xml:"stage"`
}

// codebeat:enable[TOO_MANY_IVARS]

// ConfigStage part of cruise-control.xml. @TODO better documentation
type ConfigStage struct {
	Name     string         `xml:"name,attr"`
	Approval ConfigApproval `xml:"approval,omitempty" json:",omitempty"`
	Jobs     []ConfigJob    `xml:"jobs>job"`
}

// ConfigJob part of cruise-control.xml. @TODO better documentation
type ConfigJob struct {
	Name                 string                      `xml:"name,attr"`
	EnvironmentVariables []ConfigEnvironmentVariable `xml:"environmentvariables>variable" json:",omitempty"`
	Tasks                ConfigTasks                 `xml:"tasks"`
	Resources            []string                    `xml:"resources>resource" json:",omitempty"`
	Artifacts            []ConfigArtifact            `xml:"artifacts>artifact" json:",omitempty"`
}

// ConfigArtifact part of cruise-control.xml. @TODO better documentation
type ConfigArtifact struct {
	Src         string `xml:"src,attr"`
	Destination string `xml:"dest,attr,omitempty" json:",omitempty"`
}

// ConfigApproval part of cruise-control.xml. @TODO better documentation
type ConfigApproval struct {
	Type string `xml:"type,attr,omitempty" json:",omitempty"`
}

// ConfigEnvironmentVariable part of cruise-control.xml. @TODO better documentation
type ConfigEnvironmentVariable struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value"`
}

// PipelineMaterial part of cruise-control.xml. @TODO better documentation
type PipelineMaterial struct {
	Name         string `xml:"pipelineName,attr"`
	StageName    string `xml:"stageName,attr"`
	MaterialName string `xml:"materialName,attr"`
}

// GitRepositoryMaterial part of cruise-control.xml. @TODO better documentation
type GitRepositoryMaterial struct {
	URL     string         `xml:"url,attr"`
	Filters []ConfigFilter `xml:"filter>ignore,omitempty"`
}

// ConfigFilter part of cruise-control.xml. @TODO better documentation
type ConfigFilter struct {
	Ignore string `xml:"pattern,attr,omitempty"`
}

// ConfigParam part of cruise-control.xml. @TODO better documentation
type ConfigParam struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

// ConfigRepository part of cruise-control.xml. @TODO better documentation
type ConfigRepository struct {
	Plugin string              `xml:"plugin,attr"`
	ID     string              `xml:"id,attr"`
	Git    ConfigRepositoryGit `xml:"git"`
}

// ConfigRepositoryGit part of cruise-control.xml. @TODO better documentation
type ConfigRepositoryGit struct {
	URL string `xml:"url,attr"`
}

// ConfigSCM part of cruise-control.xml. @TODO better documentation
type ConfigSCM struct {
	ID                  string                    `xml:"id,attr"`
	Name                string                    `xml:"name,attr"`
	PluginConfiguration ConfigPluginConfiguration `xml:"pluginConfiguration"`
	Configuration       []ConfigProperty          `xml:"configuration>property"`
}

// ConfigMaterialRepository part of cruise-control.xml. @TODO better documentation
type ConfigMaterialRepository struct {
	ID                  string                    `xml:"id,attr"`
	Name                string                    `xml:"name,attr"`
	PluginConfiguration ConfigPluginConfiguration `xml:"pluginConfiguration"`
	Configuration       []ConfigProperty          `xml:"configuration>property"`
	Packages            []ConfigPackage           `xml:"packages>package"`
}

// ConfigPackage part of cruise-control.xml. @TODO better documentation
type ConfigPackage struct {
	ID            string           `xml:"id,attr"`
	Name          string           `xml:"name,attr"`
	Configuration []ConfigProperty `xml:"configuration>property"`
}

// ConfigPluginConfiguration part of cruise-control.xml. @TODO better documentation
type ConfigPluginConfiguration struct {
	ID      string `xml:"id,attr"`
	Version string `xml:"version,attr"`
}

// ConfigServer part of cruise-control.xml. @TODO better documentation
// codebeat:disable[TOO_MANY_IVARS]
type ConfigServer struct {
	MailHost                  MailHost       `xml:"mailhost"`
	Security                  ConfigSecurity `xml:"security"`
	Elastic                   ConfigElastic  `xml:"elastic"`
	ArtifactsDir              string         `xml:"artifactsdir,attr"`
	SiteURL                   string         `xml:"siteUrl,attr"`
	SecureSiteURL             string         `xml:"secureSiteUrl,attr"`
	PurgeStart                string         `xml:"purgeStart,attr"`
	PurgeUpTo                 string         `xml:"purgeUpto,attr"`
	JobTimeout                int            `xml:"jobTimeout,attr"`
	AgentAutoRegisterKey      string         `xml:"agentAutoRegisterKey,attr"`
	WebhookSecret             string         `xml:"webhookSecret,attr"`
	CommandRepositoryLocation string         `xml:"commandRepositoryLocation,attr"`
	ServerID                  string         `xml:"serverId,attr"`
}

// codebeat:enable[TOO_MANY_IVARS]

// MailHost part of cruise-control.xml. @TODO better documentation
type MailHost struct {
	Hostname string `xml:"hostname,attr"`
	Port     int    `xml:"port,attr"`
	TLS      bool   `xml:"tls,attr"`
	From     string `xml:"from,attr"`
	Admin    string `xml:"admin,attr"`
}

// ConfigSecurity part of cruise-control.xml. @TODO better documentation
type ConfigSecurity struct {
	AuthConfigs  []ConfigAuthConfig `xml:"authConfigs>authConfig"`
	Roles        []ConfigRole       `xml:"roles>role"`
	Admins       []string           `xml:"admins>user"`
	PasswordFile PasswordFilePath   `xml:"passwordFile"`
}

// PasswordFilePath describes the location to set of user/passwords on disk
type PasswordFilePath struct {
	Path string `xml:"path,attr"`
}

// ConfigRole part of cruise-control.xml. @TODO better documentation
type ConfigRole struct {
	Name  string   `xml:"name,attr"`
	Users []string `xml:"users>user"`
}

// ConfigAuthConfig part of cruise-control.xml. @TODO better documentation
type ConfigAuthConfig struct {
	ID         string           `xml:"id,attr"`
	PluginID   string           `xml:"pluginId,attr"`
	Properties []ConfigProperty `xml:"property"`
}

// ConfigElastic part of cruise-control.xml. @TODO better documentation
type ConfigElastic struct {
	Profiles []ConfigElasticProfile `xml:"profiles>profile"`
}

// ConfigElasticProfile part of cruise-control.xml. @TODO better documentation
type ConfigElasticProfile struct {
	ID         string           `xml:"id,attr"`
	PluginID   string           `xml:"pluginId,attr"`
	Properties []ConfigProperty `xml:"property"`
}

// ConfigProperty part of cruise-control.xml. @TODO better documentation
type ConfigProperty struct {
	Key   string `xml:"key"`
	Value string `xml:"value"`
}

// Version part of cruise-control.xml. @TODO better documentation
type Version struct {
	Links       *HALLinks `json:"_links"`
	Version     string    `json:"version"`
	BuildNumber string    `json:"build_number"`
	GitSHA      string    `json:"git_sha"`
	FullVersion string    `json:"full_version"`
	CommitURL   string    `json:"commit_url"`
}

// Get the config.xml document from the server and... render it as JSON... 'cause... eyugh.
func (cs *ConfigurationService) Get(ctx context.Context) (cx *ConfigXML, resp *APIResponse, err error) {
	cx = &ConfigXML{}
	_, resp, err = cs.client.getAction(ctx, &APIClientRequest{
		Path:         "admin/config.xml",
		ResponseBody: cx,
		ResponseType: responseTypeXML,
	})
	return
}

// GetVersion of the GoCD server and other metadata about the software version.
func (cs *ConfigurationService) GetVersion(ctx context.Context) (v *Version, resp *APIResponse, err error) {
	v = &Version{}
	_, resp, err = cs.client.getAction(ctx, &APIClientRequest{
		Path:         "version",
		ResponseBody: v,
		APIVersion:   apiV1,
	})
	return
}
