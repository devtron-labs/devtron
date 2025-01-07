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
	"bytes"
	"encoding/json"
	"fmt"
	bean2 "github.com/devtron-labs/common-lib/imageScan/bean"
	repository1 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster/environment"
	read2 "github.com/devtron-labs/devtron/pkg/cluster/read"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/adapter"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/read"
	repository3 "github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository"
	securityBean "github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"net/http"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type PolicyService interface {
	SavePolicy(request *bean.CreateVulnerabilityPolicyRequest, userId int32) (*bean.IdVulnerabilityPolicyResult, error)
	UpdatePolicy(updatePolicyParams bean.UpdatePolicyParams, userId int32) (*bean.IdVulnerabilityPolicyResult, error)
	DeletePolicy(id int, userId int32) (*bean.IdVulnerabilityPolicyResult, error)
	GetPolicies(policyLevel securityBean.PolicyLevel, clusterId, environmentId, appId int) (*bean.GetVulnerabilityPolicyResult, error)
	GetBlockedCVEList(cves []*repository3.CveStore, clusterId, envId, appId int, isAppstore bool) ([]*repository3.CveStore, error)
	VerifyImage(verifyImageRequest *VerifyImageRequest) (map[string][]*VerifyImageResponse, error)
	GetCvePolicy(id int, userId int32) (*repository3.CvePolicy, error)
	GetApplicablePolicy(clusterId, envId, appId int, isAppstore bool) (map[string]*repository3.CvePolicy, map[securityBean.Severity]*repository3.CvePolicy, error)
	HasBlockedCVE(cves []*repository3.CveStore, cvePolicy map[string]*repository3.CvePolicy, severityPolicy map[securityBean.Severity]*repository3.CvePolicy) bool
}
type PolicyServiceImpl struct {
	environmentService            environment.EnvironmentService
	logger                        *zap.SugaredLogger
	apRepository                  repository1.AppRepository
	pipelineOverride              chartConfig.PipelineOverrideRepository
	cvePolicyRepository           repository3.CvePolicyRepository
	clusterService                cluster.ClusterService
	PipelineRepository            pipelineConfig.PipelineRepository
	scanResultRepository          repository3.ImageScanResultRepository
	imageScanDeployInfoRepository repository3.ImageScanDeployInfoRepository
	imageScanObjectMetaRepository repository3.ImageScanObjectMetaRepository
	client                        *http.Client
	ciArtifactRepository          repository.CiArtifactRepository
	ciConfig                      *types.CiCdConfig
	imageScanHistoryReadService   read.ImageScanHistoryReadService
	cveStoreRepository            repository3.CveStoreRepository
	ciTemplateRepository          pipelineConfig.CiTemplateRepository
	ClusterReadService            read2.ClusterReadService
	transactionManager            sql.TransactionWrapper
}

func NewPolicyServiceImpl(environmentService environment.EnvironmentService,
	logger *zap.SugaredLogger,
	apRepository repository1.AppRepository,
	pipelineOverride chartConfig.PipelineOverrideRepository,
	cvePolicyRepository repository3.CvePolicyRepository,
	clusterService cluster.ClusterService,
	PipelineRepository pipelineConfig.PipelineRepository,
	scanResultRepository repository3.ImageScanResultRepository,
	imageScanDeployInfoRepository repository3.ImageScanDeployInfoRepository,
	imageScanObjectMetaRepository repository3.ImageScanObjectMetaRepository, client *http.Client,
	ciArtifactRepository repository.CiArtifactRepository, ciConfig *types.CiCdConfig,
	imageScanHistoryReadService read.ImageScanHistoryReadService,
	cveStoreRepository repository3.CveStoreRepository,
	ciTemplateRepository pipelineConfig.CiTemplateRepository,
	ClusterReadService read2.ClusterReadService,
	transactionManager sql.TransactionWrapper) *PolicyServiceImpl {
	return &PolicyServiceImpl{
		environmentService:            environmentService,
		logger:                        logger,
		apRepository:                  apRepository,
		pipelineOverride:              pipelineOverride,
		cvePolicyRepository:           cvePolicyRepository,
		clusterService:                clusterService,
		PipelineRepository:            PipelineRepository,
		scanResultRepository:          scanResultRepository,
		imageScanDeployInfoRepository: imageScanDeployInfoRepository,
		imageScanObjectMetaRepository: imageScanObjectMetaRepository,
		client:                        client,
		ciArtifactRepository:          ciArtifactRepository,
		ciConfig:                      ciConfig,
		imageScanHistoryReadService:   imageScanHistoryReadService,
		cveStoreRepository:            cveStoreRepository,
		ciTemplateRepository:          ciTemplateRepository,
		ClusterReadService:            ClusterReadService,
		transactionManager:            transactionManager,
	}
}

/*
1. get environment name by ns+ cluster
2. appName = releaseName - environmentName
3. verify app-env-image combination is correct
4. get policy for app(global, cluster, env, app)
5. get vulnblity for image
6. apply policy

if releaseName is empty(unable to determine app name) get policy for environment,
if env not found apply global policy
*/

type VerifyImageRequest struct {
	Images      []string
	ReleaseName string
	Namespace   string
	ClusterName string
	PodName     string
}

type VerifyImageResponse struct {
	Name         string
	Severity     string
	Package      string
	Version      string
	FixedVersion string
}

func (impl *PolicyServiceImpl) SendEventToClairUtility(event *bean2.ImageScanEvent) error {
	reqBody, err := json.Marshal(event)
	if err != nil {
		return err
	}
	impl.logger.Debugw("request", "body", string(reqBody))
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", impl.ciConfig.ImageScannerEndpoint, "scanner/image"), bytes.NewBuffer(reqBody))
	if err != nil {
		impl.logger.Errorw("error while writing test suites", "err", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := impl.client.Do(req)
	if err != nil {
		impl.logger.Errorw("error while UpdateJiraTransition request ", "err", err)
		return err
	}
	resp.Body.Close()
	impl.logger.Debugw("response from test suit create api", "status code", resp.StatusCode)
	return err
}

func (impl *PolicyServiceImpl) VerifyImage(verifyImageRequest *VerifyImageRequest) (map[string][]*VerifyImageResponse, error) {
	var clusterId, envId, appId int
	var isAppStore bool
	env, err := impl.environmentService.FindByNamespaceAndClusterName(verifyImageRequest.Namespace, verifyImageRequest.ClusterName)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching environment detail", "err", err)
		return nil, err
	} else if err == pg.ErrNoRows {
		cluster, err := impl.clusterService.FindOneActive(verifyImageRequest.ClusterName)
		if err != nil || cluster == nil {
			impl.logger.Errorw("error in getting cluster", "err", err)
			return nil, err
		}
		clusterId = cluster.Id
	} else {
		envId = env.Id
		clusterId = env.ClusterId
	}

	appName := strings.TrimSuffix(verifyImageRequest.ReleaseName, fmt.Sprintf("-%s", env.Name))
	app, err := impl.apRepository.FindActiveByName(appName)
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	} else if app != nil {
		appId = app.Id
		isAppStore = app.AppType == helper.ChartStoreApp
	} else {
		// np app do nothing
	}

	cvePolicy, severityPolicy, err := impl.GetApplicablePolicy(clusterId, envId, appId, isAppStore)
	if err != nil {
		impl.logger.Errorw("error in generating applicable policy", "err", err)
	}

	var objectType string
	var typeId int
	if appId != 0 && isAppStore {
		objectType = repository3.ScanObjectType_CHART
	} else if appId != 0 {
		objectType = repository3.ScanObjectType_APP
	} else {
		objectType = repository3.ScanObjectType_POD
	}

	imageBlockedCves := make(map[string][]*VerifyImageResponse)
	var scanResultsId []int
	scanResultsIdMap := make(map[int]int)
	for _, image := range verifyImageRequest.Images {

		// TODO - scan only if ci scan enabled

		scanHistory, err := impl.imageScanHistoryReadService.FindByImage(image)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in fetching scan history ", "err", err)
			return nil, err
		}
		if scanHistory != nil && scanHistory.Id == 0 && objectType != repository3.ScanObjectType_APP {
			scanEvent := &bean2.ImageScanEvent{Image: image, ImageDigest: "", PipelineId: 0, UserId: 1}
			dockerReg, err := impl.ciTemplateRepository.FindByAppId(app.Id)
			if err != nil {
				impl.logger.Errorw("error in fetching docker reg ", "err", err)
				return nil, err
			}
			scanEvent.DockerRegistryId = dockerReg.DockerRegistry.Id
			err = impl.SendEventToClairUtility(scanEvent)
			if err != nil {
				impl.logger.Errorw("error in send event to image scanner ", "err", err)
				return nil, err
			}
		}

		scanResults, err := impl.scanResultRepository.FindByImage(image)
		if err != nil {
			impl.logger.Errorw("error in fetching vulnerability ", "err", err)
			return nil, err
		}
		cveNameToScanResultPackageNameMapping := make(map[string]string)
		var cveStores []*repository3.CveStore
		for _, scanResult := range scanResults {
			cveNameToScanResultPackageNameMapping[scanResult.CveStoreName] = scanResult.Package
			cveStores = append(cveStores, &scanResult.CveStore)
			if _, ok := scanResultsIdMap[scanResult.ImageScanExecutionHistoryId]; !ok {
				scanResultsIdMap[scanResult.ImageScanExecutionHistoryId] = scanResult.ImageScanExecutionHistoryId
			}
		}
		blockedCves := repository3.EnforceCvePolicy(cveStores, cvePolicy, severityPolicy)
		impl.logger.Debugw("blocked cve for image", "image", image, "blocked", blockedCves)
		for _, cve := range blockedCves {
			vr := &VerifyImageResponse{
				Name:         cve.Name,
				Severity:     cve.GetSeverity().String(),
				Package:      cve.Package,
				Version:      cve.Version,
				FixedVersion: cve.FixedVersion,
			}
			if packageName, ok := cveNameToScanResultPackageNameMapping[cve.Name]; ok {
				if len(packageName) > 0 {
					// fetch package name from image_scan_execution_result table
					vr.Package = packageName
				}

			}
			imageBlockedCves[image] = append(imageBlockedCves[image], vr)
		}
	}

	if objectType == repository3.ScanObjectType_POD {
		// TODO create entry
		imageScanObjectMeta := &repository3.ImageScanObjectMeta{
			Name:   verifyImageRequest.PodName,
			Image:  strings.Join(verifyImageRequest.Images, ","),
			Active: true,
		}
		err = impl.imageScanObjectMetaRepository.Save(imageScanObjectMeta)
		if err != nil {
			impl.logger.Errorw("error in updating imageScanObjectMetaRepository info", "err", err)
			return imageBlockedCves, nil
		}
		typeId = imageScanObjectMeta.Id
	} else {
		typeId = appId
	}
	impl.logger.Debug(typeId)

	for _, v := range scanResultsIdMap {
		scanResultsId = append(scanResultsId, v)
	}

	if len(scanResultsId) == 0 {
		for _, image := range verifyImageRequest.Images {
			scanHistory, err := impl.imageScanHistoryReadService.FindByImage(image)
			if err != nil && err != pg.ErrNoRows {
				impl.logger.Errorw("error in fetching scan history ", "err", err)
				return nil, err
			}
			if scanHistory != nil && scanHistory.Id > 0 {
				scanResultsId = append(scanResultsId, scanHistory.Id)
			}
		}
	}

	if len(scanResultsId) > 0 {
		ot, err := impl.imageScanDeployInfoRepository.FindByTypeMetaAndTypeId(typeId, objectType) // todo insure this touple unique in db
		if err != nil && err != pg.ErrNoRows {
			return nil, err
		} else if err == pg.ErrNoRows {
			imageScanDeployInfo := &repository3.ImageScanDeployInfo{
				ImageScanExecutionHistoryId: scanResultsId,
				ScanObjectMetaId:            appId,
				ObjectType:                  objectType,
				EnvId:                       envId,
				ClusterId:                   clusterId,
				AuditLog: sql.AuditLog{
					CreatedOn: time.Now(),
					CreatedBy: 1,
					UpdatedOn: time.Now(),
					UpdatedBy: 1,
				},
			}
			err := impl.imageScanDeployInfoRepository.Save(imageScanDeployInfo)
			if err != nil {
				impl.logger.Errorw("error in adding deploy info", "err", err)
			}
		} else {
			impl.logger.Debugw("pt", "ot", ot)
		}
	}
	return imageBlockedCves, nil
}

func (impl *PolicyServiceImpl) GetApplicablePolicy(clusterId, envId, appId int, isAppstore bool) (map[string]*repository3.CvePolicy, map[securityBean.Severity]*repository3.CvePolicy, error) {

	var policyLevel securityBean.PolicyLevel
	if isAppstore && appId > 0 && envId > 0 && clusterId > 0 {
		policyLevel = securityBean.Environment
	} else if appId > 0 && envId > 0 && clusterId > 0 {
		policyLevel = securityBean.Application
	} else if envId > 0 && clusterId > 0 {
		policyLevel = securityBean.Environment
	} else if clusterId > 0 {
		policyLevel = securityBean.Cluster
	} else {
		// error in case of global or other policy
		return nil, nil, fmt.Errorf("policy not identified")
	}

	cvePolicy, severityPolicy, err := impl.getPolicies(policyLevel, clusterId, envId, appId)
	return cvePolicy, severityPolicy, err
}
func (impl *PolicyServiceImpl) getApplicablePolicies(policies []*repository3.CvePolicy) (map[string]*repository3.CvePolicy, map[securityBean.Severity]*repository3.CvePolicy) {
	cvePolicy := make(map[string][]*repository3.CvePolicy)
	severityPolicy := make(map[securityBean.Severity][]*repository3.CvePolicy)
	for _, policy := range policies {
		if policy.CVEStoreId != "" {
			cvePolicy[policy.CveStore.Name] = append(cvePolicy[policy.CveStore.Name], policy)
		} else {
			severityPolicy[*policy.Severity] = append(severityPolicy[*policy.Severity], policy)
		}
	}
	applicableCvePolicy := impl.getHighestPolicy(cvePolicy)
	applicableSeverityPolicy := impl.getHighestPolicyS(severityPolicy)
	return applicableCvePolicy, applicableSeverityPolicy
}

func (impl *PolicyServiceImpl) getHighestPolicy(allPolicies map[string][]*repository3.CvePolicy) map[string]*repository3.CvePolicy {
	applicablePolicies := make(map[string]*repository3.CvePolicy)
	for key, policies := range allPolicies {
		var applicablePolicy *repository3.CvePolicy
		for _, policy := range policies {
			if applicablePolicy == nil {
				applicablePolicy = policy
			} else {
				if policy.PolicyLevel() > applicablePolicy.PolicyLevel() {
					applicablePolicy = policy
				}
			}
		}
		applicablePolicies[key] = applicablePolicy
	}
	return applicablePolicies
}
func (impl *PolicyServiceImpl) getHighestPolicyS(allPolicies map[securityBean.Severity][]*repository3.CvePolicy) map[securityBean.Severity]*repository3.CvePolicy {
	applicablePolicies := make(map[securityBean.Severity]*repository3.CvePolicy)
	for key, policies := range allPolicies {
		var applicablePolicy *repository3.CvePolicy
		for _, policy := range policies {
			if applicablePolicy == nil {
				applicablePolicy = policy
			} else {
				if policy.PolicyLevel() > applicablePolicy.PolicyLevel() {
					applicablePolicy = policy
				}
			}
		}
		applicablePolicies[key] = applicablePolicy
	}
	return applicablePolicies
}

// -----------crud api----
/*
Severity/cveId
--
 global: na
 cluster: clusterId
 environment: environmentId
 application : appId, envId
--
action


res: id,
*/

func (impl *PolicyServiceImpl) parsePolicyAction(action string) (securityBean.PolicyAction, error) {
	var policyAction securityBean.PolicyAction
	if action == "allow" {
		policyAction = securityBean.Allow
	} else if action == "block" {
		policyAction = securityBean.Block
	} else if action == "inherit" {
		policyAction = securityBean.Inherit
	} else if action == "blockiffixed" {
		policyAction = securityBean.Blockiffixed
	} else {
		return securityBean.Inherit, fmt.Errorf("unsupported action %s", action)
	}
	return policyAction, nil
}

func (impl *PolicyServiceImpl) SavePolicy(request *bean.CreateVulnerabilityPolicyRequest, userId int32) (*bean.IdVulnerabilityPolicyResult, error) {
	tx, err := impl.transactionManager.StartTx()
	if err != nil {
		impl.logger.Errorw("error in creating transaction", "request", request, "err", err)
		return nil, err
	}
	defer func() {
		err = impl.transactionManager.RollbackTx(tx)
		if err != nil {
			impl.logger.Infow("error in rolling back transaction", "request", request, "err", err)
		}
	}()

	action, err := impl.parsePolicyAction(string(*request.Action))
	if err != nil {
		impl.logger.Errorw("error in parsing policy action", "action", request.Action, "err", err)
		return nil, err
	}
	var severity securityBean.Severity
	if len(request.Severity) > 0 {
		severity, err = securityBean.SeverityStringToEnumWithError(request.Severity)
		if err != nil {
			impl.logger.Errorw("error in converting string severity to enum", "severity", request.Severity, "err", err)
			return nil, err
		}
	} else {
		cveStore, err := impl.cveStoreRepository.FindByName(request.CveId)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("error in finding cveStore by cveId", "cveId", request.CveId, "err", err)
			return nil, err
		} else if util.IsErrNoRows(err) {
			errMessage := fmt.Sprintf("cve %s not found in our database", request.CveId)
			return nil, util.NewApiError(http.StatusNotFound, errMessage, errMessage)
		}
		severity = cveStore.GetSeverity()
	}
	// mark all previous policy on cveIds deleted before saving new one
	if len(request.CveId) > 0 {
		err = impl.softDeleteOldPoliciesIfExists(tx, request, userId)
		if err != nil {
			impl.logger.Errorw("error in soft deleting policies if exists", "request", request, "err", err)
			return nil, err
		}
	}
	policy, err := impl.cvePolicyRepository.SavePolicy(tx, adapter.BuildCvePolicy(request, action, severity, time.Now(), userId))
	if err != nil {
		impl.logger.Errorw("error in saving policy", "request", request, "err", err)
		return nil, err
	}
	err = impl.transactionManager.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing tx Create material", "request", request, "err", err)
		return nil, err
	}
	return &bean.IdVulnerabilityPolicyResult{Id: policy.Id}, nil
}

func (impl *PolicyServiceImpl) softDeleteOldPoliciesIfExists(tx *pg.Tx, request *bean.CreateVulnerabilityPolicyRequest, userId int32) error {
	policiesToDelete, err := impl.cvePolicyRepository.GetActiveByCveIdAndScope(request.CveId, request.EnvId, request.AppId, request.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in getting active policies by cveId and scope", "request", request, "err", err)
		return err
	}
	for _, policy := range policiesToDelete {
		policy.UpdateDeleted(true)
		policy.AuditLog.UpdateAuditLog(userId)
	}
	err = impl.cvePolicyRepository.UpdatePoliciesInBulk(tx, policiesToDelete)
	if err != nil {
		impl.logger.Errorw("error in deleting policies", "request", request, "err", err)
		return err
	}
	return nil
}

/*
1. policy id
2. action
*/
func (impl *PolicyServiceImpl) UpdatePolicy(updatePolicyParams bean.UpdatePolicyParams, userId int32) (*bean.IdVulnerabilityPolicyResult, error) {
	policyAction, err := impl.parsePolicyAction(updatePolicyParams.Action)
	if err != nil {
		return nil, err
	}
	if policyAction == securityBean.Inherit {
		return impl.DeletePolicy(updatePolicyParams.Id, userId)
	} else {
		policy, err := impl.cvePolicyRepository.GetById(updatePolicyParams.Id)
		if err != nil {
			impl.logger.Errorw("error in fetching policy ", "id", updatePolicyParams.Id)
			return nil, err
		}
		policy.Action = policyAction
		policy.UpdatedOn = time.Now()
		policy.UpdatedBy = userId
		policy, err = impl.cvePolicyRepository.UpdatePolicy(nil, policy)
		if err != nil {
			return nil, err
		} else {
			return &bean.IdVulnerabilityPolicyResult{Id: policy.Id}, nil
		}
	}
}

/*
input : policyId
output: id
*/
func (impl *PolicyServiceImpl) DeletePolicy(id int, userId int32) (*bean.IdVulnerabilityPolicyResult, error) {
	policy, err := impl.cvePolicyRepository.GetById(id)
	if err != nil {
		impl.logger.Errorw("error in fetching policy ", "id", id)
		return nil, err
	}
	if policy.Global && policy.CVEStoreId == "" {
		return nil, fmt.Errorf("global severity policy can't be changed to inherit")
	}
	policy.Deleted = true
	policy.UpdatedOn = time.Now()
	policy.UpdatedBy = userId
	policy, err = impl.cvePolicyRepository.UpdatePolicy(nil, policy)
	if err != nil {
		return nil, err
	} else {
		return &bean.IdVulnerabilityPolicyResult{Id: policy.Id}, nil
	}
}

/*
	global: na
	cluster: clusterId
	environment: environmentId
	application : appId, envId

res:
*/
func (impl *PolicyServiceImpl) GetPolicies(policyLevel securityBean.PolicyLevel, clusterId, environmentId, appId int) (*bean.GetVulnerabilityPolicyResult, error) {

	vulnerabilityPolicyResult := &bean.GetVulnerabilityPolicyResult{
		Level: bean.ResourceLevel(policyLevel.String()),
	}
	if policyLevel == securityBean.Global {
		cvePolicy, severityPolicy, err := impl.getPolicies(policyLevel, clusterId, environmentId, appId)
		if err != nil {
			return nil, err
		}
		vulnerabilityPolicy := impl.vulnerabilityPolicyBuilder(policyLevel, cvePolicy, severityPolicy)
		vulnerabilityPolicyResult.Policies = append(vulnerabilityPolicyResult.Policies, vulnerabilityPolicy)
	} else if policyLevel == securityBean.Cluster {
		if clusterId == 0 {
			return nil, fmt.Errorf("cluster id is missing")
		}
		// get cluster name
		cluster, err := impl.ClusterReadService.FindById(clusterId)
		if err != nil {
			impl.logger.Errorw("error in fetching cluster details", "id", clusterId, "err", err)
			return nil, err
		}

		cvePolicy, severityPolicy, err := impl.getPolicies(policyLevel, clusterId, environmentId, appId)
		if err != nil {
			return nil, err
		}
		vulnerabilityPolicy := impl.vulnerabilityPolicyBuilder(policyLevel, cvePolicy, severityPolicy)
		vulnerabilityPolicy.Name = cluster.ClusterName
		vulnerabilityPolicy.ClusterId = clusterId
		vulnerabilityPolicyResult.Policies = append(vulnerabilityPolicyResult.Policies, vulnerabilityPolicy)
	} else if policyLevel == securityBean.Environment {
		if environmentId == 0 {
			return nil, fmt.Errorf("environmentId is missing")
		}
		env, err := impl.environmentService.FindById(environmentId)
		if err != nil {
			impl.logger.Errorw("error in fetching env details", "id", environmentId, "err", err)
			return nil, err
		}
		clusterId = env.ClusterId
		cvePolicy, severityPolicy, err := impl.getPolicies(policyLevel, clusterId, environmentId, appId)
		if err != nil {
			return nil, err
		}
		vulnerabilityPolicy := impl.vulnerabilityPolicyBuilder(policyLevel, cvePolicy, severityPolicy)
		vulnerabilityPolicy.Name = env.Environment
		vulnerabilityPolicy.EnvId = env.Id
		vulnerabilityPolicyResult.Policies = append(vulnerabilityPolicyResult.Policies, vulnerabilityPolicy)
	} else if policyLevel == securityBean.Application {
		if appId == 0 {
			return nil, fmt.Errorf("appId is missing")
		}
		app, err := impl.apRepository.FindById(appId)
		if err != nil {
			impl.logger.Errorw("error in fetching app", "id", appId, "err", err)
			return nil, err
		}
		pipelines, err := impl.PipelineRepository.FindActiveByAppId(appId)
		if err != nil {
			impl.logger.Errorw("error in fetching pipelines", "id", appId, "err", err)
			return nil, err
		}
		var envId []*int
		for _, pipeline := range pipelines {
			envId = append(envId, &pipeline.EnvironmentId)
		}
		envs, err := impl.environmentService.FindByIds(envId)
		if err != nil {
			impl.logger.Errorw("Error in fetching env", "envId", envId, "err", err)
			return nil, err
		}
		for _, env := range envs {
			cvePolicy, severityPolicy, err := impl.getPolicies(policyLevel, env.ClusterId, env.Id, appId)
			if err != nil {
				return nil, err
			}
			vulnerabilityPolicy := impl.vulnerabilityPolicyBuilder(policyLevel, cvePolicy, severityPolicy)
			vulnerabilityPolicy.Name = fmt.Sprintf("%s/%s", app.AppName, env.Environment)
			vulnerabilityPolicy.EnvId = env.Id
			vulnerabilityPolicy.AppId = appId
			vulnerabilityPolicyResult.Policies = append(vulnerabilityPolicyResult.Policies, vulnerabilityPolicy)
		}
	}
	return vulnerabilityPolicyResult, nil
}

func (impl *PolicyServiceImpl) vulnerabilityPolicyBuilder(policyLevel securityBean.PolicyLevel, cvePolicy map[string]*repository3.CvePolicy, severityPolicy map[securityBean.Severity]*repository3.CvePolicy) *bean.VulnerabilityPolicy {
	vulnerabilityPolicy := &bean.VulnerabilityPolicy{}
	for _, v := range severityPolicy {

		severityPolicy := &bean.SeverityPolicy{
			Id: v.Id,
			Policy: &bean.VulnerabilityPermission{
				Action:      bean.VulnerabilityAction(v.Action.String()),
				Inherited:   v.PolicyLevel() != policyLevel,
				IsOverriden: v.PolicyLevel() == policyLevel,
			},
			PolicyOrigin: v.PolicyLevel().String(),
			Severity:     v.Severity.String(),
		}
		vulnerabilityPolicy.Severities = append(vulnerabilityPolicy.Severities, severityPolicy)
	}
	for _, v := range cvePolicy {
		cvePolicy := &bean.CvePolicy{
			SeverityPolicy: bean.SeverityPolicy{
				Id: v.Id,
				Policy: &bean.VulnerabilityPermission{
					Action:      bean.VulnerabilityAction(v.Action.String()),
					Inherited:   v.PolicyLevel() != policyLevel,
					IsOverriden: v.PolicyLevel() == policyLevel,
				},
				PolicyOrigin: v.PolicyLevel().String(),
				Severity:     v.Severity.String(),
			},
			Name: v.CVEStoreId,
		}
		vulnerabilityPolicy.Cves = append(vulnerabilityPolicy.Cves, cvePolicy)
	}
	return vulnerabilityPolicy
}

func (impl *PolicyServiceImpl) getPolicies(policyLevel securityBean.PolicyLevel, clusterId, environmentId, appId int) (map[string]*repository3.CvePolicy, map[securityBean.Severity]*repository3.CvePolicy, error) {
	var policies []*repository3.CvePolicy
	var err error
	if policyLevel == securityBean.Global {
		policies, err = impl.cvePolicyRepository.GetGlobalPolicies()
	} else if policyLevel == securityBean.Cluster {
		policies, err = impl.cvePolicyRepository.GetClusterPolicies(clusterId)
	} else if policyLevel == securityBean.Environment {
		policies, err = impl.cvePolicyRepository.GetEnvPolicies(clusterId, environmentId)
	} else if policyLevel == securityBean.Application {
		policies, err = impl.cvePolicyRepository.GetAppEnvPolicies(clusterId, environmentId, appId)
	} else {
		return nil, nil, fmt.Errorf("unsupported policy level: %s", policyLevel)
	}
	if err != nil {
		impl.logger.Errorw("error in fetching policy  ", "level", policyLevel, "err", err)
		return nil, nil, err
	}
	cvePolicy, severityPolicy := impl.getApplicablePolicies(policies)
	impl.logger.Debugw("policy identified ", "policyLevel", policyLevel)
	// transform and return
	return cvePolicy, severityPolicy, nil
}

func (impl *PolicyServiceImpl) GetBlockedCVEList(cves []*repository3.CveStore, clusterId, envId, appId int, isAppstore bool) ([]*repository3.CveStore, error) {

	cvePolicy, severityPolicy, err := impl.GetApplicablePolicy(clusterId, envId, appId, isAppstore)
	if err != nil {
		return nil, err
	}
	blockedCve := repository3.EnforceCvePolicy(cves, cvePolicy, severityPolicy)
	return blockedCve, nil
}

func (impl *PolicyServiceImpl) HasBlockedCVE(cves []*repository3.CveStore, cvePolicy map[string]*repository3.CvePolicy, severityPolicy map[securityBean.Severity]*repository3.CvePolicy) bool {
	for _, cve := range cves {
		if policy, ok := cvePolicy[cve.Name]; ok {
			if policy.Action == securityBean.Allow {
				continue
			} else if (policy.Action == securityBean.Block) || (policy.Action == securityBean.Blockiffixed && cve.FixedVersion != "") {
				return true
			}
		} else {
			if severityPolicy[cve.GetSeverity()] != nil && severityPolicy[cve.GetSeverity()].Action == securityBean.Allow {
				continue
			} else if severityPolicy[cve.GetSeverity()] != nil && (severityPolicy[cve.GetSeverity()].Action == securityBean.Block || (severityPolicy[cve.GetSeverity()].Action == securityBean.Blockiffixed && cve.FixedVersion != "")) {
				return true
			}
		}
	}
	return false
}

func (impl *PolicyServiceImpl) GetCvePolicy(id int, userId int32) (*repository3.CvePolicy, error) {
	policy, err := impl.cvePolicyRepository.GetById(id)
	if err != nil {
		impl.logger.Errorw("error in fetching policy ", "id", id)
		return nil, err
	}
	return policy, nil
}
