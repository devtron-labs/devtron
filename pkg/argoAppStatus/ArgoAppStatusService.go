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
	GetAll() []AppStatusRequestResponseDto
}

type ArgoAppStatusServiceImpl struct {
	argoAppStatusRepository argoAppStatus.ArgoAppStatusRepository
	logger                  *zap.SugaredLogger
}

func (impl *ArgoAppStatusServiceImpl) GetAll(requests []AppStatusRequestResponseDto) []AppStatusRequestResponseDto {
	appIds := make([]int, 0)
	installedAppIds := make([]int, 0)
	for _, request := range requests {
		if request.AppId > 0 {
			appIds = append(appIds, request.AppId)
		} else if request.InstalledAppId > 0 {
			installedAppIds = append(installedAppIds, request.InstalledAppId)
		}
	}
	containers, err := impl.argoAppStatusRepository.GetAllWithAppIds(appIds, installedAppIds)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching argo-app-statuses from argo-app-status repository for ", "appids", appIds, "installAppIds", installedAppIds)
		res := make([]AppStatusRequestResponseDto, 0)
		return res
	}
	environmentStatusMap := make(map[int][]EnvironmentStatusContainer)
	environmentStatusMapForInstalledApps := make(map[int][]EnvironmentStatusContainer)

	for _, container := range containers {
		envContainer := EnvironmentStatusContainer{
			EnvId:  container.EnvId,
			Status: container.Status,
		}
		if container.AppType == argoAppStatus.DEVTRON_APP {
			if _, ok := environmentStatusMap[container.AppId]; !ok {
				environmentStatusMap[container.AppId] = make([]EnvironmentStatusContainer, 0)
			}
			environmentStatusMap[container.AppId] = append(environmentStatusMap[container.AppId], envContainer)
		} else if container.AppType == argoAppStatus.DEVTRON_INSTALLED_APP {
			if _, ok := environmentStatusMapForInstalledApps[container.AppId]; !ok {
				environmentStatusMapForInstalledApps[container.AppId] = make([]EnvironmentStatusContainer, 0)
			}
			environmentStatusMapForInstalledApps[container.AppId] = append(environmentStatusMap[container.AppId], envContainer)
		}
	}

	response := make([]AppStatusRequestResponseDto, len(environmentStatusMap)+len(environmentStatusMapForInstalledApps))
	var itr = 0
	for id, envContainersArray := range environmentStatusMap {
		resultDto := AppStatusRequestResponseDto{
			AppId:                       id,
			EnvironmentStatusContainers: envContainersArray,
		}
		response[itr] = resultDto
		itr++
	}

	for id, envContainersArray := range environmentStatusMapForInstalledApps {
		resultDto := AppStatusRequestResponseDto{
			InstalledAppId:              id,
			EnvironmentStatusContainers: envContainersArray,
		}
		response[itr] = resultDto
		itr++
	}

	return response
}
