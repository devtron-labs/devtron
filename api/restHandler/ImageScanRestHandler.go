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

package restHandler

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/cluster/environment"
	"github.com/devtron-labs/devtron/pkg/policyGoverance/security/imageScanning"
	securityBean "github.com/devtron-labs/devtron/pkg/policyGoverance/security/imageScanning/bean"
	security2 "github.com/devtron-labs/devtron/pkg/policyGoverance/security/imageScanning/repository"
	"net/http"
	"strconv"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
)

const (
	ObjectTypeApp   = "app"
	ObjectTypeChart = "chart"
	ObjectTypePod   = "pod"
)

type ImageScanRestHandler interface {
	ScanExecutionList(w http.ResponseWriter, r *http.Request)
	FetchExecutionDetail(w http.ResponseWriter, r *http.Request)
	FetchMinScanResultByAppIdAndEnvId(w http.ResponseWriter, r *http.Request)
	VulnerabilityExposure(w http.ResponseWriter, r *http.Request)
}

type ImageScanRestHandlerImpl struct {
	logger             *zap.SugaredLogger
	imageScanService   imageScanning.ImageScanService
	userService        user.UserService
	enforcer           casbin.Enforcer
	enforcerUtil       rbac.EnforcerUtil
	environmentService environment.EnvironmentService
}

func NewImageScanRestHandlerImpl(logger *zap.SugaredLogger,
	imageScanService imageScanning.ImageScanService, userService user.UserService, enforcer casbin.Enforcer,
	enforcerUtil rbac.EnforcerUtil, environmentService environment.EnvironmentService) *ImageScanRestHandlerImpl {
	return &ImageScanRestHandlerImpl{
		logger:             logger,
		imageScanService:   imageScanService,
		userService:        userService,
		enforcer:           enforcer,
		enforcerUtil:       enforcerUtil,
		environmentService: environmentService,
	}
}

func (impl ImageScanRestHandlerImpl) ScanExecutionList(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var request *securityBean.ImageScanRequest
	err = decoder.Decode(&request)
	if err != nil {
		impl.logger.Errorw("request err, ScanExecutionList", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	deployInfoList, err := impl.imageScanService.FetchAllDeployInfo(request)
	if err != nil {
		impl.logger.Errorw("service err, ScanExecutionList", "err", err, "payload", request)
		if util.IsErrNoRows(err) {
			responseList := make([]*securityBean.ImageScanHistoryResponse, 0)
			common.WriteJsonResp(w, nil, &securityBean.ImageScanHistoryListingResponse{ImageScanHistoryResponse: responseList}, http.StatusOK)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}

	token := r.Header.Get("token")
	var ids []int
	var appRBACObjects []string
	var envRBACObjects []string
	var podRBACObjects []string
	podRBACMap := make(map[string]int)

	IdToAppEnvPairs := make(map[int][2]int)
	for _, item := range deployInfoList {
		if item.ScanObjectMetaId > 0 && (item.ObjectType == ObjectTypeApp || item.ObjectType == ObjectTypeChart) {
			IdToAppEnvPairs[item.Id] = [2]int{item.ScanObjectMetaId, item.EnvId}
		}
	}

	appObjects, envObjects, appIdtoApp, envIdToEnv, err := impl.enforcerUtil.GetAppAndEnvRBACNamesByAppAndEnvIds(IdToAppEnvPairs)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	for _, item := range deployInfoList {
		if item.ScanObjectMetaId > 0 && (item.ObjectType == ObjectTypeApp || item.ObjectType == ObjectTypeChart) {
			appObject := appObjects[item.Id]
			envObject := envObjects[item.Id]
			if appObject != "" {
				appRBACObjects = append(appRBACObjects, appObject)
			}
			if envObject != "" {
				envRBACObjects = append(envRBACObjects, envObject)
			}
		} else if item.ScanObjectMetaId > 0 && (item.ObjectType == ObjectTypePod) {
			environments, err := impl.environmentService.GetByClusterId(item.ClusterId)
			if err != nil {
				common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
				return
			}
			for _, environment := range environments {
				podObject := environment.EnvironmentIdentifier
				podRBACObjects = append(podRBACObjects, podObject)
				podRBACMap[podObject] = item.Id
			}
		}
	}

	appResults := impl.enforcer.EnforceInBatch(token, casbin.ResourceApplications, casbin.ActionGet, appRBACObjects)
	envResults := impl.enforcer.EnforceInBatch(token, casbin.ResourceEnvironment, casbin.ActionGet, envRBACObjects)
	podResults := impl.enforcer.EnforceInBatch(token, casbin.ResourceGlobalEnvironment, casbin.ActionGet, podRBACObjects)

	for _, item := range deployInfoList {
		if impl.enforcerUtil.IsAuthorizedForAppInAppResults(item.ScanObjectMetaId, appResults, appIdtoApp) && impl.enforcerUtil.IsAuthorizedForEnvInEnvResults(item.ScanObjectMetaId, item.EnvId, envResults, appIdtoApp, envIdToEnv) {
			ids = append(ids, item.Id)
		}
	}
	for podObject, authorized := range podResults {
		if authorized {
			if itemId, exists := podRBACMap[podObject]; exists {
				ids = append(ids, itemId)
			}
		}
	}

	if ids == nil || len(ids) == 0 {
		responseList := make([]*securityBean.ImageScanHistoryResponse, 0)
		common.WriteJsonResp(w, nil, &securityBean.ImageScanHistoryListingResponse{ImageScanHistoryResponse: responseList}, http.StatusOK)
		return
	}

	results, err := impl.imageScanService.FetchScanExecutionListing(request, ids)
	if err != nil {
		impl.logger.Errorw("service err, ScanExecutionList", "err", err, "payload", request)
		if util.IsErrNoRows(err) {
			responseList := make([]*securityBean.ImageScanHistoryResponse, 0)
			common.WriteJsonResp(w, nil, &securityBean.ImageScanHistoryListingResponse{ImageScanHistoryResponse: responseList}, http.StatusOK)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	common.WriteJsonResp(w, err, results, http.StatusOK)
}

func (impl ImageScanRestHandlerImpl) FetchExecutionDetail(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	v := r.URL.Query()
	var imageScanDeployInfoId, artifactId, appId, envId int
	imageScanDeployInfoIdS := v.Get("imageScanDeployInfoId")
	if len(imageScanDeployInfoIdS) > 0 {
		imageScanDeployInfoId, err = strconv.Atoi(imageScanDeployInfoIdS)
		if err != nil {
			impl.logger.Errorw("request err, FetchExecutionDetail", "err", err, "imageScanDeployInfoIdS", imageScanDeployInfoIdS)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		}
	}
	artifactIdS := v.Get("artifactId")
	if len(artifactIdS) > 0 {
		artifactId, err = strconv.Atoi(artifactIdS)
		if err != nil {
			impl.logger.Errorw("request err, FetchExecutionDetail", "err", err, "artifactIdS", artifactIdS)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		}
	}
	appIds := v.Get("appId")
	if len(appIds) > 0 {
		appId, err = strconv.Atoi(appIds)
		if err != nil {
			impl.logger.Errorw("request err, FetchExecutionDetail", "err", err, "appIds", appIds)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		}
	}
	envIds := v.Get("envId")
	if len(envIds) > 0 {
		envId, err = strconv.Atoi(envIds)
		if err != nil {
			impl.logger.Errorw("request err, FetchExecutionDetail", "err", err, "envIds", envIds)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		}
	}
	image := v.Get("image")
	request := &securityBean.ImageScanRequest{
		ImageScanDeployInfoId: imageScanDeployInfoId,
		Image:                 image,
		ArtifactId:            artifactId,
		AppId:                 appId,
		EnvId:                 envId,
	}

	executionDetail, err := impl.imageScanService.FetchExecutionDetailResult(request)
	if err != nil {
		impl.logger.Errorw("service err, FetchExecutionDetail", "err", err, "payload", request)
		if util.IsErrNoRows(err) {
			common.WriteJsonResp(w, nil, &securityBean.ImageScanExecutionDetail{}, http.StatusOK)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}

	token := r.Header.Get("token")
	if executionDetail != nil {
		//RBAC
		if executionDetail.AppId > 0 && executionDetail.EnvId > 0 {
			object := impl.enforcerUtil.GetAppRBACNameByAppId(appId)
			if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
				common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
				return
			}
			object = impl.enforcerUtil.GetEnvRBACNameByAppId(appId, envId)
			if ok := impl.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, object); !ok {
				common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
				return
			}
		} else if executionDetail.AppId > 0 {
			object := impl.enforcerUtil.GetAppRBACNameByAppId(appId)
			if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
				common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
				return
			}
		} else {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		}
		//RBAC
	} else {
		common.WriteJsonResp(w, err, &securityBean.ImageScanExecutionDetail{}, http.StatusOK)
	}

	common.WriteJsonResp(w, err, executionDetail, http.StatusOK)
}

func (impl ImageScanRestHandlerImpl) FetchMinScanResultByAppIdAndEnvId(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	var appId, envId int
	request := &securityBean.ImageScanRequest{}
	appIds := v.Get("appId")
	if len(appIds) > 0 {
		appId, err := strconv.Atoi(appIds)
		if err != nil {
			impl.logger.Errorw("request err, FetchMinScanResultByAppIdAndEnvId", "err", err, "appIds", appIds)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		}
		request.AppId = appId
	}
	envIds := v.Get("envId")
	if len(envIds) > 0 {
		envId, err := strconv.Atoi(envIds)
		if err != nil {
			impl.logger.Errorw("request err, FetchMinScanResultByAppIdAndEnvId", "err", err, "envIds", envIds)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		}
		request.EnvId = envId
	}

	//RBAC
	token := r.Header.Get("token")
	if appId > 0 && envId > 0 {
		object := impl.enforcerUtil.GetAppRBACNameByAppId(appId)
		if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
		object = impl.enforcerUtil.GetEnvRBACNameByAppId(appId, envId)
		if ok := impl.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, object); !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
	}
	//RBAC

	result, err := impl.imageScanService.FetchMinScanResultByAppIdAndEnvId(request)
	if err != nil {
		impl.logger.Errorw("service err, FetchMinScanResultByAppIdAndEnvId", "err", err, "payload", request)
		if util.IsErrNoRows(err) {
			err = &util.ApiError{InternalMessage: err.Error(), UserMessage: "no data found"}
			common.WriteJsonResp(w, err, nil, http.StatusOK)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}
	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (impl ImageScanRestHandlerImpl) VulnerabilityExposure(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var request *security2.VulnerabilityRequest
	err = decoder.Decode(&request)
	if err != nil {
		impl.logger.Errorw("request err, VulnerabilityExposure", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	results, err := impl.imageScanService.VulnerabilityExposure(request)
	if err != nil {
		impl.logger.Errorw("service err, VulnerabilityExposure", "err", err, "payload", request)
		if util.IsErrNoRows(err) {
			responseList := make([]*securityBean.ImageScanHistoryResponse, 0)
			common.WriteJsonResp(w, nil, &securityBean.ImageScanHistoryListingResponse{ImageScanHistoryResponse: responseList}, http.StatusOK)
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		}
		return
	}

	//RBAC
	token := r.Header.Get("token")
	var vulnerabilityExposure []*security2.VulnerabilityExposure
	for _, item := range results.VulnerabilityExposure {
		object := impl.enforcerUtil.GetAppRBACNameByAppId(item.AppId)
		if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
		object = impl.enforcerUtil.GetEnvRBACNameByAppId(item.AppId, item.EnvId)
		if ok := impl.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, object); ok {
			vulnerabilityExposure = append(vulnerabilityExposure, item)
		}
	}
	//RBAC
	results.VulnerabilityExposure = vulnerabilityExposure
	common.WriteJsonResp(w, err, results, http.StatusOK)
}
