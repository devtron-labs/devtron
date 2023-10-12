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
	"github.com/caarlos0/env"
	bean2 "github.com/devtron-labs/devtron/api/bean"
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

type CiManagerService interface {
	CiPipelineConfigService
	CiMaterialConfigService
	AppArtifactManager
}

type CdManagerService interface {
	CdPipelineConfigService
	DevtronAppCMCSService
	DevtronAppStrategyService
	AppDeploymentTypeChangeManager
}

type PipelineBuilder interface {
	AppConfigService
	CiManagerService
	CdManagerService
}

type PipelineBuilderImpl struct {
	*AppConfigServiceImpl
	*BuildPipelineConfigServiceImpl
	*DeploymentPipelineConfigServiceImpl
	logger *zap.SugaredLogger
}

func NewPipelineBuilderImpl(
	pipelineConfigService *BuildPipelineConfigServiceImpl,
	deploymentConfigService *DeploymentPipelineConfigServiceImpl,
	appConfigService *AppConfigServiceImpl,
	logger *zap.SugaredLogger) *PipelineBuilderImpl {

	return &PipelineBuilderImpl{
		BuildPipelineConfigServiceImpl:      pipelineConfigService,
		DeploymentPipelineConfigServiceImpl: deploymentConfigService,
		AppConfigServiceImpl:                appConfigService,
		logger:                              logger,
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

//func (impl *PipelineBuilderImpl) getGitMaterialsForApp(appId int) ([]*bean.GitMaterial, error) {
//	materials, err := impl.materialRepo.FindByAppId(appId)
//	if err != nil {
//		impl.logger.Errorw("error in fetching materials for app", "appId", appId, "err", err)
//		return nil, err
//	}
//	var gitMaterials []*bean.GitMaterial
//
//	for _, material := range materials {
//		gitUrl := material.Url
//		if material.GitProvider.AuthMode == repository.AUTH_MODE_USERNAME_PASSWORD ||
//			material.GitProvider.AuthMode == repository.AUTH_MODE_ACCESS_TOKEN {
//			u, err := url.Parse(gitUrl)
//			if err != nil {
//				return nil, err
//			}
//			var password string
//			userName := material.GitProvider.UserName
//			if material.GitProvider.AuthMode == repository.AUTH_MODE_USERNAME_PASSWORD {
//				password = material.GitProvider.Password
//
//			} else if material.GitProvider.AuthMode == repository.AUTH_MODE_ACCESS_TOKEN {
//				password = material.GitProvider.AccessToken
//				if userName == "" {
//					userName = "devtron-boat"
//				}
//			}
//			if userName == "" || password == "" {
//				return nil, util.ApiError{}.ErrorfUser("invalid git credentials config")
//			}
//			u.User = url.UserPassword(userName, password)
//			gitUrl = u.String()
//		}
//		gitMaterial := &bean.GitMaterial{
//			Id:            material.Id,
//			Url:           gitUrl,
//			GitProviderId: material.GitProviderId,
//			Name:          material.Name[strings.Index(material.Name, "-")+1:],
//			CheckoutPath:  material.CheckoutPath,
//		}
//		gitMaterials = append(gitMaterials, gitMaterial)
//	}
//	return gitMaterials, nil
//} //no usage
//
//func (impl *PipelineBuilderImpl) isGitRepoUrlPresent(appId int) bool {
//	fetchedChart, err := impl.chartRepository.FindLatestByAppId(appId)
//
//	if err != nil || len(fetchedChart.GitRepoUrl) == 0 {
//		impl.logger.Errorw("error fetching git repo url or it is not present")
//		return false
//	}
//	return true
//} //no usage

//func (impl *PipelineBuilderImpl) handleNotHealthyAppsIfArgoDeploymentType(pipeline *pipelineConfig.Pipeline,
//	failedPipelines []*bean.DeploymentChangeStatus) ([]*bean.DeploymentChangeStatus, error) {
//
//	if pipeline.DeploymentAppType == bean.ArgoCd {
//		// check if app status is Healthy
//		status, err := impl.appStatusRepository.Get(pipeline.AppId, pipeline.EnvironmentId)
//
//		// case: missing status row in db
//		if len(status.Status) == 0 {
//			return failedPipelines, nil
//		}
//
//		// cannot delete the app from argocd if app status is Progressing
//		if err != nil || status.Status == "Progressing" {
//
//			healthCheckErr := errors.New("unable to fetch app status or app status is progressing")
//
//			impl.logger.Errorw(healthCheckErr.Error(),
//				"appId", pipeline.AppId,
//				"environmentId", pipeline.EnvironmentId,
//				"err", err)
//
//			failedPipelines = impl.handleFailedDeploymentAppChange(pipeline, failedPipelines, healthCheckErr.Error())
//
//			return failedPipelines, healthCheckErr
//		}
//		return failedPipelines, nil
//	}
//	return failedPipelines, nil
//} //no usage

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
