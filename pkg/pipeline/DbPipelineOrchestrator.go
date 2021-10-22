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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DbPipelineOrchestrator interface {
	CreateApp(createRequest *bean.CreateAppDTO) (*bean.CreateAppDTO, error)
	DeleteApp(appId int, userId int32) error
	CreateMaterials(createMaterialRequest *bean.CreateMaterialDTO) (*bean.CreateMaterialDTO, error)
	UpdateMaterial(updateMaterialRequest *bean.UpdateMaterialDTO) (*bean.UpdateMaterialDTO, error)
	CreateCiConf(createRequest *bean.CiConfigRequest, templateId int) (*bean.CiConfigRequest, error)
	CreateCDPipelines(pipelineRequest *bean.CDPipelineConfigObject, appId int, userId int32, tx *pg.Tx) (pipelineId int, err error)
	UpdateCDPipeline(pipelineRequest *bean.CDPipelineConfigObject, userId int32, tx *pg.Tx) (err error)
	DeleteCiPipeline(pipeline *pipelineConfig.CiPipeline, userId int32, tx *pg.Tx) error
	DeleteCdPipeline(pipelineId int, tx *pg.Tx) error
	PatchMaterialValue(createRequest *bean.CiPipeline, userId int32) (*bean.CiPipeline, error)
	PipelineExists(name string) (bool, error)
	GetCdPipelinesForApp(appId int) (cdPipelines *bean.CdPipelines, err error)
	GetCdPipelinesForAppAndEnv(appId int, envId int) (cdPipelines *bean.CdPipelines, err error)
	GetByEnvOverrideId(envOverrideId int) (*bean.CdPipelines, error)
}

type DbPipelineOrchestratorImpl struct {
	appRepository                pipelineConfig.AppRepository
	logger                       *zap.SugaredLogger
	materialRepository           pipelineConfig.MaterialRepository
	pipelineRepository           pipelineConfig.PipelineRepository
	ciPipelineRepository         pipelineConfig.CiPipelineRepository
	CiPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository
	GitSensorClient              gitSensor.GitSensorClient
	ciConfig                     *CiConfig
	appWorkflowRepository        appWorkflow.AppWorkflowRepository
	envRepository                cluster.EnvironmentRepository
	attributesService            attributes.AttributesService
	appListingRepository         repository.AppListingRepository
	appLabelsService             app.AppLabelService
}

func NewDbPipelineOrchestrator(
	pipelineGroupRepository pipelineConfig.AppRepository,
	logger *zap.SugaredLogger,
	materialRepository pipelineConfig.MaterialRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	CiPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	GitSensorClient gitSensor.GitSensorClient, ciConfig *CiConfig,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	envRepository cluster.EnvironmentRepository,
	attributesService attributes.AttributesService,
	appListingRepository repository.AppListingRepository,
	appLabelsService app.AppLabelService,
) *DbPipelineOrchestratorImpl {

	return &DbPipelineOrchestratorImpl{
		appRepository:                pipelineGroupRepository,
		logger:                       logger,
		materialRepository:           materialRepository,
		pipelineRepository:           pipelineRepository,
		ciPipelineRepository:         ciPipelineRepository,
		CiPipelineMaterialRepository: CiPipelineMaterialRepository,
		GitSensorClient:              GitSensorClient,
		ciConfig:                     ciConfig,
		appWorkflowRepository:        appWorkflowRepository,
		envRepository:                envRepository,
		attributesService:            attributesService,
		appListingRepository:         appListingRepository,
		appLabelsService:             appLabelsService,
	}
}

const BEFORE_DOCKER_BUILD string = "BEFORE_DOCKER_BUILD"
const AFTER_DOCKER_BUILD string = "AFTER_DOCKER_BUILD"

func (impl DbPipelineOrchestratorImpl) PatchMaterialValue(createRequest *bean.CiPipeline, userId int32) (*bean.CiPipeline, error) {
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
		Version:          createRequest.Version,
		Id:               createRequest.Id,
		DockerArgs:       string(argByte),
		Active:           createRequest.Active,
		IsManual:         createRequest.IsManual,
		IsExternal:       createRequest.IsExternal,
		Deleted:          createRequest.Deleted,
		ParentCiPipeline: createRequest.ParentCiPipeline,
		ScanEnabled:      createRequest.ScanEnabled,
		AuditLog:         models.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now()},
	}
	err = impl.ciPipelineRepository.Update(ciPipelineObject, tx)
	if err != nil {
		return nil, err
	}

	ciPipelineScripts, err := impl.ciPipelineRepository.FindCiScriptsByCiPipelineId(createRequest.Id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("failed to fetch ciPipelineScripts", "err", err)
		return nil, err
	}

	existingCiScriptMap := make(map[int]bool)
	existingCiScriptModelMap := make(map[int]*pipelineConfig.CiPipelineScript)
	for _, script := range ciPipelineScripts {
		existingCiScriptMap[script.Id] = true
		existingCiScriptModelMap[script.Id] = script
	}

	err = impl.patchCiScripts(userId, createRequest, existingCiScriptMap, existingCiScriptModelMap, tx)
	if err != nil {
		impl.logger.Errorw("error while patching ci scripts", "err", err)
		return nil, err
	}

	var materials []*pipelineConfig.CiPipelineMaterial
	var materialsAdd []*pipelineConfig.CiPipelineMaterial
	var materialsUpdate []*pipelineConfig.CiPipelineMaterial
	for _, material := range createRequest.CiMaterial {
		pipelineMaterial := &pipelineConfig.CiPipelineMaterial{
			Id:            material.Id,
			Value:         material.Source.Value,
			Type:          material.Source.Type,
			Active:        createRequest.Active,
			GitMaterialId: material.GitMaterialId,
			AuditLog:      models.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now()},
		}
		if material.Id == 0 {
			pipelineMaterial.CiPipelineId = createRequest.Id
			pipelineMaterial.CreatedBy = userId
			pipelineMaterial.CreatedOn = time.Now()
			materialsAdd = append(materialsAdd, pipelineMaterial)
		} else {
			materialsUpdate = append(materialsUpdate, pipelineMaterial)
		}
	}
	if len(materialsAdd) > 0 {
		err = impl.CiPipelineMaterialRepository.Save(tx, materialsAdd...)
		if err != nil {
			return nil, err
		}
	}
	err = impl.CiPipelineMaterialRepository.Update(tx, materialsUpdate...)
	if err != nil {
		return nil, err
	}
	materials = append(materials, materialsAdd...)
	materials = append(materials, materialsUpdate...)

	if ciPipelineObject.IsExternal {
		createRequest, err = impl.updateExternalCiDetails(createRequest, userId, tx)
		if err != nil {
			impl.logger.Errorw("err", "err", err)
			return nil, err
		}
	} else {
		err = impl.addPipelineMaterialInGitSensor(materials)
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
			AuditLog:         models.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now()},
		}
		err = impl.ciPipelineRepository.Update(ciPipelineObject, tx)
		if err != nil {
			impl.logger.Errorw("err", "err", err)
			return nil, err
		}
	}

	if len(childrenCiPipelineIds) > 0 {

		ciPipelineMaterials, err := impl.CiPipelineMaterialRepository.FindByCiPipelineIdsIn(childrenCiPipelineIds)
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
					Id:       ciPipelineMaterial.Id,
					Value:    parentMaterial.Source.Value,
					Active:   createRequest.Active,
					AuditLog: models.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now()},
				}
				linkedMaterials = append(linkedMaterials, pipelineMaterial)
			} else {
				impl.logger.Errorw("material not fount in patent", "gitMaterialId", ciPipelineMaterial.GitMaterialId)
				return nil, fmt.Errorf("error while updating linked pipeline")
			}
		}
		err = impl.CiPipelineMaterialRepository.Update(tx, linkedMaterials...)
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

func (impl DbPipelineOrchestratorImpl) patchCiScripts(userId int32, pipeline *bean.CiPipeline, existingCiScriptMap map[int]bool, existingCiScriptModelMap map[int]*pipelineConfig.CiPipelineScript, tx *pg.Tx) error {
	for _, ciScript := range pipeline.BeforeDockerBuildScripts {
		ciPipelineScript := impl.buildCiPipelineScript(userId, ciScript, BEFORE_DOCKER_BUILD, pipeline)
		if _, ok := existingCiScriptMap[ciScript.Id]; ok { // Update
			err := impl.ciPipelineRepository.UpdateCiPipelineScript(ciPipelineScript, tx)
			if err != nil {
				impl.logger.Errorw("error while updating script", "err", err)
				return err
			}
			existingCiScriptMap[ciScript.Id] = false
		} else {
			err := impl.ciPipelineRepository.SaveCiPipelineScript(ciPipelineScript, tx) // Create
			if err != nil {
				impl.logger.Errorw("error while creating script", "err", err)
				return err
			}
		}
	}

	for _, ciScript := range pipeline.AfterDockerBuildScripts {
		ciPipelineScript := impl.buildCiPipelineScript(userId, ciScript, AFTER_DOCKER_BUILD, pipeline)
		if _, ok := existingCiScriptMap[ciScript.Id]; ok { // Update
			err := impl.ciPipelineRepository.UpdateCiPipelineScript(ciPipelineScript, tx)
			if err != nil {
				impl.logger.Errorw("error while updating script", "err", err)
				return err
			}
			existingCiScriptMap[ciScript.Id] = false
		} else {
			err := impl.ciPipelineRepository.SaveCiPipelineScript(ciPipelineScript, tx) // Create
			if err != nil {
				impl.logger.Errorw("error while creating script", "err", err)
				return err
			}
		}
	}

	for k, v := range existingCiScriptMap {
		if v { // Delete
			script := existingCiScriptModelMap[k]
			script.Active = false
			err := impl.ciPipelineRepository.UpdateCiPipelineScript(script, tx)
			if err != nil {
				impl.logger.Errorw("error while deleting script", "err", err)
				return err
			}
		}
	}

	return nil
}

func (impl DbPipelineOrchestratorImpl) DeleteCiPipeline(pipeline *pipelineConfig.CiPipeline, userId int32, tx *pg.Tx) error {
	p := &pipelineConfig.CiPipeline{
		Id:       pipeline.Id,
		Deleted:  true,
		AuditLog: models.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now()},
	}
	err := impl.ciPipelineRepository.Update(p, tx)
	if err != nil {
		return err
	}
	var materials []*pipelineConfig.CiPipelineMaterial
	for _, material := range pipeline.CiPipelineMaterials {
		pipelineMaterial := &pipelineConfig.CiPipelineMaterial{
			Id:       material.Id,
			Active:   false,
			AuditLog: models.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now()},
		}
		materials = append(materials, pipelineMaterial)
	}

	rows, err := impl.deleteExternalCiDetails(p, userId, tx)
	if err != nil {
		impl.logger.Errorw("err", err)
		return err
	}

	if rows == 0 {
		err = impl.addPipelineMaterialInGitSensor(materials)
		if err != nil {
			impl.logger.Errorf("error in saving pipelineMaterials in git sensor", "materials", materials, "err", err)
			return err
		}
	}
	err = impl.CiPipelineMaterialRepository.Update(tx, materials...)
	return err
}

func (impl DbPipelineOrchestratorImpl) CreateCiConf(createRequest *bean.CiConfigRequest, templateId int) (*bean.CiConfigRequest, error) {
	//save pipeline in db start
	for _, ciPipeline := range createRequest.CiPipelines {
		scriptNames := make(map[string]bool)
		for _, s := range ciPipeline.BeforeDockerBuildScripts {
			if _, ok := scriptNames[s.Name]; ok {
				impl.logger.Errorw("duplicate script names")
				return nil, errors.New("duplicate script names")
			}
			scriptNames[s.Name] = true
		}
		for _, s := range ciPipeline.AfterDockerBuildScripts {
			if _, ok := scriptNames[s.Name]; ok {
				impl.logger.Errorw("duplicate script names")
				return nil, errors.New("duplicate script names")
			}
			scriptNames[s.Name] = true
		}

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
			AppId:            createRequest.AppId,
			IsManual:         ciPipeline.IsManual,
			IsExternal:       ciPipeline.IsExternal,
			CiTemplateId:     templateId,
			Version:          ciPipeline.Version,
			Name:             ciPipeline.Name,
			ParentCiPipeline: ciPipeline.ParentCiPipeline,
			DockerArgs:       string(argByte),
			Active:           true,
			Deleted:          false,
			ScanEnabled:      createRequest.ScanEnabled,
			AuditLog:         models.AuditLog{UpdatedBy: createRequest.UserId, CreatedBy: createRequest.UserId, UpdatedOn: time.Now(), CreatedOn: time.Now()},
		}
		err = impl.ciPipelineRepository.Save(ciPipelineObject, tx)
		ciPipeline.Id = ciPipelineObject.Id
		if err != nil {
			impl.logger.Errorw("error in saving pipeline", "ciPipelineObject", ciPipelineObject, "err", err)
			return nil, err
		}

		for i, ciScript := range ciPipeline.BeforeDockerBuildScripts {
			ciPipelineScript := impl.buildCiPipelineScript(createRequest.UserId, ciScript, BEFORE_DOCKER_BUILD, ciPipeline)
			err = impl.ciPipelineRepository.SaveCiPipelineScript(ciPipelineScript, tx)
			if err != nil {
				impl.logger.Errorw("error while saving ci script", "err", err)
				return nil, err
			}
			ciPipeline.BeforeDockerBuildScripts[i].Id = ciPipelineScript.Id
		}

		for i, ciScript := range ciPipeline.AfterDockerBuildScripts {
			ciPipelineScript := impl.buildCiPipelineScript(createRequest.UserId, ciScript, AFTER_DOCKER_BUILD, ciPipeline)
			err = impl.ciPipelineRepository.SaveCiPipelineScript(ciPipelineScript, tx)
			if err != nil {
				impl.logger.Errorw("error while saving ci script", "err", err)
				return nil, err
			}
			ciPipeline.AfterDockerBuildScripts[i].Id = ciPipelineScript.Id
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
				AuditLog:      models.AuditLog{UpdatedBy: createRequest.UserId, CreatedBy: createRequest.UserId, UpdatedOn: time.Now(), CreatedOn: time.Now()},
			}
			pipelineMaterials = append(pipelineMaterials, material)
		}
		err = impl.CiPipelineMaterialRepository.Save(tx, pipelineMaterials...)
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
			ciPipeline, err = impl.saveExternalCiDetails(ciPipeline, createRequest, tx)
			if err != nil {
				impl.logger.Errorw("err", err)
				return nil, err
			}
		} else {
			//save pipeline in db end
			err = impl.addPipelineMaterialInGitSensor(pipelineMaterials)
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
				AuditLog:      models.AuditLog{CreatedBy: createRequest.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: createRequest.UserId},
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
		for _, r := range ciPipeline.CiMaterial {
			ciMaterial, err := impl.CiPipelineMaterialRepository.GetById(r.Id)
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

func (impl DbPipelineOrchestratorImpl) buildCiPipelineScript(userId int32, ciScript *bean.CiScript, scriptStage string, ciPipeline *bean.CiPipeline) *pipelineConfig.CiPipelineScript {
	ciPipelineScript := &pipelineConfig.CiPipelineScript{
		Name:           ciScript.Name,
		Index:          ciScript.Index,
		CiPipelineId:   ciPipeline.Id,
		Script:         ciScript.Script,
		Stage:          scriptStage,
		Active:         true,
		OutputLocation: ciScript.OutputLocation,
		AuditLog: models.AuditLog{
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

func (impl DbPipelineOrchestratorImpl) generateApiKey(ciPipelineId int, ciPipelineName string, secret string) (prefix, sha string) {
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

func (impl DbPipelineOrchestratorImpl) generateExternalCiPayload(ciPipeline *bean.CiPipeline, externalCiPipeline *pipelineConfig.ExternalCiPipeline, keyPrefix string, apiKey string) *bean.CiPipeline {
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

func (impl DbPipelineOrchestratorImpl) saveExternalCiDetails(ciPipeline *bean.CiPipeline, createRequest *bean.CiConfigRequest, tx *pg.Tx) (*bean.CiPipeline, error) {
	var err error
	if ciPipeline.ParentCiPipeline == 0 {
		keyPrefix, apiKey := impl.generateApiKey(ciPipeline.Id, ciPipeline.Name, impl.ciConfig.ExternalCiApiSecret)
		externalCiPipeline := &pipelineConfig.ExternalCiPipeline{
			CiPipelineId: ciPipeline.Id,
			Active:       true,
			AuditLog:     models.AuditLog{UpdatedBy: createRequest.UserId, CreatedBy: createRequest.UserId, UpdatedOn: time.Now(), CreatedOn: time.Now()},
			AccessToken:  apiKey,
		}
		externalCiPipeline, err = impl.ciPipelineRepository.SaveExternalCi(externalCiPipeline, tx)
		ciPipeline = impl.generateExternalCiPayload(ciPipeline, externalCiPipeline, keyPrefix, apiKey)
	} else {
		externalCiPipeline := &pipelineConfig.ExternalCiPipeline{
			CiPipelineId: ciPipeline.Id,
			Active:       true,
			AuditLog:     models.AuditLog{UpdatedBy: createRequest.UserId, CreatedBy: createRequest.UserId, UpdatedOn: time.Now(), CreatedOn: time.Now()},
			AccessToken:  "",
		}
		externalCiPipeline, err = impl.ciPipelineRepository.SaveExternalCi(externalCiPipeline, tx)
	}
	return ciPipeline, err
}

func (impl DbPipelineOrchestratorImpl) updateExternalCiDetails(ciPipeline *bean.CiPipeline, userId int32, tx *pg.Tx) (*bean.CiPipeline, error) {
	var err error
	if ciPipeline.ParentCiPipeline == 0 {
		keyPrefix, apiKey := impl.generateApiKey(ciPipeline.Id, ciPipeline.Name, impl.ciConfig.ExternalCiApiSecret)
		externalCiPipeline := &pipelineConfig.ExternalCiPipeline{
			CiPipelineId: ciPipeline.Id,
			Active:       true,
			AuditLog:     models.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now()},
			AccessToken:  apiKey,
		}
		externalCiPipeline, _, err = impl.ciPipelineRepository.UpdateExternalCi(externalCiPipeline, tx)
		ciPipeline = impl.generateExternalCiPayload(ciPipeline, externalCiPipeline, keyPrefix, apiKey)
	} else {
		externalCiPipeline := &pipelineConfig.ExternalCiPipeline{
			CiPipelineId: ciPipeline.Id,
			Active:       true,
			AuditLog:     models.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now()},
			AccessToken:  "",
		}
		externalCiPipeline, _, err = impl.ciPipelineRepository.UpdateExternalCi(externalCiPipeline, tx)
	}
	return ciPipeline, err
}

func (impl DbPipelineOrchestratorImpl) deleteExternalCiDetails(ciPipeline *pipelineConfig.CiPipeline, userId int32, tx *pg.Tx) (int, error) {
	externalCiPipeline := &pipelineConfig.ExternalCiPipeline{
		CiPipelineId: ciPipeline.Id,
		Active:       false,
		AuditLog:     models.AuditLog{UpdatedBy: userId, UpdatedOn: time.Now()},
	}
	externalCiPipeline, rows, err := impl.ciPipelineRepository.UpdateExternalCi(externalCiPipeline, tx)
	return rows, err
}

func (impl DbPipelineOrchestratorImpl) addPipelineMaterialInGitSensor(pipelineMaterials []*pipelineConfig.CiPipelineMaterial) error {
	var materials []*gitSensor.CiPipelineMaterial
	for _, ciPipelineMaterial := range pipelineMaterials {
		material := &gitSensor.CiPipelineMaterial{
			Id:            ciPipelineMaterial.Id,
			Active:        ciPipelineMaterial.Active,
			Value:         ciPipelineMaterial.Value,
			GitMaterialId: ciPipelineMaterial.GitMaterialId,
			Type:          gitSensor.SourceType(ciPipelineMaterial.Type),
		}
		materials = append(materials, material)
	}

	_, err := impl.GitSensorClient.SavePipelineMaterial(materials)
	return err
}

func (impl DbPipelineOrchestratorImpl) CreateApp(createRequest *bean.CreateAppDTO) (*bean.CreateAppDTO, error) {
	dbConnection := impl.appRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	app, err := impl.createAppGroup(createRequest.AppName, createRequest.UserId, createRequest.TeamId, tx)
	if err != nil {
		return nil, err
	}
	// create labels and tags with app
	if app.Active && len(createRequest.AppLabels) > 0 {
		for _, label := range createRequest.AppLabels {
			request := &bean.AppLabelDto{
				AppId:  app.Id,
				Key:    label.Key,
				Value:  label.Value,
				UserId: createRequest.UserId,
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
	return createRequest, nil
}

func (impl DbPipelineOrchestratorImpl) DeleteApp(appId int, userId int32) error {
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
	app.Active = false
	app.UpdatedOn = time.Now()
	app.UpdatedBy = userId
	err = impl.appRepository.Update(app)
	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return err
	}
	return nil
}

func (impl DbPipelineOrchestratorImpl) CreateMaterials(createMaterialRequest *bean.CreateMaterialDTO) (*bean.CreateMaterialDTO, error) {
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

func (impl DbPipelineOrchestratorImpl) UpdateMaterial(updateMaterialDTO *bean.UpdateMaterialDTO) (*bean.UpdateMaterialDTO, error) {
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

func (impl DbPipelineOrchestratorImpl) updateRepositoryToGitSensor(material *pipelineConfig.GitMaterial) error {
	sensorMaterial := &gitSensor.GitMaterial{
		Name:             material.Name,
		Url:              material.Url,
		Id:               material.Id,
		GitProviderId:    material.GitProviderId,
		CheckoutLocation: material.CheckoutPath,
		Deleted:          !material.Active,
		FetchSubmodules:  material.FetchSubmodules,
	}
	_, err := impl.GitSensorClient.UpdateRepo(sensorMaterial)
	return err
}

func (impl DbPipelineOrchestratorImpl) addRepositoryToGitSensor(materials []*bean.GitMaterial) error {
	var sensorMaterials []*gitSensor.GitMaterial
	for _, material := range materials {
		sensorMaterial := &gitSensor.GitMaterial{
			Name:            material.Name,
			Url:             material.Url,
			Id:              material.Id,
			GitProviderId:   material.GitProviderId,
			Deleted:         false,
			FetchSubmodules: material.FetchSubmodules,
		}
		sensorMaterials = append(sensorMaterials, sensorMaterial)
	}
	_, err := impl.GitSensorClient.AddRepo(sensorMaterials)
	return err
}

//FIXME: not thread safe
func (impl DbPipelineOrchestratorImpl) createAppGroup(name string, userId int32, teamId int, tx *pg.Tx) (*pipelineConfig.App, error) {
	app, err := impl.appRepository.FindActiveByName(name)
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	if app != nil && app.Id > 0 {
		impl.logger.Warnw("app already exists", "name", name)
		err = &util.ApiError{
			Code:            constants.AppAlreadyExists.Code,
			InternalMessage: "app already exists",
			UserMessage:     constants.AppAlreadyExists.UserMessage(name),
		}
		return nil, err
	}
	pg := &pipelineConfig.App{
		Active:   true,
		AppName:  name,
		TeamId:   teamId,
		AuditLog: models.AuditLog{UpdatedBy: userId, CreatedBy: userId, UpdatedOn: time.Now(), CreatedOn: time.Now()},
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

func (impl DbPipelineOrchestratorImpl) validateCheckoutPathsForMultiGit(allPaths map[int]string) error {
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

func (impl DbPipelineOrchestratorImpl) updateMaterial(updateMaterialDTO *bean.UpdateMaterialDTO) (*pipelineConfig.GitMaterial, error) {
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
	currentMaterial.AuditLog = models.AuditLog{UpdatedBy: updateMaterialDTO.UserId, CreatedBy: currentMaterial.CreatedBy, UpdatedOn: time.Now(), CreatedOn: currentMaterial.CreatedOn}

	err = impl.materialRepository.UpdateMaterial(currentMaterial)
	if err != nil {
		impl.logger.Errorw("error in updating material", "material", currentMaterial, "err", err)
		return nil, err
	}
	return currentMaterial, nil
}

func (impl DbPipelineOrchestratorImpl) createMaterial(inputMaterial *bean.GitMaterial, appId int, userId int32) (*pipelineConfig.GitMaterial, error) {
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
		AuditLog:        models.AuditLog{UpdatedBy: userId, CreatedBy: userId, UpdatedOn: time.Now(), CreatedOn: time.Now()},
	}
	err := impl.materialRepository.SaveMaterial(material)
	if err != nil {
		impl.logger.Errorw("error in saving material", "material", material, "err", err)
		return nil, err
	}
	return material, err
}

func (impl DbPipelineOrchestratorImpl) CreateCDPipelines(pipelineRequest *bean.CDPipelineConfigObject, appId int, userId int32, tx *pg.Tx) (pipelineId int, err error) {
	preStageConfig := ""
	preTriggerType := pipelineConfig.TriggerType("")
	if len(pipelineRequest.PreStage.Config) > 0 {
		preStageConfig = pipelineRequest.PreStage.Config
		preTriggerType = pipelineRequest.PreStage.TriggerType
	}

	postStageConfig := ""
	postTriggerType := pipelineConfig.TriggerType("")
	if len(pipelineRequest.PostStage.Config) > 0 {
		postStageConfig = pipelineRequest.PostStage.Config
		postTriggerType = pipelineRequest.PostStage.TriggerType
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
		AuditLog:                      models.AuditLog{UpdatedBy: userId, CreatedBy: userId, UpdatedOn: time.Now(), CreatedOn: time.Now()},
	}
	err = impl.pipelineRepository.Save([]*pipelineConfig.Pipeline{pipeline}, tx)
	return pipeline.Id, err
}

func (impl DbPipelineOrchestratorImpl) UpdateCDPipeline(pipelineRequest *bean.CDPipelineConfigObject, userId int32, tx *pg.Tx) (err error) {
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

	postStageConfig := ""
	postTriggerType := pipelineConfig.TriggerType("")
	if len(pipelineRequest.PostStage.Config) > 0 {
		postStageConfig = pipelineRequest.PostStage.Config
		postTriggerType = pipelineRequest.PostStage.TriggerType
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
	pipeline.PreStageConfig = preStageConfig
	pipeline.PostStageConfig = postStageConfig
	pipeline.PreTriggerType = preTriggerType
	pipeline.PostTriggerType = postTriggerType
	pipeline.PreStageConfigMapSecretNames = string(preStageConfigMapSecretNames)
	pipeline.PostStageConfigMapSecretNames = string(postStageConfigMapSecretNames)
	pipeline.RunPreStageInEnv = pipelineRequest.RunPreStageInEnv
	pipeline.RunPostStageInEnv = pipelineRequest.RunPostStageInEnv
	pipeline.UpdatedBy = userId
	pipeline.UpdatedOn = time.Now()
	err = impl.pipelineRepository.Update(pipeline, tx)
	if err != nil {
		return err
	}
	return err
}

func (impl DbPipelineOrchestratorImpl) DeleteCdPipeline(pipelineId int, tx *pg.Tx) error {
	return impl.pipelineRepository.Delete(pipelineId, tx)
}

func (impl DbPipelineOrchestratorImpl) PipelineExists(name string) (bool, error) {
	return impl.pipelineRepository.PipelineExists(name)
}

func (impl DbPipelineOrchestratorImpl) GetCdPipelinesForApp(appId int) (cdPipelines *bean.CdPipelines, err error) {
	dbPipelines, err := impl.pipelineRepository.FindActiveByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching cdPipeline", "appId", appId, "err", err)
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
			CiPipelineId:                  dbPipeline.CiPipelineId,
			TriggerType:                   dbPipeline.TriggerType,
			PreStage:                      preStage,
			PostStage:                     postStage,
			RunPreStageInEnv:              dbPipeline.RunPreStageInEnv,
			RunPostStageInEnv:             dbPipeline.RunPostStageInEnv,
			PreStageConfigMapSecretNames:  preStageConfigmapSecrets,
			PostStageConfigMapSecretNames: postStageConfigmapSecrets,
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

func (impl DbPipelineOrchestratorImpl) GetCdPipelinesForAppAndEnv(appId int, envId int) (cdPipelines *bean.CdPipelines, err error) {
	dbPipelines, err := impl.pipelineRepository.FindActiveByAppIdAndEnvironmentId(appId, envId)
	if err != nil {
		impl.logger.Errorw("error in fetching cdPipeline", "appId", appId, "err", err)
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
		pipelines = append(pipelines, pipeline)
	}
	cdPipelines = &bean.CdPipelines{
		AppId:     appId,
		Pipelines: pipelines,
	}
	return cdPipelines, nil
}

func (impl DbPipelineOrchestratorImpl) GetByEnvOverrideId(envOverrideId int) (*bean.CdPipelines, error) {
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
