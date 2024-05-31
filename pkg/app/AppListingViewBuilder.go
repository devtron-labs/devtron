/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package app

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"go.uber.org/zap"
	"sort"
	"strconv"
	"strings"
)

type AppListingViewBuilder interface {
	BuildView(fetchAppListingRequest FetchAppListingRequest, appEnvMap map[string][]*bean.AppEnvironmentContainer) ([]*bean.AppContainer, error)
}

type AppListingViewBuilderImpl struct {
	Logger *zap.SugaredLogger
}

func NewAppListingViewBuilderImpl(Logger *zap.SugaredLogger) *AppListingViewBuilderImpl {
	return &AppListingViewBuilderImpl{
		Logger: Logger,
	}
}

func (impl *AppListingViewBuilderImpl) BuildView(fetchAppListingRequest FetchAppListingRequest, appEnvMap map[string][]*bean.AppEnvironmentContainer) ([]*bean.AppContainer, error) {
	// filter status
	filteredAppEnvMap := map[string][]*bean.AppEnvironmentContainer{}
	for k, v := range appEnvMap {
		for _, e := range v {
			if e.Deleted {
				//continue
				e.Status = NotDeployed
			}
			if !impl.filterStatus(fetchAppListingRequest, e.Status) {
				if _, ok := filteredAppEnvMap[k]; !ok {
					var envs []*bean.AppEnvironmentContainer
					filteredAppEnvMap[k] = envs
				}
				filteredAppEnvMap[k] = append(filteredAppEnvMap[k], e)
			}
		}
	}

	var appContainersResponses []*bean.AppContainer
	for k, v := range filteredAppEnvMap {
		appId, err := strconv.Atoi(strings.Split(k, "_")[0])
		if err != nil {
			impl.Logger.Error("err", err)
			return []*bean.AppContainer{}, nil
		}
		appName := strings.Split(k, "_")[1]
		defaultEnv := bean.AppEnvironmentContainer{}
		projectId := 0
		for _, env := range v {
			projectId = env.TeamId
			if env.Default {
				defaultEnv = *env
				break
			}
		}

		sort.Slice(v, func(i, j int) bool {
			return v[i].LastDeployedTime >= v[j].LastDeployedTime
		})

		appContainerResponse := &bean.AppContainer{
			AppId:                   appId,
			AppName:                 appName,
			AppEnvironmentContainer: v,
			DefaultEnv:              defaultEnv,
			ProjectId:               projectId,
		}
		appContainersResponses = append(appContainersResponses, appContainerResponse)
	}

	// Sort apps based on default envs
	if fetchAppListingRequest.SortBy != "" {
		if helper.AppNameSortBy == fetchAppListingRequest.SortBy {
			if fetchAppListingRequest.SortOrder == helper.Asc {
				sort.Slice(appContainersResponses, func(i, j int) bool {
					return appContainersResponses[i].AppName < appContainersResponses[j].AppName
				})
			} else if fetchAppListingRequest.SortOrder == helper.Desc {
				sort.Slice(appContainersResponses, func(i, j int) bool {
					return appContainersResponses[i].AppName > appContainersResponses[j].AppName
				})
			}
		} else if helper.LastDeployedSortBy == fetchAppListingRequest.SortBy {
			if fetchAppListingRequest.SortOrder == helper.Asc {
				sort.Slice(appContainersResponses, func(i, j int) bool {
					deployedTime1 := appContainersResponses[i].AppEnvironmentContainer[0].LastDeployedTime
					deployedTime2 := appContainersResponses[j].AppEnvironmentContainer[0].LastDeployedTime
					return deployedTime1 < deployedTime2
				})
			} else if fetchAppListingRequest.SortOrder == helper.Desc {
				sort.Slice(appContainersResponses, func(i, j int) bool {
					deployedTime1 := appContainersResponses[i].AppEnvironmentContainer[0].LastDeployedTime
					deployedTime2 := appContainersResponses[j].AppEnvironmentContainer[0].LastDeployedTime
					return deployedTime1 > deployedTime2
				})
			}
		}
	}
	return appContainersResponses, nil
}

func (impl AppListingViewBuilderImpl) filterStatus(fetchAppListingRequest FetchAppListingRequest, status string) bool {
	return len(fetchAppListingRequest.Statuses) > 0 && !arrContains(fetchAppListingRequest.Statuses, status)
}
