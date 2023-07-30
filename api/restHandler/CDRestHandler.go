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

package restHandler

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/util/ArgoUtil"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/gorilla/mux"
	"github.com/xyproto/unzip"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type CDRestHandler interface {
	FetchResourceTree(w http.ResponseWriter, r *http.Request)

	FetchPodContainerLogs(w http.ResponseWriter, r *http.Request)
	UploadKustomizeHandler(w http.ResponseWriter, r *http.Request)
}

type CDRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	resourceService ArgoUtil.ResourceService
	appService      app.AppService
	userAuthService user.UserService
}

func NewCDRestHandlerImpl(logger *zap.SugaredLogger, resourceService ArgoUtil.ResourceService, appService app.AppService, userService user.UserService) *CDRestHandlerImpl {
	cdRestHandler := &CDRestHandlerImpl{logger: logger, resourceService: resourceService, appService: appService, userAuthService: userService}
	return cdRestHandler
}

func (handler CDRestHandlerImpl) UploadKustomizeHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	_, err = strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, "Bad Request", http.StatusBadRequest)
		return
	}
	_, err = strconv.Atoi(vars["envId"])
	if err != nil {
		common.WriteJsonResp(w, err, "Bad Request", http.StatusBadRequest)
		return
	}
	r.ParseMultipartForm(10 << 20) // Set the maximum upload size to 10 MB
	file, fileHandler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	uploadDir := "/tmp/uploads"
	os.MkdirAll(uploadDir, os.ModePerm)

	zipFilePath := filepath.Join(uploadDir, fileHandler.Filename)
	outFile, err := os.Create(zipFilePath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()
	io.Copy(outFile, file)

	unzipDir := filepath.Join(uploadDir, strings.TrimSuffix(fileHandler.Filename, filepath.Ext(fileHandler.Filename)))
	err = unzip.Extract(zipFilePath, unzipDir)
	//err = unzip1(zipFilePath, unzipDir)
	if err != nil {
		http.Error(w, "Error unzipping the folder", http.StatusInternalServerError)
		return
	}
	handler.appService.UploadKustomizeData()
	fmt.Fprintf(w, "Folder uploaded and extracted successfully.")
}

func (handler CDRestHandlerImpl) FetchResourceTree(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appName := vars["app-name"]

	res, err := handler.resourceService.FetchResourceTree(appName)
	if err != nil {
		handler.logger.Errorw("request err, FetchResourceTree", "err", err, "appName", appName)
	}
	resJson, err := json.Marshal(res)
	_, err = w.Write(resJson)
	if err != nil {
		handler.logger.Errorw("request err, FetchResourceTree", "err", err, "appName", appName, "response", res)
	}
}

func (handler CDRestHandlerImpl) FetchPodContainerLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appName := vars["app-name"]
	podName := vars["pod-name"]

	res, err := handler.resourceService.FetchPodContainerLogs(appName, podName, ArgoUtil.PodContainerLogReq{})
	if err != nil {
		handler.logger.Errorw("service err, FetchPodContainerLogs", "err", err, "appName", appName, "podName", podName)
	}
	resJson, err := json.Marshal(res)
	_, err = w.Write(resJson)
	if err != nil {
		handler.logger.Errorw("service err, FetchPodContainerLogs", "err", err, "appName", appName, "podName", podName)
	}
}
