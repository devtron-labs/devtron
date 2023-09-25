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

package pipeline

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/client/gitSensor"
	app2 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/genericNotes"
	repository3 "github.com/devtron-labs/devtron/pkg/genericNotes/repository"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	history3 "github.com/devtron-labs/devtron/pkg/pipeline/history"
	repository4 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	repository5 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	bean3 "github.com/devtron-labs/devtron/pkg/user/bean"
	util2 "github.com/devtron-labs/devtron/util"
	util3 "github.com/devtron-labs/devtron/util/k8s"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/go-pg/pg"
	errors1 "github.com/juju/errors"
	"go.uber.org/zap"
)

type CiCdPipelineOrchestrator interface {
	CreateApp(createRequest *bean.CreateAppDTO) (*bean.CreateAppDTO, error)
	DeleteApp(appId int, userId int32) error
	CreateMaterials(createMaterialRequest *bean.CreateMaterialDTO) (*bean.CreateMaterialDTO, error)
	UpdateMaterial(updateMaterialRequest *bean.UpdateMaterialDTO) (*bean.UpdateMaterialDTO, error)
	CreateCiConf(createRequest *bean.CiConfigRequest, templateId int) (*bean.CiConfigRequest, error)
	CreateCDPipelines(pipelineRequest *bean.CDPipelineConfigObject, appId int, userId int32, tx *pg.Tx, appName string) (pipelineId int, err error)
	UpdateCDPipeline(pipelineRequest *bean.CDPipelineConfigObject, userId int32, tx *pg.Tx) (err error)
	DeleteCiPipeline(pipeline *pipelineConfig.CiPipeline, request *bean.CiPatchRequest, tx *pg.Tx) error
	DeleteCdPipeline(pipelineId int, userId int32, tx *pg.Tx) error
	PatchMaterialValue(createRequest *bean.CiPipeline, userId int32, oldPipeline *pipelineConfig.CiPipeline) (*bean.CiPipeline, error)
	PatchCiMaterialSource(ciPipeline *bean.CiMaterialPatchRequest, userId int32) (*bean.CiMaterialPatchRequest, error)
	PipelineExists(name string) (bool, error)
	GetCdPipelinesForApp(appId int) (cdPipelines *bean.CdPipelines, err error)
	GetCdPipelinesForAppAndEnv(appId int, envId int) (cdPipelines *bean.CdPipelines, err error)
	GetByEnvOverrideId(envOverrideId int) (*bean.CdPipelines, error)
	BuildCiPipelineScript(userId int32, ciScript *bean.CiScript, scriptStage string, ciPipeline *bean.CiPipeline) *pipelineConfig.CiPipelineScript
	AddPipelineMaterialInGitSensor(pipelineMaterials []*pipelineConfig.CiPipelineMaterial) error
	CheckStringMatchRegex(regex string, value string) bool
	CreateEcrRepo(dockerRepository, AWSRegion, AWSAccessKeyId, AWSSecretAccessKey string) error
	GetCdPipelinesForEnv(envId int, requestedAppIds []int) (cdPipelines *bean.CdPipelines, err error)
}

type CiCdPipelineOrchestratorImpl struct {
	appRepository                 app2.AppRepository
	logger                        *zap.SugaredLogger
	materialRepository            pipelineConfig.MaterialRepository
	pipelineRepository            pipelineConfig.PipelineRepository
	ciPipelineRepository          pipelineConfig.CiPipelineRepository
	ciPipelineMaterialRepository  pipelineConfig.CiPipelineMaterialRepository
	GitSensorClient               gitSensor.Client
	ciConfig                      *CiConfig
	appWorkflowRepository         appWorkflow.AppWorkflowRepository
	envRepository                 repository2.EnvironmentRepository
	attributesService             attributes.AttributesService
	appListingRepository          repository.AppListingRepository
	appLabelsService              app.AppCrudOperationService
	userAuthService               user.UserAuthService
	prePostCdScriptHistoryService history3.PrePostCdScriptHistoryService
	prePostCiScriptHistoryService history3.PrePostCiScriptHistoryService
	pipelineStageService          PipelineStageService
	//ciTemplateOverrideRepository  pipelineConfig.CiTemplateOverrideRepository
	ciTemplateService             CiTemplateService
	ciTemplateOverrideRepository  pipelineConfig.CiTemplateOverrideRepository
	gitMaterialHistoryService     history3.GitMaterialHistoryService
	ciPipelineHistoryService      history3.CiPipelineHistoryService
	dockerArtifactStoreRepository dockerRegistryRepository.DockerArtifactStoreRepository
	configMapService              ConfigMapService
	genericNoteService            genericNotes.GenericNoteService
}

func NewCiCdPipelineOrchestrator(
	pipelineGroupRepository app2.AppRepository,
	logger *zap.SugaredLogger,
	materialRepository pipelineConfig.MaterialRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	GitSensorClient gitSensor.Client, ciConfig *CiConfig,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	envRepository repository2.EnvironmentRepository,
	attributesService attributes.AttributesService,
	appListingRepository repository.AppListingRepository,
	appLabelsService app.AppCrudOperationService,
	userAuthService user.UserAuthService,
	prePostCdScriptHistoryService history3.PrePostCdScriptHistoryService,
	prePostCiScriptHistoryService history3.PrePostCiScriptHistoryService,
	pipelineStageService PipelineStageService,
	ciTemplateOverrideRepository pipelineConfig.CiTemplateOverrideRepository,
	gitMaterialHistoryService history3.GitMaterialHistoryService,
	ciPipelineHistoryService history3.CiPipelineHistoryService,
	ciTemplateService CiTemplateService,
	dockerArtifactStoreRepository dockerRegistryRepository.DockerArtifactStoreRepository,
	configMapService ConfigMapService,
	genericNoteService genericNotes.GenericNoteService) *CiCdPipelineOrchestratorImpl {
	return &CiCdPipelineOrchestratorImpl{
		appRepository:                 pipelineGroupRepository,
		logger:                        logger,
		materialRepository:            materialRepository,
		pipelineRepository:            pipelineRepository,
		ciPipelineRepository:          ciPipelineRepository,
		ciPipelineMaterialRepository:  ciPipelineMaterialRepository,
		GitSensorClient:               GitSensorClient,
		ciConfig:                      ciConfig,
		appWorkflowRepository:         appWorkflowRepository,
		envRepository:                 envRepository,
		attributesService:             attributesService,
		appListingRepository:          appListingRepository,
		appLabelsService:              appLabelsService,
		userAuthService:               userAuthService,
		prePostCdScriptHistoryService: prePostCdScriptHistoryService,
		prePostCiScriptHistoryService: prePostCiScriptHistoryService,
		pipelineStageService:          pipelineStageService,
		ciTemplateOverrideRepository:  ciTemplateOverrideRepository,
		gitMaterialHistoryService:     gitMaterialHistoryService,
		ciPipelineHistoryService:      ciPipelineHistoryService,
		ciTemplateService:             ciTemplateService,
		dockerArtifactStoreRepository: dockerArtifactStoreRepository,
		configMapService:              configMapService,
		genericNoteService:            genericNoteService,
	}
}

const BEFORE_DOCKER_BUILD string = "BEFORE_DOCKER_BUILD"
const AFTER_DOCKER_BUILD string = "AFTER_DOCKER_BUILD"

func (impl CiCdPipelineOrchestratorImpl) PatchCiMaterialSource(patchRequest *bean.CiMaterialPatchRequest, userId int32) (*bean.CiMaterialPatchRequest, error) {
	pipeline, err := impl.findUniquePipelineForAppIdAndEnvironmentId(patchRequest.AppId, patchRequest.EnvironmentId)
	if err != nil {
		return nil, err
	}
	ciPipelineMaterial, err := impl.findUniqueCiPipelineMaterial(pipeline.CiPipelineId)
	if err != nil {
		return nil, err
	}
	ciPipelineMaterial.Type = patchRequest.Source.Type
	ciPipelineMaterial.Value = patchRequest.Source.Value
	ciPipelineMaterial.Regex = patchRequest.Source.Regex
	ciPipelineMaterial.AuditLog.UpdatedBy = userId
	ciPipelineMaterial.AuditLog.UpdatedOn = time.Now()
	if err = impl.saveUpdatedMaterial([]*pipelineConfig.CiPipelineMaterial{ciPipelineMaterial}); err != nil {
		return nil, err
	}
	return patchRequest, nil
}

func (impl CiCdPipelineOrchestratorImpl) findUniquePipelineForAppIdAndEnvironmentId(appId, environmentId int) (*pipelineConfig.Pipeline, error) {
	pipeline, err := impl.pipelineRepository.FindActiveByAppIdAndEnvironmentId(appId, environmentId)
	if err != nil {
		return nil, err
	}
	if len(pipeline) != 1 {
		return nil, fmt.Errorf("unique pipeline was not found, for the given appId and environmentId")
	}
	return pipeline[0], nil
}

func (impl CiCdPipelineOrchestratorImpl) findUniqueCiPipelineMaterial(ciPipelineId int) (*pipelineConfig.CiPipelineMaterial, error) {
	ciPipelineMaterials, err := impl.ciPipelineMaterialRepository.FindByCiPipelineIdsIn([]int{ciPipelineId})
	if err != nil {
		return nil, err
	}
	if len(ciPipelineMaterials) != 1 {
		return nil, fmt.Errorf("unique ciPipelineMaterial was not found, for the given appId and environmentId")
	}
	return ciPipelineMaterials[0], nil
}

func (impl CiCdPipelineOrchestratorImpl) saveUpdatedMaterial(materialsUpdate []*pipelineConfig.CiPipelineMaterial) error {
	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = impl.ciPipelineMaterialRepository.Update(tx, materialsUpdate...); err != nil {
		return err
	}
	if err = impl.AddPipelineMaterialInGitSensor(materialsUpdate); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
func (impl CiCdPipelineOrchestratorImpl) PatchMaterialValue(createRequest *bean.CiPipeline, userId int32, oldPipeline *pipelineConfig.CiPipeline) (*bean.CiPipeline, error) {
	argByte, err := json.Marshal(createRequest.DockerArgs)
	if err != nil {
		impl.logger.Error(err)
		return nil, err
	}
	dbConnection := impl.pipelineRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	ciPipelineObject := &pipelineConfig.CiPipeline{
		Version:                  createRequest.Version,
		Id:                       createRequest.Id,
		DockerArgs:               string(argByte),
		Active:                   createRequest.Active,
		IsManual:                 createRequest.IsManual,
		IsExternal:               createRequest.IsExternal,
		Deleted:                  createRequest.Deleted,
		ParentCiPipeline:         createRequest.ParentCiPipeline,
		ScanEnabled:              createRequest.ScanEnabled,
		IsDockerConfigOverridden: createRequest.IsDockerConfigOverridden,
		AuditLog:                 sql.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now()},
	}

	createOnTimeMap := make(map[int]time.Time)
	createByMap := make(map[int]int32)
	for _, oldMaterial := range oldPipeline.CiPipelineMaterials {
		createOnTimeMap[oldMaterial.GitMaterialId] = oldMaterial.CreatedOn
		createByMap[oldMaterial.GitMaterialId] = oldMaterial.CreatedBy
	}

	err = impl.ciPipelineRepository.Update(ciPipelineObject, tx)
	if err != nil {
		return nil, err
	}
	var CiEnvMappingObject *pipelineConfig.CiEnvMapping
	if ciPipelineObject.Id != 0 {
		CiEnvMappingObject, err = impl.ciPipelineRepository.FindCiEnvMappingByCiPipelineId(ciPipelineObject.Id)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting CiEnvMappingObject ", "err", err, "ciPipelineId", createRequest.Id)
			return nil, err
		}
	}
	if err == nil && CiEnvMappingObject != nil {
		if CiEnvMappingObject.EnvironmentId != createRequest.EnvironmentId {
			CiEnvMappingObject.EnvironmentId = createRequest.EnvironmentId
			CiEnvMappingObject.AuditLog = sql.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now()}
			err = impl.ciPipelineRepository.UpdateCiEnvMapping(CiEnvMappingObject, tx)
			if err != nil {
				impl.logger.Errorw("error in getting CiEnvMappingObject ", "err", err, "ciPipelineId", createRequest.Id)
				return nil, err
			}
			if createRequest.EnvironmentId != 0 {
				createJobEnvOverrideRequest := &bean2.CreateJobEnvOverridePayload{
					AppId:  createRequest.AppId,
					EnvId:  createRequest.EnvironmentId,
					UserId: userId,
				}
				_, err = impl.configMapService.ConfigSecretEnvironmentCreate(createJobEnvOverrideRequest)
				if err != nil && !strings.Contains(err.Error(), "already exists") {
					impl.logger.Errorw("error in saving env override", "createJobEnvOverrideRequest", createJobEnvOverrideRequest, "err", err)
					return nil, err
				}
			}
		}
	}
	// marking old scripts inactive
	err = impl.ciPipelineRepository.MarkCiPipelineScriptsInactiveByCiPipelineId(createRequest.Id, tx)
	if err != nil {
		impl.logger.Errorw("error in marking ciPipelineScripts inactive", "err", err, "ciPipelineId", createRequest.Id)
		return nil, err
	}
	if createRequest.PreBuildStage != nil {
		//updating pre stage
		err = impl.pipelineStageService.UpdatePipelineStage(createRequest.PreBuildStage, repository5.PIPELINE_STAGE_TYPE_PRE_CI, createRequest.Id, userId)
		if err != nil {
			impl.logger.Errorw("error in updating pre stage", "err", err, "preBuildStage", createRequest.PreBuildStage, "ciPipelineId", createRequest.Id)
			return nil, err
		}
	}
	if createRequest.PostBuildStage != nil {
		//updating post stage
		err = impl.pipelineStageService.UpdatePipelineStage(createRequest.PostBuildStage, repository5.PIPELINE_STAGE_TYPE_POST_CI, createRequest.Id, userId)
		if err != nil {
			impl.logger.Errorw("error in updating post stage", "err", err, "postBuildStage", createRequest.PostBuildStage, "ciPipelineId", createRequest.Id)
			return nil, err
		}
	}
	for _, material := range createRequest.CiMaterial {
		if material.IsRegex == true && material.Source.Value != "" {
			material.IsRegex = false
		} else if material.IsRegex == false && material.Source.Regex != "" {
			material.IsRegex = true
		}
	}
	var materials []*pipelineConfig.CiPipelineMaterial
	var materialsAdd []*pipelineConfig.CiPipelineMaterial
	var materialsUpdate []*pipelineConfig.CiPipelineMaterial
	var materialGitMap = make(map[int]string)
	for _, material := range createRequest.CiMaterial {
		pipelineMaterial := &pipelineConfig.CiPipelineMaterial{
			Id:            material.Id,
			Value:         material.Source.Value,
			Type:          material.Source.Type,
			Active:        createRequest.Active,
			Regex:         material.Source.Regex,
			GitMaterialId: material.GitMaterialId,
			AuditLog:      sql.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now()},
		}
		if material.Source.Type == pipelineConfig.SOURCE_TYPE_BRANCH_FIXED {
			materialGitMap[material.GitMaterialId] = material.Source.Value
		}
		if material.Id == 0 {
			pipelineMaterial.CiPipelineId = createRequest.Id
			pipelineMaterial.CreatedBy = userId
			pipelineMaterial.CreatedOn = time.Now()
			materialsAdd = append(materialsAdd, pipelineMaterial)
		} else {
			pipelineMaterial.CiPipelineId = createRequest.Id
			pipelineMaterial.CreatedBy = userId
			pipelineMaterial.CreatedOn = createOnTimeMap[material.GitMaterialId]
			pipelineMaterial.UpdatedOn = time.Now()
			pipelineMaterial.UpdatedBy = userId
			materialsUpdate = append(materialsUpdate, pipelineMaterial)
		}
	}
	if len(materialsAdd) > 0 {
		err = impl.ciPipelineMaterialRepository.Save(tx, materialsAdd...)
		if err != nil {
			return nil, err
		}
	}
	if len(materialsUpdate) > 0 {
		err = impl.ciPipelineMaterialRepository.Update(tx, materialsUpdate...)
		if err != nil {
			return nil, err
		}
	}

	materials = append(materials, materialsAdd...)
	materials = append(materials, materialsUpdate...)

	if ciPipelineObject.IsExternal {

	} else {
		err = impl.AddPipelineMaterialInGitSensor(materials)
		if err != nil {
			impl.logger.Errorf("error in saving pipelineMaterials in git sensor", "materials", materials, "err", err)
			return nil, err
		}
	}

	childrenCiPipelines, err := impl.ciPipelineRepository.FindByParentCiPipelineId(createRequest.Id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("err", "err", err)
		return nil, err
	}
	var childrenCiPipelineIds []int
	for _, ci := range childrenCiPipelines {
		childrenCiPipelineIds = append(childrenCiPipelineIds, ci.Id)
		ciPipelineObject := &pipelineConfig.CiPipeline{
			Version:          createRequest.Version,
			Id:               ci.Id,
			DockerArgs:       string(argByte),
			Active:           createRequest.Active,
			IsManual:         createRequest.IsManual,
			IsExternal:       true,
			Deleted:          createRequest.Deleted,
			ParentCiPipeline: createRequest.Id,
			AuditLog:         sql.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now()},
		}
		err = impl.ciPipelineRepository.Update(ciPipelineObject, tx)
		if err != nil {
			impl.logger.Errorw("err", "err", err)
			return nil, err
		}
	}
	if !createRequest.IsExternal && createRequest.IsDockerConfigOverridden {
		//get override
		savedTemplateOverrideBean, err := impl.ciTemplateService.FindTemplateOverrideByCiPipelineId(createRequest.Id)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting templateOverride by ciPipelineId", "err", err, "ciPipelineId", createRequest.Id)
			return nil, err
		}
		ciBuildConfigBean := createRequest.DockerConfigOverride.CiBuildConfig
		templateOverrideReq := &pipelineConfig.CiTemplateOverride{
			CiPipelineId:     createRequest.Id,
			DockerRegistryId: createRequest.DockerConfigOverride.DockerRegistry,
			DockerRepository: createRequest.DockerConfigOverride.DockerRepository,
			//DockerfilePath:   createRequest.DockerConfigOverride.DockerBuildConfig.DockerfilePath,
			GitMaterialId:             ciBuildConfigBean.GitMaterialId,
			BuildContextGitMaterialId: ciBuildConfigBean.BuildContextGitMaterialId,
			Active:                    true,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: userId,
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		}

		savedTemplateOverride := savedTemplateOverrideBean.CiTemplateOverride
		err = impl.createDockerRepoIfNeeded(createRequest.DockerConfigOverride.DockerRegistry, createRequest.DockerConfigOverride.DockerRepository)
		if err != nil {
			impl.logger.Errorw("error, createDockerRepoIfNeeded", "err", err, "dockerRegistryId", createRequest.DockerConfigOverride.DockerRegistry, "dockerRegistry", createRequest.DockerConfigOverride.DockerRepository)
			return nil, err
		}
		if savedTemplateOverride != nil && savedTemplateOverride.Id > 0 {
			ciBuildConfigBean.Id = savedTemplateOverride.CiBuildConfigId
			templateOverrideReq.Id = savedTemplateOverride.Id
			templateOverrideReq.CreatedOn = savedTemplateOverride.CreatedOn
			templateOverrideReq.CreatedBy = savedTemplateOverride.CreatedBy
			ciTemplateBean := &bean2.CiTemplateBean{
				CiTemplateOverride: templateOverrideReq,
				CiBuildConfig:      ciBuildConfigBean,
				UserId:             userId,
			}
			err = impl.ciTemplateService.Update(ciTemplateBean)
			if err != nil {
				return nil, err
			}

			err = impl.ciPipelineHistoryService.SaveHistory(ciPipelineObject, materials, ciTemplateBean, repository4.TRIGGER_UPDATE)

			if err != nil {
				impl.logger.Errorw("error in saving history of ci pipeline material")
			}

		} else {
			ciTemplateBean := &bean2.CiTemplateBean{
				CiTemplateOverride: templateOverrideReq,
				CiBuildConfig:      ciBuildConfigBean,
				UserId:             userId,
			}
			err := impl.ciTemplateService.Save(ciTemplateBean)
			if err != nil {
				return nil, err
			}

			err = impl.ciPipelineHistoryService.SaveHistory(ciPipelineObject, materials, ciTemplateBean, repository4.TRIGGER_UPDATE)

			if err != nil {
				impl.logger.Errorw("error in saving history of ci pipeline material")
			}

		}
	} else {
		ciTemplateBean := &bean2.CiTemplateBean{
			CiTemplateOverride: &pipelineConfig.CiTemplateOverride{},
			CiBuildConfig:      nil,
			UserId:             userId,
		}
		err = impl.ciPipelineHistoryService.SaveHistory(ciPipelineObject, materials, ciTemplateBean, repository4.TRIGGER_UPDATE)
		if err != nil {
			impl.logger.Errorw("error in saving history of ci pipeline material")
		}
	}

	if len(childrenCiPipelineIds) > 0 {

		ciPipelineMaterials, err := impl.ciPipelineMaterialRepository.FindByCiPipelineIdsIn(childrenCiPipelineIds)
		if err != nil {
			impl.logger.Errorw("error in fetching  ciPipelineMaterials", "err", err)
			return nil, err
		}
		parentMaterialsMap := make(map[int]*bean.CiMaterial)
		for _, material := range createRequest.CiMaterial {
			parentMaterialsMap[material.GitMaterialId] = material
		}
		var linkedMaterials []*pipelineConfig.CiPipelineMaterial
		for _, ciPipelineMaterial := range ciPipelineMaterials {
			if parentMaterial, ok := parentMaterialsMap[ciPipelineMaterial.GitMaterialId]; ok {
				pipelineMaterial := &pipelineConfig.CiPipelineMaterial{
					Id:            ciPipelineMaterial.Id,
					Value:         parentMaterial.Source.Value,
					Active:        createRequest.Active,
					Regex:         parentMaterial.Source.Regex,
					AuditLog:      sql.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now(), CreatedOn: time.Now(), CreatedBy: userId},
					Type:          parentMaterial.Source.Type,
					GitMaterialId: parentMaterial.GitMaterialId,
					CiPipelineId:  ciPipelineMaterial.CiPipelineId,
				}
				linkedMaterials = append(linkedMaterials, pipelineMaterial)
			} else {
				impl.logger.Errorw("material not fount in patent", "gitMaterialId", ciPipelineMaterial.GitMaterialId)
				return nil, fmt.Errorf("error while updating linked pipeline")
			}
		}
		err = impl.ciPipelineMaterialRepository.Update(tx, linkedMaterials...)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return createRequest, nil
}

func (impl CiCdPipelineOrchestratorImpl) DeleteCiPipeline(pipeline *pipelineConfig.CiPipeline, request *bean.CiPatchRequest, tx *pg.Tx) error {

	userId := request.UserId
	CiEnvMappingObject, err := impl.ciPipelineRepository.FindCiEnvMappingByCiPipelineId(pipeline.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting CiEnvMappingObject ", "err", err, "ciPipelineId", pipeline.Id)
		return err
	}
	if err == nil && CiEnvMappingObject != nil {
		CiEnvMappingObject.AuditLog = sql.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now()}
		CiEnvMappingObject.Deleted = true
		err = impl.ciPipelineRepository.UpdateCiEnvMapping(CiEnvMappingObject, tx)
		if err != nil {
			impl.logger.Errorw("error in getting CiEnvMappingObject ", "err", err, "ciPipelineId", CiEnvMappingObject.CiPipelineId)
			return err
		}
	}

	p := &pipelineConfig.CiPipeline{
		Id:                       pipeline.Id,
		Deleted:                  true,
		ScanEnabled:              pipeline.ScanEnabled,
		IsManual:                 pipeline.IsManual,
		IsDockerConfigOverridden: pipeline.IsDockerConfigOverridden,
		AuditLog:                 sql.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now()},
	}

	err = impl.ciPipelineRepository.Update(p, tx)
	if err != nil {
		impl.logger.Errorw("error in updating ci pipeline, DeleteCiPipeline", "err", err, "pipelineId", pipeline.Id)
		return err
	}
	var materials []*pipelineConfig.CiPipelineMaterial
	for _, material := range pipeline.CiPipelineMaterials {
		materialDbObject, err := impl.ciPipelineMaterialRepository.GetById(material.Id)
		if err != nil {
			return err
		}
		materialDbObject.Active = false
		materials = append(materials, materialDbObject)
	}

	if request.CiPipeline.ExternalCiConfig.Id != 0 {
		err = impl.AddPipelineMaterialInGitSensor(materials)
		if err != nil {
			impl.logger.Errorw("error in saving pipelineMaterials in git sensor", "materials", materials, "err", err)
			return err
		}
	}

	err = impl.ciPipelineMaterialRepository.Update(tx, materials...)
	if err != nil {
		impl.logger.Errorw("error in updating ci pipeline materials, DeleteCiPipeline", "err", err, "pipelineId", pipeline.Id)
		return err
	}
	if !request.CiPipeline.IsDockerConfigOverridden || request.CiPipeline.IsExternal { //if pipeline is external or if config is not overridden then ignore override and ciBuildConfig values
		CiTemplateBean := bean2.CiTemplateBean{
			CiTemplate:         nil,
			CiTemplateOverride: &pipelineConfig.CiTemplateOverride{},
			CiBuildConfig:      &bean2.CiBuildConfigBean{},
			UserId:             userId,
		}
		err = impl.ciPipelineHistoryService.SaveHistory(p, materials, &CiTemplateBean, repository4.TRIGGER_DELETE)
		if err != nil {
			impl.logger.Errorw("error in saving delete history for ci pipeline material and ci template overridden", "err", err)
		}
	} else {
		CiTemplateBean := bean2.CiTemplateBean{
			CiTemplate: nil,
			CiTemplateOverride: &pipelineConfig.CiTemplateOverride{
				CiPipelineId:     request.CiPipeline.Id,
				DockerRegistryId: request.CiPipeline.DockerConfigOverride.DockerRegistry,
				DockerRepository: request.CiPipeline.DockerConfigOverride.DockerRepository,
				//DockerfilePath:   ciPipelineRequest.DockerConfigOverride.DockerBuildConfig.DockerfilePath,
				GitMaterialId: request.CiPipeline.DockerConfigOverride.CiBuildConfig.GitMaterialId,
				Active:        false,
				AuditLog: sql.AuditLog{
					CreatedBy: userId,
					CreatedOn: time.Now(),
					UpdatedBy: userId,
					UpdatedOn: time.Now(),
				},
			},
			CiBuildConfig: request.CiPipeline.DockerConfigOverride.CiBuildConfig,
			UserId:        userId,
		}
		err = impl.ciPipelineHistoryService.SaveHistory(p, materials, &CiTemplateBean, repository4.TRIGGER_DELETE)
		if err != nil {
			impl.logger.Errorw("error in saving delete history for ci pipeline material and ci template overridden", "err", err)
		}
	}

	return err
}

func (impl CiCdPipelineOrchestratorImpl) CreateCiConf(createRequest *bean.CiConfigRequest, templateId int) (*bean.CiConfigRequest, error) {
	//save pipeline in db start
	for _, ciPipeline := range createRequest.CiPipelines {
		argByte, err := json.Marshal(ciPipeline.DockerArgs)
		if err != nil {
			impl.logger.Errorw("err", "err", err)
			return nil, err
		}

		dbConnection := impl.pipelineRepository.GetConnection()
		tx, err := dbConnection.Begin()
		if err != nil {
			return nil, err
		}
		// Rollback tx on error.
		defer tx.Rollback()

		ciPipelineObject := &pipelineConfig.CiPipeline{
			AppId:                    createRequest.AppId,
			IsManual:                 ciPipeline.IsManual,
			IsExternal:               ciPipeline.IsExternal,
			CiTemplateId:             templateId,
			Version:                  ciPipeline.Version,
			Name:                     ciPipeline.Name,
			ParentCiPipeline:         ciPipeline.ParentCiPipeline,
			DockerArgs:               string(argByte),
			Active:                   true,
			Deleted:                  false,
			ScanEnabled:              createRequest.ScanEnabled,
			IsDockerConfigOverridden: ciPipeline.IsDockerConfigOverridden,
			AuditLog:                 sql.AuditLog{UpdatedBy: createRequest.UserId, CreatedBy: createRequest.UserId, UpdatedOn: time.Now(), CreatedOn: time.Now()},
		}
		err = impl.ciPipelineRepository.Save(ciPipelineObject, tx)
		ciPipeline.Id = ciPipelineObject.Id
		if err != nil {
			impl.logger.Errorw("error in saving pipeline", "ciPipelineObject", ciPipelineObject, "err", err)
			return nil, err
		}
		if createRequest.IsJob {
			CiEnvMapping := &pipelineConfig.CiEnvMapping{
				CiPipelineId:  ciPipeline.Id,
				EnvironmentId: ciPipeline.EnvironmentId,
				AuditLog:      sql.AuditLog{UpdatedBy: createRequest.UserId, CreatedBy: createRequest.UserId, UpdatedOn: time.Now(), CreatedOn: time.Now()},
			}
			err = impl.ciPipelineRepository.SaveCiEnvMapping(CiEnvMapping, tx)
			if err != nil {
				impl.logger.Errorw("error in saving pipeline", "CiEnvMapping", CiEnvMapping, "err", err)
				return nil, err
			}

			if ciPipeline.EnvironmentId != 0 && !createRequest.IsCloneJob {
				createJobEnvOverrideRequest := &bean2.CreateJobEnvOverridePayload{
					AppId:  createRequest.AppId,
					EnvId:  ciPipeline.EnvironmentId,
					UserId: createRequest.UserId,
				}
				_, err = impl.configMapService.ConfigSecretEnvironmentCreate(createJobEnvOverrideRequest)
				if err != nil && !strings.Contains(err.Error(), "already exists") {
					impl.logger.Errorw("error in saving env override", "createJobEnvOverrideRequest", createJobEnvOverrideRequest, "err", err)
					return nil, err
				}

			}
		}

		var pipelineMaterials []*pipelineConfig.CiPipelineMaterial
		for _, r := range ciPipeline.CiMaterial {
			material := &pipelineConfig.CiPipelineMaterial{
				GitMaterialId: r.GitMaterialId,
				ScmId:         r.ScmId,
				ScmVersion:    r.ScmVersion,
				ScmName:       r.ScmName,
				Value:         r.Source.Value,
				Type:          r.Source.Type,
				Path:          r.Path,
				CheckoutPath:  r.CheckoutPath,
				CiPipelineId:  ciPipelineObject.Id,
				Active:        true,
				Regex:         r.Source.Regex,
				AuditLog:      sql.AuditLog{UpdatedBy: createRequest.UserId, CreatedBy: createRequest.UserId, UpdatedOn: time.Now(), CreatedOn: time.Now()},
			}
			if material.Regex == "" && r.Source.Type == pipelineConfig.SOURCE_TYPE_BRANCH_REGEX {
				material.Regex = r.Source.Value
			}
			pipelineMaterials = append(pipelineMaterials, material)
		}
		err = impl.ciPipelineMaterialRepository.Save(tx, pipelineMaterials...)
		if err != nil {
			impl.logger.Errorf("error in saving pipelineMaterials in db", "materials", pipelineMaterials, "err", err)
			return nil, err
		}
		pmIds := make(map[string]int)
		for _, pm := range pipelineMaterials {
			key := fmt.Sprintf("%d-%d", pm.CiPipelineId, pm.GitMaterialId)
			pmIds[key] = pm.Id
		}
		for _, r := range ciPipeline.CiMaterial {
			key := fmt.Sprintf("%d-%d", ciPipelineObject.Id, r.GitMaterialId)
			r.Id = pmIds[key]
		}
		if ciPipeline.IsExternal {

		} else {
			//save pipeline in db end
			err = impl.AddPipelineMaterialInGitSensor(pipelineMaterials)
			if err != nil {
				impl.logger.Errorf("error in saving pipelineMaterials in git sensor", "materials", pipelineMaterials, "err", err)
				return nil, err
			}
		}

		//adding ci pipeline to workflow
		appWorkflowModel, err := impl.appWorkflowRepository.FindByIdAndAppId(createRequest.AppWorkflowId, createRequest.AppId)
		if err != nil && pg.ErrNoRows != err {
			return createRequest, err
		}
		if appWorkflowModel.Id > 0 {
			appWorkflowMap := &appWorkflow.AppWorkflowMapping{
				AppWorkflowId: appWorkflowModel.Id,
				ParentId:      0,
				ComponentId:   ciPipeline.Id,
				Type:          "CI_PIPELINE",
				Active:        true,
				ParentType:    "",
				AuditLog:      sql.AuditLog{CreatedBy: createRequest.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: createRequest.UserId},
			}
			_, err = impl.appWorkflowRepository.SaveAppWorkflowMapping(appWorkflowMap, tx)
			if err != nil {
				return createRequest, err
			}
		}
		err = tx.Commit()
		if err != nil {
			return nil, err
		}
		ciTemplateBean := &bean2.CiTemplateBean{}
		if ciPipeline.IsDockerConfigOverridden {
			//creating template override
			templateOverride := &pipelineConfig.CiTemplateOverride{
				CiPipelineId:     ciPipeline.Id,
				DockerRegistryId: ciPipeline.DockerConfigOverride.DockerRegistry,
				DockerRepository: ciPipeline.DockerConfigOverride.DockerRepository,
				//DockerfilePath:   ciPipelineRequest.DockerConfigOverride.DockerBuildConfig.DockerfilePath,
				GitMaterialId:             ciPipeline.DockerConfigOverride.CiBuildConfig.GitMaterialId,
				BuildContextGitMaterialId: ciPipeline.DockerConfigOverride.CiBuildConfig.BuildContextGitMaterialId,
				Active:                    true,
				AuditLog: sql.AuditLog{
					CreatedBy: createRequest.UserId,
					CreatedOn: time.Now(),
					UpdatedBy: createRequest.UserId,
					UpdatedOn: time.Now(),
				},
			}
			ciTemplateBean = &bean2.CiTemplateBean{
				CiTemplateOverride: templateOverride,
				CiBuildConfig:      ciPipeline.DockerConfigOverride.CiBuildConfig,
				UserId:             createRequest.UserId,
			}
			if !ciPipeline.IsExternal {
				err = impl.createDockerRepoIfNeeded(ciPipeline.DockerConfigOverride.DockerRegistry, ciPipeline.DockerConfigOverride.DockerRepository)
				if err != nil {
					impl.logger.Errorw("error, createDockerRepoIfNeeded", "err", err, "dockerRegistryId", ciPipeline.DockerConfigOverride.DockerRegistry, "dockerRegistry", ciPipeline.DockerConfigOverride.DockerRepository)
					return nil, err
				}
				err := impl.ciTemplateService.Save(ciTemplateBean)
				if err != nil {
					return nil, err
				}
			}
		}
		err = impl.ciPipelineHistoryService.SaveHistory(ciPipelineObject, pipelineMaterials, ciTemplateBean, repository4.TRIGGER_ADD)
		if err != nil {
			impl.logger.Errorw("error in saving history for ci pipeline", "err", err, "ciPipelineId", ciPipelineObject.Id)
		}

		//creating ci stages after tx commit due to FK constraints
		if ciPipeline.PreBuildStage != nil && len(ciPipeline.PreBuildStage.Steps) > 0 {
			//creating pre stage
			err = impl.pipelineStageService.CreatePipelineStage(ciPipeline.PreBuildStage, repository5.PIPELINE_STAGE_TYPE_PRE_CI, ciPipeline.Id, createRequest.UserId)
			if err != nil {
				impl.logger.Errorw("error in creating pre stage", "err", err, "preBuildStage", ciPipeline.PreBuildStage, "ciPipelineId", ciPipeline.Id)
				return nil, err
			}
		}
		if ciPipeline.PostBuildStage != nil && len(ciPipeline.PostBuildStage.Steps) > 0 {
			//creating post stage
			err = impl.pipelineStageService.CreatePipelineStage(ciPipeline.PostBuildStage, repository5.PIPELINE_STAGE_TYPE_POST_CI, ciPipeline.Id, createRequest.UserId)
			if err != nil {
				impl.logger.Errorw("error in creating post stage", "err", err, "postBuildStage", ciPipeline.PostBuildStage, "ciPipelineId", ciPipeline.Id)
				return nil, err
			}
		}
		for _, r := range ciPipeline.CiMaterial {
			ciMaterial, err := impl.ciPipelineMaterialRepository.GetById(r.Id)
			if err != nil && pg.ErrNoRows != err {
				return nil, err
			}
			if ciMaterial != nil && ciMaterial.GitMaterial != nil {
				r.GitMaterialName = ciMaterial.GitMaterial.Name[strings.Index(ciMaterial.GitMaterial.Name, "-")+1:]
			}
		}
	}
	return createRequest, nil
}

func (impl CiCdPipelineOrchestratorImpl) BuildCiPipelineScript(userId int32, ciScript *bean.CiScript, scriptStage string, ciPipeline *bean.CiPipeline) *pipelineConfig.CiPipelineScript {
	ciPipelineScript := &pipelineConfig.CiPipelineScript{
		Name:           ciScript.Name,
		Index:          ciScript.Index,
		CiPipelineId:   ciPipeline.Id,
		Script:         ciScript.Script,
		Stage:          scriptStage,
		Active:         true,
		OutputLocation: ciScript.OutputLocation,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
	if ciScript.Id != 0 {
		ciPipelineScript.Id = ciScript.Id
	}
	return ciPipelineScript
}

func (impl CiCdPipelineOrchestratorImpl) generateApiKey(ciPipelineId int, ciPipelineName string, secret string) (prefix, sha string) {
	hashData := strconv.Itoa(ciPipelineId) + "-" + ciPipelineName + "-" + time.Now().String()
	prefix = base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(ciPipelineId)))
	h := hmac.New(sha256.New, []byte(secret))
	_, err := h.Write([]byte(hashData))
	if err != nil {
		impl.logger.Error(err)
	}
	sha = hex.EncodeToString(h.Sum(nil))
	return prefix, sha
}

func (impl CiCdPipelineOrchestratorImpl) generateExternalCiPayload(ciPipeline *bean.CiPipeline, externalCiPipeline *pipelineConfig.ExternalCiPipeline, keyPrefix string, apiKey string) *bean.CiPipeline {
	if impl.ciConfig.ExternalCiWebhookUrl == "" {
		hostUrl, err := impl.attributesService.GetByKey(attributes.HostUrlKey)
		if err != nil {
			impl.logger.Errorw("there is no external ci webhook url configured", "ci pipeline", ciPipeline)
			return nil
		}
		if hostUrl != nil {
			impl.ciConfig.ExternalCiWebhookUrl = fmt.Sprintf("%s/%s", hostUrl.Value, ExternalCiWebhookPath)
		}
	}
	accessKey := keyPrefix + "." + apiKey
	ciPipeline.ExternalCiConfig = bean.ExternalCiConfig{
		WebhookUrl: impl.ciConfig.ExternalCiWebhookUrl,
		AccessKey:  accessKey,
		Payload:    impl.ciConfig.ExternalCiPayload,
	}
	return ciPipeline
}

func (impl CiCdPipelineOrchestratorImpl) AddPipelineMaterialInGitSensor(pipelineMaterials []*pipelineConfig.CiPipelineMaterial) error {
	var materials []*gitSensor.CiPipelineMaterial
	for _, ciPipelineMaterial := range pipelineMaterials {
		if ciPipelineMaterial.Type != pipelineConfig.SOURCE_TYPE_BRANCH_REGEX {
			material := &gitSensor.CiPipelineMaterial{
				Id:            ciPipelineMaterial.Id,
				Active:        ciPipelineMaterial.Active,
				Value:         ciPipelineMaterial.Value,
				GitMaterialId: ciPipelineMaterial.GitMaterialId,
				Type:          gitSensor.SourceType(ciPipelineMaterial.Type),
			}
			materials = append(materials, material)
		}
	}

	return impl.GitSensorClient.SavePipelineMaterial(context.Background(), materials)
}

func (impl CiCdPipelineOrchestratorImpl) CheckStringMatchRegex(regex string, value string) bool {
	response, err := regexp.MatchString(regex, value)
	if err != nil {
		return false
	}
	return response
}

func (impl CiCdPipelineOrchestratorImpl) CreateApp(createRequest *bean.CreateAppDTO) (*bean.CreateAppDTO, error) {
	// validate the labels key-value if propagate is true
	for _, label := range createRequest.AppLabels {
		if !label.Propagate {
			continue
		}
		labelKey := label.Key
		labelValue := label.Value
		err := util3.CheckIfValidLabel(labelKey, labelValue)
		if err != nil {
			return nil, err
		}
	}

	dbConnection := impl.appRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	app, err := impl.createAppGroup(createRequest.AppName, createRequest.UserId, createRequest.TeamId, createRequest.AppType, tx)
	if err != nil {
		return nil, err
	}
	err = impl.storeDescription(tx, createRequest, app.Id)
	if err != nil {
		impl.logger.Errorw("error in saving description", "err", err, "descriptionObj", createRequest.Description, "userId", createRequest.UserId)
		return nil, err
	}
	// create labels and tags with app
	if app.Active && len(createRequest.AppLabels) > 0 {
		for _, label := range createRequest.AppLabels {
			request := &bean.AppLabelDto{
				AppId:     app.Id,
				Key:       label.Key,
				Value:     label.Value,
				Propagate: label.Propagate,
				UserId:    createRequest.UserId,
			}
			_, err := impl.appLabelsService.Create(request, tx)
			if err != nil {
				impl.logger.Errorw("error on creating labels for app id ", "err", err, "appId", app.Id)
				return nil, err
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in commit repo", "error", err)
		return nil, err
	}
	createRequest.Id = app.Id
	if createRequest.AppType == helper.Job {
		createRequest.AppName = app.DisplayName
	}
	return createRequest, nil
}

func (impl CiCdPipelineOrchestratorImpl) storeDescription(tx *pg.Tx, createRequest *bean.CreateAppDTO, appId int) error {
	if createRequest.Description != nil && createRequest.Description.Description != "" {
		descriptionObj := repository3.GenericNote{
			Description:    createRequest.Description.Description,
			IdentifierType: repository3.AppType,
			Identifier:     appId,
		}
		note, err := impl.genericNoteService.Save(tx, &descriptionObj, createRequest.UserId)
		if err != nil {
			impl.logger.Errorw("error in saving description", "err", err, "descriptionObj", descriptionObj, "userId", createRequest.UserId)
			return err
		}
		createRequest.Description = note
	}
	return nil
}

func (impl CiCdPipelineOrchestratorImpl) DeleteApp(appId int, userId int32) error {
	// Delete git materials,call git sensor and delete app
	impl.logger.Debug("deleting materials in orchestrator")
	materials, err := impl.materialRepository.FindByAppId(appId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("err", err)
		return err
	}
	for i := range materials {
		materials[i].Active = false
		materials[i].UpdatedOn = time.Now()
		materials[i].UpdatedBy = userId
	}
	err = impl.materialRepository.Update(materials)

	if err != nil {
		impl.logger.Errorw("could not delete materials ", "err", err)
		return err
	}

	err = impl.gitMaterialHistoryService.CreateDeleteMaterialHistory(materials)

	impl.logger.Debug("deleting materials in git_sensor")
	for _, m := range materials {
		err = impl.updateRepositoryToGitSensor(m)
		if err != nil {
			impl.logger.Errorw("error in updating to git-sensor", "err", err)
			return err
		}
	}

	app, err := impl.appRepository.FindById(appId)
	if err != nil {
		impl.logger.Errorw("err", err)
		return err
	}
	dbConnection := impl.appRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in establishing connection", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	app.Active = false
	app.UpdatedOn = time.Now()
	app.UpdatedBy = userId
	err = impl.appRepository.UpdateWithTxn(app, tx)
	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return err
	}
	//deleting auth roles entries for this project
	err = impl.userAuthService.DeleteRoles(bean3.APP_TYPE, app.AppName, tx, "")
	if err != nil {
		impl.logger.Errorw("error in deleting auth roles", "err", err)
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (impl CiCdPipelineOrchestratorImpl) CreateMaterials(createMaterialRequest *bean.CreateMaterialDTO) (*bean.CreateMaterialDTO, error) {
	existingMaterials, err := impl.materialRepository.FindByAppId(createMaterialRequest.AppId)
	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return nil, err
	}
	checkoutPaths := make(map[int]string)
	impl.logger.Debugw("existing materials", "material", existingMaterials)
	for _, material := range existingMaterials {
		checkoutPaths[material.Id] = material.CheckoutPath
	}
	for i, material := range createMaterialRequest.Material {
		if material.CheckoutPath == "" {
			material.CheckoutPath = "./"
		}
		checkoutPaths[i*-1] = material.CheckoutPath
	}
	duplicatePathErr := impl.validateCheckoutPathsForMultiGit(checkoutPaths)
	if duplicatePathErr != nil {
		impl.logger.Errorw("duplicate checkout paths", "err", err)
		return nil, duplicatePathErr
	}
	var materials []*bean.GitMaterial
	for _, inputMaterial := range createMaterialRequest.Material {
		m, err := impl.createMaterial(inputMaterial, createMaterialRequest.AppId, createMaterialRequest.UserId)
		inputMaterial.Id = m.Id
		if err != nil {
			return nil, err
		}
		materials = append(materials, inputMaterial)
	}
	err = impl.addRepositoryToGitSensor(materials)
	if err != nil {
		impl.logger.Errorw("error in updating to sensor", "err", err)
		return nil, err
	}
	impl.logger.Debugw("all materials are ", "materials", materials)
	return createMaterialRequest, nil
}

func (impl CiCdPipelineOrchestratorImpl) UpdateMaterial(updateMaterialDTO *bean.UpdateMaterialDTO) (*bean.UpdateMaterialDTO, error) {
	updatedMaterial, err := impl.updateMaterial(updateMaterialDTO)
	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return nil, err
	}

	err = impl.updateRepositoryToGitSensor(updatedMaterial)
	if err != nil {
		impl.logger.Errorw("error in updating to git-sensor", "err", err)
		return nil, err
	}
	return updateMaterialDTO, nil
}

func (impl CiCdPipelineOrchestratorImpl) updateRepositoryToGitSensor(material *pipelineConfig.GitMaterial) error {
	sensorMaterial := &gitSensor.GitMaterial{
		Name:             material.Name,
		Url:              material.Url,
		Id:               material.Id,
		GitProviderId:    material.GitProviderId,
		CheckoutLocation: material.CheckoutPath,
		Deleted:          !material.Active,
		FetchSubmodules:  material.FetchSubmodules,
		FilterPattern:    material.FilterPattern,
	}
	return impl.GitSensorClient.UpdateRepo(context.Background(), sensorMaterial)
}

func (impl CiCdPipelineOrchestratorImpl) addRepositoryToGitSensor(materials []*bean.GitMaterial) error {
	var sensorMaterials []*gitSensor.GitMaterial
	for _, material := range materials {
		sensorMaterial := &gitSensor.GitMaterial{
			Name:            material.Name,
			Url:             material.Url,
			Id:              material.Id,
			GitProviderId:   material.GitProviderId,
			Deleted:         false,
			FetchSubmodules: material.FetchSubmodules,
			FilterPattern:   material.FilterPattern,
		}
		sensorMaterials = append(sensorMaterials, sensorMaterial)
	}
	return impl.GitSensorClient.AddRepo(context.Background(), sensorMaterials)
}

// FIXME: not thread safe
func (impl CiCdPipelineOrchestratorImpl) createAppGroup(name string, userId int32, teamId int, appType helper.AppType, tx *pg.Tx) (*app2.App, error) {
	app, err := impl.appRepository.FindActiveByName(name)
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	if appType != helper.Job {
		if app != nil && app.Id > 0 {
			impl.logger.Warnw("app already exists", "name", name)
			err = &util.ApiError{
				Code:            constants.AppAlreadyExists.Code,
				InternalMessage: "app already exists",
				UserMessage:     constants.AppAlreadyExists.UserMessage(name),
			}
			return nil, err
		}
	} else {
		job, err := impl.appRepository.FindJobByDisplayName(name)
		if err != nil && err != pg.ErrNoRows {
			return nil, err
		}
		if job != nil && job.Id > 0 {
			impl.logger.Warnw("job already exists", "name", name)
			err = &util.ApiError{
				Code:            constants.AppAlreadyExists.Code,
				InternalMessage: "job already exists",
				UserMessage:     constants.AppAlreadyExists.UserMessage(name),
			}
			return nil, err
		}
	}

	displayName := name
	appName := name
	if appType == helper.Job {
		appName = name + "/" + util2.Generate(8) + "J"
	}
	pg := &app2.App{
		Active:      true,
		AppName:     appName,
		DisplayName: displayName,
		TeamId:      teamId,
		AppType:     appType,
		AuditLog:    sql.AuditLog{UpdatedBy: userId, CreatedBy: userId, UpdatedOn: time.Now(), CreatedOn: time.Now()},
	}
	err = impl.appRepository.SaveWithTxn(pg, tx)
	if err != nil {
		impl.logger.Errorw("error in saving entity ", "entity", pg)
		return nil, err
	}

	apps, err := impl.appRepository.FindActiveListByName(name)
	if err != nil {
		return nil, err
	}
	appLen := len(apps)
	if appLen > 1 {
		firstElement := apps[0]
		if firstElement.Id != pg.Id {
			pg.Active = false
			err = impl.appRepository.UpdateWithTxn(pg, tx)
			if err != nil {
				impl.logger.Errorw("error in saving entity ", "entity", pg)
				return nil, err
			}
			err = &util.ApiError{
				Code:            constants.AppAlreadyExists.Code,
				InternalMessage: "app already exists",
				UserMessage:     constants.AppAlreadyExists.UserMessage(name),
			}
			return nil, err
		}
	}
	return pg, nil
}

func (impl CiCdPipelineOrchestratorImpl) validateCheckoutPathsForMultiGit(allPaths map[int]string) error {
	dockerfilePathMap := make(map[string]bool)
	impl.logger.Debugw("all paths ", "path", allPaths)
	isMulti := len(allPaths) > 1
	for _, c := range allPaths {
		if isMulti && (c == "") {
			impl.logger.Errorw("validation err", "err", "checkout path required for multi-git")
			return fmt.Errorf("checkout path required for multi-git")
		}
		if !strings.HasPrefix(c, "./") {
			impl.logger.Errorw("validation err", "err", "invalid checkout path it must start with ./")
			return fmt.Errorf("invalid checkout path it must start with ./ ")
		}
		if _, ok := dockerfilePathMap[c]; ok {
			impl.logger.Error("duplicate checkout paths found")
			return errors.New("duplicate checkout paths found")
		}
		dockerfilePathMap[c] = true
	}
	return nil
}

func (impl CiCdPipelineOrchestratorImpl) updateMaterial(updateMaterialDTO *bean.UpdateMaterialDTO) (*pipelineConfig.GitMaterial, error) {
	existingMaterials, err := impl.materialRepository.FindByAppId(updateMaterialDTO.AppId)
	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return nil, err
	}
	checkoutPaths := make(map[int]string)
	for _, material := range existingMaterials {
		checkoutPaths[material.Id] = material.CheckoutPath
	}
	var currentMaterial *pipelineConfig.GitMaterial
	for _, m := range existingMaterials {
		if m.Id == updateMaterialDTO.Material.Id {
			currentMaterial = m
			break
		}
	}
	if currentMaterial == nil {
		return nil, errors.New("material to be updated does not exist")
	}
	if updateMaterialDTO.Material.CheckoutPath == "" {
		updateMaterialDTO.Material.CheckoutPath = "./"
	}
	checkoutPaths[updateMaterialDTO.Material.Id] = updateMaterialDTO.Material.CheckoutPath
	validationErr := impl.validateCheckoutPathsForMultiGit(checkoutPaths)
	if validationErr != nil {
		impl.logger.Errorw("validation err", "err", err)
		return nil, validationErr
	}
	currentMaterial.Url = updateMaterialDTO.Material.Url
	basePath := path.Base(updateMaterialDTO.Material.Url)
	basePath = strings.TrimSuffix(basePath, ".git")

	currentMaterial.Name = strconv.Itoa(updateMaterialDTO.Material.GitProviderId) + "-" + basePath
	currentMaterial.GitProviderId = updateMaterialDTO.Material.GitProviderId
	currentMaterial.CheckoutPath = updateMaterialDTO.Material.CheckoutPath
	currentMaterial.FetchSubmodules = updateMaterialDTO.Material.FetchSubmodules
	currentMaterial.FilterPattern = updateMaterialDTO.Material.FilterPattern
	currentMaterial.AuditLog = sql.AuditLog{UpdatedBy: updateMaterialDTO.UserId, CreatedBy: currentMaterial.CreatedBy, UpdatedOn: time.Now(), CreatedOn: currentMaterial.CreatedOn}

	err = impl.materialRepository.UpdateMaterial(currentMaterial)

	if err != nil {
		impl.logger.Errorw("error in updating material", "material", currentMaterial, "err", err)
		return nil, err
	}

	err = impl.gitMaterialHistoryService.CreateMaterialHistory(currentMaterial)

	return currentMaterial, nil
}

func (impl CiCdPipelineOrchestratorImpl) createMaterial(inputMaterial *bean.GitMaterial, appId int, userId int32) (*pipelineConfig.GitMaterial, error) {
	basePath := path.Base(inputMaterial.Url)
	basePath = strings.TrimSuffix(basePath, ".git")
	material := &pipelineConfig.GitMaterial{
		Url:             inputMaterial.Url,
		AppId:           appId,
		Name:            strconv.Itoa(inputMaterial.GitProviderId) + "-" + basePath,
		GitProviderId:   inputMaterial.GitProviderId,
		Active:          true,
		CheckoutPath:    inputMaterial.CheckoutPath,
		FetchSubmodules: inputMaterial.FetchSubmodules,
		FilterPattern:   inputMaterial.FilterPattern,
		AuditLog:        sql.AuditLog{UpdatedBy: userId, CreatedBy: userId, UpdatedOn: time.Now(), CreatedOn: time.Now()},
	}
	err := impl.materialRepository.SaveMaterial(material)
	if err != nil {
		impl.logger.Errorw("error in saving material", "material", material, "err", err)
		return nil, err
	}
	err = impl.gitMaterialHistoryService.CreateMaterialHistory(material)
	return material, err
}

func (impl CiCdPipelineOrchestratorImpl) CreateCDPipelines(pipelineRequest *bean.CDPipelineConfigObject, appId int, userId int32, tx *pg.Tx, appName string) (pipelineId int, err error) {
	preStageConfig := ""
	preTriggerType := pipelineConfig.TriggerType("")
	if len(pipelineRequest.PreStage.Config) > 0 {
		preStageConfig = pipelineRequest.PreStage.Config
		preTriggerType = pipelineRequest.PreStage.TriggerType
	}

	if pipelineRequest.PreDeployStage != nil {
		preTriggerType = pipelineRequest.PreDeployStage.TriggerType
	}

	postStageConfig := ""
	postTriggerType := pipelineConfig.TriggerType("")
	if len(pipelineRequest.PostStage.Config) > 0 {
		postStageConfig = pipelineRequest.PostStage.Config
		postTriggerType = pipelineRequest.PostStage.TriggerType
	}

	if pipelineRequest.PostDeployStage != nil {
		postTriggerType = pipelineRequest.PostDeployStage.TriggerType
	}

	preStageConfigMapSecretNames, err := json.Marshal(&pipelineRequest.PreStageConfigMapSecretNames)
	if err != nil {
		impl.logger.Error(err)
		return 0, err
	}

	postStageConfigMapSecretNames, err := json.Marshal(&pipelineRequest.PostStageConfigMapSecretNames)
	if err != nil {
		impl.logger.Error(err)
		return 0, err
	}

	env, err := impl.envRepository.FindById(pipelineRequest.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in getting environment by id", "err", err)
		return 0, err
	}
	pipeline := &pipelineConfig.Pipeline{
		EnvironmentId:                 pipelineRequest.EnvironmentId,
		AppId:                         appId,
		Name:                          pipelineRequest.Name,
		Deleted:                       false,
		CiPipelineId:                  pipelineRequest.CiPipelineId,
		TriggerType:                   pipelineRequest.TriggerType,
		PreStageConfig:                preStageConfig,
		PostStageConfig:               postStageConfig,
		PreTriggerType:                preTriggerType,
		PostTriggerType:               postTriggerType,
		PreStageConfigMapSecretNames:  string(preStageConfigMapSecretNames),
		PostStageConfigMapSecretNames: string(postStageConfigMapSecretNames),
		RunPreStageInEnv:              pipelineRequest.RunPreStageInEnv,
		RunPostStageInEnv:             pipelineRequest.RunPostStageInEnv,
		DeploymentAppCreated:          false,
		DeploymentAppType:             pipelineRequest.DeploymentAppType,
		DeploymentAppName:             fmt.Sprintf("%s-%s", appName, env.Name),
		AuditLog:                      sql.AuditLog{UpdatedBy: userId, CreatedBy: userId, UpdatedOn: time.Now(), CreatedOn: time.Now()},
	}
	err = impl.pipelineRepository.Save([]*pipelineConfig.Pipeline{pipeline}, tx)
	if err != nil {
		impl.logger.Errorw("error in saving cd pipeline", "err", err, "pipeline", pipeline)
		return 0, err
	}
	if pipeline.PreStageConfig != "" {
		err = impl.prePostCdScriptHistoryService.CreatePrePostCdScriptHistory(pipeline, tx, repository4.PRE_CD_TYPE, false, 0, time.Time{})
		if err != nil {
			impl.logger.Errorw("error in creating pre cd script entry", "err", err, "pipeline", pipeline)
			return 0, err
		}
	}
	if pipeline.PostStageConfig != "" {
		err = impl.prePostCdScriptHistoryService.CreatePrePostCdScriptHistory(pipeline, tx, repository4.POST_CD_TYPE, false, 0, time.Time{})
		if err != nil {
			impl.logger.Errorw("error in creating post cd script entry", "err", err, "pipeline", pipeline)
			return 0, err
		}
	}
	return pipeline.Id, nil
}

func (impl CiCdPipelineOrchestratorImpl) UpdateCDPipeline(pipelineRequest *bean.CDPipelineConfigObject, userId int32, tx *pg.Tx) (err error) {
	pipeline, err := impl.pipelineRepository.FindById(pipelineRequest.Id)
	if err == pg.ErrNoRows {
		return fmt.Errorf("no cd pipeline found")
	} else if err != nil {
		return err
	} else if pipeline.Id == 0 {
		return fmt.Errorf("no cd pipeline found")
	}
	preStageConfig := ""
	preTriggerType := pipelineConfig.TriggerType("")
	if len(pipelineRequest.PreStage.Config) > 0 {
		preStageConfig = pipelineRequest.PreStage.Config
		preTriggerType = pipelineRequest.PreStage.TriggerType
	}
	if pipelineRequest.PreDeployStage != nil {
		preTriggerType = pipelineRequest.PreDeployStage.TriggerType
	}

	postStageConfig := ""
	postTriggerType := pipelineConfig.TriggerType("")
	if len(pipelineRequest.PostStage.Config) > 0 {
		postStageConfig = pipelineRequest.PostStage.Config
		postTriggerType = pipelineRequest.PostStage.TriggerType
	}
	if pipelineRequest.PostDeployStage != nil {
		postTriggerType = pipelineRequest.PostDeployStage.TriggerType
	}

	preStageConfigMapSecretNames, err := json.Marshal(&pipelineRequest.PreStageConfigMapSecretNames)
	if err != nil {
		impl.logger.Error(err)
		return err
	}

	postStageConfigMapSecretNames, err := json.Marshal(&pipelineRequest.PostStageConfigMapSecretNames)
	if err != nil {
		impl.logger.Error(err)
		return err
	}

	pipeline.TriggerType = pipelineRequest.TriggerType
	pipeline.PreTriggerType = preTriggerType
	pipeline.PostTriggerType = postTriggerType
	pipeline.PreStageConfig = preStageConfig
	pipeline.PostStageConfig = postStageConfig
	pipeline.PreStageConfigMapSecretNames = string(preStageConfigMapSecretNames)
	pipeline.PostStageConfigMapSecretNames = string(postStageConfigMapSecretNames)
	pipeline.RunPreStageInEnv = pipelineRequest.RunPreStageInEnv
	pipeline.RunPostStageInEnv = pipelineRequest.RunPostStageInEnv
	pipeline.UpdatedBy = userId
	pipeline.UpdatedOn = time.Now()
	err = impl.pipelineRepository.Update(pipeline, tx)
	if err != nil {
		impl.logger.Errorw("error in updating cd pipeline", "err", err, "pipeline", pipeline)
		return err
	}
	if pipeline.PreStageConfig != "" {
		err = impl.prePostCdScriptHistoryService.CreatePrePostCdScriptHistory(pipeline, tx, repository4.PRE_CD_TYPE, false, 0, time.Time{})
		if err != nil {
			impl.logger.Errorw("error in creating pre cd script entry", "err", err, "pipeline", pipeline)
			return err
		}
	}
	if pipeline.PostStageConfig != "" {
		err = impl.prePostCdScriptHistoryService.CreatePrePostCdScriptHistory(pipeline, tx, repository4.POST_CD_TYPE, false, 0, time.Time{})
		if err != nil {
			impl.logger.Errorw("error in creating post cd script entry", "err", err, "pipeline", pipeline)
			return err
		}
	}

	if pipelineRequest.PreDeployStage != nil {
		//updating pre stage
		err = impl.pipelineStageService.UpdatePipelineStage(pipelineRequest.PreDeployStage, repository5.PIPELINE_STAGE_TYPE_PRE_CD, pipelineRequest.Id, userId)
		if err != nil {
			impl.logger.Errorw("error in updating pre stage", "err", err, "preDeployStage", pipelineRequest.PreDeployStage, "cdPipelineId", pipelineRequest.Id)
			return err
		}
	}
	if pipelineRequest.PostDeployStage != nil {
		//updating post stage
		err = impl.pipelineStageService.UpdatePipelineStage(pipelineRequest.PostDeployStage, repository5.PIPELINE_STAGE_TYPE_POST_CD, pipelineRequest.Id, userId)
		if err != nil {
			impl.logger.Errorw("error in updating post stage", "err", err, "postDeployStage", pipelineRequest.PostDeployStage, "cdPipelineId", pipelineRequest.Id)
			return err
		}
	}
	return err
}

func (impl CiCdPipelineOrchestratorImpl) DeleteCdPipeline(pipelineId int, userId int32, tx *pg.Tx) error {
	return impl.pipelineRepository.Delete(pipelineId, userId, tx)
}
func (impl CiCdPipelineOrchestratorImpl) getPipelineIdAndPrePostStageMapping(dbPipelines []*pipelineConfig.Pipeline) (map[int][]*bean2.PipelineStageDto, error) {
	var err error
	pipelineIdAndPrePostStageMapping := make(map[int][]*bean2.PipelineStageDto)
	var dbPipelineIds []int
	for _, pipeline := range dbPipelines {
		dbPipelineIds = append(dbPipelineIds, pipeline.Id)
	}
	if len(dbPipelineIds) > 0 {
		pipelineIdAndPrePostStageMapping, err = impl.pipelineStageService.GetCdPipelineStageDataDeepCopyForPipelineIds(dbPipelineIds)
		if err != nil {
			impl.logger.Errorw("error in fetching pipelinePrePostStageMapping", "err", err, "cdPipelineIds", dbPipelineIds)
			return pipelineIdAndPrePostStageMapping, err
		}
	}
	return pipelineIdAndPrePostStageMapping, nil
}

func (impl CiCdPipelineOrchestratorImpl) PipelineExists(name string) (bool, error) {
	return impl.pipelineRepository.PipelineExists(name)
}

func (impl CiCdPipelineOrchestratorImpl) GetCdPipelinesForApp(appId int) (cdPipelines *bean.CdPipelines, err error) {
	dbPipelines, err := impl.pipelineRepository.FindActiveByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching cdPipeline", "appId", appId, "err", err)
	}
	pipelineIdAndPrePostStageMapping, err := impl.getPipelineIdAndPrePostStageMapping(dbPipelines)
	if err != nil {
		impl.logger.Errorw("error in fetching pipelineIdAndPrePostStageMapping", "err", err)
		return nil, err
	}
	var pipelines []*bean.CDPipelineConfigObject
	for _, dbPipeline := range dbPipelines {
		preStage := bean.CdStage{}
		if len(dbPipeline.PreStageConfig) > 0 {
			preStage.Name = "Pre-Deployment"
			preStage.Config = dbPipeline.PreStageConfig
			preStage.TriggerType = dbPipeline.PreTriggerType
		}
		postStage := bean.CdStage{}
		if len(dbPipeline.PostStageConfig) > 0 {
			postStage.Name = "Post-Deployment"
			postStage.Config = dbPipeline.PostStageConfig
			postStage.TriggerType = dbPipeline.PostTriggerType
		}

		preStageConfigmapSecrets := bean.PreStageConfigMapSecretNames{}
		postStageConfigmapSecrets := bean.PostStageConfigMapSecretNames{}

		if dbPipeline.PreStageConfigMapSecretNames != "" {
			err = json.Unmarshal([]byte(dbPipeline.PreStageConfigMapSecretNames), &preStageConfigmapSecrets)
			if err != nil {
				impl.logger.Error(err)
				return nil, err
			}
		}
		if dbPipeline.PostStageConfigMapSecretNames != "" {
			err = json.Unmarshal([]byte(dbPipeline.PostStageConfigMapSecretNames), &postStageConfigmapSecrets)
			if err != nil {
				impl.logger.Error(err)
				return nil, err
			}
		}

		pipeline := &bean.CDPipelineConfigObject{
			Id:                            dbPipeline.Id,
			Name:                          dbPipeline.Name,
			EnvironmentId:                 dbPipeline.EnvironmentId,
			EnvironmentName:               dbPipeline.Environment.Name,
			Description:                   dbPipeline.Environment.Description,
			CiPipelineId:                  dbPipeline.CiPipelineId,
			TriggerType:                   dbPipeline.TriggerType,
			PreStage:                      preStage,
			PostStage:                     postStage,
			RunPreStageInEnv:              dbPipeline.RunPreStageInEnv,
			RunPostStageInEnv:             dbPipeline.RunPostStageInEnv,
			PreStageConfigMapSecretNames:  preStageConfigmapSecrets,
			PostStageConfigMapSecretNames: postStageConfigmapSecrets,
			DeploymentAppType:             dbPipeline.DeploymentAppType,
			DeploymentAppDeleteRequest:    dbPipeline.DeploymentAppDeleteRequest,
			IsVirtualEnvironment:          dbPipeline.Environment.IsVirtualEnvironment,
		}
		if pipelineStages, ok := pipelineIdAndPrePostStageMapping[dbPipeline.Id]; ok {
			pipeline.PreDeployStage = pipelineStages[0]
			pipeline.PostDeployStage = pipelineStages[1]
		}
		pipelines = append(pipelines, pipeline)
	}
	cdPipelines = &bean.CdPipelines{
		AppId:     appId,
		Pipelines: pipelines,
	}
	if len(pipelines) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no cd pipeline found"}
	} else {
		err = nil
	}
	return cdPipelines, err
}

func (impl CiCdPipelineOrchestratorImpl) GetCdPipelinesForEnv(envId int, requestedAppIds []int) (cdPipelines *bean.CdPipelines, err error) {
	var dbPipelines []*pipelineConfig.Pipeline
	if len(requestedAppIds) > 0 {
		dbPipelines, err = impl.pipelineRepository.FindActiveByInFilter(envId, requestedAppIds)
	} else {
		dbPipelines, err = impl.pipelineRepository.FindActiveByEnvId(envId)
	}
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines", "envId", envId, "err", err)
		return nil, err
	}

	var appIds []int
	for _, pipeline := range dbPipelines {
		appIds = append(appIds, pipeline.AppId)
	}
	if len(appIds) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no cd pipeline found"}
		return cdPipelines, err
	}

	// fetch other environments also which are linked with this app
	dbPipelines, err = impl.pipelineRepository.FindActiveByAppIds(appIds)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error fetching pipelines for env id", "err", err)
		return nil, err
	}
	pipelineIdAndPrePostStageMapping, err := impl.getPipelineIdAndPrePostStageMapping(dbPipelines)
	if err != nil {
		impl.logger.Errorw("error in fetching pipelineIdAndPrePostStageMapping", "err", err)
		return nil, err
	}

	var pipelines []*bean.CDPipelineConfigObject
	for _, dbPipeline := range dbPipelines {
		pipeline := &bean.CDPipelineConfigObject{
			Id:                    dbPipeline.Id,
			Name:                  dbPipeline.Name,
			EnvironmentId:         dbPipeline.EnvironmentId,
			EnvironmentName:       dbPipeline.Environment.Name,
			CiPipelineId:          dbPipeline.CiPipelineId,
			TriggerType:           dbPipeline.TriggerType,
			RunPreStageInEnv:      dbPipeline.RunPreStageInEnv,
			RunPostStageInEnv:     dbPipeline.RunPostStageInEnv,
			DeploymentAppType:     dbPipeline.DeploymentAppType,
			AppName:               dbPipeline.App.AppName,
			AppId:                 dbPipeline.AppId,
			TeamId:                dbPipeline.App.TeamId,
			EnvironmentIdentifier: dbPipeline.Environment.EnvironmentIdentifier,
			IsVirtualEnvironment:  dbPipeline.Environment.IsVirtualEnvironment,
		}
		if len(dbPipeline.PreStageConfig) > 0 {
			preStage := bean.CdStage{}
			preStage.Name = "Pre-Deployment"
			preStage.Config = dbPipeline.PreStageConfig
			preStage.TriggerType = dbPipeline.PreTriggerType
			pipeline.PreStage = preStage
		}
		if len(dbPipeline.PostStageConfig) > 0 {
			postStage := bean.CdStage{}
			postStage.Name = "Post-Deployment"
			postStage.Config = dbPipeline.PostStageConfig
			postStage.TriggerType = dbPipeline.PostTriggerType
			pipeline.PostStage = postStage
		}
		if pipelineStages, ok := pipelineIdAndPrePostStageMapping[dbPipeline.Id]; ok {
			pipeline.PreDeployStage = pipelineStages[0]
			pipeline.PostDeployStage = pipelineStages[1]
		}

		if dbPipeline.PreStageConfigMapSecretNames != "" {
			preStageConfigmapSecrets := bean.PreStageConfigMapSecretNames{}
			err = json.Unmarshal([]byte(dbPipeline.PreStageConfigMapSecretNames), &preStageConfigmapSecrets)
			if err != nil {
				impl.logger.Errorw("unmarshal error", "err", err)
				return nil, err
			}
			pipeline.PreStageConfigMapSecretNames = preStageConfigmapSecrets
		}
		if dbPipeline.PostStageConfigMapSecretNames != "" {
			postStageConfigmapSecrets := bean.PostStageConfigMapSecretNames{}
			err = json.Unmarshal([]byte(dbPipeline.PostStageConfigMapSecretNames), &postStageConfigmapSecrets)
			if err != nil {
				impl.logger.Errorw("unmarshal error", "err", err)
				return nil, err
			}
			pipeline.PostStageConfigMapSecretNames = postStageConfigmapSecrets
		}

		pipelines = append(pipelines, pipeline)
	}
	cdPipelines = &bean.CdPipelines{
		Pipelines: pipelines,
	}
	if len(pipelines) == 0 {
		err = &util.ApiError{Code: "404", HttpStatusCode: 200, UserMessage: "no cd pipeline found"}
	} else {
		err = nil
	}
	return cdPipelines, err
}

func (impl CiCdPipelineOrchestratorImpl) GetCdPipelinesForAppAndEnv(appId int, envId int) (cdPipelines *bean.CdPipelines, err error) {
	dbPipelines, err := impl.pipelineRepository.FindActiveByAppIdAndEnvironmentId(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in fetching cdPipeline", "appId", appId, "err", err)
	}
	pipelineIdAndPrePostStageMapping, err := impl.getPipelineIdAndPrePostStageMapping(dbPipelines)
	if err != nil {
		impl.logger.Errorw("error in fetching pipelineIdAndPrePostStageMapping", "err", err)
		return nil, err
	}
	var pipelines []*bean.CDPipelineConfigObject
	for _, dbPipeline := range dbPipelines {
		preStage := bean.CdStage{}
		if len(dbPipeline.PreStageConfig) > 0 {
			preStage.Name = "Pre-Deployment"
			preStage.Config = dbPipeline.PreStageConfig
			preStage.TriggerType = dbPipeline.PreTriggerType
		}
		postStage := bean.CdStage{}
		if len(dbPipeline.PostStageConfig) > 0 {
			postStage.Name = "Post-Deployment"
			postStage.Config = dbPipeline.PostStageConfig
			postStage.TriggerType = dbPipeline.PostTriggerType
		}

		preStageConfigmapSecrets := bean.PreStageConfigMapSecretNames{}
		postStageConfigmapSecrets := bean.PostStageConfigMapSecretNames{}

		if dbPipeline.PreStageConfigMapSecretNames != "" {
			err = json.Unmarshal([]byte(dbPipeline.PreStageConfigMapSecretNames), &preStageConfigmapSecrets)
			if err != nil {
				impl.logger.Error(err)
				return nil, err
			}
		}
		if dbPipeline.PostStageConfigMapSecretNames != "" {
			err = json.Unmarshal([]byte(dbPipeline.PostStageConfigMapSecretNames), &postStageConfigmapSecrets)
			if err != nil {
				impl.logger.Error(err)
				return nil, err
			}
		}
		env, err := impl.envRepository.FindById(envId)
		if err != nil {
			impl.logger.Error(err)
			return nil, err
		}
		pipeline := &bean.CDPipelineConfigObject{
			Id:                            dbPipeline.Id,
			Name:                          dbPipeline.Name,
			EnvironmentId:                 dbPipeline.EnvironmentId,
			CiPipelineId:                  dbPipeline.CiPipelineId,
			TriggerType:                   dbPipeline.TriggerType,
			PreStage:                      preStage,
			PostStage:                     postStage,
			PreStageConfigMapSecretNames:  preStageConfigmapSecrets,
			PostStageConfigMapSecretNames: postStageConfigmapSecrets,
			RunPreStageInEnv:              dbPipeline.RunPreStageInEnv,
			RunPostStageInEnv:             dbPipeline.RunPostStageInEnv,
			CdArgoSetup:                   env.Cluster.CdArgoSetup,
		}
		if pipelineStages, ok := pipelineIdAndPrePostStageMapping[dbPipeline.Id]; ok {
			pipeline.PreDeployStage = pipelineStages[0]
			pipeline.PostDeployStage = pipelineStages[1]
		}
		pipelines = append(pipelines, pipeline)
	}
	cdPipelines = &bean.CdPipelines{
		AppId:     appId,
		Pipelines: pipelines,
	}
	return cdPipelines, nil
}

func (impl CiCdPipelineOrchestratorImpl) GetByEnvOverrideId(envOverrideId int) (*bean.CdPipelines, error) {
	dbPipelines, err := impl.pipelineRepository.GetByEnvOverrideId(envOverrideId)
	if err != nil {
		impl.logger.Errorw("error in fetching cdPipeline", "envOverrideId", envOverrideId, "err", err)
	}
	var pipelines []*bean.CDPipelineConfigObject
	for _, dbPipeline := range dbPipelines {
		pipeline := &bean.CDPipelineConfigObject{
			Id:            dbPipeline.Id,
			Name:          dbPipeline.Name,
			EnvironmentId: dbPipeline.EnvironmentId,
			CiPipelineId:  dbPipeline.CiPipelineId,
			TriggerType:   dbPipeline.TriggerType,
		}
		pipelines = append(pipelines, pipeline)
	}
	cdPipelines := &bean.CdPipelines{
		//AppId:     appId,
		Pipelines: pipelines,
	}
	return cdPipelines, nil
}

func (impl CiCdPipelineOrchestratorImpl) createDockerRepoIfNeeded(dockerRegistryId, dockerRepository string) error {
	dockerArtifactStore, err := impl.dockerArtifactStoreRepository.FindOne(dockerRegistryId)
	if err != nil {
		impl.logger.Errorw("error in fetching DockerRegistry  for update", "err", err, "registry", dockerRegistryId)
		return err
	}
	if dockerArtifactStore.RegistryType == dockerRegistryRepository.REGISTRYTYPE_ECR {
		err := impl.CreateEcrRepo(dockerRepository, dockerArtifactStore.AWSRegion, dockerArtifactStore.AWSAccessKeyId, dockerArtifactStore.AWSSecretAccessKey)
		if err != nil {
			impl.logger.Errorw("ecr repo creation failed while updating ci template", "err", err, "repo", dockerRepository)
			return err
		}
	}
	return nil
}
func (impl CiCdPipelineOrchestratorImpl) CreateEcrRepo(dockerRepository, AWSRegion, AWSAccessKeyId, AWSSecretAccessKey string) error {
	impl.logger.Debugw("attempting ecr repo creation ", "repo", dockerRepository)
	if !impl.ciConfig.TryCreatingEcrRepo {
		impl.logger.Warnw("not creating ecr repo, set TRY_CREATING_ECR_REPO flag true to enable")
		return nil
	}
	err := util.CreateEcrRepo(dockerRepository, AWSRegion, AWSAccessKeyId, AWSSecretAccessKey)
	if err != nil {
		if errors1.IsAlreadyExists(err) {
			impl.logger.Warnw("this repo already exists!!, skipping repo creation", "repo", dockerRepository)
		} else {
			impl.logger.Errorw("ecr repo creation failed, it might be due to authorization or any other external "+
				"dependency. please create repo manually before triggering ci", "repo", dockerRepository, "err", err)
			return err
		}
	}
	return nil
}
