package scanningResultsParser

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
)

type Service interface {
	GetScanResults(appId, envId int)
}

type ServiceImpl struct {
	// FindLatestCdWorkflowRunnerByEnvironmentIdAndRunnerType
	cdWorkflowRepo              pipelineConfig.CdWorkflowRepository
	artifactRepo                repository.CiArtifactRepository
	imageScanningDeployInfoRepo security.ImageScanDeployInfoRepository
	imageScanHistoryRepository  security.ImageScanHistoryRepository
}

func NewServiceImpl(cdWorkflowRepo pipelineConfig.CdWorkflowRepository,
	artifactRepo repository.CiArtifactRepository,
	imageScanningDeployInfoRepo security.ImageScanDeployInfoRepository,
	imageScanHistoryRepository security.ImageScanHistoryRepository,
) *ServiceImpl {
	return &ServiceImpl{
		cdWorkflowRepo:              cdWorkflowRepo,
		artifactRepo:                artifactRepo,
		imageScanningDeployInfoRepo: imageScanningDeployInfoRepo,
		imageScanHistoryRepository:  imageScanHistoryRepository,
	}
}

func (impl ServiceImpl) GetScanResults(appId, envId int) {
	cdWfRunner, err := impl.cdWorkflowRepo.FindLatestCdWorkflowRunnerByEnvironmentIdAndRunnerType(appId, envId, bean.CD_WORKFLOW_TYPE_DEPLOY)
	if err != nil {
		return
	}

	ciArtifact, err := impl.artifactRepo.Get(cdWfRunner.CdWorkflow.CiArtifactId)
	if err != nil {
		return
	}

	//  for image(image built by us(devtron)) and code scan result fetching
	ciWorkflowId := ciArtifact.WorkflowId
	imageScanDeployInfo, err := impl.imageScanningDeployInfoRepo.FindByTypeMetaAndTypeId(*ciWorkflowId, security.ScanObjectType_CI_Workflow)
	if err != nil {
		return
	}
	imageScanHistoryIds := imageScanDeployInfo.ImageScanExecutionHistoryId

	// todo: for image(images present in manifests and not built by us) and k8s manifest scan results

	// get the scan results for all the history ids
	scanHistories, err := impl.imageScanHistoryRepository.FindByIds(imageScanHistoryIds)
	if err != nil {
		return
	}

}
