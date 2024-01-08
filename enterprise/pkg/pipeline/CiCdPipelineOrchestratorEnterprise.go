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
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/enterprise/pkg/globalTag"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	app2 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/genericNotes"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	history3 "github.com/devtron-labs/devtron/pkg/pipeline/history"
	repository5 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"go.uber.org/zap"
)

type CiCdPipelineOrchestratorEnterpriseImpl struct {
	globalTagService globalTag.GlobalTagService
	*pipeline.CiCdPipelineOrchestratorImpl
}

func NewCiCdPipelineOrchestratorEnterpriseImpl(pipelineGroupRepository app2.AppRepository,
	logger *zap.SugaredLogger,
	materialRepository pipelineConfig.MaterialRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	GitSensorClient gitSensor.Client, ciConfig *types.CiCdConfig,
	appWorkflowRepository appWorkflow.AppWorkflowRepository,
	envRepository repository2.EnvironmentRepository,
	attributesService attributes.AttributesService,
	appListingRepository repository.AppListingRepository,
	appLabelsService app.AppCrudOperationService,
	userAuthService user.UserAuthService,
	prePostCdScriptHistoryService history3.PrePostCdScriptHistoryService,
	prePostCiScriptHistoryService history3.PrePostCiScriptHistoryService,
	pipelineStageService pipeline.PipelineStageService,
	ciTemplateOverrideRepository pipelineConfig.CiTemplateOverrideRepository,
	gitMaterialHistoryService history3.GitMaterialHistoryService,
	ciPipelineHistoryService history3.CiPipelineHistoryService,
	ciTemplateService pipeline.CiTemplateService,
	dockerArtifactStoreRepository dockerRegistryRepository.DockerArtifactStoreRepository,
	globalTagService globalTag.GlobalTagService,
	PipelineOverrideRepository chartConfig.PipelineOverrideRepository, CiArtifactRepository repository.CiArtifactRepository, manifestPushConfigRepository repository5.ManifestPushConfigRepository,
	configMapService pipeline.ConfigMapService,
	genericNoteService genericNotes.GenericNoteService,
	customTagService pipeline.CustomTagService,
	devtronResourceService devtronResource.DevtronResourceService) *CiCdPipelineOrchestratorEnterpriseImpl {
	return &CiCdPipelineOrchestratorEnterpriseImpl{
		CiCdPipelineOrchestratorImpl: pipeline.NewCiCdPipelineOrchestrator(pipelineGroupRepository, logger, materialRepository, pipelineRepository,
			ciPipelineRepository, ciPipelineMaterialRepository, GitSensorClient, ciConfig, appWorkflowRepository, envRepository, attributesService,
			appListingRepository, appLabelsService, userAuthService, prePostCdScriptHistoryService, prePostCiScriptHistoryService,
			pipelineStageService, ciTemplateOverrideRepository, gitMaterialHistoryService, ciPipelineHistoryService, ciTemplateService,
			dockerArtifactStoreRepository, PipelineOverrideRepository, CiArtifactRepository, manifestPushConfigRepository, configMapService, customTagService, genericNoteService,
			devtronResourceService),
		globalTagService: globalTagService,
	}
}

func (impl *CiCdPipelineOrchestratorEnterpriseImpl) CreateApp(createRequest *bean.CreateAppDTO) (*bean.CreateAppDTO, error) {
	// validate mandatory labels against project
	labelsMap := make(map[string]string)
	for _, label := range createRequest.AppLabels {
		labelsMap[label.Key] = label.Value
	}
	err := impl.globalTagService.ValidateMandatoryLabelsForProject(createRequest.TeamId, labelsMap)
	if err != nil {
		return nil, err
	}

	// call forward
	return impl.CiCdPipelineOrchestratorImpl.CreateApp(createRequest)
}
