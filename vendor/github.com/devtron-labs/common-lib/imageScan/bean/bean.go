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

import (
	"github.com/devtron-labs/common-lib/constants"
	git "github.com/devtron-labs/common-lib/git-manager"
)

type ManifestData struct {
	ChartData  []byte `json:"chartData"`
	ValuesYaml []byte `json:"valuesYaml"`
}

type ImageScanEvent struct {
	Image            string                  `json:"image"`
	ImageDigest      string                  `json:"imageDigest"`
	AppId            int                     `json:"appId"`
	EnvId            int                     `json:"envId"`
	PipelineId       int                     `json:"pipelineId"`
	CiArtifactId     int                     `json:"ciArtifactId"`
	UserId           int                     `json:"userId"`
	AccessKey        string                  `json:"accessKey"`
	SecretKey        string                  `json:"secretKey"`
	Token            string                  `json:"token"`
	AwsRegion        string                  `json:"awsRegion"`
	DockerRegistryId string                  `json:"dockerRegistryId"`
	DockerConnection string                  `json:"dockerConnection"`
	DockerCert       string                  `json:"dockerCert"`
	CiProjectDetails []git.CiProjectDetails  `json:"ciProjectDetails"`
	SourceType       constants.SourceType    `json:"sourceType"`
	SourceSubType    constants.SourceSubType `json:"sourceSubType"`
	CiWorkflowId     int                     `json:"ciWorkflowId"`
	CdWorkflowId     int                     `json:"cdWorkflowId"`
	ChartHistoryId   int                     `json:"chartHistoryId"`
	ManifestData     *ManifestData           `json:"manifestData"`
	ReScan           bool                    `json:"reScan"`
}

func (r *ImageScanEvent) IsManifest() bool {
	return r.SourceType == constants.SourceTypeCode && r.SourceSubType == constants.SourceSubTypeManifest
}

func (r *ImageScanEvent) IsImageFromManifest() bool {
	return r.SourceType == constants.SourceTypeImage && r.SourceSubType == constants.SourceSubTypeManifest
}

func (r *ImageScanEvent) IsBuiltImage() bool {
	return r.SourceType == constants.SourceTypeImage && r.SourceSubType == constants.SourceSubTypeCi
}

type ScanResultPayload struct {
	ImageScanEvent       *ImageScanEvent
	ScanToolId           int                      `json:"scanToolId"`
	SourceScanningResult string                   `json:"sourceScanningResult"`
	Sbom                 string                   `json:"sbom"`
	ImageScanOutput      []*ImageScanOutputObject `json:"imageScanOutput"`
}

type ImageScanOutputObject struct {
	TargetName     string `json:"targetName"`
	Class          string `json:"class"`
	Type           string `json:"type"`
	Name           string `json:"name"`
	Package        string `json:"package"`
	PackageVersion string `json:"packageVersion"`
	FixedInVersion string `json:"fixedInVersion"`
	Severity       string `json:"severity"`
}
