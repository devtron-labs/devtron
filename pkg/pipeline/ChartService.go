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
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"

	repository4 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	dirCopy "github.com/otiai10/copy"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	repository2 "github.com/argoproj/argo-cd/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/devtron/client/argocdServer/repository"
	"github.com/devtron-labs/devtron/internal/sql/models"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/ghodss/yaml"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

type TemplateRequest struct {
	Id                      int             `json:"id"  validate:"number"`
	AppId                   int             `json:"appId,omitempty"  validate:"number,required"`
	RefChartTemplate        string          `json:"refChartTemplate,omitempty"`
	RefChartTemplateVersion string          `json:"refChartTemplateVersion,omitempty"`
	ChartRepositoryId       int             `json:"chartRepositoryId,omitempty"`
	ValuesOverride          json.RawMessage `json:"valuesOverride,omitempty" validate:"required"` //json format user value
	DefaultAppOverride      json.RawMessage `json:"defaultAppOverride,omitempty"`                 //override values available
	ChartRefId              int             `json:"chartRefId,omitempty"  validate:"number"`
	Latest                  bool            `json:"latest"`
	IsAppMetricsEnabled     bool            `json:"isAppMetricsEnabled"`
	Schema                  json.RawMessage `json:"schema"`
	Readme                  string          `json:"readme"`
	UserId                  int32           `json:"-"`
}

type AppMetricEnableDisableRequest struct {
	AppId               int   `json:"appId,omitempty"`
	EnvironmentId       int   `json:"environmentId,omitempty"`
	IsAppMetricsEnabled bool  `json:"isAppMetricsEnabled"`
	UserId              int32 `json:"-"`
}

type ChartUpgradeRequest struct {
	ChartRefId int   `json:"chartRefId"  validate:"number"`
	All        bool  `json:"all"`
	AppIds     []int `json:"appIds"`
	UserId     int32 `json:"-"`
}

type PipelineConfigRequest struct {
	Id                   int             `json:"id"  validate:"number"`
	AppId                int             `json:"appId,omitempty"  validate:"number,required"`
	EnvConfigOverrideId  int             `json:"envConfigOverrideId,omitempty"`
	PipelineConfigValues json.RawMessage `json:"pipelineConfigValues,omitempty" validate:"required"` //json format user value
	PipelineId           int             `json:"PipelineId,omitempty"`
	Latest               bool            `json:"latest"`
	Previous             bool            `json:"previous"`
	EnvId                int             `json:"envId,omitempty"`
	ManualReviewed       bool            `json:"manualReviewed" validate:"required"`
	UserId               int32           `json:"-"`
}
type PipelineConfigRequestResponse struct {
	LatestPipelineConfigRequest   PipelineConfigRequest `json:"latestPipelineConfigRequest"`
	PreviousPipelineConfigRequest PipelineConfigRequest `json:"previousPipelineConfigRequest"`
}

type AppConfigResponse struct {
	//DefaultAppConfig  json.RawMessage `json:"defaultAppConfig"`
	//AppConfig         TemplateRequest            `json:"appConfig"`
	LatestAppConfig   TemplateRequest `json:"latestAppConfig"`
	PreviousAppConfig TemplateRequest `json:"previousAppConfig"`
}

type RefChartDir string
type DefaultChart string

type ChartService interface {
	Create(templateRequest TemplateRequest, ctx context.Context) (chart *TemplateRequest, err error)
	CreateChartFromEnvOverride(templateRequest TemplateRequest, ctx context.Context) (chart *TemplateRequest, err error)
	FindLatestChartForAppByAppId(appId int) (chartTemplate *TemplateRequest, err error)
	GetByAppIdAndChartRefId(appId int, chartRefId int) (chartTemplate *TemplateRequest, err error)
	GetAppOverrideForDefaultTemplate(chartRefId int) (map[string]interface{}, error)
	UpdateAppOverride(templateRequest *TemplateRequest) (*TemplateRequest, error)
	IsReadyToTrigger(appId int, envId int, pipelineId int) (IsReady, error)
	ChartRefAutocomplete() ([]chartRef, error)
	ChartRefAutocompleteForAppOrEnv(appId int, envId int) (*chartRefResponse, error)
	FindPreviousChartByAppId(appId int) (chartTemplate *TemplateRequest, err error)
	UpgradeForApp(appId int, chartRefId int, newAppOverride map[string]interface{}, userId int32, ctx context.Context) (bool, error)
	AppMetricsEnableDisable(appMetricRequest AppMetricEnableDisableRequest) (*AppMetricEnableDisableRequest, error)
	DeploymentTemplateValidate(templatejson interface{}, chartRefId int) (bool, error)
	JsonSchemaExtractFromFile(chartRefId int) (map[string]interface{}, error)
	GetSchemaAndReadmeForTemplateByChartRefId(chartRefId int) (schema []byte, readme []byte, err error)
	ExtractChartIfMissing(chartData []byte, refChartDir string, location string) (*ChartDataInfo, error)
	CheckChartExists(chartRefId int) error
	GetLocationFromChartNameAndVersion(chartName string, chartVersion string) string
	ValidateUploadedFileFormat(fileName string) error
	ReadChartMetaDataForLocation(chartDir string, fileName string) (string, string, error)
}
type ChartServiceImpl struct {
	chartRepository                  chartRepoRepository.ChartRepository
	logger                           *zap.SugaredLogger
	repoRepository                   chartRepoRepository.ChartRepoRepository
	chartTemplateService             util.ChartTemplateService
	pipelineGroupRepository          app.AppRepository
	mergeUtil                        util.MergeUtil
	repositoryService                repository.ServiceClient
	refChartDir                      RefChartDir
	defaultChart                     DefaultChart
	chartRefRepository               chartRepoRepository.ChartRefRepository
	envOverrideRepository            chartConfig.EnvConfigOverrideRepository
	pipelineConfigRepository         chartConfig.PipelineConfigRepository
	configMapRepository              chartConfig.ConfigMapRepository
	environmentRepository            repository4.EnvironmentRepository
	pipelineRepository               pipelineConfig.PipelineRepository
	appLevelMetricsRepository        repository3.AppLevelMetricsRepository
	envLevelAppMetricsRepository     repository3.EnvLevelAppMetricsRepository
	client                           *http.Client
	deploymentTemplateHistoryService history.DeploymentTemplateHistoryService
}

func NewChartServiceImpl(chartRepository chartRepoRepository.ChartRepository,
	logger *zap.SugaredLogger,
	chartTemplateService util.ChartTemplateService,
	repoRepository chartRepoRepository.ChartRepoRepository,
	pipelineGroupRepository app.AppRepository,
	refChartDir RefChartDir,
	defaultChart DefaultChart,
	mergeUtil util.MergeUtil,
	repositoryService repository.ServiceClient,
	chartRefRepository chartRepoRepository.ChartRefRepository,
	envOverrideRepository chartConfig.EnvConfigOverrideRepository,
	pipelineConfigRepository chartConfig.PipelineConfigRepository,
	configMapRepository chartConfig.ConfigMapRepository,
	environmentRepository repository4.EnvironmentRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	appLevelMetricsRepository repository3.AppLevelMetricsRepository,
	envLevelAppMetricsRepository repository3.EnvLevelAppMetricsRepository,
	client *http.Client,
	deploymentTemplateHistoryService history.DeploymentTemplateHistoryService) *ChartServiceImpl {
	return &ChartServiceImpl{
		chartRepository:                  chartRepository,
		logger:                           logger,
		chartTemplateService:             chartTemplateService,
		repoRepository:                   repoRepository,
		pipelineGroupRepository:          pipelineGroupRepository,
		mergeUtil:                        mergeUtil,
		refChartDir:                      refChartDir,
		defaultChart:                     defaultChart,
		repositoryService:                repositoryService,
		chartRefRepository:               chartRefRepository,
		envOverrideRepository:            envOverrideRepository,
		pipelineConfigRepository:         pipelineConfigRepository,
		configMapRepository:              configMapRepository,
		environmentRepository:            environmentRepository,
		pipelineRepository:               pipelineRepository,
		appLevelMetricsRepository:        appLevelMetricsRepository,
		envLevelAppMetricsRepository:     envLevelAppMetricsRepository,
		client:                           client,
		deploymentTemplateHistoryService: deploymentTemplateHistoryService,
	}
}

func (impl ChartServiceImpl) GetSchemaAndReadmeForTemplateByChartRefId(chartRefId int) ([]byte, []byte, error) {
	refChart, _, err, _ := impl.getRefChart(TemplateRequest{ChartRefId: chartRefId})
	if err != nil {
		impl.logger.Errorw("error in getting refChart", "err", err, "chartRefId", chartRefId)
		return nil, nil, err
	}
	var schemaByte []byte
	var readmeByte []byte
	schemaByte, err = ioutil.ReadFile(filepath.Clean(filepath.Join(refChart, "schema.json")))
	if err != nil {
		impl.logger.Errorw("error in reading schema.json file for refChart", "err", err, "chartRefId", chartRefId)
	}
	readmeByte, err = ioutil.ReadFile(filepath.Clean(filepath.Join(refChart, "README.md")))
	if err != nil {
		impl.logger.Errorw("error in reading readme file for refChart", "err", err, "chartRefId", chartRefId)
	}
	return schemaByte, readmeByte, nil
}

func (impl ChartServiceImpl) GetAppOverrideForDefaultTemplate(chartRefId int) (map[string]interface{}, error) {
	err := impl.CheckChartExists(chartRefId)
	if err != nil {
		impl.logger.Errorw("error in getting missing chart for chartRefId", "err", err, "chartRefId")
		return nil, err
	}

	refChart, _, err, _ := impl.getRefChart(TemplateRequest{ChartRefId: chartRefId})
	if err != nil {
		return nil, err
	}
	appOverrideByte, err := ioutil.ReadFile(filepath.Clean(filepath.Join(refChart, "app-values.yaml")))
	if err != nil {
		return nil, err
	}
	appOverrideByte, err = yaml.YAMLToJSON(appOverrideByte)
	if err != nil {
		return nil, err
	}
	envOverrideByte, err := ioutil.ReadFile(filepath.Clean(filepath.Join(refChart, "env-values.yaml")))
	if err != nil {
		return nil, err
	}
	envOverrideByte, err = yaml.YAMLToJSON(envOverrideByte)
	if err != nil {
		return nil, err
	}

	merged, err := impl.mergeUtil.JsonPatch(appOverrideByte, []byte(envOverrideByte))
	if err != nil {
		return nil, err
	}

	appOverride := json.RawMessage(merged)
	messages := make(map[string]interface{})
	messages["defaultAppOverride"] = appOverride
	return messages, nil
}

type AppMetricsEnabled struct {
	AppMetrics bool `json:"app-metrics"`
}

func (impl ChartServiceImpl) Create(templateRequest TemplateRequest, ctx context.Context) (*TemplateRequest, error) {
	err := impl.CheckChartExists(templateRequest.ChartRefId)
	if err != nil {
		impl.logger.Errorw("error in getting missing chart for chartRefId", "err", err, "chartRefId")
		return nil, err
	}
	chartMeta, err := impl.getChartMetaData(templateRequest)
	if err != nil {
		return nil, err
	}

	//save chart
	// 1. create chart, 2. push in repo, 3. add value of chart variable 4. save chart
	chartRepo, err := impl.getChartRepo(templateRequest)
	if err != nil {
		impl.logger.Errorw("error in fetching chart repo detail", "req", templateRequest)
		return nil, err
	}

	refChart, templateName, err, chartversion := impl.getRefChart(templateRequest)
	if err != nil {
		return nil, err
	}

	chartMajorVersion, chartMinorVersion, err := util2.ExtractChartVersion(chartversion)
	if err != nil {
		impl.logger.Errorw("chart version parsing", "err", err)
		return nil, err
	}

	existingChart, _ := impl.chartRepository.FindChartByAppIdAndRefId(templateRequest.AppId, templateRequest.ChartRefId)
	if existingChart != nil && existingChart.Id > 0 {
		return nil, fmt.Errorf("this reference chart already has added to appId %d refId %d", templateRequest.AppId, templateRequest.Id)
	}

	// STARTS
	currentLatestChart, err := impl.chartRepository.FindLatestChartForAppByAppId(templateRequest.AppId)
	if err != nil && pg.ErrNoRows != err {
		return nil, err
	}
	impl.logger.Debugw("current latest chart in db", "chartId", currentLatestChart.Id)
	if currentLatestChart.Id > 0 {
		impl.logger.Debugw("updating env and pipeline config which are currently latest in db", "chartId", currentLatestChart.Id)

		impl.logger.Debug("updating all other charts which are not latest but may be set previous true, setting previous=false")
		//step 2
		noLatestCharts, err := impl.chartRepository.FindNoLatestChartForAppByAppId(templateRequest.AppId)
		for _, noLatestChart := range noLatestCharts {
			if noLatestChart.Id != templateRequest.Id {

				noLatestChart.Latest = false // these are already false by d way
				noLatestChart.Previous = false
				err = impl.chartRepository.Update(noLatestChart)
				if err != nil {
					return nil, err
				}
			}
		}

		impl.logger.Debug("now going to update latest entry in db to false and previous flag = true")
		// now finally update latest entry in db to false and previous true
		currentLatestChart.Latest = false // these are already false by d way
		currentLatestChart.Previous = true
		err = impl.chartRepository.Update(currentLatestChart)
		if err != nil {
			return nil, err
		}
	}
	// ENDS

	impl.logger.Debug("now finally create new chart and make it latest entry in db and previous flag = true")

	version, err := impl.getNewVersion(chartRepo.Name, chartMeta.Name, refChart)
	chartMeta.Version = version
	if err != nil {
		return nil, err
	}
	chartValues, chartGitAttr, err := impl.chartTemplateService.CreateChart(chartMeta, refChart, templateName)
	if err != nil {
		return nil, err
	}
	override, err := templateRequest.ValuesOverride.MarshalJSON()
	if err != nil {
		return nil, err
	}
	valuesJson, err := yaml.YAMLToJSON([]byte(chartValues.Values))
	if err != nil {
		return nil, err
	}
	merged, err := impl.mergeUtil.JsonPatch(valuesJson, []byte(templateRequest.ValuesOverride))
	if err != nil {
		return nil, err
	}

	dst := new(bytes.Buffer)
	err = json.Compact(dst, override)
	if err != nil {
		return nil, err
	}
	override = dst.Bytes()

	err = impl.registerInArgo(chartGitAttr, ctx)
	if err != nil {
		return nil, err
	}

	chart := &chartRepoRepository.Chart{
		AppId:                   templateRequest.AppId,
		ChartRepoId:             chartRepo.Id,
		Values:                  string(merged),
		GlobalOverride:          string(override),
		ReleaseOverride:         chartValues.ReleaseOverrides, //image descriptor template
		PipelineOverride:        chartValues.PipelineOverrides,
		ImageDescriptorTemplate: chartValues.ImageDescriptorTemplate,
		ChartName:               chartMeta.Name,
		ChartRepo:               chartRepo.Name,
		ChartRepoUrl:            chartRepo.Url,
		ChartVersion:            chartMeta.Version,
		Status:                  models.CHARTSTATUS_NEW,
		Active:                  true,
		ChartLocation:           chartGitAttr.ChartLocation,
		GitRepoUrl:              chartGitAttr.RepoUrl,
		ReferenceTemplate:       templateName,
		ChartRefId:              templateRequest.ChartRefId,
		Latest:                  true,
		Previous:                false,
		AuditLog:                sql.AuditLog{CreatedBy: templateRequest.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: templateRequest.UserId},
	}

	err = impl.chartRepository.Save(chart)
	if err != nil {
		impl.logger.Errorw("error in saving chart ", "chart", chart, "error", err)
		//If found any error, rollback chart museum
		return nil, err
	}

	//creating history entry for deployment template
	//TODO : remove default value(false) app metrics flag for history when app metrics request is combined with template save request
	appLevelAppMetricsEnabled := false
	appLevelMetrics, err := impl.appLevelMetricsRepository.FindByAppId(templateRequest.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting app level metrics app level", "error", err)
	} else if err == nil {
		appLevelAppMetricsEnabled = appLevelMetrics.AppMetrics
	}
	err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromGlobalTemplate(chart, nil, appLevelAppMetricsEnabled)
	if err != nil {
		impl.logger.Errorw("error in creating entry for deployment template history", "err", err, "chart", chart)
		return nil, err
	}
	if !(chartMajorVersion >= 3 && chartMinorVersion >= 1) {
		appMetricsRequest := AppMetricEnableDisableRequest{UserId: templateRequest.UserId, AppId: templateRequest.AppId, IsAppMetricsEnabled: false}
		_, err = impl.updateAppLevelMetrics(&appMetricsRequest)
		if err != nil {
			impl.logger.Errorw("err while disable app metrics for lower versions", "err", err)
			return nil, err
		}
	}

	chartVal, err := impl.chartAdaptor(chart, appLevelMetrics)
	return chartVal, err
}

func (impl ChartServiceImpl) CreateChartFromEnvOverride(templateRequest TemplateRequest, ctx context.Context) (*TemplateRequest, error) {
	err := impl.CheckChartExists(templateRequest.ChartRefId)
	if err != nil {
		impl.logger.Errorw("error in getting missing chart for chartRefId", "err", err, "chartRefId")
		return nil, err
	}

	chartMeta, err := impl.getChartMetaData(templateRequest)
	if err != nil {
		return nil, err
	}

	appMetrics := templateRequest.IsAppMetricsEnabled

	//save chart
	// 1. create chart, 2. push in repo, 3. add value of chart variable 4. save chart
	chartRepo, err := impl.getChartRepo(templateRequest)
	if err != nil {
		impl.logger.Errorw("error in fetching chart repo detail", "req", templateRequest, "err", err)
		return nil, err
	}

	refChart, templateName, err, chartversion := impl.getRefChart(templateRequest)
	if err != nil {
		return nil, err
	}

	chartMajorVersion, chartMinorVersion, err := util2.ExtractChartVersion(chartversion)
	if err != nil {
		impl.logger.Errorw("chart version parsing", "err", err)
		return nil, err
	}

	if appMetrics && !(chartMajorVersion >= 3 && chartMinorVersion >= 1) {
		impl.logger.Error("cannot enable app metrics for older chart versions < 3.1.0")
		return nil, errors.New("cannot enable app metrics for older chart versions < 3.1.0")
	}

	impl.logger.Debug("now finally create new chart and make it latest entry in db and previous flag = true")

	version, err := impl.getNewVersion(chartRepo.Name, chartMeta.Name, refChart)
	chartMeta.Version = version
	if err != nil {
		return nil, err
	}
	chartValues, chartGitAttr, err := impl.chartTemplateService.CreateChart(chartMeta, refChart, templateName)

	if err != nil {
		return nil, err
	}
	override, err := templateRequest.ValuesOverride.MarshalJSON()
	if err != nil {
		return nil, err
	}
	valuesJson, err := yaml.YAMLToJSON([]byte(chartValues.Values))
	if err != nil {
		return nil, err
	}
	merged, err := impl.mergeUtil.JsonPatch(valuesJson, []byte(templateRequest.ValuesOverride))
	if err != nil {
		return nil, err
	}

	dst := new(bytes.Buffer)
	err = json.Compact(dst, override)
	if err != nil {
		return nil, err
	}
	override = dst.Bytes()

	err = impl.registerInArgo(chartGitAttr, ctx)
	if err != nil {
		return nil, err
	}

	chart := &chartRepoRepository.Chart{
		AppId:                   templateRequest.AppId,
		ChartRepoId:             chartRepo.Id,
		Values:                  string(merged),
		GlobalOverride:          string(override),
		ReleaseOverride:         chartValues.ReleaseOverrides,
		PipelineOverride:        chartValues.PipelineOverrides,
		ImageDescriptorTemplate: chartValues.ImageDescriptorTemplate,
		ChartName:               chartMeta.Name,
		ChartRepo:               chartRepo.Name,
		ChartRepoUrl:            chartRepo.Url,
		ChartVersion:            chartMeta.Version,
		Status:                  models.CHARTSTATUS_NEW,
		Active:                  true,
		ChartLocation:           chartGitAttr.ChartLocation,
		GitRepoUrl:              chartGitAttr.RepoUrl,
		ReferenceTemplate:       templateName,
		ChartRefId:              templateRequest.ChartRefId,
		Latest:                  false,
		Previous:                false,
		AuditLog:                sql.AuditLog{CreatedBy: templateRequest.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: templateRequest.UserId},
	}

	err = impl.chartRepository.Save(chart)
	if err != nil {
		impl.logger.Errorw("error in saving chart ", "chart", chart, "error", err)
		return nil, err
	}
	//creating history entry for deployment template
	//TODO : remove default value(false) app metrics flag for history when app metrics request is combined with template save request
	appLevelAppMetricsEnabled := false
	appLevelMetrics, err := impl.appLevelMetricsRepository.FindByAppId(templateRequest.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting app level metrics app level", "error", err)
	} else if err == nil {
		appLevelAppMetricsEnabled = appLevelMetrics.AppMetrics
	}
	err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromGlobalTemplate(chart, nil, appLevelAppMetricsEnabled)
	if err != nil {
		impl.logger.Errorw("error in creating entry for deployment template history", "err", err, "chart", chart)
		return nil, err
	}
	chartVal, err := impl.chartAdaptor(chart, appLevelMetrics)
	return chartVal, err
}

func (impl ChartServiceImpl) registerInArgo(chartGitAttribute *util.ChartGitAttribute, ctx context.Context) error {
	repo := &v1alpha1.Repository{
		Repo: chartGitAttribute.RepoUrl,
	}
	repo, err := impl.repositoryService.Create(ctx, &repository2.RepoCreateRequest{Repo: repo, Upsert: true})
	if err != nil {
		impl.logger.Errorw("error in creating argo Repository ", "err", err)
	}
	impl.logger.Infow("repo registered in argo", "name", chartGitAttribute.RepoUrl)
	return err
}

//converts db object to bean
func (impl ChartServiceImpl) chartAdaptor(chart *chartRepoRepository.Chart, appLevelMetrics *repository3.AppLevelMetrics) (*TemplateRequest, error) {
	var appMetrics bool
	if chart == nil || chart.Id == 0 {
		return &TemplateRequest{}, &util.ApiError{UserMessage: "no chart found"}
	}
	if appLevelMetrics != nil {
		appMetrics = appLevelMetrics.AppMetrics
	}
	return &TemplateRequest{
		RefChartTemplate:        chart.ReferenceTemplate,
		Id:                      chart.Id,
		AppId:                   chart.AppId,
		ChartRepositoryId:       chart.ChartRepoId,
		DefaultAppOverride:      json.RawMessage(chart.GlobalOverride),
		RefChartTemplateVersion: impl.getParentChartVersion(chart.ChartVersion),
		Latest:                  chart.Latest,
		ChartRefId:              chart.ChartRefId,
		IsAppMetricsEnabled:     appMetrics,
	}, nil
}

func (impl ChartServiceImpl) getChartMetaData(templateRequest TemplateRequest) (*chart.Metadata, error) {
	pg, err := impl.pipelineGroupRepository.FindById(templateRequest.AppId)
	if err != nil {
		impl.logger.Errorw("error in fetching pg", "id", templateRequest.AppId, "err", err)
	}
	metadata := &chart.Metadata{
		Name: pg.AppName,
	}
	return metadata, err
}
func (impl ChartServiceImpl) getRefChart(templateRequest TemplateRequest) (string, string, error, string) {
	var template string
	var version string

	if templateRequest.ChartRefId > 0 {
		chartRef, err := impl.chartRefRepository.FindById(templateRequest.ChartRefId)
		if err != nil {
			chartRef, err = impl.chartRefRepository.GetDefault()
			if err != nil {
				return "", "", err, ""
			}
		}
		template = chartRef.Location
		version = chartRef.Version
	} else {
		chartRef, err := impl.chartRefRepository.GetDefault()
		if err != nil {
			return "", "", err, ""
		}
		template = chartRef.Location
		version = chartRef.Version
	}

	//TODO VIKI- fetch from chart ref table
	chartPath := path.Join(string(impl.refChartDir), template)
	valid, err := chartutil.IsChartDir(chartPath)
	if err != nil || !valid {
		impl.logger.Errorw("invalid base chart", "dir", chartPath, "err", err)
		return "", "", err, ""
	}
	return chartPath, template, nil, version
}

func (impl ChartServiceImpl) getRefChartVersion(templateRequest TemplateRequest) (string, error) {
	var version string
	if templateRequest.ChartRefId > 0 {
		chartRef, err := impl.chartRefRepository.FindById(templateRequest.ChartRefId)
		if err != nil {
			chartRef, err = impl.chartRefRepository.GetDefault()
			if err != nil {
				return "", err
			}
		}
		version = chartRef.Version
	} else {
		chartRef, err := impl.chartRefRepository.GetDefault()
		if err != nil {
			return "", err
		}
		version = chartRef.Location
	}
	return version, nil
}

func (impl ChartServiceImpl) getChartRepo(templateRequest TemplateRequest) (*chartRepoRepository.ChartRepo, error) {
	if templateRequest.ChartRepositoryId == 0 {
		chartRepo, err := impl.repoRepository.GetDefault()
		if err != nil {
			impl.logger.Errorw("error in fetching default repo", "err", err)
			return nil, err
		}
		return chartRepo, err
	} else {
		chartRepo, err := impl.repoRepository.FindById(templateRequest.ChartRepositoryId)
		if err != nil {
			impl.logger.Errorw("error in fetching chart repo", "err", err, "id", templateRequest.ChartRepositoryId)
			return nil, err
		}
		return chartRepo, err
	}
}

func (impl ChartServiceImpl) getParentChartVersion(childVersion string) string {
	placeholders := strings.Split(childVersion, ".")
	return fmt.Sprintf("%s.%s.0", placeholders[0], placeholders[1])
}

//this method is not thread safe
func (impl ChartServiceImpl) getNewVersion(chartRepo, chartName, refChartLocation string) (string, error) {
	parentVersion, err := impl.chartTemplateService.GetChartVersion(refChartLocation)
	if err != nil {
		return "", err
	}
	placeholders := strings.Split(parentVersion, ".")
	if len(placeholders) != 3 || placeholders[2] != "0" {
		return "", fmt.Errorf("invalid parent chart version %s", parentVersion)
	}

	currentVersion, err := impl.chartRepository.FindCurrentChartVersion(chartRepo, chartName, placeholders[0]+"."+placeholders[1])
	if err != nil {
		return placeholders[0] + "." + placeholders[1] + ".1", nil
	}
	patch := strings.Split(currentVersion, ".")[2]
	count, err := strconv.ParseInt(patch, 10, 32)
	if err != nil {
		return "", err
	}
	count += 1

	return placeholders[0] + "." + placeholders[1] + "." + strconv.FormatInt(count, 10), nil
}

func (impl ChartServiceImpl) FindLatestChartForAppByAppId(appId int) (chartTemplate *TemplateRequest, err error) {
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching chart ", "appId", appId, "err", err)
		return nil, err
	}

	appMetrics, err := impl.appLevelMetricsRepository.FindByAppId(appId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching app-metrics", "appId", appId, "err", err)
		return nil, err
	}

	chartTemplate, err = impl.chartAdaptor(chart, appMetrics)
	return chartTemplate, err
}

func (impl ChartServiceImpl) GetByAppIdAndChartRefId(appId int, chartRefId int) (chartTemplate *TemplateRequest, err error) {
	chart, err := impl.chartRepository.FindChartByAppIdAndRefId(appId, chartRefId)
	if err != nil {
		impl.logger.Errorw("error in fetching chart ", "appId", appId, "err", err)
		return nil, err
	}
	appLevelMetrics, err := impl.appLevelMetricsRepository.FindByAppId(appId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching app metrics flag", "err", err)
		return nil, err
	}
	chartTemplate, err = impl.chartAdaptor(chart, appLevelMetrics)
	return chartTemplate, err
}

func (impl ChartServiceImpl) UpdateAppOverride(templateRequest *TemplateRequest) (*TemplateRequest, error) {

	template, err := impl.chartRepository.FindById(templateRequest.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching chart config", "id", templateRequest.Id, "err", err)
		return nil, err
	}

	chartMajorVersion, chartMinorVersion, err := util2.ExtractChartVersion(template.ChartVersion)
	if err != nil {
		impl.logger.Errorw("chart version parsing", "err", err)
		return nil, err
	}

	//STARTS
	currentLatestChart, err := impl.chartRepository.FindLatestChartForAppByAppId(templateRequest.AppId)
	if err != nil {
		return nil, err
	}
	if currentLatestChart.Id > 0 && currentLatestChart.Id == templateRequest.Id {

	} else if currentLatestChart.Id != templateRequest.Id {
		impl.logger.Debug("updating env and pipeline config which are currently latest in db", "chartId", currentLatestChart.Id)

		impl.logger.Debugw("updating request chart env config and pipeline config - making configs latest", "chartId", templateRequest.Id)

		impl.logger.Debug("updating all other charts which are not latest but may be set previous true, setting previous=false")
		//step 3
		noLatestCharts, err := impl.chartRepository.FindNoLatestChartForAppByAppId(templateRequest.AppId)
		for _, noLatestChart := range noLatestCharts {
			if noLatestChart.Id != templateRequest.Id {

				noLatestChart.Latest = false // these are already false by d way
				noLatestChart.Previous = false
				err = impl.chartRepository.Update(noLatestChart)
				if err != nil {
					return nil, err
				}
			}
		}

		impl.logger.Debug("now going to update latest entry in db to false and previous flag = true")
		// now finally update latest entry in db to false and previous true
		currentLatestChart.Latest = false // these are already false by d way
		currentLatestChart.Previous = true
		err = impl.chartRepository.Update(currentLatestChart)
		if err != nil {
			return nil, err
		}

	} else {
		return nil, nil
	}
	//ENDS

	impl.logger.Debug("now finally update request chart in db to latest and previous flag = false")
	values, err := impl.mergeUtil.JsonPatch([]byte(template.Values), templateRequest.ValuesOverride)
	if err != nil {
		return nil, err
	}
	template.Values = string(values)
	template.UpdatedOn = time.Now()
	template.UpdatedBy = templateRequest.UserId
	template.GlobalOverride = string(templateRequest.ValuesOverride)
	template.Latest = true
	template.Previous = false
	err = impl.chartRepository.Update(template)
	if err != nil {
		return nil, err
	}

	if !(chartMajorVersion >= 3 && chartMinorVersion >= 1) {
		appMetricRequest := AppMetricEnableDisableRequest{UserId: templateRequest.UserId, AppId: templateRequest.AppId, IsAppMetricsEnabled: false}
		_, err := impl.updateAppLevelMetrics(&appMetricRequest)
		if err != nil {
			impl.logger.Errorw("error in disable app metric flag", "error", err)
			return nil, err
		}
	}
	//creating history entry for deployment template
	//TODO : remove fetching and setting app metrics flag for history when app metrics update request is combined with template update request
	appLevelAppMetricsEnabled := false
	appLevelMetrics, err := impl.appLevelMetricsRepository.FindByAppId(template.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting app level metrics app level", "error", err)
	} else if err == nil {
		appLevelAppMetricsEnabled = appLevelMetrics.AppMetrics
	}
	err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromGlobalTemplate(template, nil, appLevelAppMetricsEnabled)
	if err != nil {
		impl.logger.Errorw("error in creating entry for deployment template history", "err", err, "chart", template)
		return nil, err
	}
	return templateRequest, nil
}

func (impl ChartServiceImpl) updateAppLevelMetrics(appMetricRequest *AppMetricEnableDisableRequest) (*repository3.AppLevelMetrics, error) {
	existingAppLevelMetrics, err := impl.appLevelMetricsRepository.FindByAppId(appMetricRequest.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in app metrics app level flag", "error", err)
		return nil, err
	}
	if existingAppLevelMetrics != nil && existingAppLevelMetrics.Id != 0 {
		existingAppLevelMetrics.AppMetrics = appMetricRequest.IsAppMetricsEnabled
		err := impl.appLevelMetricsRepository.Update(existingAppLevelMetrics)
		if err != nil {
			impl.logger.Errorw("failed to update app level metrics flag", "error", err)
			return nil, err
		}
		return existingAppLevelMetrics, nil
	} else {
		appLevelMetricsNew := &repository3.AppLevelMetrics{
			AppId:        appMetricRequest.AppId,
			AppMetrics:   appMetricRequest.IsAppMetricsEnabled,
			InfraMetrics: true,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				UpdatedOn: time.Now(),
				CreatedBy: appMetricRequest.UserId,
				UpdatedBy: appMetricRequest.UserId,
			},
		}
		err = impl.appLevelMetricsRepository.Save(appLevelMetricsNew)
		if err != nil {
			impl.logger.Errorw("error in saving app level metrics flag", "error", err)
			return appLevelMetricsNew, err
		}
		return appLevelMetricsNew, nil
	}
}

type IsReady struct {
	Flag    bool   `json:"flag"`
	Message string `json:"message"`
}

func (impl ChartServiceImpl) IsReadyToTrigger(appId int, envId int, pipelineId int) (IsReady, error) {
	isReady := IsReady{Flag: false}
	envOverride, err := impl.envOverrideRepository.ActiveEnvConfigOverride(appId, envId)
	if err != nil {
		impl.logger.Errorf("invalid state", "err", err, "envId", envId)
		isReady.Message = "Something went wrong"
		return isReady, err
	}

	if envOverride.Latest == false {
		impl.logger.Error("chart is updated for this app, may be environment or pipeline config is older")
		isReady.Message = "chart is updated for this app, may be environment or pipeline config is older"
		return isReady, nil
	}

	strategy, err := impl.pipelineConfigRepository.GetDefaultStrategyByPipelineId(pipelineId)
	if err != nil {
		impl.logger.Errorw("invalid state", "err", err, "req", strategy)
		if errors.IsNotFound(err) {
			isReady.Message = "no strategy found for request pipeline in this environment"
			return isReady, fmt.Errorf("no pipeline config found for request pipeline in this environment")
		}
		isReady.Message = "Something went wrong"
		return isReady, err
	}

	isReady.Flag = true
	isReady.Message = "Pipeline is well enough configured for trigger"
	return isReady, nil
}

type chartRef struct {
	Id      int    `json:"id"`
	Version string `json:"version"`
	Name    string `json:"name"`
}

type chartRefResponse struct {
	ChartRefs         []chartRef `json:"chartRefs"`
	LatestChartRef    int        `json:"latestChartRef"`
	LatestAppChartRef int        `json:"latestAppChartRef"`
	LatestEnvChartRef int        `json:"latestEnvChartRef,omitempty"`
}

type ChartYamlStruct struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type ChartDataInfo struct {
	ChartLocation   string `json:"chartLocation"`
	ChartName       string `json:"chartName"`
	ChartVersion    string `json:"chartVersion"`
	TemporaryFolder string `json:"temporaryFolder"`
}

func (impl ChartServiceImpl) ChartRefAutocomplete() ([]chartRef, error) {
	var chartRefs []chartRef
	results, err := impl.chartRefRepository.GetAll()
	if err != nil {
		impl.logger.Errorw("error in fetching chart config", "err", err)
		return chartRefs, err
	}

	for _, result := range results {
		chartRefs = append(chartRefs, chartRef{Id: result.Id, Version: result.Version})
	}

	return chartRefs, nil
}

func (impl ChartServiceImpl) ChartRefAutocompleteForAppOrEnv(appId int, envId int) (*chartRefResponse, error) {
	chartRefResponse := &chartRefResponse{}
	var chartRefs []chartRef
	results, err := impl.chartRefRepository.GetAll()
	if err != nil {
		impl.logger.Errorw("error in fetching chart config", "err", err)
		return chartRefResponse, err
	}

	var LatestAppChartRef int
	for _, result := range results {
		if len(result.Name) == 0 {
			result.Name = "Rollout Deployment"
		}
		chartRefs = append(chartRefs, chartRef{Id: result.Id, Version: result.Version, Name: result.Name})
		if result.Default == true {
			LatestAppChartRef = result.Id
		}
	}
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching latest chart", "err", err)
		return chartRefResponse, err
	}

	if envId > 0 {
		envOverride, err := impl.envOverrideRepository.FindLatestChartForAppByAppIdAndEnvId(appId, envId)
		if err != nil && !errors.IsNotFound(err) {
			impl.logger.Errorw("error in fetching latest chart", "err", err)
			return chartRefResponse, err
		}
		if envOverride != nil && envOverride.Chart != nil {
			chartRefResponse.LatestEnvChartRef = envOverride.Chart.ChartRefId
		} else {
			chartRefResponse.LatestEnvChartRef = chart.ChartRefId
		}
	}
	chartRefResponse.LatestAppChartRef = chart.ChartRefId
	chartRefResponse.ChartRefs = chartRefs
	chartRefResponse.LatestChartRef = LatestAppChartRef
	return chartRefResponse, nil
}

func (impl ChartServiceImpl) FindPreviousChartByAppId(appId int) (chartTemplate *TemplateRequest, err error) {
	chart, err := impl.chartRepository.FindPreviousChartByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching chart ", "appId", appId, "err", err)
		return nil, err
	}
	chartTemplate, err = impl.chartAdaptor(chart, nil)
	return chartTemplate, err
}

func (impl ChartServiceImpl) UpgradeForApp(appId int, chartRefId int, newAppOverride map[string]interface{}, userId int32, ctx context.Context) (bool, error) {

	currentChart, err := impl.FindLatestChartForAppByAppId(appId)
	if err != nil && pg.ErrNoRows != err {
		impl.logger.Error(err)
		return false, err
	}
	if pg.ErrNoRows == err {
		impl.logger.Errorw("no chart configured for this app", "appId", appId)
		return false, fmt.Errorf("no chart configured for this app, skip it for upgrade")
	}

	templateRequest := TemplateRequest{}
	templateRequest.ChartRefId = chartRefId
	templateRequest.AppId = appId
	templateRequest.ChartRepositoryId = currentChart.ChartRepositoryId
	templateRequest.DefaultAppOverride = newAppOverride["defaultAppOverride"].(json.RawMessage)
	templateRequest.ValuesOverride = currentChart.DefaultAppOverride
	templateRequest.UserId = userId

	upgradedChartReq, err := impl.Create(templateRequest, ctx)
	if err != nil {
		impl.logger.Error(err)
		return false, err
	}
	if upgradedChartReq == nil || upgradedChartReq.Id == 0 {
		impl.logger.Infow("unable to upgrade app", "appId", appId)
		return false, fmt.Errorf("unable to upgrade app, got no error on creating chart but unable to complete")
	}
	updatedChart, err := impl.chartRepository.FindById(upgradedChartReq.Id)
	if err != nil {
		return false, err
	}

	//STEP 2 - env upgrade
	impl.logger.Debugw("creating env and pipeline config for app", "appId", appId)
	//step 1
	envOverrides, err := impl.envOverrideRepository.GetEnvConfigByChartId(currentChart.Id)
	if err != nil && envOverrides == nil {
		return false, err
	}
	for _, envOverride := range envOverrides {

		//STEP 4 = create environment config
		env, err := impl.environmentRepository.FindById(envOverride.TargetEnvironment)
		if err != nil {
			return false, err
		}
		envOverrideNew := &chartConfig.EnvConfigOverride{
			Active:            true,
			ManualReviewed:    true,
			Status:            models.CHARTSTATUS_SUCCESS,
			EnvOverrideValues: string(envOverride.EnvOverrideValues),
			TargetEnvironment: envOverride.TargetEnvironment,
			ChartId:           updatedChart.Id,
			AuditLog:          sql.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now(), CreatedOn: time.Now(), CreatedBy: userId},
			Namespace:         env.Namespace,
			Latest:            true,
			Previous:          false,
		}
		err = impl.envOverrideRepository.Save(envOverrideNew)
		if err != nil {
			impl.logger.Errorw("error in creating env config", "data", envOverride, "error", err)
			return false, err
		}
		//creating history entry for deployment template
		//TODO : remove fetching and setting app metrics flag for history when app metrics update request is combined with template update request
		isAppMetricsEnabled := false
		envLevelAppMetrics, err := impl.envLevelAppMetricsRepository.FindByAppIdAndEnvId(appId, envOverrideNew.TargetEnvironment)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting env level app metrics", "err", err, "appId", appId, "envId", envOverrideNew.TargetEnvironment)
			return false, err
		} else if err == pg.ErrNoRows {
			appLevelAppMetrics, err := impl.appLevelMetricsRepository.FindByAppId(appId)
			if err != nil && err != pg.ErrNoRows {
				impl.logger.Errorw("error in getting app level app metrics", "err", err, "appId", appId)
				return false, err
			} else if err == nil {
				isAppMetricsEnabled = appLevelAppMetrics.AppMetrics
			}
		} else {
			isAppMetricsEnabled = *envLevelAppMetrics.AppMetrics
		}
		err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromEnvOverrideTemplate(envOverrideNew, nil, isAppMetricsEnabled, 0)
		if err != nil {
			impl.logger.Errorw("error in creating entry for env deployment template history", "err", err, "envOverride", envOverrideNew)
			return false, err
		}
	}

	return true, nil
}

//Deprecated
func (impl ChartServiceImpl) filterDeploymentTemplateForBackground(deploymentTemplate pipelineConfig.DeploymentTemplate, pipelineOverride string) (string, error) {
	var deploymentType DeploymentType
	err := json.Unmarshal([]byte(pipelineOverride), &deploymentType)
	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return "", err
	}
	if pipelineConfig.DEPLOYMENT_TEMPLATE_BLUE_GREEN == deploymentTemplate {
		newDeploymentType := DeploymentType{
			Deployment: Deployment{
				Strategy: Strategy{
					BlueGreen: deploymentType.Deployment.Strategy.BlueGreen,
				},
			},
		}
		pipelineOverrideBytes, err := json.Marshal(newDeploymentType)
		if err != nil {
			impl.logger.Errorw("err", "err", err)
			return "", err
		}
		pipelineOverride = string(pipelineOverrideBytes)
	} else if pipelineConfig.DEPLOYMENT_TEMPLATE_ROLLING == deploymentTemplate {
		newDeploymentType := DeploymentType{
			Deployment: Deployment{
				Strategy: Strategy{
					Rolling: deploymentType.Deployment.Strategy.Rolling,
				},
			},
		}
		pipelineOverrideBytes, err := json.Marshal(newDeploymentType)
		if err != nil {
			impl.logger.Errorw("err", "err", err)
			return "", err
		}
		pipelineOverride = string(pipelineOverrideBytes)
		return pipelineOverride, nil
	}
	return pipelineOverride, nil
}

func (impl ChartServiceImpl) AppMetricsEnableDisable(appMetricRequest AppMetricEnableDisableRequest) (*AppMetricEnableDisableRequest, error) {
	currentChart, err := impl.chartRepository.FindLatestChartForAppByAppId(appMetricRequest.AppId)
	if err != nil && pg.ErrNoRows != err {
		impl.logger.Error(err)
		return nil, err
	}
	if pg.ErrNoRows == err {
		impl.logger.Errorw("no chart configured for this app", "appId", appMetricRequest.AppId)
		err = &util.ApiError{
			HttpStatusCode:  http.StatusNotFound,
			InternalMessage: "no chart configured for this app",
			UserMessage:     "no chart configured for this app",
		}
		return nil, err
	}
	// validate app metrics compatibility
	if appMetricRequest.IsAppMetricsEnabled == true {
		chartMajorVersion, chartMinorVersion, err := util2.ExtractChartVersion(currentChart.ChartVersion)
		if err != nil {
			impl.logger.Errorw("chart version parsing", "err", err)
			return nil, err
		}

		if !(chartMajorVersion >= 3 && chartMinorVersion >= 1) {
			err = &util.ApiError{
				InternalMessage: "chart version in not compatible for app metrics",
				UserMessage:     "chart version in not compatible for app metrics",
			}
			return nil, err
		}
	}
	//update or create app level app metrics
	appLevelMetrics, err := impl.updateAppLevelMetrics(&appMetricRequest)
	if err != nil {
		impl.logger.Errorw("error in saving app level metrics flag", "error", err)
		return nil, err
	}
	//creating history entry for deployment template
	err = impl.deploymentTemplateHistoryService.CreateDeploymentTemplateHistoryFromGlobalTemplate(currentChart, nil, appMetricRequest.IsAppMetricsEnabled)
	if err != nil {
		impl.logger.Errorw("error in creating entry for deployment template history", "err", err, "chart", currentChart)
		return nil, err
	}
	if appLevelMetrics.Id > 0 {
		return &appMetricRequest, nil
	}
	return nil, err
}

const memoryPattern = `"1000Mi" or "1Gi"`
const cpuPattern = `"50m" or "0.05"`
const cpu = "cpu"
const memory = "memory"

func (impl ChartServiceImpl) DeploymentTemplateValidate(templatejson interface{}, chartRefId int) (bool, error) {
	schemajson, err := impl.JsonSchemaExtractFromFile(chartRefId)
	if err != nil {
		impl.logger.Errorw("Json Schema not found err, FindJsonSchema", "err", err)
		return true, nil
	}
	//if err != nil && chartRefId >= 9 {
	//	impl.logger.Errorw("Json Schema not found err, FindJsonSchema", "err", err)
	//	return false, err
	//} else if err != nil {
	//	impl.logger.Errorw("Json Schema not found err, FindJsonSchema", "err", err)
	//	return true, nil
	//}
	schemaLoader := gojsonschema.NewGoLoader(schemajson)
	documentLoader := gojsonschema.NewGoLoader(templatejson)
	marshalTemplatejson, err := json.Marshal(templatejson)
	if err != nil {
		impl.logger.Errorw("json template marshal err, DeploymentTemplateValidate", "err", err)
		return false, err
	}
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		impl.logger.Errorw("result validate err, DeploymentTemplateValidate", "err", err)
		return false, err
	}
	if result.Valid() {
		var dat map[string]interface{}
		if err := json.Unmarshal(marshalTemplatejson, &dat); err != nil {
			impl.logger.Errorw("json template unmarshal err, DeploymentTemplateValidate", "err", err)
			return false, err
		}

		_, err := util2.CompareLimitsRequests(dat)
		if err != nil {
			impl.logger.Errorw("LimitRequestCompare err, DeploymentTemplateValidate", "err", err)
			return false, err
		}
		_, err = util2.AutoScale(dat)
		if err != nil {
			impl.logger.Errorw("LimitRequestCompare err, DeploymentTemplateValidate", "err", err)
			return false, err
		}

		return true, nil
	} else {
		var stringerror string
		for _, err := range result.Errors() {
			impl.logger.Errorw("result err, DeploymentTemplateValidate", "err", err.Details())
			if err.Details()["format"] == cpu {
				stringerror = stringerror + err.Field() + ": Format should be like " + cpuPattern + "\n"
			} else if err.Details()["format"] == memory {
				stringerror = stringerror + err.Field() + ": Format should be like " + memoryPattern + "\n"
			} else {
				stringerror = stringerror + err.String() + "\n"
			}
		}
		return false, errors.New(stringerror)
	}
}

func (impl ChartServiceImpl) JsonSchemaExtractFromFile(chartRefId int) (map[string]interface{}, error) {
	refChartDir, _, err, _ := impl.getRefChart(TemplateRequest{ChartRefId: chartRefId})
	if err != nil {
		impl.logger.Errorw("refChartDir Not Found err, JsonSchemaExtractFromFile", err)
		return nil, err
	}
	fileStatus := filepath.Join(refChartDir, "schema.json")
	if _, err := os.Stat(fileStatus); os.IsNotExist(err) {
		impl.logger.Errorw("Schema File Not Found err, JsonSchemaExtractFromFile", err)
		return nil, err
	} else {
		jsonFile, err := os.Open(fileStatus)
		if err != nil {
			impl.logger.Errorw("jsonfile open err, JsonSchemaExtractFromFile", "err", err)
			return nil, err
		}
		byteValueJsonFile, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			impl.logger.Errorw("byteValueJsonFile read err, JsonSchemaExtractFromFile", "err", err)
			return nil, err
		}

		var schemajson map[string]interface{}
		err = json.Unmarshal([]byte(byteValueJsonFile), &schemajson)
		if err != nil {
			impl.logger.Errorw("Unmarshal err in byteValueJsonFile, DeploymentTemplateValidate", "err", err)
			return nil, err
		}
		return schemajson, nil
	}
}

func (impl ChartServiceImpl) CheckChartExists(chartRefId int) error {
	chartRef, err := impl.chartRefRepository.FindById(chartRefId)
	if err != nil {
		impl.logger.Errorw("error in finding ref chart by id", "err", err)
		return err
	}
	refChartDir := filepath.Clean(filepath.Join(string(impl.refChartDir), chartRef.Location))
	if _, err := os.Stat(refChartDir); os.IsNotExist(err) {
		chartInfo, err := impl.ExtractChartIfMissing(chartRef.ChartData, string(impl.refChartDir), chartRef.Location)
		if chartInfo.TemporaryFolder != "" {
			err1 := os.RemoveAll(chartInfo.TemporaryFolder)
			if err1 != nil {
				impl.logger.Errorw("error in deleting temp dir ", "err", err)
			}
		}
		return err
	}
	return nil
}

func (impl *ChartServiceImpl) GetLocationFromChartNameAndVersion(chartName string, chartVersion string) string {
	var chartLocation string

	chartname := strings.ReplaceAll(chartName, ".", "-")
	chartname = strings.ReplaceAll(chartname, " ", "_")
	chartversion := strings.ReplaceAll(chartVersion, ".", "-")
	if !strings.Contains(chartname, chartversion) {
		chartLocation = chartname + "_" + chartversion
	} else {
		chartLocation = chartname
	}
	return chartLocation
}

func (impl *ChartServiceImpl) ValidateUploadedFileFormat(fileName string) error {
	if !strings.HasSuffix(fileName, ".tar.gz") {
		return errors.New("unsupported format")
	}
	return nil
}

func (impl ChartServiceImpl) ReadChartMetaDataForLocation(chartDir string, fileName string) (string, string, error) {
	chartLocation := filepath.Clean(filepath.Join(chartDir, fileName))

	chartYamlPath := filepath.Clean(filepath.Join(chartLocation, "Chart.yaml"))
	if _, err := os.Stat(chartYamlPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("Chart.yaml file not present in the directory")
	}

	data, err := ioutil.ReadFile(chartYamlPath)
	if err != nil {
		impl.logger.Errorw("failed reading data from file", "err", err)
		return "", "", err
	}
	//println(data)
	var chartYaml ChartYamlStruct
	err = yaml.Unmarshal(data, &chartYaml)
	if err != nil {
		impl.logger.Errorw("Unmarshal error of yaml file", "err", err)
		return "", "", err
	}
	if chartYaml.Name == "" || chartYaml.Version == "" {
		impl.logger.Errorw("Missing values in yaml file either name or version", "err", err)
		return "", "", errors.New("Missing values in yaml file either name or version")
	}
	return chartYaml.Name, chartYaml.Version, nil
}

func (impl ChartServiceImpl) ExtractChartIfMissing(chartData []byte, refChartDir string, location string) (*ChartDataInfo, error) {
	binaryDataReader := bytes.NewReader(chartData)
	dir := impl.chartTemplateService.GetDir()
	chartInfo := &ChartDataInfo{
		ChartName:       "",
		ChartVersion:    "",
		ChartLocation:   "",
		TemporaryFolder: "",
	}
	temporaryChartWorkingDir := filepath.Clean(filepath.Join(refChartDir, dir))
	err := os.MkdirAll(temporaryChartWorkingDir, os.ModePerm)
	if err != nil {
		impl.logger.Errorw("error in creating directory, CallbackConfigMap", "err", err)
		return chartInfo, err
	}
	chartInfo.TemporaryFolder = temporaryChartWorkingDir
	err = util2.ExtractTarGz(binaryDataReader, temporaryChartWorkingDir)
	if err != nil {
		impl.logger.Errorw("error in extracting binary data of charts", "err", err)
		return chartInfo, err
	}

	var chartLocation string
	var chartName string
	var chartVersion string
	var fileName string

	files, err := ioutil.ReadDir(temporaryChartWorkingDir)
	if err != nil {
		impl.logger.Errorw("error in reading err dir", "err", err)
		return chartInfo, err
	}

	fileName = files[0].Name()
	if strings.HasPrefix(files[0].Name(), ".") {
		fileName = files[1].Name()
	}

	currentChartWorkingDir := filepath.Clean(filepath.Join(temporaryChartWorkingDir, fileName))

	if location == "" {
		chartName, chartVersion, err = impl.ReadChartMetaDataForLocation(temporaryChartWorkingDir, fileName)
		if err != nil {
			impl.logger.Errorw("Chart yaml file not found")
			return chartInfo, err
		}
		exists, err := impl.chartRefRepository.CheckIfDataExists(chartName, chartVersion)
		if exists {
			impl.logger.Errorw("request err, chart name and version exists already in the database")
			return chartInfo, errors.New("chart name and version exists already in the database")
		}
		if err != nil {
			impl.logger.Errorw("Error in searching the database")
			return chartInfo, err
		}
		chartLocation = impl.GetLocationFromChartNameAndVersion(chartName, chartVersion)
		if err != nil {
			impl.logger.Errorw("error in fetching name and version in Chart yaml", "err", err)
			return chartInfo, err
		}
		err = util2.CheckForMissingFiles(currentChartWorkingDir)
		if err != nil {
			impl.logger.Errorw("Missing files in the folder", "err", err)
			return chartInfo, err
		}
		location = chartLocation
	}
	err = dirCopy.Copy(currentChartWorkingDir, filepath.Clean(filepath.Join(refChartDir, location)))
	if err != nil {
		impl.logger.Errorw("error in copying chart from temp dir to ref chart dir", "err", err)
		return chartInfo, err
	}
	chartInfo.ChartLocation = location
	chartInfo.ChartName = chartName
	chartInfo.ChartVersion = chartVersion
	return chartInfo, nil
}
