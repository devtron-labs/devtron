package scoop

import (
	"context"
	"encoding/json"
	"fmt"
	client2 "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/autoRemediation"
	"github.com/devtron-labs/devtron/pkg/autoRemediation/repository"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/util"
	util5 "github.com/devtron-labs/devtron/util/event"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"time"
)

type Service interface {
	HandleInterceptedEvent(ctx context.Context, event *InterceptedEvent) error
	HandleNotificationEvent(ctx context.Context, clusterId int, notification map[string]interface{}) error
}

type ServiceImpl struct {
	logger                       *zap.SugaredLogger
	watcherService               autoRemediation.WatcherService
	ciHandler                    pipeline.CiHandler
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository
	interceptedEventsRepository  repository.InterceptedEventsRepository
	clusterService               cluster.ClusterService
	eventClient                  client2.EventClient
	eventFactory                 client2.EventFactory
}

func NewServiceImpl(logger *zap.SugaredLogger,
	watcherService autoRemediation.WatcherService,
	ciHandler pipeline.CiHandler,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	interceptedEventsRepository repository.InterceptedEventsRepository,
	clusterService cluster.ClusterService,
	eventClient client2.EventClient,
	eventFactory client2.EventFactory,
) *ServiceImpl {
	return &ServiceImpl{
		logger:                       logger,
		watcherService:               watcherService,
		ciHandler:                    ciHandler,
		ciPipelineMaterialRepository: ciPipelineMaterialRepository,
		interceptedEventsRepository:  interceptedEventsRepository,
		clusterService:               clusterService,
		eventClient:                  eventClient,
		eventFactory:                 eventFactory,
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
			interceptEventExec := impl.triggerJob(trigger, event)
			interceptEventExec.ClusterId = interceptedEvent.ClusterId
			interceptEventExec.Event = event
			interceptEventExec.InvolvedObject = involvedObj
			interceptEventExec.InterceptedAt = interceptedEvent.InterceptedAt
			interceptEventExec.Namespace = interceptedEvent.Namespace
			interceptEventExec.Message = interceptedEvent.Message
			interceptEventExec.MessageType = interceptedEvent.MessageType
			interceptEventExec.AuditLog = sql.NewDefaultAuditLog(1)
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
			rollbackErr := impl.interceptedEventsRepository.RollbackTx(tx)
			if err != nil {
				impl.logger.Errorw("error in rolling back db transaction while saving intercepted event executions", "interceptEventExecs", interceptEventExecs, "err", rollbackErr)
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

func (impl ServiceImpl) triggerJob(trigger *autoRemediation.Trigger, event string) *repository.InterceptedEventExecution {
	runtimeParams := bean.RuntimeParameters{
		EnvVariables: make(map[string]string),
	}
	for _, param := range trigger.Data.RuntimeParameters {
		runtimeParams.EnvVariables[param.Key] = param.Value
	}

	runtimeParams.EnvVariables["event"] = event
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

func (impl ServiceImpl) HandleNotificationEvent(ctx context.Context, clusterId int, notification map[string]interface{}) error {
	cluster, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in finding cluster information by id", "clusterId", clusterId, "err", err)
		return err
	}

	var configType string
	var ok bool
	var configName string

	if configType, ok = notification["configType"].(string); !ok {
		return errors.New("config type not set")
	}

	if !(string(util5.Slack) == configType || string(util5.Webhook) == configType || string(util5.SES) == configType || string(util5.SMTP) == configType) {
		return errors.New("un-supported config type")
	}

	if configNameIf, ok := notification["configName"]; ok {
		if configName, ok = configNameIf.(string); ok {
			return errors.New("un-supported config name")
		}

		if (string(util5.Slack) == configType || string(util5.Webhook) == configType) && configName == "" {
			return errors.New("config name is required for webhook/slack")
		}
	}

	notification["cluster"] = cluster.ClusterName
	notification["eventTypeId"] = util5.ScoopNotification
	notification["eventTime"] = time.Now()
	notification["correlationId"] = fmt.Sprintf("%s", uuid.NewV4())

	emailIfs, _ := notification["emailIds"].([]interface{})
	emails := util2.Map(emailIfs, func(emailIf interface{}) string {
		return emailIf.(string)
	})
	payload, err := impl.eventFactory.BuildScoopNotificationEventProviders(util5.Channel(configType), configName, emails)
	if err != nil {
		impl.logger.Errorw("error in constructing event payload ", "clusterId", clusterId, "notification", notification, "err", err)
		return err
	}

	notification["payload"] = payload
	_, err = impl.eventClient.SendAnyEvent(notification)
	if err != nil {
		impl.logger.Errorw("error in sending scoop event notification", "clusterId", clusterId, "notification", notification, "err", err)
	}
	return err
}
