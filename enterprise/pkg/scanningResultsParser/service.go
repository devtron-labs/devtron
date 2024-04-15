package scanningResultsParser

import (
	"errors"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type Service interface {
	GetScanResults(appId, envId int) (Response, error)
}

type ServiceImpl struct {
	// FindLatestCdWorkflowRunnerByEnvironmentIdAndRunnerType
	logger                      *zap.SugaredLogger
	cdWorkflowRepo              pipelineConfig.CdWorkflowRepository
	artifactRepo                repository.CiArtifactRepository
	imageScanningDeployInfoRepo security.ImageScanDeployInfoRepository
	imageScanHistoryRepository  security.ImageScanHistoryRepository
}

func NewServiceImpl(cdWorkflowRepo pipelineConfig.CdWorkflowRepository,
	artifactRepo repository.CiArtifactRepository,
	imageScanningDeployInfoRepo security.ImageScanDeployInfoRepository,
	imageScanHistoryRepository security.ImageScanHistoryRepository,
	logger *zap.SugaredLogger,
) *ServiceImpl {
	return &ServiceImpl{
		cdWorkflowRepo:              cdWorkflowRepo,
		artifactRepo:                artifactRepo,
		imageScanningDeployInfoRepo: imageScanningDeployInfoRepo,
		imageScanHistoryRepository:  imageScanHistoryRepository,
		logger:                      logger,
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
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching image scan deploy info for cd workflow", "err", err, "appId", appId, "envId", envId)
		return resp, err
	}
	imageScanHistoryIds = append(imageScanHistoryIds, imageScanDeployInfo.ImageScanExecutionHistoryId...)

	// get the scan results for all the history ids
	scanHistories, err := impl.imageScanHistoryRepository.FetchWithHistoryIds(imageScanHistoryIds)
	if err != nil {
		impl.logger.Errorw("error in fetching resource scan result with history ids given", "err", err, "appId", appId, "envId", envId)
		return resp, err
	}

	var imageScanExecs = make(map[string]*security.ExecutionData)
	var codeScanExec *security.ExecutionData
	var k8sManifestScanExec *security.ExecutionData
	for _, scanHistory := range scanHistories {
		if scanHistory.IsBuiltImage() || scanHistory.IsManifestImage() {
			imageScanExecs[scanHistory.Image] = scanHistory
		} else if scanHistory.IsManifest() {
			k8sManifestScanExec = scanHistory
		} else {
			codeScanExec = scanHistory

		}
	}

	if codeScanExec != nil {
		if parseCodePtr := ParseCodeScanResult(codeScanExec.ScanDataJson); parseCodePtr != nil {
			resp.CodeScan = *parseCodePtr
			resp.CodeScan.Metadata = Metadata{
				ScanToolName: codeScanExec.ScanToolName,
				StartedOn:    codeScanExec.StartedOn,
				Status:       codeScanExec.Status.String(),
			}
		}
	}

	if k8sManifestScanExec != nil {
		if parseManifestPtr := ParseK8sConfigScanResult(k8sManifestScanExec.ScanDataJson); parseManifestPtr != nil {
			resp.KubernetesManifest = *parseManifestPtr
			resp.KubernetesManifest.Metadata = Metadata{
				ScanToolName: k8sManifestScanExec.ScanToolName,
				StartedOn:    k8sManifestScanExec.StartedOn,
				Status:       k8sManifestScanExec.Status.String(),
			}
		}
	}

	if len(imageScanExecs) > 0 {
		if imageScanResponse := getImageScanResult(imageScanExecs); imageScanResponse != nil {
			resp.ImageScan = *imageScanResponse
		}
	}
	return resp, err
}

func getImageScanResult(imageScanExecs map[string]*security.ExecutionData) *ImageScanResponse {
	vulnerabilityResponse := &VulnerabilityResponse{}
	licensesResponse := &LicenseResponse{}
	for image, imageScanExec := range imageScanExecs {
		if imageScanResp := ParseImageScanResult(imageScanExec.ScanDataJson); imageScanResp != nil {

			// collect vulnerabilities
			vulnerabilities := ImageVulnerability{
				Image: image,
				Metadata: Metadata{
					ScanToolName: imageScanExec.ScanToolName,
					StartedOn:    imageScanExec.StartedOn,
					Status:       imageScanExec.Status.String(),
				},
			}

			if imageScanResp.Vulnerability != nil {
				vulnerabilities.Vulnerabilities = *imageScanResp.Vulnerability
				vulnerabilityResponse.append(vulnerabilities)
			}

			// collect licenses
			licenses := ImageLicenses{
				Image: image,
				Metadata: Metadata{
					ScanToolName: imageScanExec.ScanToolName,
					StartedOn:    imageScanExec.StartedOn,
					Status:       imageScanExec.Status.String(),
				},
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
