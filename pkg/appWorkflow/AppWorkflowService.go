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

package appWorkflow

import (
	"errors"
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/app/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	bean3 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	resourceGroup2 "github.com/devtron-labs/devtron/pkg/resourceGroup"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/variables/utils"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

const (
	CI_PIPELINE_TYPE = "CI_PIPELINE"
	CD_PIPELINE_TYPE = "CD_PIPELINE"
)

type AppWorkflowService interface {
	CreateAppWorkflow(req AppWorkflowDto) (AppWorkflowDto, error)
	FindAppWorkflows(appId int) ([]AppWorkflowDto, error)
	FindAppWorkflowById(Id int, appId int) (AppWorkflowDto, error)
	DeleteAppWorkflow(appWorkflowId int, userId int32) error

	SaveAppWorkflowMapping(wf AppWorkflowMappingDto) (AppWorkflowMappingDto, error)
	FindAppWorkflowMapping(workflowId int) ([]AppWorkflowMappingDto, error)
	FindAllAppWorkflowMapping(workflowIds []int) (map[int][]AppWorkflowMappingDto, error)
	FindAppWorkflowMappingByComponent(id int, compType string) ([]*appWorkflow.AppWorkflowMapping, error)
	CheckCdPipelineByCiPipelineId(id int) bool
	FindAppWorkflowByName(name string, appId int) (AppWorkflowDto, error)

	FindAllWorkflowsComponentDetails(appId int) (*AllAppWorkflowComponentDetails, error)
	FindAppWorkflowsByEnvironmentId(request resourceGroup2.ResourceGroupingRequest, token string) ([]*AppWorkflowDto, error)
	FindAllWorkflowsForApps(request WorkflowNamesRequest) (*WorkflowNamesResponse, error)
	FilterWorkflows(triggerViewConfig *TriggerViewWorkflowConfig, envIds []int) (*TriggerViewWorkflowConfig, error)
	FindCdPipelinesByAppId(appId int) (*bean.CdPipelines, error)
	FindAppWorkflowByCiPipelineId(ciPipelineId int) ([]*appWorkflow.AppWorkflowMapping, error)
}

type AppWorkflowServiceImpl struct {
	Logger                   *zap.SugaredLogger
	appWorkflowRepository    appWorkflow.AppWorkflowRepository
	ciCdPipelineOrchestrator pipeline.CiCdPipelineOrchestrator
	ciPipelineRepository     pipelineConfig.CiPipelineRepository
	pipelineRepository       pipelineConfig.PipelineRepository
	resourceGroupService     resourceGroup2.ResourceGroupService
	appRepository            appRepository.AppRepository
	enforcerUtil             rbac.EnforcerUtil
	userAuthService          user.UserAuthService
	chartService             chart.ChartService
}

type AppWorkflowDto struct {
	Id                    int                     `json:"id,omitempty"`
	Name                  string                  `json:"name"`
	AppId                 int                     `json:"appId"`
	AppWorkflowMappingDto []AppWorkflowMappingDto `json:"tree,omitempty"`
	UserId                int32                   `json:"-"`
}

type TriggerViewWorkflowConfig struct {
	Workflows        []AppWorkflowDto          `json:"workflows"`
	CiConfig         *bean.TriggerViewCiConfig `json:"ciConfig"`
	CdPipelines      *bean.CdPipelines         `json:"cdConfig"`
	ExternalCiConfig []*bean.ExternalCiConfig  `json:"externalCiConfig"`
}

type AppWorkflowMappingDto struct {
	Id                         int        `json:"id,omitempty"`
	AppWorkflowId              int        `json:"appWorkflowId"`
	Type                       string     `json:"type"`
	ComponentId                int        `json:"componentId"`
	ParentId                   int        `json:"parentId"`
	ParentType                 string     `json:"parentType"`
	DeploymentAppDeleteRequest bool       `json:"deploymentAppDeleteRequest"`
	UserId                     int32      `json:"-"`
	IsLast                     bool       `json:"isLast"`
	ChildPipelinesIds          mapset.Set `json:"-"`
}

func (dto AppWorkflowMappingDto) getPipelineIdentifier() PipelineIdentifier {
	return PipelineIdentifier{
		PipelineType: dto.Type,
		PipelineId:   dto.ComponentId,
	}
}

func (dto AppWorkflowMappingDto) getParentPipelineIdentifier() PipelineIdentifier {
	return PipelineIdentifier{
		PipelineType: dto.ParentType,
		PipelineId:   dto.ParentId,
	}
}

type AllAppWorkflowComponentDetails struct {
	Workflows []*WorkflowComponentNamesDto `json:"workflows"`
}

type WorkflowComponentNamesDto struct {
	Id             int      `json:"id"`
	Name           string   `json:"name"`
	CiPipelineId   int      `json:"ciPipelineId"`
	CiPipelineName string   `json:"ciPipelineName"`
	CdPipelines    []string `json:"cdPipelines"`
}

type WorkflowNamesResponse struct {
	AppIdWorkflowNamesMapping map[string][]string `json:"appIdWorkflowNamesMapping"`
}

type WorkflowNamesRequest struct {
	AppNames []string `json:"appNames"`
}

type WorkflowCloneRequest struct {
	WorkflowName  string `json:"workflowName,omitempty"`
	AppId         int    `json:"appId,omitempty"`
	EnvironmentId int    `json:"environmentId,omitempty"`
	WorkflowId    int    `json:"workflowId,omitempty"`
	UserId        int32  `json:"-"`
}

type PipelineIdentifier struct {
	PipelineType string
	PipelineId   int
}

func NewAppWorkflowServiceImpl(logger *zap.SugaredLogger, appWorkflowRepository appWorkflow.AppWorkflowRepository,
	ciCdPipelineOrchestrator pipeline.CiCdPipelineOrchestrator, ciPipelineRepository pipelineConfig.CiPipelineRepository,
	pipelineRepository pipelineConfig.PipelineRepository, enforcerUtil rbac.EnforcerUtil, resourceGroupService resourceGroup2.ResourceGroupService,
	appRepository appRepository.AppRepository, userAuthService user.UserAuthService, chartService chart.ChartService) *AppWorkflowServiceImpl {
	return &AppWorkflowServiceImpl{
		Logger:                   logger,
		appWorkflowRepository:    appWorkflowRepository,
		ciCdPipelineOrchestrator: ciCdPipelineOrchestrator,
		ciPipelineRepository:     ciPipelineRepository,
		pipelineRepository:       pipelineRepository,
		enforcerUtil:             enforcerUtil,
		resourceGroupService:     resourceGroupService,
		appRepository:            appRepository,
		userAuthService:          userAuthService,
		chartService:             chartService,
	}
}

func (impl AppWorkflowServiceImpl) CreateAppWorkflow(req AppWorkflowDto) (AppWorkflowDto, error) {
	var wf *appWorkflow.AppWorkflow
	var savedAppWf *appWorkflow.AppWorkflow
	var err error

	if req.Id != 0 {
		wf = &appWorkflow.AppWorkflow{
			Id:     req.Id,
			Name:   req.Name,
			Active: true,
			AuditLog: sql.AuditLog{
				UpdatedOn: time.Now(),
				UpdatedBy: req.UserId,
			},
		}
		savedAppWf, err = impl.appWorkflowRepository.UpdateAppWorkflow(wf)
	} else {
		workflow, err := impl.appWorkflowRepository.FindByNameAndAppId(req.Name, req.AppId)
		if err != nil && err != pg.ErrNoRows {
			impl.Logger.Errorw("error in finding workflow by app id and name", "name", req.Name, "appId", req.AppId)
			return req, err
		}
		if workflow.Id != 0 {
			impl.Logger.Errorw("workflow with this name already exist", "err", err)
			return req, errors.New(bean2.WORKFLOW_EXIST_ERROR)
		}
		wf := &appWorkflow.AppWorkflow{
			Name:   req.Name,
			AppId:  req.AppId,
			Active: true,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				UpdatedOn: time.Now(),
				CreatedBy: req.UserId,
				UpdatedBy: req.UserId,
			},
		}
		savedAppWf, err = impl.appWorkflowRepository.SaveAppWorkflow(wf)
	}
	if err != nil {
		impl.Logger.Errorw("err", err)
		return req, err
	}
	req.Id = savedAppWf.Id
	return req, nil
}

func (impl AppWorkflowServiceImpl) FindAppWorkflows(appId int) ([]AppWorkflowDto, error) {
	appWorkflows, err := impl.appWorkflowRepository.FindByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error occurred while fetching app workflows", "appId", appId, "err", err)
		return nil, err
	}
	var workflows []AppWorkflowDto
	var wfIds []int
	for _, appWf := range appWorkflows {
		wfIds = append(wfIds, appWf.Id)
	}

	wfrIdVsMappings, err := impl.FindAllAppWorkflowMapping(wfIds)
	if err != nil {
		return nil, err
	}

	for _, w := range appWorkflows {
		workflow := AppWorkflowDto{
			Id:    w.Id,
			Name:  w.Name,
			AppId: w.AppId,
		}
		workflow.AppWorkflowMappingDto = wfrIdVsMappings[w.Id]
		workflows = append(workflows, workflow)
	}

	return workflows, err
}

func (impl AppWorkflowServiceImpl) FindAppWorkflowById(Id int, appId int) (AppWorkflowDto, error) {
	appWorkflow, err := impl.appWorkflowRepository.FindByIdAndAppId(Id, appId)
	if err != nil {
		impl.Logger.Errorw("err", "error", err)
		return AppWorkflowDto{}, err
	}
	wfrIdVsMappings, err := impl.FindAllAppWorkflowMapping([]int{appWorkflow.Id})
	if err != nil {
		return AppWorkflowDto{}, err
	}

	appWorkflowDto := &AppWorkflowDto{
		AppId:                 appWorkflow.AppId,
		Id:                    appWorkflow.Id,
		Name:                  appWorkflow.Name,
		AppWorkflowMappingDto: wfrIdVsMappings[appWorkflow.Id],
	}
	return *appWorkflowDto, err
}

func (impl AppWorkflowServiceImpl) DeleteAppWorkflow(appWorkflowId int, userId int32) error {
	impl.Logger.Debugw("Deleting app-workflow: ", "appWorkflowId", appWorkflowId)
	wf, err := impl.appWorkflowRepository.FindById(appWorkflowId)
	if err != nil {
		impl.Logger.Errorw("err", err)
		return err
	}
	app, err := impl.appRepository.FindById(wf.AppId)
	if err != nil {
		impl.Logger.Errorw("error in finding app by app id", "err", err, "appId", wf.AppId)
		return err
	}

	mappingForCI, err := impl.appWorkflowRepository.FindWFAllMappingByWorkflowId(wf.Id)
	if err != nil {
		impl.Logger.Errorw("err", err)
		return err
	}
	if len(mappingForCI) > 0 {
		return &util.ApiError{
			InternalMessage:   "Workflow contains pipelines. First delete all pipelines in the workflow.",
			UserDetailMessage: fmt.Sprintf("Workflow contains pipelines. First delete all pipelines in the workflow."),
			UserMessage:       fmt.Sprintf("Workflow contains pipelines. First delete all pipelines in the workflow.")}
	}

	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	// Deleting workflow
	err = impl.appWorkflowRepository.DeleteAppWorkflow(wf, tx)
	if err != nil {
		impl.Logger.Errorw("err", err)
		return err
	}
	// Delete app workflow mapping
	mapping, err := impl.appWorkflowRepository.FindWFAllMappingByWorkflowId(wf.Id)
	for _, item := range mapping {
		err := impl.appWorkflowRepository.DeleteAppWorkflowMapping(item, tx)
		if err != nil {
			impl.Logger.Errorw("error in deleting workflow mapping", "err", err)
			return err
		}
	}
	err = impl.userAuthService.DeleteRoles(bean3.WorkflowType, app.AppName, tx, "", wf.Name)
	if err != nil {
		impl.Logger.Errorw("error in deleting auth roles", "err", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (impl AppWorkflowServiceImpl) SaveAppWorkflowMapping(req AppWorkflowMappingDto) (AppWorkflowMappingDto, error) {
	appWorkflow := &appWorkflow.AppWorkflowMapping{
		ParentId:      req.ParentId,
		AppWorkflowId: req.AppWorkflowId,
		ComponentId:   req.ComponentId,
		ParentType:    req.ParentType,
		Type:          req.Type,
		Active:        true,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			UpdatedOn: time.Now(),
			CreatedBy: req.UserId,
			UpdatedBy: req.UserId,
		},
	}
	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return AppWorkflowMappingDto{}, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	appWorkflow, err = impl.appWorkflowRepository.SaveAppWorkflowMapping(appWorkflow, tx)
	if err != nil {
		impl.Logger.Errorw("err", err)
		return AppWorkflowMappingDto{}, err
	}
	req.Id = appWorkflow.Id

	err = tx.Commit()
	if err != nil {
		return AppWorkflowMappingDto{}, err
	}

	return AppWorkflowMappingDto{}, nil
}

func (impl AppWorkflowServiceImpl) FindAllAppWorkflowMapping(workflowIds []int) (map[int][]AppWorkflowMappingDto, error) {
	appWorkflowMappings, err := impl.appWorkflowRepository.FindByWorkflowIds(workflowIds)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error occurred while fetching app wf mapping", "workflowIds", workflowIds, "err", err)
		return nil, err
	}
	parentPipelineIdsSet := mapset.NewSet()
	for _, w := range appWorkflowMappings {
		if w.ParentType == CD_PIPELINE_TYPE {
			parentPipelineIdsSet.Add(w.ParentId)
		}
	}
	var workflowMappingDtos []AppWorkflowMappingDto
	var cdPipelineIds []int
	for _, w := range appWorkflowMappings {
		workflow := AppWorkflowMappingDto{
			Id:            w.Id,
			ParentId:      w.ParentId,
			ComponentId:   w.ComponentId,
			Type:          w.Type,
			AppWorkflowId: w.AppWorkflowId,
			ParentType:    w.ParentType,
		}
		if w.Type == CD_PIPELINE_TYPE {
			if !parentPipelineIdsSet.Contains(w.ComponentId) {
				workflow.IsLast = true
			}
			cdPipelineIds = append(cdPipelineIds, w.ComponentId)
		}
		workflowMappingDtos = append(workflowMappingDtos, workflow)
	}
	if len(cdPipelineIds) > 0 {
		var cdPipelineIdMap map[int]bool
		cdPipelineIdMap, err = impl.pipelineRepository.FilterDeploymentDeleteRequestedPipelineIds(cdPipelineIds)
		if err != nil && err != pg.ErrNoRows {
			impl.Logger.Errorw("error occurred while filtering out delete request pipelines", "cdPipelineIds", cdPipelineIds, "err", err)
			return nil, err
		}
		for _, workflowMapping := range workflowMappingDtos {
			if workflowMapping.Type == "CD_PIPELINE" && cdPipelineIdMap[workflowMapping.ComponentId] {
				workflowMapping.DeploymentAppDeleteRequest = true
			}
		}
	}
	wfIdVsMappings := make(map[int][]AppWorkflowMappingDto)
	for _, workflowMappingDto := range workflowMappingDtos {
		appWorkflowId := workflowMappingDto.AppWorkflowId
		workflowMappings := wfIdVsMappings[appWorkflowId]
		workflowMappings = append(workflowMappings, workflowMappingDto)
		wfIdVsMappings[appWorkflowId] = workflowMappings
	}
	return wfIdVsMappings, err

}

func (impl AppWorkflowServiceImpl) FindAppWorkflowMapping(workflowId int) ([]AppWorkflowMappingDto, error) {
	appWorkflowMapping, err := impl.appWorkflowRepository.FindByWorkflowId(workflowId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("err", err)
		return nil, err
	}
	var workflows []AppWorkflowMappingDto
	for _, w := range appWorkflowMapping {
		workflow := AppWorkflowMappingDto{
			Id:            w.Id,
			ParentId:      w.ParentId,
			ComponentId:   w.ComponentId,
			Type:          w.Type,
			AppWorkflowId: w.AppWorkflowId,
			ParentType:    w.ParentType,
		}
		if w.Type == "CD_PIPELINE" {
			pipeline, err := impl.pipelineRepository.FindById(w.ComponentId)
			if err != nil && err != pg.ErrNoRows {
				impl.Logger.Errorw("err", "err", err)
				return nil, err
			}
			if pipeline != nil {
				workflow.DeploymentAppDeleteRequest = pipeline.DeploymentAppDeleteRequest
			}
		}
		workflows = append(workflows, workflow)
	}
	return workflows, err
}

func (impl AppWorkflowServiceImpl) FindAppWorkflowMappingForEnv(appIds []int) (map[int]*AppWorkflowDto, error) {
	appWorkflowMappings, err := impl.appWorkflowRepository.FindMappingByAppIds(appIds)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("err", err)
		return nil, err
	}
	pipelineIds := make([]int, 0)
	for _, w := range appWorkflowMappings {
		if w.Type == "CD_PIPELINE" {
			pipelineIds = append(pipelineIds, w.ComponentId)
		}
	}
	pipelines, err := impl.pipelineRepository.FindByIdsIn(pipelineIds)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("err", "err", err)
		return nil, err
	}
	pipelineMap := make(map[int]*pipelineConfig.Pipeline)
	for _, pipeline := range pipelines {
		pipelineMap[pipeline.Id] = pipeline
	}
	workflowMappings := make(map[int][]AppWorkflowMappingDto)
	workflows := make(map[int]*AppWorkflowDto)
	for _, w := range appWorkflowMappings {
		if _, ok := workflows[w.AppWorkflowId]; !ok {
			workflows[w.AppWorkflowId] = &AppWorkflowDto{
				Id:    w.AppWorkflowId,
				AppId: w.AppWorkflow.AppId,
			}
		}
		workflow := AppWorkflowMappingDto{
			Id:            w.Id,
			ParentId:      w.ParentId,
			ComponentId:   w.ComponentId,
			Type:          w.Type,
			AppWorkflowId: w.AppWorkflowId,
			ParentType:    w.ParentType,
		}
		if w.Type == "CD_PIPELINE" {
			workflow.DeploymentAppDeleteRequest = pipelineMap[w.ComponentId].DeploymentAppDeleteRequest
		}
		workflowMappings[w.AppWorkflowId] = append(workflowMappings[w.AppWorkflowId], workflow)
		workflows[w.AppWorkflowId].AppWorkflowMappingDto = workflowMappings[w.AppWorkflowId]
	}
	return workflows, err
}

func (impl AppWorkflowServiceImpl) FindAppWorkflowMappingByComponent(id int, compType string) ([]*appWorkflow.AppWorkflowMapping, error) {
	appWorkflowMappings, err := impl.appWorkflowRepository.FindByComponent(id, compType)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("err", err)
		return nil, err
	}
	return appWorkflowMappings, err
}

func (impl AppWorkflowServiceImpl) FindAppWorkflowByName(name string, appId int) (AppWorkflowDto, error) {
	appWorkflow, err := impl.appWorkflowRepository.FindByNameAndAppId(name, appId)
	if err != nil {
		impl.Logger.Errorw("err", err)
		return AppWorkflowDto{}, err
	}
	wfrIdVsMappings, err := impl.FindAllAppWorkflowMapping([]int{appWorkflow.Id})
	if err != nil {
		return AppWorkflowDto{}, err
	}

	appWorkflowDto := &AppWorkflowDto{
		AppId:                 appWorkflow.AppId,
		Id:                    appWorkflow.Id,
		Name:                  appWorkflow.Name,
		AppWorkflowMappingDto: wfrIdVsMappings[appWorkflow.Id],
	}
	return *appWorkflowDto, err
}

func (impl AppWorkflowServiceImpl) CheckCdPipelineByCiPipelineId(id int) bool {
	appWorkflowMapping, err := impl.appWorkflowRepository.FindWFCDMappingByCIPipelineId(id)

	if err == nil && appWorkflowMapping != nil {
		return true
	}
	return false
}

func (impl AppWorkflowServiceImpl) FindAllWorkflowsComponentDetails(appId int) (*AllAppWorkflowComponentDetails, error) {
	//get all workflows
	appWorkflows, err := impl.appWorkflowRepository.FindByAppId(appId)
	if err != nil {
		impl.Logger.Errorw("error in getting app workflows by appId", "err", err, "appId", appId)
		return nil, err
	}
	appWorkflowMappings, err := impl.appWorkflowRepository.FindAllWFMappingsByAppId(appId)
	if err != nil {
		impl.Logger.Errorw("error in getting appWorkflowMappings by appId", "err", err, "appId", appId)
		return nil, err
	}
	var wfComponentDetails []*WorkflowComponentNamesDto
	wfIdAndComponentDtoIndexMap := make(map[int]int)
	for i, appWf := range appWorkflows {
		wfIdAndComponentDtoIndexMap[appWf.Id] = i
		wfComponentDetail := &WorkflowComponentNamesDto{
			Id:   appWf.Id,
			Name: appWf.Name,
		}
		wfComponentDetails = append(wfComponentDetails, wfComponentDetail)
	}

	//getting all ciPipelines by appId
	ciPipelines, err := impl.ciPipelineRepository.FindByAppId(appId)
	if err != nil {
		impl.Logger.Errorw("error in getting ciPipelines by appId", "err", err, "appId", appId)
		return nil, err
	}
	ciPipelineIdNameMap := make(map[int]string, len(ciPipelines))
	for _, ciPipeline := range ciPipelines {
		ciPipelineIdNameMap[ciPipeline.Id] = ciPipeline.Name
	}

	//getting all ciPipelines by appId
	cdPipelines, err := impl.pipelineRepository.FindActiveByAppId(appId)
	if err != nil {
		impl.Logger.Errorw("error in getting cdPipelines by appId", "err", err, "appId", appId)
		return nil, err
	}
	cdPipelineIdNameMap := make(map[int]string, len(cdPipelines))
	for _, cdPipeline := range cdPipelines {
		cdPipelineIdNameMap[cdPipeline.Id] = cdPipeline.Environment.Name
	}

	for _, appWfMapping := range appWorkflowMappings {
		if index, ok := wfIdAndComponentDtoIndexMap[appWfMapping.AppWorkflowId]; ok {
			if appWfMapping.Type == CI_PIPELINE_TYPE {
				wfComponentDetails[index].CiPipelineId = appWfMapping.ComponentId
				if name, ok1 := ciPipelineIdNameMap[appWfMapping.ComponentId]; ok1 {
					wfComponentDetails[index].CiPipelineName = name
				}
			} else if appWfMapping.Type == CD_PIPELINE_TYPE {
				if envName, ok1 := cdPipelineIdNameMap[appWfMapping.ComponentId]; ok1 {
					wfComponentDetails[index].CdPipelines = append(wfComponentDetails[index].CdPipelines, envName)
				}
			}
		}
	}
	resp := &AllAppWorkflowComponentDetails{
		Workflows: wfComponentDetails,
	}
	return resp, nil
}

func (impl AppWorkflowServiceImpl) FindAppWorkflowsByEnvironmentId(request resourceGroup2.ResourceGroupingRequest, token string) ([]*AppWorkflowDto, error) {
	workflows := make([]*AppWorkflowDto, 0)
	if request.ResourceGroupId > 0 {
		appIds, err := impl.resourceGroupService.GetResourceIdsByResourceGroupId(request.ResourceGroupId)
		if err != nil {
			return nil, err
		}
		//override AppIds if already provided app group id in request.
		request.ResourceIds = appIds
	}
	var pipelines []*pipelineConfig.Pipeline
	var err error
	if len(request.ResourceIds) > 0 {
		pipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.ParentResourceId, request.ResourceIds)
	} else {
		pipelines, err = impl.pipelineRepository.FindActivePipelineByEnvId(request.ParentResourceId)
	}
	if err != nil {
		impl.Logger.Errorw("error in fetching pipelines", "envId", request.ParentResourceId, "err", err)
		return nil, err
	}

	pipelineMap := make(map[int]bool)
	appNamesMap := make(map[int]string)
	var appIds []int
	//authorization block starts here
	pipelineIds := make([]int, 0)
	for _, pipeline := range pipelines {
		pipelineIds = append(pipelineIds, pipeline.Id)
	}
	if len(pipelineIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no matching pipeline found"}
		return workflows, err
	}
	var appObjectArr []string
	var envObjectArr []string
	objects := impl.enforcerUtil.GetAppAndEnvObjectByPipelineIds(pipelineIds)
	pipelineIds = []int{}
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
		envObjectArr = append(envObjectArr, object[1])
	}
	appResults, envResults := request.CheckAuthBatch(token, appObjectArr, envObjectArr)
	for _, pipeline := range pipelines {
		appObject := objects[pipeline.Id][0]
		envObject := objects[pipeline.Id][1]
		if !(appResults[appObject] && envResults[envObject]) {
			//if user unauthorized, skip items
			continue
		}
		appIds = append(appIds, pipeline.AppId)
		appNamesMap[pipeline.AppId] = pipeline.App.AppName
		pipelineMap[pipeline.Id] = true
	}
	//authorization block ends here

	if len(appIds) == 0 {
		impl.Logger.Warnw("there is no app id found for fetching app workflows", "req", request)
		return workflows, nil
	}
	appWorkflows, err := impl.FindAppWorkflowMappingForEnv(appIds)
	if err != nil {
		impl.Logger.Errorw("error fetching app workflow mapping by wf id", "err", err)
		return nil, err
	}
	for _, appWorkflow := range appWorkflows {
		appName := appNamesMap[appWorkflow.AppId]
		appWorkflow.Name = appName
		mappings := appWorkflow.AppWorkflowMappingDto
		valid := false
		for _, mapping := range mappings {
			if mapping.Type == CD_PIPELINE_TYPE {
				if _, ok := pipelineMap[mapping.ComponentId]; ok {
					valid = true
				}
			}
		}
		//if there is no matching pipeline for requested environment, skip from workflow listing
		if valid {
			workflows = append(workflows, appWorkflow)
		}
	}
	return workflows, err
}

func (impl AppWorkflowServiceImpl) FindAllWorkflowsForApps(request WorkflowNamesRequest) (*WorkflowNamesResponse, error) {
	if len(request.AppNames) == 0 {
		return &WorkflowNamesResponse{}, nil
	}
	appIdNameMapping, appIds, err := impl.appRepository.FetchAppIdsByDisplayNamesForJobs(request.AppNames)
	if err != nil {
		impl.Logger.Errorw("error in getting apps by appNames", "err", err, "appNames", request.AppNames)
		return nil, err
	}
	appWorkflows, err := impl.appWorkflowRepository.FindByAppIds(appIds)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error occurred while fetching app workflows", "AppIds", appIds, "err", err)
		return nil, err
	}
	appIdWorkflowMap := make(map[string][]string)
	for _, workflow := range appWorkflows {
		if workflows, ok := appIdWorkflowMap[appIdNameMapping[workflow.AppId]]; ok {
			workflows = append(workflows, workflow.Name)
			appIdWorkflowMap[appIdNameMapping[workflow.AppId]] = workflows
		} else {
			appIdWorkflowMap[appIdNameMapping[workflow.AppId]] = []string{workflow.Name}

		}
	}
	workflowResp := &WorkflowNamesResponse{
		AppIdWorkflowNamesMapping: appIdWorkflowMap,
	}
	return workflowResp, err
}

func (impl AppWorkflowServiceImpl) FilterWorkflows(triggerViewConfig *TriggerViewWorkflowConfig, envIds []int) (*TriggerViewWorkflowConfig, error) {
	cdPipelines := triggerViewConfig.CdPipelines.Pipelines
	cdPipelineIdsFiltered := mapset.NewSet()
	//cdPipelinesIds list corresponding to env ids
	for _, cdPipeline := range cdPipelines {
		if slices.Contains(envIds, cdPipeline.EnvironmentId) {
			cdPipelineIdsFiltered.Add(cdPipeline.Id)
		}
	}

	filteredWorkflows := make([]AppWorkflowDto, 0)
	for index, workflow := range triggerViewConfig.Workflows {
		isPresent := false
		for _, appWorkflowMapping := range workflow.AppWorkflowMappingDto {
			if appWorkflowMapping.Type == CD_PIPELINE_TYPE && cdPipelineIdsFiltered.Contains(appWorkflowMapping.ComponentId) {
				isPresent = true
				break
			}
		}
		//filter out all those env which not exist in cdPipelineIdsFiltered
		if !isPresent {
			continue
		}

		identifierToFilteredWorkflowMapping, leafPipelines, _ := processWorkflowMappingTree(workflow.AppWorkflowMappingDto)

		identifierToFilteredWorkflowMapping = filterMappingOnFilteredCdPipelineIds(identifierToFilteredWorkflowMapping, leafPipelines, cdPipelineIdsFiltered)

		triggerViewConfig.Workflows[index].AppWorkflowMappingDto = extractOutFilteredWorkflowMappings(workflow.AppWorkflowMappingDto, identifierToFilteredWorkflowMapping)

		filteredWorkflows = append(filteredWorkflows, triggerViewConfig.Workflows[index])
	}
	triggerViewConfig.Workflows = filteredWorkflows

	return triggerViewConfig, nil
}

// extractOutFilteredWorkflowMappings extracts out those AppWorkflowMappingDto from identifierToFilteredWorkflowMapping
// which have already been filtered out by the env filtering.
func extractOutFilteredWorkflowMappings(appWorkflowMappings []AppWorkflowMappingDto, identifierToFilteredWorkflowMapping map[PipelineIdentifier]*AppWorkflowMappingDto) []AppWorkflowMappingDto {
	newAppWorkflowMappingDto := make([]AppWorkflowMappingDto, 0)
	for _, appWorkflowMapping := range appWorkflowMappings {
		if _, ok := identifierToFilteredWorkflowMapping[appWorkflowMapping.getPipelineIdentifier()]; ok {
			newAppWorkflowMappingDto = append(newAppWorkflowMappingDto, *identifierToFilteredWorkflowMapping[appWorkflowMapping.getPipelineIdentifier()])
		}
	}
	return newAppWorkflowMappingDto
}

// processWorkflowMappingTree function processed the wf mapping array into a tree structure
// returns a map of identifier to mapping, leaf nodes and the root node
func processWorkflowMappingTree(appWorkflowMappings []AppWorkflowMappingDto) (map[PipelineIdentifier]*AppWorkflowMappingDto, []AppWorkflowMappingDto, *AppWorkflowMappingDto) {
	identifierToFilteredWorkflowMapping := make(map[PipelineIdentifier]*AppWorkflowMappingDto)
	leafPipelines := make([]AppWorkflowMappingDto, 0)
	var rootPipeline *AppWorkflowMappingDto
	//initializing the nodes with empty children and collecting leaf
	for i, appWorkflowMapping := range appWorkflowMappings {
		appWorkflowMappings[i].ChildPipelinesIds = mapset.NewSet()
		identifierToFilteredWorkflowMapping[appWorkflowMapping.getPipelineIdentifier()] = &appWorkflowMappings[i]

		//collecting leaf pipelines
		if appWorkflowMapping.IsLast {
			leafPipelines = append(leafPipelines, appWorkflowMapping)
		}
	}

	for _, appWorkflowMapping := range identifierToFilteredWorkflowMapping {
		// populating children in parent nodes
		parentId := appWorkflowMapping.getParentPipelineIdentifier()
		componentId := appWorkflowMapping.ComponentId
		if parentMapping, hasParent := identifierToFilteredWorkflowMapping[parentId]; hasParent && !parentMapping.ChildPipelinesIds.Contains(componentId) {
			parentMapping.ChildPipelinesIds.Add(componentId)
		} else if !hasParent {
			rootPipeline = appWorkflowMapping
		}
	}
	return identifierToFilteredWorkflowMapping, leafPipelines, rootPipeline
}

// filterMappingOnFilteredCdPipelineIds iterates over all leaf cd-pipelines, if that leaf cd-pipeline is present in the
// cdPipelineIdsFiltered then we want to preserve all it's parents cd, but if at a
// stage where one leaf cd-pipeline is not in cdPipelineIdsFiltered then we can delete the trailing leaf
// cd-pipeline from componentIdWorkflowMapping's list of AppWorkflowMappingDto and also truncate the child
// cd-pipeline id present in the parent's ChildPipelinesIds object inside AppWorkflowMappingDto.
func filterMappingOnFilteredCdPipelineIds(identifierToFilteredWorkflowMapping map[PipelineIdentifier]*AppWorkflowMappingDto,
	leafPipelines []AppWorkflowMappingDto, cdPipelineIdsFiltered mapset.Set) map[PipelineIdentifier]*AppWorkflowMappingDto {

	leafPipelineSize := len(leafPipelines)
	for i := 0; i < leafPipelineSize; i++ {
		if cdPipelineIdsFiltered.Contains(leafPipelines[i].ComponentId) {
			continue
		} else {
			delete(identifierToFilteredWorkflowMapping, leafPipelines[i].getPipelineIdentifier())
			parent := leafPipelines[i].getParentPipelineIdentifier()
			identifierToFilteredWorkflowMapping[parent].ChildPipelinesIds.Remove(leafPipelines[i].ComponentId)
		}
		parentPipelineIdentifier := leafPipelines[i].getParentPipelineIdentifier()
		childPipelineIds := identifierToFilteredWorkflowMapping[parentPipelineIdentifier].ChildPipelinesIds
		if childPipelineIds.Cardinality() == 0 {
			//this means this pipeline has become leaf, so append this pipelineId in leafPipelines for further processing
			leafPipelines = append(leafPipelines, *identifierToFilteredWorkflowMapping[leafPipelines[i].getParentPipelineIdentifier()])
			leafPipelineSize += 1
		}

	}
	return identifierToFilteredWorkflowMapping
}

func (impl AppWorkflowServiceImpl) FindCdPipelinesByAppId(appId int) (*bean.CdPipelines, error) {
	dbPipelines, err := impl.pipelineRepository.FindActiveByAppId(appId)
	if err != nil {
		impl.Logger.Errorw("FindCdPipelinesByAppId, error in fetching cdPipeline", "appId", appId, "err", err)
		return nil, err
	}
	cdPipelines := &bean.CdPipelines{
		AppId: appId,
	}

	isAppLevelGitOpsConfigured, err := impl.chartService.IsGitOpsRepoConfiguredForDevtronApps(appId)
	if err != nil {
		impl.Logger.Errorw("error in fetching latest chart details for app by appId")
		return nil, err
	}

	for _, pipeline := range dbPipelines {
		cdPipelineConfigObj := &bean.CDPipelineConfigObject{
			Id:                        pipeline.Id,
			EnvironmentId:             pipeline.EnvironmentId,
			EnvironmentName:           pipeline.Environment.Name,
			CiPipelineId:              pipeline.CiPipelineId,
			TriggerType:               pipeline.TriggerType,
			Name:                      pipeline.Name,
			DeploymentAppType:         pipeline.DeploymentAppType,
			AppName:                   pipeline.DeploymentAppName,
			AppId:                     pipeline.AppId,
			IsGitOpsRepoNotConfigured: !isAppLevelGitOpsConfigured,
		}
		cdPipelines.Pipelines = append(cdPipelines.Pipelines, cdPipelineConfigObj)
	}

	return cdPipelines, nil
}

func (impl AppWorkflowServiceImpl) FindAppWorkflowByCiPipelineId(ciPipelineId int) ([]*appWorkflow.AppWorkflowMapping, error) {
	appWorkflowMapping, err := impl.appWorkflowRepository.FindByComponentId(ciPipelineId)
	if err != nil {
		impl.Logger.Errorw("error in getting app workflow mappings from component id", "err", err, "componentId", ciPipelineId)
		return nil, err
	}
	return appWorkflowMapping, nil

}

// LevelWiseSort performs level wise sort for workflow mappings starting from leaves
// This will break if ever the workflow mappings array break the assumption of being a DAG with one root node
func LevelWiseSort(appWorkflowMappings []AppWorkflowMappingDto) []AppWorkflowMappingDto {

	if len(appWorkflowMappings) < 2 {
		return appWorkflowMappings
	}

	identifierToNodeMapping, _, root := processWorkflowMappingTree(appWorkflowMappings)

	result := make([]AppWorkflowMappingDto, 0)
	nodesInCurrentLevel := append(make([]AppWorkflowMappingDto, 0), *root)
	for len(result) != len(appWorkflowMappings) {
		result = append(result, nodesInCurrentLevel...)
		childrenOfCurrentLevel := make([]AppWorkflowMappingDto, 0)
		for _, node := range nodesInCurrentLevel {
			childrenOfCurrentLevel = append(childrenOfCurrentLevel, getMappingsFromIds(identifierToNodeMapping, utils.ToIntArray(node.ChildPipelinesIds.ToSlice()))...)
		}
		// cloning slice elements
		nodesInCurrentLevel = append(childrenOfCurrentLevel, []AppWorkflowMappingDto{}...)
	}

	return result
}

func getMappingsFromIds(identifierToNodeMapping map[PipelineIdentifier]*AppWorkflowMappingDto, ids []int) []AppWorkflowMappingDto {
	result := make([]AppWorkflowMappingDto, 0)
	for _, id := range ids {
		identifier := PipelineIdentifier{
			PipelineType: CD_PIPELINE_TYPE,
			PipelineId:   id,
		}
		result = append(result, *identifierToNodeMapping[identifier])
	}
	return result
}
