/*
 * Copyright (c) 2024. Devtron Inc.
 */

package bean

const (
	HIGH     string = "high"
	CRITICAL string = "critical"
	SAFE     string = "safe"
	LOW      string = "low"
	MEDIUM   string = "medium"
	MODERATE string = "moderate"
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
)

// Handling for future use
func (d Severity) ValuesOf(severity string) Severity {
	if severity == CRITICAL || severity == HIGH {
		return Critical
	} else if severity == MODERATE || severity == MEDIUM {
		return Medium
	} else if severity == LOW || severity == SAFE {
		return Low
	}
	return Low
}

// Updating it for future use(not in use for standard severity)
func (d Severity) String() string {
	return [...]string{LOW, MODERATE, CRITICAL, HIGH, SAFE}[d]
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
