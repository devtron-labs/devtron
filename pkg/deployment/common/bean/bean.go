package bean

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"log"
	"strconv"
	"strings"
)

type Destination struct {
	Namespace string `json:"namespace,omitempty"` // deployed application namespace
	Server    string `json:"server,omitempty"`    // deployed application cluster url
}

type Source struct {
	RepoURL        string `json:"repoURL,omitempty"`
	ChartPath      string `json:"chartPath,omitempty"`
	ValuesFilePath string `json:"valuesFilePath,omitempty"`
	TargetRevision string `json:"targetRevision,omitempty"` //target branch
}

type SyncPolicy struct {
	SyncPolicy string `json:"syncPolicy,omitempty"`
}

type ArgoCDSpec struct {
	ClusterId   int          `json:"clusterId,omitempty"` // Application object cluster
	Namespace   string       `json:"namespace,omitempty"` // Application object namespace
	Destination *Destination `json:"destination,omitempty"`
	Source      *Source      `json:"source,omitempty"`
	SyncPolicy  *SyncPolicy  `json:"syncPolicy,omitempty"`
}

type ReleaseConfiguration struct {
	ArgoCDSpec ArgoCDSpec `json:"argoCDSpec"`
}

func (r *ReleaseConfiguration) JSON() []byte {
	if r == nil {
		return nil
	}
	releaseConfigJson, err := json.Marshal(r)
	if err != nil {
		log.Print("error in marshaling releaseConfiguration", "err")
		return nil
	}
	return releaseConfigJson
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

func (d *DeploymentConfig) IsArgoCdClientSupported() bool {
	return d.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_ACD && d.ReleaseMode != util.PIPELINE_RELEASE_MODE_LINK
}

func (d *DeploymentConfig) IsArgoAppSyncAndRefreshSupported() bool {
	return d.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_ACD && d.ReleaseMode != util.PIPELINE_RELEASE_MODE_LINK
}

func (d *DeploymentConfig) IsArgoAppPatchSupported() bool {
	return d.DeploymentAppType == util.PIPELINE_DEPLOYMENT_TYPE_ACD && d.ReleaseMode != util.PIPELINE_RELEASE_MODE_LINK
}

func (d *DeploymentConfig) IsArgoAppCreationRequired(deploymentAppCreated bool) bool {
	if d.DeploymentAppType != util.PIPELINE_DEPLOYMENT_TYPE_ACD {
		return false
	}
	if deploymentAppCreated {
		return false
	}
	if d.ReleaseMode == util.PIPELINE_RELEASE_MODE_LINK {
		return false
	}
	return true
}

func (d *DeploymentConfig) IsEmpty() bool {
	return d == nil || d.Id == 0
}

func (d *DeploymentConfig) GetRepoURL() string {
	return d.ReleaseConfiguration.ArgoCDSpec.Source.RepoURL
}

func (d *DeploymentConfig) GetTargetRevision() string {
	return d.ReleaseConfiguration.ArgoCDSpec.Source.TargetRevision
}

func (d *DeploymentConfig) GetValuesFilePath() string {
	return d.ReleaseConfiguration.ArgoCDSpec.Source.ValuesFilePath
}

func (d *DeploymentConfig) GetChartLocation() string {
	return d.ReleaseConfiguration.ArgoCDSpec.Source.ChartPath
}

func (d *DeploymentConfig) SetRepoURL(repoURL string) {
	d.ReleaseConfiguration.ArgoCDSpec.Source.RepoURL = repoURL
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
