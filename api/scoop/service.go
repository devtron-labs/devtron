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
	event, involvedObj, triggers, err := impl.getTriggersAndEventData(interceptedEvent)
	if err != nil {
		impl.logger.Errorw("error in getting triggers and intercepted event data", "interceptedEvent", interceptedEvent, "err", err)
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

	err = impl.saveInterceptedEvents(interceptEventExecs)
	if err != nil {
		impl.logger.Errorw("error in saving intercepted event executions", "interceptedEvent", interceptedEvent, "err", err)
	}
	return err
}

func (impl ServiceImpl) getTriggersAndEventData(interceptedEvent *InterceptedEvent) (event string, involvedObj string, triggers []*autoRemediation.Trigger, err error) {
	eventBytes, err := json.Marshal(&interceptedEvent.Event)
	if err != nil {
		return event, involvedObj, triggers, err
	}
	event = string(eventBytes)
	involvedObjBytes, err := json.Marshal(&interceptedEvent.InvolvedObject)
	if err != nil {
		return event, involvedObj, triggers, err
	}
	involvedObj = string(involvedObjBytes)
	watchersMap := make(map[int]*Watcher)
	for _, watcher := range interceptedEvent.Watchers {
		watchersMap[watcher.Id] = watcher
	}

	watcherIds := maps.Keys(watchersMap)
	triggers, err = impl.watcherService.GetTriggerByWatcherIds(watcherIds)
	if err != nil {
		impl.logger.Errorw("error in getting triggers with watcher ids", "watcherIds", watcherIds, "err", err)
	}
	return
}

func (impl ServiceImpl) saveInterceptedEvents(interceptEventExecs []*repository.InterceptedEventExecution) error {
	tx, err := impl.interceptedEventsRepository.StartTx()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			impl.logger.Debugw("rolling back db tx")
			err = impl.interceptedEventsRepository.RollbackTx(tx)
			if err != nil {
				impl.logger.Errorw("error in rolling back db transaction while saving intercepted event executions", "interceptEventExecs", interceptEventExecs, "err", err)
			}
		}
	}()

	_, err = impl.interceptedEventsRepository.Save(interceptEventExecs, tx)
	if err != nil {
		impl.logger.Errorw("error in saving intercepted event executions ", "interceptEventExecs", interceptEventExecs, "err", err)
		return err
	}

	err = impl.interceptedEventsRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction while saving intercepted event executions", "interceptEventExecs", interceptEventExecs, "err", err)
		return err
	}
	return nil
}
