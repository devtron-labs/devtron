package argoAppStatus

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/argoAppStatus"
	"go.uber.org/zap"
)

type AppStatusRequestResponseDto struct {
	AppId                       int                          `json:"appId"`
	InstalledAppId              int                          `json:"installedAppId"`
	EnvironmentStatusContainers []EnvironmentStatusContainer `json:"environmentStatusContainers"`
}

type EnvironmentStatusContainer struct {
	EnvId  int `json:"envId"`
	Status int `json:"status"`
}

type ArgoAppStatusService interface {
	GetAllDevtronAppStatuses(requests []AppStatusRequestResponseDto) ([]AppStatusRequestResponseDto, error)
	GetAllInstalledAppStatuses(requests []AppStatusRequestResponseDto) ([]AppStatusRequestResponseDto, error)
}

type ArgoAppStatusServiceImpl struct {
	argoAppStatusRepository argoAppStatus.ArgoAppStatusRepository
	logger                  *zap.SugaredLogger
}

func NewArgoAppStatusServiceImpl(argoAppStatusRepository argoAppStatus.ArgoAppStatusRepository, logger *zap.SugaredLogger) *ArgoAppStatusServiceImpl {
	return &ArgoAppStatusServiceImpl{
		argoAppStatusRepository: argoAppStatusRepository,
		logger:                  logger,
	}

}

func (impl *ArgoAppStatusServiceImpl) GetAllDevtronAppStatuses(requests []AppStatusRequestResponseDto) ([]AppStatusRequestResponseDto, error) {
	appIds := make([]int, 0)
	for _, request := range requests {
		if request.AppId > 0 {
			appIds = append(appIds, request.AppId)
		}
	}

	containers, err := impl.argoAppStatusRepository.GetAllDevtronAppStatuses(appIds)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching argo-app-statuses from argo-app-status repository for ", "appids", appIds)
		res := make([]AppStatusRequestResponseDto, 0)
		return res, err
	}
	environmentStatusMap := make(map[int][]EnvironmentStatusContainer)

	for _, container := range containers {
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

func (impl *ArgoAppStatusServiceImpl) GetAllInstalledAppStatuses(requests []AppStatusRequestResponseDto) ([]AppStatusRequestResponseDto, error) {
	installedAppIds := make([]int, 0)
	for _, request := range requests {
		if request.InstalledAppId > 0 {
			installedAppIds = append(installedAppIds, request.InstalledAppId)
		}
	}

	containers, err := impl.argoAppStatusRepository.GetAllInstalledAppStatuses(installedAppIds)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching argo-app-statuses from argo-app-status repository for ", "installAppIds", installedAppIds)
		res := make([]AppStatusRequestResponseDto, 0)
		return res, err
	}

	environmentStatusMapForInstalledApps := make(map[int][]EnvironmentStatusContainer)
	for _, container := range containers {
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
