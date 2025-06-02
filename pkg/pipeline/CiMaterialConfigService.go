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

package pipeline

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/constants"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/build/git/gitMaterial/read"
	"github.com/devtron-labs/devtron/pkg/build/git/gitMaterial/repository"
	"github.com/devtron-labs/devtron/pkg/build/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util/sliceUtil"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type CiMaterialConfigService interface {
	//CreateMaterialsForApp : Delegating the request to ciCdPipelineOrchestrator for Material creation
	CreateMaterialsForApp(request *bean.CreateMaterialDTO) (*bean.CreateMaterialDTO, error)
	//UpdateMaterialsForApp : Delegating the request to ciCdPipelineOrchestrator for updating Material
	UpdateMaterialsForApp(request *bean.UpdateMaterialDTO) (*bean.UpdateMaterialDTO, error)
	DeleteMaterial(request *bean.UpdateMaterialDTO) error
	//PatchCiMaterialSource : Delegating the request to ciCdPipelineOrchestrator for updating source
	PatchCiMaterialSource(ciPipeline *bean.CiMaterialPatchRequest, userId int32) (*bean.CiMaterialPatchRequest, error)
	//BulkPatchCiMaterialSource : Delegating the request to ciCdPipelineOrchestrator for bulk updating source
	BulkPatchCiMaterialSource(ciPipelines *bean.CiMaterialBulkPatchRequest, userId int32, token string, checkAppSpecificAccess func(token, action string, appId int) (bool, error)) (*bean.CiMaterialBulkPatchResponse, error)
	//GetMaterialsForAppId : Retrieve material for given appId
	GetMaterialsForAppId(appId int) []*bean.GitMaterial
}

type CiMaterialConfigServiceImpl struct {
	logger                       *zap.SugaredLogger
	materialRepo                 repository.MaterialRepository
	gitMaterialReadService       read.GitMaterialReadService
	ciTemplateService            pipeline.CiTemplateReadService
	ciCdPipelineOrchestrator     CiCdPipelineOrchestrator
	ciPipelineRepository         pipelineConfig.CiPipelineRepository
	gitMaterialHistoryService    history.GitMaterialHistoryService
	pipelineRepository           pipelineConfig.PipelineRepository
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository
	transactionManager           sql.TransactionWrapper
}

func NewCiMaterialConfigServiceImpl(
	logger *zap.SugaredLogger,
	materialRepo repository.MaterialRepository,
	ciTemplateService pipeline.CiTemplateReadService,
	ciCdPipelineOrchestrator CiCdPipelineOrchestrator,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	gitMaterialHistoryService history.GitMaterialHistoryService,
	pipelineRepository pipelineConfig.PipelineRepository,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	transactionManager sql.TransactionWrapper,
	gitMaterialReadService read.GitMaterialReadService) *CiMaterialConfigServiceImpl {

	return &CiMaterialConfigServiceImpl{
		logger:                       logger,
		materialRepo:                 materialRepo,
		ciTemplateService:            ciTemplateService,
		ciCdPipelineOrchestrator:     ciCdPipelineOrchestrator,
		ciPipelineRepository:         ciPipelineRepository,
		gitMaterialHistoryService:    gitMaterialHistoryService,
		pipelineRepository:           pipelineRepository,
		ciPipelineMaterialRepository: ciPipelineMaterialRepository,
		transactionManager:           transactionManager,
		gitMaterialReadService:       gitMaterialReadService,
	}
}

func (impl *CiMaterialConfigServiceImpl) CreateMaterialsForApp(request *bean.CreateMaterialDTO) (*bean.CreateMaterialDTO, error) {
	res, err := impl.ciCdPipelineOrchestrator.CreateMaterials(request)
	if err != nil {
		impl.logger.Errorw("error in saving create materials req", "req", request, "err", err)
	}
	return res, err
}

func (impl *CiMaterialConfigServiceImpl) UpdateMaterialsForApp(request *bean.UpdateMaterialDTO) (*bean.UpdateMaterialDTO, error) {
	res, err := impl.ciCdPipelineOrchestrator.UpdateMaterial(request)
	if err != nil {
		impl.logger.Errorw("error in updating materials req", "req", request, "err", err)
	}
	return res, err
}

func (impl *CiMaterialConfigServiceImpl) DeleteMaterial(request *bean.UpdateMaterialDTO) error {
	//finding ci pipelines for this app; if found any, will not delete git material
	pipelines, err := impl.ciPipelineRepository.FindByAppId(request.AppId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in deleting git material", "gitMaterial", request.Material, "err", err)
		return err
	}
	if len(pipelines) > 0 {
		//pipelines are present, in this case we will check if this material is used in docker config
		//if it is used, then we won't delete
		ciTemplateBean, err := impl.ciTemplateService.FindByAppId(request.AppId)
		if err != nil && err == errors.NotFoundf(err.Error()) {
			impl.logger.Errorw("err in getting docker registry", "appId", request.AppId, "err", err)
			return err
		}
		if ciTemplateBean != nil {
			ciTemplate := ciTemplateBean.CiTemplate
			if ciTemplate != nil && ciTemplate.GitMaterialId == request.Material.Id {
				return fmt.Errorf("cannot delete git material, is being used in docker config")
			}
		}
		pipelineIds := sliceUtil.NewSliceFromFuncExec(pipelines, func(dbPipeline *pipelineConfig.CiPipeline) int {
			return dbPipeline.Id
		})
		exist, err := impl.ciTemplateService.CheckIfTemplateOverrideExists(pipelineIds, request.Material.Id)
		if err != nil {
			impl.logger.Errorw("error in checking if template override exists", "pipelineIds", pipelineIds, "gitMaterialId", request.Material.Id, "err", err)
			return err
		}
		if exist {
			return util2.GetApiErrorAdapter(http.StatusBadRequest, "400", "cannot delete git material, is being used in overridden ci template", "cannot delete git material, is being used in overridden ci template")
		}
	}
	existingMaterial, err := impl.gitMaterialReadService.FindById(request.Material.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete", "gitMaterial", request.Material)
		return err
	}
	existingMaterial.UpdatedOn = time.Now()
	existingMaterial.UpdatedBy = request.UserId

	tx, err := impl.transactionManager.StartTx()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = impl.materialRepo.MarkMaterialDeleted(tx, existingMaterial)
	if err != nil {
		impl.logger.Errorw("error in deleting git material", "gitMaterial", existingMaterial)
		return err
	}

	err = impl.gitMaterialHistoryService.MarkMaterialDeletedAndCreateHistory(tx, existingMaterial)

	var materials []*pipelineConfig.CiPipelineMaterial
	for _, pipeline := range pipelines {
		materialDbObject, err := impl.ciPipelineMaterialRepository.GetByPipelineIdAndGitMaterialId(pipeline.Id, request.Material.Id)
		if err != nil {
			return err
		}
		if len(materialDbObject) == 0 {
			continue
		}
		materialDbObject[0].Active = false
		materials = append(materials, materialDbObject[0])
	}

	if len(materials) != 0 {
		err = impl.ciPipelineMaterialRepository.Update(tx, materials...)
		if err != nil {
			impl.logger.Errorw("error while updating ci pipeline material", "appId", request.AppId, "err", err)
			return err
		}
	}

	//err = impl.ciPipelineMaterialRepository.Update(tx, materials...)
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing changes in ci pipeline material", "appId", request.AppId, "err", err)
		return err
	}

	return nil
}

func (impl *CiMaterialConfigServiceImpl) PatchCiMaterialSource(ciPipeline *bean.CiMaterialPatchRequest, userId int32) (*bean.CiMaterialPatchRequest, error) {
	return impl.ciCdPipelineOrchestrator.PatchCiMaterialSource(ciPipeline, userId)
}

func (impl *CiMaterialConfigServiceImpl) BulkPatchCiMaterialSource(ciPipelines *bean.CiMaterialBulkPatchRequest, userId int32, token string, checkAppSpecificAccess func(token, action string, appId int) (bool, error)) (*bean.CiMaterialBulkPatchResponse, error) {
	response := &bean.CiMaterialBulkPatchResponse{}
	var ciPipelineMaterials []*pipelineConfig.CiPipelineMaterial
	for _, appId := range ciPipelines.AppIds {
		ciPipeline := &bean.CiMaterialValuePatchRequest{
			AppId:         appId,
			EnvironmentId: ciPipelines.EnvironmentId,
		}
		ciPipelineMaterial, err := impl.ciCdPipelineOrchestrator.PatchCiMaterialSourceValue(ciPipeline, userId, ciPipelines.Value, token, checkAppSpecificAccess)

		if err == nil {
			ciPipelineMaterial.Type = constants.SOURCE_TYPE_BRANCH_FIXED
			ciPipelineMaterials = append(ciPipelineMaterials, ciPipelineMaterial)
		}
		response.Apps = append(response.Apps, bean.CiMaterialPatchResponse{
			AppId:   appId,
			Status:  getPatchStatus(err),
			Message: getPatchMessage(err),
		})
	}
	if len(ciPipelineMaterials) == 0 {
		return response, nil
	}
	if err := impl.ciCdPipelineOrchestrator.UpdateCiPipelineMaterials(ciPipelineMaterials); err != nil {
		return nil, err
	}
	return response, nil
}

func (impl *CiMaterialConfigServiceImpl) GetMaterialsForAppId(appId int) []*bean.GitMaterial {
	materials, err := impl.gitMaterialReadService.FindByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching materials", "appId", appId, "err", err)
	}

	ciTemplateBean, err := impl.ciTemplateService.FindByAppId(appId)
	if err != nil && err != errors.NotFoundf(err.Error()) {
		impl.logger.Errorw("err in getting ci-template", "appId", appId, "err", err)
	}

	var gitMaterials []*bean.GitMaterial
	for _, material := range materials {
		gitMaterial := &bean.GitMaterial{
			Url:             material.Url,
			Name:            material.Name[strings.Index(material.Name, "-")+1:],
			Id:              material.Id,
			GitProviderId:   material.GitProviderId,
			CheckoutPath:    material.CheckoutPath,
			FetchSubmodules: material.FetchSubmodules,
			FilterPattern:   material.FilterPattern,
		}
		//check if git material is deletable or not
		if ciTemplateBean != nil {
			ciTemplate := ciTemplateBean.CiTemplate
			if ciTemplate != nil && (ciTemplate.GitMaterialId == material.Id || ciTemplate.BuildContextGitMaterialId == material.Id) {
				gitMaterial.IsUsedInCiConfig = true
			}
		}
		gitMaterials = append(gitMaterials, gitMaterial)
	}
	return gitMaterials
}
