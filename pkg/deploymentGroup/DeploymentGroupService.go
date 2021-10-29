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

package deploymentGroup

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"go.uber.org/zap"
)

type DeploymentGroupRequest struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	CiPipelineId  int    `json:"ciPipelineId"`
	EnvironmentId int    `json:"environmentId"`
	AppIds        []int  `json:"appIds"`
	UserId        int32  `json:"-"`
}

type CiPipelineResponseForDG struct {
	Name         string             `json:"name"`
	CiPipelineId int                `json:"ciPipelineId"`
	Connections  int                `json:"connections"`
	Repositories []*bean.CiMaterial `json:"repositories"`
}

type EnvironmentAppListForDG struct {
	Id   int                `json:"id"`
	Name string             `json:"name"`
	Apps []pipeline.AppBean `json:"apps"`
}

type DeploymentGroupTriggerRequest struct {
	DeploymentGroupId int   `json:"deploymentGroupId"`
	UserId            int32 `json:"userId"`
	CiArtifactId      int   `json:"ciArtifactId"`
}

type DeploymentGroupHibernateRequest struct {
	DeploymentGroupId int   `json:"deploymentGroupId"`
	UserId            int32 `json:"userId"`
	//CiArtifactId      int   `json:"ciArtifactId"`
}

type DeploymentGroupService interface {
	CreateDeploymentGroup(deploymentGroupRequest *DeploymentGroupRequest) (*DeploymentGroupRequest, error)
	FetchParentCiForDG(deploymentGroupId int) ([]*CiPipelineResponseForDG, error)
	FetchEnvApplicationsForDG(ciPipelineId int) ([]*EnvironmentAppListForDG, error)
	FetchAllDeploymentGroups() ([]DeploymentGroupDTO, error)
	DeleteDeploymentGroup(deploymentGroupId int) (bool, error)
	FindById(id int) (*DeploymentGroupDTO, error)
	TriggerReleaseForDeploymentGroup(triggerRequest *DeploymentGroupTriggerRequest) (interface{}, error)
	UpdateDeploymentGroup(deploymentGroupRequest *DeploymentGroupRequest) (*DeploymentGroupRequest, error)
	GetArtifactsByCiPipeline(ciPipelineId int) (bean.CiArtifactResponse, error)
	GetDeploymentGroupById(deploymentGroupId int) (*DeploymentGroupRequest, error)
}

type DeploymentGroupDTO struct {
	Id              int             `json:"id"`
	Name            string          `json:"name"`
	AppCount        int             `json:"appCount"`
	NoOfApps        string          `json:"noOfApps"`
	EnvironmentId   int             `json:"environmentId"`
	CiPipelineId    int             `json:"ciPipelineId"`
	CiPipelineName  string          `json:"ciPipelineName"`
	CiMaterialDTOs  []CiMaterialDTO `json:"ciMaterialDTOs"`
	EnvironmentName string          `json:"environmentName"`
}

type CiMaterialDTO struct {
	Name        string `json:"name"`
	SourceType  string `json:"type"`
	SourceValue string `json:"value"`
}

type DeploymentGroupServiceImpl struct {
	appRepository                pipelineConfig.AppRepository
	logger                       *zap.SugaredLogger
	pipelineRepository           pipelineConfig.PipelineRepository
	ciPipelineRepository         pipelineConfig.CiPipelineRepository
	deploymentGroupRepository    repository.DeploymentGroupRepository
	environmentRepository        cluster.EnvironmentRepository
	deploymentGroupAppRepository repository.DeploymentGroupAppRepository
	ciArtifactRepository         repository.CiArtifactRepository
	appWorkflowRepository        appWorkflow.AppWorkflowRepository
	workflowDagExecutor          pipeline.WorkflowDagExecutor
}

func NewDeploymentGroupServiceImpl(appRepository pipelineConfig.AppRepository, logger *zap.SugaredLogger,
	pipelineRepository pipelineConfig.PipelineRepository, ciPipelineRepository pipelineConfig.CiPipelineRepository,
	deploymentGroupRepository repository.DeploymentGroupRepository, environmentRepository cluster.EnvironmentRepository,
	deploymentGroupAppRepository repository.DeploymentGroupAppRepository,
	ciArtifactRepository repository.CiArtifactRepository,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	workflowDagExecutor pipeline.WorkflowDagExecutor) *DeploymentGroupServiceImpl {
	return &DeploymentGroupServiceImpl{
		appRepository:                appRepository,
		logger:                       logger,
		pipelineRepository:           pipelineRepository,
		ciPipelineRepository:         ciPipelineRepository,
		deploymentGroupRepository:    deploymentGroupRepository,
		environmentRepository:        environmentRepository,
		deploymentGroupAppRepository: deploymentGroupAppRepository,
		ciArtifactRepository:         ciArtifactRepository,
		appWorkflowRepository:        appWorkflowRepository,
		workflowDagExecutor:          workflowDagExecutor,
	}
}

func (impl *DeploymentGroupServiceImpl) FindById(id int) (*DeploymentGroupDTO, error) {
	impl.logger.Debug("fetching deployment group details")
	dg, err := impl.deploymentGroupRepository.GetById(id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("err", err)
		return nil, err
	}

	environment, err := impl.environmentRepository.FindById(dg.EnvironmentId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("err", err)
		return nil, err
	}

	dgResp := &DeploymentGroupDTO{
		Id:              dg.Id,
		Name:            dg.Name,
		AppCount:        dg.AppCount,
		NoOfApps:        dg.NoOfApps,
		EnvironmentId:   dg.EnvironmentId,
		EnvironmentName: environment.Name,
		CiPipelineId:    dg.CiPipelineId,
	}
	ciPipeline, err := impl.ciPipelineRepository.FindById(dg.CiPipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("err", err)
		return nil, err
	}
	dgResp.CiPipelineName = ciPipeline.Name

	var ciMaterials []CiMaterialDTO
	for _, m := range ciPipeline.CiPipelineMaterials {
		ciMaterialDTO := CiMaterialDTO{
			Name:        m.GitMaterial.Name[strings.Index(m.GitMaterial.Name, "-")+1:],
			SourceType:  string(m.Type),
			SourceValue: m.Value,
		}
		ciMaterials = append(ciMaterials, ciMaterialDTO)
	}

	dgResp.CiMaterialDTOs = ciMaterials
	return dgResp, nil
}

func (impl *DeploymentGroupServiceImpl) CreateDeploymentGroup(deploymentGroupRequest *DeploymentGroupRequest) (*DeploymentGroupRequest, error) {

	//TODO - WIRING
	model := &repository.DeploymentGroup{}
	model.Name = deploymentGroupRequest.Name
	model.EnvironmentId = deploymentGroupRequest.EnvironmentId
	model.CiPipelineId = deploymentGroupRequest.CiPipelineId
	model.Active = true
	model.CreatedBy = deploymentGroupRequest.UserId
	model.UpdatedBy = deploymentGroupRequest.UserId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()
	model.AppCount = len(deploymentGroupRequest.AppIds)
	model, err := impl.deploymentGroupRepository.Create(model)
	if err != nil {
		impl.logger.Errorw("error in creating DG", "error", err)
		return nil, err
	}

	for _, item := range deploymentGroupRequest.AppIds {
		modelMap := &repository.DeploymentGroupApp{}
		modelMap.DeploymentGroupId = model.Id
		modelMap.AppId = item
		modelMap.CreatedBy = deploymentGroupRequest.UserId
		modelMap.UpdatedBy = deploymentGroupRequest.UserId
		modelMap.Active = true
		modelMap.CreatedOn = time.Now()
		modelMap.UpdatedOn = time.Now()
		modelMap, err := impl.deploymentGroupAppRepository.Create(modelMap)
		if err != nil {
			impl.logger.Errorw("error in creating DG map", "error", err)
			return nil, err
		}
	}
	deploymentGroupRequest.Id = model.Id
	return deploymentGroupRequest, nil
}

type DGParentCiResponse struct {
	ParentCis     []*CiPipelineResponseForDG
	ParentCiForDg *DeploymentGroupDTO
}

func (impl *DeploymentGroupServiceImpl) FetchParentCiForDG(deploymentGroupId int) ([]*CiPipelineResponseForDG, error) {

	var results []*CiPipelineResponseForDG
	parentCiIds, err := impl.ciPipelineRepository.FetchParentCiPipelinesForDG()
	if err != nil {
		impl.logger.Errorw("error in fetching parent ci pipelines", "error", err)
	}

	parentIdsMap := make(map[int][]int)
	for _, item := range parentCiIds {
		list := parentIdsMap[item.ParentCiPipeline]
		if len(list) == 0 {
			var ids []int
			ids = append(ids, item.Id)
			parentIdsMap[item.ParentCiPipeline] = ids
		} else {
			list = append(list, item.Id)
			parentIdsMap[item.ParentCiPipeline] = list
		}
	}

	if deploymentGroupId != 0 {
		dg, err := impl.FindById(deploymentGroupId)
		if err != nil && util.IsErrNoRows(err) {
			return nil, err
		}
		found := false
		for _, item := range parentCiIds {
			if item.Id == dg.CiPipelineId {
				found = true
				break
			}
		}
		if !found {
			ids := make([]int, 0)
			parentIdsMap[dg.CiPipelineId] = ids
		}
	}

	for key, value := range parentIdsMap {
		ciPipeline, count, err := impl.ciPipelineRepository.FetchCiPipelinesForDG(key, value)
		if err != nil {
			impl.logger.Errorw("error in creating DG", "error", err)
		}
		pipeline := &CiPipelineResponseForDG{}
		pipeline.CiPipelineId = ciPipeline.Id
		pipeline.Name = ciPipeline.Name
		pipeline.Connections = count

		var materialTemp []*bean.CiMaterial
		for _, material := range ciPipeline.CiPipelineMaterials {
			ciMaterial := bean.CiMaterial{
				Id:              material.Id,
				CheckoutPath:    material.CheckoutPath,
				Path:            material.Path,
				ScmId:           material.ScmId,
				GitMaterialId:   material.GitMaterialId,
				GitMaterialName: material.GitMaterial.Name[strings.Index(material.GitMaterial.Name, "-")+1:],
				ScmName:         material.ScmName,
				ScmVersion:      material.ScmVersion,
				Source:          &bean.SourceTypeConfig{Type: material.Type, Value: material.Value},
			}
			materialTemp = append(materialTemp, &ciMaterial)
		}
		pipeline.Repositories = materialTemp
		results = append(results, pipeline)

	}
	return results, err
}

func (impl *DeploymentGroupServiceImpl) FetchEnvApplicationsForDG(ciPipelineId int) ([]*EnvironmentAppListForDG, error) {
	var results []*EnvironmentAppListForDG
	envs, err := impl.environmentRepository.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err)
		return nil, err
	}
	childrenCi, err := impl.ciPipelineRepository.FindByParentCiPipelineId(ciPipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("err", err)
		return nil, err
	}
	var childrenCiIds []int
	for _, ci := range childrenCi {
		childrenCiIds = append(childrenCiIds, ci.Id)
	}
	childrenCiIds = append(childrenCiIds, ciPipelineId)
	childrenCds, err := impl.pipelineRepository.FindByCiPipelineIdsIn(childrenCiIds)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	appIdMap := make(map[int]bool)
	for _, cd := range childrenCds {
		appIdMap[cd.AppId] = true
	}

	for _, e := range envs {
		environmentAppListForDG := &EnvironmentAppListForDG{}
		environmentAppListForDG.Id = e.Id
		environmentAppListForDG.Name = e.Name

		apps, err := impl.appRepository.FindAppsByEnvironmentId(e.Id)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("error in fetching apps", "error", err)
			return nil, err
		}
		for _, app := range apps {
			if _, ok := appIdMap[app.Id]; ok {
				environmentAppListForDG.Apps = append(environmentAppListForDG.Apps, pipeline.AppBean{Id: app.Id, Name: app.AppName})
			}
		}
		if len(environmentAppListForDG.Apps) > 0 {
			results = append(results, environmentAppListForDG)
		}
	}

	return results, nil
}

func (impl *DeploymentGroupServiceImpl) FetchAllDeploymentGroups() ([]DeploymentGroupDTO, error) {
	impl.logger.Debug("fetching all deployment groups")
	deploymentGroups, err := impl.deploymentGroupRepository.GetAll()
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("err", err)
		return nil, err
	}

	var deploymentGroupsResp []DeploymentGroupDTO
	deploymentGroupsResp = []DeploymentGroupDTO{}
	for _, dg := range deploymentGroups {
		if !dg.Active {
			continue
		}
		ciPipeline, err := impl.ciPipelineRepository.FindById(dg.CiPipelineId)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("err", err)
			return nil, err
		}

		var ciMaterials []CiMaterialDTO
		for _, m := range ciPipeline.CiPipelineMaterials {
			ciMaterialDTO := CiMaterialDTO{
				Name:        m.GitMaterial.Name[strings.Index(m.GitMaterial.Name, "-")+1:],
				SourceType:  string(m.Type),
				SourceValue: m.Value,
			}
			ciMaterials = append(ciMaterials, ciMaterialDTO)
		}

		env, err := impl.environmentRepository.FindById(dg.EnvironmentId)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("err", err)
			return nil, err
		}

		resp := DeploymentGroupDTO{
			Id:              dg.Id,
			Name:            dg.Name,
			AppCount:        dg.AppCount,
			NoOfApps:        dg.NoOfApps,
			EnvironmentId:   dg.EnvironmentId,
			EnvironmentName: env.Name,
			CiPipelineId:    dg.CiPipelineId,
			CiMaterialDTOs:  ciMaterials,
		}
		deploymentGroupsResp = append(deploymentGroupsResp, resp)
	}
	return deploymentGroupsResp, nil
}

func (impl *DeploymentGroupServiceImpl) DeleteDeploymentGroup(deploymentGroupId int) (bool, error) {
	model, err := impl.deploymentGroupRepository.GetById(deploymentGroupId)
	if err != nil {
		impl.logger.Errorw("error in delete DG", "error", err)
		return false, err
	}
	model.Active = false
	model, err = impl.deploymentGroupRepository.Update(model)
	if err != nil {
		impl.logger.Errorw("error in delete DG", "error", err)
		return false, err
	}

	modelApps, err := impl.deploymentGroupAppRepository.GetByDeploymentGroup(deploymentGroupId)
	if err != nil {
		impl.logger.Errorw("error in delete DG App", "error", err)
		return false, err
	}
	for _, modelApp := range modelApps {
		modelApp.Active = false
		_, err = impl.deploymentGroupAppRepository.Update(modelApp)
		if err != nil {
			impl.logger.Errorw("error in delete DG App map", "error", err)
			return false, err
		}
	}

	return true, nil
}

func (impl *DeploymentGroupServiceImpl) TriggerReleaseForDeploymentGroup(triggerRequest *DeploymentGroupTriggerRequest) (interface{}, error) {
	group, err := impl.deploymentGroupRepository.FindByIdWithApp(triggerRequest.DeploymentGroupId)
	if err != nil {
		impl.logger.Errorw("error in fetching deployment group", "err", err)
		return nil, err
	}
	var appIds []int
	for _, groupApp := range group.DeploymentGroupApps {
		appIds = append(appIds, groupApp.AppId)
	}
	if len(appIds) == 0 {
		impl.logger.Errorw("no app found", "req", triggerRequest)
		return nil, fmt.Errorf("no app found corresponding to deployment group %d", triggerRequest.DeploymentGroupId)
	}
	ciPipelines, err := impl.ciPipelineRepository.FinDByParentCiPipelineAndAppId(group.CiPipelineId, appIds)
	if err != nil {
		impl.logger.Errorw("error in fetching ci pipelines", "triggerRequest", triggerRequest, "err", err)
		return nil, err
	}
	impl.logger.Debugw("ci pipelines identified", "pipeline", ciPipelines)
	//get artifact ids
	var ciPipelineIds []int
	for _, ci := range ciPipelines {
		ciPipelineIds = append(ciPipelineIds, ci.Id)
	}
	if len(ciPipelineIds) == 0 {
		impl.logger.Errorw("no ciPipelineIds found", "req", triggerRequest)
		return nil, fmt.Errorf("no ciPipeline found corresponding to deployment group %d", triggerRequest.DeploymentGroupId)
	}
	ciArtifacts, err := impl.ciArtifactRepository.FinDByParentCiArtifactAndCiId(triggerRequest.CiArtifactId, ciPipelineIds)
	if err != nil {
		impl.logger.Errorw("error in getting ci artifacts", "err", err, "parent", triggerRequest.CiArtifactId)
		return nil, err
	}
	//get cd pipeline id
	appwfMappings, err := impl.appWorkflowRepository.FindWFCDMappingByCIPipelineIds(ciPipelineIds)
	if err != nil {
		impl.logger.Errorw("error in getting wf mappings", "err", err, "ciPipelineIds", ciPipelineIds)
		return nil, err
	}
	var cdPipelineIds []int
	for _, wf := range appwfMappings {
		cdPipelineIds = append(cdPipelineIds, wf.ComponentId)
	}
	if len(cdPipelineIds) == 0 {
		impl.logger.Errorw("no cdPipelineIds found", "req", triggerRequest)
		return nil, fmt.Errorf("no cdPipelineIds found corresponding to deployment group %d", triggerRequest.DeploymentGroupId)
	}
	cdPipelines, err := impl.pipelineRepository.FindByIdsInAndEnvironment(cdPipelineIds, group.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in fetching cdPipelines ", "triggerRequest", triggerRequest, "err", err)
		return nil, err
	}
	var requests []*pipeline.BulkTriggerRequest
	ciArtefactMapping := make(map[int]*repository.CiArtifact)
	for _, ciArtefact := range ciArtifacts {
		ciArtefactMapping[ciArtefact.PipelineId] = ciArtefact
	}
	for _, cdPipeline := range cdPipelines {
		if val, ok := ciArtefactMapping[cdPipeline.CiPipelineId]; ok {
			//do something here
			req := &pipeline.BulkTriggerRequest{
				CiArtifactId: val.Id,
				PipelineId:   cdPipeline.Id,
			}
			requests = append(requests, req)
		} else {
			impl.logger.Warnw("no artifact found", "cdPipeline", cdPipeline)
		}
	}
	//trigger
	// apply mapping
	_, err = impl.workflowDagExecutor.TriggerBulkDeploymentAsync(requests, triggerRequest.UserId)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (impl *DeploymentGroupServiceImpl) UpdateDeploymentGroup(deploymentGroupRequest *DeploymentGroupRequest) (*DeploymentGroupRequest, error) {

	model, err := impl.deploymentGroupRepository.GetById(deploymentGroupRequest.Id)
	if err != nil {
		impl.logger.Errorw("error in updating DG", "error", err)
		return nil, err
	}

	model.Name = deploymentGroupRequest.Name
	model.EnvironmentId = deploymentGroupRequest.EnvironmentId
	model.CiPipelineId = deploymentGroupRequest.CiPipelineId
	model.Active = true
	model.UpdatedBy = deploymentGroupRequest.UserId
	model.UpdatedOn = time.Now()
	model.AppCount = len(deploymentGroupRequest.AppIds)
	model, err = impl.deploymentGroupRepository.Update(model)
	if err != nil {
		impl.logger.Errorw("error in updating DG", "error", err)
		return nil, err
	}

	dgMapping, err := impl.deploymentGroupAppRepository.GetByDeploymentGroup(deploymentGroupRequest.Id)
	if err != nil {
		impl.logger.Errorw("error in updating DG map", "error", err)
		return nil, err
	}

	appIds := make(map[int]bool)
	for _, item := range deploymentGroupRequest.AppIds {
		appIds[item] = true
	}

	existingAppIds := make(map[int]bool)
	//var existingAppIds []int
	for _, item := range dgMapping {
		existingAppIds[item.AppId] = true
		if _, ok := appIds[item.AppId]; ok {
			// DO NOTHING
		} else {
			// DELETE ENTRY AS NOT PROVIDED IN REQUEST
			err = impl.deploymentGroupAppRepository.Delete(item)
			if err != nil {
				impl.logger.Errorw("error in delete DG map", "error", err)
				return nil, err
			}
		}
	}

	for _, item := range deploymentGroupRequest.AppIds {
		if _, ok := existingAppIds[item]; ok {
			// DO NOTHING, ALREADY PROCESSED
		} else {
			//CREATE NEW MAP
			modelMap := &repository.DeploymentGroupApp{}
			modelMap.DeploymentGroupId = model.Id
			modelMap.AppId = item
			modelMap.CreatedBy = deploymentGroupRequest.UserId
			modelMap.UpdatedBy = deploymentGroupRequest.UserId
			modelMap.Active = true
			modelMap.CreatedOn = time.Now()
			modelMap.UpdatedOn = time.Now()
			modelMap, err := impl.deploymentGroupAppRepository.Create(modelMap)
			if err != nil {
				impl.logger.Errorw("error in updating DG map", "error", err)
				return nil, err
			}
		}
	}

	return deploymentGroupRequest, nil
}

func (impl *DeploymentGroupServiceImpl) GetArtifactsByCiPipeline(ciPipelineId int) (bean.CiArtifactResponse, error) {
	var ciArtifacts []bean.CiArtifactBean
	var ciArtifactsResponse bean.CiArtifactResponse
	artifacts, err := impl.ciArtifactRepository.GetArtifactsByCiPipelineId(ciPipelineId)
	if err != nil {
		return ciArtifactsResponse, err
	}

	for _, artifact := range artifacts {
		mInfo, err := impl.parseMaterialInfo([]byte(artifact.MaterialInfo), artifact.DataSource)
		if err != nil {
			mInfo = []byte("[]")
			impl.logger.Errorw("error on parse material", "err", err)
		}

		ciArtifacts = append(ciArtifacts, bean.CiArtifactBean{
			Id:           artifact.Id,
			Image:        artifact.Image,
			MaterialInfo: mInfo,
			Latest:       artifact.Latest,
		})
	}

	if ciArtifacts == nil {
		ciArtifacts = []bean.CiArtifactBean{}
	}
	ciArtifactsResponse.CiArtifacts = ciArtifacts
	return ciArtifactsResponse, nil
}

func (impl *DeploymentGroupServiceImpl) parseMaterialInfo(materialInfo json.RawMessage, source string) (json.RawMessage, error) {
	if source != "GOCD" && source != "CI-RUNNER" && source != "EXTERNAL" {
		return nil, fmt.Errorf("datasource: %s not supported", source)
	}
	var ciMaterials []repository.CiMaterialInfo
	err := json.Unmarshal(materialInfo, &ciMaterials)
	if err != nil {
		impl.logger.Errorw("unmarshal error for material info", "material info", materialInfo, "err", err)
	}
	var scmMapList []map[string]string
	scmMap := map[string]string{}
	for _, material := range ciMaterials {
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

			scmMap["url"] = url
			scmMap["revision"] = revision
			scmMap["modifiedTime"] = _modification.ModifiedTime
			scmMap["author"] = _modification.Author
			scmMap["message"] = _modification.Message
			scmMap["tag"] = _modification.Tag
			scmMap["webhookData"] = _webhookDataStr
		}
		scmMapList = append(scmMapList, scmMap)
	}
	mInfo, err := json.Marshal(scmMapList)
	if err != nil {
		impl.logger.Errorw("unmarshal error for material info", "scmMapList", scmMapList, "err", err)
	}
	return mInfo, err
}

func (impl *DeploymentGroupServiceImpl) GetDeploymentGroupById(deploymentGroupId int) (*DeploymentGroupRequest, error) {

	model, err := impl.deploymentGroupRepository.GetById(deploymentGroupId)
	if err != nil {
		impl.logger.Errorw("error in updating DG", "error", err)
		return nil, err
	}

	deploymentGroupRequest := &DeploymentGroupRequest{}
	deploymentGroupRequest.Id = model.Id
	deploymentGroupRequest.Name = model.Name
	deploymentGroupRequest.EnvironmentId = model.EnvironmentId
	deploymentGroupRequest.CiPipelineId = model.CiPipelineId

	var appIds []int
	dgMapping, err := impl.deploymentGroupAppRepository.GetByDeploymentGroup(model.Id)
	if err != nil {
		impl.logger.Errorw("error in updating DG map", "error", err)
		return nil, err
	}

	for _, item := range dgMapping {
		appIds = append(appIds, item.AppId)
	}
	deploymentGroupRequest.AppIds = appIds
	return deploymentGroupRequest, err
}
