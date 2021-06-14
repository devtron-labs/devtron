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

/*
	@description: app listing view
*/
package repository

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

type AppListingRepository interface {
	FetchAppsByEnvironment(appListingFilter helper.AppListingFilter) ([]*bean.AppEnvironmentContainer, error)
	DeploymentDetailsByAppIdAndEnvId(appId int, envId int) (bean.DeploymentDetailContainer, error)
	FetchAppDetail(appId int, envId int) (bean.AppDetailContainer, error)

	FetchAppTriggerView(appId int) ([]bean.TriggerView, error)
	FetchAppStageStatus(appId int) ([]bean.AppStageStatus, error)

	//Not in used
	PrometheusApiByEnvId(id int) (*string, error)

	FetchOtherEnvironment(appId int) ([]*bean.Environment, error)

	SaveNewDeployment(deploymentStatus *DeploymentStatus, tx *pg.Tx) error
	FindLastDeployedStatus(appName string) (DeploymentStatus, error)
	FindLastDeployedStatuses(appNames []string) ([]DeploymentStatus, error)
	FindLastDeployedStatusesForAllApps() ([]DeploymentStatus, error)
	DeploymentDetailByArtifactId(ciArtifactId int) (bean.DeploymentDetailContainer, error)
	FindAppCount(isProd bool) (int, error)
}

type DeploymentStatus struct {
	tableName struct{}  `sql:"deployment_status" pg:",discard_unknown_columns"`
	Id        int       `sql:"id,pk"`
	AppName   string    `sql:"app_name,notnull"`
	Status    string    `sql:"status,notnull"`
	AppId     int       `sql:"app_id"`
	EnvId     int       `sql:"env_id"`
	CreatedOn time.Time `sql:"created_on"`
	UpdatedOn time.Time `sql:"updated_on"`
}

const NewDeployment string = "Deployment Initiated"

type AppListingRepositoryImpl struct {
	dbConnection                     *pg.DB
	Logger                           *zap.SugaredLogger
	appListingRepositoryQueryBuilder helper.AppListingRepositoryQueryBuilder
}

func NewAppListingRepositoryImpl(Logger *zap.SugaredLogger, dbConnection *pg.DB, appListingRepositoryQueryBuilder helper.AppListingRepositoryQueryBuilder) *AppListingRepositoryImpl {
	return &AppListingRepositoryImpl{dbConnection: dbConnection, Logger: Logger, appListingRepositoryQueryBuilder: appListingRepositoryQueryBuilder}
}

/**
It will return the list of filtered apps with details related to each env
*/
func (impl AppListingRepositoryImpl) FetchAppsByEnvironment(appListingFilter helper.AppListingFilter) ([]*bean.AppEnvironmentContainer, error) {
	impl.Logger.Debug("reached at FetchAppsByEnvironment:")
	var appEnvArr []*bean.AppEnvironmentContainer

	query := impl.appListingRepositoryQueryBuilder.BuildAppListingQueryLastDeploymentTime()
	impl.Logger.Debugw("basic app detail query: ", query)
	var lastDeployedTimeDTO []*bean.AppEnvironmentContainer
	lastDeployedTimeMap := map[int]*bean.AppEnvironmentContainer{}
	_, err := impl.dbConnection.Query(&lastDeployedTimeDTO, query)
	if err != nil {
		impl.Logger.Error(err)
		return appEnvArr, err
	}
	for _, item := range lastDeployedTimeDTO {
		if _, ok := lastDeployedTimeMap[item.PipelineId]; ok {
			continue
		}
		lastDeployedTimeMap[item.PipelineId] = &bean.AppEnvironmentContainer{
			LastDeployedTime: item.LastDeployedTime,
			DataSource:       item.DataSource,
			MaterialInfoJson: item.MaterialInfoJson,
			CiArtifactId:     item.CiArtifactId,
		}
	}

	var appEnvContainer []*bean.AppEnvironmentContainer
	appsEnvquery := impl.appListingRepositoryQueryBuilder.BuildAppListingQuery(appListingFilter)
	impl.Logger.Debugw("basic app detail query: ", appsEnvquery)
	_, appsErr := impl.dbConnection.Query(&appEnvContainer, appsEnvquery)
	if appsErr != nil {
		impl.Logger.Error(appsErr)
		return appEnvContainer, appsErr
	}

	latestDeploymentStatusMap := map[string]*bean.AppEnvironmentContainer{}
	for _, item := range appEnvContainer {
		if item.EnvironmentId > 0 && item.PipelineId > 0 && item.Active == false {
			// skip adding apps which have linked with cd pipeline and that environment has marked as deleted.
			continue
		}
		// include only apps which are not linked with any cd pipeline + those linked with cd pipeline and env has active.
		key := strconv.Itoa(item.AppId) + "_" + strconv.Itoa(item.EnvironmentId)
		if _, ok := latestDeploymentStatusMap[key]; ok {
			continue
		}

		if lastDeployedTime, ok := lastDeployedTimeMap[item.PipelineId]; ok {
			item.LastDeployedTime = lastDeployedTime.LastDeployedTime
			item.DataSource = lastDeployedTime.DataSource
			item.MaterialInfoJson = lastDeployedTime.MaterialInfoJson
			item.CiArtifactId = lastDeployedTime.CiArtifactId
		}

		if len(item.DataSource) > 0 {
			mInfo, err := parseMaterialInfo([]byte(item.MaterialInfoJson), item.DataSource)
			if err == nil {
				item.MaterialInfo = mInfo
			} else {
				item.MaterialInfo = []byte("[]")
			}
			item.MaterialInfoJson = ""
		} else {
			item.MaterialInfo = []byte("[]")
			item.MaterialInfoJson = ""
		}
		appEnvArr = append(appEnvArr, item)
		latestDeploymentStatusMap[key] = item
	}

	return appEnvArr, nil
}

//It will return the deployment detail of any cd pipeline which is latest triggered for Environment of any App
func (impl AppListingRepositoryImpl) DeploymentDetailsByAppIdAndEnvId(appId int, envId int) (bean.DeploymentDetailContainer, error) {
	impl.Logger.Debugf("reached at AppListingRepository:")
	var deploymentDetail bean.DeploymentDetailContainer
	query := "SELECT env.environment_name, a.app_name, ceco.namespace, u.email_id as last_deployed_by" +
		" , cia.material_info, pco.created_on AS last_deployed_time, pco.pipeline_release_counter as release_version" +
		" , env.default, cia.data_source, p.pipeline_name as last_deployed_pipeline" +
		" FROM chart_env_config_override ceco" +
		" INNER JOIN environment env ON env.id=ceco.target_environment" +
		" INNER JOIN pipeline_config_override pco ON pco.env_config_override_id = ceco.id" +
		" INNER JOIN pipeline p on p.id = pco.pipeline_id" +
		" INNER JOIN ci_artifact cia on cia.id = pco.ci_artifact_id" +
		" INNER JOIN app a ON a.id=p.app_id" +
		" LEFT JOIN users u on u.id=pco.created_by" +
		" WHERE a.app_store is false AND a.id=? AND env.id=? AND p.deleted = FALSE AND env.active = TRUE" +
		" ORDER BY pco.created_on desc limit 1;"
	impl.Logger.Debugf("query:", query)
	_, err := impl.dbConnection.Query(&deploymentDetail, query, appId, envId)
	if err != nil {
		impl.Logger.Errorw("Exception caught:", err)
		return deploymentDetail, err
	}

	mInfo, err := parseMaterialInfo(deploymentDetail.MaterialInfo, deploymentDetail.DataSource)
	if err == nil {
		deploymentDetail.MaterialInfo = mInfo
	} else {
		deploymentDetail.MaterialInfo = []byte("[]")
	}
	deploymentDetail.EnvironmentId = envId
	return deploymentDetail, nil
}

func parseMaterialInfo(materialInfo json.RawMessage, source string) (json.RawMessage, error) {
	if source != "GOCD" && source != "CI-RUNNER" && source != "EXTERNAL" {
		return nil, fmt.Errorf("datasource: %s not supported", source)
	}
	if materialInfo == nil {
		return []byte("[]"), nil
	}
	var ciMaterials []*CiMaterialInfo
	err := json.Unmarshal(materialInfo, &ciMaterials)
	if err != nil {
		return []byte("[]"), err
	}
	var scmMaps []map[string]string
	for _, material := range ciMaterials {
		materialMap := map[string]string{}
		var url string
		if material.Material.Type == "git" {
			url = material.Material.GitConfiguration.URL
		} else if material.Material.Type == "scm" {
			url = material.Material.ScmConfiguration.URL
		} else {
			return nil, fmt.Errorf("unknown material type:%s ", material.Material.Type)
		}

		if material.Modifications != nil && len(material.Modifications) > 0 {
			revision := material.Modifications[0].Revision
			url = strings.TrimSpace(url)
			materialMap["url"] = url
			materialMap["revision"] = revision
			materialMap["modifiedTime"] = material.Modifications[0].ModifiedTime
			materialMap["author"] = material.Modifications[0].Author
			materialMap["message"] = material.Modifications[0].Message
			materialMap["branch"] = material.Modifications[0].Branch
		}
		scmMaps = append(scmMaps, materialMap)
	}
	mInfo, err := json.Marshal(scmMaps)
	return mInfo, err
}

func (impl AppListingRepositoryImpl) FetchAppDetail(appId int, envId int) (bean.AppDetailContainer, error) {
	impl.Logger.Debugf("reached at AppListingRepository:")
	var appDetailContainer bean.AppDetailContainer
	//Fetch deployment detail of cd pipeline latest triggered withing env of any App.
	deploymentDetail, err := impl.DeploymentDetailsByAppIdAndEnvId(appId, envId)
	if err != nil {
		impl.Logger.Warn("unable to fetch deployment detail for app")
	}

	// other environment tab
	var otherEnvironments []bean.Environment
	query := "SELECT p.environment_id,env.environment_name from pipeline p" +
		" INNER JOIN environment env on env.id=p.environment_id" +
		" where p.app_id=? and p.deleted = FALSE AND env.active = TRUE GROUP by 1,2"
	impl.Logger.Debugw("other env query:", query)
	_, err = impl.dbConnection.Query(&otherEnvironments, query, appId)

	//linkOuts, err := impl.fetchLinkOutsByAppIdAndEnvId(appId, deploymentDetail.EnvironmentId)

	//appDetailContainer.LinkOuts = linkOuts
	appDetailContainer.Environments = otherEnvironments
	appDetailContainer.DeploymentDetailContainer = deploymentDetail
	return appDetailContainer, nil
}

func (impl AppListingRepositoryImpl) fetchLinkOutsByAppIdAndEnvId(appId int, envId int) ([]string, error) {
	impl.Logger.Debug("reached at AppListingRepository:")

	var linkOuts []string
	query := "SELECT ael.link from app_env_linkouts ael where ael.app_id=? and ael.environment_id=?"
	impl.Logger.Debugw("lingOut query:", query)
	_, err := impl.dbConnection.Query(&linkOuts, query, appId, envId)
	if err != nil {
		impl.Logger.Errorw("err", err)
	}
	return linkOuts, err
}

func (impl AppListingRepositoryImpl) PrometheusApiByEnvId(id int) (*string, error) {
	impl.Logger.Debug("reached at PrometheusApiByEnvId:")
	var prometheusEndpoint string
	query := "SELECT env.prometheus_endpoint from environment env" +
		" WHERE env.id = ? AND env.active = TRUE"
	impl.Logger.Debugw("query", query)
	//environments := []string{"QA"}
	_, err := impl.dbConnection.Query(&prometheusEndpoint, query, id)
	if err != nil {
		impl.Logger.Error("Exception caught:", err)
		return nil, err
	}
	return &prometheusEndpoint, nil
}

func (impl AppListingRepositoryImpl) FetchAppTriggerView(appId int) ([]bean.TriggerView, error) {
	impl.Logger.Debug("reached at FetchAppTriggerView:")
	var triggerView []bean.TriggerView
	var triggerViewResponse []bean.TriggerView
	query := "" +
		" SELECT cp.id as ci_pipeline_id,cp.name as ci_pipeline_name,p.id as cd_pipeline_id," +
		" p.pipeline_name as cd_pipeline_name, a.app_name, env.environment_name" +
		" FROM pipeline p" +
		" INNER JOIN ci_pipeline cp on cp.id = p.ci_pipeline_id" +
		" INNER JOIN app a ON a.id = p.app_id" +
		" INNER JOIN environment env on env.id = p.environment_id" +
		" WHERE p.app_id=? and p.deleted=false and a.app_store is false AND env.active = TRUE;"

	impl.Logger.Debugw("query", query)
	_, err := impl.dbConnection.Query(&triggerView, query, appId)
	if err != nil {
		impl.Logger.Errorw("Exception caught:", err)
		return nil, err
	}

	for _, item := range triggerView {
		if item.CdPipelineId > 0 {
			var tView bean.TriggerView
			statusQuery := "SELECT p.id as cd_pipeline_id, pco.created_on as last_deployed_time," +
				" u.email_id as last_deployed_by, pco.pipeline_release_counter as release_version," +
				" cia.material_info, cia.data_source, evt.reason as status" +
				" FROM pipeline p" +
				" INNER JOIN pipeline_config_override pco ON pco.pipeline_id = p.id" +
				" INNER JOIN ci_artifact cia on cia.id = pco.ci_artifact_id" +
				" LEFT JOIN events evt ON evt.release_version=CAST(pco.pipeline_release_counter AS varchar)" +
				" AND evt.pipeline_name=p.pipeline_name" +
				" LEFT JOIN users u on u.id=pco.created_by" +
				" WHERE p.id = ?" +
				" ORDER BY evt.created_on desc, pco.created_on desc LIMIT 1;"

			impl.Logger.Debugw("statusQuery", statusQuery)
			_, err := impl.dbConnection.Query(&tView, statusQuery, item.CdPipelineId)
			if err != nil {
				impl.Logger.Errorw("error while fetching trigger detail", "err", err)
			}
			if err == nil && tView.CdPipelineId > 0 {
				if tView.Status != "" {
					item.Status = tView.Status
					item.StatusMessage = tView.StatusMessage
				} else {
					item.Status = "Initiated"
				}

				item.LastDeployedTime = tView.LastDeployedTime
				item.LastDeployedBy = tView.LastDeployedBy
				item.ReleaseVersion = tView.ReleaseVersion
				item.DataSource = tView.DataSource
				item.MaterialInfo = tView.MaterialInfo
				mInfo, err := parseMaterialInfo(tView.MaterialInfo, tView.DataSource)
				if err == nil {
					item.MaterialInfo = mInfo
				} else {
					item.MaterialInfo = []byte("[]")
				}
			}
		}
		triggerViewResponse = append(triggerViewResponse, item)
	}

	return triggerViewResponse, nil
}

func (impl AppListingRepositoryImpl) FetchAppStageStatus(appId int) ([]bean.AppStageStatus, error) {
	impl.Logger.Debug("reached at AppListingRepository:")
	var appStageStatus []bean.AppStageStatus

	var stages struct {
		AppId        int  `json:"app_id,omitempty"`
		CiTemplateId int  `json:"ci_template_id,omitempty"`
		CiPipelineId int  `json:"ci_pipeline_id,omitempty"`
		ChartId      int  `json:"chart_id,omitempty"`
		PipelineId   int  `json:"pipeline_id,omitempty"`
		YamlStatus   int  `json:"yaml_status,omitempty"`
		YamlReviewed bool `json:"yaml_reviewed,omitempty"`
	}

	query := "SELECT " +
		" app.id as app_id, ct.id as ci_template_id, cp.id as ci_pipeline_id, ch.id as chart_id," +
		" p.id as pipeline_id, ceco.status as yaml_status, ceco.reviewed as yaml_reviewed " +
		" FROM app app" +
		" LEFT JOIN ci_template ct on ct.app_id=app.id" +
		" LEFT JOIN ci_pipeline cp on cp.app_id=app.id" +
		" LEFT JOIN charts ch on ch.app_id=app.id" +
		" LEFT JOIN pipeline p on p.app_id=app.id" +
		" LEFT JOIN chart_env_config_override ceco on ceco.chart_id=ch.id" +
		" WHERE app.id=? and app.app_store is false;"

	impl.Logger.Debugw("last app stages status query:", "query", query)

	_, err := impl.dbConnection.Query(&stages, query, appId)
	if err != nil {
		impl.Logger.Errorw("error:", err)
		return appStageStatus, err
	}

	var isMaterialExists bool
	materialQuery := "select exists(select 1 from git_material gm where gm.app_id=? and gm.active is TRUE)"
	impl.Logger.Debugw("material stage status query:", "query", query)

	_, err = impl.dbConnection.Query(&isMaterialExists, materialQuery, appId)
	if err != nil {
		impl.Logger.Errorw("error:", err)
		return appStageStatus, err
	}
	materialExists := 0
	if isMaterialExists {
		materialExists = 1
	}

	appStageStatus = append(appStageStatus, impl.makeAppStageStatus(0, "APP", stages.AppId))
	appStageStatus = append(appStageStatus, impl.makeAppStageStatus(1, "MATERIAL", materialExists))
	appStageStatus = append(appStageStatus, impl.makeAppStageStatus(2, "TEMPLATE", stages.CiTemplateId))
	appStageStatus = append(appStageStatus, impl.makeAppStageStatus(3, "CI_PIPELINE", stages.CiPipelineId))
	appStageStatus = append(appStageStatus, impl.makeAppStageStatus(4, "CHART", stages.ChartId))
	appStageStatus = append(appStageStatus, impl.makeAppStageStatus(5, "CD_PIPELINE", stages.PipelineId))
	appStageStatus = append(appStageStatus, bean.AppStageStatus{
		Stage:     6,
		StageName: "CHART_ENV_CONFIG",
		Status: func() bool {
			if stages.YamlStatus == 3 && stages.YamlReviewed == true {
				return true
			} else {
				return false
			}
		}(),
		Required: true,
	})

	return appStageStatus, nil
}

func (impl AppListingRepositoryImpl) makeAppStageStatus(stage int, stageName string, id int) bean.AppStageStatus {
	return bean.AppStageStatus{
		Stage:     stage,
		StageName: stageName,
		Status: func() bool {
			if id > 0 {
				return true
			} else {
				return false
			}
		}(),
		Required: true,
	}
}

func (impl AppListingRepositoryImpl) FetchOtherEnvironment(appId int) ([]*bean.Environment, error) {
	impl.Logger.Debug("reached at FetchOtherEnvironment:")

	// other environment tab
	var otherEnvironments []*bean.Environment
	query := "SELECT p.environment_id,env.environment_name, env_app_m.app_metrics, env.default as prod, env_app_m.infra_metrics from pipeline p" +
		" INNER JOIN environment env on env.id=p.environment_id" +
		" LEFT JOIN env_level_app_metrics env_app_m on env.id=env_app_m.env_id and p.app_id = env_app_m.app_id" +
		" where p.app_id=? and p.deleted = FALSE AND env.active = TRUE GROUP by 1,2,3,4, 5"
	impl.Logger.Debugw("other env query:", query)
	_, err := impl.dbConnection.Query(&otherEnvironments, query, appId)
	if err != nil {
		impl.Logger.Error("error in fetching other environment", "error", err)
	}
	return otherEnvironments, nil
}

func (impl AppListingRepositoryImpl) SaveNewDeployment(deploymentStatus *DeploymentStatus, tx *pg.Tx) error {
	err := tx.Insert(deploymentStatus)
	return err
}

func (impl AppListingRepositoryImpl) FindLastDeployedStatus(appName string) (DeploymentStatus, error) {
	var deployment DeploymentStatus
	err := impl.dbConnection.Model(&deployment).
		Where("app_name = ?", appName).
		Order("id Desc").
		Limit(1).
		Select()
	return deployment, err
}

func (impl AppListingRepositoryImpl) FindLastDeployedStatuses(appNames []string) ([]DeploymentStatus, error) {
	if len(appNames) == 0 {
		return []DeploymentStatus{}, nil
	}
	var deploymentStatuses []DeploymentStatus

	var preparedAppNames []string
	for _, a := range appNames {
		preparedAppNames = append(preparedAppNames, "'"+a+"'")
	}

	query := "select distinct app_name, status, max(id) as id from deployment_status where app_name in (" + strings.Join(preparedAppNames[:], ",") + ") group by app_name, status order by id desc"
	_, err := impl.dbConnection.Query(&deploymentStatuses, query)
	if err != nil {
		impl.Logger.Error("err", err)
		return []DeploymentStatus{}, err
	}
	return deploymentStatuses, nil
}

func (impl AppListingRepositoryImpl) FindLastDeployedStatusesForAllApps() ([]DeploymentStatus, error) {
	var deploymentStatuses []DeploymentStatus
	query := "select distinct app_name, status, max(id) as id, app_id, env_id, created_on, updated_on from deployment_status group by app_name, status, app_id, env_id, created_on, updated_on order by id desc"
	_, err := impl.dbConnection.Query(&deploymentStatuses, query)
	if err != nil {
		impl.Logger.Error("err", err)
		return []DeploymentStatus{}, err
	}
	return deploymentStatuses, nil
}

func (impl AppListingRepositoryImpl) DeploymentDetailByArtifactId(ciArtifactId int) (bean.DeploymentDetailContainer, error) {
	impl.Logger.Debug("reached at AppListingRepository:")
	var deploymentDetail bean.DeploymentDetailContainer
	query := "SELECT env.id AS environment_id, env.environment_name, env.default, pco.created_on as last_deployed_time, a.app_name" +
		" FROM pipeline_config_override pco" +
		" INNER JOIN pipeline p on p.id = pco.pipeline_id" +
		" INNER JOIN environment env ON env.id=p.environment_id" +
		" INNER JOIN app a on a.id = p.app_id" +
		" WHERE pco.ci_artifact_id = ? and p.deleted=false AND env.active = TRUE" +
		" ORDER BY pco.pipeline_release_counter desc LIMIT 1;"
	impl.Logger.Debugw("last success full deployed artifact query:", query)

	_, err := impl.dbConnection.Query(&deploymentDetail, query, ciArtifactId)
	if err != nil {
		impl.Logger.Errorw("Exception caught:", err)
		return deploymentDetail, err
	}

	return deploymentDetail, nil
}

func (impl AppListingRepositoryImpl) FindAppCount(isProd bool) (int, error) {
	var count int
	query := "SELECT count(distinct pipeline.app_id) from pipeline pipeline " +
		" INNER JOIN environment env on env.id=pipeline.environment_id" +
		" INNER JOIN app app on app.id=pipeline.app_id" +
		" WHERE env.default = ? and app.active=true;"
	_, err := impl.dbConnection.Query(&count, query, isProd)
	if err != nil {
		impl.Logger.Errorw("exception caught inside repository for fetching app count:", "err", err)
		return count, err
	}

	return count, nil
}
