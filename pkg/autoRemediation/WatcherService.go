package autoRemediation

import (
	json2 "encoding/json"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/autoRemediation/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"gopkg.in/square/go-jose.v2/json"
)

type WatcherService interface {
	CreateWatcher(watcherRequest *WatcherDto, userId int32) (int, error)
	GetWatcherById(watcherId int) (*WatcherDto, error)
	DeleteWatcherById(watcherId int, userId int32) error
	UpdateWatcherById(watcherId int, watcherRequest *WatcherDto, userId int32) error
	// RetrieveInterceptedEvents(offset int, size int, sortOrder string, searchString string, from time.Time, to time.Time, watchers []string, clusters []string, namespaces []string) (EventsResponse, error)
	FindAllWatchers(offset int, search string, size int, sortOrder string, sortOrderBy string) (WatchersResponse, error)
	GetTriggerByWatcherIds(watcherIds []int) ([]*Trigger, error)
}

type ScoopConfig struct {
	WatcherUrl string `env:"SCOOP_WATCHER_URL" envDefault:"http://scoop.utils:8081"`
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
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService
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
	resourceQualifierMappingService resourceQualifiers.QualifierMappingService,
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
		resourceQualifierMappingService: resourceQualifierMappingService,
		logger:                          logger,
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
		Description:      watcherRequest.Description,
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

	envs, err := impl.getEnvsMap(watcherRequest.EventConfiguration.getEnvsFromSelectors())
	if err != nil {
		impl.logger.Errorw("error in getting envs using env names", "envNames", envs, "err", err)
		return 0, err
	}
	envSelectionIdentifiers := getEnvSelectionIdentifiers(envs)
	err = impl.resourceQualifierMappingService.CreateMappings(tx, userId, resourceQualifiers.K8sEventWatcher, []int{watcher.Id}, resourceQualifiers.EnvironmentSelector, envSelectionIdentifiers)
	if err != nil {
		impl.logger.Errorw("error in mapping watchers to the given envs", "watcher", watcher, "envSelectionIdentifiers", envSelectionIdentifiers, "err", err)
		return 0, err
	}
	err = impl.triggerRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to create trigger", "error", err)
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
	var triggersJobsForWatcher []*Trigger
	for _, trigger := range watcherRequest.Triggers {
		if trigger.IdentifierType == repository.DEVTRON_JOB {
			triggersJobsForWatcher = append(triggersJobsForWatcher, &trigger)
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
		impl.logger.Errorw("error in retrieving details of job pipeline environment", "error", err)
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
			impl.logger.Errorw("error in trigger data ", "error", err)
			return err
		}
		triggerRes := &repository.Trigger{
			WatcherId: watcherId,
			Type:      repository.DEVTRON_JOB,
			Data:      jsonData,
			Active:    true,
			AuditLog:  sql.NewDefaultAuditLog(userId),
		}
		triggersResult = append(triggersResult, triggerRes)
	}
	_, err = impl.triggerRepository.SaveInBulk(triggersResult, tx)
	if err != nil {
		impl.logger.Errorw("error in saving trigger", "error", err)
		return err
	}
	return nil
}

func (impl *WatcherServiceImpl) getJobEnvPipelineDetailsForWatcher(triggers []*Trigger) (*jobDetails, error) {
	var jobNames, envNames, pipelineNames []string
	for _, trig := range triggers {
		jobNames = append(jobNames, trig.Data.JobName)
		envNames = append(envNames, trig.Data.ExecutionEnvironment)
		pipelineNames = append(pipelineNames, trig.Data.PipelineName)
	}
	apps, err := impl.appRepository.FetchAppByDisplayNamesForJobs(jobNames)
	if err != nil {
		impl.logger.Errorw("error in fetching apps", "error", err)
		return &jobDetails{}, err
	}
	var jobIds []int
	for _, app := range apps {
		jobIds = append(jobIds, app.Id)
	}
	pipelines, err := impl.ciPipelineRepository.FindByNames(pipelineNames, jobIds)
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines", "error", err)
		return &jobDetails{}, err
	}
	var pipelinesId []int
	for _, pipeline := range pipelines {
		pipelinesId = append(pipelinesId, pipeline.Id)
	}
	envs, err := impl.environmentRepository.FindByNames(envNames)
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "error", err)
		return &jobDetails{}, err
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
		return &jobDetails{}, err
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
		Description: watcher.Description,
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

	selectors, err := impl.getEnvSelectors(watcherId)
	if err != nil {
		impl.logger.Errorw("error in getting selectors for the watcher", "watcherId", watcherId, "error", err)
		return nil, err
	}

	watcherResponse.EventConfiguration.Selectors = selectors
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

	return nil
}

func (impl *WatcherServiceImpl) FindAllWatchers(offset int, search string, size int, sortOrder string, sortOrderBy string) (WatchersResponse, error) {
	params := repository.WatcherQueryParams{
		Offset:      offset,
		Size:        size,
		Search:      search,
		SortOrderBy: sortOrderBy,
		SortOrder:   sortOrder,
	}
	watchers, err := impl.watcherRepository.FindAllWatchersByQueryName(params)
	if err != nil {
		impl.logger.Errorw("error in retrieving watchers ", "error", err)
		return WatchersResponse{}, err
	}
	var watcherIds []int
	for _, watcher := range watchers {
		watcherIds = append(watcherIds, watcher.Id)
	}
	triggers, err := impl.triggerRepository.GetTriggerByWatcherIds(watcherIds)
	if err != nil {
		impl.logger.Errorw("error in retrieving triggers ", "error", err)
		return WatchersResponse{}, err
	}
	var triggerIds []int
	watcherIdToTrigger := make(map[int]repository.Trigger)
	for _, trigger := range triggers {
		triggerIds = append(triggerIds, trigger.Id)
		watcherIdToTrigger[trigger.WatcherId] = *trigger
	}

	watcherResponses := WatchersResponse{
		Size:   params.Size,
		Offset: params.Offset,
		Total:  len(watchers),
	}
	var pipelineIds []int
	for _, watcher := range watchers {
		var triggerResp TriggerData
		if err := json.Unmarshal(watcherIdToTrigger[watcher.Id].Data, &triggerResp); err != nil {
			impl.logger.Errorw("error in unmarshalling trigger data", "error", err)
			return WatchersResponse{}, err
		}
		pipelineIds = append(pipelineIds, triggerResp.PipelineId)
		watcherResponses.List = append(watcherResponses.List, WatcherItem{
			Name:            watcher.Name,
			Description:     watcher.Description,
			JobPipelineName: triggerResp.PipelineName,
			JobPipelineId:   triggerResp.PipelineId,
		})
	}
	workflows, err := impl.appWorkflowMappingRepository.FindWFCIMappingByCIPipelineIds(pipelineIds)
	if err != nil {
		impl.logger.Errorw("error in retrieving workflows ", "error", err)
		return WatchersResponse{}, err
	}
	var pipelineIdtoAppworkflow map[int]int
	for _, workflow := range workflows {
		pipelineIdtoAppworkflow[workflow.ComponentId] = workflow.AppWorkflowId
	}
	for _, watcherList := range watcherResponses.List {
		watcherList.WorkflowId = pipelineIdtoAppworkflow[watcherList.JobPipelineId]
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
	mappings, err := impl.resourceQualifierMappingService.GetResourceMappingsForResources(resourceQualifiers.K8sEventWatcher, []int{watcherId}, resourceQualifiers.EnvironmentSelector)
	if err != nil {
		return nil, err
	}

	envNames := make([]string, 0, len(mappings))
	for _, mapping := range mappings {
		envNames = append(envNames, mapping.SelectionIdentifier.SelectionIdentifierName.EnvironmentName)
	}

	envs, err := impl.environmentRepository.GetWithClusterByNames(envNames)
	if err != nil {
		return nil, err
	}

	clusterWiseEnvs := make(map[string][]string)
	for _, env := range envs {
		list, ok := clusterWiseEnvs[env.Cluster.ClusterName]
		if !ok {
			list = make([]string, 0)
		}
		list = append(list, env.Name)
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
