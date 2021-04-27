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

package external

import (
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/appstore"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/external"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/user"
	"go.uber.org/zap"
	"time"
)

type ExternalAppsDto struct {
	Id             int       `json:"id"`
	AppName        string    `json:"appName"`
	Label          string    `json:"label"`
	ChartName      string    `json:"chartName"`
	Namespace      string    `json:"namespace"`
	ClusterId      int       `json:"clusterId"`
	LastDeployedOn time.Time `json:"lastDeployedOn"`
	Active         bool      `json:"active"`
	UserId         int32     `json:"-"`
}

type ExternalAppsService interface {
	SearchExternalAppsByFilter() ([]*ExternalAppsDto, error)
	Create(request *ExternalAppsDto) (*ExternalAppsDto, error)
	Update(request *ExternalAppsDto) (*ExternalAppsDto, error)
	FindById(id int) (*ExternalAppsDto, error)
	FindAll() ([]*ExternalAppsDto, error)
}

type ExternalAppsServiceImpl struct {
	logger                        *zap.SugaredLogger
	appStoreRepository            appstore.AppStoreRepository
	appStoreApplicationRepository appstore.AppStoreApplicationVersionRepository
	installedAppRepository        appstore.InstalledAppRepository
	userService                   user.UserService
	repoRepository                chartConfig.ChartRepoRepository
	K8sUtil                       *util.K8sUtil
	clusterService                cluster.ClusterService
	envService                    cluster.EnvironmentService
	versionService                argocdServer.VersionService
	aCDAuthConfig                 *user.ACDAuthConfig
	externalAppsRepository        external.ExternalAppsRepository
}

func NewExternalAppsServiceImpl(logger *zap.SugaredLogger, appStoreRepository appstore.AppStoreRepository,
	appStoreApplicationRepository appstore.AppStoreApplicationVersionRepository, installedAppRepository appstore.InstalledAppRepository,
	userService user.UserService, repoRepository chartConfig.ChartRepoRepository, K8sUtil *util.K8sUtil,
	clusterService cluster.ClusterService, envService cluster.EnvironmentService,
	versionService argocdServer.VersionService, aCDAuthConfig *user.ACDAuthConfig, externalAppsRepository external.ExternalAppsRepository) *ExternalAppsServiceImpl {
	return &ExternalAppsServiceImpl{
		logger:                        logger,
		appStoreRepository:            appStoreRepository,
		appStoreApplicationRepository: appStoreApplicationRepository,
		installedAppRepository:        installedAppRepository,
		userService:                   userService,
		repoRepository:                repoRepository,
		K8sUtil:                       K8sUtil,
		clusterService:                clusterService,
		envService:                    envService,
		versionService:                versionService,
		aCDAuthConfig:                 aCDAuthConfig,
		externalAppsRepository:        externalAppsRepository,
	}
}

func (impl *ExternalAppsServiceImpl) SearchExternalAppsByFilter() ([]*ExternalAppsDto, error) {

	return nil, nil
}

func (impl *ExternalAppsServiceImpl) Create(request *ExternalAppsDto) (*ExternalAppsDto, error) {

	externalApps := &external.ExternalApps{AuditLog: models.AuditLog{CreatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: request.UserId},}
	externalApps.AppName = request.AppName
	externalApps.Label = request.Label
	externalApps.ChartName = request.ChartName
	externalApps.Namespace = request.Namespace
	externalApps.LastDeployedOn = request.LastDeployedOn
	externalApps.Active = request.Active
	externalApps.Active = true
	externalApps, err := impl.externalAppsRepository.Create(externalApps)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}

	return request, nil
}

func (impl *ExternalAppsServiceImpl) Update(request *ExternalAppsDto) (*ExternalAppsDto, error) {

	externalApp, err := impl.externalAppsRepository.FindById(request.Id)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	externalApp.AppName = request.AppName
	externalApp.Label = request.Label
	externalApp.ChartName = request.ChartName
	externalApp.Namespace = request.Namespace
	externalApp.LastDeployedOn = request.LastDeployedOn
	externalApp.Active = request.Active
	externalApp.UpdatedOn = time.Now()
	externalApp, err = impl.externalAppsRepository.Update(externalApp)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}

	return request, nil
}

func (impl *ExternalAppsServiceImpl) FindById(id int) (*ExternalAppsDto, error) {
	externalApp := &ExternalAppsDto{}
	externalAppModel, err := impl.externalAppsRepository.FindById(id)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	externalApp.AppName = externalAppModel.AppName
	externalApp.Label = externalAppModel.Label
	externalApp.ChartName = externalAppModel.ChartName
	externalApp.Namespace = externalAppModel.Namespace
	externalApp.LastDeployedOn = externalAppModel.LastDeployedOn
	externalApp.Active = externalAppModel.Active
	return externalApp, nil
}

func (impl *ExternalAppsServiceImpl) FindAll() ([]*ExternalAppsDto, error) {
	var externalApps []*ExternalAppsDto
	models, err := impl.externalAppsRepository.FindAll()
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	for _, externalAppModel := range models {
		externalApp := &ExternalAppsDto{}
		externalApp.AppName = externalAppModel.AppName
		externalApp.Label = externalAppModel.Label
		externalApp.ChartName = externalAppModel.ChartName
		externalApp.Namespace = externalAppModel.Namespace
		externalApp.LastDeployedOn = externalAppModel.LastDeployedOn
		externalApp.Active = externalAppModel.Active
		externalApps = append(externalApps, externalApp)
	}
	return externalApps, nil
}
