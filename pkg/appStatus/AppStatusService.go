package appStatus

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/appStatus"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

const (
	HealthStatusSuspended   string = "Suspended"
	HealthStatusHibernating string = "HIBERNATING"
)

type AppStatusRequestResponseDto struct {
	AppId                       int                          `json:"appId"`
	InstalledAppId              int                          `json:"installedAppId"`
	EnvironmentStatusContainers []EnvironmentStatusContainer `json:"environmentStatusContainers"`
}

type EnvironmentStatusContainer struct {
	EnvId  int    `json:"envId"`
	Status string `json:"status"`
}

type AppStatusService interface {
	UpdateStatusWithAppIdEnvId(appIdEnvId, envId int, status string) error
	DeleteWithAppIdEnvId(tx *pg.Tx, appId, envId int) error
}

type AppStatusServiceImpl struct {
	appStatusRepository appStatus.AppStatusRepository
	logger              *zap.SugaredLogger
	enforcer            casbin.Enforcer
	enforcerUtil        rbac.EnforcerUtil
}

func NewAppStatusServiceImpl(appStatusRepository appStatus.AppStatusRepository, logger *zap.SugaredLogger, enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil) *AppStatusServiceImpl {
	return &AppStatusServiceImpl{
		appStatusRepository: appStatusRepository,
		logger:              logger,
		enforcer:            enforcer,
		enforcerUtil:        enforcerUtil,
	}

}

func (impl *AppStatusServiceImpl) UpdateStatusWithAppIdEnvId(appId, envId int, status string) error {
	container, err := impl.appStatusRepository.Get(appId, envId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting app-status for", "appId", appId, "envId", envId, "err", err)
		return err
	}

	if status == HealthStatusSuspended {
		status = HealthStatusHibernating
	}
	if container.AppId == 0 {
		container.AppId = appId
		container.EnvId = envId
		container.Status = status
		err = impl.appStatusRepository.Create(container)
		if err != nil {
			impl.logger.Errorw("error in Creating appStatus", "appId", appId, "envId", envId, "err", err)
			return err
		}
	} else if container.Status != status {
		container.Status = status
		err = impl.appStatusRepository.Update(container)
		if err != nil {
			impl.logger.Errorw("error in Updating appStatus", "appId", appId, "envId", envId, "err", err)
			return err
		}
	}

	return nil
}

func (impl *AppStatusServiceImpl) DeleteWithAppIdEnvId(tx *pg.Tx, appId, envId int) error {
	err := impl.appStatusRepository.Delete(tx, appId, envId)
	if err != nil {
		impl.logger.Errorw("error in deleting appStatus", "appId", appId, "envId", envId, "err", err)
		return err
	}
	return nil
}
