package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/bean"
	resourceGroup2 "github.com/devtron-labs/devtron/pkg/resourceGroup"
)

type AppConfigService interface {
	//CreateApp : This function creates applications of type Job as well as Devtronapps
	// In case of error response object is nil
	CreateApp(request *bean.CreateAppDTO) (*bean.CreateAppDTO, error)
	//DeleteApp : This function deletes applications of type Job as well as DevtronApps
	DeleteApp(appId int, userId int32) error
	//GetApp : Gets Application along with Git materials for given appId.
	//If the application type is a 'Chart Store App', it doesnt provide any detail.
	//For application types like Jobs and DevtronApps, it retrieves Git materials associated with the application.
	//In case of error response object is nil
	GetApp(appId int) (application *bean.CreateAppDTO, err error)
	//FindByIds : Find applications by given IDs, delegating the request to the appRepository.
	// It queries the repository for applications corresponding to the given IDs and constructs
	//a list of AppBean objects containing ID, name, and team ID.
	//It returns the list of AppBean instances.
	//In case of error,AppBean is returned as nil.
	FindByIds(ids []*int) ([]*AppBean, error)
	//GetAppList : Retrieve and return a list of applications after converting in proper bean object.
	//In case of any error , []AppBean is returned as nil.
	GetAppList() ([]AppBean, error)
	//FindAllMatchesByAppName : Find and return applications matching the given name and type.
	//Internally,It performs a case-insensitive search based on the applicationName("%"+appName+"%") and type.
	//In case of error,[]*AppBean is returned as nil.
	FindAllMatchesByAppName(appName string, appType helper.AppType) ([]*AppBean, error)
	//GetAppListForEnvironment : Retrieves a list of applications (AppBean) based on the provided ResourceGroupingRequest.
	// It first determines the relevant application and environment objects based on the active pipelines fetched from the repository.
	//The function then performs authorization checks on these objects for the given user.
	//Finally , the corresponding AppBean objects are added to the applicationList and then returned.
	//In case of error,[]*AppBean is returned as nil.
	GetAppListForEnvironment(request resourceGroup2.ResourceGroupingRequest) ([]*AppBean, error)
	//FindAppsByTeamId : Retrieves applications (AppBean) associated with the provided teamId
	//It queries the repository for applications belonging to the specified team(project) and
	//constructs a list of AppBean instances containing ID and name.
	//The function returns the list of applications in valid case.
	//In case of error,[]*AppBean is returned as nil.
	FindAppsByTeamId(teamId int) ([]*AppBean, error)
	//FindAppsByTeamName : Retrieves applications (AppBean) associated with the provided teamName
	// It queries the repository for applications belonging to the specified team(project) and
	// constructs a list of AppBean instances containing ID and name.
	// The function returns the list of applications in valid case.
	// In case of error,[]*AppBean is returned as nil.
	FindAppsByTeamName(teamName string) ([]AppBean, error)
}

func (impl *PipelineBuilderImpl) CreateApp(request *bean.CreateAppDTO) (*bean.CreateAppDTO, error) {
	impl.logger.Debugw("app create request received", "req", request)

	res, err := impl.ciCdPipelineOrchestrator.CreateApp(request)
	if err != nil {
		impl.logger.Errorw("error in saving create app req", "req", request, "err", err)
	}
	return res, err
}

func (impl *PipelineBuilderImpl) DeleteApp(appId int, userId int32) error {
	impl.logger.Debugw("app delete request received", "app", appId)
	err := impl.ciCdPipelineOrchestrator.DeleteApp(appId, userId)
	return err
}

func (impl *PipelineBuilderImpl) GetApp(appId int) (application *bean.CreateAppDTO, err error) {
	app, err := impl.appRepo.FindById(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching app", "id", appId, "err", err)
		return nil, err
	}
	application = &bean.CreateAppDTO{
		Id:      app.Id,
		AppName: app.AppName,
		TeamId:  app.TeamId,
		AppType: app.AppType,
	}
	if app.AppType == helper.ChartStoreApp {
		return application, nil
	}
	gitMaterials := impl.GetMaterialsForAppId(appId)
	application.Material = gitMaterials
	if app.AppType == helper.Job {
		app.AppName = app.DisplayName
	}
	application.AppType = app.AppType
	return application, nil
}

func (impl *PipelineBuilderImpl) FindByIds(ids []*int) ([]*AppBean, error) {
	var appsRes []*AppBean
	apps, err := impl.appRepo.FindByIds(ids)
	if err != nil {
		impl.logger.Errorw("error while fetching app", "err", err)
		return nil, err
	}
	for _, app := range apps {
		appsRes = append(appsRes, &AppBean{Id: app.Id, Name: app.AppName, TeamId: app.TeamId})
	}
	return appsRes, err
}

func (impl *PipelineBuilderImpl) GetAppList() ([]AppBean, error) {
	var appsRes []AppBean
	apps, err := impl.appRepo.FindAll()
	if err != nil {
		impl.logger.Errorw("error while fetching app", "err", err)
		return nil, err
	}
	for _, app := range apps {
		appsRes = append(appsRes, AppBean{Id: app.Id, Name: app.AppName})
	}
	return appsRes, err
}

func (impl *PipelineBuilderImpl) FindAllMatchesByAppName(appName string, appType helper.AppType) ([]*AppBean, error) {
	var appsRes []*AppBean
	var apps []*app.App
	var err error
	if len(appName) == 0 {
		apps, err = impl.appRepo.FindAll()
	} else {
		apps, err = impl.appRepo.FindAllMatchesByAppName(appName, appType)
	}
	if err != nil {
		impl.logger.Errorw("error while fetching app", "err", err)
		return nil, err
	}
	for _, app := range apps {
		name := app.AppName
		if appType == helper.Job {
			name = app.DisplayName
		}
		appsRes = append(appsRes, &AppBean{Id: app.Id, Name: name})
	}
	return appsRes, err
}

func (impl PipelineBuilderImpl) GetAppListForEnvironment(request resourceGroup2.ResourceGroupingRequest) ([]*AppBean, error) {
	var applicationList []*AppBean
	var cdPipelines []*pipelineConfig.Pipeline
	var err error
	if request.ResourceGroupId > 0 {
		appIds, err := impl.resourceGroupService.GetResourceIdsByResourceGroupId(request.ResourceGroupId)
		if err != nil {
			return nil, err
		}
		//override appIds if already provided app group id in request.
		request.ResourceIds = appIds
	}
	if len(request.ResourceIds) > 0 {
		cdPipelines, err = impl.pipelineRepository.FindActiveByInFilter(request.ParentResourceId, request.ResourceIds)
	} else {
		cdPipelines, err = impl.pipelineRepository.FindActiveByEnvId(request.ParentResourceId)
	}
	if err != nil {
		impl.logger.Errorw("error in fetching pipelines", "request", request, "err", err)
		return nil, err
	}
	if len(cdPipelines) == 0 {
		return applicationList, nil
	}
	var appObjectArr []string
	var envObjectArr []string
	objects := impl.enforcerUtil.GetAppAndEnvObjectByDbPipeline(cdPipelines)
	for _, object := range objects {
		appObjectArr = append(appObjectArr, object[0])
		envObjectArr = append(envObjectArr, object[1])
	}
	appResults, envResults := request.CheckAuthBatch(request.EmailId, appObjectArr, envObjectArr)
	for _, pipeline := range cdPipelines {
		appObject := objects[pipeline.Id][0]
		envObject := objects[pipeline.Id][1]
		if !(appResults[appObject] && envResults[envObject]) {
			//if user unauthorized, skip items
			continue
		}
		applicationList = append(applicationList, &AppBean{Id: pipeline.AppId, Name: pipeline.App.AppName})
	}
	return applicationList, err
}

func (impl *PipelineBuilderImpl) FindAppsByTeamId(teamId int) ([]*AppBean, error) {
	var appsRes []*AppBean
	apps, err := impl.appRepo.FindAppsByTeamId(teamId)
	if err != nil {
		impl.logger.Errorw("error while fetching app", "err", err)
		return nil, err
	}
	for _, app := range apps {
		appsRes = append(appsRes, &AppBean{Id: app.Id, Name: app.AppName})
	}
	return appsRes, err
}

func (impl *PipelineBuilderImpl) FindAppsByTeamName(teamName string) ([]AppBean, error) {
	var appsRes []AppBean
	apps, err := impl.appRepo.FindAppsByTeamName(teamName)
	if err != nil {
		impl.logger.Errorw("error while fetching app", "err", err)
		return nil, err
	}
	for _, app := range apps {
		appsRes = append(appsRes, AppBean{Id: app.Id, Name: app.AppName})
	}
	return appsRes, err
}
