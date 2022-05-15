/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/sql"
	"net/http"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type PolicyService interface {
	SavePolicy(request bean.CreateVulnerabilityPolicyRequest, userId int32) (*bean.IdVulnerabilityPolicyResult, error)
	UpdatePolicy(updatePolicyParams bean.UpdatePolicyParams, userId int32) (*bean.IdVulnerabilityPolicyResult, error)
	DeletePolicy(id int, userId int32) (*bean.IdVulnerabilityPolicyResult, error)
	GetPolicies(policyLevel security.PolicyLevel, clusterId, environmentId, appId int) (*bean.GetVulnerabilityPolicyResult, error)
	GetBlockedCVEList(cves []*security.CveStore, clusterId, envId, appId int, isAppstore bool) ([]*security.CveStore, error)
	VerifyImage(verifyImageRequest *VerifyImageRequest) (map[string][]*VerifyImageResponse, error)
	GetCvePolicy(id int, userId int32) (*security.CvePolicy, error)
	GetApplicablePolicy(clusterId, envId, appId int, isAppstore bool) (map[string]*security.CvePolicy, map[security.Severity]*security.CvePolicy, error)
	HasBlockedCVE(cves []*security.CveStore, cvePolicy map[string]*security.CvePolicy, severityPolicy map[security.Severity]*security.CvePolicy) bool
}
type PolicyServiceImpl struct {
	environmentService            cluster.EnvironmentService
	logger                        *zap.SugaredLogger
	apRepository                  app.AppRepository
	pipelineOverride              chartConfig.PipelineOverrideRepository
	cvePolicyRepository           security.CvePolicyRepository
	clusterService                cluster.ClusterService
	PipelineRepository            pipelineConfig.PipelineRepository
	scanResultRepository          security.ImageScanResultRepository
	imageScanDeployInfoRepository security.ImageScanDeployInfoRepository
	imageScanObjectMetaRepository security.ImageScanObjectMetaRepository
	client                        *http.Client
	ciArtifactRepository          repository.CiArtifactRepository
	ciConfig                      *pipeline.CiConfig
	scanHistoryRepository         security.ImageScanHistoryRepository
	cveStoreRepository            security.CveStoreRepository
	ciTemplateRepository          pipelineConfig.CiTemplateRepository
}

func NewPolicyServiceImpl(environmentService cluster.EnvironmentService,
	logger *zap.SugaredLogger,
	apRepository app.AppRepository,
	pipelineOverride chartConfig.PipelineOverrideRepository,
	cvePolicyRepository security.CvePolicyRepository,
	clusterService cluster.ClusterService,
	PipelineRepository pipelineConfig.PipelineRepository,
	scanResultRepository security.ImageScanResultRepository,
	imageScanDeployInfoRepository security.ImageScanDeployInfoRepository,
	imageScanObjectMetaRepository security.ImageScanObjectMetaRepository, client *http.Client,
	ciArtifactRepository repository.CiArtifactRepository, ciConfig *pipeline.CiConfig,
	scanHistoryRepository security.ImageScanHistoryRepository, cveStoreRepository security.CveStoreRepository,
	ciTemplateRepository pipelineConfig.CiTemplateRepository) *PolicyServiceImpl {
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
		scanHistoryRepository:         scanHistoryRepository,
		cveStoreRepository:            cveStoreRepository,
		ciTemplateRepository:          ciTemplateRepository,
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

type ScanEvent struct {
	Image            string `json:"image"`
	ImageDigest      string `json:"imageDigest"`
	AppId            int    `json:"appId"`
	EnvId            int    `json:"envId"`
	PipelineId       int    `json:"pipelineId"`
	CiArtifactId     int    `json:"ciArtifactId"`
	UserId           int    `json:"userId"`
	AccessKey        string `json:"accessKey"`
	SecretKey        string `json:"secretKey"`
	Token            string `json:"token"`
	AwsRegion        string `json:"awsRegion"`
	DockerRegistryId string `json:"dockerRegistryId"`
}

func (impl *PolicyServiceImpl) SendEventToClairUtility(event *ScanEvent) error {
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
		isAppStore = app.AppStore
	} else {
		//np app do nothing
	}

	cvePolicy, severityPolicy, err := impl.GetApplicablePolicy(clusterId, envId, appId, isAppStore)
	if err != nil {
		impl.logger.Errorw("error in generating applicable policy", "err", err)
	}

	var objectType string
	var typeId int
	if appId != 0 && isAppStore {
		objectType = security.ScanObjectType_CHART
	} else if appId != 0 {
		objectType = security.ScanObjectType_APP
	} else {
		objectType = security.ScanObjectType_POD
	}

	imageBlockedCves := make(map[string][]*VerifyImageResponse)
	var scanResultsId []int
	scanResultsIdMap := make(map[int]int)
	for _, image := range verifyImageRequest.Images {

		//TODO - scan only if ci scan enabled

		scanHistory, err := impl.scanHistoryRepository.FindByImage(image)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in fetching scan history ", "err", err)
			return nil, err
		}
		if scanHistory != nil && scanHistory.Id == 0 && objectType != security.ScanObjectType_APP {
			scanEvent := &ScanEvent{Image: image, ImageDigest: "", PipelineId: 0, UserId: 1}
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
		var cveStores []*security.CveStore
		for _, scanResult := range scanResults {
			cveStores = append(cveStores, &scanResult.CveStore)
			if _, ok := scanResultsIdMap[scanResult.ImageScanExecutionHistoryId]; !ok {
				scanResultsIdMap[scanResult.ImageScanExecutionHistoryId] = scanResult.ImageScanExecutionHistoryId
			}
		}
		blockedCves := impl.enforceCvePolicy(cveStores, cvePolicy, severityPolicy)
		impl.logger.Debugw("blocked cve for image", "image", image, "blocked", blockedCves)
		for _, cve := range blockedCves {
			vr := &VerifyImageResponse{
				Name:         cve.Name,
				Severity:     cve.Severity.String(),
				Package:      cve.Package,
				Version:      cve.Version,
				FixedVersion: cve.FixedVersion,
			}
			imageBlockedCves[image] = append(imageBlockedCves[image], vr)
		}
	}

	if objectType == security.ScanObjectType_POD {
		//TODO create entry
		imageScanObjectMeta := &security.ImageScanObjectMeta{
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
			scanHistory, err := impl.scanHistoryRepository.FindByImage(image)
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
		ot, err := impl.imageScanDeployInfoRepository.FindByTypeMetaAndTypeId(typeId, objectType) //todo insure this touple unique in db
		if err != nil && err != pg.ErrNoRows {
			return nil, err
		} else if err == pg.ErrNoRows {
			imageScanDeployInfo := &security.ImageScanDeployInfo{
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

//image(cve), appId, envId
func (impl *PolicyServiceImpl) enforceCvePolicy(cves []*security.CveStore, cvePolicy map[string]*security.CvePolicy, severityPolicy map[security.Severity]*security.CvePolicy) (blockedCVE []*security.CveStore) {

	for _, cve := range cves {
		if policy, ok := cvePolicy[cve.Name]; ok {
			if policy.Action == security.Allow {
				continue
			} else {
				blockedCVE = append(blockedCVE, cve)
			}
		} else {
			if severityPolicy[cve.Severity] != nil && severityPolicy[cve.Severity].Action == security.Allow {
				continue
			} else {
				blockedCVE = append(blockedCVE, cve)
			}
		}
	}
	return blockedCVE
}

func (impl *PolicyServiceImpl) GetApplicablePolicy(clusterId, envId, appId int, isAppstore bool) (map[string]*security.CvePolicy, map[security.Severity]*security.CvePolicy, error) {

	var policyLevel security.PolicyLevel
	if isAppstore && appId > 0 && envId > 0 && clusterId > 0 {
		policyLevel = security.Environment
	} else if appId > 0 && envId > 0 && clusterId > 0 {
		policyLevel = security.Application
	} else if envId > 0 && clusterId > 0 {
		policyLevel = security.Environment
	} else if clusterId > 0 {
		policyLevel = security.Cluster
	} else {
		//error in case of global or other policy
		return nil, nil, fmt.Errorf("policy not identified")
	}

	cvePolicy, severityPolicy, err := impl.getPolicies(policyLevel, clusterId, envId, appId)
	return cvePolicy, severityPolicy, err
}
func (impl *PolicyServiceImpl) getApplicablePolicies(policies []*security.CvePolicy) (map[string]*security.CvePolicy, map[security.Severity]*security.CvePolicy) {
	cvePolicy := make(map[string][]*security.CvePolicy)
	severityPolicy := make(map[security.Severity][]*security.CvePolicy)
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

func (impl *PolicyServiceImpl) getHighestPolicy(allPolicies map[string][]*security.CvePolicy) map[string]*security.CvePolicy {
	applicablePolicies := make(map[string]*security.CvePolicy)
	for key, policies := range allPolicies {
		var applicablePolicy *security.CvePolicy
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
func (impl *PolicyServiceImpl) getHighestPolicyS(allPolicies map[security.Severity][]*security.CvePolicy) map[security.Severity]*security.CvePolicy {
	applicablePolicies := make(map[security.Severity]*security.CvePolicy)
	for key, policies := range allPolicies {
		var applicablePolicy *security.CvePolicy
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

//-----------crud api----
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

func (impl *PolicyServiceImpl) parsePolicyAction(action string) (security.PolicyAction, error) {
	var policyAction security.PolicyAction
	if action == "allow" {
		policyAction = security.Allow
	} else if action == "block" {
		policyAction = security.Block
	} else if action == "inherit" {
		policyAction = security.Inherit
	} else {
		return security.Inherit, fmt.Errorf("unsupported action %s", action)
	}
	return policyAction, nil
}

func (impl *PolicyServiceImpl) SavePolicy(request bean.CreateVulnerabilityPolicyRequest, userId int32) (*bean.IdVulnerabilityPolicyResult, error) {
	isGlobal := false
	if request.ClusterId == 0 && request.EnvId == 0 && request.AppId == 0 {
		isGlobal = true
	}
	action, err := impl.parsePolicyAction(string(*request.Action))
	if err != nil {
		return nil, err
	}
	var severity security.Severity
	if len(request.Severity) > 0 {
		if request.Severity == "critical" {
			severity = security.Critical
		} else if request.Severity == "moderate" {
			severity = security.Moderate
		} else if request.Severity == "low" {
			severity = security.Low
		} else {
			return nil, fmt.Errorf("unsupported Severity %s", request.Severity)
		}
	} else {
		cveStore, err := impl.cveStoreRepository.FindByName(request.CveId)
		if err != nil {
			return nil, err
		}
		severity = cveStore.Severity
	}
	policy := &security.CvePolicy{
		Global:        isGlobal,
		ClusterId:     request.ClusterId,
		EnvironmentId: request.EnvId,
		AppId:         request.AppId,
		CVEStoreId:    request.CveId,
		Action:        action,
		Severity:      &severity,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
	policy, err = impl.cvePolicyRepository.SavePolicy(policy)
	if err != nil {
		impl.logger.Errorw("error in saving policy", "err", err)
		return nil, fmt.Errorf("error in saving policy")
	}
	return &bean.IdVulnerabilityPolicyResult{Id: policy.Id}, nil
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
	if policyAction == security.Inherit {
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
		policy, err = impl.cvePolicyRepository.UpdatePolicy(policy)
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
	policy, err = impl.cvePolicyRepository.UpdatePolicy(policy)
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
func (impl *PolicyServiceImpl) GetPolicies(policyLevel security.PolicyLevel, clusterId, environmentId, appId int) (*bean.GetVulnerabilityPolicyResult, error) {

	vulnerabilityPolicyResult := &bean.GetVulnerabilityPolicyResult{
		Level: bean.ResourceLevel(policyLevel.String()),
	}
	if policyLevel == security.Global {
		cvePolicy, severityPolicy, err := impl.getPolicies(policyLevel, clusterId, environmentId, appId)
		if err != nil {
			return nil, err
		}
		vulnerabilityPolicy := impl.vulnerabilityPolicyBuilder(policyLevel, cvePolicy, severityPolicy)
		vulnerabilityPolicyResult.Policies = append(vulnerabilityPolicyResult.Policies, vulnerabilityPolicy)
	} else if policyLevel == security.Cluster {
		if clusterId == 0 {
			return nil, fmt.Errorf("cluster id is missing")
		}
		//get cluster name
		cluster, err := impl.clusterService.FindById(clusterId)
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
	} else if policyLevel == security.Environment {
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
	} else if policyLevel == security.Application {
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
		for _, env := range envs {
			cvePolicy, severityPolicy, err := impl.getPolicies(policyLevel, clusterId, environmentId, appId)
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

func (impl *PolicyServiceImpl) vulnerabilityPolicyBuilder(policyLevel security.PolicyLevel, cvePolicy map[string]*security.CvePolicy, severityPolicy map[security.Severity]*security.CvePolicy) *bean.VulnerabilityPolicy {
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

func (impl *PolicyServiceImpl) getPolicies(policyLevel security.PolicyLevel, clusterId, environmentId, appId int) (map[string]*security.CvePolicy, map[security.Severity]*security.CvePolicy, error) {
	var policies []*security.CvePolicy
	var err error
	if policyLevel == security.Global {
		policies, err = impl.cvePolicyRepository.GetGlobalPolicies()
	} else if policyLevel == security.Cluster {
		policies, err = impl.cvePolicyRepository.GetClusterPolicies(clusterId)
	} else if policyLevel == security.Environment {
		policies, err = impl.cvePolicyRepository.GetEnvPolicies(clusterId, environmentId)
	} else if policyLevel == security.Application {
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
	//transform and return
	return cvePolicy, severityPolicy, nil
}

func (impl *PolicyServiceImpl) GetBlockedCVEList(cves []*security.CveStore, clusterId, envId, appId int, isAppstore bool) ([]*security.CveStore, error) {

	cvePolicy, severityPolicy, err := impl.GetApplicablePolicy(clusterId, envId, appId, isAppstore)
	if err != nil {
		return nil, err
	}
	blockedCve := impl.enforceCvePolicy(cves, cvePolicy, severityPolicy)
	return blockedCve, nil
}

func (impl *PolicyServiceImpl) HasBlockedCVE(cves []*security.CveStore, cvePolicy map[string]*security.CvePolicy, severityPolicy map[security.Severity]*security.CvePolicy) bool {
	for _, cve := range cves {
		if policy, ok := cvePolicy[cve.Name]; ok {
			if policy.Action == security.Allow {
				continue
			} else {
				return true
			}
		} else {
			if severityPolicy[cve.Severity] != nil && severityPolicy[cve.Severity].Action == security.Allow {
				continue
			} else {
				return true
			}
		}
	}
	return false
}

func (impl *PolicyServiceImpl) GetCvePolicy(id int, userId int32) (*security.CvePolicy, error) {
	policy, err := impl.cvePolicyRepository.GetById(id)
	if err != nil {
		impl.logger.Errorw("error in fetching policy ", "id", id)
		return nil, err
	}
	return policy, nil
}
