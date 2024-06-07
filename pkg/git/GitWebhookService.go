/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package git

import (
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"go.uber.org/zap"
)

type GitWebhookService interface {
	HandleGitWebhook(gitWebhookRequest gitSensor.CiPipelineMaterial) (int, error)
}

type GitWebhookServiceImpl struct {
	logger               *zap.SugaredLogger
	ciHandler            pipeline.CiHandler
	gitWebhookRepository repository.GitWebhookRepository
}

func NewGitWebhookServiceImpl(Logger *zap.SugaredLogger, ciHandler pipeline.CiHandler, gitWebhookRepository repository.GitWebhookRepository) *GitWebhookServiceImpl {
	return &GitWebhookServiceImpl{
		logger:               Logger,
		ciHandler:            ciHandler,
		gitWebhookRepository: gitWebhookRepository,
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

	if string(gitWebhookRequest.Type) == string(pipelineConfig.SOURCE_TYPE_WEBHOOK) {
		webhookData := gitWebhookRequest.GitCommit.WebhookData
		ciPipelineMaterial.GitCommit.WebhookData = pipelineConfig.WebhookData{
			Id:              webhookData.Id,
			EventActionType: webhookData.EventActionType,
			Data:            webhookData.Data,
		}
	}

	resp, err := impl.ciHandler.HandleCIWebhook(bean.GitCiTriggerRequest{
		CiPipelineMaterial:        ciPipelineMaterial,
		TriggeredBy:               1, // Automatic trigger, userId is 1
		ExtraEnvironmentVariables: gitWebhookRequest.ExtraEnvironmentVariables,
	})
	if err != nil {
		impl.logger.Errorw("failed HandleCIWebhook", "err", err)
		return 0, err
	}
	return resp, nil
}
