package appStatus

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/appStatus"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"strings"
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
	GetAllDevtronAppStatuses(requests []AppStatusRequestResponseDto, token string) ([]AppStatusRequestResponseDto, error)
	GetAllInstalledAppStatuses(requests []AppStatusRequestResponseDto, token string) ([]AppStatusRequestResponseDto, error)
	UpdateStatusWithAppIdEnvId(appIdEnvId, envId int, status string) error
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

func (impl *AppStatusServiceImpl) getRBACObjectMap(containers []appStatus.AppStatusContainer, userEmailId string) map[string]bool {
	noOfContainers := len(containers)
	objectArray := make([]string, noOfContainers)
	for i := 0; i < noOfContainers; i++ {
		objectArray[i] = fmt.Sprintf("%s/%s", strings.ToLower(containers[i].EnvIdentifier), strings.ToLower(containers[i].AppName))
	}
	rbacObjectMap := impl.enforcer.EnforceByEmailInBatch(userEmailId, casbin.ResourceApplications, casbin.ActionGet, objectArray)
	return rbacObjectMap
}

func (impl *AppStatusServiceImpl) GetAllDevtronAppStatuses(requests []AppStatusRequestResponseDto, userEmailId string) ([]AppStatusRequestResponseDto, error) {
	appIds := make([]int, 0)
	for _, request := range requests {
		if request.AppId > 0 {
			appIds = append(appIds, request.AppId)
		}
	}

	containers, err := impl.appStatusRepository.GetAllDevtronAppStatuses(appIds)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching argo-app-statuses from argo-app-status repository for ", "appids", appIds)
		res := make([]AppStatusRequestResponseDto, 0)
		return res, err
	}

	//Rbac
	rbacObjectMap := impl.getRBACObjectMap(containers, userEmailId)
	environmentStatusMap := make(map[int][]EnvironmentStatusContainer)

	for _, container := range containers {
		object := fmt.Sprintf("%s/%s", strings.ToLower(container.EnvIdentifier), strings.ToLower(container.AppName))
		ok := rbacObjectMap[object]
		if !ok {
			continue
		}
		envContainer := EnvironmentStatusContainer{
			EnvId:  container.EnvId,
			Status: container.Status,
		}
		if !container.AppStore {
			if _, ok := environmentStatusMap[container.AppId]; !ok {
				environmentStatusMap[container.AppId] = make([]EnvironmentStatusContainer, 0)
			}
			environmentStatusMap[container.AppId] = append(environmentStatusMap[container.AppId], envContainer)
		}
	}

	response := make([]AppStatusRequestResponseDto, len(environmentStatusMap))
	var itr = 0
	for id, envContainersArray := range environmentStatusMap {
		resultDto := AppStatusRequestResponseDto{
			AppId:                       id,
			EnvironmentStatusContainers: envContainersArray,
		}
		response[itr] = resultDto
		itr++
	}

	return response, nil
}

func (impl *AppStatusServiceImpl) GetAllInstalledAppStatuses(requests []AppStatusRequestResponseDto, userEmailId string) ([]AppStatusRequestResponseDto, error) {
	installedAppIds := make([]int, 0)
	for _, request := range requests {
		if request.InstalledAppId > 0 {
			installedAppIds = append(installedAppIds, request.InstalledAppId)
		}
	}

	containers, err := impl.appStatusRepository.GetAllInstalledAppStatuses(installedAppIds)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching argo-app-statuses from argo-app-status repository for ", "installAppIds", installedAppIds)
		res := make([]AppStatusRequestResponseDto, 0)
		return res, err
	}
	rbacObjectMap := impl.getRBACObjectMap(containers, userEmailId)
	environmentStatusMapForInstalledApps := make(map[int][]EnvironmentStatusContainer)
	for _, container := range containers {
		object := fmt.Sprintf("%s/%s", strings.ToLower(container.EnvIdentifier), strings.ToLower(container.AppName))
		ok := rbacObjectMap[object]
		if !ok {
			continue
		}
		envContainer := EnvironmentStatusContainer{
			EnvId:  container.EnvId,
			Status: container.Status,
		}
		if container.AppStore {
			if _, ok := environmentStatusMapForInstalledApps[container.InstalledAppId]; !ok {
				environmentStatusMapForInstalledApps[container.InstalledAppId] = make([]EnvironmentStatusContainer, 0)
			}
			environmentStatusMapForInstalledApps[container.InstalledAppId] = append(environmentStatusMapForInstalledApps[container.InstalledAppId], envContainer)
		}
	}

	response := make([]AppStatusRequestResponseDto, len(environmentStatusMapForInstalledApps))
	var itr = 0

	for installAppId, envContainersArray := range environmentStatusMapForInstalledApps {
		resultDto := AppStatusRequestResponseDto{
			InstalledAppId:              installAppId,
			EnvironmentStatusContainers: envContainersArray,
		}
		response[itr] = resultDto
		itr++
	}

	return response, nil
}

func (impl *AppStatusServiceImpl) UpdateStatusWithAppIdEnvId(appId, envId int, status string) error {
	container, err := impl.appStatusRepository.Get(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in getting app-status for", "appId", appId, "envId", envId, "err", err)
		return err
	}
	tx, err := impl.appStatusRepository.GetConnection().Begin()
	if err != nil {
		impl.logger.Errorw("error in creating transaction", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	container.Status = status
	if container.AppId == 0 {
		container.AppId = appId
		container.EnvId = envId
		err = impl.appStatusRepository.Create(tx, container)
		if err != nil {
			impl.logger.Errorw("error in Creating appStatus", "appId", appId, "envId", envId, "err", err)
			return err
		}
	} else {
		err = impl.appStatusRepository.Update(tx, container)
		if err != nil {
			impl.logger.Errorw("error in Updating appStatus", "appId", appId, "envId", envId, "err", err)
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error occurred while committing db transaction", "err", err)
		return err
	}
	return nil
}
