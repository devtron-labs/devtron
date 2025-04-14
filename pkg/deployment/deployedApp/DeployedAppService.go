/*
 * Copyright (c) 2024. Devtron Inc.
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

package deployedApp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	util5 "github.com/devtron-labs/common-lib/utils/k8s"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean6 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	bean5 "github.com/devtron-labs/devtron/pkg/deployment/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/deployedApp/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps"
	bean3 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/k8s"
	bean4 "github.com/devtron-labs/devtron/pkg/k8s/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DeployedAppService interface {
	StopStartApp(ctx context.Context, stopRequest *bean.StopAppRequest, userMetadata *bean6.UserMetadata) (int, error)
	RotatePods(ctx context.Context, podRotateRequest *bean.PodRotateRequest, userMetadata *bean6.UserMetadata) (*bean4.RotatePodResponse, error)
	StopStartAppV1(ctx context.Context, stopRequest *bean.StopAppRequest, userMetadata *bean6.UserMetadata) (int, error)
	HibernationPatch(ctx context.Context, appId, envId int, userMetadata *bean6.UserMetadata) (*bean.HibernationPatchResponse, error)
}

type DeployedAppServiceImpl struct {
	logger               *zap.SugaredLogger
	k8sCommonService     k8s.K8sCommonService
	cdTriggerService     devtronApps.TriggerService
	envRepository        repository.EnvironmentRepository
	pipelineRepository   pipelineConfig.PipelineRepository
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository
}

func NewDeployedAppServiceImpl(logger *zap.SugaredLogger,
	k8sCommonService k8s.K8sCommonService,
	cdTriggerService devtronApps.TriggerService,
	envRepository repository.EnvironmentRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository) *DeployedAppServiceImpl {
	return &DeployedAppServiceImpl{
		logger:               logger,
		k8sCommonService:     k8sCommonService,
		cdTriggerService:     cdTriggerService,
		envRepository:        envRepository,
		pipelineRepository:   pipelineRepository,
		cdWorkflowRepository: cdWorkflowRepository,
	}
}

func (impl *DeployedAppServiceImpl) StopStartApp(ctx context.Context, stopRequest *bean.StopAppRequest, userMetadata *bean6.UserMetadata) (int, error) {
	return impl.stopStartApp(ctx, stopRequest, userMetadata)
}

func (impl *DeployedAppServiceImpl) stopStartApp(ctx context.Context, stopRequest *bean.StopAppRequest, userMetadata *bean6.UserMetadata) (int, error) {
	pipelines, err := impl.pipelineRepository.FindActiveByAppIdAndEnvironmentId(stopRequest.AppId, stopRequest.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in fetching pipeline", "app", stopRequest.AppId, "env", stopRequest.EnvironmentId, "err", err)
		return 0, err
	}
	if len(pipelines) == 0 {
		return 0, fmt.Errorf("no pipeline found")
	}
	pipeline := pipelines[0]

	//find pipeline with default
	var pipelineIds []int
	for _, p := range pipelines {
		impl.logger.Debugw("adding pipelineId", "pipelineId", p.Id)
		pipelineIds = append(pipelineIds, p.Id)
		//FIXME
	}
	wf, err := impl.cdWorkflowRepository.FindLatestCdWorkflowByPipelineId(pipelineIds)
	if errors.Is(err, pg.ErrNoRows) {
		return 0, errors.New("no deployment history found,this app was never deployed")
	}
	if err != nil {
		impl.logger.Errorw("error in fetching latest release", "err", err)
		return 0, err
	}
	err = impl.checkForFeasibilityBeforeStartStop(ctx, stopRequest.AppId, stopRequest.EnvironmentId, userMetadata)
	if err != nil {
		impl.logger.Errorw("error in checking for feasibility before hibernating and un hibernating", "stopRequest", stopRequest, "err", err)
		return 0, err
	}
	overrideRequest := &bean2.ValuesOverrideRequest{
		PipelineId:     pipeline.Id,
		AppId:          stopRequest.AppId,
		CiArtifactId:   wf.CiArtifactId,
		UserId:         stopRequest.UserId,
		CdWorkflowType: bean2.CD_WORKFLOW_TYPE_DEPLOY,
	}
	if stopRequest.RequestType == bean.STOP {
		err = impl.setStopTemplate(stopRequest)
		if err != nil {
			impl.logger.Errorw("error in configuring stopTemplate stopStartApp", "stopRequest", stopRequest, "err", err)
			return 0, err
		}
		overrideRequest.AdditionalOverride = json.RawMessage([]byte(stopRequest.StopPatch))
		overrideRequest.DeploymentType = models.DEPLOYMENTTYPE_STOP
	} else if stopRequest.RequestType == bean.START {
		overrideRequest.DeploymentType = models.DEPLOYMENTTYPE_START
	} else {
		return 0, fmt.Errorf("unsupported operation %s", stopRequest.RequestType)
	}
	triggerContext := bean3.TriggerContext{
		Context:     ctx,
		ReferenceId: stopRequest.ReferenceId,
	}
	id, _, _, err := impl.cdTriggerService.ManualCdTrigger(triggerContext, overrideRequest, userMetadata)
	if err != nil {
		impl.logger.Errorw("error in stopping app", "err", err, "appId", stopRequest.AppId, "envId", stopRequest.EnvironmentId)
		return 0, err
	}
	return id, err
}

func (impl *DeployedAppServiceImpl) RotatePods(ctx context.Context, podRotateRequest *bean.PodRotateRequest, userMetadata *bean6.UserMetadata) (*bean4.RotatePodResponse, error) {
	impl.logger.Infow("rotate pod request", "payload", podRotateRequest)
	//extract cluster id and namespace from env id
	environmentId := podRotateRequest.EnvironmentId
	environment, err := impl.envRepository.FindById(environmentId)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching env details", "envId", environmentId, "err", err)
		return nil, err
	}
	err = impl.checkForFeasibilityBeforeStartStop(ctx, podRotateRequest.AppId, podRotateRequest.EnvironmentId, userMetadata)
	if err != nil {
		impl.logger.Errorw("error in checking for feasibility in Rotating pods", "podRotateRequest", podRotateRequest, "err", err)
		return nil, err
	}
	var resourceIdentifiers []util5.ResourceIdentifier
	for _, resourceIdentifier := range podRotateRequest.ResourceIdentifiers {
		resourceIdentifier.Namespace = environment.Namespace
		resourceIdentifiers = append(resourceIdentifiers, resourceIdentifier)
	}
	rotatePodRequest := &bean4.RotatePodRequest{
		ClusterId: environment.ClusterId,
		Resources: resourceIdentifiers,
	}
	response, err := impl.k8sCommonService.RotatePods(ctx, rotatePodRequest)
	if err != nil {
		return nil, err
	}
	//TODO KB: make entry in cd workflow runner
	return response, nil
}
func (impl *DeployedAppServiceImpl) setStopTemplate(stopRequest *bean.StopAppRequest) error {
	var stopTemplate string
	var err error
	if stopRequest.IsHibernationPatchConfigured {
		stopTemplate, err = impl.getTemplate(stopRequest)
		if err != nil {
			impl.logger.Errorw("error in getting hibernation patch configuration", "stopRequest", stopRequest, "err", err)
			return err
		}
		impl.logger.Debugw("stop template fetched from scope", "stopTemplate", stopTemplate)
	} else {
		stopTemplate, err = impl.getTemplateDefault()
		if err != nil {
			impl.logger.Errorw("error in getting hibernation patch configuration", "stopRequest", stopRequest, "err", err)
			return err
		}
	}
	stopRequest.StopPatch = stopTemplate
	return nil
}

func (impl *DeployedAppServiceImpl) getTemplateDefault() (string, error) {
	stopTemplate := bean5.DefaultStopTemplate
	return stopTemplate, nil
}
