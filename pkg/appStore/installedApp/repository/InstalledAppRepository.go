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

package repository

import (
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s/health"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/timelineStatus"
	util2 "github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	util3 "github.com/devtron-labs/devtron/pkg/appStore/util"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/gitUtil"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
	"strconv"
	"time"
)

// InstalledApps TODO: remove the deprecated column GitOpsRepoName
type InstalledApps struct {
	TableName                  struct{}                              `sql:"installed_apps" pg:",discard_unknown_columns"`
	Id                         int                                   `sql:"id,pk"`
	AppId                      int                                   `sql:"app_id,notnull"`
	EnvironmentId              int                                   `sql:"environment_id,notnull"`
	Active                     bool                                  `sql:"active, notnull"`
	GitOpsRepoName             string                                `sql:"git_ops_repo_name"` // Deprecated; currently in use for backward compatibility
	GitOpsRepoUrl              string                                `sql:"git_ops_repo_url"`  // GitOpsRepoName has been migrated to GitOpsRepoUrl; Make sure to migrate from GitOpsRepoName for future flows
	IsCustomRepository         bool                                  `sql:"is_custom_repository"`
	DeploymentAppType          string                                `sql:"deployment_app_type"`
	Status                     appStoreBean.AppstoreDeploymentStatus `sql:"status"`
	DeploymentAppDeleteRequest bool                                  `sql:"deployment_app_delete_request"`
	Notes                      string                                `json:"notes"`
	App                        app.App
	Environment                repository.Environment
	sql.AuditLog
}

func (model *InstalledApps) MarkActive() {
	model.Active = true
}

func (model *InstalledApps) MarkInActive() {
	model.Active = false
}

func (model *InstalledApps) UpdateStatus(status appStoreBean.AppstoreDeploymentStatus) {
	model.Status = status
}

func (model *InstalledApps) UpdateGitOpsRepository(gitOpsRepoUrl string, isCustomRepository bool) {
	if gitOpsRepoUrl == "" {
		return
	}
	model.GitOpsRepoUrl = gitOpsRepoUrl
	model.IsCustomRepository = isCustomRepository
	model.GitOpsRepoName = gitUtil.GetGitRepoNameFromGitRepoUrl(gitOpsRepoUrl) // Handled for backward compatibility
}

func (model *InstalledApps) ChangeAppNameToDisplayName() {
	model.App.AppName = model.App.DisplayName
}

func (model *InstalledApps) GetUniqueAppNameIdentifier() string {
	if util3.IsExternalChartStoreApp(model.App.DisplayName) {
		//if display name is set then that installedApp's app is already migrated
		return model.App.AppName
	}
	return fmt.Sprintf("%s-%s-%s", model.App.AppName, model.Environment.Namespace, strconv.Itoa(model.Environment.ClusterId))
}

type InstalledAppVersions struct {
	TableName                    struct{} `sql:"installed_app_versions" pg:",discard_unknown_columns"`
	Id                           int      `sql:"id,pk"`
	InstalledAppId               int      `sql:"installed_app_id,notnull"`
	AppStoreApplicationVersionId int      `sql:"app_store_application_version_id,notnull"`
	ValuesYaml                   string   `sql:"values_yaml_raw"`
	Active                       bool     `sql:"active, notnull"`
	ReferenceValueId             int      `sql:"reference_value_id"`
	ReferenceValueKind           string   `sql:"reference_value_kind"`
	sql.AuditLog
	InstalledApp               InstalledApps
	AppStoreApplicationVersion appStoreDiscoverRepository.AppStoreApplicationVersion
}

func (model *InstalledAppVersions) MarkActive() {
	model.Active = true
}

func (model *InstalledAppVersions) MarkInActive() {
	model.Active = false
}

type InstalledAppRepository interface {
	CreateInstalledApp(model *InstalledApps, tx *pg.Tx) (*InstalledApps, error)
	CreateInstalledAppVersion(model *InstalledAppVersions, tx *pg.Tx) (*InstalledAppVersions, error)
	UpdateInstalledApp(model *InstalledApps, tx *pg.Tx) (*InstalledApps, error)
	UpdateInstalledAppVersion(model *InstalledAppVersions, tx *pg.Tx) (*InstalledAppVersions, error)
	GetInstalledApp(id int) (*InstalledApps, error)
	GetInstalledAppsByAppId(appId int) (InstalledApps, error)
	GetInstalledAppVersion(id int) (*InstalledAppVersions, error)
	GetInstalledAppVersionAny(id int) (*InstalledAppVersions, error)
	GetAllInstalledApps(filter *appStoreBean.AppStoreFilter) ([]InstalledAppsWithChartDetails, error)
	GetAllInstalledAppsByAppStoreId(appStoreId int) ([]InstalledAppAndEnvDetails, error)
	GetAllInstalledAppsByChartRepoId(chartRepoId int) ([]InstalledAppAndEnvDetails, error)
	GetInstalledAppVersionByInstalledAppIdAndEnvId(installedAppId int, envId int) (*InstalledAppVersions, error)
	FetchNotes(installedAppId int) (*InstalledApps, error)
	GetInstalledAppVersionByAppStoreId(appStoreId int) ([]*InstalledAppVersions, error)
	DeleteInstalledApp(model *InstalledApps) (*InstalledApps, error)
	DeleteInstalledAppVersion(model *InstalledAppVersions) (*InstalledAppVersions, error)
	GetInstalledAppVersionByInstalledAppId(id int) ([]*InstalledAppVersions, error)
	GetConnection() (dbConnection *pg.DB)
	GetInstalledAppVersionByInstalledAppIdMeta(installedAppId int) ([]*InstalledAppVersions, error)
	GetActiveInstalledAppVersionByInstalledAppId(installedAppId int) (*InstalledAppVersions, error)
	GetLatestInstalledAppVersionByGitHash(gitHash string) (*InstalledAppVersions, error)
	GetClusterComponentByClusterId(clusterId int) ([]*InstalledApps, error)     //unused
	GetClusterComponentByClusterIds(clusterIds []int) ([]*InstalledApps, error) //unused
	GetInstalledAppVersionByAppIdAndEnvId(appId int, envId int) (*InstalledAppVersions, error)
	GetInstalledAppVersionByClusterIds(clusterIds []int) ([]*InstalledAppVersions, error) //unused
	GetInstalledAppVersionByClusterIdsV2(clusterIds []int) ([]*InstalledAppVersions, error)
	GetInstalledApplicationByClusterIdAndNamespaceAndAppName(clusterId int, namespace string, appName string) (*InstalledApps, error)
	GetAppAndEnvDetailsForDeploymentAppTypeInstalledApps(deploymentAppType string, clusterIds []int) ([]*InstalledApps, error)
	GetDeploymentSuccessfulStatusCountForTelemetry() (int, error)
	GetGitOpsInstalledAppsWhereArgoAppDeletedIsTrue(installedAppId int, envId int) (InstalledApps, error)
	GetInstalledAppByGitHash(gitHash string) (InstallAppDeleteRequest, error)
	GetInstalledAppByAppName(appName string) (*InstalledApps, error)
	GetInstalledAppByAppIdAndDeploymentType(appId int, deploymentAppType string) (InstalledApps, error)
	GetInstalledAppByInstalledAppVersionId(installedAppVersionId int) (InstalledApps, error)
	// GetInstalledAppByGitOpsAppName will return all the active installed_apps with matching `app_name-environment_name`
	GetInstalledAppByGitOpsAppName(acdAppName string) (*InstalledApps, error)
	GetInstalledAppByGitRepoUrl(repoName, repoUrl string) (*InstalledApps, error)

	GetArgoPipelinesHavingLatestTriggerStuckInNonTerminalStatusesForAppStore(getPipelineDeployedBeforeMinutes int, getPipelineDeployedWithinHours int) ([]*InstalledAppVersions, error)
	GetArgoPipelinesHavingTriggersStuckInLastPossibleNonTerminalTimelinesForAppStore(pendingSinceSeconds int, timeForDegradation int) ([]*InstalledAppVersions, error)
	GetHelmReleaseStatusConfigByInstalledAppId(installedAppVersionHistoryId int) (string, string, error)

	GetActiveInstalledAppByEnvIdAndDeploymentType(envId int, deploymentType string, excludeAppIds []string, includeAppIds []string) ([]*InstalledApps, error)
	UpdateDeploymentAppTypeInInstalledApp(deploymentAppType string, installedAppIdIncludes []int, userId int32, deployStatus int) error
	FindInstalledAppByIds(ids []int) ([]*InstalledApps, error)
	// FindInstalledAppsByAppId returns multiple installed apps for an appId, this only happens for external-apps with same name installed in diff namespaces
	FindInstalledAppsByAppId(appId int) ([]*InstalledApps, error)
}

type InstalledAppRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewInstalledAppRepositoryImpl(Logger *zap.SugaredLogger, dbConnection *pg.DB) *InstalledAppRepositoryImpl {
	return &InstalledAppRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

func (impl InstalledAppRepositoryImpl) CreateInstalledApp(model *InstalledApps, tx *pg.Tx) (*InstalledApps, error) {
	err := tx.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl InstalledAppRepositoryImpl) CreateInstalledAppVersion(model *InstalledAppVersions, tx *pg.Tx) (*InstalledAppVersions, error) {
	err := tx.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl InstalledAppRepositoryImpl) UpdateInstalledApp(model *InstalledApps, tx *pg.Tx) (*InstalledApps, error) {
	err := tx.Update(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl InstalledAppRepositoryImpl) UpdateInstalledAppVersion(model *InstalledAppVersions, tx *pg.Tx) (*InstalledAppVersions, error) {
	var err error
	if tx == nil {
		err = impl.dbConnection.Update(model)
	} else {
		err = tx.Update(model)
	}
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}
func (impl InstalledAppRepositoryImpl) FetchNotes(installedAppId int) (*InstalledApps, error) {
	model := &InstalledApps{}
	err := impl.dbConnection.Model(model).
		Column("installed_apps.*", "App").
		Where("installed_apps.id = ?", installedAppId).Where("installed_apps.active = true").Select()
	return model, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledApp(id int) (*InstalledApps, error) {
	model := &InstalledApps{}
	err := impl.dbConnection.Model(model).
		Column("installed_apps.*", "App", "Environment", "App.Team", "Environment.Cluster").
		Where("installed_apps.id = ?", id).Where("installed_apps.active = true").Select()
	return model, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppsByAppId(appId int) (InstalledApps, error) {
	var model InstalledApps
	err := impl.dbConnection.Model(&model).
		Column("installed_apps.*", "App", "Environment", "App.Team", "Environment.Cluster").
		Where("installed_apps.app_id = ?", appId).Where("installed_apps.active = true").Select()
	return model, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppVersionByAppStoreId(appStoreId int) ([]*InstalledAppVersions, error) {
	var model []*InstalledAppVersions
	err := impl.dbConnection.Model(&model).
		Column("installed_app_versions.*", "InstalledApp", "InstalledApp.App", "InstalledApp.Environment", "AppStoreApplicationVersion").
		Where("app_store_application_version.app_store_id = ?", appStoreId).
		Where("installed_app_versions.active = true").Select()
	if err != nil {
		return model, err
	}
	for _, installedAppVersion := range model {
		appStore := &appStoreDiscoverRepository.AppStore{}
		err = impl.dbConnection.
			Model(appStore).
			Column("app_store.*", "ChartRepo", "DockerArtifactStore", "DockerArtifactStore.OCIRegistryConfig").
			Where("app_store.id = ? ", installedAppVersion.AppStoreApplicationVersion.AppStoreId).
			Relation("DockerArtifactStore.OCIRegistryConfig", func(q *orm.Query) (query *orm.Query, err error) {
				return q.Where("deleted IS FALSE and " +
					"repository_type='CHART' and " +
					"(repository_action='PULL' or repository_action='PULL/PUSH')"), nil
			}).
			Select()
		if err != nil {
			return model, err
		}
		installedAppVersion.AppStoreApplicationVersion.AppStore = appStore
	}
	return model, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppVersionByInstalledAppIdMeta(installedAppId int) ([]*InstalledAppVersions, error) {
	var model []*InstalledAppVersions
	err := impl.dbConnection.Model(&model).
		Column("installed_app_versions.*", "InstalledApp", "InstalledApp.App", "InstalledApp.Environment", "AppStoreApplicationVersion").
		Where("installed_app_versions.installed_app_id = ?", installedAppId).
		Order("installed_app_versions.id desc").
		Select()
	if err != nil {
		return model, err
	}
	for _, installedAppVersion := range model {
		appStore := &appStoreDiscoverRepository.AppStore{}
		err = impl.dbConnection.
			Model(appStore).
			Column("app_store.*", "ChartRepo", "DockerArtifactStore", "DockerArtifactStore.OCIRegistryConfig").
			Where("app_store.id = ? ", installedAppVersion.AppStoreApplicationVersion.AppStoreId).
			Relation("DockerArtifactStore.OCIRegistryConfig", func(q *orm.Query) (query *orm.Query, err error) {
				return q.Where("deleted IS FALSE and " +
					"repository_type='CHART' and " +
					"(repository_action='PULL' or repository_action='PULL/PUSH')"), nil
			}).
			Select()
		if err != nil {
			return model, err
		}
		installedAppVersion.AppStoreApplicationVersion.AppStore = appStore
	}
	return model, err
}

func (impl InstalledAppRepositoryImpl) GetActiveInstalledAppVersionByInstalledAppId(installedAppId int) (*InstalledAppVersions, error) {
	model := &InstalledAppVersions{}
	err := impl.dbConnection.Model(model).
		Column("installed_app_versions.*", "InstalledApp", "InstalledApp.App", "InstalledApp.Environment", "AppStoreApplicationVersion").
		Where("installed_app_versions.installed_app_id = ?", installedAppId).
		Where("installed_app_versions.active = true").Order("installed_app_versions.id desc").Limit(1).Select()
	if err != nil {
		return model, err
	}
	appStore := &appStoreDiscoverRepository.AppStore{}
	err = impl.dbConnection.
		Model(appStore).
		Column("app_store.*", "ChartRepo", "DockerArtifactStore", "DockerArtifactStore.OCIRegistryConfig").
		Where("app_store.id = ? ", model.AppStoreApplicationVersion.AppStoreId).
		Relation("DockerArtifactStore.OCIRegistryConfig", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("deleted IS FALSE and " +
				"repository_type='CHART' and " +
				"(repository_action='PULL' or repository_action='PULL/PUSH')"), nil
		}).
		Select()
	if err != nil {
		return model, err
	}
	model.AppStoreApplicationVersion.AppStore = appStore
	return model, err
}

func (impl InstalledAppRepositoryImpl) GetLatestInstalledAppVersionByGitHash(gitHash string) (*InstalledAppVersions, error) {
	model := &InstalledAppVersions{}
	err := impl.dbConnection.Model(model).
		Column("installed_app_versions.*", "InstalledApp").
		Column("AppStoreApplicationVersion.AppStore.ChartRepo").
		Where("installed_app_versions.git_hash = ?", gitHash).
		Where("installed_app_versions.active = true").Order("installed_app_versions.id desc").Limit(1).Select()
	return model, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppVersion(id int) (*InstalledAppVersions, error) {
	model := &InstalledAppVersions{}
	err := impl.dbConnection.Model(model).
		Column("installed_app_versions.*", "InstalledApp", "InstalledApp.App", "InstalledApp.Environment", "InstalledApp.Environment.Cluster", "AppStoreApplicationVersion", "InstalledApp.App.Team").
		Where("installed_app_versions.id = ?", id).Where("installed_app_versions.active = true").Select()
	if err != nil {
		return model, err
	}
	appStore := &appStoreDiscoverRepository.AppStore{}
	err = impl.dbConnection.
		Model(appStore).
		Column("app_store.*", "ChartRepo", "DockerArtifactStore", "DockerArtifactStore.OCIRegistryConfig").
		Where("app_store.id = ? ", model.AppStoreApplicationVersion.AppStoreId).
		Relation("DockerArtifactStore.OCIRegistryConfig", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("deleted IS FALSE and " +
				"repository_type='CHART' and " +
				"(repository_action='PULL' or repository_action='PULL/PUSH')"), nil
		}).
		Select()
	if err != nil {
		return model, err
	}
	model.AppStoreApplicationVersion.AppStore = appStore
	return model, err
}

// it returns enable and disabled both version
func (impl InstalledAppRepositoryImpl) GetInstalledAppVersionAny(id int) (*InstalledAppVersions, error) {
	model := &InstalledAppVersions{}
	err := impl.dbConnection.Model(model).
		Column("installed_app_versions.*", "InstalledApp", "InstalledApp.App", "AppStoreApplicationVersion").
		Where("installed_app_versions.id = ?", id).Select()
	if err != nil {
		return model, err
	}
	appStore := &appStoreDiscoverRepository.AppStore{}
	err = impl.dbConnection.
		Model(appStore).
		Column("app_store.*", "ChartRepo", "DockerArtifactStore", "DockerArtifactStore.OCIRegistryConfig").
		Where("app_store.id = ? ", model.AppStoreApplicationVersion.AppStoreId).
		Relation("DockerArtifactStore.OCIRegistryConfig", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("deleted IS FALSE and " +
				"repository_type='CHART' and " +
				"(repository_action='PULL' or repository_action='PULL/PUSH')"), nil
		}).
		Select()
	if err != nil {
		return model, err
	}
	model.AppStoreApplicationVersion.AppStore = appStore
	return model, err
}

func (impl InstalledAppRepositoryImpl) GetAllInstalledApps(filter *appStoreBean.AppStoreFilter) ([]InstalledAppsWithChartDetails, error) {
	var installedAppsWithChartDetails []InstalledAppsWithChartDetails
	var query string
	query = "select iav.updated_on, iav.id as installed_app_version_id, ch.name as chart_repo_name, das.id as docker_artifact_store_id,"
	query = query + " env.environment_name, env.id as environment_id, env.is_virtual_environment, a.app_name, a.display_name, a.app_offering_mode, asav.icon, asav.name as app_store_application_name,"
	query = query + " env.namespace, cluster.cluster_name, a.team_id, cluster.id as cluster_id, "
	query = query + " asav.id as app_store_application_version_id, ia.id , asav.deprecated , app_status.status as app_status, ia.deployment_app_delete_request"
	query = query + " from installed_app_versions iav"
	query = query + " inner join installed_apps ia on iav.installed_app_id = ia.id"
	query = query + " inner join app a on a.id = ia.app_id"
	query = query + " inner join environment env on ia.environment_id = env.id"
	query = query + " inner join cluster on env.cluster_id = cluster.id"
	query = query + " inner join app_store_application_version asav on iav.app_store_application_version_id = asav.id"
	query = query + " inner join app_store aps on aps.id = asav.app_store_id"
	query = query + " left join chart_repo ch on ch.id = aps.chart_repo_id"
	query = query + " left join docker_artifact_store das on das.id = aps.docker_artifact_store_id"
	query = query + " left join app_status on app_status.app_id = ia.app_id and ia.environment_id = app_status.env_id"
	query = query + " where ia.active = true and iav.active = true"
	if filter.OnlyDeprecated {
		query = query + " AND asav.deprecated = TRUE"
	}
	if len(filter.AppStoreName) > 0 {
		query = query + " AND aps.name LIKE '%" + filter.AppStoreName + "%'"
	}
	if len(filter.AppName) > 0 {
		query = query + " AND a.app_name LIKE '%" + filter.AppName + "%'"
	}
	if len(filter.ChartRepoId) > 0 {
		query = query + " AND ch.id IN (" + sqlIntSeq(filter.ChartRepoId) + ")"
	}
	if len(filter.EnvIds) > 0 {
		query = query + " AND env.id IN (" + sqlIntSeq(filter.EnvIds) + ")"
	}
	if len(filter.ClusterIds) > 0 {
		query = query + " AND cluster.id IN (" + sqlIntSeq(filter.ClusterIds) + ")"
	}
	if len(filter.AppStatuses) > 0 {
		appStatuses := util.ProcessAppStatuses(filter.AppStatuses)
		query = query + " and app_status.status IN (" + appStatuses + ") "
	}
	query = query + " ORDER BY aps.name ASC"
	if filter.Size > 0 {
		query = query + " OFFSET " + strconv.Itoa(filter.Offset) + " LIMIT " + strconv.Itoa(filter.Size) + ""
	}
	query = query + ";"
	var err error
	_, err = impl.dbConnection.Query(&installedAppsWithChartDetails, query)
	if err != nil {
		return nil, err
	}
	return installedAppsWithChartDetails, err
}

func (impl InstalledAppRepositoryImpl) GetAllInstalledAppsByAppStoreId(appStoreId int) ([]InstalledAppAndEnvDetails, error) {
	var installedAppAndEnvDetails []InstalledAppAndEnvDetails
	var queryTemp = "select env.environment_name, env.id as environment_id, a.app_name, a.display_name, a.app_offering_mode, ia.updated_on, u.email_id," +
		" asav.id as app_store_application_version_id, iav.id as installed_app_version_id, ia.id as installed_app_id, ia.app_id, ia.deployment_app_type, app_status.status as app_status" +
		" from installed_app_versions iav inner join installed_apps ia on iav.installed_app_id = ia.id" +
		" inner join app a on a.id = ia.app_id " +
		" inner join app_store_application_version asav on iav.app_store_application_version_id = asav.id " +
		" inner join app_store aps on asav.app_store_id = aps.id " +
		" inner join environment env on ia.environment_id = env.id " +
		" left join users u on u.id = ia.updated_by " +
		" left join app_status on app_status.app_id = ia.app_id and ia.environment_id = app_status.env_id\n" +
		" where aps.id = " + strconv.Itoa(appStoreId) + " and ia.active=true and iav.active=true and env.active=true"
	_, err := impl.dbConnection.Query(&installedAppAndEnvDetails, queryTemp)
	if err != nil {
		return nil, err
	}
	return installedAppAndEnvDetails, err
}

func (impl InstalledAppRepositoryImpl) GetAllInstalledAppsByChartRepoId(chartRepoId int) ([]InstalledAppAndEnvDetails, error) {
	var installedAppAndEnvDetails []InstalledAppAndEnvDetails
	var queryTemp = "select env.environment_name, env.id as environment_id, a.app_name, ia.updated_on, u.email_id, asav.id as app_store_application_version_id, iav.id as installed_app_version_id, ia.id as installed_app_id " +
		" from installed_app_versions iav inner join installed_apps ia on iav.installed_app_id = ia.id" +
		" inner join app a on a.id = ia.app_id " +
		" inner join app_store_application_version asav on iav.app_store_application_version_id = asav.id " +
		" inner join app_store aps on asav.app_store_id = aps.id " +
		" inner join environment env on ia.environment_id = env.id " +
		" left join users u on u.id = ia.updated_by " +
		" where aps.chart_repo_id = ? and ia.active=true and iav.active=true and env.active=true"
	_, err := impl.dbConnection.Query(&installedAppAndEnvDetails, queryTemp, chartRepoId)
	if err != nil {
		return nil, err
	}
	return installedAppAndEnvDetails, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppVersionByInstalledAppIdAndEnvId(installedAppId int, envId int) (*InstalledAppVersions, error) {
	installedAppVersion := &InstalledAppVersions{}
	err := impl.dbConnection.
		Model(installedAppVersion).
		Column("installed_app_versions.*", "InstalledApp", "InstalledApp.App", "InstalledApp.Environment", "InstalledApp.Environment.Cluster", "AppStoreApplicationVersion").
		Join("inner join installed_apps ia on ia.id = installed_app_versions.installed_app_id").
		Where("ia.id = ?", installedAppId).
		Where("ia.environment_id = ?", envId).
		Where("ia.active = true").Where("installed_app_versions.active = true").
		Limit(1).
		Select()
	if err != nil {
		return installedAppVersion, err
	}
	appStore := &appStoreDiscoverRepository.AppStore{}
	err = impl.dbConnection.
		Model(appStore).
		Column("app_store.*", "ChartRepo", "DockerArtifactStore", "DockerArtifactStore.OCIRegistryConfig").
		Where("app_store.id = ? ", installedAppVersion.AppStoreApplicationVersion.AppStoreId).
		Relation("DockerArtifactStore.OCIRegistryConfig", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("deleted IS FALSE and " +
				"repository_type='CHART' and " +
				"(repository_action='PULL' or repository_action='PULL/PUSH')"), nil
		}).
		Select()
	if err != nil {
		return installedAppVersion, err
	}
	installedAppVersion.AppStoreApplicationVersion.AppStore = appStore
	return installedAppVersion, err
}

func sqlIntSeq(ns []int) string {
	if len(ns) == 0 {
		return ""
	}
	estimate := len(ns) * 4
	b := make([]byte, 0, estimate)
	for _, n := range ns {
		b = strconv.AppendInt(b, int64(n), 10)
		b = append(b, ',')
	}
	b = b[:len(b)-1]
	return string(b)
}

func (impl InstalledAppRepositoryImpl) DeleteInstalledApp(model *InstalledApps) (*InstalledApps, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl InstalledAppRepositoryImpl) DeleteInstalledAppVersion(model *InstalledAppVersions) (*InstalledAppVersions, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppVersionByInstalledAppId(installedAppId int) ([]*InstalledAppVersions, error) {
	model := make([]*InstalledAppVersions, 0)
	err := impl.dbConnection.Model(&model).
		Column("installed_app_versions.*").
		Where("installed_app_versions.installed_app_id = ?", installedAppId).
		Where("installed_app_versions.active = true").Select()

	return model, err
}

func (impl *InstalledAppRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return impl.dbConnection
}

func (impl InstalledAppRepositoryImpl) GetClusterComponentByClusterId(clusterId int) ([]*InstalledApps, error) {
	var models []*InstalledApps
	err := impl.dbConnection.Model(&models).
		Column("installed_apps.*", "App", "Environment").
		Where("environment.cluster_id = ?", clusterId).
		Where("installed_apps.active = ?", true).
		Where("environment.active = ?", true).
		Select()
	return models, err
}

func (impl InstalledAppRepositoryImpl) GetClusterComponentByClusterIds(clusterIds []int) ([]*InstalledApps, error) {
	var models []*InstalledApps
	err := impl.dbConnection.Model(&models).
		Column("installed_apps.*", "App", "Environment").
		Where("environment.cluster_id in (?)", pg.In(clusterIds)).
		Where("installed_apps.active = ?", true).
		Where("environment.active = ?", true).
		Select()
	return models, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppVersionByAppIdAndEnvId(appId int, envId int) (*InstalledAppVersions, error) {
	installedAppVersion := &InstalledAppVersions{}
	err := impl.dbConnection.
		Model(installedAppVersion).
		Column("installed_app_versions.*", "InstalledApp", "InstalledApp.App", "InstalledApp.Environment", "AppStoreApplicationVersion").
		Join("inner join installed_apps ia on ia.id = installed_app_versions.installed_app_id").
		Where("ia.app_id = ?", appId).
		Where("ia.environment_id = ?", envId).
		Where("ia.active = true").Where("installed_app_versions.active = true").
		Order("installed_app_versions.id DESC").
		Limit(1).
		Select()
	if err != nil {
		return installedAppVersion, err
	}
	appStore := &appStoreDiscoverRepository.AppStore{}
	err = impl.dbConnection.
		Model(appStore).
		Column("app_store.*", "ChartRepo", "DockerArtifactStore", "DockerArtifactStore.OCIRegistryConfig").
		Where("app_store.id = ? ", installedAppVersion.AppStoreApplicationVersion.AppStoreId).
		Relation("DockerArtifactStore.OCIRegistryConfig", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("deleted IS FALSE and " +
				"repository_type='CHART' and " +
				"(repository_action='PULL' or repository_action='PULL/PUSH')"), nil
		}).
		Select()
	if err != nil {
		return installedAppVersion, err
	}
	installedAppVersion.AppStoreApplicationVersion.AppStore = appStore
	return installedAppVersion, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppVersionByClusterIds(clusterIds []int) ([]*InstalledAppVersions, error) {
	var installedAppVersions []*InstalledAppVersions
	err := impl.dbConnection.
		Model(&installedAppVersions).
		Column("installed_app_versions.*", "InstalledApp", "InstalledApp.App", "InstalledApp.Environment", "AppStoreApplicationVersion").
		Join("inner join installed_apps ia on ia.id = installed_app_versions.installed_app_id").
		Join("inner join environment env on env.id = ia.environment_id").
		Where("ia.active = true").Where("installed_app_versions.active = true").
		Where("env.cluster_id in (?)", pg.In(clusterIds)).Where("env.active = ?", true).
		Order("installed_app_versions.id desc").
		Select()
	if err != nil {
		return installedAppVersions, err
	}
	for _, installedAppVersion := range installedAppVersions {
		appStore := &appStoreDiscoverRepository.AppStore{}
		err = impl.dbConnection.
			Model(appStore).
			Column("app_store.*", "ChartRepo", "DockerArtifactStore", "DockerArtifactStore.OCIRegistryConfig").
			Where("app_store.id = ? ", installedAppVersion.AppStoreApplicationVersion.AppStoreId).
			Relation("DockerArtifactStore.OCIRegistryConfig", func(q *orm.Query) (query *orm.Query, err error) {
				return q.Where("deleted IS FALSE and " +
					"repository_type='CHART' and " +
					"(repository_action='PULL' or repository_action='PULL/PUSH')"), nil
			}).
			Select()
		if err != nil {
			return installedAppVersions, err
		}
		installedAppVersion.AppStoreApplicationVersion.AppStore = appStore
	}
	return installedAppVersions, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppVersionByClusterIdsV2(clusterIds []int) ([]*InstalledAppVersions, error) {
	var installedAppVersions []*InstalledAppVersions
	err := impl.dbConnection.
		Model(&installedAppVersions).
		Column("installed_app_versions.*", "InstalledApp", "InstalledApp.App", "InstalledApp.Environment", "AppStoreApplicationVersion").
		Join("inner join installed_apps ia on ia.id = installed_app_versions.installed_app_id").
		Join("inner join cluster_installed_apps cia on cia.installed_app_id = ia.id").
		Where("ia.active = true").Where("installed_app_versions.active = true").Where("cia.cluster_id in (?)", pg.In(clusterIds)).
		Order("installed_app_versions.id desc").
		Select()
	if err != nil {
		return installedAppVersions, err
	}
	for _, installedAppVersion := range installedAppVersions {
		appStore := &appStoreDiscoverRepository.AppStore{}
		err = impl.dbConnection.
			Model(appStore).
			Column("app_store.*", "ChartRepo", "DockerArtifactStore", "DockerArtifactStore.OCIRegistryConfig").
			Where("app_store.id = ? ", installedAppVersion.AppStoreApplicationVersion.AppStoreId).
			Relation("DockerArtifactStore.OCIRegistryConfig", func(q *orm.Query) (query *orm.Query, err error) {
				return q.Where("deleted IS FALSE and " +
					"repository_type='CHART' and " +
					"(repository_action='PULL' or repository_action='PULL/PUSH')"), nil
			}).
			Select()
		if err != nil {
			return installedAppVersions, err
		}
		installedAppVersion.AppStoreApplicationVersion.AppStore = appStore
	}
	return installedAppVersions, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledApplicationByClusterIdAndNamespaceAndAppName(clusterId int, namespace string, appName string) (*InstalledApps, error) {
	model := &InstalledApps{}
	err := impl.dbConnection.Model(model).
		Column("installed_apps.*", "App", "Environment", "App.Team").
		Where("environment.cluster_id = ?", clusterId).
		Where("environment.namespace = ?", namespace).
		Where("app.app_name = ?", appName).
		Where("installed_apps.active = ?", true).
		Where("app.active = ?", true).
		Where("environment.active = ?", true).
		Select()
	return model, err
}

func (impl InstalledAppRepositoryImpl) GetAppAndEnvDetailsForDeploymentAppTypeInstalledApps(deploymentAppType string, clusterIds []int) ([]*InstalledApps, error) {
	var installedApps []*InstalledApps
	err := impl.dbConnection.
		Model(&installedApps).
		Column("installed_apps.id", "App.app_name", "App.display_name", "Environment.cluster_id", "Environment.namespace").
		Where("environment.cluster_id in (?)", pg.In(clusterIds)).
		Where("installed_apps.deployment_app_type = ?", deploymentAppType).
		Where("app.active = ?", true).
		Where("installed_apps.active = ?", true).
		Select()
	return installedApps, err
}

func (impl InstalledAppRepositoryImpl) GetDeploymentSuccessfulStatusCountForTelemetry() (int, error) {

	countQuery := "select count(Id) from installed_apps where status=?;"
	var count int
	_, err := impl.dbConnection.Query(&count, countQuery, appStoreBean.DEPLOY_SUCCESS)
	if err != nil {
		impl.Logger.Errorw("unable to get deployment count of successfully deployed Helm apps")
	}
	return count, err
}

func (impl InstalledAppRepositoryImpl) GetGitOpsInstalledAppsWhereArgoAppDeletedIsTrue(installedAppId int, envId int) (InstalledApps, error) {
	var installedApps InstalledApps
	err := impl.dbConnection.Model(&installedApps).
		Column("installed_apps.*", "App.app_name", "Environment.namespace", "Environment.cluster_id", "Environment.environment_name").
		Where("deployment_app_delete_request = ?", true).
		Where("installed_apps.active = ?", true).
		Where("installed_apps.id = ?", installedAppId).
		Where("installed_apps.environment_id = ?", envId).
		Where("deployment_app_type = ?", util2.PIPELINE_DEPLOYMENT_TYPE_ACD).
		Select()
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in fetching pipeline while udating delete status", "err", err)
		return installedApps, err
	}
	return installedApps, nil
}
func (impl InstalledAppRepositoryImpl) GetInstalledAppByGitHash(gitHash string) (InstallAppDeleteRequest, error) {
	model := InstallAppDeleteRequest{}
	query := "select iv.installed_app_id, a.app_name, i.app_id, i.environment_id, a.app_offering_mode, e.cluster_id, e.namespace " +
		" from app a inner join installed_apps i on a.id=i.app_id  " +
		"inner join installed_app_versions iv on i.id=iv.installed_app_id " +
		"inner join installed_app_version_history ivh on ivh.installed_app_version_id=iv.id " +
		"inner join environment e on e.id=i.environment_id where ivh.git_hash=?;"
	_, err := impl.dbConnection.Query(&model, query, gitHash)
	if err != nil {
		impl.Logger.Errorw("error in getting delete request data", "err", err)
		return model, err
	}
	return model, nil
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppByAppIdAndDeploymentType(appId int, deploymentAppType string) (InstalledApps, error) {
	var installedApps InstalledApps
	queryString := `select * from installed_apps where active=? and app_id=? and deployment_app_type=?;`
	_, err := impl.dbConnection.Query(&installedApps, queryString, true, appId, deploymentAppType)
	if err != nil {
		impl.Logger.Errorw("error in fetching InstalledApp", "err", err)
		return installedApps, err
	}

	return installedApps, nil
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppByAppName(appName string) (*InstalledApps, error) {
	model := &InstalledApps{}
	err := impl.dbConnection.Model(model).
		Column("installed_apps.*", "App", "Environment").
		Where("app.app_name = ?", appName).
		Where("installed_apps.active = true").
		Limit(1).
		Select()
	return model, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppByInstalledAppVersionId(installedAppVersionId int) (InstalledApps, error) {
	var installedApps InstalledApps
	queryString := `select ia.* from installed_apps ia inner join installed_app_versions iav on ia.id=iav.installed_app_id
         			where iav.active=? and iav.id=? and ia.deployment_app_type=?;`
	_, err := impl.dbConnection.Query(&installedApps, queryString, true, installedAppVersionId, util2.PIPELINE_DEPLOYMENT_TYPE_ACD)
	if err != nil {
		impl.Logger.Errorw("error in fetching InstalledApp", "err", err)
		return installedApps, err
	}

	return installedApps, nil
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppByGitOpsAppName(acdAppName string) (*InstalledApps, error) {
	model := &InstalledApps{}
	err := impl.dbConnection.Model(model).
		Column("installed_apps.*", "App", "Environment").
		// TODO add deployment_app_name filed in installed_apps table
		Where("CONCAT(app.app_name, ?, environment.environment_name) = ?", "-", acdAppName).
		Where("installed_apps.active = true").
		Where("environment.active = true").
		Limit(1).
		Select()
	return model, err
}

func (impl InstalledAppRepositoryImpl) GetInstalledAppByGitRepoUrl(repoName, repoUrl string) (*InstalledApps, error) {
	model := &InstalledApps{}
	err := impl.dbConnection.Model(model).
		Column("installed_apps.*", "App").
		Where("installed_apps.git_ops_repo_name = ? OR installed_apps.git_ops_repo_url = ?", repoName, repoUrl).
		Where("installed_apps.active = true").
		Where("app.active = true").
		Limit(1).
		Select()
	return model, err
}

func (impl InstalledAppRepositoryImpl) GetArgoPipelinesHavingLatestTriggerStuckInNonTerminalStatusesForAppStore(getPipelineDeployedBeforeMinutes int, getPipelineDeployedWithinHours int) ([]*InstalledAppVersions, error) {
	var installedAppVersions []*InstalledAppVersions
	queryString := `select iav.* from installed_app_versions iav 
    				inner join installed_apps ia on iav.installed_app_id=ia.id 
    				inner join installed_app_version_history iavh on iavh.installed_app_version_id=iav.id 
             		where iavh.id in (select DISTINCT ON (installed_app_version_id) max(id) as id from installed_app_version_history 
                         where updated_on < NOW() - INTERVAL '? minutes' and updated_on > NOW() - INTERVAL '? hours' and status not in (?)
                         group by installed_app_version_id, id order by installed_app_version_id, id desc ) and ia.deployment_app_type=? and iav.active=?;`

	_, err := impl.dbConnection.Query(&installedAppVersions, queryString, getPipelineDeployedBeforeMinutes, getPipelineDeployedWithinHours,
		pg.In([]string{pipelineConfig.WorkflowAborted, pipelineConfig.WorkflowFailed, pipelineConfig.WorkflowSucceeded, string(health.HealthStatusHealthy), string(health.HealthStatusDegraded)}),
		util2.PIPELINE_DEPLOYMENT_TYPE_ACD, true)
	if err != nil {
		impl.Logger.Errorw("error in GetArgoPipelinesHavingLatestTriggerStuckInNonTerminalStatusesForAppStore", "err", err)
		return nil, err
	}
	return installedAppVersions, nil
}

func (impl InstalledAppRepositoryImpl) GetArgoPipelinesHavingTriggersStuckInLastPossibleNonTerminalTimelinesForAppStore(pendingSinceSeconds int, timeForDegradation int) ([]*InstalledAppVersions, error) {
	var installedAppVersions []*InstalledAppVersions
	queryString := `select iav.* from installed_app_versions iav inner join installed_apps ia on iav.installed_app_id=ia.id 
					inner join installed_app_version_history iavh on iavh.installed_app_version_id=iav.id
					where iavh.id in (select DISTINCT ON (installed_app_version_history_id) max(id) as id from pipeline_status_timeline
					                    where status in (?) and status_time < NOW() - INTERVAL '? seconds'
										group by installed_app_version_history_id, id order by installed_app_version_history_id, id desc)
					and iavh.updated_on > NOW() - INTERVAL '? minutes' and ia.deployment_app_type=? and iav.active=?;`

	_, err := impl.dbConnection.Query(&installedAppVersions, queryString,
		pg.In([]timelineStatus.TimelineStatus{timelineStatus.TIMELINE_STATUS_KUBECTL_APPLY_SYNCED,
			timelineStatus.TIMELINE_STATUS_FETCH_TIMED_OUT, timelineStatus.TIMELINE_STATUS_UNABLE_TO_FETCH_STATUS}),
		pendingSinceSeconds, timeForDegradation, util2.PIPELINE_DEPLOYMENT_TYPE_ACD, true)
	if err != nil {
		impl.Logger.Errorw("error in GetArgoPipelinesHavingTriggersStuckInLastPossibleNonTerminalTimelinesForAppStore", "err", err)
		return nil, err
	}
	return installedAppVersions, nil
}

func (impl InstalledAppRepositoryImpl) GetHelmReleaseStatusConfigByInstalledAppId(installedAppVersionHistoryId int) (string, string, error) {
	installStatus := struct {
		HelmReleaseStatusConfig string
		Status                  string
	}{}
	queryString := `select helm_release_status_config, installed_app_version_history.status  from installed_app_version_history inner join installed_app_versions on installed_app_version_history.installed_app_version_id=installed_app_versions.id inner join installed_apps on installed_apps.id=installed_app_versions.installed_app_id where installed_apps.id = ? order by installed_app_version_history.created_on desc limit 1;`
	_, err := impl.dbConnection.Query(&installStatus, queryString, installedAppVersionHistoryId)
	if err != nil {
		impl.Logger.Errorw("error in GetAllGitOpsDeploymentAppName", "err", err)
		return installStatus.HelmReleaseStatusConfig, "", err
	}
	return installStatus.HelmReleaseStatusConfig, installStatus.Status, nil
}

func (impl InstalledAppRepositoryImpl) GetActiveInstalledAppByEnvIdAndDeploymentType(envId int, deploymentType string,
	excludeAppIds []string, includeAppIds []string) ([]*InstalledApps, error) {
	var installedApps []*InstalledApps

	query := impl.dbConnection.
		Model(&installedApps).
		Column("installed_apps.*", "App", "Environment").
		Join("inner join app a on installed_apps.app_id = a.id").
		Where("installed_apps.environment_id = ?", envId).
		Where("installed_apps.deployment_app_type = ?", deploymentType).
		Where("installed_apps.active = ?", true)

	if len(excludeAppIds) > 0 {
		query.Where("installed_apps.app_id not in (?)", pg.In(excludeAppIds))
	}

	if len(includeAppIds) > 0 {
		query.Where("installed_apps.app_id in (?)", pg.In(includeAppIds))
	}

	err := query.Select()
	if err != nil {
		return nil, err
	}
	return installedApps, nil
}

// UpdateDeploymentAppTypeInInstalledApp takes in deploymentAppType and list of installedAppIds and
// updates the deployment_app_type in the table for given ids.
func (impl InstalledAppRepositoryImpl) UpdateDeploymentAppTypeInInstalledApp(deploymentAppType string, installedAppIdIncludes []int, userId int32, deployStatus int) error {
	query := "update installed_apps set deployment_app_type = ?,updated_by = ?, updated_on = ?, status = ? where id in (?);"
	var installedApp *InstalledApps
	_, err := impl.dbConnection.Query(installedApp, query, deploymentAppType, userId, time.Now(), deployStatus, pg.In(installedAppIdIncludes))

	return err
}

func (impl InstalledAppRepositoryImpl) FindInstalledAppByIds(ids []int) ([]*InstalledApps, error) {
	var installedApps []*InstalledApps
	err := impl.dbConnection.Model(&installedApps).
		Column("installed_apps.*", "App", "Environment", "Environment.Cluster").
		Join("inner join app a on installed_apps.app_id = a.id").
		Join("inner join environment e on installed_apps.environment_id = e.id").
		Join("inner join cluster c on c.id = e.cluster_id").
		Where("installed_apps.id in (?)", pg.In(ids)).
		Where("installed_apps.active = true").
		Select()
	if err != nil {
		impl.Logger.Errorw("error on fetching installed apps by ids", "ids", ids)
	}
	return installedApps, err
}

func (impl InstalledAppRepositoryImpl) FindInstalledAppsByAppId(appId int) ([]*InstalledApps, error) {
	var installedApps []*InstalledApps
	err := impl.dbConnection.Model(&installedApps).
		Column("installed_apps.*", "App", "Environment").
		Join("inner join app a on installed_apps.app_id = a.id").
		Join("inner join environment e on installed_apps.environment_id = e.id").
		Where("installed_apps.app_id = ?", appId).
		Where("installed_apps.active = true").
		Select()
	if err != nil {
		impl.Logger.Errorw("error on fetching installed apps by appId", "appId", appId)
	}
	return installedApps, err
}
