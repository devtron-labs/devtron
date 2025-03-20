package read

import (
	"encoding/json"
	"fmt"
	apiGitOpsBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/chart/bean"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/common"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics"
	bean3 "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"
	chartRefRead "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/read"
	util2 "github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"strings"
)

type ChartReadService interface {
	GetByAppIdAndChartRefId(appId int, chartRefId int) (chartTemplate *bean.TemplateRequest, err error)
	IsGitOpsRepoConfiguredForDevtronApps(appIds []int) (map[int]bool, error)
	FindLatestChartForAppByAppId(appId int) (chartTemplate *bean.TemplateRequest, err error)
	GetChartRefConfiguredForApp(appId int) (*bean3.ChartRefDto, error)
}

type ChartReadServiceImpl struct {
	logger                    *zap.SugaredLogger
	chartRepository           chartRepoRepository.ChartRepository
	deploymentConfigService   common.DeploymentConfigService
	deployedAppMetricsService deployedAppMetrics.DeployedAppMetricsService
	gitOpsConfigReadService   config.GitOpsConfigReadService
	ChartRefReadService       chartRefRead.ChartRefReadService
}

func NewChartReadServiceImpl(logger *zap.SugaredLogger,
	chartRepository chartRepoRepository.ChartRepository,
	deploymentConfigService common.DeploymentConfigService,
	deployedAppMetricsService deployedAppMetrics.DeployedAppMetricsService,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	ChartRefReadService chartRefRead.ChartRefReadService) *ChartReadServiceImpl {
	return &ChartReadServiceImpl{
		logger:                    logger,
		chartRepository:           chartRepository,
		deploymentConfigService:   deploymentConfigService,
		deployedAppMetricsService: deployedAppMetricsService,
		gitOpsConfigReadService:   gitOpsConfigReadService,
		ChartRefReadService:       ChartRefReadService,
	}

}

func (impl *ChartReadServiceImpl) GetByAppIdAndChartRefId(appId int, chartRefId int) (chartTemplate *bean.TemplateRequest, err error) {
	chart, err := impl.chartRepository.FindChartByAppIdAndRefId(appId, chartRefId)
	if err != nil {
		impl.logger.Errorw("error in fetching chart ", "appId", appId, "err", err)
		return nil, err
	}
	isAppMetricsEnabled, err := impl.deployedAppMetricsService.GetMetricsFlagByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching app-metrics", "appId", appId, "err", err)
		return nil, err
	}
	deploymentConfig, err := impl.deploymentConfigService.GetConfigForDevtronApps(appId, 0)
	if err != nil {
		impl.logger.Errorw("error in fetching deployment config by appId", "appId", appId, "err", err)
		return nil, err
	}
	chartTemplate, err = impl.chartAdaptor(chart, isAppMetricsEnabled, deploymentConfig)
	return chartTemplate, err
}

func (impl *ChartReadServiceImpl) IsGitOpsRepoConfiguredForDevtronApps(appIds []int) (map[int]bool, error) {
	gitOpsConfigStatus, err := impl.gitOpsConfigReadService.IsGitOpsConfigured()
	if err != nil {
		impl.logger.Errorw("error in fetching latest chart for app by appId")
		return nil, err
	}
	appIdRepoConfiguredMap := make(map[int]bool, len(appIds))
	for _, appId := range appIds {
		if !gitOpsConfigStatus.IsGitOpsConfiguredAndArgoCdInstalled() {
			appIdRepoConfiguredMap[appId] = false
		} else if !gitOpsConfigStatus.AllowCustomRepository {
			appIdRepoConfiguredMap[appId] = true
		} else {
			latestChartConfiguredInApp, err := impl.FindLatestChartForAppByAppId(appId)
			if err != nil {
				impl.logger.Errorw("error in fetching latest chart for app by appId")
				return nil, err
			}
			appIdRepoConfiguredMap[appId] = !apiGitOpsBean.IsGitOpsRepoNotConfigured(latestChartConfiguredInApp.GitRepoUrl)
		}
	}
	return appIdRepoConfiguredMap, nil
}

func (impl *ChartReadServiceImpl) FindLatestChartForAppByAppId(appId int) (chartTemplate *bean.TemplateRequest, err error) {
	chart, err := impl.chartRepository.FindLatestChartForAppByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching chart ", "appId", appId, "err", err)
		return nil, err
	}

	deploymentConfig, err := impl.deploymentConfigService.GetConfigForDevtronApps(appId, 0)
	if err != nil {
		impl.logger.Errorw("error in fetching deployment config by appId", "appId", appId, "err", err)
		return nil, err
	}

	isAppMetricsEnabled, err := impl.deployedAppMetricsService.GetMetricsFlagByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching app-metrics", "appId", appId, "err", err)
		return nil, err
	}
	chartTemplate, err = impl.chartAdaptor(chart, isAppMetricsEnabled, deploymentConfig)
	return chartTemplate, err
}

// converts db object to bean
func (impl *ChartReadServiceImpl) chartAdaptor(chartInput *chartRepoRepository.Chart, isAppMetricsEnabled bool, deploymentConfig *bean2.DeploymentConfig) (*bean.TemplateRequest, error) {
	if chartInput == nil || chartInput.Id == 0 {
		return &bean.TemplateRequest{}, &util.ApiError{UserMessage: "no chartInput found"}
	}
	gitRepoUrl := ""
	targetRevision := util2.GetDefaultTargetRevision()
	if !apiGitOpsBean.IsGitOpsRepoNotConfigured(deploymentConfig.GetRepoURL()) {
		gitRepoUrl = deploymentConfig.GetRepoURL()
		targetRevision = deploymentConfig.GetTargetRevision()
	}
	templateRequest := &bean.TemplateRequest{
		RefChartTemplate:        chartInput.ReferenceTemplate,
		Id:                      chartInput.Id,
		AppId:                   chartInput.AppId,
		ChartRepositoryId:       chartInput.ChartRepoId,
		DefaultAppOverride:      json.RawMessage(chartInput.GlobalOverride),
		RefChartTemplateVersion: impl.getParentChartVersion(chartInput.ChartVersion),
		Latest:                  chartInput.Latest,
		ChartRefId:              chartInput.ChartRefId,
		IsAppMetricsEnabled:     isAppMetricsEnabled,
		IsBasicViewLocked:       chartInput.IsBasicViewLocked,
		CurrentViewEditor:       chartInput.CurrentViewEditor,
		GitRepoUrl:              gitRepoUrl,
		IsCustomGitRepository:   deploymentConfig.ConfigType == bean2.CUSTOM.String(),
		ImageDescriptorTemplate: chartInput.ImageDescriptorTemplate,
		TargetRevision:          targetRevision,
	}
	if chartInput.Latest {
		templateRequest.LatestChartVersion = chartInput.ChartVersion
	}
	return templateRequest, nil
}

func (impl *ChartReadServiceImpl) GetChartRefConfiguredForApp(appId int) (*bean3.ChartRefDto, error) {
	latestChart, err := impl.FindLatestChartForAppByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in finding latest chart by appId", "appId", appId, "err", err)
		return nil, nil
	}
	chartRef, err := impl.ChartRefReadService.FindById(latestChart.ChartRefId)
	if err != nil {
		impl.logger.Errorw("error in finding latest chart by appId", "appId", appId, "err", err)
		return nil, nil
	}
	return chartRef, nil
}

func (impl *ChartReadServiceImpl) getParentChartVersion(childVersion string) string {
	placeholders := strings.Split(childVersion, ".")
	return fmt.Sprintf("%s.%s.0", placeholders[0], placeholders[1])
}
