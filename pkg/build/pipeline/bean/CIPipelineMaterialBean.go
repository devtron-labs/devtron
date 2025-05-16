package bean

import (
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"time"
)

type CiPipelineMaterialResponse struct {
	Id              int                    `json:"id"`
	GitMaterialId   int                    `json:"gitMaterialId"`
	GitMaterialUrl  string                 `json:"gitMaterialUrl"`
	GitMaterialName string                 `json:"gitMaterialName"`
	Type            string                 `json:"type"`
	Value           string                 `json:"value"`
	Active          bool                   `json:"active"`
	History         []*gitSensor.GitCommit `json:"history,omitempty"`
	LastFetchTime   time.Time              `json:"lastFetchTime"`
	IsRepoError     bool                   `json:"isRepoError"`
	RepoErrorMsg    string                 `json:"repoErrorMsg"`
	IsBranchError   bool                   `json:"isBranchError"`
	BranchErrorMsg  string                 `json:"branchErrorMsg"`
	Url             string                 `json:"url"`
	Regex           string                 `json:"regex"`
}

type MaterialTriggerInfo struct {
	GitTriggers map[int]pipelineConfig.GitCommit `json:"gitTriggers"`
	CiMaterials []CiPipelineMaterialResponse     `json:"ciMaterials"`
}
