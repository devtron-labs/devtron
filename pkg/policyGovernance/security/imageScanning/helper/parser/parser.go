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
	"github.com/tidwall/gjson"
	"strings"
)

type JsonKey string
type JsonVal string

func (jp JsonKey) string() string {
	return string(jp)
}

func (jv JsonVal) string() string {
	return string(jv)
}

const Results JsonKey = "Results"

// License parameters json path
const (
	ClassificationKey JsonKey = "Category"
	SeverityKey       JsonKey = "Severity"
	LicenseKey        JsonKey = "Name"
	PackageKey        JsonKey = "PkgName"
	SourceKey         JsonKey = "FilePath"
	ClassKey          JsonKey = "Class"
)

const (
	TypeKey JsonKey = "Type"
)

// Vulnerabilities paths
const (
	VulnerabilitiesKey JsonKey = "Vulnerabilities"
	CVEIdKey           JsonKey = "VulnerabilityID"
	CurrentVersionKey  JsonKey = "InstalledVersion"
	FixedInVersionKey  JsonKey = "FixedVersion"
	TargetKey          JsonKey = "Target"
)

func parseVulnerabilities(scanResult string, severityToSkipMap map[string]bool) *Vulnerabilities {
	vulnerabilitiesRes := &Vulnerabilities{}
	if results := gjson.Get(scanResult, Results.string()); results.IsArray() {
		results.ForEach(func(_, val gjson.Result) bool {
			target := val.Get(TargetKey.string()).String()
			class := val.Get(ClassKey.string()).String()
			typeName := val.Get(TypeKey.string()).String()
			if vulnerabilities := val.Get(VulnerabilitiesKey.string()); vulnerabilities.IsArray() {
				vulnerabilities.ForEach(func(_, vulnerability gjson.Result) bool {
					license := Vulnerability{
						CVEId:          vulnerability.Get(CVEIdKey.string()).String(),
						Severity:       Severity(vulnerability.Get(SeverityKey.string()).String()),
						CurrentVersion: vulnerability.Get(CurrentVersionKey.string()).String(),
						Package:        vulnerability.Get(PackageKey.string()).String(),
						FixedInVersion: vulnerability.Get(FixedInVersionKey.string()).String(),
						Target:         target,
						Class:          class,
						Type:           typeName,
					}
					if _, ok := severityToSkipMap[strings.ToLower(license.Severity.ToString())]; !ok {
						vulnerabilitiesRes.Vulnerabilities = append(vulnerabilitiesRes.Vulnerabilities, license)
					}
					return true
				})
			}
			return true
		})
	}
	if vulnerabilitiesRes != nil {
		vulnerabilitiesRes.Summary = BuildVulnerabilitySummary(vulnerabilitiesRes.Vulnerabilities)
	}
	return vulnerabilitiesRes
}

func BuildVulnerabilitySummary(allVulnerabilities []Vulnerability) Summary {
	summary := make(map[Severity]int)
	for _, vulnerability := range allVulnerabilities {
		summary[vulnerability.Severity] = summary[vulnerability.Severity] + 1
	}
	return Summary{
		Severities: summary,
	}
}
