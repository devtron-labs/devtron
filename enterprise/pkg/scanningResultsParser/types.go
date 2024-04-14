package scanningResultsParser

import (
	"fmt"
	"github.com/tidwall/gjson"
)

type Severity string

const (
	LOW      Severity = "Low"
	MEDIUM   Severity = "Medium"
	HIGH     Severity = "High"
	CRITICAL Severity = "Critical"
	UNKNOWN  Severity = "Unknown"
)

type SeveritySummary struct {
	Severities map[Severity]int
}

func (summary SeveritySummary) String() string {
	return fmt.Sprintf("%d Critical, %d High, %d Medium, %d Low, %d Unknown", summary.Severities[CRITICAL], summary.Severities[HIGH], summary.Severities[MEDIUM], summary.Severities[LOW], summary.Severities[UNKNOWN])
}

type Licenses struct {
	Summary  SeveritySummary
	Licenses []License
}

type License struct {
	Classification string   // Category
	Severity       Severity // Severity
	License        string   // Name
	Package        string   // PkgName
	Source         string   // FilePath
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
	Summary         SeveritySummary
	Vulnerabilities []Vulnerability
}

type Vulnerability struct {
	CVEId          string   // VulnerabilityID
	Severity       Severity // Severity
	Package        string   // PkgName
	CurrentVersion string   // InstalledVersion
	FixedInVersion string   // FixedVersion
}

type MisConfigurationSummary struct {
	Success    int64
	Fail       int64
	Exceptions int64
}

func (summary MisConfigurationSummary) String() string {
	return fmt.Sprintf("%d Successes, %d Failures, %d Exceptions", summary.Success, summary.Fail, summary.Exceptions)
}

type Line struct {
	Number    int64  // Number
	Content   string // Content
	IsCause   bool   // IsCause
	Truncated bool   // Truncated
}

type CauseMetadata struct {
	StartLine int64  // StartLine
	EndLine   int64  // EndLine
	Lines     []Line // Code.Lines
}

type Configuration struct {
	Id            string        // ID
	Title         string        // Title
	Message       string        // Message
	Resolution    string        // Resolution
	Status        string        // Status
	Severity      Severity      // Severity
	CauseMetadata CauseMetadata // CauseMetadata
}

type MisConfigurations struct {
	FilePath       string                  // Target
	Type           string                  // Type
	MisConfSummary MisConfigurationSummary // MisConfSummary
	Summary        SeveritySummary
	Configurations []Configuration
}

type Secret struct {
	Severity Severity
	CauseMetadata
}

type ExposedSecrets struct {
	FilePath string // target and class: secret
	Summary  SeveritySummary
	Secrets  []Secret
}
