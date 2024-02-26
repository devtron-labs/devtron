package deploymentTemplate

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/util"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
	bean4 "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	chart2 "k8s.io/helm/pkg/proto/hapi/chart"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type DeploymentTemplateService interface {
	BuildChartAndGetPath(appName string, envOverride *chartConfig.EnvConfigOverride, ctx context.Context) (string, error)
}

type DeploymentTemplateServiceImpl struct {
	logger               *zap.SugaredLogger
	chartRefService      chartRef.ChartRefService
	chartTemplateService util.ChartTemplateService

	chartRepository chartRepoRepository.ChartRepository
}

func NewDeploymentTemplateServiceImpl(logger *zap.SugaredLogger,
	chartRefService chartRef.ChartRefService,
	chartTemplateService util.ChartTemplateService,
	chartRepository chartRepoRepository.ChartRepository) *DeploymentTemplateServiceImpl {
	return &DeploymentTemplateServiceImpl{
		logger:               logger,
		chartRefService:      chartRefService,
		chartTemplateService: chartTemplateService,
		chartRepository:      chartRepository,
	}
}

func (impl *DeploymentTemplateServiceImpl) BuildChartAndGetPath(appName string, envOverride *chartConfig.EnvConfigOverride, ctx context.Context) (string, error) {
	if !strings.HasSuffix(envOverride.Chart.ChartLocation, fmt.Sprintf("%s%s", "/", envOverride.Chart.ChartVersion)) {
		_, span := otel.Tracer("orchestrator").Start(ctx, "autoHealChartLocationInChart")
		err := impl.autoHealChartLocationInChart(ctx, envOverride)
		span.End()
		if err != nil {
			return "", err
		}
	}
	chartMetaData := &chart2.Metadata{
		Name:    appName,
		Version: envOverride.Chart.ChartVersion,
	}
	referenceTemplatePath := path.Join(bean4.RefChartDirPath, envOverride.Chart.ReferenceTemplate)
	// Load custom charts to referenceTemplatePath if not exists
	if _, err := os.Stat(referenceTemplatePath); os.IsNotExist(err) {
		chartRefValue, err := impl.chartRefService.FindById(envOverride.Chart.ChartRefId)
		if err != nil {
			impl.logger.Errorw("error in fetching ChartRef data", "err", err)
			return "", err
		}
		if chartRefValue.ChartData != nil {
			chartInfo, err := impl.chartRefService.ExtractChartIfMissing(chartRefValue.ChartData, bean4.RefChartDirPath, chartRefValue.Location)
			if chartInfo != nil && chartInfo.TemporaryFolder != "" {
				err1 := os.RemoveAll(chartInfo.TemporaryFolder)
				if err1 != nil {
					impl.logger.Errorw("error in deleting temp dir ", "err", err)
				}
			}
			return "", err
		}
	}
	_, span := otel.Tracer("orchestrator").Start(ctx, "chartTemplateService.BuildChart")
	tempReferenceTemplateDir, err := impl.chartTemplateService.BuildChart(ctx, chartMetaData, referenceTemplatePath)
	span.End()
	if err != nil {
		return "", err
	}
	return tempReferenceTemplateDir, nil
}

func (impl *DeploymentTemplateServiceImpl) autoHealChartLocationInChart(ctx context.Context, envOverride *chartConfig.EnvConfigOverride) error {
	chartId := envOverride.Chart.Id
	impl.logger.Infow("auto-healing: Chart location in chart not correct. modifying ", "chartId", chartId,
		"current chartLocation", envOverride.Chart.ChartLocation, "current chartVersion", envOverride.Chart.ChartVersion)

	// get chart from DB (getting it from DB because envOverride.Chart does not have full row of DB)
	_, span := otel.Tracer("orchestrator").Start(ctx, "chartRepository.FindById")
	chart, err := impl.chartRepository.FindById(chartId)
	span.End()
	if err != nil {
		impl.logger.Errorw("error occurred while fetching chart from DB", "chartId", chartId, "err", err)
		return err
	}

	// get chart ref from DB (to get location)
	chartRefId := chart.ChartRefId
	_, span = otel.Tracer("orchestrator").Start(ctx, "chartRefRepository.FindById")
	chartRefDto, err := impl.chartRefService.FindById(chartRefId)
	span.End()
	if err != nil {
		impl.logger.Errorw("error occurred while fetching chartRef from DB", "chartRefId", chartRefId, "err", err)
		return err
	}

	// build new chart location
	newChartLocation := filepath.Join(chartRefDto.Location, envOverride.Chart.ChartVersion)
	impl.logger.Infow("new chart location build", "chartId", chartId, "newChartLocation", newChartLocation)

	// update chart in DB
	chart.ChartLocation = newChartLocation
	_, span = otel.Tracer("orchestrator").Start(ctx, "chartRepository.Update")
	err = impl.chartRepository.Update(chart)
	span.End()
	if err != nil {
		impl.logger.Errorw("error occurred while saving chart into DB", "chartId", chartId, "err", err)
		return err
	}

	// update newChartLocation in model
	envOverride.Chart.ChartLocation = newChartLocation
	return nil
}
