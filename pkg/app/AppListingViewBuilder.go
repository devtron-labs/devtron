/*
 * Copyright (c) 2020-2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package app

import (
	"errors"
	"github.com/devtron-labs/devtron/api/bean/AppView"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"go.uber.org/zap"
	"sort"
	"strconv"
	"strings"
)

type AppListingViewBuilder interface {
	BuildView(fetchAppListingRequest FetchAppListingRequest, appEnvMap map[string][]*AppView.AppEnvironmentContainer) ([]*AppView.AppContainer, error)
}

type AppListingViewBuilderImpl struct {
	Logger *zap.SugaredLogger
}

func NewAppListingViewBuilderImpl(Logger *zap.SugaredLogger) *AppListingViewBuilderImpl {
	return &AppListingViewBuilderImpl{
		Logger: Logger,
	}
}

func (impl *AppListingViewBuilderImpl) BuildView(fetchAppListingRequest FetchAppListingRequest, appEnvMap map[string][]*AppView.AppEnvironmentContainer) ([]*AppView.AppContainer, error) {
	// filter status
	filteredAppEnvMap := map[string][]*AppView.AppEnvironmentContainer{}
	for k, v := range appEnvMap {
		for _, e := range v {
			if e.Deleted {
				//continue
				e.Status = NotDeployed
			}
			if !impl.filterStatus(fetchAppListingRequest, e.Status) {
				if _, ok := filteredAppEnvMap[k]; !ok {
					var envs []*AppView.AppEnvironmentContainer
					filteredAppEnvMap[k] = envs
				}
				filteredAppEnvMap[k] = append(filteredAppEnvMap[k], e)
			}
		}
	}

	var appContainersResponses []*AppView.AppContainer
	for k, v := range filteredAppEnvMap {
		appIdAndName := strings.Split(k, "_")
		if len(appIdAndName) != 2 {
			return []*AppView.AppContainer{}, errors.New("invalid format for app id and name. It should be in format <appId>_<appName>")
		}
		appId, err := strconv.Atoi(appIdAndName[0])
		if err != nil {
			impl.Logger.Error("err", err)
			return []*AppView.AppContainer{}, nil
		}
		appName := appIdAndName[1]
		defaultEnv := AppView.AppEnvironmentContainer{}
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

		appContainerResponse := &AppView.AppContainer{
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
