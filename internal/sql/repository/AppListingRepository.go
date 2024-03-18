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
	"context"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/internal/middleware"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"go.opentelemetry.io/otel"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type AppListingRepository interface {
	FetchJobs(appIds []int, statuses []string, environmentIds []int, sortOrder string) ([]*bean.JobListingContainer, error)
	FetchOverviewCiPipelines(jobId int) ([]*bean.JobListingContainer, error)
	FetchJobsLastSucceededOn(ciPipelineIDs []int) ([]*bean.CiPipelineLastSucceededTime, error)
	FetchAppDetail(ctx context.Context, appId int, envId int) (bean.AppDetailContainer, error)

	FetchAppTriggerView(appId int) ([]bean.TriggerView, error)
	FetchAppStageStatus(appId int, appType int) ([]bean.AppStageStatus, error)

	// Not in used
	PrometheusApiByEnvId(id int) (*string, error)

	FetchOtherEnvironment(appId int) ([]*bean.Environment, error)
	FetchMinDetailOtherEnvironment(appId int) ([]*bean.Environment, error)
	DeploymentDetailByArtifactId(ciArtifactId int, envId int) (bean.DeploymentDetailContainer, error)
	FindAppCount(isProd bool) (int, error)
	FetchAppsByEnvironmentV2(appListingFilter helper.AppListingFilter) ([]*bean.AppEnvironmentContainer, int, error)
	FetchOverviewAppsByEnvironment(envId, limit, offset int) ([]*bean.AppEnvironmentContainer, error)
	FetchLastDeployedImage(appId, envId int) (*LastDeployed, error)
}

// below table is deprecated, not being used anywhere
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

type AppNameTypeIdContainerDBResponse struct {
	AppName string `sql:"app_name"`
	AppId   int    `sql:"id"`
}

type LastDeployed struct {
	ParentCiArtifact  int    `sql:"parent_ci_artifact_id"`
	CiArtifactId      int    `sql:"ci_artifact_id"`
	LastDeployedBy    string `sql:"last_deployed_by"`
	LastDeployedImage string `sql:"last_deployed_image"`
}

type AppListingRepositoryImpl struct {
	dbConnection                     *pg.DB
	Logger                           *zap.SugaredLogger
	appListingRepositoryQueryBuilder helper.AppListingRepositoryQueryBuilder
	environmentRepository            repository2.EnvironmentRepository
	gitOpsRepository                 GitOpsConfigRepository
}

type CiWorkflowRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewAppListingRepositoryImpl(
	Logger *zap.SugaredLogger,
	dbConnection *pg.DB,
	appListingRepositoryQueryBuilder helper.AppListingRepositoryQueryBuilder,
	environmentRepository repository2.EnvironmentRepository,
	gitOpsRepository GitOpsConfigRepository) *AppListingRepositoryImpl {
	return &AppListingRepositoryImpl{
		dbConnection:                     dbConnection,
		Logger:                           Logger,
		appListingRepositoryQueryBuilder: appListingRepositoryQueryBuilder,
		environmentRepository:            environmentRepository,
		gitOpsRepository:                 gitOpsRepository,
	}
}

func (impl AppListingRepositoryImpl) FetchJobs(appIds []int, statuses []string, environmentIds []int, sortOrder string) ([]*bean.JobListingContainer, error) {
	var jobContainers []*bean.JobListingContainer
	if len(appIds) == 0 {
		return jobContainers, nil
	}
	jobsQuery := impl.appListingRepositoryQueryBuilder.BuildJobListingQuery(appIds, statuses, environmentIds, sortOrder)

	impl.Logger.Debugw("basic app detail query: ", jobsQuery)
	_, appsErr := impl.dbConnection.Query(&jobContainers, jobsQuery)
	if appsErr != nil {
		impl.Logger.Error(appsErr)
		return jobContainers, appsErr
	}
	jobContainers = impl.extractEnvironmentNameFromId(jobContainers)
	return jobContainers, nil
}

func (impl AppListingRepositoryImpl) FetchOverviewCiPipelines(jobId int) ([]*bean.JobListingContainer, error) {
	var jobContainers []*bean.JobListingContainer
	jobsQuery := impl.appListingRepositoryQueryBuilder.OverviewCiPipelineQuery()
	impl.Logger.Debugw("basic app detail query: ", jobsQuery)
	_, appsErr := impl.dbConnection.Query(&jobContainers, jobsQuery, jobId)
	if appsErr != nil {
		impl.Logger.Error(appsErr)
		return jobContainers, appsErr
	}
	jobContainers = impl.extractEnvironmentNameFromId(jobContainers)
	return jobContainers, nil
}

func (impl AppListingRepositoryImpl) FetchOverviewAppsByEnvironment(envId, limit, offset int) ([]*bean.AppEnvironmentContainer, error) {
	query := " SELECT a.id as app_id,a.app_name,aps.status as app_status, ld.last_deployed_time " +
		" FROM app a " +
		" INNER JOIN pipeline p ON p.app_id = a.id and p.deleted = false and p.environment_id = ? " +
		" LEFT JOIN app_status aps ON aps.app_id = a.id and aps.env_id = ? " +
		" LEFT JOIN " +
		" (SELECT pco.pipeline_id,MAX(pco.created_on) as last_deployed_time from pipeline_config_override pco " +
		" GROUP BY pco.pipeline_id) ld ON ld.pipeline_id = p.id " +
		" WHERE a.active = true " +
		" ORDER BY a.app_name "
	if limit > 0 {
		query += fmt.Sprintf("LIMIT %v", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf("OFFSET %v", offset)
	}
	var envContainers []*bean.AppEnvironmentContainer
	_, err := impl.dbConnection.Query(&envContainers, query, envId, envId)
	return envContainers, err
}

func (impl AppListingRepositoryImpl) FetchLastDeployedImage(appId, envId int) (*LastDeployed, error) {
	var lastDeployed []*LastDeployed
	// we are adding a case in the query to concatenate the string "(inactive)" to the users' email id when user is inactive
	query := `select ca.id as ci_artifact_id, ca.parent_ci_artifact as parent_ci_artifact_id, ca.image as last_deployed_image, 
			    case
					when u.active = false then u.email_id || ' (inactive)'
					else u.email_id
				end as last_deployed_by 
				from pipeline p
                join cd_workflow cw on cw.pipeline_id = p.id
			  	join cd_workflow_runner cwr on cwr.cd_workflow_id = cw.id
				join ci_artifact ca on ca.id = cw.ci_artifact_id
				join users u on u.id = cwr.triggered_by
				where p.app_id = ? and p.environment_id = ? and p.deleted = false order by cwr.created_on desc;`
	_, err := impl.dbConnection.Query(&lastDeployed, query, appId, envId)
	if len(lastDeployed) > 0 {
		return lastDeployed[0], err
	}
	return nil, err
}

func (impl AppListingRepositoryImpl) FetchJobsLastSucceededOn(CiPipelineIDs []int) ([]*bean.CiPipelineLastSucceededTime, error) {
	var lastSucceededTimeArray []*bean.CiPipelineLastSucceededTime
	if len(CiPipelineIDs) == 0 {
		return lastSucceededTimeArray, nil
	}
	jobsLastFinishedOnQuery := impl.appListingRepositoryQueryBuilder.JobsLastSucceededOnTimeQuery(CiPipelineIDs)
	impl.Logger.Debugw("basic app detail query: ", jobsLastFinishedOnQuery)
	_, appsErr := impl.dbConnection.Query(&lastSucceededTimeArray, jobsLastFinishedOnQuery)
	if appsErr != nil {
		impl.Logger.Errorw("error in fetching lastSucceededTimeArray", "error", appsErr, jobsLastFinishedOnQuery)
		return lastSucceededTimeArray, appsErr
	}
	return lastSucceededTimeArray, nil
}

func getRequiredAppIdsInSequence(appIds []int) []int {
	resIDs := make([]int, 0)
	appIdsSet := make(map[int]bool)
	for _, appId := range appIds {
		if _, ok := appIdsSet[appId]; !ok {
			resIDs = append(resIDs, appId)
			appIdsSet[appId] = true
		}
	}
	return resIDs
}

func (impl AppListingRepositoryImpl) FetchAppsByEnvironmentV2(appListingFilter helper.AppListingFilter) ([]*bean.AppEnvironmentContainer, int, error) {
	impl.Logger.Debugw("reached at FetchAppsByEnvironment ", "appListingFilter", appListingFilter)
	var appEnvArr []*bean.AppEnvironmentContainer
	appsSize := 0
	lastDeployedTimeMap := make(map[int]string)
	var appEnvContainer []*bean.AppEnvironmentContainer
	var lastDeployedTimeDTO = make([]*bean.AppEnvironmentContainer, 0)

	if string(appListingFilter.SortBy) == helper.LastDeployedSortBy {

		query := impl.appListingRepositoryQueryBuilder.GetAppIdsQueryWithPaginationForLastDeployedSearch(appListingFilter)
		impl.Logger.Debug("GetAppIdsQueryWithPaginationForLastDeployedSearch query ", query)
		start := time.Now()
		_, err := impl.dbConnection.Query(&lastDeployedTimeDTO, query)
		middleware.AppListingDuration.WithLabelValues("getAppIdsQueryWithPaginationForLastDeployedSearch", "devtron").Observe(time.Since(start).Seconds())
		if err != nil || len(lastDeployedTimeDTO) == 0 {
			if err != nil {
				impl.Logger.Errorw("error in getting appIds with appList filter from db", "err", err, "filter", appListingFilter, "query", query)
			}
			return appEnvArr, appsSize, err
		}

		appsSize = lastDeployedTimeDTO[0].TotalCount
		appIdsFound := make([]int, len(lastDeployedTimeDTO))
		for i, obj := range lastDeployedTimeDTO {
			appIdsFound[i] = obj.AppId
		}
		appListingFilter.AppIds = appIdsFound
		appContainerQuery := impl.appListingRepositoryQueryBuilder.GetQueryForAppEnvContainerss(appListingFilter)
		impl.Logger.Debug("GetQueryForAppEnvContainerss query ", query)
		_, err = impl.dbConnection.Query(&appEnvContainer, appContainerQuery)
		if err != nil {
			impl.Logger.Errorw("error in getting appEnvContainers with appList filter from db", "err", err, "filter", appListingFilter, "query", appContainerQuery)
			return appEnvArr, appsSize, err
		}

	} else {

		// to get all the appIds in appEnvs allowed for user and filtered by the appListing filter and sorted by name
		appIdCountDtos := make([]*bean.AppEnvironmentContainer, 0)
		appIdCountQuery := impl.appListingRepositoryQueryBuilder.GetAppIdsQueryWithPaginationForAppNameSearch(appListingFilter)
		impl.Logger.Debug("GetAppIdsQueryWithPaginationForAppNameSearch query ", appIdCountQuery)
		start := time.Now()
		_, appsErr := impl.dbConnection.Query(&appIdCountDtos, appIdCountQuery)
		middleware.AppListingDuration.WithLabelValues("getAppIdsQueryWithPaginationForAppNameSearch", "devtron").Observe(time.Since(start).Seconds())
		if appsErr != nil || len(appIdCountDtos) == 0 {
			if appsErr != nil {
				if appsErr != nil {
					impl.Logger.Errorw("error in getting appIds with appList filter from db", "err", appsErr, "filter", appListingFilter, "query", appIdCountQuery)
				}
			}
			return appEnvContainer, appsSize, appsErr
		}
		appsSize = appIdCountDtos[0].TotalCount
		uniqueAppIds := make([]int, len(appIdCountDtos))
		for i, obj := range appIdCountDtos {
			uniqueAppIds[i] = obj.AppId
		}
		appListingFilter.AppIds = uniqueAppIds
		// set appids required for this page in the filter and get the appEnv containers of these apps
		appListingFilter.AppIds = uniqueAppIds
		appsEnvquery := impl.appListingRepositoryQueryBuilder.GetQueryForAppEnvContainerss(appListingFilter)
		impl.Logger.Debug("GetQueryForAppEnvContainerss query: ", appsEnvquery)
		start = time.Now()
		_, appsErr = impl.dbConnection.Query(&appEnvContainer, appsEnvquery)
		middleware.AppListingDuration.WithLabelValues("buildAppListingQuery", "devtron").Observe(time.Since(start).Seconds())
		if appsErr != nil {
			impl.Logger.Errorw("error in getting appEnvContainers with appList filter from db", "err", appsErr, "filter", appListingFilter, "query", appsEnvquery)
			return appEnvContainer, appsSize, appsErr
		}

	}

	// filter out unique pipelineIds from the above result and get the deployment times for them
	// some items don't have pipelineId if no pipeline is configured for the app in the appEnv container
	pipelineIdsSet := make(map[int]bool)
	pipelineIds := make([]int, 0)
	for _, item := range appEnvContainer {
		pId := item.PipelineId
		if _, ok := pipelineIdsSet[pId]; !ok && pId > 0 {
			pipelineIds = append(pipelineIds, pId)
			pipelineIdsSet[pId] = true
		}
	}

	// if any pipeline found get the latest deployment time
	if len(pipelineIds) > 0 {
		query := impl.appListingRepositoryQueryBuilder.BuildAppListingQueryLastDeploymentTimeV2(pipelineIds)
		impl.Logger.Debugw("basic app detail query: ", query)
		start := time.Now()
		_, err := impl.dbConnection.Query(&lastDeployedTimeDTO, query)
		middleware.AppListingDuration.WithLabelValues("buildAppListingQueryLastDeploymentTime", "devtron").Observe(time.Since(start).Seconds())
		if err != nil {
			impl.Logger.Errorw("error in getting latest deployment time for given pipelines", "err", err, "pipelines", pipelineIds, "query", query)
			return appEnvArr, appsSize, err
		}
	}

	// get the last deployment time for all the items
	for _, item := range lastDeployedTimeDTO {
		if _, ok := lastDeployedTimeMap[item.PipelineId]; ok {
			continue
		}
		lastDeployedTimeMap[item.PipelineId] = item.LastDeployedTime

	}

	// set the time for corresponding appEnv container
	for _, item := range appEnvContainer {
		if lastDeployedTime, ok := lastDeployedTimeMap[item.PipelineId]; ok {
			item.LastDeployedTime = lastDeployedTime
		}
		appEnvArr = append(appEnvArr, item)
	}

	return appEnvArr, appsSize, nil
}

// DeploymentDetailsByAppIdAndEnvId It will return the deployment detail of any cd pipeline which is latest triggered for Environment of any App
func (impl AppListingRepositoryImpl) deploymentDetailsByAppIdAndEnvId(ctx context.Context, appId int, envId int) (bean.DeploymentDetailContainer, error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "DeploymentDetailsByAppIdAndEnvId")
	defer span.End()
	var deploymentDetail bean.DeploymentDetailContainer
	query := "SELECT" +
		" a.app_name," +
		" env.environment_name," +
		" env.namespace," +
		" env.default," +
		" p.deployment_app_type," +
		" p.ci_pipeline_id," +
		" p.deployment_app_delete_request," +
		" cia.data_source," +
		" cia.id as ci_artifact_id," +
		" cia.parent_ci_artifact as parent_artifact_id," +
		" cl.k8s_version," +
		" env.cluster_id," +
		" env.is_virtual_environment," +
		" cl.cluster_name" +
		" FROM pipeline p" +
		" INNER JOIN pipeline_config_override pco on pco.pipeline_id=p.id" +
		" INNER JOIN environment env ON env.id=p.environment_id" +
		" INNER JOIN cluster cl on cl.id=env.cluster_id" +
		" INNER JOIN ci_artifact cia on cia.id = pco.ci_artifact_id" +
		" INNER JOIN app a ON a.id=p.app_id" +
		" WHERE a.app_type = 0 AND a.id=? AND env.id=? AND p.deleted = FALSE AND env.active = TRUE" +
		" ORDER BY pco.created_on DESC LIMIT 1;"
	_, err := impl.dbConnection.Query(&deploymentDetail, query, appId, envId)
	if err != nil {
		impl.Logger.Errorw("Exception caught:", "err", err)
		return deploymentDetail, err
	}
	deploymentDetail.EnvironmentId = envId
	return deploymentDetail, nil
}

func parseMaterialInfo(materialInfo string, source string) (json.RawMessage, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PARSEMATERIALINFO_MATERIAL_RECOVER,  materialInfo: %s,  source: %s, err: %s \n", materialInfo, source, r)
		}
	}()
	if source != GOCD && source != CI_RUNNER && source != WEBHOOK && source != EXT {
		return nil, fmt.Errorf("datasource: %s not supported", source)
	}
	if materialInfo == "" {
		return []byte("[]"), nil
	}
	var ciMaterials []*CiMaterialInfo
	err := json.Unmarshal([]byte(materialInfo), &ciMaterials)
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
			_modification := material.Modifications[0]

			revision := _modification.Revision
			url = strings.TrimSpace(url)

			_webhookDataStr := ""
			_webhookDataByteArr, err := json.Marshal(_modification.WebhookData)
			if err == nil {
				_webhookDataStr = string(_webhookDataByteArr)
			}

			materialMap["url"] = url
			materialMap["revision"] = revision
			materialMap["modifiedTime"] = _modification.ModifiedTime
			materialMap["author"] = _modification.Author
			materialMap["message"] = _modification.Message
			materialMap["branch"] = _modification.Branch
			materialMap["webhookData"] = _webhookDataStr
		}
		scmMaps = append(scmMaps, materialMap)
	}
	mInfo, err := json.Marshal(scmMaps)
	return mInfo, err
}

func (impl AppListingRepositoryImpl) FetchAppDetail(ctx context.Context, appId int, envId int) (bean.AppDetailContainer, error) {
	impl.Logger.Debugf("reached at AppListingRepository:")
	var appDetailContainer bean.AppDetailContainer
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "DeploymentDetailsByAppIdAndEnvId")
	defer span.End()
	// Fetch deployment detail of cd pipeline latest triggered within env of any App.
	deploymentDetail, err := impl.deploymentDetailsByAppIdAndEnvId(newCtx, appId, envId)
	if err != nil {
		impl.Logger.Warn("unable to fetch deployment detail for app")
	}
	appDetailContainer.DeploymentDetailContainer = deploymentDetail
	return appDetailContainer, nil
}

func (impl AppListingRepositoryImpl) PrometheusApiByEnvId(id int) (*string, error) {
	impl.Logger.Debug("reached at PrometheusApiByEnvId:")
	var prometheusEndpoint string
	query := "SELECT env.prometheus_endpoint from environment env" +
		" WHERE env.id = ? AND env.active = TRUE"
	impl.Logger.Debugw("query", query)
	// environments := []string{"QA"}
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
		" WHERE p.app_id=? and p.deleted=false and a.app_type = 0 AND env.active = TRUE;"

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
				" cia.material_info as material_info_json_string, cia.data_source, evt.reason as status" +
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
				mInfo, err := parseMaterialInfo(tView.MaterialInfoJsonString, tView.DataSource)
				if err == nil && len(mInfo) > 0 {
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

func (impl AppListingRepositoryImpl) FetchAppStageStatus(appId int, appType int) ([]bean.AppStageStatus, error) {
	impl.Logger.Debug("reached at AppListingRepository:")
	var appStageStatus []bean.AppStageStatus

	var stages struct {
		AppId           int    `json:"app_id,omitempty"`
		CiTemplateId    int    `json:"ci_template_id,omitempty"`
		CiPipelineId    int    `json:"ci_pipeline_id,omitempty"`
		ChartId         int    `json:"chart_id,omitempty"`
		ChartGitRepoUrl string `json:"chart_git_repo_url,omitempty"`
		PipelineId      int    `json:"pipeline_id,omitempty"`
		YamlStatus      int    `json:"yaml_status,omitempty"`
		YamlReviewed    bool   `json:"yaml_reviewed,omitempty"`
	}

	query := "SELECT " +
		" app.id as app_id, ct.id as ci_template_id, cp.id as ci_pipeline_id, ch.id as chart_id, ch.git_repo_url as chart_git_repo_url," +
		" p.id as pipeline_id, ceco.status as yaml_status, ceco.reviewed as yaml_reviewed " +
		" FROM app app" +
		" LEFT JOIN ci_template ct on ct.app_id=app.id" +
		" LEFT JOIN ci_pipeline cp on cp.app_id=app.id" +
		" LEFT JOIN charts ch on ch.app_id=app.id" +
		" LEFT JOIN pipeline p on p.app_id=app.id" +
		" LEFT JOIN chart_env_config_override ceco on ceco.chart_id=ch.id" +
		" WHERE app.id=? and app.app_type = ? limit 1;"

	impl.Logger.Debugw("last app stages status query:", "query", query)

	_, err := impl.dbConnection.Query(&stages, query, appId, appType)
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
	isCustomGitopsRepoUrl := false
	model, err := impl.gitOpsRepository.GetGitOpsConfigActive()
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error while getting GetGitOpsConfigActive", "err", err)
		return appStageStatus, err
	}
	if model != nil && model.Id > 0 && model.AllowCustomRepository {
		isCustomGitopsRepoUrl = true
	}
	if gitOps.IsGitOpsRepoNotConfigured(stages.ChartGitRepoUrl) && stages.CiPipelineId == 0 {
		stages.ChartGitRepoUrl = ""
	}
	appStageStatus = append(appStageStatus, impl.makeAppStageStatus(0, "APP", stages.AppId, true),
		impl.makeAppStageStatus(1, "MATERIAL", materialExists, true),
		impl.makeAppStageStatus(2, "TEMPLATE", stages.CiTemplateId, true),
		impl.makeAppStageStatus(3, "CI_PIPELINE", stages.CiPipelineId, true),
		impl.makeAppStageStatus(4, "CHART", stages.ChartId, true),
		impl.makeAppStageStatus(5, "GITOPS_CONFIG", len(stages.ChartGitRepoUrl), isCustomGitopsRepoUrl),
		impl.makeAppStageStatus(6, "CD_PIPELINE", stages.PipelineId, true),
		impl.makeAppStageChartEnvConfigStatus(7, "CHART_ENV_CONFIG", stages.YamlStatus == 3 && stages.YamlReviewed),
	)

	return appStageStatus, nil
}

func (impl AppListingRepositoryImpl) makeAppStageChartEnvConfigStatus(stage int, stageName string, status bool) bean.AppStageStatus {
	return bean.AppStageStatus{Stage: stage, StageName: stageName, Status: status, Required: true}
}

func (impl AppListingRepositoryImpl) makeAppStageStatus(stage int, stageName string, id int, isRequired bool) bean.AppStageStatus {
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
		Required: isRequired,
	}
}

func (impl AppListingRepositoryImpl) FetchOtherEnvironment(appId int) ([]*bean.Environment, error) {
	// other environment tab
	var otherEnvironments []*bean.Environment
	//TODO: remove infra metrics from query as it is not being used from here
	query := `select pcwr.pipeline_id, pcwr.last_deployed, pcwr.latest_cd_workflow_runner_id, pcwr.environment_id, pcwr.deployment_app_delete_request,   
       			e.cluster_id, e.environment_name, e.default as prod, e.description, ca.image as last_deployed_image, 
      			u.email_id as last_deployed_by, elam.app_metrics, elam.infra_metrics, ap.status as app_status, ca.id as ci_artifact_id, ca.parent_ci_artifact as parent_ci_artifact_id
    			from (select * 
      				from (select p.id as pipeline_id, p.app_id, cwr.started_on as last_deployed, cwr.triggered_by, cwr.id as latest_cd_workflow_runner_id,  
                  	 	cw.ci_artifact_id, p.environment_id, p.deployment_app_delete_request, 
                  		row_number() over (partition by p.id order by cwr.started_on desc) as max_started_on_rank  
            			from (select * from pipeline where app_id = ? and deleted=?) as p 
                     	left join cd_workflow cw on cw.pipeline_id = p.id 
                     	left join cd_workflow_runner cwr on cwr.cd_workflow_id = cw.id 
            			where cwr.workflow_type = ? or cwr.workflow_type is null) pcwrraw  
      				where max_started_on_rank = 1) pcwr 
         		INNER JOIN environment e on e.id = pcwr.environment_id 
         		LEFT JOIN ci_artifact ca on ca.id = pcwr.ci_artifact_id 
         		LEFT JOIN users u on u.id = pcwr.triggered_by 
        		LEFT JOIN env_level_app_metrics elam on pcwr.environment_id = elam.env_id and pcwr.app_id = elam.app_id 
        		LEFT JOIN app_status ap ON pcwr.environment_id = ap.env_id and pcwr.app_id=ap.app_id;`
	_, err := impl.dbConnection.Query(&otherEnvironments, query, appId, false, "DEPLOY")
	if err != nil {
		impl.Logger.Error("error in fetching other environment", "error", err)
		return nil, err
	}
	return otherEnvironments, nil
}

func (impl AppListingRepositoryImpl) FetchMinDetailOtherEnvironment(appId int) ([]*bean.Environment, error) {
	impl.Logger.Debug("reached at FetchMinDetailOtherEnvironment:")
	var otherEnvironments []*bean.Environment
	//TODO: remove infra metrics from query as it is not being used from here
	query := `SELECT p.environment_id,env.environment_name,env.description,env.is_virtual_environment, env.cluster_id, env.default as prod, p.deployment_app_delete_request,
       			env_app_m.app_metrics,env_app_m.infra_metrics from 
 				(SELECT pl.id,pl.app_id,pl.environment_id,pl.deleted, pl.deployment_app_delete_request from pipeline pl 
  					LEFT JOIN pipeline_config_override pco on pco.pipeline_id = pl.id where pl.app_id = ? and pl.deleted = FALSE 
  					GROUP BY pl.id) p INNER JOIN environment env on env.id=p.environment_id 
                	LEFT JOIN env_level_app_metrics env_app_m on env.id=env_app_m.env_id and p.app_id = env_app_m.app_id 
                    where p.app_id=? and p.deleted = FALSE AND env.active = TRUE;`
	_, err := impl.dbConnection.Query(&otherEnvironments, query, appId, appId)
	if err != nil {
		impl.Logger.Error("error in fetching other environment", "error", err)
	}
	return otherEnvironments, nil
}

func (impl AppListingRepositoryImpl) DeploymentDetailByArtifactId(ciArtifactId int, envId int) (bean.DeploymentDetailContainer, error) {
	impl.Logger.Debug("reached at AppListingRepository:")
	var deploymentDetail bean.DeploymentDetailContainer
	query := "SELECT env.id AS environment_id, env.environment_name, env.default, pco.created_on as last_deployed_time, a.app_name" +
		" FROM pipeline_config_override pco" +
		" INNER JOIN pipeline p on p.id = pco.pipeline_id" +
		" INNER JOIN environment env ON env.id=p.environment_id" +
		" INNER JOIN app a on a.id = p.app_id" +
		" WHERE pco.ci_artifact_id = ? and p.deleted=false AND env.active = TRUE AND env.id = ?" +
		" ORDER BY pco.pipeline_release_counter desc LIMIT 1;"
	impl.Logger.Debugw("last success full deployed artifact query:", query)

	_, err := impl.dbConnection.Query(&deploymentDetail, query, ciArtifactId, envId)
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

func (impl AppListingRepositoryImpl) extractEnvironmentNameFromId(jobContainers []*bean.JobListingContainer) []*bean.JobListingContainer {
	var envIds []*int
	for _, job := range jobContainers {
		if job.EnvironmentId != 0 {
			envIds = append(envIds, &job.EnvironmentId)
		}
		if job.LastTriggeredEnvironmentId != 0 {
			envIds = append(envIds, &job.LastTriggeredEnvironmentId)
		}
	}
	envs, _ := impl.environmentRepository.FindByIds(envIds)

	envIdNameMap := make(map[int]string)

	for _, env := range envs {
		envIdNameMap[env.Id] = env.Name
	}

	for _, job := range jobContainers {
		if job.EnvironmentId != 0 {
			job.EnvironmentName = envIdNameMap[job.EnvironmentId]
		}
		if job.LastTriggeredEnvironmentId != 0 {
			job.LastTriggeredEnvironmentName = envIdNameMap[job.LastTriggeredEnvironmentId]
		}
	}

	return jobContainers
}
