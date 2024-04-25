package scoop

import (
	"context"
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/autoRemediation"
	"github.com/devtron-labs/devtron/pkg/autoRemediation/repository"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

type Service interface {
	HandleInterceptedEvent(ctx context.Context, event *InterceptedEvent) error
}

type ServiceImpl struct {
	logger                       *zap.SugaredLogger
	watcherService               autoRemediation.WatcherService
	ciHandler                    pipeline.CiHandler
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository
	interceptedEventsRepository  repository.InterceptedEventsRepository
}

func NewServiceImpl(logger *zap.SugaredLogger,
	watcherService autoRemediation.WatcherService,
	ciHandler pipeline.CiHandler,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	interceptedEventsRepository repository.InterceptedEventsRepository,
) *ServiceImpl {
	return &ServiceImpl{
		logger:                       logger,
		watcherService:               watcherService,
		ciHandler:                    ciHandler,
		ciPipelineMaterialRepository: ciPipelineMaterialRepository,
		interceptedEventsRepository:  interceptedEventsRepository,
	}
}

func (impl ServiceImpl) HandleInterceptedEvent(ctx context.Context, interceptedEvent *InterceptedEvent) error {

	eventBytes, err := json.Marshal(&interceptedEvent.Event)
	if err != nil {
		return err
	}
	event := string(eventBytes)
	involvedObjBytes, err := json.Marshal(&interceptedEvent.InvolvedObject)
	if err != nil {
		return err
	}
	involvedObj := string(involvedObjBytes)
	watchersMap := make(map[int]*Watcher)
	for _, watcher := range interceptedEvent.Watchers {
		watchersMap[watcher.Id] = watcher
	}

	triggers, err := impl.watcherService.GetTriggerByWatcherIds(maps.Keys(watchersMap))
	if err != nil {
		return err
	}

	jobPipelineIds := make([]int, 0, len(triggers))
	for _, trigger := range triggers {
		jobPipelineIds = append(jobPipelineIds, trigger.Data.PipelineId)
	}

	ciPipelineMaterialMap := make(map[int][]*pipelineConfig.CiPipelineMaterial)

	ciPipelineMaterials, err := impl.ciPipelineMaterialRepository.FindByCiPipelineIdsIn(jobPipelineIds)
	if err != nil {
		impl.logger.Errorw("error in fetching ci pipeline material ")
	}

	for _, ciPipelineMaterial := range ciPipelineMaterials {
		ciPipelineId := ciPipelineMaterial.CiPipelineId
		if _, ok := ciPipelineMaterialMap[ciPipelineId]; !ok {
			ciPipelineMaterialMap[ciPipelineId] = make([]*pipelineConfig.CiPipelineMaterial, 0)
		}
		ciPipelineMaterialMap[ciPipelineId] = append(ciPipelineMaterialMap[ciPipelineId], ciPipelineMaterial)
	}

	interceptEventExecs := make([]*repository.InterceptedEventExecution, 0, len(triggers))
	for _, trigger := range triggers {
		ciMaterials := ciPipelineMaterialMap[trigger.Data.PipelineId]
		if len(ciMaterials) == 0 {
			continue
		}
		runtimeparams := bean.RuntimeParameters{
			EnvVariables: make(map[string]string),
		}
		for _, param := range trigger.Data.RuntimeParameters {
			runtimeparams.EnvVariables[param.Key] = param.Value
		}

		request := bean.CiTriggerRequest{
			PipelineId: trigger.Data.PipelineId,
			// system user
			TriggeredBy:   1,
			EnvironmentId: trigger.Data.ExecutionEnvironmentId,
			PipelineType:  "CI_BUILD",
			CiPipelineMaterial: []bean.CiPipelineMaterial{{
				Id: ciMaterials[0].Id,
				// todo: check how to fetch the git material info
			}},
			RuntimeParams: &runtimeparams,
		}

		ciWorkflowId, err := impl.ciHandler.HandleCIManual(request)
		status := repository.Progressing
		executionMessage := ""
		if err != nil {
			impl.logger.Errorw("error in trigger job ci pipeline", "triggerRequest", request, "err", err)
			executionMessage = err.Error()
			status = repository.Errored
		}

		interceptEventExec := &repository.InterceptedEventExecution{
			ClusterId:          interceptedEvent.ClusterId,
			Event:              event,
			InvolvedObject:     involvedObj,
			InterceptedAt:      interceptedEvent.InterceptedAt,
			Namespace:          interceptedEvent.Namespace,
			Message:            interceptedEvent.Message,
			MessageType:        interceptedEvent.MessageType,
			TriggerId:          trigger.Id,
			TriggerExecutionId: ciWorkflowId,
			Status:             status,
			ExecutionMessage:   executionMessage,
		}

		interceptEventExecs = append(interceptEventExecs, interceptEventExec)
	}
	tx, err := impl.interceptedEventsRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in creating transaction while saving ", "interceptedEvent", interceptedEvent, "err", err)
		return err
	}

	defer impl.interceptedEventsRepository.RollbackTx(tx)
	_, err = impl.interceptedEventsRepository.Save(interceptEventExecs, tx)
	if err != nil {
		impl.logger.Errorw("error in saving intercepted event executions ", "interceptEventExecs", interceptEventExecs, "err", err)
		return err
	}

	err = impl.interceptedEventsRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction while saving ", "interceptedEvent", interceptedEvent, "err", err)
		return err
	}
	return nil
}
