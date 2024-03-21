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

package appList

import (
	"github.com/caarlos0/env/v6"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/cluster"
	bean2 "github.com/devtron-labs/devtron/pkg/cluster/repository/bean"
	"github.com/devtron-labs/devtron/pkg/team"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type AppFilteringRestHandler interface {
	GetClusterTeamAndEnvListForAutocomplete(w http.ResponseWriter, r *http.Request)
}

type AppFilteringRestHandlerImpl struct {
	logger                            *zap.SugaredLogger
	teamService                       team.TeamService
	enforcer                          casbin.Enforcer
	userService                       user.UserService
	clusterService                    cluster.ClusterService
	environmentClusterMappingsService cluster.EnvironmentService
	cfg                               *bean.Config
}

func NewAppFilteringRestHandlerImpl(logger *zap.SugaredLogger,
	teamService team.TeamService,
	enforcer casbin.Enforcer,
	userService user.UserService,
	clusterService cluster.ClusterService,
	environmentClusterMappingsService cluster.EnvironmentService,
) *AppFilteringRestHandlerImpl {
	cfg := &bean.Config{}
	err := env.Parse(cfg)
	if err != nil {
		logger.Errorw("error occurred while parsing config ", "err", err)
		cfg.IgnoreAuthCheck = false
	}
	logger.Infow("app listing rest handler initialized", "ignoreAuthCheckValue", cfg.IgnoreAuthCheck)
	appFilteringRestHandler := &AppFilteringRestHandlerImpl{
		logger:                            logger,
		teamService:                       teamService,
		enforcer:                          enforcer,
		userService:                       userService,
		clusterService:                    clusterService,
		environmentClusterMappingsService: environmentClusterMappingsService,
		cfg:                               cfg,
	}
	return appFilteringRestHandler
}

func (handler AppFilteringRestHandlerImpl) GetClusterTeamAndEnvListForAutocomplete(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	clusterMapping := make(map[string]cluster.ClusterBean)
	start := time.Now()
	clusterList, err := handler.clusterService.FindAllForAutoComplete()
	dbOperationTime := time.Since(start)
	if err != nil {
		handler.logger.Errorw("service err, FindAllForAutoComplete in clusterService layer", "error", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	var granterClusters []cluster.ClusterBean
	v := r.URL.Query()
	authEnabled := true
	auth := v.Get("auth")
	if len(auth) > 0 {
		authEnabled, err = strconv.ParseBool(auth)
		if err != nil {
			authEnabled = true
			err = nil
			//ignore error, apply rbac by default
		}
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	start = time.Now()
	for _, item := range clusterList {
		clusterMapping[strings.ToLower(item.ClusterName)] = item
		if authEnabled == true {
			if ok := handler.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionGet, item.ClusterName); ok {
				granterClusters = append(granterClusters, item)
			}
		} else {
			granterClusters = append(granterClusters, item)
		}

	}
	handler.logger.Infow("Cluster elapsed Time for enforcer", "dbElapsedTime", dbOperationTime, "enforcerTime", time.Since(start), "envSize", len(granterClusters))
	//RBAC enforcer Ends

	if len(granterClusters) == 0 {
		granterClusters = make([]cluster.ClusterBean, 0)
	}

	//getting environment for autocomplete
	start = time.Now()
	environments, err := handler.environmentClusterMappingsService.GetEnvironmentOnlyListForAutocomplete()
	if err != nil {
		handler.logger.Errorw("service err, GetEnvironmentListForAutocomplete at environmentClusterMappingsService layer", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	dbElapsedTime := time.Since(start)
	var grantedEnvironment = environments
	start = time.Now()
	println(dbElapsedTime, grantedEnvironment)
	if !handler.cfg.IgnoreAuthCheck {
		grantedEnvironment = make([]bean2.EnvironmentBean, 0)
		// RBAC enforcer applying
		var envIdentifierList []string
		for index, item := range environments {
			clusterName := strings.ToLower(strings.Split(item.EnvironmentIdentifier, "__")[0])
			if clusterMapping[clusterName].Id != 0 {
				environments[index].CdArgoSetup = clusterMapping[clusterName].IsCdArgoSetup
				environments[index].ClusterName = clusterMapping[clusterName].ClusterName
			}
			envIdentifierList = append(envIdentifierList, strings.ToLower(item.EnvironmentIdentifier))
		}

		result := handler.enforcer.EnforceInBatch(token, casbin.ResourceGlobalEnvironment, casbin.ActionGet, envIdentifierList)
		for _, item := range environments {

			var hasAccess bool
			EnvironmentIdentifier := item.ClusterName + "__" + item.Namespace
			if item.EnvironmentIdentifier != EnvironmentIdentifier {
				// fix for futuristic case
				hasAccess = result[strings.ToLower(EnvironmentIdentifier)] || result[strings.ToLower(item.EnvironmentIdentifier)]
			} else {
				hasAccess = result[strings.ToLower(item.EnvironmentIdentifier)]
			}
			if hasAccess {
				grantedEnvironment = append(grantedEnvironment, item)
			}
		}
		//RBAC enforcer Ends
	}
	elapsedTime := time.Since(start)
	handler.logger.Infow("Env elapsed Time for enforcer", "dbElapsedTime", dbElapsedTime, "elapsedTime",
		elapsedTime, "envSize", len(grantedEnvironment))

	//getting teams for autocomplete
	start = time.Now()
	teams, err := handler.teamService.FetchForAutocomplete()
	if err != nil {
		handler.logger.Errorw("service err, FetchForAutocomplete at teamService layer", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	dbElapsedTime = time.Since(start)
	var grantedTeams = teams
	start = time.Now()
	if !handler.cfg.IgnoreAuthCheck {
		grantedTeams = make([]team.TeamRequest, 0)
		// RBAC enforcer applying
		var teamNameList []string
		for _, item := range teams {
			teamNameList = append(teamNameList, strings.ToLower(item.Name))
		}

		result := handler.enforcer.EnforceInBatch(token, casbin.ResourceTeam, casbin.ActionGet, teamNameList)

		for _, item := range teams {
			if hasAccess := result[strings.ToLower(item.Name)]; hasAccess {
				grantedTeams = append(grantedTeams, item)
			}
		}
	}
	handler.logger.Infow("Team elapsed Time for enforcer", "dbElapsedTime", dbElapsedTime, "elapsedTime", time.Since(start),
		"envSize", len(grantedTeams))

	//RBAC enforcer Ends
	resp := &AppAutocomplete{
		Teams:        grantedTeams,
		Environments: grantedEnvironment,
		Clusters:     granterClusters,
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)

}
