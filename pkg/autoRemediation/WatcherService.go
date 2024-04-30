package autoRemediation

import (
	"context"
	"fmt"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/autoRemediation/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/k8s/application"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util"
	scoopClient "github.com/devtron-labs/scoop/client"
	"github.com/devtron-labs/scoop/pkg/watcherEvents"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"gopkg.in/square/go-jose.v2/json"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sort"
	"time"
)

type WatcherService interface {
	CreateWatcher(watcherRequest *WatcherDto, userId int32) (int, error)
	GetWatcherById(watcherId int) (*WatcherDto, error)
	DeleteWatcherById(watcherId int, userId int32) error
	UpdateWatcherById(watcherId int, watcherRequest *WatcherDto, userId int32) error
	RetrieveInterceptedEvents(params repository.InterceptedEventQueryParams) (*InterceptedResponse, error)
	FindAllWatchers(offset int, search string, size int, sortOrder string, sortOrderBy string) (WatchersResponse, error)
	GetTriggerByWatcherIds(watcherIds []int) ([]*Trigger, error)

	GetWatchersByClusterId(clusterId int) ([]*watcherEvents.Watcher, error)
}

type WatcherServiceImpl struct {
	watcherRepository               repository.WatcherRepository
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

func NewWatcherServiceImpl(watcherRepository repository.WatcherRepository,
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
	logger *zap.SugaredLogger) *WatcherServiceImpl {

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

func (impl *WatcherServiceImpl) CreateWatcher(watcherRequest *WatcherDto, userId int32) (int, error) {

	gvks, err := fetchGvksFromK8sResources(watcherRequest.EventConfiguration.K8sResources)
	if err != nil {
		impl.logger.Errorw("error in fetching gvks", "error", err)
		return 0, err
	}
	watcher := &repository.Watcher{
		Name:             watcherRequest.Name,
		Description:      watcherRequest.Description,
		FilterExpression: watcherRequest.EventConfiguration.EventExpression,
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

	envs, err := impl.getEnvsMap(watcherRequest.EventConfiguration.getEnvsFromSelectors())
	if err != nil {
		impl.logger.Errorw("error in getting envs using selectors", "err", err)
		return 0, err
	}
	envSelectionIdentifiers := getEnvSelectionIdentifiers(envs)
	err = impl.resourceQualifierMappingService.CreateMappings(tx, userId, resourceQualifiers.K8sEventWatcher, []int{watcher.Id}, resourceQualifiers.EnvironmentSelector, envSelectionIdentifiers)
	if err != nil {
		impl.logger.Errorw("error in mapping watchers to the given envs", "watcher", watcher, "envSelectionIdentifiers", envSelectionIdentifiers, "err", err)
		return 0, err
	}

	watcherRequest.Id = watcher.Id
	err = impl.informScoops(envs, watcherEvents.ADD, watcherRequest)
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

func fetchGvksFromK8sResources(resources []*K8sResource) (string, error) {
	gvks, err := json.Marshal(resources)
	if err != nil {
		return "", err
	}
	return string(gvks), nil
}

func (impl *WatcherServiceImpl) createTriggerForWatcher(watcherRequest *WatcherDto, watcherId int, userId int32, tx *pg.Tx) error {
	var triggersJobsForWatcher []*Trigger
	if watcherRequest.Triggers == nil {
		return nil
	}
	for i, _ := range watcherRequest.Triggers {
		if watcherRequest.Triggers[i].IdentifierType == repository.DEVTRON_JOB {
			triggersJobsForWatcher = append(triggersJobsForWatcher, &watcherRequest.Triggers[i])
		}
	}
	err := impl.createTriggerJobsForWatcher(triggersJobsForWatcher, watcherId, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in creating triggers for watcher", "error", err)
		return err
	}
	return nil
}

type jobDetails struct {
	displayNameToId         map[string]int
	pipelineNameToId        map[string]int
	envNameToId             map[string]int
	pipelineIdtoAppworkflow map[int]int
}

func (impl *WatcherServiceImpl) createTriggerJobsForWatcher(triggers []*Trigger, watcherId int, userId int32, tx *pg.Tx) error {
	jobInfo, err := impl.getJobEnvPipelineDetailsForWatcher(triggers)
	if err != nil {
		impl.logger.Errorw("error in retrieving details of trigger type job", "error", err)
		return err
	}
	var triggersResult []*repository.Trigger
	for _, res := range triggers {
		triggerData := TriggerData{
			RuntimeParameters:      res.Data.RuntimeParameters,
			JobId:                  jobInfo.displayNameToId[res.Data.JobName],
			JobName:                res.Data.JobName,
			PipelineId:             jobInfo.pipelineNameToId[res.Data.PipelineName],
			PipelineName:           res.Data.PipelineName,
			ExecutionEnvironment:   res.Data.ExecutionEnvironment,
			ExecutionEnvironmentId: jobInfo.envNameToId[res.Data.ExecutionEnvironment],
			WorkflowId:             jobInfo.pipelineIdtoAppworkflow[jobInfo.pipelineNameToId[res.Data.PipelineName]],
		}
		jsonData, err := json.Marshal(triggerData)
		if err != nil {
			impl.logger.Errorw("error in marshalling trigger data ", "error", err)
			return err
		}
		triggerRes := &repository.Trigger{
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

func (impl *WatcherServiceImpl) getJobEnvPipelineDetailsForWatcher(triggers []*Trigger) (*jobDetails, error) {
	var jobsDetails *jobDetails
	var jobNames, envNames, pipelineNames []string

	for _, trig := range triggers {
		jobNames = append(jobNames, trig.Data.JobName)
		envNames = append(envNames, trig.Data.ExecutionEnvironment)
		pipelineNames = append(pipelineNames, trig.Data.PipelineName)
	}
	apps, err := impl.appRepository.FetchAppByDisplayNamesForJobs(jobNames)
	if err != nil {
		impl.logger.Errorw("error in fetching apps", "jobNames", jobNames, "error", err)
		return jobsDetails, err
	}
	var jobIds []int
	for _, app := range apps {
		jobIds = append(jobIds, app.Id)
	}
	pipelines, err := impl.ciPipelineRepository.FindByNames(pipelineNames, jobIds)
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines", "pipelineNames", pipelineNames, "error", err)
		return jobsDetails, err
	}
	var pipelinesId []int
	for _, pipeline := range pipelines {
		pipelinesId = append(pipelinesId, pipeline.Id)
	}
	envs, err := impl.environmentRepository.FindByNames(envNames)
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "envNames", envNames, "error", err)
		return jobsDetails, err
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
		impl.logger.Errorw("error in retrieving workflows for pipelineIds", pipelinesId, "error", err)
		return jobsDetails, err
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
func (impl *WatcherServiceImpl) GetWatcherById(watcherId int) (*WatcherDto, error) {
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
	watcherResponse := WatcherDto{
		Name:        watcher.Name,
		Description: watcher.Description,
		EventConfiguration: EventConfiguration{
			K8sResources:    k8sResources,
			EventExpression: watcher.FilterExpression,
		},
	}

	triggers, err := impl.triggerRepository.GetTriggerByWatcherId(watcherId)
	if err != nil {
		impl.logger.Errorw("error in getting triggers for watcher id", "watcherId", watcherId, "error", err)
		return &WatcherDto{}, err
	}
	if len(triggers) == 0 {
		watcherResponse.Triggers = []Trigger{}
	}
	for _, trigger := range triggers {
		triggerResp, err := impl.getTriggerDataFromJson(trigger.Data)
		if err != nil {
			impl.logger.Errorw("error in unmarshalling trigger data", "error", err)
			return nil, err
		}
		triggerResponse := Trigger{
			Id:             trigger.Id,
			IdentifierType: trigger.Type,
			Data:           triggerResp,
		}
		watcherResponse.Triggers = append(watcherResponse.Triggers, triggerResponse)
	}

	selectors, err := impl.getEnvSelectors(watcherId)
	if err != nil {
		impl.logger.Errorw("error in getting selectors for the watcher", "watcherId", watcherId, "error", err)
		return nil, err
	}

	watcherResponse.EventConfiguration.Selectors = selectors
	return &watcherResponse, nil

}

func (impl *WatcherServiceImpl) getTriggerDataFromJson(data string) (TriggerData, error) {
	var triggerResp TriggerData
	if err := json.Unmarshal([]byte(data), &triggerResp); err != nil {
		impl.logger.Errorw("error in unmarshalling trigger data", "error", err)
		return TriggerData{}, err
	}
	return triggerResp, nil
}

func getK8sResourcesFromGvks(gvks string) ([]*K8sResource, error) {
	var k8sResources []*K8sResource
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

	err = impl.resourceQualifierMappingService.DeleteAllQualifierMappingsByResourceTypeAndId(resourceQualifiers.K8sEventWatcher, watcherId, sql.NewDefaultAuditLog(userId), tx)
	if err != nil {
		impl.logger.Errorw("error in envs mappings for the watcher", "watcherId", watcherId, "err", err)
		return err
	}

	// err = impl.informScoops(envs, watcherRequest)
	// if err != nil {
	// 	impl.logger.Errorw("error in informing respective scoops about this watcher creation", "err", err, "watcherRequest", watcherRequest)
	// 	return err
	// }

	err = impl.triggerRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing db request in watcher delete request", "watcherId", watcherId, "err", err)
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
	watcher.Description = watcherRequest.Description
	watcher.FilterExpression = watcherRequest.EventConfiguration.EventExpression
	watcher.Gvks = gvks
	watcher.AuditLog = sql.NewDefaultAuditLog(userId)
	tx, err := impl.triggerRepository.StartTx()
	if err != nil {
		impl.logger.Errorw("error in creating transaction for creating trigger", watcherId, "error", err)
		return err
	}

	err = impl.watcherRepository.Update(tx, watcher, userId)
	if err != nil {
		impl.logger.Errorw("error in updating watcher", "error", err)
		return err
	}

	err = impl.triggerRepository.DeleteTriggerByWatcherId(tx, watcher.Id)
	if err != nil {
		impl.logger.Errorw("error in deleting trigger by watcher id", watcherId, "error", err)
		return err
	}

	err = impl.createTriggerForWatcher(watcherRequest, watcherId, userId, tx)
	if err != nil {
		impl.logger.Errorw("error in creating trigger by watcher id", watcherId, "error", err)
		return err
	}

	err = impl.resourceQualifierMappingService.DeleteAllQualifierMappingsByResourceTypeAndId(resourceQualifiers.K8sEventWatcher, watcherId, sql.NewDefaultAuditLog(userId), tx)
	if err != nil {
		impl.logger.Errorw("error in envs mappings for the watcher", "watcherId", watcherId, "err", err)
		return err
	}

	envs, err := impl.getEnvsMap(watcherRequest.EventConfiguration.getEnvsFromSelectors())
	if err != nil {
		impl.logger.Errorw("error in getting envs using env names", "envNames", envs, "err", err)
		return err
	}

	envSelectionIdentifiers := getEnvSelectionIdentifiers(envs)
	err = impl.resourceQualifierMappingService.CreateMappings(tx, userId, resourceQualifiers.K8sEventWatcher, []int{watcher.Id}, resourceQualifiers.EnvironmentSelector, envSelectionIdentifiers)
	if err != nil {
		impl.logger.Errorw("error in mapping watchers to the given envs", "watcher", watcher, "envSelectionIdentifiers", envSelectionIdentifiers, "err", err)
		return err
	}
	watcherRequest.Id = watcher.Id
	err = impl.informScoops(envs, watcherEvents.UPDATE, watcherRequest)
	if err != nil {
		impl.logger.Errorw("error in informing respective scoops about this watcher creation", "err", err, "watcherRequest", watcherRequest)
		return err
	}
	err = impl.triggerRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to create trigger", "error", err)
		return err
	}
	defer impl.triggerRepository.RollbackTx(tx)
	return nil
}

func indexOfWatcher(watcherID int, watchersList []*repository.Watcher) int {
	for i, watcher := range watchersList {
		if watcher.Id == watcherID {
			return i
		}
	}
	return -1
}

type CombinedData struct {
	time    time.Time
	Trigger *Trigger
	Watcher *repository.Watcher
}

func sortByWatcher(combinedData []CombinedData, watchersList []*repository.Watcher) []CombinedData {
	sort.Slice(combinedData, func(i, j int) bool {
		indexI := indexOfWatcher(combinedData[i].Watcher.Id, watchersList)
		indexJ := indexOfWatcher(combinedData[j].Watcher.Id, watchersList)
		return indexI < indexJ
	})
	return combinedData
}

// Define a function to check if a watcher is already present in combinedData
func watcherExists(watcher *repository.Watcher, combinedData []CombinedData) bool {
	for _, data := range combinedData {
		if data.Watcher.Id == watcher.Id {
			return true
		}
	}
	return false
}

func (impl *WatcherServiceImpl) FindAllWatchers(offset int, search string, size int, sortOrder string, sortOrderBy string) (WatchersResponse, error) {
	//TODO : implemented under assumption of having one trigger for one watcher of type JOB only
	params := repository.WatcherQueryParams{
		Offset:      offset,
		Size:        size,
		Search:      search,
		SortOrderBy: sortOrderBy,
		SortOrder:   sortOrder,
	}
	watchers, total, err := impl.watcherRepository.FindAllWatchersByQueryName(params)
	if err != nil {
		impl.logger.Errorw("error in retrieving watchers ", "error", err)
		return WatchersResponse{}, err
	}
	if len(watchers) == 0 {
		return WatchersResponse{
			Size:   size,
			Offset: offset,
			Total:  0,
			List:   []WatcherItem{},
		}, nil
	}
	var watcherIds []int
	watcherData := make(map[int]*repository.Watcher)
	for _, watcher := range watchers {
		watcherIds = append(watcherIds, watcher.Id)
		watcherData[watcher.Id] = watcher
	}
	triggers, err := impl.triggerRepository.GetTriggerByWatcherIds(watcherIds)

	if err != nil {
		impl.logger.Errorw("error in retrieving triggers ", "error", err)
		return WatchersResponse{}, err
	}
	triggerIdWatcherId := make(map[int]int)
	for _, trigger := range triggers {
		triggerIdWatcherId[trigger.Id] = trigger.WatcherId
	}
	var jobPipelineIds []int
	// watcherData := make(map[int]repository.Watcher)
	triggerDto := make(map[int]Trigger)
	for _, trigger := range triggers {
		triggerResp, err := impl.getTriggerDataFromJson(trigger.Data)
		if err != nil {
			impl.logger.Errorw("error in unmarshalling trigger data", "error", err)
			return WatchersResponse{}, err
		}
		jobPipelineIds = append(jobPipelineIds, triggerResp.PipelineId)
		triggerDto[triggerResp.PipelineId] = Trigger{
			Id:             trigger.Id,
			IdentifierType: trigger.Type,
			Data: TriggerData{
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
	var ciWorkflows []*pipelineConfig.CiWorkflow
	if len(jobPipelineIds) != 0 {
		ciWorkflows, err = impl.ciWorkflowRepository.FindLastOneTriggeredWorkflowByCiIds(jobPipelineIds, sortOrder)
		if err != nil {
			impl.logger.Errorw("error in fetching last triggered workflow by ci ids", jobPipelineIds, "error", err)
			return WatchersResponse{}, err
		}
	}

	// var sortedPipe []int
	var combinedData []CombinedData
	for _, workflow := range ciWorkflows {
		trigger := triggerDto[workflow.CiPipelineId]
		watcher := watcherData[triggerIdWatcherId[trigger.Id]]
		combinedData = append(combinedData, CombinedData{
			time:    workflow.StartedOn,
			Trigger: &trigger,
			Watcher: watcher,
		})
	}
	for _, watcher := range watchers {
		if !watcherExists(watcher, combinedData) {
			combinedData = append(combinedData, CombinedData{
				Trigger: nil,
				Watcher: watcher,
			})
		}
	}
	if sortOrderBy == "name" {
		combinedData = sortByWatcher(combinedData, watchers)
	}
	watcherResponses := WatchersResponse{
		Size:   params.Size,
		Offset: params.Offset,
		Total:  total,
	}

	for _, cd := range combinedData {
		if cd.Trigger != nil {
			watcherResponses.List = append(watcherResponses.List, WatcherItem{
				Id:              cd.Watcher.Id,
				Name:            cd.Watcher.Name,
				Description:     cd.Watcher.Description,
				TriggeredAt:     cd.time,
				JobPipelineName: cd.Trigger.Data.PipelineName,
				JobPipelineId:   cd.Trigger.Data.PipelineId,
				WorkflowId:      cd.Trigger.Data.WorkflowId,
				JobId:           cd.Trigger.Data.JobId,
			})
		} else {
			watcherResponses.List = append(watcherResponses.List, WatcherItem{
				Id:          cd.Watcher.Id,
				Name:        cd.Watcher.Name,
				Description: cd.Watcher.Description,
				TriggeredAt: cd.time,
			})
		}

	}
	return watcherResponses, nil
}

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

func (impl *WatcherServiceImpl) getEnvSelectors(watcherId int) ([]Selector, error) {
	mappings, err := impl.resourceQualifierMappingService.GetQualifierMappingsByResourceId(watcherId, resourceQualifiers.K8sEventWatcher)
	if err != nil {
		return nil, err
	}

	envNames := make([]string, 0, len(mappings))
	for _, mapping := range mappings {
		// currently assuming all the mappings are of identifier type environment
		envNames = append(envNames, mapping.IdentifierValueString)
	}
	var envs []*repository2.Environment
	if len(envNames) != 0 {
		envs, err = impl.environmentRepository.GetWithClusterByNames(envNames)
		if err != nil {
			return nil, err
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

	selectors := make([]Selector, 0, len(clusterWiseEnvs))
	for clusterName, _ := range clusterWiseEnvs {
		selectors = append(selectors, Selector{
			Type:      EnvironmentSelector,
			GroupName: clusterName,
			Names:     clusterWiseEnvs[clusterName],
		})
	}
	return selectors, nil
}
func (impl WatcherServiceImpl) RetrieveInterceptedEvents(params repository.InterceptedEventQueryParams) (*InterceptedResponse, error) {
	var clustersFetched []*repository2.Cluster
	if len(params.Clusters) != 0 {
		var err error
		clustersFetched, err = impl.clusterRepository.FindByNames(params.Clusters)
		if err != nil {
			impl.logger.Errorw("error in retrieving clusters ", "err", err)
			return &InterceptedResponse{}, err
		}
	}

	var clusterId []int
	for _, cluster := range clustersFetched {
		clusterId = append(clusterId, cluster.Id)
	}
	parameters := repository.InterceptedEventQuery{
		ClusterIds:      clusterId,
		Offset:          params.Offset,
		Size:            params.Size,
		SortOrder:       params.SortOrder,
		SearchString:    params.SearchString,
		From:            params.From,
		To:              params.To,
		Watchers:        params.Watchers,
		Namespaces:      params.Namespaces,
		ExecutionStatus: params.ExecutionStatus,
	}
	interceptedEventData, total, err := impl.interceptedEventsRepository.FindAllInterceptedEvents(&parameters)
	if err != nil {
		impl.logger.Errorw("error in retrieving intercepted events", "err", err)
		return &InterceptedResponse{}, err
	}
	if len(interceptedEventData) == 0 {
		return &InterceptedResponse{
			Size:   params.Size,
			Offset: params.Offset,
			Total:  0,
			List:   []InterceptedEventsDto{},
		}, nil
	}
	var clusterIds []int
	for _, event := range interceptedEventData {
		clusterIds = append(clusterIds, event.ClusterId)
	}
	var clusters []repository2.Cluster
	if len(clusterIds) != 0 {
		clusters, err = impl.clusterRepository.FindByIds(clusterIds)
		if err != nil {
			impl.logger.Errorw("error in retrieving cluster names ", "err", err)
			return &InterceptedResponse{}, err
		}
	}
	clusterIdtoName := make(map[int]string)
	for _, cluster := range clusters {
		clusterIdtoName[cluster.Id] = cluster.ClusterName
	}
	var interceptedEvents []InterceptedEventsDto
	for _, event := range interceptedEventData {
		interceptedEvent := InterceptedEventsDto{
			MessageType:        event.MessageType,
			Message:            event.Message,
			Event:              event.Event,
			InvolvedObject:     event.InvolvedObject,
			ClusterName:        clusterIdtoName[event.ClusterId],
			ClusterId:          event.ClusterId,
			Namespace:          event.Namespace,
			WatcherName:        event.WatcherName,
			InterceptedTime:    (event.InterceptedAt).String(),
			ExecutionStatus:    event.Status,
			TriggerId:          event.TriggerId,
			TriggerExecutionId: event.TriggerExecutionId,
		}
		var triggerData TriggerData
		if err := json.Unmarshal([]byte(event.TriggerData), &triggerData); err != nil {
			impl.logger.Errorw("error in unmarshalling trigger data", "error", err)
			return nil, err
		}
		interceptedEvent.EnvironmentName = triggerData.ExecutionEnvironment
		interceptedEvent.Trigger = Trigger{
			Id:             event.TriggerId,
			IdentifierType: event.TriggerType,
			Data:           triggerData,
		}
		interceptedEvents = append(interceptedEvents, interceptedEvent)
	}
	interceptedResponse := InterceptedResponse{
		Offset: params.Offset,
		Size:   params.Size,
		Total:  total,
		List:   interceptedEvents,
	}
	return &interceptedResponse, nil
}

func (impl *WatcherServiceImpl) GetWatchersByClusterId(clusterId int) ([]*watcherEvents.Watcher, error) {
	mappings, err := impl.resourceQualifierMappingService.GetQualifierMappingsByResourceType(resourceQualifiers.K8sEventWatcher)
	if err != nil {
		impl.logger.Errorw("error in getting watchers by clusterId", "clusterId", clusterId, "err", err)
		return nil, err
	}

	watcherEnvMap := make(map[int][]string)
	envNames := util.Map(mappings, func(mapping *resourceQualifiers.QualifierMapping) string {
		envIds := watcherEnvMap[mapping.ResourceId]
		if envIds == nil {
			envIds = make([]string, 0)
		}
		envIds = append(envIds, mapping.IdentifierValueString)
		watcherEnvMap[mapping.ResourceId] = envIds

		return mapping.IdentifierValueString
	})

	watcherIds := maps.Keys(watcherEnvMap)
	watchers, err := impl.watcherRepository.GetWatcherByIds(watcherIds)
	if err != nil {
		impl.logger.Errorw("error in getting watchers by watcherIds", "watcherIds", watcherIds, "err", err)
		return nil, err
	}

	envMap, err := impl.getEnvsMap(envNames)
	if err != nil {
		impl.logger.Errorw("error in getting environment details by env names", "envNames", envNames, "err", err)
		return nil, err
	}

	watchersResponse := make([]*watcherEvents.Watcher, 0, len(watchers))
	for _, watcher := range watchers {
		nsMap := make(map[string]bool)
		for _, envId := range watcherEnvMap[watcher.Id] {
			if env, ok := envMap[envId]; ok {
				nsMap[env.Namespace] = true
			}
		}

		k8sResources, err := getK8sResourcesFromGvks(watcher.Gvks)
		if err != nil {
			impl.logger.Errorw("error in unmarshalling gvk string ", "gvk", watcher.Gvks, "err", err)
			continue
		}
		watcherResp := &watcherEvents.Watcher{
			Id:                    watcher.Id,
			Name:                  watcher.Name,
			EventFilterExpression: watcher.FilterExpression,
			Namespaces:            nsMap,
			GVKs: util.Map(k8sResources, func(k8Resource *K8sResource) schema.GroupVersionKind {
				return k8Resource.GetGVK()
			}),
		}

		watchersResponse = append(watchersResponse, watcherResp)
	}

	return watchersResponse, nil
}

func (impl *WatcherServiceImpl) informScoops(envsMap map[string]*repository2.Environment, action watcherEvents.Action, watcherRequest *WatcherDto) error {
	clusterEnvMap := make(map[int][]*repository2.Environment)
	for _, env := range envsMap {
		namespaces := clusterEnvMap[env.ClusterId]
		if namespaces == nil {
			namespaces = make([]*repository2.Environment, 0)
		}
		namespaces = append(namespaces, env)
		clusterEnvMap[env.ClusterId] = namespaces
	}

	for clusterId, envDetails := range clusterEnvMap {
		nsMap := make(map[string]bool)
		for _, env := range envDetails {
			nsMap[env.Namespace] = true
		}

		watcher := &watcherEvents.Watcher{
			Id:                    watcherRequest.Id,
			Name:                  watcherRequest.Name,
			GVKs:                  watcherRequest.EventConfiguration.getK8sResources(),
			EventFilterExpression: watcherRequest.EventConfiguration.EventExpression,
			Namespaces:            nsMap,
		}

		port, scoopConfig, err := impl.k8sApplicationService.GetScoopPort(context.Background(), clusterId)
		if err != nil && errors.Is(err, application.ScoopNotConfiguredErr) {
			impl.logger.Errorw("error in informing to scoop", "clusterId", clusterId, "scoopConfig", scoopConfig, "err", err)
			// not returning the error as we have to continue updating other scoops
			continue
		}
		scoopUrl := fmt.Sprintf("http://127.0.0.1:%d", port)
		err = scoopClient.UpdateWatcherConfig(scoopUrl, scoopConfig.PassKey, action, watcher)
		if err != nil {
			impl.logger.Errorw("error in informing to scoop by a REST call", "watcher", watcher, "action", action, "scoopUrl", scoopUrl, "err", err)
			// not returning the error as we have to continue updating other scoops
			continue
		}
	}

	return nil
}
