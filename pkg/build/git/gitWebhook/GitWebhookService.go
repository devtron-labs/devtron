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

package gitWebhook

import (
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/sql/constants"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/build/git/gitWebhook/repository"
	"github.com/devtron-labs/devtron/pkg/build/trigger"
	"go.uber.org/zap"
)

type GitWebhookService interface {
	HandleGitWebhook(gitWebhookRequest gitSensor.CiPipelineMaterial) (int, error)
}

type GitWebhookServiceImpl struct {
	logger               *zap.SugaredLogger
	gitWebhookRepository repository.GitWebhookRepository
	ciHandlerService     trigger.HandlerService
}

func NewGitWebhookServiceImpl(Logger *zap.SugaredLogger, gitWebhookRepository repository.GitWebhookRepository,
	ciHandlerService trigger.HandlerService) *GitWebhookServiceImpl {
	return &GitWebhookServiceImpl{
		logger:               Logger,
		gitWebhookRepository: gitWebhookRepository,
		ciHandlerService:     ciHandlerService,
	}
}

func (impl *GitWebhookServiceImpl) HandleGitWebhook(gitWebhookRequest gitSensor.CiPipelineMaterial) (int, error) {
	ciPipelineMaterial := bean.CiPipelineMaterial{
		Id:            gitWebhookRequest.Id,
		GitMaterialId: gitWebhookRequest.GitMaterialId,
		Type:          string(gitWebhookRequest.Type),
		Value:         gitWebhookRequest.Value,
		Active:        gitWebhookRequest.Active,
		GitCommit: pipelineConfig.GitCommit{
			Commit:  gitWebhookRequest.GitCommit.Commit,
			Author:  gitWebhookRequest.GitCommit.Author,
			Date:    gitWebhookRequest.GitCommit.Date,
			Message: gitWebhookRequest.GitCommit.Message,
			Changes: gitWebhookRequest.GitCommit.Changes,
		},
	}

	if string(gitWebhookRequest.Type) == string(constants.SOURCE_TYPE_WEBHOOK) {
		webhookData := gitWebhookRequest.GitCommit.WebhookData
		ciPipelineMaterial.GitCommit.WebhookData = pipelineConfig.WebhookData{
			Id:              webhookData.Id,
			EventActionType: webhookData.EventActionType,
			Data:            webhookData.Data,
		}
	}

	resp, err := impl.ciHandlerService.HandleCIWebhook(bean.GitCiTriggerRequest{
		CiPipelineMaterial:        ciPipelineMaterial,
		TriggeredBy:               bean2.SYSTEM_USER_ID, // Automatic trigger, system user
		ExtraEnvironmentVariables: gitWebhookRequest.ExtraEnvironmentVariables,
	})
	if err != nil {
		impl.logger.Errorw("failed HandleCIWebhook", "err", err)
		return 0, err
	}
	return resp, nil
}
