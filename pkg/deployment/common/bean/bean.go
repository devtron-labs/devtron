package bean

import (
	"encoding/json"
	"fmt"
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

func (c *DeploymentConfig) GetRepoURL() string {
	return c.ReleaseConfiguration.ArgoCDSpec.Source.RepoURL
}

func (c *DeploymentConfig) GetChartLocation() string {
	return c.ReleaseConfiguration.ArgoCDSpec.Source.ChartPath
}

func (c *DeploymentConfig) SetRepoURL(repoURL string) {
	c.ReleaseConfiguration.ArgoCDSpec.Source.RepoURL = repoURL
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
