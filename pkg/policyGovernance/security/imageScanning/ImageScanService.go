/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package imageScanning

import (
	"context"
	bean4 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster/environment"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/helper/parser"
	repository3 "github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository"
	securityBean "github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/policyGovernance/security/scanTool/repository"
	"github.com/devtron-labs/devtron/pkg/workflow/cd/read"
	"go.opentelemetry.io/otel"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository"
	repository1 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ImageScanService interface {
	FetchAllDeployInfo(request *bean3.ImageScanRequest) ([]*repository3.ImageScanDeployInfo, error)
	FetchScanExecutionListing(request *bean3.ImageScanRequest, ids []int) (*bean3.ImageScanHistoryListingResponse, error)
	FetchExecutionDetailResult(request *bean3.ImageScanRequest) (*bean3.ImageScanExecutionDetail, error)
	FetchMinScanResultByAppIdAndEnvId(request *bean3.ImageScanRequest) (*bean3.ImageScanExecutionDetail, error)
	VulnerabilityExposure(request *repository3.VulnerabilityRequest) (*repository3.VulnerabilityExposureListingResponse, error)
	GetArtifactVulnerabilityStatus(ctx context.Context, request *bean2.VulnerabilityCheckRequest) (bool, error)
	IsImageScanExecutionCompleted(image, imageDigest string) (bool, error)
	// resource scanning functions below
	GetScanResults(resourceScanQueryParams *bean3.ResourceScanQueryParams) (parser.ResourceScanResponseDto, error)
	FilterDeployInfoByScannedArtifactsDeployedInEnv(deployInfoList []*repository3.ImageScanDeployInfo) ([]*repository3.ImageScanDeployInfo, error)
}

type ImageScanServiceImpl struct {
	Logger                                    *zap.SugaredLogger
	scanHistoryRepository                     repository3.ImageScanHistoryRepository
	scanResultRepository                      repository3.ImageScanResultRepository
	scanObjectMetaRepository                  repository3.ImageScanObjectMetaRepository
	cveStoreRepository                        repository3.CveStoreRepository
	imageScanDeployInfoRepository             repository3.ImageScanDeployInfoRepository
	userService                               user.UserService
	appRepository                             repository1.AppRepository
	envService                                environment.EnvironmentService
	ciArtifactRepository                      repository.CiArtifactRepository
	policyService                             PolicyService
	pipelineRepository                        pipelineConfig.PipelineRepository
	ciPipelineRepository                      pipelineConfig.CiPipelineRepository
	scanToolMetaDataRepository                repository2.ScanToolMetadataRepository
	scanToolExecutionHistoryMappingRepository repository3.ScanToolExecutionHistoryMappingRepository
	cvePolicyRepository                       repository3.CvePolicyRepository
	cdWorkflowReadService                     read.CdWorkflowReadService
}

func NewImageScanServiceImpl(Logger *zap.SugaredLogger, scanHistoryRepository repository3.ImageScanHistoryRepository,
	scanResultRepository repository3.ImageScanResultRepository, scanObjectMetaRepository repository3.ImageScanObjectMetaRepository,
	cveStoreRepository repository3.CveStoreRepository, imageScanDeployInfoRepository repository3.ImageScanDeployInfoRepository,
	userService user.UserService,
	appRepository repository1.AppRepository,
	envService environment.EnvironmentService, ciArtifactRepository repository.CiArtifactRepository, policyService PolicyService,
	pipelineRepository pipelineConfig.PipelineRepository, ciPipelineRepository pipelineConfig.CiPipelineRepository, scanToolMetaDataRepository repository2.ScanToolMetadataRepository, scanToolExecutionHistoryMappingRepository repository3.ScanToolExecutionHistoryMappingRepository,
	cvePolicyRepository repository3.CvePolicyRepository,
	cdWorkflowReadService read.CdWorkflowReadService) *ImageScanServiceImpl {
	return &ImageScanServiceImpl{Logger: Logger, scanHistoryRepository: scanHistoryRepository, scanResultRepository: scanResultRepository,
		scanObjectMetaRepository: scanObjectMetaRepository, cveStoreRepository: cveStoreRepository,
		imageScanDeployInfoRepository:             imageScanDeployInfoRepository,
		userService:                               userService,
		appRepository:                             appRepository,
		envService:                                envService,
		ciArtifactRepository:                      ciArtifactRepository,
		policyService:                             policyService,
		pipelineRepository:                        pipelineRepository,
		ciPipelineRepository:                      ciPipelineRepository,
		scanToolMetaDataRepository:                scanToolMetaDataRepository,
		scanToolExecutionHistoryMappingRepository: scanToolExecutionHistoryMappingRepository,
		cvePolicyRepository:                       cvePolicyRepository,
		cdWorkflowReadService:                     cdWorkflowReadService,
	}
}

func (impl ImageScanServiceImpl) FetchAllDeployInfo(request *bean3.ImageScanRequest) ([]*repository3.ImageScanDeployInfo, error) {
	deployedList, err := impl.imageScanDeployInfoRepository.FindAll()
	if err != nil {
		impl.Logger.Errorw("error while fetching scan execution result", "err", err)
		return nil, err
	}
	return deployedList, nil
}

func (impl ImageScanServiceImpl) FetchScanExecutionListing(request *bean3.ImageScanRequest, deployInfoIds []int) (*bean3.ImageScanHistoryListingResponse, error) {
	groupByList, err := impl.imageScanDeployInfoRepository.ScanListingWithFilter(&request.ImageScanFilter, request.Size, request.Offset, deployInfoIds)
	if err != nil {
		impl.Logger.Errorw("error while fetching scan execution result", "err", err)
		return nil, err
	}
	var ids []int
	totalCount := 0
	for _, item := range groupByList {
		totalCount = item.TotalCount
		ids = append(ids, item.Id)
	}
	if len(ids) == 0 {
		impl.Logger.Debugw("no image scan deploy info exists", "err", err)
		responseList := make([]*bean3.ImageScanHistoryResponse, 0)
		return &bean3.ImageScanHistoryListingResponse{ImageScanHistoryResponse: responseList}, nil
	}
	deployedList, err := impl.imageScanDeployInfoRepository.FindByIds(ids)
	if err != nil {
		impl.Logger.Errorw("error while fetching scan execution result", "err", err)
		return nil, err
	}

	groupByListMap := make(map[int]*repository3.ImageScanDeployInfo)
	executionHistoryIds := make([]int, 0, len(deployedList)*2)
	for _, item := range deployedList {
		groupByListMap[item.Id] = item
		executionHistoryIds = append(executionHistoryIds, item.ImageScanExecutionHistoryId...)
	}

	// fetching all execution history in bulk for updating last check time in case when no vul are found(no results will be saved)
	mapOfExecutionHistoryIdVsLastExecTime, err := impl.fetchImageExecutionHistoryMapByIds(executionHistoryIds)
	if err != nil {
		// made it non-blocking as it is not critical
		impl.Logger.Errorw("error encountered in FetchScanExecutionListing", "err", err)
	}

	var finalResponseList []*bean3.ImageScanHistoryResponse
	for _, item := range groupByList {
		imageScanHistoryResponse := &bean3.ImageScanHistoryResponse{}
		var lastChecked time.Time

		criticalCount := 0
		highCount := 0
		moderateCount := 0
		lowCount, unkownCount := 0, 0
		imageScanDeployInfo := groupByListMap[item.Id]
		if imageScanDeployInfo != nil {
			scanResultList, err := impl.scanResultRepository.FetchByScanExecutionIds(imageScanDeployInfo.ImageScanExecutionHistoryId)
			if err != nil && err != pg.ErrNoRows {
				impl.Logger.Errorw("error while fetching scan execution result", "err", err)
				//return nil, err
			}
			if err == pg.ErrNoRows {
				impl.Logger.Errorw("no scan execution data found, but has image scan deployed info", "err", err)
				return nil, err
			}

			for _, item := range scanResultList {
				lastChecked = item.ImageScanExecutionHistory.ExecutionTime
				criticalCount, highCount, moderateCount, lowCount, unkownCount = impl.updateCount(item.CveStore.GetSeverity(), criticalCount, highCount, moderateCount, lowCount, unkownCount)
			}
			// updating in case when no vul are found (no results)
			if lastChecked.IsZero() && len(imageScanDeployInfo.ImageScanExecutionHistoryId) > 0 && mapOfExecutionHistoryIdVsLastExecTime != nil {
				// len is always 1 of image execution history array
				historyId := imageScanDeployInfo.ImageScanExecutionHistoryId[0]
				if val, ok := mapOfExecutionHistoryIdVsLastExecTime[historyId]; ok {
					lastChecked = val
				}
			}
		}
		severityCount := &bean3.SeverityCount{
			Critical: criticalCount,
			High:     highCount,
			Medium:   moderateCount,
			Low:      lowCount,
			Unknown:  unkownCount,
		}
		imageScanHistoryResponse.ImageScanDeployInfoId = item.Id
		if imageScanDeployInfo != nil {
			imageScanHistoryResponse.LastChecked = &lastChecked
		}
		imageScanHistoryResponse.SeverityCount = severityCount
		if imageScanDeployInfo != nil {
			imageScanHistoryResponse.EnvId = imageScanDeployInfo.EnvId
		}
		imageScanHistoryResponse.Environment = item.EnvironmentName
		if len(request.AppName) > 0 || len(request.ObjectName) > 0 {
			imageScanHistoryResponse.Name = item.ObjectName
			imageScanHistoryResponse.Type = item.ObjectType
			if len(request.AppName) > 0 {
				imageScanHistoryResponse.AppId = item.ScanObjectMetaId
			}
		} else {
			if item.ObjectType == repository3.ScanObjectType_APP || item.ObjectType == repository3.ScanObjectType_CHART {
				app, err := impl.appRepository.FindById(item.ScanObjectMetaId)
				if err != nil && err != pg.ErrNoRows {
					return nil, err
				}
				if err == pg.ErrNoRows {
					continue
				}
				imageScanHistoryResponse.AppId = app.Id
				imageScanHistoryResponse.Name = app.AppName
				imageScanHistoryResponse.Type = item.ObjectType
			} else if item.ObjectType == repository3.ScanObjectType_POD {
				scanObjectMeta, err := impl.scanObjectMetaRepository.FindOne(item.ScanObjectMetaId)
				if err != nil && err != pg.ErrNoRows {
					return nil, err
				}
				if err == pg.ErrNoRows {
					continue
				}
				imageScanHistoryResponse.Name = scanObjectMeta.Name
				imageScanHistoryResponse.Type = item.ObjectType
			}
		}
		finalResponseList = append(finalResponseList, imageScanHistoryResponse)
	}

	finalResponse := &bean3.ImageScanHistoryListingResponse{
		Offset:                   request.Offset,
		Size:                     request.Size,
		ImageScanHistoryResponse: finalResponseList,
		Total:                    totalCount,
	}

	/*
		1) fetch from image_deployment_info, group by object id and type,
		2) on iteration collect image scan execution id and fetch its result and put in map
		3) merge 1st result with map
	*/

	return finalResponse, err
}

func (impl ImageScanServiceImpl) fetchImageExecutionHistoryMapByIds(historyIds []int) (map[int]time.Time, error) {
	mapOfExecutionHistoryIdVsExecutionTime := make(map[int]time.Time)
	if len(historyIds) > 0 {
		imageScanExecutionHistories, err := impl.scanHistoryRepository.FindByIds(historyIds)
		if err != nil {
			impl.Logger.Errorw("error encountered in fetchImageExecutionHistoryMapByIds", "historyIds", historyIds, "err", err)
			return nil, err
		}
		for _, h := range imageScanExecutionHistories {
			mapOfExecutionHistoryIdVsExecutionTime[h.Id] = h.ExecutionTime
		}
	}
	return mapOfExecutionHistoryIdVsExecutionTime, nil
}
func (impl ImageScanServiceImpl) fetchLatestArtifactIdForAppAndEnv(appId, envId int) (int, error) {
	cdWfRunner, err := impl.cdWorkflowReadService.FindLatestCdWorkflowRunnerByEnvironmentIdAndRunnerType(appId, envId, bean4.CD_WORKFLOW_TYPE_DEPLOY)
	if err != nil {
		impl.Logger.Errorw("error in fetching latest cd workflow runner of type DEPLOY for appId and envId", "appId", appId, "envId", envId, "err", err)
		return 0, err
	}
	var ciArtifactId int
	if cdWfRunner.CdWorkflow != nil && cdWfRunner.CdWorkflow.CiArtifact != nil {
		// for images built by us or provided in external ci case.
		ciArtifactId = cdWfRunner.CdWorkflow.CiArtifactId
		if cdWfRunner.CdWorkflow.CiArtifact.ParentCiArtifact > 0 {
			//linked ci case, parent ci artifact already scanned
			ciArtifactId = cdWfRunner.CdWorkflow.CiArtifact.ParentCiArtifact
		}
	}

	return ciArtifactId, nil
}

func (impl ImageScanServiceImpl) FetchExecutionDetailResult(request *bean3.ImageScanRequest) (*bean3.ImageScanExecutionDetail, error) {
	//var scanExecution *security.ImageScanExecutionHistory
	var scanExecutionIds []int
	var executionTime time.Time
	imageScanResponse := &bean3.ImageScanExecutionDetail{}
	isRegularApp := false
	if request.AppId > 0 && request.EnvId > 0 {
		artifactId, err := impl.fetchLatestArtifactIdForAppAndEnv(request.AppId, request.EnvId)
		if err != nil {
			impl.Logger.Errorw("error while fetching latest artifact for app and env", "appId", request.AppId, "envId", request.EnvId, "err", err)
			return nil, err
		}
		// setting latest artifact deploy on app and env
		request.ArtifactId = artifactId
	}
	if request.ImageScanDeployInfoId > 0 {
		// TODO: this flow is to be removed after sometime as not in use from now (23rd dec)
		// scan detail for deployed images
		scanDeployInfo, err := impl.imageScanDeployInfoRepository.FindOne(request.ImageScanDeployInfoId)
		if err != nil {
			impl.Logger.Errorw("error while fetching scan execution result", "err", err)
			return nil, err
		}

		scanExecutionIds = append(scanExecutionIds, scanDeployInfo.ImageScanExecutionHistoryId...)

		if scanDeployInfo.ObjectType == repository3.ScanObjectType_APP || scanDeployInfo.ObjectType == repository3.ScanObjectType_CHART {
			request.AppId = scanDeployInfo.ScanObjectMetaId
		} else if scanDeployInfo.ObjectType == repository3.ScanObjectType_POD {
			request.ObjectId = scanDeployInfo.ScanObjectMetaId
		}
		request.EnvId = scanDeployInfo.EnvId
		if scanDeployInfo.ObjectType == repository3.ScanObjectType_APP {
			isRegularApp = true
		}
		imageScanResponse.ObjectType = scanDeployInfo.ObjectType
		if !isRegularApp {
			imageScanResponse.ScanEnabled = true
			imageScanResponse.Scanned = true
		}
	} else if request.ArtifactId > 0 {
		// scan detail for artifact weather it is deployed or not, used here for ci build history
		ciArtifact, err := impl.ciArtifactRepository.Get(request.ArtifactId)
		if err != nil {
			impl.Logger.Errorw("error while fetching scan execution result", "err", err)
			return nil, err
		}
		imageScanResponse.Image = ciArtifact.Image
		scanExecution, err := impl.scanHistoryRepository.FindByImageAndDigestWithHistoryMapping(ciArtifact.ImageDigest, ciArtifact.Image)
		if err != nil && !util.IsErrNoRows(err) {
			impl.Logger.Errorw("error while fetching scan execution result", "err", err)
			return nil, err
		} else if util.IsErrNoRows(err) {
			// image not scanned
			imageScanResponse.Scanned = false
			return imageScanResponse, nil
		}
		ciPipeline, err := impl.ciPipelineRepository.FindByIdIncludingInActive(ciArtifact.PipelineId)
		if err != nil {
			impl.Logger.Errorw("error while fetching scan execution result", "err", err)
			return nil, err
		}
		imageScanResponse.AppId = ciPipeline.AppId

		scanExecutionIds = append(scanExecutionIds, scanExecution.Id)
		executionTime = scanExecution.ExecutionTime
		imageScanResponse.ScanEnabled = ciArtifact.ScanEnabled
		imageScanResponse.Scanned = ciArtifact.Scanned
		if scanExecution.ScanToolExecutionHistoryMapping != nil {
			imageScanResponse.Status = scanExecution.ScanToolExecutionHistoryMapping.State
		}
		if ciArtifact.ScanEnabled == false {
			impl.Logger.Debugw("returning without result as scan disabled for this artifact", "ciArtifact", ciArtifact)
			return imageScanResponse, nil
		}
		imageScanResponse.ObjectType = repository3.ScanObjectType_APP
	}

	var vulnerabilities []*bean3.Vulnerabilities
	var criticalCount, highCount, moderateCount, lowCount, unkownCount int
	var cveStores []*repository3.CveStore
	imageDigests := make(map[string]string)
	if len(scanExecutionIds) > 0 {
		//var imageScanResultFinal []*security.ImageScanExecutionResult
		imageScanResult, err := impl.scanResultRepository.FetchByScanExecutionIds(scanExecutionIds)
		if err != nil {
			impl.Logger.Errorw("error while fetching scan execution result", "err", err)
			return nil, err
		}

		for _, item := range imageScanResult {
			vulnerability := &bean3.Vulnerabilities{
				CVEName:  item.CveStore.Name,
				CVersion: item.CveStore.Version,
				FVersion: item.FixedVersion,
				Package:  item.CveStore.Package,
				Severity: item.CveStore.GetSeverity().String(),
				Target:   item.Target,
				Type:     item.Type,
				Class:    item.Class,
				//Permission: "BLOCK", TODO
			}
			// data already migrated hence get package, version and fixedVersion from image_scan_execution_result
			if len(item.Package) > 0 {
				// data already migrated hence get package from image_scan_execution_result
				vulnerability.Package = item.Package
			}
			if len(item.Version) > 0 {
				vulnerability.CVersion = item.Version
			}
			criticalCount, highCount, moderateCount, lowCount, unkownCount = impl.updateCount(item.CveStore.GetSeverity(), criticalCount, highCount, moderateCount, lowCount, unkownCount)
			vulnerabilities = append(vulnerabilities, vulnerability)
			cveStores = append(cveStores, &item.CveStore)
			if _, ok := imageDigests[item.ImageScanExecutionHistory.ImageHash]; !ok {
				imageDigests[item.ImageScanExecutionHistory.ImageHash] = item.ImageScanExecutionHistory.ImageHash
			}
			executionTime = item.ImageScanExecutionHistory.ExecutionTime
		}
		if len(imageScanResult) > 0 {
			imageScanResponse.ScanToolId = imageScanResult[0].ScanToolId
		} else {
			toolIdFromExecutionHistory, err := impl.getScanToolIdFromExecutionHistory(scanExecutionIds)
			if err != nil {
				impl.Logger.Errorw("error in getting scan tool id from exection history", "err", err, "")
				return nil, err
			}
			if toolIdFromExecutionHistory != -1 {
				imageScanResponse.ScanToolId = toolIdFromExecutionHistory
			}
		}
	}
	severityCount := &bean3.SeverityCount{
		Critical: criticalCount,
		High:     highCount,
		Medium:   moderateCount,
		Low:      lowCount,
	}
	imageScanResponse.ImageScanDeployInfoId = request.ImageScanDeployInfoId
	if len(vulnerabilities) == 0 {
		vulnerabilities = make([]*bean3.Vulnerabilities, 0)
	}
	imageScanResponse.Vulnerabilities = vulnerabilities
	imageScanResponse.SeverityCount = severityCount
	imageScanResponse.ExecutionTime = executionTime

	// scanned enabled or not only for when we don't ask for direct artifact id
	if request.ImageScanDeployInfoId > 0 && request.ArtifactId == 0 && isRegularApp {
		for _, v := range imageDigests {
			ciArtifact, err := impl.ciArtifactRepository.GetByImageDigest(v)
			if err != nil {
				impl.Logger.Errorw("error while fetching scan execution result", "err", err)
				return nil, err
			}
			imageScanResponse.ScanEnabled = ciArtifact.ScanEnabled
			imageScanResponse.Scanned = ciArtifact.Scanned
			imageScanResponse.Image = ciArtifact.Image
		}
	}

	if request.AppId > 0 && request.EnvId > 0 {
		app, err := impl.appRepository.FindById(request.AppId)
		if err != nil {
			impl.Logger.Errorw("error while fetching env", "err", err)
			return nil, err
		}
		imageScanResponse.AppId = request.AppId
		imageScanResponse.AppName = app.AppName
		env, err := impl.envService.FindById(request.EnvId)
		if err != nil {
			impl.Logger.Errorw("error while fetching env", "err", err)
			return nil, err
		}
		imageScanResponse.EnvId = request.EnvId
		imageScanResponse.EnvName = env.Environment

		blockCveList, err := impl.policyService.GetBlockedCVEList(cveStores, env.ClusterId, env.Id, request.AppId, app.AppType == helper.ChartStoreApp)
		if err != nil {
			impl.Logger.Errorw("error while fetching env", "err", err)
			//return nil, err
			//TODO - review @nishant
		}
		if blockCveList != nil {
			vulnerabilityPermissionMap := make(map[string]string)
			for _, cve := range blockCveList {
				vulnerabilityPermissionMap[cve.Name] = bean3.BLOCK
			}
			var updatedVulnerabilities []*bean3.Vulnerabilities
			for _, vulnerability := range imageScanResponse.Vulnerabilities {
				if _, ok := vulnerabilityPermissionMap[vulnerability.CVEName]; ok {
					vulnerability.Permission = bean3.BLOCK
				} else {
					vulnerability.Permission = bean3.WHITELISTED
				}
				updatedVulnerabilities = append(updatedVulnerabilities, vulnerability)
			}
			if len(updatedVulnerabilities) == 0 {
				updatedVulnerabilities = make([]*bean3.Vulnerabilities, 0)
			}
			imageScanResponse.Vulnerabilities = updatedVulnerabilities
		} else {
			for _, vulnerability := range imageScanResponse.Vulnerabilities {
				vulnerability.Permission = bean3.WHITELISTED
			}
		}
	}
	// setting scan tool name if scan tool id is present
	if imageScanResponse.ScanToolId > 0 {
		scanToolName, scanToolUrl, err := impl.scanToolMetaDataRepository.FindNameAndUrlById(imageScanResponse.ScanToolId)
		if err != nil {
			impl.Logger.Errorw("error in getting scan tool name by id", "scanToolId", imageScanResponse.ScanToolId, "err", err)
			return nil, err
		}
		imageScanResponse.ScanToolName = scanToolName
		imageScanResponse.ScanToolUrl = scanToolUrl
	}
	return imageScanResponse, nil
}

func (impl *ImageScanServiceImpl) FetchMinScanResultByAppIdAndEnvId(request *bean3.ImageScanRequest) (*bean3.ImageScanExecutionDetail, error) {
	//var scanExecution *security.ImageScanExecutionHistory
	var scanExecutionIds []int
	var executionTime time.Time

	var objectType []string
	objectType = append(objectType, repository3.ScanObjectType_APP, repository3.ScanObjectType_CHART)
	scanDeployInfo, err := impl.imageScanDeployInfoRepository.FetchByAppIdAndEnvId(request.AppId, request.EnvId, objectType)
	if err != nil && pg.ErrNoRows != err {
		impl.Logger.Errorw("error while fetching scan execution result", "err", err)
		return nil, err
	}
	if scanDeployInfo == nil || scanDeployInfo.Id == 0 || err == pg.ErrNoRows {
		return nil, err
	}
	scanExecutionIds = append(scanExecutionIds, scanDeployInfo.ImageScanExecutionHistoryId...)

	var criticalCount, highCount, moderateCount, lowCount, unkownCount, scantoolId int
	if len(scanExecutionIds) > 0 {
		imageScanResult, err := impl.scanResultRepository.FetchByScanExecutionIds(scanExecutionIds)
		if err != nil {
			impl.Logger.Errorw("error while fetching scan execution result", "err", err)
			return nil, err
		}
		for _, item := range imageScanResult {
			executionTime = item.ImageScanExecutionHistory.ExecutionTime
			criticalCount, highCount, moderateCount, lowCount, unkownCount = impl.updateCount(item.CveStore.GetSeverity(), criticalCount, highCount, moderateCount, lowCount, unkownCount)
		}
		if len(imageScanResult) > 0 {
			scantoolId = imageScanResult[0].ScanToolId
		} else {
			toolIdFromExecutionHistory, err := impl.getScanToolIdFromExecutionHistory(scanExecutionIds)
			if err != nil || toolIdFromExecutionHistory == -1 {
				impl.Logger.Errorw("error in getting scan tool id from exection history", "err", err, "")
				return nil, err
			}
			scantoolId = toolIdFromExecutionHistory
		}
	}
	severityCount := &bean3.SeverityCount{
		Critical: criticalCount,
		High:     highCount,
		Medium:   moderateCount,
		Low:      lowCount,
		Unknown:  unkownCount,
	}
	imageScanResponse := &bean3.ImageScanExecutionDetail{
		ImageScanDeployInfoId: scanDeployInfo.Id,
		SeverityCount:         severityCount,
		ExecutionTime:         executionTime,
		ObjectType:            scanDeployInfo.ObjectType,
		ScanEnabled:           true,
		Scanned:               true,
		ScanToolId:            scantoolId,
	}
	return imageScanResponse, nil
}

func (impl *ImageScanServiceImpl) getScanToolIdFromExecutionHistory(scanExecutionIds []int) (int, error) {
	scanToolHistoryMappings, err := impl.scanToolExecutionHistoryMappingRepository.GetAllScanHistoriesByExecutionHistoryIds(scanExecutionIds)
	if err != nil {
		if err == pg.ErrNoRows {
			impl.Logger.Errorw("got no rows for scanToolHistoryMappings", "err", err)
		} else {
			impl.Logger.Errorw("error in getting scanToolHistoryMappings", "err", err)
			return -1, err
		}
	}
	if len(scanToolHistoryMappings) > 0 {
		return scanToolHistoryMappings[0].ScanToolId, nil
	}
	return -1, err
}

func (impl *ImageScanServiceImpl) VulnerabilityExposure(request *repository3.VulnerabilityRequest) (*repository3.VulnerabilityExposureListingResponse, error) {
	vulnerabilityExposureListingResponse := &repository3.VulnerabilityExposureListingResponse{
		Offset: request.Offset,
		Size:   request.Size,
	}
	size := request.Size
	request.Size = 0
	count, err := impl.cveStoreRepository.VulnerabilityExposure(request)
	if err != nil {
		impl.Logger.Errorw("error while fetching vulnerability exposure", "err", err)
		return nil, err
	}
	request.Size = size
	vulnerabilityExposureListingResponse.Total = len(count)
	vulnerabilityExposureList, err := impl.cveStoreRepository.VulnerabilityExposure(request)
	if err != nil {
		impl.Logger.Errorw("error while fetching vulnerability exposure", "err", err)
		return nil, err
	}

	var cveStores []*repository3.CveStore
	cveStore, err := impl.cveStoreRepository.FindByName(request.CveName)
	if err != nil {
		impl.Logger.Errorw("error while fetching cve store", "err", err)
		return nil, err
	}

	envMap := make(map[int]bean.EnvironmentBean)
	environments, err := impl.envService.GetAllActive()
	if err != nil {
		impl.Logger.Errorw("error while fetching vulnerability exposure", "err", err)
		return nil, err
	}
	for _, item := range environments {
		envMap[item.Id] = item
	}

	cveStores = append(cveStores, cveStore)
	for _, item := range vulnerabilityExposureList {
		envId := 0
		if item.AppType == helper.ChartStoreApp {
			envId = item.ChartEnvId
		} else if item.AppType == helper.CustomApp {
			envId = item.PipelineEnvId
		}
		env := envMap[envId]
		item.EnvId = envId
		item.EnvName = env.Environment
		var appStore bool
		appStore = item.AppType == helper.ChartStoreApp
		blockCveList, err := impl.policyService.GetBlockedCVEList(cveStores, env.ClusterId, envId, item.AppId, appStore)
		if err != nil {
			impl.Logger.Errorw("error while fetching blocked list", "err", err)
			return nil, err
		}
		if len(blockCveList) > 0 {
			item.Blocked = true
		}
	}
	vulnerabilityExposureListingResponse.VulnerabilityExposure = vulnerabilityExposureList
	return vulnerabilityExposureListingResponse, nil
}

func (impl *ImageScanServiceImpl) CalculateSeverityCountInfo(vulnerabilities []*bean3.Vulnerabilities) *bean3.SeverityCount {
	diff := bean3.SeverityCount{}
	for _, vulnerability := range vulnerabilities {
		if vulnerability.IsCritical() {
			diff.Critical += 1
		} else if vulnerability.IsHigh() {
			diff.High += 1
		} else if vulnerability.IsMedium() {
			diff.Medium += 1
		} else if vulnerability.IsLow() {
			diff.Low += 1
		} else if vulnerability.IsUnknown() {
			diff.Unknown += 1
		}
	}
	return &diff
}

func (impl *ImageScanServiceImpl) GetArtifactVulnerabilityStatus(ctx context.Context, request *bean2.VulnerabilityCheckRequest) (bool, error) {
	isVulnerable := false
	if len(request.ImageDigest) > 0 {
		var cveStores []*repository3.CveStore
		_, span := otel.Tracer("orchestrator").Start(ctx, "scanResultRepository.FindByImageDigest")
		imageScanResult, err := impl.scanResultRepository.FindByImageDigest(request.ImageDigest)
		span.End()
		if err != nil && err != pg.ErrNoRows {
			impl.Logger.Errorw("error fetching image digest", "digest", request.ImageDigest, "err", err)
			return false, err
		}
		for _, item := range imageScanResult {
			cveStores = append(cveStores, &item.CveStore)
		}
		_, span = otel.Tracer("orchestrator").Start(ctx, "cvePolicyRepository.GetBlockedCVEList")
		if request.CdPipeline.Environment.ClusterId == 0 {
			envDetails, err := impl.envService.GetDetailsById(request.CdPipeline.EnvironmentId)
			if err != nil {
				impl.Logger.Errorw("error fetching cluster details by env, GetArtifactVulnerabilityStatus", "envId", request.CdPipeline.EnvironmentId, "err", err)
				return false, err
			}
			request.CdPipeline.Environment = *envDetails
		}
		blockCveList, err := impl.cvePolicyRepository.GetBlockedCVEList(cveStores, request.CdPipeline.Environment.ClusterId, request.CdPipeline.EnvironmentId, request.CdPipeline.AppId, false)
		span.End()
		if err != nil {
			impl.Logger.Errorw("error encountered in GetArtifactVulnerabilityStatus", "clusterId", request.CdPipeline.Environment.ClusterId, "envId", request.CdPipeline.EnvironmentId, "appId", request.CdPipeline.AppId, "err", err)
			return false, err
		}
		if len(blockCveList) > 0 {
			isVulnerable = true
		}
	}
	return isVulnerable, nil
}

func (impl ImageScanServiceImpl) updateCount(severity securityBean.Severity, criticalCount int, highCount int, moderateCount int, lowCount int, unkownCount int) (int, int, int, int, int) {
	if severity == securityBean.Critical {
		criticalCount += 1
	} else if severity == securityBean.High {
		highCount = highCount + 1
	} else if severity == securityBean.Medium {
		moderateCount = moderateCount + 1
	} else if severity == securityBean.Low {
		lowCount = lowCount + 1
	} else if severity == securityBean.Unknown {
		unkownCount += 1
	}
	return criticalCount, highCount, moderateCount, lowCount, unkownCount
}

func (impl ImageScanServiceImpl) IsImageScanExecutionCompleted(image, imageDigest string) (bool, error) {
	var isScanningCompleted bool
	allScanHistoryMappings, err := impl.scanToolExecutionHistoryMappingRepository.FetchScanHistoryMappingsUsingImageAndImageDigest(image, imageDigest)
	if err != nil {
		impl.Logger.Errorw("error in fetching all scan execution history mapping", "image", image, "imageDigest", imageDigest, "err", err)
		return false, err
	}

	for _, scanHistoryMapping := range allScanHistoryMappings {
		if scanHistoryMapping.State == repository3.ScanExecutionProcessStateCompleted {
			isScanningCompleted = true
		}
	}
	return isScanningCompleted, nil
}

func (impl ImageScanServiceImpl) FilterDeployInfoByScannedArtifactsDeployedInEnv(deployInfoList []*repository3.ImageScanDeployInfo) ([]*repository3.ImageScanDeployInfo, error) {
	filteredDeployInfoList := make([]*repository3.ImageScanDeployInfo, 0, len(deployInfoList))
	appVsEnvIdMap := make(map[int][]int, len(deployInfoList))
	for _, imageScanDeployInfo := range deployInfoList {
		if imageScanDeployInfo.IsObjectTypeApp() {
			if _, ok := appVsEnvIdMap[imageScanDeployInfo.ScanObjectMetaId]; ok {
				appVsEnvIdMap[imageScanDeployInfo.ScanObjectMetaId] = append(appVsEnvIdMap[imageScanDeployInfo.ScanObjectMetaId], imageScanDeployInfo.EnvId)
			} else {
				appVsEnvIdMap[imageScanDeployInfo.ScanObjectMetaId] = make([]int, 0, len(deployInfoList))
				appVsEnvIdMap[imageScanDeployInfo.ScanObjectMetaId] = append(appVsEnvIdMap[imageScanDeployInfo.ScanObjectMetaId], imageScanDeployInfo.EnvId)
			}
		}
	}
	appEnvToCiArtifactMap, ciArtifactIdToScannedMap, err := impl.fetchLatestArtifactMetadataDeployedOnAllEnvsAcrossApps(appVsEnvIdMap)
	if err != nil {
		impl.Logger.Errorw("error in fetching latest artifacts metadata for deployed on all envs and apps", "appVsEnvIdMap", appVsEnvIdMap, "err", err)
		return nil, err
	}
	for _, imageScanDeployInfo := range deployInfoList {
		if imageScanDeployInfo.IsObjectTypeApp() {
			if ciArtifactId, ok := appEnvToCiArtifactMap[bean3.NewAppEnvMetadata(imageScanDeployInfo.ScanObjectMetaId, imageScanDeployInfo.EnvId)]; ok {
				if ciArtifactIdToScannedMap[ciArtifactId] {
					// if this ci artifact is scanned then append in final resp
					filteredDeployInfoList = append(filteredDeployInfoList, imageScanDeployInfo)
				}
			}
		}

	}
	return filteredDeployInfoList, nil
}

func (impl ImageScanServiceImpl) fetchLatestArtifactMetadataDeployedOnAllEnvsAcrossApps(appVsEnvIdMap map[int][]int) (map[bean3.AppEnvMetadata]int, map[int]bool, error) {
	appEnvToCiArtifactMap := make(map[bean3.AppEnvMetadata]int)
	ciArtifactIdToScannedMap := make(map[int]bool)
	parentCiArtifactIds := make([]int, 0)
	cdwfArtifactsMetadata, err := impl.cdWorkflowReadService.FindLatestCdWorkflowRunnerArtifactMetadataForAppAndEnvIds(appVsEnvIdMap, bean4.CD_WORKFLOW_TYPE_DEPLOY)
	if err != nil {
		impl.Logger.Errorw("error in FindLatestCdWorkflowRunnersByAppAndEnvIds", "appVsEnvIdMap", appVsEnvIdMap, "err", err)
		return nil, nil, err
	}
	for _, item := range cdwfArtifactsMetadata {
		ciArtifactId := item.CiArtifactId
		if item.ParentCiArtifact > 0 {
			parentCiArtifactIds = append(parentCiArtifactIds, item.ParentCiArtifact)
			ciArtifactId = item.ParentCiArtifact
		} else {
			ciArtifactIdToScannedMap[ciArtifactId] = item.Scanned
		}
		appEnvToCiArtifactMap[bean3.NewAppEnvMetadata(item.AppId, item.EnvId)] = ciArtifactId
	}
	if len(parentCiArtifactIds) > 0 {
		parentCiArtifacts, err := impl.ciArtifactRepository.GetByIds(parentCiArtifactIds)
		if err != nil {
			impl.Logger.Errorw("error in getting artifacts by ids", "ids", parentCiArtifactIds, "err", err)
			return nil, nil, err
		}
		for _, parentCiArtifact := range parentCiArtifacts {
			// for linked ci case
			ciArtifactIdToScannedMap[parentCiArtifact.Id] = parentCiArtifact.Scanned
		}
	}
	return appEnvToCiArtifactMap, ciArtifactIdToScannedMap, nil
}
