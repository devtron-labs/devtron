package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/history"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ChartsHistoryService interface {
	CreateChartsHistoryFromGlobalCharts(chart *chartConfig.Chart, tx *pg.Tx) error
	CreateChartsHistoryFromEnvOverrideCharts(envOverride *chartConfig.EnvConfigOverride, tx *pg.Tx) error
}

type ChartsHistoryServiceImpl struct {
	logger                 *zap.SugaredLogger
	chartHistoryRepository history.ChartHistoryRepository
	pipelineRepository     pipelineConfig.PipelineRepository
	chartRepository        chartConfig.ChartRepository
}

func NewChartsHistoryServiceImpl(logger *zap.SugaredLogger, chartHistoryRepository history.ChartHistoryRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	chartRepository chartConfig.ChartRepository) *ChartsHistoryServiceImpl {
	return &ChartsHistoryServiceImpl{
		logger:                 logger,
		chartHistoryRepository: chartHistoryRepository,
		pipelineRepository:     pipelineRepository,
		chartRepository:        chartRepository,
	}
}

func (impl ChartsHistoryServiceImpl) CreateChartsHistoryFromGlobalCharts(chart *chartConfig.Chart, tx *pg.Tx) (err error) {
	//getting all pipelines without overridden charts
	pipelines, err := impl.pipelineRepository.FindAllPipelinesByChartsOverride(false, chart.AppId, chart.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in getting pipelines, CreateChartsHistoryFromGlobalCharts", "err", err, "chart", chart)
		return err
	}
	for _, pipeline := range pipelines {
		historyModel := &history.ChartsHistory{
			PipelineId:              pipeline.Id,
			ImageDescriptorTemplate: chart.ImageDescriptorTemplate,
			Template:                chart.GlobalOverride,
			Deployed:                false,
			AuditLog: sql.AuditLog{
				CreatedOn: chart.CreatedOn,
				CreatedBy: chart.CreatedBy,
				UpdatedOn: chart.UpdatedOn,
				UpdatedBy: chart.UpdatedBy,
			},
		}
		//creating new entry
		if tx != nil {
			_, err = impl.chartHistoryRepository.CreateHistoryWithTxn(historyModel, tx)
		} else {
			_, err = impl.chartHistoryRepository.CreateHistory(historyModel)
		}
		if err != nil {
			impl.logger.Errorw("err in creating history entry for charts", "err", err, "history", historyModel)
			return err
		}
	}
	return err
}

func (impl ChartsHistoryServiceImpl) CreateChartsHistoryFromEnvOverrideCharts(envOverride *chartConfig.EnvConfigOverride, tx *pg.Tx) (err error) {
	//getting all pipelines without overridden charts
	pipelines, err := impl.pipelineRepository.GetByEnvOverrideId(envOverride.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in getting pipelines, CreateChartsHistoryFromEnvOverrideCharts", "err", err, "envOverrideId", envOverride.Id)
		return err
	}
	chart, err := impl.chartRepository.FindById(envOverride.ChartId)
	if err != nil {
		impl.logger.Errorw("err in getting global chart", "err", err, "chart", chart)
		return err
	}
	for _, pipeline := range pipelines {
		historyModel := &history.ChartsHistory{
			PipelineId:              pipeline.Id,
			ImageDescriptorTemplate: chart.ImageDescriptorTemplate,
			Deployed:                false,
			AuditLog: sql.AuditLog{
				CreatedOn: envOverride.CreatedOn,
				CreatedBy: envOverride.CreatedBy,
				UpdatedOn: envOverride.UpdatedOn,
				UpdatedBy: envOverride.UpdatedBy,
			},
		}
		if envOverride.IsOverride {
			historyModel.Template = envOverride.EnvOverrideValues
		} else {
			//this is for the case when env override is created for new cd pipelines with template = "{}"
			historyModel.Template = chart.GlobalOverride
		}
		//creating new entry
		if tx != nil {
			_, err = impl.chartHistoryRepository.CreateHistoryWithTxn(historyModel, tx)
		} else {
			_, err = impl.chartHistoryRepository.CreateHistory(historyModel)
		}
		if err != nil {
			impl.logger.Errorw("err in creating history entry for charts", "err", err, "history", historyModel)
			return err
		}
	}
	return err
}
