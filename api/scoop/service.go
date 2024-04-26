package scoop

import (
	"context"
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/autoRemediation"
	"github.com/devtron-labs/devtron/pkg/autoRemediation/repository"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/pkg/errors"
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

	interceptEventExecs := make([]*repository.InterceptedEventExecution, 0, len(triggers))
	for _, trigger := range triggers {
		switch trigger.IdentifierType {
		case repository.DEVTRON_JOB:
			interceptEventExec := impl.triggerJob(trigger)
			interceptEventExec.ClusterId = interceptedEvent.ClusterId
			interceptEventExec.Event = event
			interceptEventExec.InvolvedObject = involvedObj
			interceptEventExec.InterceptedAt = interceptedEvent.InterceptedAt
			interceptEventExec.Namespace = interceptedEvent.Namespace
			interceptEventExec.Message = interceptedEvent.Message
			interceptEventExec.MessageType = interceptedEvent.MessageType
			interceptEventExecs = append(interceptEventExecs, interceptEventExec)
		}
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

func (impl ServiceImpl) triggerJob(trigger *autoRemediation.Trigger) *repository.InterceptedEventExecution {
	runtimeParams := bean.RuntimeParameters{
		EnvVariables: make(map[string]string),
	}
	for _, param := range trigger.Data.RuntimeParameters {
		runtimeParams.EnvVariables[param.Key] = param.Value
	}

	request := bean.CiTriggerRequest{
		PipelineId: trigger.Data.PipelineId,
		// system user
		TriggeredBy:   1,
		EnvironmentId: trigger.Data.ExecutionEnvironmentId,
		PipelineType:  "CI_BUILD",
		RuntimeParams: &runtimeParams,
	}

	ciWorkflowId := 0
	status := repository.Progressing
	executionMessage := ""

	// get the commit for this pipeline as we need it during trigger
	// this call internally fetches the commits from git-sensor.
	gitCommits, err := impl.ciHandler.FetchMaterialsByPipelineId(trigger.Data.PipelineId, true)

	// if errored or no git commits are find, we should not trigger the job as, it will eventually fail.
	if err != nil || len(gitCommits) == 0 || len(gitCommits[0].History) == 0 {
		if err == nil {
			err = errors.New("no git commits found")
		}
		impl.logger.Errorw("error in getting git commits for ci pipeline", "ciPipelineId", trigger.Data.PipelineId, "err", err)
		executionMessage = err.Error()
		status = repository.Errored
	} else {

		request.CiPipelineMaterial = []bean.CiPipelineMaterial{
			{
				Id: gitCommits[0].Id,
				GitCommit: pipelineConfig.GitCommit{
					Commit: gitCommits[0].History[0].Commit,
				},
			},
		}

		// trigger job pipeline
		ciWorkflowId, err = impl.ciHandler.HandleCIManual(request)
		if err != nil {
			impl.logger.Errorw("error in trigger job ci pipeline", "triggerRequest", request, "err", err)
			executionMessage = err.Error()
			status = repository.Errored
		}

	}

	interceptEventExec := &repository.InterceptedEventExecution{
		TriggerId:          trigger.Id,
		TriggerExecutionId: ciWorkflowId,
		Status:             status,
		// store the error here if something goes wrong before actually triggering the job even
		ExecutionMessage: executionMessage,
	}

	return interceptEventExec
}
