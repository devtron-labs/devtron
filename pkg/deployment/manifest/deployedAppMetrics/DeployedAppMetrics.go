package deployedAppMetrics

import (
	"context"
	interalRepo "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"time"
)

type DeployedAppMetricsService interface {
	GetMetricsFlagByAppIdEvenIfNotInDb(appId int) (bool, error)
	CheckAndUpdateAppLevelMetrics(ctx context.Context, req *bean.DeployedAppMetricsRequest) error
}

type DeployedAppMetricsServiceImpl struct {
	logger                    *zap.SugaredLogger
	chartRefRepository        chartRepoRepository.ChartRefRepository
	appLevelMetricsRepository interalRepo.AppLevelMetricsRepository
	envLevelMetricsRepository interalRepo.EnvLevelAppMetricsRepository
}

func NewDeployedAppMetricsServiceImpl(logger *zap.SugaredLogger,
	chartRefRepository chartRepoRepository.ChartRefRepository,
	appLevelMetricsRepository interalRepo.AppLevelMetricsRepository,
	envLevelMetricsRepository interalRepo.EnvLevelAppMetricsRepository) *DeployedAppMetricsServiceImpl {
	return &DeployedAppMetricsServiceImpl{
		logger:                    logger,
		chartRefRepository:        chartRefRepository,
		appLevelMetricsRepository: appLevelMetricsRepository,
		envLevelMetricsRepository: envLevelMetricsRepository,
	}
}

func (impl *DeployedAppMetricsServiceImpl) GetMetricsFlagByAppIdEvenIfNotInDb(appId int) (bool, error) {
	isAppMetricsEnabled := false
	appMetrics, err := impl.appLevelMetricsRepository.FindByAppId(appId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error in fetching app level metrics", "appId", appId, "err", err)
		return isAppMetricsEnabled, err
	}
	if appMetrics != nil {
		isAppMetricsEnabled = appMetrics.AppMetrics
	}
	return isAppMetricsEnabled, nil
}

// CheckAndUpdateAppLevelMetrics - this method checks whether chart being used supports metrics or not and update accordingly
func (impl *DeployedAppMetricsServiceImpl) CheckAndUpdateAppLevelMetrics(ctx context.Context, req *bean.DeployedAppMetricsRequest) error {
	isAppMetricsSupported, err := impl.checkIsAppMetricsSupported(req.ChartRefId)
	if err != nil {
		return err
	}
	if !(isAppMetricsSupported) {
		//chart does not have metrics support, disabling
		req.EnableMetrics = false
	}
	_, span := otel.Tracer("orchestrator").Start(ctx, "updateAppLevelMetrics")
	_, err = impl.updateAppLevelMetrics(req)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in disable app metric flag", "error", err)
		return err
	}
	return nil
}

func (impl *DeployedAppMetricsServiceImpl) checkIsAppMetricsSupported(chartRefId int) (bool, error) {
	chartRefValue, err := impl.chartRefRepository.FindById(chartRefId)
	if err != nil {
		impl.logger.Errorw("error in finding reference chart by id", "err", err)
		return false, nil
	}
	return chartRefValue.IsAppMetricsSupported, nil
}

func (impl *DeployedAppMetricsServiceImpl) updateAppLevelMetrics(req *bean.DeployedAppMetricsRequest) (*interalRepo.AppLevelMetrics, error) {
	existingAppLevelMetrics, err := impl.appLevelMetricsRepository.FindByAppId(req.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in app metrics app level flag", "error", err)
		return nil, err
	}
	if existingAppLevelMetrics != nil && existingAppLevelMetrics.Id != 0 {
		existingAppLevelMetrics.AppMetrics = req.EnableMetrics
		existingAppLevelMetrics.UpdatedBy = req.UserId
		existingAppLevelMetrics.UpdatedOn = time.Now()
		err := impl.appLevelMetricsRepository.Update(existingAppLevelMetrics)
		if err != nil {
			impl.logger.Errorw("error in to updating app level metrics", "error", err, "model", existingAppLevelMetrics)
			return nil, err
		}
		return existingAppLevelMetrics, nil
	} else {
		appLevelMetricsNew := &interalRepo.AppLevelMetrics{
			AppId:        req.AppId,
			AppMetrics:   req.EnableMetrics,
			InfraMetrics: true,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				UpdatedOn: time.Now(),
				CreatedBy: req.UserId,
				UpdatedBy: req.UserId,
			},
		}
		err = impl.appLevelMetricsRepository.Save(appLevelMetricsNew)
		if err != nil {
			impl.logger.Errorw("error in saving app level metrics flag", "error", err, "model", appLevelMetricsNew)
			return appLevelMetricsNew, err
		}
		return appLevelMetricsNew, nil
	}
}
