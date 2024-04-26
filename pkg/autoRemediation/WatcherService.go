package autoRemediation

import (
	json2 "encoding/json"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/autoRemediation/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"gopkg.in/square/go-jose.v2/json"
)

type WatcherService interface {
	CreateWatcher(watcherRequest *WatcherDto, userId int32) (int, error)
	GetWatcherById(watcherId int) (*WatcherDto, error)
	DeleteWatcherById(watcherId int) error
	UpdateWatcherById(watcherId int, watcherRequest *WatcherDto, userId int32) error
	// RetrieveInterceptedEvents(offset int, size int, sortOrder string, searchString string, from time.Time, to time.Time, watchers []string, clusters []string, namespaces []string) (EventsResponse, error)
	//FindAllWatchers(offset int, search string, size int, sortOrder string, sortOrderBy string) (WatchersResponse, error)
	GetTriggerByWatcherIds(watcherIds []int) ([]*Trigger, error)
}

type ScoopConfig struct {
	WatcherUrl string ``
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

func NewWatcherServiceImpl(watcherRepository repository.WatcherRepository,
	triggerRepository repository.TriggerRepository,
	interceptedEventsRepository repository.InterceptedEventsRepository,
	appRepository appRepository.AppRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	environmentRepository repository2.EnvironmentRepository,
	appWorkflowMappingRepository appWorkflow.AppWorkflowRepository,
	clusterRepository repository2.ClusterRepository,
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
func (impl *WatcherServiceImpl) CreateWatcher(watcherRequest *WatcherDto, userId int32) (int, error) {

	gvks, err := fetchGvksFromK8sResources(watcherRequest.EventConfiguration.K8sResources)
	if err != nil {
		impl.logger.Errorw("error in creating fetching gvks", "error", err)
		return 0, err
	}
	watcher := &repository.Watcher{
		Name:             watcherRequest.Name,
		Desc:             watcherRequest.Description,
		FilterExpression: watcherRequest.EventConfiguration.EventExpression,
		Gvks:             gvks,
		Active:           true,
		AuditLog:         sql.NewDefaultAuditLog(userId),
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
	err = impl.createTriggerForWatcher(watcherRequest, watcher.Id, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in saving triggers", "error", err)
		return 0, err
	}

	return watcher.Id, nil
}

func fetchGvksFromK8sResources(resources []K8sResource) (string, error) {
	gvks, err := json.Marshal(resources)
	if err != nil {
		return "", err
	}
	return string(gvks), nil
}

func (impl *WatcherServiceImpl) createTriggerForWatcher(watcherRequest *WatcherDto, watcherId int, userId int32, tx *pg.Tx) error {
	for _, trigger := range watcherRequest.Triggers {
		if trigger.IdentifierType == repository.DEVTRON_JOB {
			err := impl.createTriggerJobsForWatcher(watcherRequest, watcherId, userId, tx)
			if err != nil {
				impl.logger.Errorw("error in creating triggers for watcher", "error", err)
				return err
			}
		}
	}
	return nil
}

func (impl *WatcherServiceImpl) createTriggerJobsForWatcher(watcherRequest *WatcherDto, watcherId int, userId int32, tx *pg.Tx) error {
	displayNameToId, pipelineNameToId, envNameToId, pipelineIdtoAppworkflow, err := impl.getJobEnvPipelineDetailsForWatcher(watcherRequest.Triggers)
	if err != nil {
		impl.logger.Errorw("error in retrieving details of job pipeline environment", "error", err)
		return err
	}
	var triggers []*repository.Trigger
	for _, res := range watcherRequest.Triggers {
		triggerData := TriggerData{
			RuntimeParameters:      res.Data.RuntimeParameters,
			JobId:                  displayNameToId[res.Data.JobName],
			JobName:                res.Data.JobName,
			PipelineId:             pipelineNameToId[res.Data.PipelineName],
			PipelineName:           res.Data.PipelineName,
			ExecutionEnvironment:   res.Data.ExecutionEnvironment,
			ExecutionEnvironmentId: envNameToId[res.Data.ExecutionEnvironment],
			WorkflowId:             pipelineIdtoAppworkflow[pipelineNameToId[res.Data.PipelineName]],
		}
		jsonData, err := json.Marshal(triggerData)
		if err != nil {
			impl.logger.Errorw("error in trigger data ", "error", err)
			return err
		}
		trigger := &repository.Trigger{
			WatcherId: watcherId,
			Type:      repository.DEVTRON_JOB,
			Data:      jsonData,
			Active:    true,
			AuditLog:  sql.NewDefaultAuditLog(userId),
		}
		triggers = append(triggers, trigger)
	}
	_, err = impl.triggerRepository.SaveInBulk(triggers, tx)
	if err != nil {
		impl.logger.Errorw("error in saving trigger", "error", err)
		return err
	}
	defer impl.triggerRepository.RollbackTx(tx)
	err = impl.triggerRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to create trigger", "error", err)
		return err
	}
	return nil
}
func (impl *WatcherServiceImpl) getJobEnvPipelineDetailsForWatcher(triggers []Trigger) (map[string]int, map[string]int, map[string]int, map[int]int, error) {
	var jobNames, envNames, pipelineNames []string
	for _, trig := range triggers {
		jobNames = append(jobNames, trig.Data.JobName)
		envNames = append(envNames, trig.Data.ExecutionEnvironment)
		pipelineNames = append(pipelineNames, trig.Data.PipelineName)
	}
	apps, err := impl.appRepository.FetchAppByDisplayNamesForJobs(jobNames)
	if err != nil {
		impl.logger.Errorw("error in fetching apps", "error", err)
		return nil, nil, nil, nil, err
	}
	var jobIds []int
	for _, app := range apps {
		jobIds = append(jobIds, app.Id)
	}
	pipelines, err := impl.ciPipelineRepository.FindByNames(pipelineNames, jobIds)
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines", "error", err)
		return nil, nil, nil, nil, err
	}
	var pipelinesId []int
	for _, pipeline := range pipelines {
		pipelinesId = append(pipelinesId, pipeline.Id)
	}
	envs, err := impl.environmentRepository.FindByNames(envNames)
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "error", err)
		return nil, nil, nil, nil, err
	}
	displayNameToId := make(map[string]int)
	for _, app := range apps {
		displayNameToId[app.DisplayName] = app.Id
	}
	pipelineNameToId := make(map[string]int)
	for _, pipeline := range pipelines {
		pipelineNameToId[pipeline.Name] = pipeline.Id
	}
	workflows, err := impl.appWorkflowMappingRepository.FindWFCIMappingByCIPipelineIds(pipelinesId)
	if err != nil {
		impl.logger.Errorw("error in retrieving workflows ", "error", err)
		return nil, nil, nil, nil, err
	}
	var pipelineIdtoAppworkflow map[int]int
	for _, workflow := range workflows {
		pipelineIdtoAppworkflow[workflow.ComponentId] = workflow.AppWorkflowId
	}
	envNameToId := make(map[string]int)
	for _, env := range envs {
		envNameToId[env.Name] = env.Id
	}
	return displayNameToId, pipelineNameToId, envNameToId, pipelineIdtoAppworkflow, nil
}
func (impl *WatcherServiceImpl) GetWatcherById(watcherId int) (*WatcherDto, error) {
	watcher, err := impl.watcherRepository.GetWatcherById(watcherId)
	if err != nil {
		impl.logger.Errorw("error in getting watcher", "error", err)
		return nil, err
	}
	k8sResources, err := getK8sResourcesFromGvks(watcher.Gvks)
	if err != nil {
		impl.logger.Errorw("error in getting k8sResources from gvks", "error", err)
		return nil, err
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
	for _, trigger := range triggers {
		triggerResp, err := impl.getTriggerDataFromJson(trigger.Data)
		if err != nil {
			impl.logger.Errorw("error in unmarshalling trigger data", "error", err)
			return nil, err
		}
		triggerResp.IdentifierType = trigger.Type
		watcherResponse.Triggers = append(watcherResponse.Triggers, triggerResp)
	}
	return &watcherResponse, nil

}

func (impl *WatcherServiceImpl) getTriggerDataFromJson(data json2.RawMessage) (Trigger, error) {
	var triggerResp Trigger
	if err := json.Unmarshal(data, &triggerResp); err != nil {
		impl.logger.Errorw("error in unmarshalling trigger data", "error", err)
		return Trigger{}, err
	}
	return triggerResp, nil
}

func getK8sResourcesFromGvks(gvks string) ([]K8sResource, error) {
	var k8sResources []K8sResource
	if err := json.Unmarshal([]byte(gvks), &k8sResources); err != nil {
		return nil, err
	}
	return k8sResources, nil
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

func (impl *WatcherServiceImpl) UpdateWatcherById(watcherId int, watcherRequest *WatcherDto, userId int32) error {
	watcher, err := impl.watcherRepository.GetWatcherById(watcherId)
	if err != nil {
		impl.logger.Errorw("error in retrieving watcher by id", watcherId, "error", err)
		return err
	}
	gvks, err := fetchGvksFromK8sResources(watcherRequest.EventConfiguration.K8sResources)
	watcher.Name = watcherRequest.Name
	watcher.Desc = watcherRequest.Description
	watcher.FilterExpression = watcherRequest.EventConfiguration.EventExpression
	watcher.Gvks = gvks
	_, err = impl.watcherRepository.Update(watcher)
	if err != nil {
		impl.logger.Errorw("error in updating watcher", "error", err)
		return err
	}
	err = impl.triggerRepository.DeleteTriggerByWatcherId(watcher.Id)
	if err != nil {
		impl.logger.Errorw("error in deleting trigger by watcher id", watcherId, "error", err)
		return err
	}
	tx, err := impl.triggerRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in creating transaction for creating trigger", watcherId, "error", err)
		return err
	}
	err = impl.createTriggerForWatcher(watcherRequest, watcherId, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in creating trigger by watcher id", watcherId, "error", err)
		return err
	}
	return nil
}

//func (impl *WatcherServiceImpl) FindAllWatchers(offset int, search string, size int, sortOrder string, sortOrderBy string) (WatchersResponse, error) {
//	search = strings.ToLower(search)
//	params := WatcherQueryParams{
//		Offset:      offset,
//		Size:        size,
//		Search:      search,
//		SortOrderBy: sortOrderBy,
//		SortOrder:   sortOrder,
//	}
//	watchers, err := impl.watcherRepository.FindAllWatchersByQueryName(params)
//	if err != nil {
//		impl.logger.Errorw("error in retrieving watchers ", "error", err)
//		return WatchersResponse{}, err
//	}
//	var watcherIds []int
//	for _, watcher := range watchers {
//		watcherIds = append(watcherIds, watcher.Id)
//	}
//	triggers, err := impl.triggerRepository.GetTriggerByWatcherIds(watcherIds)
//	if err != nil {
//		impl.logger.Errorw("error in retrieving triggers ", "error", err)
//		return WatchersResponse{}, err
//	}
//	var triggerIds []int
//	watcherIdToTrigger := make(map[int]repository.Trigger)
//	for _, trigger := range triggers {
//		triggerIds = append(triggerIds, trigger.Id)
//		watcherIdToTrigger[trigger.WatcherId] = *trigger
//	}
//
//	watcherResponses := WatchersResponse{
//		Size:   params.Size,
//		Offset: params.Offset,
//		Total:  len(watchers),
//	}
//	var pipelineIds []int
//	for _, watcher := range watchers {
//		var triggerResp TriggerData
//		if err := json.Unmarshal(watcherIdToTrigger[watcher.Id].Data, &triggerResp); err != nil {
//			impl.logger.Errorw("error in unmarshalling trigger data", "error", err)
//			return WatchersResponse{}, err
//		}
//		pipelineIds = append(pipelineIds, triggerResp.PipelineId)
//		watcherResponses.List = append(watcherResponses.List, WatcherItem{
//			Name:            watcher.Name,
//			Description:     watcher.Desc,
//			JobPipelineName: triggerResp.PipelineName,
//			JobPipelineId:   triggerResp.PipelineId,
//		})
//	}
//	workflows, err := impl.appWorkflowMappingRepository.FindWFCIMappingByCIPipelineIds(pipelineIds)
//	if err != nil {
//		impl.logger.Errorw("error in retrieving workflows ", "error", err)
//		return WatchersResponse{}, err
//	}
//	var pipelineIdtoAppworkflow map[int]int
//	for _, workflow := range workflows {
//		pipelineIdtoAppworkflow[workflow.ComponentId] = workflow.AppWorkflowId
//	}
//	for _, watcherList := range watcherResponses.List {
//		watcherList.WorkflowId = pipelineIdtoAppworkflow[watcherList.JobPipelineId]
//	}
//
//	return watcherResponses, nil
//}

func (impl *WatcherServiceImpl) GetTriggerByWatcherIds(watcherIds []int) ([]*Trigger, error) {
	triggers, err := impl.triggerRepository.GetTriggerByWatcherIds(watcherIds)
	if err != nil {
		impl.logger.Errorw("error in getting triggers by watcher ids", "watcherIds", watcherIds, "err", err)
		return nil, err
	}

	triggersResult := make([]*Trigger, 0, len(triggers))
	for _, trigger := range triggers {
		triggerResp := Trigger{}
		triggerResp.Id = trigger.Id
		triggerResp.IdentifierType = trigger.Type
		triggerData := TriggerData{}
		if err := json.Unmarshal(trigger.Data, &triggerData); err != nil {
			impl.logger.Errorw("error in unmarshalling trigger data", "error", err)
			return nil, err
		}
		triggerResp.Data.JobName = triggerData.JobName
		triggerResp.Data.PipelineName = triggerData.PipelineName
		triggerResp.Data.RuntimeParameters = triggerData.RuntimeParameters
		triggerResp.Data.ExecutionEnvironment = triggerData.ExecutionEnvironment
		triggerResp.Data.PipelineId = triggerData.PipelineId
		triggerResp.Data.JobId = triggerData.JobId
		triggerResp.Data.ExecutionEnvironmentId = triggerData.ExecutionEnvironmentId

		triggersResult = append(triggersResult, &triggerResp)
	}

	return triggersResult, nil
}
