package autoRemediation

import (
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/autoRemediation/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"go.uber.org/zap"
	"gopkg.in/square/go-jose.v2/json"
	"strings"
	"time"
)

type WatcherService interface {
	CreateWatcher(watcherRequest WatcherDto) (int, error)
	GetWatcherById(watcherId int) (*WatcherDto, error)
	DeleteWatcherById(watcherId int) error
	RetrieveInterceptedEvents() (*[]InterceptedEventsDto, error)
	UpdateWatcherById(watcherId int, watcherRequest WatcherDto) error
	FindAll(offset int, size int, sortOrder string, searchString string, from time.Time, to time.Time, watchers []string, clusters []string, namespaces []string) ([]EventsResponse, int, error)
}
type WatcherServiceImpl struct {
	watcherRepository            repository.WatcherRepository
	triggerRepository            repository.TriggerRepository
	interceptedEventsRepository  repository.InterceptedEventsRepository
	appRepository                appRepository.AppRepository
	ciPipelineRepository         pipelineConfig.CiPipelineRepository
	environmentRepository        repository2.EnvironmentRepository
	appWorkflowMappingRepository appWorkflow.AppWorkflowRepository
	clusterRepository            repository2.ClusterRepository
	logger                       *zap.SugaredLogger
}

func NewWatcherServiceImpl(watcherRepository repository.WatcherRepository, triggerRepository repository.TriggerRepository, interceptedEventsRepository repository.InterceptedEventsRepository, appRepository appRepository.AppRepository, ciPipelineRepository pipelineConfig.CiPipelineRepository, environmentRepository repository2.EnvironmentRepository, appWorkflowMappingRepository appWorkflow.AppWorkflowRepository, clusterRepository repository2.ClusterRepository,
	logger *zap.SugaredLogger) *WatcherServiceImpl {
	return &WatcherServiceImpl{
		watcherRepository:            watcherRepository,
		triggerRepository:            triggerRepository,
		interceptedEventsRepository:  interceptedEventsRepository,
		appRepository:                appRepository,
		ciPipelineRepository:         ciPipelineRepository,
		environmentRepository:        environmentRepository,
		appWorkflowMappingRepository: appWorkflowMappingRepository,
		clusterRepository:            clusterRepository,
		logger:                       logger,
	}
}
func (impl *WatcherServiceImpl) CreateWatcher(watcherRequest WatcherDto) (int, error) {

	var gvks []string
	for _, res := range watcherRequest.EventConfiguration.K8sResources {
		jsonString, _ := json.Marshal(res)
		gvks = append(gvks, string(jsonString))
	}
	strings.Join(gvks, ",")
	watcher := &repository.Watcher{
		Name:             watcherRequest.Name,
		Desc:             watcherRequest.Description,
		FilterExpression: watcherRequest.EventConfiguration.EventExpression,
		Gvks:             gvks,
	}
	tx, err := impl.watcherRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in creating watcher", "error", err)
		return 0, err
	}
	defer impl.watcherRepository.RollbackTx(tx)
	watcher, err = impl.watcherRepository.Save(watcher, tx)
	if err != nil {
		impl.logger.Errorw("error in saving watcher", "error", err)
		return 0, err
	}
	err = impl.createTriggerForWatcher(watcherRequest, watcher.Id)
	if err != nil {
		impl.logger.Errorw("error in saving triggers", "error", err)
		return 0, err
	}
	return watcher.Id, nil
}
func (impl *WatcherServiceImpl) createTriggerForWatcher(watcherRequest WatcherDto, watcherId int) error {
	var jsonData []byte
	for _, res := range watcherRequest.Triggers {
		job, err := impl.appRepository.FindJobByDisplayName(res.Data.JobName)
		if err != nil {
			impl.logger.Errorw("error in fetching job by job name ", "error", err)
			return err
		}
		pipeline, err := impl.ciPipelineRepository.FindByName(res.Data.PipelineName)
		if err != nil {
			impl.logger.Errorw("error in fetching pipeline by pipeline name ", "error", err)
			return err
		}
		environment, err := impl.environmentRepository.FindOne(res.Data.ExecutionEnvironment)
		if err != nil {
			impl.logger.Errorw("error in fetching environmenr by environment name ", "error", err)
			return err
		}
		triggerData := TriggerData{
			RuntimeParameters:      res.Data.RuntimeParameters,
			JobId:                  job.Id,
			JobName:                res.Data.JobName,
			PipelineId:             pipeline.Id,
			PipelineName:           res.Data.PipelineName,
			ExecutionEnvironment:   res.Data.ExecutionEnvironment,
			ExecutionEnvironmentId: environment.Id,
		}
		jsonData, err = json.Marshal(triggerData)
		if err != nil {
			impl.logger.Errorw("error in trigger data ", "error", err)
			return err
		}
		trigger := &repository.Trigger{
			WatcherId: watcherId,
			Data:      jsonData,
		}
		if res.IdentifierType == string(repository.DEVTRON_JOB) {
			trigger.Type = repository.DEVTRON_JOB
		}
		tx, err := impl.triggerRepository.StartTx()
		if err != nil {
			impl.logger.Errorw("error in creating trigger", "error", err)
			return err
		}
		defer impl.triggerRepository.RollbackTx(tx)
		_, err = impl.triggerRepository.Save(trigger, tx)
		if err != nil {
			impl.logger.Errorw("error in saving trigger", "error", err)
			return err
		}
	}
	return nil
}
func (impl *WatcherServiceImpl) GetWatcherById(watcherId int) (*WatcherDto, error) {
	watcher, err := impl.watcherRepository.GetWatcherById(watcherId)
	if err != nil {
		impl.logger.Errorw("error in getting watcher", "error", err)
		return &WatcherDto{}, err
	}
	var k8sResources []K8sResource
	for _, gvksString := range watcher.Gvks {
		var res K8sResource
		if err := json.Unmarshal([]byte(gvksString), &res); err != nil {
			impl.logger.Errorw("error in unmarshalling gvks", "error", err)
			return &WatcherDto{}, err
		}
		k8sResources = append(k8sResources, res)
	}
	watcherResponse := WatcherDto{
		Name:        watcher.Name,
		Description: watcher.Desc,
		EventConfiguration: EventConfiguration{
			K8sResources:    k8sResources,
			EventExpression: watcher.FilterExpression,
		},
	}
	triggers, err := impl.triggerRepository.GetTriggerByWatcherId(watcherId)
	if err != nil {
		impl.logger.Errorw("error in getting trigger for watcher id", "watcherId", watcherId, "error", err)
		return &WatcherDto{}, err
	}
	for _, trigger := range *triggers {
		var triggerResp Trigger
		if err := json.Unmarshal(trigger.Data, &triggerResp); err != nil {
			impl.logger.Errorw("error in unmarshalling trigger data", "error", err)
			return &WatcherDto{}, err
		}
		triggerResp.IdentifierType = string(trigger.Type)
		watcherResponse.Triggers = append(watcherResponse.Triggers, triggerResp)
	}
	return &watcherResponse, nil

}
func (impl *WatcherServiceImpl) DeleteWatcherById(watcherId int) error {
	err := impl.triggerRepository.DeleteTriggerByWatcherId(watcherId)
	if err != nil {
		impl.logger.Errorw("error in deleting trigger by watcher id", "watcherId", watcherId, "error", err)
		return err
	}
	err = impl.watcherRepository.DeleteWatcherById(watcherId)
	if err != nil {
		impl.logger.Errorw("error in deleting watcher by its id", watcherId, "error", err)
		return err
	}
	return nil
}
func (impl *WatcherServiceImpl) RetrieveInterceptedEvents() (*[]InterceptedEventsDto, error) {
	// message type?
	var interceptedEventsResponse []InterceptedEventsDto
	interceptedEvents, err := impl.interceptedEventsRepository.GetAllInterceptedEvents()
	if err != nil {
		impl.logger.Errorw("error in retrieving intercepted events", "error", err)
		return &[]InterceptedEventsDto{}, err
	}
	for _, interceptedEvent := range interceptedEvents {
		cluster, err := impl.clusterRepository.FindById(interceptedEvent.ClusterId)
		if err != nil {
			impl.logger.Errorw("error in retrieving cluster name ", "error", err)
			return &[]InterceptedEventsDto{}, err
		}
		interceptedEventResponse := InterceptedEventsDto{
			Message:         interceptedEvent.Message,
			MessageType:     interceptedEvent.MessageType,
			Event:           interceptedEvent.Event,
			InvolvedObject:  interceptedEvent.InvolvedObject,
			ClusterName:     cluster.ClusterName,
			Namespace:       interceptedEvent.Namespace,
			InterceptedTime: (interceptedEvent.InterceptedAt).String(),
			ExecutionStatus: string(interceptedEvent.Status),
			TriggerId:       interceptedEvent.TriggerId,
		}
		triggerResp := Trigger{}
		trigger, err := impl.triggerRepository.GetTriggerById(interceptedEventResponse.TriggerId)
		if err != nil {
			impl.logger.Errorw("error in retrieving intercepted events", "error", err)
			return &[]InterceptedEventsDto{}, err
		}
		triggerResp.IdentifierType = string(trigger.Type)
		triggerRespData := TriggerData{}
		if err := json.Unmarshal(trigger.Data, &triggerRespData); err != nil {
			impl.logger.Errorw("error in unmarshalling trigger data", "error", err)
			return &[]InterceptedEventsDto{}, err
		}
		triggerResp.Data.JobName = triggerRespData.JobName
		triggerResp.Data.PipelineName = triggerRespData.PipelineName
		triggerResp.Data.RuntimeParameters = triggerRespData.RuntimeParameters
		triggerResp.Data.ExecutionEnvironment = triggerRespData.ExecutionEnvironment
		triggerResp.Data.PipelineId = triggerRespData.PipelineId
		triggerResp.Data.JobId = triggerRespData.JobId
		triggerResp.Data.ExecutionEnvironmentId = triggerRespData.ExecutionEnvironmentId
		interceptedEventResponse.Trigger = triggerResp
		interceptedEventsResponse = append(interceptedEventsResponse, interceptedEventResponse)
	}
	return &interceptedEventsResponse, nil
}
func (impl *WatcherServiceImpl) UpdateWatcherById(watcherId int, watcherRequest WatcherDto) error {
	watcher, err := impl.watcherRepository.GetWatcherById(watcherId)
	if err != nil {
		impl.logger.Errorw("error in retrieving watcher by id", watcherId, "error", err)
		return err
	}
	var gvks []string
	for _, res := range watcherRequest.EventConfiguration.K8sResources {
		jsonString, _ := json.Marshal(res)
		gvks = append(gvks, string(jsonString))
	}
	strings.Join(gvks, ",")
	watcher.Name = watcherRequest.Name
	watcher.Desc = watcherRequest.Description
	watcher.FilterExpression = watcherRequest.EventConfiguration.EventExpression
	watcher.Gvks = gvks
	err = impl.triggerRepository.DeleteTriggerByWatcherId(watcher.Id)
	if err != nil {
		impl.logger.Errorw("error in deleting trigger by watcher id", watcherId, "error", err)
		return err
	}
	err = impl.createTriggerForWatcher(watcherRequest, watcherId)
	if err != nil {
		impl.logger.Errorw("error in creating trigger by watcher id", watcherId, "error", err)
		return err
	}
	return nil
}
func (impl *WatcherServiceImpl) FindAll(offset int, size int, sortOrder string, searchString string, from time.Time, to time.Time, watchers []string, clusters []string, namespaces []string) (EventsResponse, error) {
	events, err := impl.interceptedEventsRepository.FindAll(offset, size, sortOrder, searchString, from, to, watchers, clusters, namespaces)
	if err != nil {
		impl.logger.Errorw("error while fetching events", "err", err)
		return EventsResponse{}, err
	}
	eventResponse := EventsResponse{
		Size:   size,
		Offset: offset,
		Total:  len(events),
	}
	for _, event := range events {
		watcher, err := impl.triggerRepository.GetWatcherByTriggerId(event.TriggerId)
		if err != nil {
			impl.logger.Errorw("error while fetching events", "err", err)
			return EventsResponse{}, err
		}
		trigger, err := impl.triggerRepository.GetTriggerById(event.TriggerId)
		if err != nil {
			impl.logger.Errorw("error while fetching trigger", "err", err)
			return EventsResponse{}, err
		}
		triggerRespData := TriggerData{}
		if err := json.Unmarshal(trigger.Data, &triggerRespData); err != nil {
			impl.logger.Errorw("error in unmarshalling trigger data", "error", err)
			return EventsResponse{}, err
		}
		pipeline, err := impl.ciPipelineRepository.FindByName(triggerRespData.PipelineName)
		if err != nil {
			impl.logger.Errorw("error in fetching pipeline by pipeline name ", "error", err)
			return EventsResponse{}, err
		}
		ciWorkflow, err := impl.appWorkflowMappingRepository.FindWFCIMappingByCIPipelineId(pipeline.Id)
		eventsItem := EventsItem{
			Id:              event.Id,
			Name:            watcher.Name,
			Description:     watcher.Desc,
			JobPipelineName: pipeline.Name,
			JobPipelineId:   pipeline.Id,
			WorkflowId:      ciWorkflow[0].AppWorkflowId,
		}
		eventResponse.List = append(eventResponse.List, eventsItem)
	}
	return eventResponse, nil
}
