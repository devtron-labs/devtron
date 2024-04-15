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
	ScannedOn  time.Time        `json:"startedOn"`
	Severities map[Severity]int `json:"severities"`
}

func (summary *Summary) String() string {
	return fmt.Sprintf("%d Critical, %d High, %d Medium, %d Low, %d Unknown", summary.Severities[CRITICAL], summary.Severities[HIGH], summary.Severities[MEDIUM], summary.Severities[LOW], summary.Severities[UNKNOWN])
}

type Licenses struct {
	Summary  Summary   `json:"summary"`
	Licenses []License `json:"licenses"`
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
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
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
	Severities map[string]int64 `json:"severities"`
}

func (summary *MisConfigurationSummary) load() {
	severities := map[string]int64{
		"success":    summary.success,
		"fail":       summary.success,
		"exceptions": summary.success,
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
	Configurations []Configuration         `json:"configurations"`
}

type Secret struct {
	Severity Severity `json:"severity"`
	CauseMetadata
}

type ExposedSecret struct {
	FilePath string   `json:"filePath"` // target and class: secret
	Summary  Summary  `json:"summary"`
	Secrets  []Secret `json:"secrets"`
}

type MisConfigurations struct {
	Summary           MisConfigurationSummary `json:"summary"`
	MisConfigurations []*MisConfiguration     `json:"misConfigurations"`
}

type ExposedSecrets struct {
	Summary        Summary          `json:"summary"`
	ExposedSecrets []*ExposedSecret `json:"exposedSecrets"`
}

type ImageScanResult struct {
	Vulnerability *Vulnerabilities `json:"vulnerability"`
	License       *Licenses        `json:"license"`
}

type CodeScanResult struct {
	Vulnerability     *Vulnerabilities   `json:"vulnerability"`
	License           *Licenses          `json:"license"`
	MisConfigurations *MisConfigurations `json:"misConfigurations"`
	ExposedSecrets    *ExposedSecrets    `json:"exposedSecrets"`
	Metadata
}

type K8sManifestScanResult struct {
	MisConfigurations *MisConfigurations `json:"misConfigurations"`
	Metadata
}

type Metadata struct {
	Status    string    `json:"status"`
	StartedOn time.Time `json:"StartedOn"`
}
