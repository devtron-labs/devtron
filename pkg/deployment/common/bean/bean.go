package bean

import (
	"fmt"
	apiGitOpsBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/internal/util"
	util2 "github.com/devtron-labs/devtron/util"
	"strconv"
	"strings"
)

type ReleaseConfigVersion string

const (
	Version ReleaseConfigVersion = "v1.0.0"
)

type ReleaseConfiguration struct {
	Version    ReleaseConfigVersion `json:"version"`
	ArgoCDSpec ArgoCDSpec           `json:"argoCDSpec"`
}

type ArgoCDSpec struct {
	Metadata ApplicationMetadata `json:"metadata"`
	Spec     ApplicationSpec     `json:"spec"`
}

func (a *ArgoCDSpec) SetApplicationObjectClusterId(clusterId int) {
	a.Metadata.ClusterId = clusterId
}

type ApplicationMetadata struct {
	ClusterId int    `json:"clusterId"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type ApplicationSpec struct {
	Destination *Destination         `json:"destination,omitempty"`
	Source      *ApplicationSource   `json:"source,omitempty"`
	SyncPolicy  *SyncPolicyAutomated `json:"syncPolicy,omitempty"`
}

type ApplicationSource struct {
	// RepoURL is the URL to the repository (Git or Helm) that contains the application manifests
	RepoURL string `json:"repoURL"`
	// Path is a directory path within the Git repository, and is only valid for applications sourced from Git.
	Path string `json:"path,omitempty"`
	// TargetRevision defines the revision of the source to sync the application to.
	// In case of Git, this can be commit, tag, or branch. If omitted, will equal to HEAD.
	// In case of Helm, this is a semver tag for the Chart's version.
	TargetRevision string `json:"targetRevision,omitempty"`
	// Helm holds helm specific options
	Helm *ApplicationSourceHelm `json:"helm,omitempty"`
	// Chart is a Helm chart name, and must be specified for applications sourced from a Helm repo.
	Chart string `json:"chart,omitempty"`
	// Ref is reference to another source within sources field. This field will not be used if used with a `source` tag.
	Ref string `json:"ref,omitempty"`
}

// ApplicationSourceHelm holds helm specific options
type ApplicationSourceHelm struct {
	// ValuesFiles is a list of Helm value files to use when generating a template
	ValueFiles []string `json:"valueFiles,omitempty"`
	// Parameters is a list of Helm parameters which are passed to the helm template command upon manifest generation
	Parameters []HelmParameter `json:"parameters,omitempty"`
	// ReleaseName is the Helm release name to use. If omitted it will use the application name
	ReleaseName string `json:"releaseName,omitempty"`
	// Values specifies Helm values to be passed to helm template, typically defined as a block
	Values string `json:"values,omitempty"`
	// FileParameters are file parameters to the helm template
	FileParameters []HelmFileParameter `json:"fileParameters,omitempty"`
	// Version is the Helm version to use for templating ("3")
	Version string `json:"version,omitempty"`
	// PassCredentials pass credentials to all domains (Helm's --pass-credentials)
	PassCredentials bool `json:"passCredentials,omitempty"`
	// IgnoreMissingValueFiles prevents helm template from failing when valueFiles do not exist locally by not appending them to helm template --values
	IgnoreMissingValueFiles bool `json:"ignoreMissingValueFiles,omitempty"`
	// SkipCrds skips custom resource definition installation step (Helm's --skip-crds)
	SkipCrds bool `json:"skipCrds,omitempty"`
}

type HelmParameter struct {
	// Name is the name of the Helm parameter
	Name string `json:"name,omitempty"`
	// Value is the value for the Helm parameter
	Value string `json:"value,omitempty"`
	// ForceString determines whether to tell Helm to interpret booleans and numbers as strings
	ForceString bool `json:"forceString,omitempty"`
}

// HelmFileParameter is a file parameter that's passed to helm template during manifest generation
type HelmFileParameter struct {
	// Name is the name of the Helm parameter
	Name string `json:"name,omitempty"`
	// Path is the path to the file containing the values for the Helm parameter
	Path string `json:"path,omitempty"`
}

type Destination struct {
	Namespace string `json:"namespace,omitempty"` // deployed application namespace
	Server    string `json:"server,omitempty"`    // deployed application cluster url
}

type Automated struct {
	Prune bool `json:"prune"`
}

type SyncPolicy struct {
	Automated                *SyncPolicyAutomated      `json:"automated,omitempty"`
	SyncOptions              SyncOptions               `json:"syncOptions,omitempty"`
	Retry                    *RetryStrategy            `json:"retry,omitempty"`
	ManagedNamespaceMetadata *ManagedNamespaceMetadata `json:"managedNamespaceMetadata,omitempty"`
}

type SyncPolicyAutomated struct {
	Prune      bool `json:"prune,omitempty"`
	SelfHeal   bool `json:"selfHeal,omitempty"`
	AllowEmpty bool `json:"allowEmpty,omitempty"`
}

type SyncOptions []string

// RetryStrategy contains information about the strategy to apply when a sync failed
type RetryStrategy struct {
	// Limit is the maximum number of attempts for retrying a failed sync. If set to 0, no retries will be performed.
	Limit int64 `json:"limit,omitempty"`
	// Backoff controls how to backoff on subsequent retries of failed syncs
	Backoff *Backoff `json:"backoff,omitempty"`
}

// Backoff is the backoff strategy to use on subsequent retries for failing syncs
type Backoff struct {
	// Duration is the amount to back off. Default unit is seconds, but could also be a duration (e.g. "2m", "1h")
	Duration string `json:"duration,omitempty"`
	// Factor is a factor to multiply the base duration after each failed retry
	Factor *int64 `json:"factor,omitempty"`
	// MaxDuration is the maximum amount of time allowed for the backoff strategy
	MaxDuration string `json:"maxDuration,omitempty"`
}

type ManagedNamespaceMetadata struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type DeploymentConfigMin struct {
	DeploymentAppType      string
	ReleaseMode            string
	GitRepoUrl             string
	IsGitOpsRepoConfigured bool
}

func (d *DeploymentConfigMin) IsLinkedRelease() bool {
	return d.ReleaseMode == util.PIPELINE_RELEASE_MODE_LINK
}

type DeploymentConfig struct {
	Id                   int
	AppId                int
	EnvironmentId        int
	ConfigType           string
	DeploymentAppType    string
	ReleaseMode          string
	RepoURL              string // DEPRECATED;
	RepoName             string
	Active               bool
	ReleaseConfiguration *ReleaseConfiguration
}

func (d *DeploymentConfig) IsAcdRelease() bool {
	return d.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_ACD
}

func (d *DeploymentConfig) IsLinkedRelease() bool {
	return d.ReleaseMode == util.PIPELINE_RELEASE_MODE_LINK
}

func (d *DeploymentConfig) IsArgoCdClientSupported() bool {
	return d.IsAcdRelease() && !d.IsLinkedRelease()
}

func (d *DeploymentConfig) IsArgoAppSyncAndRefreshSupported() bool {
	return d.IsAcdRelease() && !d.IsLinkedRelease()
}

func (d *DeploymentConfig) IsArgoAppPatchSupported() bool {
	return d.IsAcdRelease() && !d.IsLinkedRelease()
}

func (d *DeploymentConfig) IsArgoAppCreationRequired(deploymentAppCreated bool) bool {
	if !d.IsAcdRelease() {
		return false
	}
	if deploymentAppCreated {
		return false
	}
	if d.IsLinkedRelease() {
		return false
	}
	return true
}

func (d *DeploymentConfig) IsEmpty() bool {
	return d == nil || d.Id == 0
}

func (d *DeploymentConfig) IsPipelineGitOpsRepoConfigured(isAppLevelGitOpsConfigured bool) bool {
	return isAppLevelGitOpsConfigured || !apiGitOpsBean.IsGitOpsRepoNotConfigured(d.GetRepoURL())
}

func (d *DeploymentConfig) GetRepoURL() string {
	if d.ReleaseConfiguration == nil || d.ReleaseConfiguration.ArgoCDSpec.Spec.Source == nil {
		return d.RepoURL
	}
	return d.ReleaseConfiguration.ArgoCDSpec.Spec.Source.RepoURL
}

func (d *DeploymentConfig) GetTargetRevision() string {
	if d.ReleaseConfiguration == nil || d.ReleaseConfiguration.ArgoCDSpec.Spec.Source == nil {
		return util2.GetDefaultTargetRevision()
	}
	return d.ReleaseConfiguration.ArgoCDSpec.Spec.Source.TargetRevision
}

func (d *DeploymentConfig) GetValuesFilePath() string {
	if d.ReleaseConfiguration == nil || d.ReleaseConfiguration.ArgoCDSpec.Spec.Source == nil {
		return ""
	}
	// currently we only support a single value file
	if len(d.ReleaseConfiguration.ArgoCDSpec.Spec.Source.Helm.ValueFiles) != 0 {
		return d.ReleaseConfiguration.ArgoCDSpec.Spec.Source.Helm.ValueFiles[0]
	} else {
		return ""
	}
}

func (d *DeploymentConfig) GetChartLocation() string {
	if d.ReleaseConfiguration == nil || d.ReleaseConfiguration.ArgoCDSpec.Spec.Source == nil {
		return ""
	}
	return d.ReleaseConfiguration.ArgoCDSpec.Spec.Source.Path
}

func (d *DeploymentConfig) SetRepoURL(repoURL string) *DeploymentConfig {
	d.RepoURL = repoURL // maintain for backward compatibility
	if d.ReleaseConfiguration == nil || d.ReleaseConfiguration.ArgoCDSpec.Spec.Source == nil {
		return d
	}
	d.ReleaseConfiguration.ArgoCDSpec.Spec.Source.RepoURL = repoURL
	return d
}

func (d *DeploymentConfig) SetChartLocation(chartLocation string) {
	if d.ReleaseConfiguration == nil || d.ReleaseConfiguration.ArgoCDSpec.Spec.Source == nil {
		return
	}
	d.ReleaseConfiguration.ArgoCDSpec.Spec.Source.Path = chartLocation
}

func (d *DeploymentConfig) GetRevision() string {
	if d.ReleaseConfiguration == nil || d.ReleaseConfiguration.ArgoCDSpec.Spec.Source == nil {
		return ""
	}
	return d.ReleaseConfiguration.ArgoCDSpec.Spec.Source.TargetRevision
}

func (d *DeploymentConfig) GetAcdAppName() string {
	if d.ReleaseConfiguration == nil {
		return ""
	}
	return d.ReleaseConfiguration.ArgoCDSpec.Metadata.Name
}

func (d *DeploymentConfig) GetValuesFileName() string {
	if d.ReleaseConfiguration == nil || d.ReleaseConfiguration.ArgoCDSpec.Spec.Source == nil ||
		d.ReleaseConfiguration.ArgoCDSpec.Spec.Source.Helm == nil {
		return ""
	}
	return d.ReleaseConfiguration.ArgoCDSpec.Spec.Source.Helm.ValueFiles[0]
}

func (d *DeploymentConfig) GetDestinationClusterURL() string {
	if d.ReleaseConfiguration == nil || d.ReleaseConfiguration.ArgoCDSpec.Spec.Destination == nil {
		return ""
	}
	return d.ReleaseConfiguration.ArgoCDSpec.Spec.Destination.Server
}

func (d *DeploymentConfig) GetDestinationNamespace() string {
	if d.ReleaseConfiguration == nil || d.ReleaseConfiguration.ArgoCDSpec.Spec.Destination == nil {
		return ""
	}
	return d.ReleaseConfiguration.ArgoCDSpec.Spec.Destination.Namespace
}

func (d *DeploymentConfig) SetApplicationObjectClusterId(id int) {
	if d.ReleaseConfiguration == nil {
		return
	}
	d.ReleaseConfiguration.ArgoCDSpec.SetApplicationObjectClusterId(id)
}

func (d *DeploymentConfig) GetApplicationObjectClusterId() int {
	if d.ReleaseConfiguration == nil {
		return 0
	}
	return d.ReleaseConfiguration.ArgoCDSpec.Metadata.ClusterId
}

func (d *DeploymentConfig) GetApplicationObjectNamespace() string {
	if d.ReleaseConfiguration == nil {
		return ""
	}
	return d.ReleaseConfiguration.ArgoCDSpec.Metadata.Namespace
}

type UniqueDeploymentConfigIdentifier string

type DeploymentConfigSelector struct {
	AppId         int
	EnvironmentId int
	CDPipelineId  int
}

func (u UniqueDeploymentConfigIdentifier) String() string {
	return string(u)
}

func GetConfigUniqueIdentifier(appId, envId int) UniqueDeploymentConfigIdentifier {
	return UniqueDeploymentConfigIdentifier(fmt.Sprintf("%d-%d", appId, envId))

}

func (u *UniqueDeploymentConfigIdentifier) GetAppAndEnvId() (appId, envId int) {
	splitArr := strings.Split(u.String(), "-")
	appIdStr, envIdStr := splitArr[0], splitArr[1]
	appId, _ = strconv.Atoi(appIdStr)
	envId, _ = strconv.Atoi(envIdStr)
	return appId, envId
}

type DeploymentConfigType string

const (
	CUSTOM           DeploymentConfigType = "custom"
	SYSTEM_GENERATED DeploymentConfigType = "system_generated"
)

func (d DeploymentConfigType) String() string {
	return string(d)
}

type DeploymentConfigCredentialType string

const (
	GitOps DeploymentConfigCredentialType = "gitOps"
)

func (d DeploymentConfigCredentialType) String() string {
	return string(d)
}

type ExternalReleaseType string

func (e ExternalReleaseType) IsArgoApplication() bool {
	return e == ArgoApplication
}

const (
	ArgoApplication ExternalReleaseType = "argoApplication"
	HelmRelease     ExternalReleaseType = "helmRelease"
	Undefined       ExternalReleaseType = ""
)

func (d *DeploymentConfig) GetMigratedFrom() (migratedFrom ExternalReleaseType, isLinked bool) {
	if d.IsLinkedRelease() {
		if d.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_ACD {
			return ArgoApplication, true
		} else if d.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_HELM {
			return HelmRelease, true
		}
	}
	return Undefined, false
}
