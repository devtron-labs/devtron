package scanningResultsParser

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"go.uber.org/zap"
)

type Service interface {
	GetScanResults(appId, envId int) (resp Response, err error)
}

type ServiceImpl struct {
	// FindLatestCdWorkflowRunnerByEnvironmentIdAndRunnerType
	logger                       *zap.SugaredLogger
	cdWorkflowRepo               pipelineConfig.CdWorkflowRepository
	artifactRepo                 repository.CiArtifactRepository
	imageScanningDeployInfoRepo  security.ImageScanDeployInfoRepository
	imageScanHistoryRepository   security.ImageScanHistoryRepository
	resourceScanResultRepository security.ResourceScanResultRepository
}

func NewServiceImpl(cdWorkflowRepo pipelineConfig.CdWorkflowRepository,
	artifactRepo repository.CiArtifactRepository,
	imageScanningDeployInfoRepo security.ImageScanDeployInfoRepository,
	imageScanHistoryRepository security.ImageScanHistoryRepository,
	resourceScanResultRepository security.ResourceScanResultRepository,
	logger *zap.SugaredLogger,
) *ServiceImpl {
	return &ServiceImpl{
		cdWorkflowRepo:               cdWorkflowRepo,
		artifactRepo:                 artifactRepo,
		imageScanningDeployInfoRepo:  imageScanningDeployInfoRepo,
		imageScanHistoryRepository:   imageScanHistoryRepository,
		resourceScanResultRepository: resourceScanResultRepository,
		logger:                       logger,
	}
}

func (impl ServiceImpl) GetScanResults(appId, envId int) (resp Response, err error) {
	cdWfRunner, err := impl.cdWorkflowRepo.FindLatestCdWorkflowRunnerByEnvironmentIdAndRunnerType(appId, envId, bean.CD_WORKFLOW_TYPE_DEPLOY)
	if err != nil {
		impl.logger.Errorw("error in fetching cd workflow runner  ", "err", err, "appId", appId, "envId", envId)
		return resp, err
	}

	ciArtifact, err := impl.artifactRepo.Get(cdWfRunner.CdWorkflow.CiArtifactId)
	if err != nil {
		impl.logger.Errorw("error in fetching ci artifact", "err", err, "appId", appId, "envId", envId)
		return resp, err
	}
	//  for image(image built by us(devtron)) and code scan result fetching
	imageScanDeployInfo, err := impl.imageScanningDeployInfoRepo.FindByTypeMetaAndTypeId(*ciArtifact.WorkflowId, security.ScanObjectType_CI_Workflow)
	if err != nil {
		impl.logger.Errorw("error in fetching image scan deploy info for ci workflow", "err", err, "appId", appId, "envId", envId)
		return resp, err
	}
	imageScanHistoryIds := imageScanDeployInfo.ImageScanExecutionHistoryId

	// for image(images present in manifests and not built by us) and k8s manifest scan results
	imageScanDeployInfo, err = impl.imageScanningDeployInfoRepo.FindByTypeMetaAndTypeId(cdWfRunner.Id, security.ScanObjectType_CD_Workflow)
	if err != nil {
		impl.logger.Errorw("error in fetching image scan deploy info for cd workflow", "err", err, "appId", appId, "envId", envId)
		return resp, err
	}
	imageScanHistoryIds = append(imageScanHistoryIds, imageScanDeployInfo.ImageScanExecutionHistoryId...)

	// get the scan results for all the history ids
	scanHistories, err := impl.resourceScanResultRepository.FetchWithHistoryIds(imageScanHistoryIds)
	if err != nil {
		impl.logger.Errorw("error in fetching resource scan result with history ids given", "err", err, "appId", appId, "envId", envId)
		return resp, err
	}
	var scanInfoImages []security.ResourceScanResult
	var scanInfoCode security.ResourceScanResult
	var scanInfoManifest security.ResourceScanResult
	for _, scanHistory := range *scanHistories {
		if (scanHistory.ImageScanExecutionHistory.SourceType == 1 && scanHistory.ImageScanExecutionHistory.SourceSubType == 1) || (scanHistory.ImageScanExecutionHistory.SourceType == 1 && scanHistory.ImageScanExecutionHistory.SourceSubType == 2) {
			scanInfoImages = append(scanInfoImages, scanHistory)
		} else if scanHistory.ImageScanExecutionHistory.SourceType == 2 && scanHistory.ImageScanExecutionHistory.SourceSubType == 1 {
			scanInfoCode = scanHistory
		} else {
			scanInfoManifest = scanHistory
		}
	}

	if parseCodePtr := ParseCodeScanResult(scanInfoCode.ScanDataJson); parseCodePtr != nil {
		resp.CodeScan = *parseCodePtr
	}

	if parseManifestPtr := ParseK8sConfigScanResult(scanInfoManifest.ScanDataJson); parseManifestPtr != nil {
		resp.KubernetesManifest = *parseManifestPtr
	}

	return resp, err
}

func getImageScanResult() {
	// var parseImage []ImageScanResult
	// for _, scanInfo := range scanInfoImages {
	// 	if imageScanResp := ParseImageScanResult(scanInfo.ScanDataJson); imageScanResp != nil {
	// 		imageScanResp.
	// 			parseImage = append(parseImage, *ParseImageScanResult(scanInfo.ScanDataJson))
	// 	}
	// }

}
