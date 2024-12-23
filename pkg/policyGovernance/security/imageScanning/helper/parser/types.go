/*
 * Copyright (c) 2024. Devtron Inc.
 */

package parser

import (
	"fmt"
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

func (r Severity) ToString() string {
	return string(r)
}

type Summary struct {
	Severities map[Severity]int `json:"severities"`
}

func (summary *Summary) String() string {
	return fmt.Sprintf("%d Critical, %d High, %d Medium, %d Low, %d Unknown", summary.Severities[CRITICAL], summary.Severities[HIGH], summary.Severities[MEDIUM], summary.Severities[LOW], summary.Severities[UNKNOWN])
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
	Target         string   `json:"target"`         // Target
	Class          string   `json:"class"`          // Class
	Type           string   `json:"type"`           // Type
	Permission     string   `json:"permission"`
}

type ImageScanResult struct {
	Vulnerability *Vulnerabilities `json:"vulnerability"`
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

func (vr *VulnerabilityResponse) Append(iv ImageVulnerability) {
	if vr.Summary.Severities == nil {
		vr.Summary.Severities = make(map[Severity]int)
	}

	vr.List = append(vr.List, iv)
	for key, val := range iv.Summary.Severities {
		vr.Summary.Severities[key] += val
	}
}

type ImageVulnerability struct {
	Image string `json:"image"`
	Vulnerabilities
	Metadata
}

type ImageScanResponse struct {
	Vulnerability *VulnerabilityResponse `json:"vulnerability"`
}

type ResourceScanResponseDto struct {
	Scanned            bool               `json:"scanned"`
	IsImageScanEnabled bool               `json:"isImageScanEnabled"`
	ImageScan          *ImageScanResponse `json:"imageScan"`
}
