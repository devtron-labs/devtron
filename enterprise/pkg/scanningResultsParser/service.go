package scanningResultsParser

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"go.uber.org/zap"
)

type Service interface {
	GetScanResults(appId, envId int) (Response, error)
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

	imageScanHistoryIds := make([]int, 0)
	ciWorkflowId := ciArtifact.WorkflowId
	if ciWorkflowId != nil {
		//  for image(image built by us(devtron)) and code scan result fetching
		imageScanDeployInfo, err := impl.imageScanningDeployInfoRepo.FindByTypeMetaAndTypeId(*ciWorkflowId, security.ScanObjectType_CI_Workflow)
		if err != nil {
			impl.logger.Errorw("error in fetching image scan deploy info for ci workflow", "err", err, "appId", appId, "envId", envId)
			return resp, err
		}
		imageScanHistoryIds = imageScanDeployInfo.ImageScanExecutionHistoryId
	}

	// for image(images present in manifests and not built by us) and k8s manifest scan results
	imageScanDeployInfo, err := impl.imageScanningDeployInfoRepo.FindByTypeMetaAndTypeId(cdWfRunner.Id, security.ScanObjectType_CD_Workflow)
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

	var imageScanJsons map[string]*string
	var codeScanJson *string
	var k8sManifestScanJson *string
	for _, scanHistory := range scanHistories {
		if scanHistory.ImageScanExecutionHistory.IsBuiltImage() || scanHistory.ImageScanExecutionHistory.IsManifestImage() {
			imageScanJsons[scanHistory.ImageScanExecutionHistory.Image] = &scanHistory.ScanDataJson
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

	if imageScanResponse := getImageScanResult(imageScanJsons); imageScanResponse != nil {
		resp.ImageScan = *imageScanResponse
	}
	return resp, err
}

func getImageScanResult(imageScanJsons map[string]*string) *ImageScanResponse {
	vulnerabilityResponse := &VulnerabilityResponse{}
	licensesResponse := &LicenseResponse{}
	for image, imageScanJson := range imageScanJsons {
		if imageScanResp := ParseImageScanResult(*imageScanJson); imageScanResp != nil {

			// collect vulnerabilities
			vulnerabilities := ImageVulnerability{
				Image: image,
			}
			if imageScanResp.Vulnerability != nil {
				vulnerabilities.Vulnerabilities = *imageScanResp.Vulnerability
				vulnerabilityResponse.append(vulnerabilities)
			}

			// collect licenses
			licenses := ImageLicenses{
				Image: image,
			}

			if imageScanResp.License != nil {
				licenses.Licenses = *imageScanResp.License
				vulnerabilityResponse.append(vulnerabilities)
				licensesResponse.append(licenses)
			}

		}
	}

	return &ImageScanResponse{
		Vulnerability: *vulnerabilityResponse,
		License:       *licensesResponse,
	}
}
