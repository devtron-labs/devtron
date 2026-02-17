package bean

import (
	"time"

	workflowConstants "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/constants"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/helper/parser"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository/bean"
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
	ImageScanDeployInfoId  int            `json:"imageScanDeployInfoId"`
	AppId                  int            `json:"appId"`
	EnvId                  int            `json:"envId"`
	Name                   string         `json:"name"`
	Type                   string         `json:"type"`
	Environment            string         `json:"environment"`
	LastChecked            *time.Time     `json:"lastChecked"`
	Image                  string         `json:"image,omitempty"`
	SeverityCount          *SeverityCount `json:"severityCount,omitempty"`
	FixableVulnerabilities int            `json:"fixableVulnerabilities"`
	ScanStatus             string         `json:"scanStatus,omitempty"` // "scanned" or "not-scanned"
}

// VulnerabilitySummary represents the summary of all vulnerabilities across all scanned apps/envs
type VulnerabilitySummary struct {
	TotalVulnerabilities      int            `json:"totalVulnerabilities"`
	SeverityCount             *SeverityCount `json:"severityCount"`
	FixableVulnerabilities    int            `json:"fixableVulnerabilities"`
	NotFixableVulnerabilities int            `json:"notFixableVulnerabilities"`
}

// VulnerabilitySummaryRequest represents the request for vulnerability summary with filters
// Same filters as VulnerabilityListingRequest (except pagination and sorting)
type VulnerabilitySummaryRequest struct {
	EnvironmentIds  []int                  `json:"envIds"`          // Filter by environment IDs
	ClusterIds      []int                  `json:"clusterIds"`      // Filter by cluster IDs
	AppIds          []int                  `json:"appIds"`          // Filter by application IDs
	Severity        []int                  `json:"severity"`        // Filter by severity
	FixAvailability []FixAvailabilityType  `json:"fixAvailability"` // Filter by fix availability (multi-select: fixAvailable, fixNotAvailable)
	AgeOfDiscovery  []VulnerabilityAgeType `json:"ageOfDiscovery"`  // Filter by vulnerability age (multi-select)
}

// VulnerabilityListingRequest represents the request for vulnerability listing with filters
type VulnerabilityListingRequest struct {
	CVEName         string                 `json:"cveName"`         // Search by CVE name
	Severity        []int                  `json:"severity"`        // Filter by severity
	EnvironmentIds  []int                  `json:"envIds"`          // Filter by environment IDs
	ClusterIds      []int                  `json:"clusterIds"`      // Filter by cluster IDs
	AppIds          []int                  `json:"appIds"`          // Filter by application IDs
	FixAvailability []FixAvailabilityType  `json:"fixAvailability"` // Filter by fix availability (multi-select: fixAvailable, fixNotAvailable)
	AgeOfDiscovery  []VulnerabilityAgeType `json:"ageOfDiscovery"`  // Filter by vulnerability age (multi-select)
	SortBy          VulnerabilitySortBy    `json:"sortBy"`          // Sort by field
	SortOrder       SortOrder              `json:"sortOrder"`       // Sort order (ASC/DESC)
	Offset          int                    `json:"offset"`          // Pagination offset
	Size            int                    `json:"size"`            // Pagination size
}

// FixAvailabilityType represents fix availability filter options
type FixAvailabilityType string

const (
	FixAvailable    FixAvailabilityType = "fixAvailable"    // CVEs with fixes available
	FixNotAvailable FixAvailabilityType = "fixNotAvailable" // CVEs without fixes
)

// VulnerabilityAgeType represents vulnerability age filter
type VulnerabilityAgeType string

const (
	VulnAgeLessThan30Days VulnerabilityAgeType = "lt_30d" // Less than 30 days old
	VulnAge30To60Days     VulnerabilityAgeType = "30_60d" // 30 to 60 days old
	VulnAge60To90Days     VulnerabilityAgeType = "60_90d" // 60 to 90 days old
	VulnAgeMoreThan90Days VulnerabilityAgeType = "gt_90d" // More than 90 days old
)

// VulnerabilitySortBy represents sort field for vulnerability listing
type VulnerabilitySortBy string

const (
	VulnSortByCveName        VulnerabilitySortBy = "cveName"
	VulnSortByCurrentVersion VulnerabilitySortBy = "currentVersion"
	VulnSortByFixedVersion   VulnerabilitySortBy = "fixedVersion"
	VulnSortByDiscoveredAt   VulnerabilitySortBy = "discoveredAt"
	VulnSortBySeverity       VulnerabilitySortBy = "severity"
)

// SortOrder represents sort order
// Type alias to repository constants to avoid circular imports
type SortOrder = workflowConstants.SortOrder

const (
	SortOrderAsc  = workflowConstants.SortOrderAsc
	SortOrderDesc = workflowConstants.SortOrderDesc
)

// VulnerabilityListingResponse represents the response for vulnerability listing
type VulnerabilityListingResponse struct {
	Offset          int                    `json:"offset"`
	Size            int                    `json:"size"`
	Total           int                    `json:"total"`
	Vulnerabilities []*VulnerabilityDetail `json:"list"`
}

// VulnerabilityDetail represents detailed information about a single CVE
type VulnerabilityDetail struct {
	CVEName        string    `json:"cveName"`
	Severity       string    `json:"severity"`
	AppName        string    `json:"appName"`
	AppId          int       `json:"appId"`
	EnvName        string    `json:"envName"`
	EnvId          int       `json:"envId"`
	DiscoveredAt   time.Time `json:"discoveredAt"`   // First time this CVE was discovered
	Package        string    `json:"package"`        // Vulnerable package name
	CurrentVersion string    `json:"currentVersion"` // Current vulnerable version
	FixedVersion   string    `json:"fixedVersion"`   // Fixed version (empty if not fixable)
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
	ScanToolUrl           string                               `json:"scanToolUrl,omitempty"`
	Status                repository.ScanExecutionProcessState `json:"status,omitempty"`
}

type AppEnvMetadata struct {
	AppId int
	EnvId int
}

func NewAppEnvMetadata(appId, envId int) AppEnvMetadata {
	return AppEnvMetadata{AppId: appId, EnvId: envId}
}
