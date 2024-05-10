package autoRemediation

import (
	"context"
	"fmt"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/autoRemediation/repository"
	types2 "github.com/devtron-labs/devtron/pkg/autoRemediation/types"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/k8s/application"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util"
	scoopClient "github.com/devtron-labs/scoop/client"
	"github.com/devtron-labs/scoop/types"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/square/go-jose.v2/json"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sort"
	"time"
)

type WatcherService interface {
	CreateWatcher(watcherRequest *types2.WatcherDto, userId int32) (int, error)
	GetWatcherById(watcherId int) (*types2.WatcherDto, error)
	DeleteWatcherById(watcherId int, userId int32) error
	UpdateWatcherById(watcherId int, watcherRequest *types2.WatcherDto, userId int32) error
	RetrieveInterceptedEvents(params *types2.InterceptedEventQueryParams) (*types2.InterceptedResponse, error)
	FindAllWatchers(params types2.WatcherQueryParams) (types2.WatchersResponse, error)
	GetTriggerByWatcherIds(watcherIds []int) ([]*types2.Trigger, error)
	GetWatchersByClusterId(clusterId int) ([]*types.Watcher, error)
	GetInterceptedEventById(interceptedEventId int) (*types2.InterceptedEventData, error)
}

type WatcherServiceImpl struct {
	watcherRepository               repository.K8sEventWatcherRepository
	triggerRepository               repository.TriggerRepository
	interceptedEventsRepository     repository.InterceptedEventsRepository
	appRepository                   appRepository.AppRepository
	ciPipelineRepository            pipelineConfig.CiPipelineRepository
	environmentRepository           repository2.EnvironmentRepository
	appWorkflowMappingRepository    appWorkflow.AppWorkflowRepository
	clusterRepository               repository2.ClusterRepository
	ciWorkflowRepository            pipelineConfig.CiWorkflowRepository
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService
	k8sApplicationService           application.K8sApplicationService
	logger                          *zap.SugaredLogger
}

func NewWatcherServiceImpl(watcherRepository repository.K8sEventWatcherRepository,
	triggerRepository repository.TriggerRepository,
	interceptedEventsRepository repository.InterceptedEventsRepository,
	appRepository appRepository.AppRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	environmentRepository repository2.EnvironmentRepository,
	appWorkflowMappingRepository appWorkflow.AppWorkflowRepository,
	clusterRepository repository2.ClusterRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository,
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService,
	k8sApplicationService application.K8sApplicationService,
	logger *zap.SugaredLogger,
) *WatcherServiceImpl {

	return &WatcherServiceImpl{
		watcherRepository:               watcherRepository,
		triggerRepository:               triggerRepository,
		interceptedEventsRepository:     interceptedEventsRepository,
		appRepository:                   appRepository,
		ciPipelineRepository:            ciPipelineRepository,
		environmentRepository:           environmentRepository,
		appWorkflowMappingRepository:    appWorkflowMappingRepository,
		clusterRepository:               clusterRepository,
		ciWorkflowRepository:            ciWorkflowRepository,
		resourceQualifierMappingService: resourceQualifierMappingService,
		k8sApplicationService:           k8sApplicationService,
		logger:                          logger,
	}
}

func (impl *WatcherServiceImpl) CreateWatcher(watcherRequest *types2.WatcherDto, userId int32) (int, error) {

	gvks, err := fetchGvksFromK8sResources(watcherRequest.EventConfiguration.K8sResources)
	if err != nil {
		impl.logger.Errorw("error in fetching gvks", "error", err)
		return 0, err
	}

	selectors, err := getSelectorJson(watcherRequest.EventConfiguration.Selectors)
	watcher := &repository.K8sEventWatcher{
		Name:             watcherRequest.Name,
		Description:      watcherRequest.Description,
		FilterExpression: watcherRequest.EventConfiguration.EventExpression,
		SelectedActions:  watcherRequest.EventConfiguration.SelectedActions,
		Selectors:        selectors,
		Gvks:             gvks,
		Active:           true,
		AuditLog:         sql.NewDefaultAuditLog(userId),
	}
	tx, err := impl.watcherRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction of watcher", "error", err)
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
		impl.logger.Errorw("error in creating triggers", "error", err)
		return 0, err
	}

	watcherRequest.Id = watcher.Id
	err = impl.informScoops(types.ADD, watcherRequest)
	if err != nil {
		impl.logger.Errorw("error in informing respective scoops about this watcher creation", "err", err, "watcherRequest", watcherRequest)
		return 0, err
	}
	err = impl.triggerRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to create trigger", "error", err)
		return 0, err
	}
	return watcher.Id, nil
}

func fetchGvksFromK8sResources(resources []*types2.K8sResource) (string, error) {
	gvks, err := json.Marshal(resources)
	if err != nil {
		return "", err
	}
	return string(gvks), nil
}

func getSelectorJson(selectors []types2.Selector) (string, error) {
	selectorBytes, err := json.Marshal(&selectors)
	return string(selectorBytes), err
}

func getSelectorsFromJson(selectorsJson string) ([]types2.Selector, error) {
	selectors := make([]types2.Selector, 0)
	err := json.Unmarshal([]byte(selectorsJson), &selectors)
	return selectors, err
}

func getClusterSelector(clusterName string, selectors []types2.Selector) *types2.Selector {
	for _, selector := range selectors {
		if selector.GroupName == types2.AllClusterGroup || selector.GroupName == clusterName {
			return &selector
		}
	}

	return nil
}

func (impl *WatcherServiceImpl) createTriggerForWatcher(watcherRequest *types2.WatcherDto, watcherId int, userId int32, tx *pg.Tx) error {
	var triggersForTypeJobs []*types2.Trigger
	if watcherRequest.Triggers == nil {
		return nil
	}
	for i, _ := range watcherRequest.Triggers {
		if watcherRequest.Triggers[i].IdentifierType == types2.DEVTRON_JOB {
			triggersForTypeJobs = append(triggersForTypeJobs, &watcherRequest.Triggers[i])
		}
	}
	if len(triggersForTypeJobs) != 0 { // if trigger type is job then job's data is processed
		err := impl.createJobsForTriggerOfWatcher(triggersForTypeJobs, watcherId, userId, tx)
		if err != nil {
			impl.logger.Errorw("error in creating triggers for watcher", "error", err)
			return err
		}
	}
	return nil
}

type jobDetails struct {
	displayNameToId         map[string]int
	pipelineNameToId        map[string]int
	envNameToId             map[string]int
	pipelineIdtoAppworkflow map[int]int
}

func (impl *WatcherServiceImpl) createJobsForTriggerOfWatcher(triggers []*types2.Trigger, watcherId int, userId int32, tx *pg.Tx) error {
	jobInfo, err := impl.getJobEnvPipelineDetailsForWatcher(triggers)
	if err != nil {
		impl.logger.Errorw("error in retrieving details of trigger type job", "error", err)
		return err
	}
	var triggersResult []*repository.AutoRemediationTrigger
	for _, res := range triggers {

		triggerData := types2.TriggerData{
			RuntimeParameters: res.Data.RuntimeParameters,
		}
		if jobInfo.displayNameToId[res.Data.JobName] != 0 && res.Data.PipelineName == "" {
			triggerData.JobId = jobInfo.displayNameToId[res.Data.JobName]
			triggerData.JobName = res.Data.JobName
		}
		if jobInfo.displayNameToId[res.Data.JobName] != 0 && jobInfo.pipelineNameToId[res.Data.PipelineName] != 0 && jobInfo.pipelineIdtoAppworkflow[jobInfo.pipelineNameToId[res.Data.PipelineName]] != 0 {
			triggerData.JobId = jobInfo.displayNameToId[res.Data.JobName]
			triggerData.JobName = res.Data.JobName
			triggerData.PipelineId = jobInfo.pipelineNameToId[res.Data.PipelineName]
			triggerData.PipelineName = res.Data.PipelineName
			triggerData.WorkflowId = jobInfo.pipelineIdtoAppworkflow[jobInfo.pipelineNameToId[res.Data.PipelineName]]
		}
		triggerData.ExecutionEnvironment = res.Data.ExecutionEnvironment
		if jobInfo.envNameToId[res.Data.ExecutionEnvironment] != 0 {
			triggerData.ExecutionEnvironmentId = jobInfo.envNameToId[res.Data.ExecutionEnvironment]
		}
		jsonData, err := json.Marshal(triggerData)
		if err != nil {
			impl.logger.Errorw("error in marshalling trigger data ", "error", err)
			return err
		}
		triggerRes := &repository.AutoRemediationTrigger{
			WatcherId: watcherId,
			Type:      repository.DEVTRON_JOB,
			Data:      string(jsonData),
			Active:    true,
			AuditLog:  sql.NewDefaultAuditLog(userId),
		}
		triggersResult = append(triggersResult, triggerRes)
	}
	_, err = impl.triggerRepository.SaveInBulk(triggersResult, tx)
	if err != nil {
		impl.logger.Errorw("error in saving triggers in bulk", "error", err)
		return err
	}
	return nil
}

func (impl *WatcherServiceImpl) getJobEnvPipelineDetailsForWatcher(triggers []*types2.Trigger) (*jobDetails, error) {
	var jobsDetails *jobDetails
	var jobNames, envNames, pipelineNames []string

	for _, trig := range triggers {
		jobNames = append(jobNames, trig.Data.JobName)
		envNames = append(envNames, trig.Data.ExecutionEnvironment)
		pipelineNames = append(pipelineNames, trig.Data.PipelineName)
	}
	var apps []*appRepository.AppDto
	var err error
	if len(jobNames) != 0 {
		apps, err = impl.appRepository.FetchAppByDisplayNamesForJobs(jobNames)
		if err != nil {
			impl.logger.Errorw("error in fetching apps", "jobNames", jobNames, "error", err)
			return jobsDetails, err
		}
	}
	var jobIds []int
	for _, app := range apps {
		jobIds = append(jobIds, app.Id)
	}
	var pipelines []*pipelineConfig.CiPipeline
	if len(jobIds) != 0 {
		pipelines, err = impl.ciPipelineRepository.FindByNames(pipelineNames, jobIds)
		if err != nil {
			impl.logger.Errorw("error in fetching pipelines", "pipelineNames", pipelineNames, "error", err)
			return jobsDetails, err
		}
	}

	var pipelinesId []int
	for _, pipeline := range pipelines {
		pipelinesId = append(pipelinesId, pipeline.Id)
	}
	var envs []*repository2.Environment
	if len(envNames) != 0 {
		envs, err = impl.environmentRepository.FindByNames(envNames)
		if err != nil {
			impl.logger.Errorw("error in fetching environment", "envNames", envNames, "error", err)
			return jobsDetails, err
		}
	}
	displayNameToId := make(map[string]int)
	for _, app := range apps {
		displayNameToId[app.DisplayName] = app.Id
	}
	pipelineNameToId := make(map[string]int)
	for _, pipeline := range pipelines {
		pipelineNameToId[pipeline.Name] = pipeline.Id
	}
	var workflows []*appWorkflow.AppWorkflowMapping
	if len(pipelinesId) != 0 {
		workflows, err = impl.appWorkflowMappingRepository.FindWFCIMappingByCIPipelineIds(pipelinesId)
		if err != nil {
			impl.logger.Errorw("error in retrieving workflows for pipelineIds", pipelinesId, "error", err)
			return jobsDetails, err
		}
	}
	pipelineIdtoAppworkflow := make(map[int]int)
	for _, workflow := range workflows {
		pipelineIdtoAppworkflow[workflow.ComponentId] = workflow.AppWorkflowId
	}
	envNameToId := make(map[string]int)
	for _, env := range envs {
		envNameToId[env.Name] = env.Id
	}
	return &jobDetails{
		pipelineNameToId:        pipelineNameToId,
		displayNameToId:         displayNameToId,
		envNameToId:             envNameToId,
		pipelineIdtoAppworkflow: pipelineIdtoAppworkflow,
	}, nil

}

func (impl *WatcherServiceImpl) GetWatcherById(watcherId int) (*types2.WatcherDto, error) {
	watcher, err := impl.watcherRepository.GetWatcherById(watcherId)
	if err != nil {
		impl.logger.Errorw("error in getting watcher by id", watcherId, "error", err)
		return nil, err
	}
	k8sResources, err := getK8sResourcesFromGvks(watcher.Gvks)
	if err != nil {
		impl.logger.Errorw("error in getting k8sResources from gvks", "error", err)
		return nil, err
	}

	selectors, err := getSelectorsFromJson(watcher.Selectors)
	if err != nil {
		impl.logger.Errorw("error in getting selectors from selectors json", "error", err)
		return nil, err
	}
	watcherResponse := types2.WatcherDto{
		Name:        watcher.Name,
		Description: watcher.Description,
		EventConfiguration: types2.EventConfiguration{
			K8sResources:    k8sResources,
			EventExpression: watcher.FilterExpression,
			SelectedActions: watcher.SelectedActions,
			Selectors:       selectors,
		},
	}

	triggers, err := impl.triggerRepository.GetTriggerByWatcherId(watcherId)
	if err != nil {
		impl.logger.Errorw("error in getting triggers for watcher id", "watcherId", watcherId, "error", err)
		return &types2.WatcherDto{}, err
	}
	if len(triggers) == 0 {
		watcherResponse.Triggers = []types2.Trigger{}
	}
	for _, trigger := range triggers {
		triggerResp, err := getTriggerDataFromJson(trigger.Data)
		if err != nil {
			impl.logger.Errorw("error in unmarshalling trigger data", "error", err)
			return nil, err
		}
		triggerResponse := types2.Trigger{
			Id:             trigger.Id,
			IdentifierType: types2.TriggerType(trigger.Type),
			Data:           triggerResp,
		}
		watcherResponse.Triggers = append(watcherResponse.Triggers, triggerResponse)
	}

	// selectors, _, err = impl.getEnvSelectors(watcherId)
	// if err != nil {
	// 	impl.logger.Errorw("error in getting selectors for the watcher", "watcherId", watcherId, "error", err)
	// 	return nil, err
	// }
	//
	// watcherResponse.EventConfiguration.Selectors = selectors
	return &watcherResponse, nil

}
func (impl *WatcherServiceImpl) GetInterceptedEventById(interceptedEventId int) (*types2.InterceptedEventData, error) {
	interceptedEvent, err := impl.interceptedEventsRepository.GetInterceptedEventById(interceptedEventId)
	if err != nil {
		impl.logger.Errorw("error in getting intercepted event by id", interceptedEventId, "error", err)
		return nil, err
	}

	return interceptedEvent, nil

}
func getTriggerDataFromJson(data string) (types2.TriggerData, error) {
	var triggerResp types2.TriggerData
	if err := json.Unmarshal([]byte(data), &triggerResp); err != nil {

		return types2.TriggerData{}, err
	}
	return triggerResp, nil
}

func getK8sResourcesFromGvks(gvks string) ([]*types2.K8sResource, error) {
	var k8sResources []*types2.K8sResource
	if err := json.Unmarshal([]byte(gvks), &k8sResources); err != nil {
		return nil, err
	}
	return k8sResources, nil
}

func (impl *WatcherServiceImpl) DeleteWatcherById(watcherId int, userId int32) error {

	tx, err := impl.watcherRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in creating watcher", "error", err)
		return err
	}
	defer func() {
		if err != nil {
			rollbackErr := impl.watcherRepository.RollbackTx(tx)
			if rollbackErr != nil {
				impl.logger.Errorw("error in rolling back in watcher delete request", "watcherId", watcherId, "err", rollbackErr)
			}
		}
	}()

	watcher, err := impl.watcherRepository.GetWatcherById(watcherId)
	if err != nil {
		impl.logger.Errorw("error in retrieving watcher by id", "watcherId", watcherId, "error", err)
		return err
	}

	selectors, err := getSelectorsFromJson(watcher.Selectors)
	if err != nil {
		impl.logger.Errorw("error in retrieving selectors from watcher", "watcherId", watcherId, "error", err)
		return err
	}

	err = impl.triggerRepository.DeleteTriggerByWatcherId(tx, watcherId)
	if err != nil {
		impl.logger.Errorw("error in deleting trigger by watcher id", "watcherId", watcherId, "error", err)
		return err
	}
	err = impl.watcherRepository.DeleteWatcherById(watcherId)
	if err != nil {
		impl.logger.Errorw("error in deleting watcher by its id", watcherId, "error", err)
		return err
	}

	// err = impl.resourceQualifierMappingService.DeleteAllQualifierMappingsByResourceTypeAndId(resourceQualifiers.K8sEventWatcher, watcherId, sql.NewDefaultAuditLog(userId), tx)
	// if err != nil {
	// 	impl.logger.Errorw("error in envs mappings for the watcher", "watcherId", watcherId, "err", err)
	// 	return err
	// }

	// _, envsMap, err := impl.getEnvSelectors(watcherId)
	// if err != nil {
	// 	impl.logger.Errorw("error in getting selectors for the watcher", "watcherId", watcherId, "error", err)
	// 	return err
	// }

	err = impl.informScoops(types.Delete, &types2.WatcherDto{Id: watcherId,
		EventConfiguration: types2.EventConfiguration{
			Selectors: selectors,
		},
	})
	if err != nil {
		impl.logger.Errorw("error in informing respective scoops about this watcher creation", "err", err, "watcherId", watcherId)
		return err
	}

	err = impl.triggerRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing db request in watcher delete request", "watcherId", watcherId, "err", err)
		return err
	}
	return nil
}

func (impl *WatcherServiceImpl) UpdateWatcherById(watcherId int, watcherRequest *types2.WatcherDto, userId int32) error {
	watcher, err := impl.watcherRepository.GetWatcherById(watcherId)
	if err != nil {
		impl.logger.Errorw("error in retrieving watcher by id", "watcherId", watcherId, "error", err)
		return err
	}
	gvks, err := fetchGvksFromK8sResources(watcherRequest.EventConfiguration.K8sResources)
	if err != nil {
		impl.logger.Errorw("error in retrieving gvks", "gvks", watcherRequest.EventConfiguration.K8sResources, "error", err)
		return err
	}
	selectors, err := getSelectorJson(watcherRequest.EventConfiguration.Selectors)
	if err != nil {
		impl.logger.Errorw("error in retrieving watcher by id", watcherId, "error", err)
		return err
	}
	watcher.Name = watcherRequest.Name
	watcher.Description = watcherRequest.Description
	watcher.FilterExpression = watcherRequest.EventConfiguration.EventExpression
	watcher.SelectedActions = watcherRequest.EventConfiguration.SelectedActions
	watcher.Selectors = selectors
	watcher.Gvks = gvks
	watcher.UpdateAuditLog(userId)
	tx, err := impl.triggerRepository.StartTx()
	defer impl.triggerRepository.RollbackTx(tx)
	if err != nil {
		impl.logger.Errorw("error in creating transaction for creating trigger", watcherId, "error", err)
		return err
	}

	err = impl.watcherRepository.Update(tx, watcher)
	if err != nil {
		impl.logger.Errorw("error in updating watcher", "error", err)
		return err
	}

	err = impl.triggerRepository.DeleteTriggerByWatcherId(tx, watcher.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in deleting trigger by watcher id", watcherId, "error", err)
		return err
	}

	err = impl.createTriggerForWatcher(watcherRequest, watcherId, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in creating trigger by watcher id", watcherId, "error", err)
		return err
	}

	watcherRequest.Id = watcher.Id
	err = impl.informScoops(types.UPDATE, watcherRequest)
	if err != nil {
		impl.logger.Errorw("error in informing respective scoops about this watcher creation", "err", err, "watcherRequest", watcherRequest)
		return err
	}
	err = impl.triggerRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to create trigger", "error", err)
		return err
	}

	return nil
}

func indexOfWatcher(watcherID int, watchersList []*repository.K8sEventWatcher) int {
	for i, watcher := range watchersList {
		if watcher.Id == watcherID {
			return i
		}
	}
	return -1
}

type WatcherTriggerData struct {
	time    time.Time
	Trigger *types2.Trigger
	Watcher *repository.K8sEventWatcher
}

func sortByWatcherNameOrder(combinedData []WatcherTriggerData, watchersList []*repository.K8sEventWatcher) []WatcherTriggerData {
	sort.Slice(combinedData, func(i, j int) bool {
		indexI := indexOfWatcher(combinedData[i].Watcher.Id, watchersList)
		indexJ := indexOfWatcher(combinedData[j].Watcher.Id, watchersList)
		return indexI < indexJ
	})
	return combinedData
}

func sortByTime(combinedData []WatcherTriggerData, sortOrder string) {
	less := func(i, j int) bool {
		if sortOrder == "asc" {
			return combinedData[i].time.Before(combinedData[j].time)
		}
		return combinedData[i].time.After(combinedData[j].time)
	}
	sort.Slice(combinedData, less)
}
func (impl *WatcherServiceImpl) FindAllWatchers(params types2.WatcherQueryParams) (types2.WatchersResponse, error) {
	// implemented under assumption of having one trigger for one watcher of type JOB only
	watchers, total, err := impl.watcherRepository.FindAllWatchersByQueryName(params)
	if err != nil {
		impl.logger.Errorw("error in retrieving watchers ", "error", err)
		return types2.WatchersResponse{}, err
	}
	if len(watchers) == 0 {
		return types2.WatchersResponse{
			Size:   params.Size,
			Offset: params.Offset,
			Total:  0,
			List:   []types2.WatcherItem{},
		}, nil
	}
	var watcherIds []int
	watcherData := make(map[int]*repository.K8sEventWatcher)
	for _, watcher := range watchers {
		watcherIds = append(watcherIds, watcher.Id)
		watcherData[watcher.Id] = watcher
	}
	triggers, err := impl.triggerRepository.GetTriggerByWatcherIds(watcherIds)

	if err != nil {
		impl.logger.Errorw("error in retrieving triggers ", "error", err)
		return types2.WatchersResponse{}, err
	}
	triggerIdToWatcherId := make(map[int]int)
	for _, trigger := range triggers {
		triggerIdToWatcherId[trigger.Id] = trigger.WatcherId
	}
	jobPipelineIds, watcherIdToTriggerData, err := impl.getTriggerDataAndPipelineIds(triggers)
	if err != nil {
		impl.logger.Errorw("error in retrieving triggers ", "error", err)
		return types2.WatchersResponse{}, err
	}
	var ciWorkflows []*pipelineConfig.CiWorkflow
	if len(jobPipelineIds) != 0 {
		ciWorkflows, err = impl.ciWorkflowRepository.FindLastOneTriggeredWorkflowByCiIds(jobPipelineIds)
		if err != nil {
			impl.logger.Errorw("error in fetching last triggered workflow by ci ids", jobPipelineIds, "error", err)
			return types2.WatchersResponse{}, err
		}
	}
	pipelineIdToTriggerTime := make(map[int]time.Time)
	for _, workflow := range ciWorkflows {
		pipelineIdToTriggerTime[workflow.CiPipelineId] = workflow.StartedOn
	}
	watcherTriggerData := populateWatcherTriggerData(watcherIds, watcherIdToTriggerData, watcherData, triggerIdToWatcherId, pipelineIdToTriggerTime)
	// introducing column triggeredAt would avoid these sorts
	if params.SortOrderBy == "triggeredAt" {
		sortByTime(watcherTriggerData, params.SortOrder)
	} else {
		sortByWatcherNameOrder(watcherTriggerData, watchers)
	}
	watcherResponses := types2.WatchersResponse{
		Size:   params.Size,
		Offset: params.Offset,
		Total:  total,
	}

	for _, data := range watcherTriggerData {
		if data.Trigger != nil {
			watcherResponses.List = append(watcherResponses.List, types2.WatcherItem{
				Id:              data.Watcher.Id,
				Name:            data.Watcher.Name,
				Description:     data.Watcher.Description,
				TriggeredAt:     data.time,
				JobPipelineName: data.Trigger.Data.PipelineName,
				JobPipelineId:   data.Trigger.Data.PipelineId,
				WorkflowId:      data.Trigger.Data.WorkflowId,
				JobId:           data.Trigger.Data.JobId,
			})
		} else {
			watcherResponses.List = append(watcherResponses.List, types2.WatcherItem{
				Id:          data.Watcher.Id,
				Name:        data.Watcher.Name,
				Description: data.Watcher.Description,
				TriggeredAt: data.time,
			})
		}

	}
	return watcherResponses, nil
}
func (impl *WatcherServiceImpl) getTriggerDataAndPipelineIds(triggers []*repository.AutoRemediationTrigger) ([]int, map[int]types2.Trigger, error) {
	var jobPipelineIds []int
	watcherIdToTriggerData := make(map[int]types2.Trigger)
	for _, trigger := range triggers {
		triggerResp, err := getTriggerDataFromJson(trigger.Data)
		if err != nil {
			impl.logger.Errorw("error in unmarshalling trigger data", "error", err)
			return []int{}, map[int]types2.Trigger{}, err
		}
		jobPipelineIds = append(jobPipelineIds, triggerResp.PipelineId)
		watcherIdToTriggerData[trigger.WatcherId] = types2.Trigger{
			Id:             trigger.Id,
			IdentifierType: types2.TriggerType(trigger.Type),
			Data: types2.TriggerData{
				RuntimeParameters:      triggerResp.RuntimeParameters,
				JobId:                  triggerResp.JobId,
				JobName:                triggerResp.JobName,
				PipelineId:             triggerResp.PipelineId,
				PipelineName:           triggerResp.PipelineName,
				ExecutionEnvironment:   triggerResp.ExecutionEnvironment,
				ExecutionEnvironmentId: triggerResp.ExecutionEnvironmentId,
				WorkflowId:             triggerResp.WorkflowId,
			},
		}
	}
	return jobPipelineIds, watcherIdToTriggerData, nil
}
func populateWatcherTriggerData(watcherIds []int, watcherIdToTriggerData map[int]types2.Trigger, watcherData map[int]*repository.K8sEventWatcher, triggerIdToWatcherId map[int]int, pipelineIdToTriggerTime map[int]time.Time) []WatcherTriggerData {
	var watcherTriggerData []WatcherTriggerData
	for _, id := range watcherIds {
		if _, exists := watcherIdToTriggerData[id]; exists {
			trigger := watcherIdToTriggerData[id]
			watcher := watcherData[triggerIdToWatcherId[trigger.Id]]
			if trigger.Id != 0 {
				watcherTriggerData = append(watcherTriggerData, WatcherTriggerData{
					time:    pipelineIdToTriggerTime[trigger.Data.PipelineId],
					Trigger: &trigger,
					Watcher: watcher,
				})
			} else {
				watcherTriggerData = append(watcherTriggerData, WatcherTriggerData{
					Trigger: nil,
					Watcher: watcher,
				})
			}
		}
	}
	return watcherTriggerData
}
func (impl *WatcherServiceImpl) GetTriggerByWatcherIds(watcherIds []int) ([]*types2.Trigger, error) {
	triggers, err := impl.triggerRepository.GetTriggerByWatcherIds(watcherIds)
	if err != nil {
		impl.logger.Errorw("error in getting triggers by watcher ids", "watcherIds", watcherIds, "err", err)
		return nil, err
	}

	triggersResult := make([]*types2.Trigger, 0, len(triggers))
	for _, trigger := range triggers {
		triggerResp := types2.Trigger{}
		triggerResp.Id = trigger.Id
		triggerResp.IdentifierType = types2.TriggerType(trigger.Type)
		triggerData := types2.TriggerData{}
		if err := json.Unmarshal([]byte(trigger.Data), &triggerData); err != nil {
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
		triggerResp.WatcherId = trigger.WatcherId
		triggersResult = append(triggersResult, &triggerResp)
	}

	return triggersResult, nil
}

func (impl *WatcherServiceImpl) getEnvsMap(envs []string) (map[string]*repository2.Environment, error) {
	envObjs, err := impl.environmentRepository.GetWithClusterByNames(envs)
	if err != nil {
		impl.logger.Errorw("error in finding envs with envNames", "envNames", envs, "err", err)
		return nil, err
	}

	envsMap := make(map[string]*repository2.Environment)
	for _, envObj := range envObjs {
		envsMap[envObj.Name] = envObj
	}
	return envsMap, nil
}

func (impl *WatcherServiceImpl) getEnvSelectors(watcherId int) ([]types2.Selector, map[string]*repository2.Environment, error) {
	mappings, err := impl.resourceQualifierMappingService.GetQualifierMappingsByResourceId(watcherId, resourceQualifiers.K8sEventWatcher)
	if err != nil {
		return nil, nil, err
	}

	envNames := make([]string, 0, len(mappings))
	for _, mapping := range mappings {
		// currently assuming all the mappings are of identifier type environment
		envNames = append(envNames, mapping.IdentifierValueString)
	}
	var envs []*repository2.Environment
	envsMap := make(map[string]*repository2.Environment)
	if len(envNames) != 0 {
		envs, err = impl.environmentRepository.GetWithClusterByNames(envNames)
		if err != nil {
			return nil, nil, err
		}
		for _, envObj := range envs {
			envsMap[envObj.Name] = envObj
		}
	}
	clusterWiseEnvs := make(map[string][]string)
	for _, env := range envs {
		list, ok := clusterWiseEnvs[env.Cluster.ClusterName]
		if !ok {
			list = make([]string, 0)
		}
		list = append(list, env.Name)
		clusterWiseEnvs[env.Cluster.ClusterName] = list
	}

	selectors := make([]types2.Selector, 0, len(clusterWiseEnvs))
	for clusterName, _ := range clusterWiseEnvs {
		selectors = append(selectors, types2.Selector{
			Type:      types2.EnvironmentSelector,
			GroupName: clusterName,
			Names:     clusterWiseEnvs[clusterName],
		})
	}
	return selectors, envsMap, nil
}
func (impl *WatcherServiceImpl) RetrieveInterceptedEvents(params *types2.InterceptedEventQueryParams) (*types2.InterceptedResponse, error) {

	interceptedEventData, total, err := impl.interceptedEventsRepository.FindAllInterceptedEvents(params)
	if err != nil {
		impl.logger.Errorw("error in retrieving intercepted events", "err", err)
		return &types2.InterceptedResponse{}, err
	}
	if len(interceptedEventData) == 0 {
		return &types2.InterceptedResponse{
			Size:   params.Size,
			Offset: params.Offset,
			Total:  0,
			List:   []types2.InterceptedEventsDto{},
		}, nil
	}

	clusterIds := util.Map(interceptedEventData, func(event *types2.InterceptedEventData) int {
		return event.ClusterId
	})

	clusterIdToClusterNameMap, err := impl.getClusterInfoOfParamsCluster(clusterIds)
	if err != nil {
		impl.logger.Errorw("error in retrieving clusters ", "err", err)
		return &types2.InterceptedResponse{}, err
	}

	interceptedEvents, triggerExecutionIds, err := populateInterceptedEventsAndFetchExecutionIds(interceptedEventData, clusterIdToClusterNameMap)
	if err != nil {
		impl.logger.Errorw("error in populating intercepted events and fetching execution ids ", "err", err)
		return nil, err
	}
	interceptedEvents, err = impl.overrideInterceptedEventsStatus(interceptedEvents, triggerExecutionIds)
	if err != nil {
		impl.logger.Errorw("error in overriding intercepted event's status with workflow's status ", "err", err)
		return nil, err
	}
	interceptedResponse := types2.InterceptedResponse{
		Offset: params.Offset,
		Size:   params.Size,
		Total:  total,
		List:   interceptedEvents,
	}
	return &interceptedResponse, nil
}

func populateInterceptedEventsAndFetchExecutionIds(interceptedEventData []*types2.InterceptedEventData, clusterIdToClusterName map[int]string) ([]types2.InterceptedEventsDto, []int, error) {
	var interceptedEvents []types2.InterceptedEventsDto
	var triggerExecutionIds []int
	for _, event := range interceptedEventData {
		interceptedEvent := types2.InterceptedEventsDto{
			InterceptedEventId: event.InterceptedEventId,
			Action:             event.Action,
			InvolvedObjects:    event.InvolvedObjects,
			Metadata:           event.Metadata,
			ClusterName:        clusterIdToClusterName[event.ClusterId],
			ClusterId:          event.ClusterId,
			Namespace:          event.Namespace,
			EnvironmentName:    event.Environment,
			WatcherName:        event.WatcherName,
			InterceptedTime:    (event.InterceptedAt).String(),
			ExecutionStatus:    types2.Status(event.Status),
			TriggerId:          event.TriggerId,
			TriggerExecutionId: event.TriggerExecutionId,
			ExecutionMessage:   event.ExecutionMessage,
		}
		triggerData, err := getTriggerDataFromJson(event.TriggerData)
		if err != nil {
			return []types2.InterceptedEventsDto{}, []int{}, err
		}
		interceptedEvent.Trigger = types2.Trigger{
			Id:             event.TriggerId,
			IdentifierType: types2.TriggerType(event.TriggerType),
			Data:           triggerData,
		}
		if interceptedEvent.Trigger.IdentifierType == types2.DEVTRON_JOB {
			triggerExecutionIds = append(triggerExecutionIds, event.TriggerExecutionId)
		}
		interceptedEvents = append(interceptedEvents, interceptedEvent)
	}
	return interceptedEvents, triggerExecutionIds, nil
}
func (impl *WatcherServiceImpl) overrideInterceptedEventsStatus(interceptedEvents []types2.InterceptedEventsDto, triggerExecutionIds []int) ([]types2.InterceptedEventsDto, error) {
	triggerExecutionIdToStatus, err := impl.getStatusForJobs(triggerExecutionIds)
	if err != nil {
		impl.logger.Errorw("error in fetching status from ci workflows", "err", err)
		return []types2.InterceptedEventsDto{}, err
	}
	for i := range interceptedEvents {
		interceptedEvent := &interceptedEvents[i]
		status := triggerExecutionIdToStatus[interceptedEvent.TriggerExecutionId]
		if status != "" {
			interceptedEvent.ExecutionStatus = types2.Status(status)
		}
	}
	return interceptedEvents, nil
}

func (impl *WatcherServiceImpl) getClusterInfoOfParamsCluster(clusters []int) (map[int]string, error) {
	var clustersFetched []repository2.Cluster
	if len(clusters) != 0 {
		var err error
		clustersFetched, err = impl.clusterRepository.FindByIds(clusters)
		if err != nil {
			impl.logger.Errorw("error in retrieving clusters ", "err", err)
			return nil, err
		}
	}

	var clusterId []int
	clusterIdToClusterName := make(map[int]string)
	for _, cluster := range clustersFetched {
		clusterId = append(clusterId, cluster.Id)
		clusterIdToClusterName[cluster.Id] = cluster.ClusterName
	}
	return clusterIdToClusterName, nil
}

func (impl *WatcherServiceImpl) getStatusForJobs(triggerExecutionIds []int) (map[int]string, error) {
	ciWorkflows, err := impl.ciWorkflowRepository.FindCiWorkflowGitTriggersByIds(triggerExecutionIds) // function should have been FindCiWorkflowByIds instead of FindCiWorkflowGitTriggersByIds
	if err != nil {
		impl.logger.Errorw("error in getting ci workflows", "err", err)
		return nil, err
	}
	triggerExecutionIdToStatus := make(map[int]string)
	for _, workflow := range ciWorkflows {
		triggerExecutionIdToStatus[workflow.Id] = workflow.Status
	}
	return triggerExecutionIdToStatus, nil
}

func (impl *WatcherServiceImpl) GetWatchersByClusterId(clusterId int) ([]*types.Watcher, error) {

	cluster, err := impl.clusterRepository.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting cluster by clusterId", "clusterId", clusterId, "err", err)
		return nil, err
	}
	watchers, _, err := impl.watcherRepository.FindAllWatchersByQueryName(types2.WatcherQueryParams{})
	if err != nil {
		impl.logger.Errorw("error in getting watchers ", "err", err)
		return nil, err
	}

	watchersResponse := make([]*types.Watcher, 0, len(watchers))
	for _, watcher := range watchers {
		selectors, err := getSelectorsFromJson(watcher.Selectors)
		if err != nil {
			impl.logger.Errorw("error in unmarshalling selectors string ", "gvk", watcher.Gvks, "err", err)
			continue
		}

		selector := getClusterSelector(cluster.ClusterName, selectors)
		if selector == nil {
			continue
		}

		k8sResources, err := getK8sResourcesFromGvks(watcher.Gvks)
		if err != nil {
			impl.logger.Errorw("error in unmarshalling gvk string ", "gvk", watcher.Gvks, "err", err)
			continue
		}
		watcherResp := &types.Watcher{
			Id:                    watcher.Id,
			Name:                  watcher.Name,
			EventFilterExpression: watcher.FilterExpression,
			// Namespaces:            nsMap,
			GVKs: util.Map(k8sResources, func(k8Resource *types2.K8sResource) schema.GroupVersionKind {
				return k8Resource.GetGVK()
			}),
			SelectedActions: watcher.SelectedActions,
			Selectors:       types2.GetNamespaceSelector(*selector),
			ClusterId:       cluster.Id,
		}

		watchersResponse = append(watchersResponse, watcherResp)
	}

	return watchersResponse, nil
}

func (impl *WatcherServiceImpl) informScoops(action types.Action, watcherRequest *types2.WatcherDto) error {
	clusterNames := make([]string, 0)
	allClusters := false
	for _, selector := range watcherRequest.EventConfiguration.Selectors {
		if selector.GroupName == types2.AllClusterGroup {
			allClusters = true
			break
		}
		clusterNames = append(clusterNames, selector.GroupName)

	}

	clusterMap := make(map[string]*repository2.Cluster)
	if allClusters {
		clusters, err := impl.clusterRepository.FindAll()
		if err != nil {
			impl.logger.Errorw("error in getting all active clusters")
			return err
		}

		for i, cluster := range clusters {
			clusterMap[cluster.ClusterName] = &clusters[i]
		}
	}

	if len(clusterNames) > 0 {
		clusters, err := impl.clusterRepository.FindByNames(clusterNames)
		if err != nil {
			impl.logger.Errorw("error in getting all active clusters")
			return err
		}

		for i, cluster := range clusters {
			clusterMap[cluster.ClusterName] = clusters[i]
		}
	}

	triggerConfigured := false
	if len(watcherRequest.Triggers) > 0 {
		triggerConfigured = true
	}

	for clusterName, cluster := range clusterMap {

		selector := getClusterSelector(clusterName, watcherRequest.EventConfiguration.Selectors)
		// this case should never happen, just a safety check
		if selector == nil {
			continue
		}
		namespaceSelector := types2.GetNamespaceSelector(*selector)

		watcher := &types.Watcher{
			Id:                    watcherRequest.Id,
			Name:                  watcherRequest.Name,
			GVKs:                  watcherRequest.EventConfiguration.GetK8sResources(),
			EventFilterExpression: watcherRequest.EventConfiguration.EventExpression,
			// Namespaces:            nsMap,
			SelectedActions: watcherRequest.EventConfiguration.SelectedActions,
			JobConfigured:   triggerConfigured,
			ClusterId:       cluster.Id,
			Selectors:       namespaceSelector,
		}

		port, scoopConfig, err := impl.k8sApplicationService.GetScoopPort(context.Background(), cluster.Id)
		if err != nil && errors.Is(err, application.ScoopNotConfiguredErr) {
			impl.logger.Errorw("error in informing to scoop", "clusterId", cluster.Id, "scoopConfig", scoopConfig, "err", err)
			// not returning the error as we have to continue updating other scoops
			continue
		}
		scoopUrl := fmt.Sprintf("http://127.0.0.1:%d", port)
		scoopClient, _ := scoopClient.NewScoopClientImpl(impl.logger, scoopUrl, scoopConfig.PassKey)
		err = scoopClient.UpdateWatcherConfig(context.Background(), action, watcher)
		if err != nil {
			impl.logger.Errorw("error in informing to scoop by a REST call", "watcher", watcher, "action", action, "scoopUrl", scoopUrl, "err", err)
			// not returning the error as we have to continue updating other scoops
			continue
		}
	}

	return nil
}
