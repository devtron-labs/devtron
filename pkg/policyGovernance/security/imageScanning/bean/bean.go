package bean

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/helper/parser"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository/bean"
	"time"
)

const (
	BLOCK       string = "BLOCK"
	WHITELISTED        = "WHITELISTED"
)

type Vulnerabilities struct {
	CVEName    string `json:"cveName"`
	Severity   string `json:"severity"`
	Package    string `json:"package,omitempty"`
	CVersion   string `json:"currentVersion"`
	FVersion   string `json:"fixedVersion"`
	Permission string `json:"permission"`
	Target     string `json:"target"`
	Class      string `json:"class"`
	Type       string `json:"type"`
}

func (vul *Vulnerabilities) ToSeverity() parser.Severity {
	return parser.Severity(vul.Severity)
}

func (vul *Vulnerabilities) IsCritical() bool {
	return vul.Severity == bean.CRITICAL
}

func (vul *Vulnerabilities) IsHigh() bool {
	return vul.Severity == bean.HIGH
}

func (vul *Vulnerabilities) IsMedium() bool {
	return vul.Severity == bean.MODERATE || vul.Severity == bean.MEDIUM
}

func (vul *Vulnerabilities) IsLow() bool {
	return vul.Severity == bean.LOW
}

func (vul *Vulnerabilities) IsUnknown() bool {
	return vul.Severity == bean.UNKNOWN
}

type SeverityCount struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Unknown  int `json:"unknown"`
}

type ImageScanRequest struct {
	ScanExecutionId       int    `json:"ScanExecutionId"`
	ImageScanDeployInfoId int    `json:"imageScanDeployInfo"`
	AppId                 int    `json:"appId"`
	EnvId                 int    `json:"envId"`
	ObjectId              int    `json:"objectId"`
	ArtifactId            int    `json:"artifactId"`
	Image                 string `json:"image"`
	bean.ImageScanFilter
}

type ImageScanHistoryListingResponse struct {
	Offset                   int                         `json:"offset"`
	Size                     int                         `json:"size"`
	Total                    int                         `json:"total"`
	ImageScanHistoryResponse []*ImageScanHistoryResponse `json:"scanList"`
}

type ImageScanHistoryResponse struct {
	ImageScanDeployInfoId int            `json:"imageScanDeployInfoId"`
	AppId                 int            `json:"appId"`
	EnvId                 int            `json:"envId"`
	Name                  string         `json:"name"`
	Type                  string         `json:"type"`
	Environment           string         `json:"environment"`
	LastChecked           *time.Time     `json:"lastChecked"`
	Image                 string         `json:"image,omitempty"`
	SeverityCount         *SeverityCount `json:"severityCount,omitempty"`
}

type ImageScanExecutionDetail struct {
	ImageScanDeployInfoId int                                  `json:"imageScanDeployInfoId"`
	AppId                 int                                  `json:"appId,omitempty"`
	EnvId                 int                                  `json:"envId,omitempty"`
	AppName               string                               `json:"appName,omitempty"`
	EnvName               string                               `json:"envName,omitempty"`
	ArtifactId            int                                  `json:"artifactId,omitempty"`
	Image                 string                               `json:"image,omitempty"`
	PodName               string                               `json:"podName,omitempty"`
	ReplicaSet            string                               `json:"replicaSet,omitempty"`
	Vulnerabilities       []*Vulnerabilities                   `json:"vulnerabilities,omitempty"`
	SeverityCount         *SeverityCount                       `json:"severityCount,omitempty"`
	ExecutionTime         time.Time                            `json:"executionTime,omitempty"`
	ScanEnabled           bool                                 `json:"scanEnabled,notnull"`
	Scanned               bool                                 `json:"scanned,notnull"`
	ObjectType            string                               `json:"objectType,notnull"`
	ScanToolId            int                                  `json:"scanToolId,omitempty"`
	ScanToolName          string                               `json:"scanToolName,omitempty"`
	Status                repository.ScanExecutionProcessState `json:"status,omitempty"`
}

type AppEnvMetadata struct {
	AppId int
	EnvId int
}

func NewAppEnvMetadata(appId, envId int) AppEnvMetadata {
	return AppEnvMetadata{AppId: appId, EnvId: envId}
}
