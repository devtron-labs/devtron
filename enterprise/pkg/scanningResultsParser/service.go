/*
 * Copyright (c) 2024. Devtron Inc.
 */

package scanningResultsParser

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	repository4 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/go-errors/errors"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type Service interface {
	GetScanResults(appId, envId, ciWorkflowId, installedAppId int) (Response, error)
}

type ServiceImpl struct {
	// FindLatestCdWorkflowRunnerByEnvironmentIdAndRunnerType
	logger                               *zap.SugaredLogger
	cdWorkflowRepo                       pipelineConfig.CdWorkflowRepository
	artifactRepo                         repository.CiArtifactRepository
	imageScanningDeployInfoRepo          security.ImageScanDeployInfoRepository
	imageScanHistoryRepository           security.ImageScanHistoryRepository
	installedAppVersionHistoryRepository repository4.InstalledAppVersionHistoryRepository
	// appStoreService             service.AppStoreService
	// appStoreDeploymentService   service2.AppStoreDeploymentService
}

func NewServiceImpl(cdWorkflowRepo pipelineConfig.CdWorkflowRepository,
	artifactRepo repository.CiArtifactRepository,
	imageScanningDeployInfoRepo security.ImageScanDeployInfoRepository,
	imageScanHistoryRepository security.ImageScanHistoryRepository,
	logger *zap.SugaredLogger,
	installedAppVersionHistoryRepository repository4.InstalledAppVersionHistoryRepository) *ServiceImpl {
	return &ServiceImpl{
		cdWorkflowRepo:                       cdWorkflowRepo,
		artifactRepo:                         artifactRepo,
		imageScanningDeployInfoRepo:          imageScanningDeployInfoRepo,
		imageScanHistoryRepository:           imageScanHistoryRepository,
		logger:                               logger,
		installedAppVersionHistoryRepository: installedAppVersionHistoryRepository,
	}
}

func (impl ServiceImpl) GetScanResults(appId, envId, ciWorkflowId, installedAppId int) (resp Response, err error) {

	imageScanHistoryIds := make([]int, 0)
	if installedAppId > 0 {
		versionHistory, err := impl.installedAppVersionHistoryRepository.GetLatestInstalledAppVersionHistoryByInstalledAppId(installedAppId)
		if err != nil {
			impl.logger.Errorw("error occurred while fetching latest installed app version history", "installedAppId", installedAppId, "err", err)
			return resp, err
		}
		imageScanDeployInfo, err := impl.imageScanningDeployInfoRepo.FindByTypeMetaAndTypeId(versionHistory.Id, security.ScanObjectType_CHART_HISTORY)
		if err != nil && !errors.Is(err, pg.ErrNoRows) {
			impl.logger.Errorw("error in fetching image scan deploy info for cd workflow", "err", err, "appId", appId, "envId", envId)
			return resp, err
		}
		imageScanHistoryIds = append(imageScanHistoryIds, imageScanDeployInfo.ImageScanExecutionHistoryId...)

	} else {
		var cdWfRunner pipelineConfig.CdWorkflowRunner
		if appId > 0 && envId > 0 {
			cdWfRunner, err = impl.cdWorkflowRepo.FindLatestCdWorkflowRunnerByEnvironmentIdAndRunnerType(appId, envId, bean.CD_WORKFLOW_TYPE_DEPLOY)
			if err != nil {
				impl.logger.Errorw("error in fetching cd workflow runner  ", "err", err, "appId", appId, "envId", envId)
				return resp, err
			}
		}

		// if ciWorkflowId is given , don't need to fetch the
		if ciWorkflowId == 0 {
			ciArtifact, err := impl.artifactRepo.Get(cdWfRunner.CdWorkflow.CiArtifactId)
			if err != nil {
				impl.logger.Errorw("error in fetching ci artifact", "err", err, "appId", appId, "envId", envId)
				return resp, err
			}
			if ciArtifact.WorkflowId != nil {
				ciWorkflowId = *ciArtifact.WorkflowId
			}
		}

		// if ciWorkflowId is still 0 , which means the ciArtifact is not built in devtron, (webhook)
		if ciWorkflowId != 0 {
			//  for image(image built by us(devtron)) and code scan result fetching
			imageScanDeployInfo, err := impl.imageScanningDeployInfoRepo.FindByTypeMetaAndTypeId(ciWorkflowId, security.ScanObjectType_CI_Workflow)
			if err != nil && !errors.Is(err, pg.ErrNoRows) {
				impl.logger.Errorw("error in fetching image scan deploy info for ci workflow", "err", err, "appId", appId, "envId", envId)
				return resp, err
			}
			imageScanHistoryIds = imageScanDeployInfo.ImageScanExecutionHistoryId
		}

		// for image(images present in manifests and not built by us) and k8s manifest scan results
		imageScanDeployInfo, err := impl.imageScanningDeployInfoRepo.FindByTypeMetaAndTypeId(cdWfRunner.CdWorkflowId, security.ScanObjectType_CD_Workflow)
		if err != nil && !errors.Is(err, pg.ErrNoRows) {
			impl.logger.Errorw("error in fetching image scan deploy info for cd workflow", "err", err, "appId", appId, "envId", envId)
			return resp, err
		}
		imageScanHistoryIds = append(imageScanHistoryIds, imageScanDeployInfo.ImageScanExecutionHistoryId...)
	}

	// no histories found,so not scanned
	if len(imageScanHistoryIds) == 0 {
		resp.Scanned = false
		return resp, nil
	}

	// get the scan results for all the history ids
	scanHistories, err := impl.imageScanHistoryRepository.FetchWithHistoryIds(imageScanHistoryIds)
	if err != nil {
		impl.logger.Errorw("error in fetching resource scan result with history ids given", "err", err, "appId", appId, "envId", envId)
		return resp, err
	}

	// no histories found,so not scanned
	if len(scanHistories) == 0 {
		resp.Scanned = false
		return resp, nil
	}

	var imageScanExecs = make(map[string]*security.ExecutionData)
	var codeScanExec *security.ExecutionData
	var k8sManifestMisConfigScanExec *security.ExecutionData
	var k8sManifestSecretScanExec *security.ExecutionData
	for _, scanHistory := range scanHistories {
		if scanHistory.IsBuiltImage() || scanHistory.IsManifestImage() {
			imageScanExecs[scanHistory.Image] = scanHistory
		} else if scanHistory.IsManifest() && scanHistory.ContainsType(security.Config) {
			k8sManifestMisConfigScanExec = scanHistory
		} else if scanHistory.IsManifest() && scanHistory.ContainsType(security.Secrets) {
			k8sManifestSecretScanExec = scanHistory
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

	if k8sManifestMisConfigScanExec != nil {

		var manifestMisconfigDataJson string
		if k8sManifestMisConfigScanExec != nil {
			manifestMisconfigDataJson = k8sManifestMisConfigScanExec.ScanDataJson
		}
		var manifestSecretDataJson string
		if k8sManifestSecretScanExec != nil {
			manifestSecretDataJson = k8sManifestSecretScanExec.ScanDataJson
		}

		if parseManifestPtr := ParseK8sConfigScanResult(manifestMisconfigDataJson, manifestSecretDataJson); parseManifestPtr != nil {
			resp.KubernetesManifest = *parseManifestPtr
			resp.KubernetesManifest.Metadata = Metadata{
				ScanToolName: k8sManifestMisConfigScanExec.ScanToolName,
				StartedOn:    k8sManifestMisConfigScanExec.StartedOn,
				Status:       k8sManifestMisConfigScanExec.Status.String(),
			}
		}
	}

	if len(imageScanExecs) > 0 {
		if imageScanResponse := getImageScanResult(imageScanExecs); imageScanResponse != nil {
			resp.ImageScan = *imageScanResponse
		}
	}
	resp.Scanned = true
	return impl.sanitizeResponse(resp), err
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
				licensesResponse.append(licenses)
			}

		}
	}

	return &ImageScanResponse{
		Vulnerability: vulnerabilityResponse,
		License:       licensesResponse,
	}
}

// sanitizeResponse converting empty array to nil for consistency
func (impl ServiceImpl) sanitizeResponse(resp Response) Response {
	if resp.CodeScan.License != nil && len(resp.CodeScan.License.Licenses) == 0 {
		resp.CodeScan.License = nil
	}
	if resp.CodeScan.Vulnerability != nil && len(resp.CodeScan.Vulnerability.Vulnerabilities) == 0 {
		resp.CodeScan.Vulnerability = nil
	}

	if resp.CodeScan.ExposedSecrets != nil && len(resp.CodeScan.ExposedSecrets.ExposedSecrets) == 0 {
		resp.CodeScan.ExposedSecrets = nil
	}

	if resp.CodeScan.MisConfigurations != nil && len(resp.CodeScan.MisConfigurations.MisConfigurations) == 0 {
		resp.CodeScan.MisConfigurations = nil
	}

	if resp.ImageScan.License != nil && len(resp.ImageScan.License.List) == 0 {
		resp.ImageScan.License = nil
	}
	if resp.ImageScan.Vulnerability != nil && len(resp.ImageScan.Vulnerability.List) == 0 {
		resp.ImageScan.Vulnerability = nil
	}
	return resp
}
