/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
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
		for _, env := range v {
			if env.Default {
				defaultEnv = *env
				break
			}
		}
		appContainerResponse := &bean.AppContainer{
			AppId:                   appId,
			AppName:                 appName,
			AppEnvironmentContainer: v,
			DefaultEnv:              defaultEnv,
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
		}
	}
	return appContainersResponses, nil
}

func (impl AppListingViewBuilderImpl) filterStatus(fetchAppListingRequest FetchAppListingRequest, status string) bool {
	return len(fetchAppListingRequest.Statuses) > 0 && !arrContains(fetchAppListingRequest.Statuses, status)
}
