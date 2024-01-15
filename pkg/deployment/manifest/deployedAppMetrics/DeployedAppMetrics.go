package deployedAppMetrics

import (
	"context"
	interalRepo "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deployedAppMetrics/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"time"
)

type DeployedAppMetricsService interface {
	GetMetricsFlagByAppIdEvenIfNotInDb(appId int) (bool, error)
	GetMetricsFlagByAppIdAndEnvId(appId, envId int) (bool, error)
	GetMetricsFlagForAPipelineByAppIdAndEnvId(appId, envId int) (bool, error)
	CheckAndUpdateAppOrEnvLevelMetrics(ctx context.Context, req *bean.DeployedAppMetricsRequest) error
	DeleteEnvLevelMetricsIfPresent(appId, envId int) error
}

type DeployedAppMetricsServiceImpl struct {
	logger                    *zap.SugaredLogger
	appLevelMetricsRepository interalRepo.AppLevelMetricsRepository
	envLevelMetricsRepository interalRepo.EnvLevelAppMetricsRepository
	chartRefService           chartRef.ChartRefService
}

func NewDeployedAppMetricsServiceImpl(logger *zap.SugaredLogger,
	appLevelMetricsRepository interalRepo.AppLevelMetricsRepository,
	envLevelMetricsRepository interalRepo.EnvLevelAppMetricsRepository,
	chartRefService chartRef.ChartRefService) *DeployedAppMetricsServiceImpl {
	return &DeployedAppMetricsServiceImpl{
		logger:                    logger,
		appLevelMetricsRepository: appLevelMetricsRepository,
		envLevelMetricsRepository: envLevelMetricsRepository,
		chartRefService:           chartRefService,
	}
}

func (impl *DeployedAppMetricsServiceImpl) GetMetricsFlagByAppIdEvenIfNotInDb(appId int) (bool, error) {
	isAppMetricsEnabled := false
	appMetrics, err := impl.appLevelMetricsRepository.FindByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching app level metrics", "appId", appId, "err", err)
		return isAppMetricsEnabled, err
	}
	if appMetrics != nil {
		isAppMetricsEnabled = appMetrics.AppMetrics
	}
	return isAppMetricsEnabled, nil
}

func (impl *DeployedAppMetricsServiceImpl) GetMetricsFlagByAppIdAndEnvId(appId, envId int) (bool, error) {
	isAppMetricsEnabled := false
	envLevelAppMetrics, err := impl.envLevelMetricsRepository.FindByAppIdAndEnvId(appId, envId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting env level app metrics", "err", err, "appId", appId, "envId", envId)
		return isAppMetricsEnabled, err
	}
	if envLevelAppMetrics != nil {
		isAppMetricsEnabled = *envLevelAppMetrics.AppMetrics
	}
	return isAppMetricsEnabled, nil
}

// GetMetricsFlagForAPipelineByAppIdAndEnvId - this function returns metrics flag for pipeline after resolving override and app level values
func (impl *DeployedAppMetricsServiceImpl) GetMetricsFlagForAPipelineByAppIdAndEnvId(appId, envId int) (bool, error) {
	isAppMetricsEnabled := false
	envLevelAppMetrics, err := impl.envLevelMetricsRepository.FindByAppIdAndEnvId(appId, envId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting env level app metrics", "err", err, "appId", appId, "envId", envId)
		return isAppMetricsEnabled, err
	} else if err == pg.ErrNoRows {
		isAppLevelMetricsEnabled, err := impl.GetMetricsFlagByAppIdEvenIfNotInDb(appId)
		if err != nil {
			impl.logger.Errorw("error, GetMetricsFlagByAppIdEvenIfNotInDb", "err", err, "appId", appId)
			return false, err
		}
		isAppMetricsEnabled = isAppLevelMetricsEnabled
	} else if envLevelAppMetrics != nil && envLevelAppMetrics.AppMetrics != nil {
		isAppMetricsEnabled = *envLevelAppMetrics.AppMetrics
	}
	return isAppMetricsEnabled, nil
}

// CheckAndUpdateAppOrEnvLevelMetrics - this method checks whether chart being used supports metrics or not, is app level or env level and updates accordingly
func (impl *DeployedAppMetricsServiceImpl) CheckAndUpdateAppOrEnvLevelMetrics(ctx context.Context, req *bean.DeployedAppMetricsRequest) error {
	isAppMetricsSupported, err := impl.checkIsAppMetricsSupported(req.ChartRefId)
	if err != nil {
		return err
	}
	if !(isAppMetricsSupported) {
		//chart does not have metrics support, disabling
		req.EnableMetrics = false
		return nil
	}
	if req.EnvId == 0 {
		_, span := otel.Tracer("orchestrator").Start(ctx, "createOrUpdateAppLevelMetrics")
		_, err = impl.createOrUpdateAppLevelMetrics(req)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in disable app metric flag", "error", err, "req", req)
			return err
		}
	} else {
		_, span := otel.Tracer("orchestrator").Start(ctx, "createOrUpdateEnvLevelMetrics")
		_, err = impl.createOrUpdateEnvLevelMetrics(req)
		span.End()
		if err != nil {
			impl.logger.Errorw("error in disable env level app metric flag", "error", err, "req", req)
			return err
		}
	}
	return nil
}

func (impl *DeployedAppMetricsServiceImpl) DeleteEnvLevelMetricsIfPresent(appId, envId int) error {
	envLevelAppMetrics, err := impl.envLevelMetricsRepository.FindByAppIdAndEnvId(appId, envId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error while fetching env level app metric", "err", err, "appId", appId, "envId", envId)
		return err
	}
	if envLevelAppMetrics != nil && envLevelAppMetrics.Id > 0 {
		err = impl.envLevelMetricsRepository.Delete(envLevelAppMetrics)
		if err != nil {
			impl.logger.Errorw("error while deletion of app metric at env level", "err", err, "model", envLevelAppMetrics)
			return err
		}
	}
	return nil
}

func (impl *DeployedAppMetricsServiceImpl) checkIsAppMetricsSupported(chartRefId int) (bool, error) {
	chartRefValue, err := impl.chartRefService.FindById(chartRefId)
	if err != nil {
		impl.logger.Errorw("error in finding reference chart by id", "err", err)
		return false, nil
	}
	return chartRefValue.IsAppMetricsSupported, nil
}

func (impl *DeployedAppMetricsServiceImpl) createOrUpdateAppLevelMetrics(req *bean.DeployedAppMetricsRequest) (*interalRepo.AppLevelMetrics, error) {
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

func (impl *DeployedAppMetricsServiceImpl) createOrUpdateEnvLevelMetrics(req *bean.DeployedAppMetricsRequest) (*interalRepo.EnvLevelAppMetrics, error) {
	// update and create env level app metrics
	envLevelAppMetrics, err := impl.envLevelMetricsRepository.FindByAppIdAndEnvId(req.AppId, req.EnvId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Error("err", err)
		return nil, err
	}
	if envLevelAppMetrics == nil || envLevelAppMetrics.Id == 0 {
		infraMetrics := true
		envLevelAppMetrics = &interalRepo.EnvLevelAppMetrics{
			AppId:        req.AppId,
			EnvId:        req.EnvId,
			AppMetrics:   &req.EnableMetrics,
			InfraMetrics: &infraMetrics,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				UpdatedOn: time.Now(),
				CreatedBy: req.UserId,
				UpdatedBy: req.UserId,
			},
		}
		err = impl.envLevelMetricsRepository.Save(envLevelAppMetrics)
		if err != nil {
			impl.logger.Error("err", err)
			return nil, err
		}
	} else {
		envLevelAppMetrics.AppMetrics = &req.EnableMetrics
		envLevelAppMetrics.UpdatedOn = time.Now()
		envLevelAppMetrics.UpdatedBy = req.UserId
		err = impl.envLevelMetricsRepository.Update(envLevelAppMetrics)
		if err != nil {
			impl.logger.Error("err", err)
			return nil, err
		}
	}
	return envLevelAppMetrics, err
}
