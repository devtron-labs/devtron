/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package parser

import (
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
	ScanToolUrl  string    `json:"scanToolUrl"`
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
	*Metadata
}

type ImageScanResponse struct {
	Vulnerability *VulnerabilityResponse `json:"vulnerability"`
}

type ResourceScanResponseDto struct {
	Scanned            bool               `json:"scanned"`
	IsImageScanEnabled bool               `json:"isImageScanEnabled"`
	ImageScan          *ImageScanResponse `json:"imageScan"`
}
