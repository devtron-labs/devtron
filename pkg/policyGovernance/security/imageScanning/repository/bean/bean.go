/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bean

import "fmt"

const (
	HIGH     string = "high"
	CRITICAL string = "critical"
	SAFE     string = "safe"
	LOW      string = "low"
	MEDIUM   string = "medium"
	MODERATE string = "moderate"
	UNKNOWN  string = "unknown"
)

type PolicyAction int

const (
	Inherit PolicyAction = iota
	Allow
	Block
	Blockiffixed
)

func (d PolicyAction) String() string {
	return [...]string{"inherit", "allow", "block", "blockiffixed"}[d]
}

// ------------------
type Severity int

const (
	Low Severity = iota
	Medium
	Critical
	High
	Safe
	Unknown
)

func SeverityStringToEnumWithError(severity string) (Severity, error) {
	if severity == LOW {
		return Low, nil
	} else if severity == MEDIUM || severity == MODERATE {
		return Medium, nil
	} else if severity == HIGH {
		return High, nil
	} else if severity == CRITICAL {
		return Critical, nil
	} else if severity == SAFE {
		return Safe, nil
	} else if severity == UNKNOWN {
		return Unknown, nil
	}
	return 0, fmt.Errorf("unsupported Severity %s", severity)
}

//// Handling for future use
//func (d Severity) ValuesOf(severity string) Severity {
//	if severity == CRITICAL || severity == HIGH {
//		return Critical
//	} else if severity == MODERATE || severity == MEDIUM {
//		return Medium
//	} else if severity == LOW || severity == SAFE {
//		return Low
//	}
//	return Low
//}

// Updating it for future use(not in use for standard severity)
func (d Severity) String() string {
	return [...]string{LOW, MEDIUM, CRITICAL, HIGH, SAFE, UNKNOWN}[d]
}

// ----------------
type PolicyLevel int

const (
	Global PolicyLevel = iota
	Cluster
	Environment
	Application
)

func (d PolicyLevel) String() string {
	return [...]string{"global", "cluster", "environment", "application"}[d]
}

type ImageScanFilter struct {
	Offset  int    `json:"offset"`
	Size    int    `json:"size"`
	CVEName string `json:"cveName"`
	AppName string `json:"appName"`
	// ObjectName deprecated
	ObjectName     string    `json:"objectName"`
	EnvironmentIds []int     `json:"envIds"`
	ClusterIds     []int     `json:"clusterIds"`
	Severity       []int     `json:"severity"`
	SortOrder      SortOrder `json:"sortOrder"`
	SortBy         SortBy    `json:"sortBy"` // sort by objectName,envName,lastChecked
}

type SortBy string
type SortOrder string

const (
	Asc  SortOrder = "ASC"
	Desc SortOrder = "DESC"
)
