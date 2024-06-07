/*
 * Copyright (c) 2024. Devtron Inc.
 */

package bean

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/security/bean"
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
	"time"
)

const (
	CLAIR ScannedBy = "Clair"
	TRIVY ScannedBy = "Trivy"
)

const (
	BLOCK       string = "BLOCK"
	WHITELISTED        = "WHITELISTED"
)

type ScannedBy string

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

type ImageScanExecutionInfo struct {
	ScannedAt       time.Time          `json:"scannedAt"`
	ScannedBy       string             `json:"scannedBy"`
	Vulnerabilities []*Vulnerabilities `json:"vulnerabilities"`
	SeverityCount   *SeverityCount     `json:"severityCount"`
}

type Vulnerabilities struct {
	CVEName    string `json:"cveName"`
	Severity   string `json:"severity"`
	Package    string `json:"package,omitempty"`
	CVersion   string `json:"currentVersion"`
	FVersion   string `json:"fixedVersion"`
	Permission string `json:"permission"`
}

func (vul *Vulnerabilities) IsCritical() bool {
	return vul.Severity == bean.CRITICAL
}

func (vul *Vulnerabilities) IsModerate() bool {
	return vul.Severity == bean.MODERATE
}

func (vul *Vulnerabilities) IsLow() bool {
	return vul.Severity == bean.LOW
}

type SeverityCount struct {
	High     int `json:"high"`
	Moderate int `json:"moderate"`
	Low      int `json:"low"`
}

type ImageScanFilter struct {
	Offset         int    `json:"offset"`
	Size           int    `json:"size"`
	CVEName        string `json:"cveName"`
	AppName        string `json:"appName"`
	ObjectName     string `json:"objectName"`
	EnvironmentIds []int  `json:"envIds"`
	ClusterIds     []int  `json:"clusterIds"`
	Severity       []int  `json:"severity"`
}

type ImageScanRequest struct {
	ScanExecutionId       int    `json:"ScanExecutionId"`
	ImageScanDeployInfoId int    `json:"imageScanDeployInfo"`
	AppId                 int    `json:"appId"`
	EnvId                 int    `json:"envId"`
	ObjectId              int    `json:"objectId"`
	ArtifactId            int    `json:"artifactId"`
	Image                 string `json:"image"`
	ImageScanFilter
}

type ImageScanExecutionDetail struct {
	ImageScanDeployInfoId int    `json:"imageScanDeployInfoId"`
	AppId                 int    `json:"appId,omitempty"`
	EnvId                 int    `json:"envId,omitempty"`
	AppName               string `json:"appName,omitempty"`
	EnvName               string `json:"envName,omitempty"`
	ArtifactId            int    `json:"artifactId,omitempty"`
	Image                 string `json:"image,omitempty"`
	PodName               string `json:"podName,omitempty"`
	ReplicaSet            string `json:"replicaSet,omitempty"`
	ScanEnabled           bool   `json:"scanEnabled,notnull"`
	Scanned               bool   `json:"scanned,notnull"`
	ObjectType            string `json:"objectType,notnull"`
	ScanResult
}

type ImageScanResult struct {
	ScanResult ScanResult                           `json:"scanResult"`
	Image      string                               `json:"image"`
	State      serverBean.ScanExecutionProcessState `json:"state"`
	Error      string                               `json:"error"`
}

type ScanResult struct {
	Vulnerabilities []*Vulnerabilities `json:"vulnerabilities,omitempty"`
	SeverityCount   *SeverityCount     `json:"severityCount,omitempty"`
	ExecutionTime   time.Time          `json:"executionTime,omitempty"`
	ScanToolId      int                `json:"scanToolId,omitempty""`
}
