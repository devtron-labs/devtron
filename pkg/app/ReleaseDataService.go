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

package app

import (
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/client/lens"
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"go.uber.org/zap"
)

type ReleaseDataService interface {
	TriggerEventForAllRelease(appId, environmentId int) error
	GetDeploymentMetrics(request *lens.MetricRequest) (resBody []byte, resCode *lens.StatusCode, err error)
}
type ReleaseDataServiceImpl struct {
	pipelineOverrideRepository   chartConfig.PipelineOverrideRepository
	logger                       *zap.SugaredLogger
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository
	eventClient                  client.EventClient
	lensClient                   lens.LensClient
}

func NewReleaseDataServiceImpl(
	pipelineOverrideRepository chartConfig.PipelineOverrideRepository,
	logger *zap.SugaredLogger,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	eventClient client.EventClient,
	lensClient lens.LensClient) *ReleaseDataServiceImpl {
	return &ReleaseDataServiceImpl{
		pipelineOverrideRepository:   pipelineOverrideRepository,
		logger:                       logger,
		ciPipelineMaterialRepository: ciPipelineMaterialRepository,
		eventClient:                  eventClient,
		lensClient:                   lensClient,
	}

}

func (impl *ReleaseDataServiceImpl) TriggerEventForAllRelease(appId, environmentId int) error {
	releases, err := impl.pipelineOverrideRepository.GetAllRelease(appId, environmentId)
	if err != nil {
		impl.logger.Errorw("error in getting release pipeline", "app", appId, "env", environmentId, "err", err)
		return err
	}
	var ciPipelineMaterials []*pipelineConfig.CiPipelineMaterial
	var deployments []*DeploymentEvent
	for _, release := range releases {
		deployment := &DeploymentEvent{
			ApplicationId:      release.Pipeline.AppId,
			EnvironmentId:      release.Pipeline.EnvironmentId,
			ReleaseId:          release.PipelineReleaseCounter,
			PipelineOverrideId: release.Id,
			TriggerTime:        release.CreatedOn,
			PipelineMaterials:  nil,
			CiArtifactId:       release.CiArtifactId,
		}
		if len(ciPipelineMaterials) == 0 {
			ciPipelineMaterials, err = impl.ciPipelineMaterialRepository.GetByPipelineId(release.CiArtifact.PipelineId)
			if err != nil {
				impl.logger.Errorw("error in getting pipeline materials ", "err", err)
				return err
			}
		}
		materialInfoMap, err := release.CiArtifact.ParseMaterialInfo()
		if err != nil {
			impl.logger.Errorw("error in parsing material", "err", err)
			//return err
		}
		for _, ciPipelineMaterial := range ciPipelineMaterials {
			hash := materialInfoMap[ciPipelineMaterial.GitMaterial.Url]
			pipelineMaterialInfo := &PipelineMaterialInfo{PipelineMaterialId: ciPipelineMaterial.Id, CommitHash: hash}
			deployment.PipelineMaterials = append(deployment.PipelineMaterials, pipelineMaterialInfo)
		}
		deployments = append(deployments, deployment)
	}
	for _, deploymentEvent := range deployments {
		impl.logger.Infow("triggering deployment event", "event", deploymentEvent)
		err = impl.eventClient.WriteNatsEvent(pubsub.CD_SUCCESS, deploymentEvent)
		if err != nil {
			impl.logger.Errorw("error in writing cd trigger event", "err", err)
			return err
		}
	}
	return nil
}

func (impl *ReleaseDataServiceImpl) GetDeploymentMetrics(request *lens.MetricRequest) (resBody []byte, resCode *lens.StatusCode, err error) {
	resBody, resCode, err = impl.lensClient.GetAppMetrics(request)
	return resBody, resCode, err
}
