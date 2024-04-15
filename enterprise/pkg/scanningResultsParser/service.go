package scanningResultsParser

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
)

type Service interface {
	GetScanResults(appId, envId int) error
}

type ServiceImpl struct {
	// FindLatestCdWorkflowRunnerByEnvironmentIdAndRunnerType
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
) *ServiceImpl {
	return &ServiceImpl{
		cdWorkflowRepo:               cdWorkflowRepo,
		artifactRepo:                 artifactRepo,
		imageScanningDeployInfoRepo:  imageScanningDeployInfoRepo,
		imageScanHistoryRepository:   imageScanHistoryRepository,
		resourceScanResultRepository: resourceScanResultRepository,
	}
}

func (impl ServiceImpl) GetScanResults(appId, envId int) (resp Response, err error) {
	cdWfRunner, err := impl.cdWorkflowRepo.FindLatestCdWorkflowRunnerByEnvironmentIdAndRunnerType(appId, envId, bean.CD_WORKFLOW_TYPE_DEPLOY)
	if err != nil {
		return resp, err
	}

	ciArtifact, err := impl.artifactRepo.Get(cdWfRunner.CdWorkflow.CiArtifactId)
	if err != nil {
		return resp, err
	}

	imageScanHistoryIds := make([]int, 0)
	ciWorkflowId := ciArtifact.WorkflowId
	if ciWorkflowId != nil {
		//  for image(image built by us(devtron)) and code scan result fetching
		imageScanDeployInfo, err := impl.imageScanningDeployInfoRepo.FindByTypeMetaAndTypeId(*ciWorkflowId, security.ScanObjectType_CI_Workflow)
		if err != nil {
			return resp, err
		}
		imageScanHistoryIds = imageScanDeployInfo.ImageScanExecutionHistoryId
	}

	// for image(images present in manifests and not built by us) and k8s manifest scan results
	imageScanDeployInfo, err := impl.imageScanningDeployInfoRepo.FindByTypeMetaAndTypeId(cdWfRunner.Id, security.ScanObjectType_CD_Workflow)
	if err != nil {
		return resp, err
	}
	imageScanHistoryIds = append(imageScanHistoryIds, imageScanDeployInfo.ImageScanExecutionHistoryId...)

	// get the scan results for all the history ids
	scanHistories, err := impl.resourceScanResultRepository.FetchWithHistoryIds(imageScanHistoryIds)
	if err != nil {
		return resp, err
	}

	var imageScanJsons []*string
	var codeScanJson *string
	var k8sManifestScanJson *string
	for _, scanHistory := range scanHistories {
		if scanHistory.ImageScanExecutionHistory.IsBuiltImage() || scanHistory.ImageScanExecutionHistory.IsManifestImage() {
			imageScanJsons = append(imageScanJsons, &scanHistory.ScanDataJson)
		} else if scanHistory.ImageScanExecutionHistory.SourceType == 2 && scanHistory.ImageScanExecutionHistory.SourceSubType == 1 {
			codeScanJson = &scanHistory.ScanDataJson
		} else {
			k8sManifestScanJson = &scanHistory.ScanDataJson
		}
	}

	if parseCodePtr := ParseCodeScanResult(*codeScanJson); parseCodePtr != nil {
		resp.CodeScan = *parseCodePtr
	}

	if parseManifestPtr := ParseK8sConfigScanResult(*k8sManifestScanJson); parseManifestPtr != nil {
		resp.KubernetesManifest = *parseManifestPtr
	}

	// if imageScanResponse := getImageScanResult(imageScanJsons); imageScanResponse != nil {
	// 	resp.ImageScan = *imageScanResponse
	// }
	return resp, err
}

// func getImageScanResult(imageScanJsons []*string) *ImageScanResponse {
// 	var parseImage []ImageScanResult
// 	for _, imageScanJson := range imageScanJsons {
// 		if imageScanResp := ParseImageScanResult(*imageScanJson); imageScanResp != nil {
//
// 		}
// 	}
//
// }
