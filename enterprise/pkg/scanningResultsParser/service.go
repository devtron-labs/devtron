package scanningResultsParser

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"io/ioutil"
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

func (impl ServiceImpl) GetScanResults(appId, envId int) error {
	cdWfRunner, err := impl.cdWorkflowRepo.FindLatestCdWorkflowRunnerByEnvironmentIdAndRunnerType(appId, envId, bean.CD_WORKFLOW_TYPE_DEPLOY)
	if err != nil {
		return err
	}

	ciArtifact, err := impl.artifactRepo.Get(cdWfRunner.CdWorkflow.CiArtifactId)
	if err != nil {
		return err
	}

	//  for image(image built by us(devtron)) and code scan result fetching
	ciWorkflowId := ciArtifact.WorkflowId
	imageScanDeployInfo, err := impl.imageScanningDeployInfoRepo.FindByTypeMetaAndTypeId(*ciWorkflowId, security.ScanObjectType_CI_Workflow)
	if err != nil {
		return err
	}
	imageScanHistoryIds := imageScanDeployInfo.ImageScanExecutionHistoryId
	cdWorkflowId := cdWfRunner.Id
	imageScanInfo, err := impl.imageScanningDeployInfoRepo.FindByTypeMetaAndTypeId(cdWorkflowId, security.ScanObjectType_CD_Workflow)
	if err != nil {
		return err
	}
	imageScanHistoryIds = append(imageScanHistoryIds, imageScanInfo.ImageScanExecutionHistoryId...)
	// todo: for image(images present in manifests and not built by us) and k8s manifest scan results

	// get the scan results for all the history ids
	scanHistories, err := impl.resourceScanResultRepository.FetchWithHistoryIds(imageScanHistoryIds)
	if err != nil {
		return err
	}
	var scanInfoImages []security.ResourceScanResult
	var scanInfoCode security.ResourceScanResult
	var scanInfoManifest []security.ResourceScanResult
	for _, scanHistory := range *scanHistories {
		if (scanHistory.ImageScanExecutionHistory.SourceType == 1 && scanHistory.ImageScanExecutionHistory.SourceSubType == 1) && (scanHistory.ImageScanExecutionHistory.SourceType == 1 && scanHistory.ImageScanExecutionHistory.SourceSubType == 2) {
			scanInfoImages = append(scanInfoImages, scanHistory)
		} else if scanHistory.ImageScanExecutionHistory.SourceType == 2 && scanHistory.ImageScanExecutionHistory.SourceSubType == 1 {
			scanInfoCode = scanHistory
		} else {
			scanInfoManifest = append(scanInfoManifest, scanHistory)
		}
	}
	var parseImage []ImageScanResult
	for _, scanInfo := range scanInfoImages {
		if ParseImageScanResult(scanInfo.ScanDataJson) != nil {
			parseImage = append(parseImage, *ParseImageScanResult(scanInfo.ScanDataJson))
		}
	}
	var parseCode CodeScanResult
	if ParseCodeScanResult(scanInfoCode.ScanDataJson) != nil {
		parseCode = *ParseCodeScanResult(scanInfoCode.ScanDataJson)
	}
	var parseManifest []K8sManifestScanResult
	for _, scanInfo := range scanInfoManifest {
		if ParseK8sConfigScanResult(scanInfo.ScanDataJson) != nil {
			parseManifest = append(parseManifest, *ParseK8sConfigScanResult(scanInfo.ScanDataJson))
		}
	}

	fmt.Println(scanHistories)
	return nil
}

func (impl ServiceImpl) GetTestData(appId, envId int) Response {
	response := Response{}
	if imageData := getImageScanData(); imageData != nil {

	}
	if codeScanData := getCodeScanData(); codeScanData != nil {
		response.CodeScan = *codeScanData
	}
	if manifestData := getK8sManifestScanData(); manifestData != nil {
		response.KubernetesManifest = *manifestData
	}

	return response
}

func loadData(fileName string) string {

	jsonBytes, _ := ioutil.ReadFile(fileName)
	return string(jsonBytes)

}

func getImageScanData() *ImageScanResult {
	jsonStr := loadData("image_scan.json")
	data := ParseImageScanResult(jsonStr)
	return data
}

func getK8sManifestScanData() *K8sManifestScanResponse {
	jsonStr := loadData("code_scan.json")
	data := ParseK8sConfigScanResult(jsonStr)
	return data
}

func getCodeScanData() *CodeScanResponse {
	jsonStr := loadData("code_scan.json")
	data := ParseCodeScanResult(jsonStr)
	return data
}
