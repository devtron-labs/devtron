package scanningResultsParser

import (
	"fmt"
	"github.com/tidwall/gjson"
	"time"
)

type Severity string

const (
	LOW      Severity = "Low"
	MEDIUM   Severity = "Medium"
	HIGH     Severity = "High"
	CRITICAL Severity = "Critical"
	UNKNOWN  Severity = "Unknown"
)

type Summary struct {
	Severities map[Severity]int `json:"severities"`
}

func (summary *Summary) String() string {
	return fmt.Sprintf("%d Critical, %d High, %d Medium, %d Low, %d Unknown", summary.Severities[CRITICAL], summary.Severities[HIGH], summary.Severities[MEDIUM], summary.Severities[LOW], summary.Severities[UNKNOWN])
}

type Licenses struct {
	Summary  Summary   `json:"summary"`
	Licenses []License `json:"list"`
}

type License struct {
	Classification string   `json:"classification"` // Category
	Severity       Severity `json:"severity"`       // Severity
	License        string   `json:"license"`        // Name
	Package        string   `json:"package"`        // PkgName
	Source         string   `json:"source"`         // FilePath
}

func getLicense(licenseJson string) License {
	return License{
		Classification: gjson.Get(licenseJson, "Category").String(),
		Severity:       Severity(gjson.Get(licenseJson, "Severity").String()),
		License:        gjson.Get(licenseJson, "Name").String(),
		Package:        gjson.Get(licenseJson, "PkgName").String(),
		Source:         gjson.Get(licenseJson, "FilePath").String(),
	}
}

type Vulnerabilities struct {
	Summary         Summary         `json:"summary"`
	Vulnerabilities []Vulnerability `json:"list"`
}

type Vulnerability struct {
	CVEId          string   `json:"cveId"`          // VulnerabilityID
	Severity       Severity `json:"severity"`       // Severity
	Package        string   `json:"package"`        // PkgName
	CurrentVersion string   `json:"currentVersion"` // InstalledVersion
	FixedInVersion string   `json:"fixedInVersion"` // FixedVersion
}

type MisConfigurationSummary struct {
	success    int64
	fail       int64
	exceptions int64
	Severities map[string]int64 `json:"status"`
}

func (summary *MisConfigurationSummary) load() {
	severities := map[string]int64{
		"success":    summary.success,
		"fail":       summary.fail,
		"exceptions": summary.exceptions,
	}
	summary.Severities = severities
}

func (summary *MisConfigurationSummary) String() string {
	return fmt.Sprintf("%d Successes, %d Failures, %d Exceptions", summary.success, summary.fail, summary.exceptions)
}

type Line struct {
	Number    int64  `json:"number"`    // Number
	Content   string `json:"content"`   // Content
	IsCause   bool   `json:"isCause"`   // IsCause
	Truncated bool   `json:"truncated"` // Truncated
}

type CauseMetadata struct {
	StartLine int64  `json:"startLine"` // StartLine
	EndLine   int64  `json:"EndLine"`   // EndLine
	Lines     []Line `json:"lines"`     // Code.Lines
}

type Configuration struct {
	Id            string        `json:"id"`            // ID
	Title         string        `json:"title"`         // Title
	Message       string        `json:"message"`       // Message
	Resolution    string        `json:"resolution"`    // Resolution
	Status        string        `json:"status"`        // Status
	Severity      Severity      `json:"severity"`      // Severity
	CauseMetadata CauseMetadata `json:"causeMetadata"` // CauseMetadata
}

type MisConfiguration struct {
	FilePath       string                  `json:"filePath"`       // Target
	Type           string                  `json:"type"`           // Type
	MisConfSummary MisConfigurationSummary `json:"misConfSummary"` // MisConfSummary
	Summary        Summary                 `json:"summary"`
	Configurations []Configuration         `json:"list"`
}

type Secret struct {
	Severity Severity `json:"severity"`
	RuleId   string   `json:"ruleId"`
	CauseMetadata
}

type ExposedSecret struct {
	FilePath string   `json:"filePath"` // target and class: secret
	Summary  Summary  `json:"summary"`
	Secrets  []Secret `json:"list"`
}

type MisConfigurations struct {
	Summary           MisConfigurationSummary `json:"misConfSummary"`
	MisConfigurations []*MisConfiguration     `json:"list"`
}

type ExposedSecrets struct {
	Summary        Summary          `json:"summary"`
	ExposedSecrets []*ExposedSecret `json:"list"`
}

type ImageScanResult struct {
	Vulnerability *Vulnerabilities `json:"vulnerability"`
	License       *Licenses        `json:"license"`
}

type CodeScanResponse struct {
	Vulnerability     *Vulnerabilities   `json:"vulnerability"`
	License           *Licenses          `json:"license"`
	MisConfigurations *MisConfigurations `json:"misConfigurations"`
	ExposedSecrets    *ExposedSecrets    `json:"exposedSecrets"`
	Metadata
}

type K8sManifestScanResponse struct {
	MisConfigurations *MisConfigurations `json:"misConfigurations"`
	ExposedSecrets    *ExposedSecrets    `json:"exposedSecrets"`
	Metadata
}

type Metadata struct {
	Status       string    `json:"status"`
	StartedOn    time.Time `json:"StartedOn"`
	ScanToolName string    `json:"scanToolName"`
}

type VulnerabilityResponse struct {
	Summary Summary              `json:"summary"`
	List    []ImageVulnerability `json:"list"`
}

func (vr *VulnerabilityResponse) append(iv ImageVulnerability) {
	vr.List = append(vr.List, iv)
	summary := iv.Summary
	for key, val := range summary.Severities {
		summary.Severities[key] = val
		if vr.Summary.Severities != nil {
			summary.Severities[key] += vr.Summary.Severities[key]
		}
	}
	vr.Summary = summary
}

type LicenseResponse struct {
	Summary Summary         `json:"summary"`
	List    []ImageLicenses `json:"list"`
}

func (lr *LicenseResponse) append(li ImageLicenses) {
	lr.List = append(lr.List, li)
	summary := li.Summary
	for key, val := range summary.Severities {
		summary.Severities[key] = val
		if lr.Summary.Severities != nil {
			summary.Severities[key] += lr.Summary.Severities[key]
		}
	}
	lr.Summary = summary
}

type ImageVulnerability struct {
	Image string `json:"image"`
	Vulnerabilities
	Metadata
}

type ImageLicenses struct {
	Image string `json:"image"`
	Licenses
	Metadata
}

type ImageScanResponse struct {
	Vulnerability VulnerabilityResponse `json:"vulnerability"`
	License       LicenseResponse       `json:"license"`
}

type Response struct {
	Scanned            bool                    `json:"scanned"`
	ImageScan          ImageScanResponse       `json:"imageScan"`
	CodeScan           CodeScanResponse        `json:"codeScan"`
	KubernetesManifest K8sManifestScanResponse `json:"kubernetesManifest"`
}
