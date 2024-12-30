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

func (r *ImageScanEvent) IsManifestImage() bool {
	return r.SourceType == constants.SourceTypeImage && r.SourceSubType == constants.SourceSubTypeManifest
}
