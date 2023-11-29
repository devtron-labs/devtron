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

package pipeline

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/caarlos0/env"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"go.uber.org/zap"
)

const DashboardConfigMap = "dashboard-cm"
const SECURITY_SCANNING = "FORCE_SECURITY_SCANNING"

var DefaultPipelineValue = []byte(`{"ConfigMaps":{"enabled":false},"ConfigSecrets":{"enabled":false},"ContainerPort":[],"EnvVariables":[],"GracePeriod":30,"LivenessProbe":{},"MaxSurge":1,"MaxUnavailable":0,"MinReadySeconds":60,"ReadinessProbe":{},"Spec":{"Affinity":{"Values":"nodes","key":""}},"app":"13","appMetrics":false,"args":{},"autoscaling":{},"command":{"enabled":false,"value":[]},"containers":[],"dbMigrationConfig":{"enabled":false},"deployment":{"strategy":{"rolling":{"maxSurge":"25%","maxUnavailable":1}}},"deploymentType":"ROLLING","env":"1","envoyproxy":{"configMapName":"","image":"","resources":{"limits":{"cpu":"50m","memory":"50Mi"},"requests":{"cpu":"50m","memory":"50Mi"}}},"image":{"pullPolicy":"IfNotPresent"},"ingress":{},"ingressInternal":{"annotations":{},"enabled":false,"host":"","path":"","tls":[]},"initContainers":[],"pauseForSecondsBeforeSwitchActive":30,"pipelineName":"","prometheus":{"release":"monitoring"},"rawYaml":[],"releaseVersion":"1","replicaCount":1,"resources":{"limits":{"cpu":"0.05","memory":"50Mi"},"requests":{"cpu":"0.01","memory":"10Mi"}},"secret":{"data":{},"enabled":false},"server":{"deployment":{"image":"","image_tag":""}},"service":{"annotations":{},"type":"ClusterIP"},"servicemonitor":{"additionalLabels":{}},"tolerations":[],"volumeMounts":[],"volumes":[],"waitForSecondsBeforeScalingDown":30}`)

type EcrConfig struct {
	EcrPrefix string `env:"ECR_REPO_NAME_PREFIX" envDefault:"test/"`
}

func GetEcrConfig() (*EcrConfig, error) {
	cfg := &EcrConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

type DeploymentServiceTypeConfig struct {
	IsInternalUse bool `env:"IS_INTERNAL_USE" envDefault:"false"`
}

type SecurityConfig struct {
	//FORCE_SECURITY_SCANNING flag is being maintained in both dashboard and orchestrator CM's
	//TODO: rishabh will remove FORCE_SECURITY_SCANNING from dashboard's CM.
	ForceSecurityScanning bool `env:"FORCE_SECURITY_SCANNING" envDefault:"false"`
}

func GetDeploymentServiceTypeConfig() (*DeploymentServiceTypeConfig, error) {
	cfg := &DeploymentServiceTypeConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

type PipelineBuilder interface {
	DevtronAppConfigService
	CiPipelineConfigService
	CiMaterialConfigService
	AppArtifactManager
	CdPipelineConfigService
	DevtronAppCMCSService
	DevtronAppStrategyService
	AppDeploymentTypeChangeManager
}
type PipelineBuilderImpl struct {
	logger          *zap.SugaredLogger
	materialRepo    pipelineConfig.MaterialRepository
	chartRepository chartRepoRepository.ChartRepository
	CiPipelineConfigService
	CiMaterialConfigService
	AppArtifactManager
	DevtronAppCMCSService
	DevtronAppStrategyService
	AppDeploymentTypeChangeManager
	CdPipelineConfigService
	DevtronAppConfigService
}

func NewPipelineBuilderImpl(
	logger *zap.SugaredLogger,
	materialRepo pipelineConfig.MaterialRepository,
	chartRepository chartRepoRepository.ChartRepository,
	ciPipelineConfigService CiPipelineConfigService,
	ciMaterialConfigService CiMaterialConfigService,
	appArtifactManager AppArtifactManager,
	devtronAppCMCSService DevtronAppCMCSService,
	devtronAppStrategyService DevtronAppStrategyService,
	appDeploymentTypeChangeManager AppDeploymentTypeChangeManager,
	cdPipelineConfigService CdPipelineConfigService,
	devtronAppConfigService DevtronAppConfigService) *PipelineBuilderImpl {

	securityConfig := &SecurityConfig{}
	err := env.Parse(securityConfig)
	if err != nil {
		logger.Errorw("error in parsing securityConfig,setting  ForceSecurityScanning to default value", "defaultValue", securityConfig.ForceSecurityScanning, "err", err)
	}
	return &PipelineBuilderImpl{
		logger:                         logger,
		materialRepo:                   materialRepo,
		chartRepository:                chartRepository,
		CiPipelineConfigService:        ciPipelineConfigService,
		CiMaterialConfigService:        ciMaterialConfigService,
		AppArtifactManager:             appArtifactManager,
		DevtronAppCMCSService:          devtronAppCMCSService,
		DevtronAppStrategyService:      devtronAppStrategyService,
		AppDeploymentTypeChangeManager: appDeploymentTypeChangeManager,
		CdPipelineConfigService:        cdPipelineConfigService,
		DevtronAppConfigService:        devtronAppConfigService,
	}
}

// internal use only
const (
	teamIdKey                string = "teamId"
	teamNameKey              string = "teamName"
	appIdKey                 string = "appId"
	appNameKey               string = "appName"
	environmentIdKey         string = "environmentId"
	environmentNameKey       string = "environmentName"
	environmentIdentifierKey string = "environmentIdentifier"
)

func formatDate(t time.Time, layout string) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(layout)
}

/*
   1. create pipelineGroup
   2. save material (add credential provider support)

*/

func (impl *PipelineBuilderImpl) getGitMaterialsForApp(appId int) ([]*bean.GitMaterial, error) {
	materials, err := impl.materialRepo.FindByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching materials for app", "appId", appId, "err", err)
		return nil, err
	}
	var gitMaterials []*bean.GitMaterial

	for _, material := range materials {
		gitUrl := material.Url
		if material.GitProvider.AuthMode == repository.AUTH_MODE_USERNAME_PASSWORD ||
			material.GitProvider.AuthMode == repository.AUTH_MODE_ACCESS_TOKEN {
			u, err := url.Parse(gitUrl)
			if err != nil {
				return nil, err
			}
			var password string
			userName := material.GitProvider.UserName
			if material.GitProvider.AuthMode == repository.AUTH_MODE_USERNAME_PASSWORD {
				password = material.GitProvider.Password

			} else if material.GitProvider.AuthMode == repository.AUTH_MODE_ACCESS_TOKEN {
				password = material.GitProvider.AccessToken
				if userName == "" {
					userName = "devtron-boat"
				}
			}
			if userName == "" || password == "" {
				return nil, util.ApiError{}.ErrorfUser("invalid git credentials config")
			}
			u.User = url.UserPassword(userName, password)
			gitUrl = u.String()
		}
		gitMaterial := &bean.GitMaterial{
			Id:            material.Id,
			Url:           gitUrl,
			GitProviderId: material.GitProviderId,
			Name:          material.Name[strings.Index(material.Name, "-")+1:],
			CheckoutPath:  material.CheckoutPath,
		}
		gitMaterials = append(gitMaterials, gitMaterial)
	}
	return gitMaterials, nil
}

func getPatchStatus(err error) bean.CiPatchStatus {
	if err != nil {
		if err.Error() == string(bean.CI_PATCH_NOT_AUTHORIZED_MESSAGE) {
			return bean.CI_PATCH_NOT_AUTHORIZED
		}
		return bean.CI_PATCH_FAILED
	}
	return bean.CI_PATCH_SUCCESS
}

func getPatchMessage(err error) bean.CiPatchMessage {
	if err != nil {
		return bean.CiPatchMessage(err.Error())
	}
	return ""
}

func allDeploymentConfigTrue(deploymentConfig map[string]bool) bool {
	for _, value := range deploymentConfig {
		if !value {
			return false
		}
	}
	return true
}

func validDeploymentConfigReceived(deploymentConfig map[string]bool, deploymentTypeSent string) bool {
	for key, value := range deploymentConfig {
		if value && key == deploymentTypeSent {
			return true
		}
	}
	return false
}

func (impl *PipelineBuilderImpl) isGitRepoUrlPresent(appId int) bool {
	fetchedChart, err := impl.chartRepository.FindLatestByAppId(appId)

	if err != nil || len(fetchedChart.GitRepoUrl) == 0 {
		impl.logger.Errorw("error fetching git repo url or it is not present")
		return false
	}
	return true
}

type DeploymentType struct {
	Deployment Deployment `json:"deployment"`
}

type Deployment struct {
	Strategy map[string]interface{} `json:"strategy"`
}

type ConfigMapSecretsResponse struct {
	Maps    []bean2.ConfigSecretMap `json:"maps"`
	Secrets []bean2.ConfigSecretMap `json:"secrets"`
}

func parseMaterialInfo(materialInfo json.RawMessage, source string) (json.RawMessage, error) {
	if source != repository.GOCD && source != repository.CI_RUNNER && source != repository.WEBHOOK && source != repository.PRE_CD && source != repository.POST_CD && source != repository.POST_CI {
		return nil, fmt.Errorf("datasource: %s not supported", source)
	}
	var ciMaterials []repository.CiMaterialInfo
	err := json.Unmarshal(materialInfo, &ciMaterials)
	if err != nil {
		println("material info", materialInfo)
		println("unmarshal error for material info", "err", err)
	}
	var scmMapList []map[string]string

	for _, material := range ciMaterials {
		scmMap := map[string]string{}
		var url string
		if material.Material.Type == "git" {
			url = material.Material.GitConfiguration.URL
		} else if material.Material.Type == "scm" {
			url = material.Material.ScmConfiguration.URL
		} else {
			return nil, fmt.Errorf("unknown material type:%s ", material.Material.Type)
		}
		if material.Modifications != nil && len(material.Modifications) > 0 {
			_modification := material.Modifications[0]

			revision := _modification.Revision
			url = strings.TrimSpace(url)

			_webhookDataStr := ""
			_webhookDataByteArr, err := json.Marshal(_modification.WebhookData)
			if err == nil {
				_webhookDataStr = string(_webhookDataByteArr)
			}

			scmMap["url"] = url
			scmMap["revision"] = revision
			scmMap["modifiedTime"] = _modification.ModifiedTime
			scmMap["author"] = _modification.Author
			scmMap["message"] = _modification.Message
			scmMap["tag"] = _modification.Tag
			scmMap["webhookData"] = _webhookDataStr
			scmMap["branch"] = _modification.Branch
		}
		scmMapList = append(scmMapList, scmMap)
	}
	mInfo, err := json.Marshal(scmMapList)
	return mInfo, err
}

type TeamAppBean struct {
	ProjectId   int        `json:"projectId"`
	ProjectName string     `json:"projectName"`
	AppList     []*AppBean `json:"appList"`
}

type AppBean struct {
	Id     int    `json:"id"`
	Name   string `json:"name,notnull"`
	TeamId int    `json:"teamId,omitempty"`
}

type PipelineStrategiesResponse struct {
	PipelineStrategy []PipelineStrategy `json:"pipelineStrategy"`
}
type PipelineStrategy struct {
	DeploymentTemplate chartRepoRepository.DeploymentStrategy `json:"deploymentTemplate,omitempty"` //
	Config             json.RawMessage                        `json:"config"`
	Default            bool                                   `json:"default"`
}

func checkAppReleaseNotExist(err error) bool {
	// RELEASE_NOT_EXIST check for helm App and NOT_FOUND check for argo app
	return strings.Contains(err.Error(), bean.NOT_FOUND) || strings.Contains(err.Error(), bean.RELEASE_NOT_EXIST)
}
