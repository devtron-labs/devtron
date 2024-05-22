/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package app

import (
	"encoding/json"
	"fmt"
	client "github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/EAMode"
	util2 "github.com/devtron-labs/devtron/pkg/appStore/util"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/deployment/manifest/bean"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/devtron-labs/common-lib/utils/k8s"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	repository2 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/genericNotes"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

const (
	ZERO_INSTALLED_APP_ID = 0
	ZERO_ENVIRONMENT_ID   = 0
)

type AppCrudOperationService interface {
	Create(request *bean.AppLabelDto, tx *pg.Tx) (*bean.AppLabelDto, error)
	FindById(id int) (*bean.AppLabelDto, error)
	FindAll() ([]*bean.AppLabelDto, error)
	GetAppMetaInfo(appId int, installedAppId int, envId int) (*bean.AppMetaInfoDto, error)
	GetHelmAppMetaInfo(appId string) (*bean.AppMetaInfoDto, error)
	GetAppLabelsForDeployment(appId int, appName, envName string) ([]byte, error)
	GetLabelsByAppId(appId int) (map[string]string, error)
	UpdateApp(request *bean.CreateAppDTO) (*bean.CreateAppDTO, error)
	UpdateProjectForApps(request *bean.UpdateProjectBulkAppsRequest) (*bean.UpdateProjectBulkAppsRequest, error)
	GetAppMetaInfoByAppName(appName string) (*bean.AppMetaInfoDto, error)
	GetAppListByTeamIds(teamIds []int, appType string) ([]*TeamAppBean, error)
}

type AppCrudOperationServiceImpl struct {
	logger                 *zap.SugaredLogger
	appLabelRepository     pipelineConfig.AppLabelRepository
	appRepository          appRepository.AppRepository
	userRepository         repository.UserRepository
	installedAppRepository repository2.InstalledAppRepository
	teamRepository         team.TeamRepository
	genericNoteService     genericNotes.GenericNoteService
	gitMaterialRepository  pipelineConfig.MaterialRepository
	installedAppDbService  EAMode.InstalledAppDBService
}

func NewAppCrudOperationServiceImpl(appLabelRepository pipelineConfig.AppLabelRepository,
	logger *zap.SugaredLogger, appRepository appRepository.AppRepository,
	userRepository repository.UserRepository, installedAppRepository repository2.InstalledAppRepository,
	teamRepository team.TeamRepository, genericNoteService genericNotes.GenericNoteService,
	gitMaterialRepository pipelineConfig.MaterialRepository,
	installedAppDbService EAMode.InstalledAppDBService) *AppCrudOperationServiceImpl {
	return &AppCrudOperationServiceImpl{
		appLabelRepository:     appLabelRepository,
		logger:                 logger,
		appRepository:          appRepository,
		userRepository:         userRepository,
		installedAppRepository: installedAppRepository,
		teamRepository:         teamRepository,
		genericNoteService:     genericNoteService,
		gitMaterialRepository:  gitMaterialRepository,
		installedAppDbService:  installedAppDbService,
	}
}

type AppBean struct {
	Id     int    `json:"id"`
	Name   string `json:"name,notnull"`
	TeamId int    `json:"teamId,omitempty"`
}

type TeamAppBean struct {
	ProjectId   int        `json:"projectId"`
	ProjectName string     `json:"projectName"`
	AppList     []*AppBean `json:"appList"`
}

func (impl AppCrudOperationServiceImpl) UpdateApp(request *bean.CreateAppDTO) (*bean.CreateAppDTO, error) {
	// validate the labels key-value if propagate is true
	for _, label := range request.AppLabels {
		if !label.Propagate {
			continue
		}
		labelKey := label.Key
		labelValue := label.Value
		err := k8s.CheckIfValidLabel(labelKey, labelValue)
		if err != nil {
			return nil, err
		}
	}

	dbConnection := impl.appRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	app, err := impl.appRepository.FindById(request.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching app", "error", err)
		return nil, err
	}
	app.TeamId = request.TeamId
	app.UpdatedOn = time.Now()
	app.UpdatedBy = request.UserId
	err = impl.appRepository.Update(app)
	if err != nil {
		impl.logger.Errorw("error in updating app", "error", err)
		return nil, err
	}

	_, err = impl.UpdateLabelsInApp(request, tx)
	if err != nil {
		impl.logger.Errorw("error in updating app labels", "error", err)
		return nil, err
	}

	//updating description
	err = impl.appRepository.SetDescription(app.Id, request.Description, request.UserId)
	if err != nil {
		impl.logger.Errorw("error in setting app description", "err", err, "appId", app.Id)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in commit db transaction", "error", err)
		return nil, err
	}

	return request, nil
}

func (impl AppCrudOperationServiceImpl) UpdateProjectForApps(request *bean.UpdateProjectBulkAppsRequest) (*bean.UpdateProjectBulkAppsRequest, error) {
	dbConnection := impl.appRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	apps, err := impl.appRepository.FindAppsByTeamId(request.TeamId)
	if err != nil {
		impl.logger.Errorw("error in fetching apps", "error", err)
		return nil, err
	}
	for _, app := range apps {
		app.TeamId = request.TeamId
		app.UpdatedOn = time.Now()
		app.UpdatedBy = request.UserId
		err = impl.appRepository.Update(app)
		if err != nil {
			impl.logger.Errorw("error in updating app", "error", err)
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in commit db transaction", "error", err)
		return nil, err
	}
	return nil, nil
}

func (impl AppCrudOperationServiceImpl) Create(request *bean.AppLabelDto, tx *pg.Tx) (*bean.AppLabelDto, error) {
	_, err := impl.appLabelRepository.FindByAppIdAndKeyAndValue(request.AppId, request.Key, request.Value)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching app label", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		model := &pipelineConfig.AppLabel{
			Key:       request.Key,
			Value:     request.Value,
			Propagate: request.Propagate,
			AppId:     request.AppId,
		}
		model.CreatedBy = request.UserId
		model.UpdatedBy = request.UserId
		model.CreatedOn = time.Now()
		model.UpdatedOn = time.Now()
		_, err = impl.appLabelRepository.Create(model, tx)
		if err != nil {
			impl.logger.Errorw("error in creating new app labels", "error", err)
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("duplicate key found for app %d, %s", request.AppId, request.Key)
	}
	return request, nil
}

func (impl AppCrudOperationServiceImpl) UpdateLabelsInApp(request *bean.CreateAppDTO, tx *pg.Tx) (*bean.CreateAppDTO, error) {
	appLabels, err := impl.appLabelRepository.FindAllByAppId(request.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching app label", "error", err)
		return nil, err
	}
	appLabelMap := make(map[string]*pipelineConfig.AppLabel)
	for _, appLabel := range appLabels {
		uniqueLabelExists := fmt.Sprintf("%s:%s:%t", appLabel.Key, appLabel.Value, appLabel.Propagate)
		if _, ok := appLabelMap[uniqueLabelExists]; !ok {
			appLabelMap[uniqueLabelExists] = appLabel
		}
	}
	appLabelDeleteMap := make(map[string]bool, 0)
	for _, label := range request.AppLabels {
		uniqueLabelRequest := fmt.Sprintf("%s:%s:%t", label.Key, label.Value, label.Propagate)
		if _, ok := appLabelMap[uniqueLabelRequest]; !ok {
			// create new
			model := &pipelineConfig.AppLabel{
				Key:       label.Key,
				Value:     label.Value,
				Propagate: label.Propagate,
				AppId:     request.Id,
			}
			model.CreatedBy = request.UserId
			model.UpdatedBy = request.UserId
			model.CreatedOn = time.Now()
			model.UpdatedOn = time.Now()
			_, err = impl.appLabelRepository.Create(model, tx)
			if err != nil {
				impl.logger.Errorw("error in creating new app labels", "error", err)
				return nil, err
			}
		} else {
			// storing this unique so that item remain live, all other item will be delete from this app
			appLabelDeleteMap[uniqueLabelRequest] = true
		}
	}
	for labelReq, _ := range appLabelDeleteMap {
		delete(appLabelMap, labelReq)
	}
	for _, appLabel := range appLabelMap {
		err = impl.appLabelRepository.Delete(appLabel, tx)
		if err != nil {
			impl.logger.Errorw("error in delete app label", "error", err)
			return nil, err
		}
	}
	return request, nil
}

func (impl AppCrudOperationServiceImpl) FindById(id int) (*bean.AppLabelDto, error) {
	model, err := impl.appLabelRepository.FindById(id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching app labels", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		return &bean.AppLabelDto{}, nil
	}
	label := &bean.AppLabelDto{
		Key:       model.Key,
		Value:     model.Value,
		Propagate: model.Propagate,
		AppId:     model.AppId,
	}
	return label, nil
}

func (impl AppCrudOperationServiceImpl) FindAll() ([]*bean.AppLabelDto, error) {
	results := make([]*bean.AppLabelDto, 0)
	models, err := impl.appLabelRepository.FindAll()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching FindAll app labels", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		return results, nil
	}
	for _, model := range models {
		dto := &bean.AppLabelDto{
			AppId:     model.AppId,
			Key:       model.Key,
			Value:     model.Value,
			Propagate: model.Propagate,
		}
		results = append(results, dto)
	}
	return results, nil
}

// GetAppMetaInfo here envId is for installedApp
func (impl AppCrudOperationServiceImpl) GetAppMetaInfo(appId int, installedAppId int, envId int) (*bean.AppMetaInfoDto, error) {
	app, err := impl.appRepository.FindAppAndProjectByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching GetAppMetaInfo", "error", err)
		return nil, err
	}
	labels := make([]*bean.Label, 0)
	models, err := impl.appLabelRepository.FindAllByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching GetAppMetaInfo", "error", err)
		return nil, err
	}
	if err == pg.ErrNoRows {
		impl.logger.Infow("no labels found for app", "app", app)
	} else {
		for _, model := range models {
			dto := &bean.Label{
				Key:       model.Key,
				Value:     model.Value,
				Propagate: model.Propagate,
			}
			labels = append(labels, dto)
		}
	}

	user, err := impl.userRepository.GetByIdIncludeDeleted(app.CreatedBy)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching user for app meta info", "error", err)
		return nil, err
	}
	userEmailId := ""
	if user != nil && user.Id > 0 {
		if user.Active {
			userEmailId = fmt.Sprintf(user.EmailId)
		} else {
			userEmailId = fmt.Sprintf("%s (inactive)", user.EmailId)
		}
	}
	appName := app.AppName
	if app.IsAppJobOrExternalType() {
		appName = app.DisplayName
	}
	noteResp, err := impl.genericNoteService.GetGenericNotesForAppIds([]int{app.Id})
	if err != nil {
		impl.logger.Errorw("error in fetching description", "err", err, "appId", app.Id)
		return nil, err
	}

	info := &bean.AppMetaInfoDto{
		AppId:       app.Id,
		AppName:     appName,
		Description: app.Description,
		ProjectId:   app.TeamId,
		ProjectName: app.Team.Name,
		CreatedBy:   userEmailId,
		CreatedOn:   app.CreatedOn,
		Labels:      labels,
		Active:      app.Active,
		Note:        noteResp[app.Id],
	}
	if installedAppId > 0 {
		installedAppVersion, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppIdAndEnvId(installedAppId, envId)
		if err != nil {
			impl.logger.Error(err)
			return nil, err
		}
		var chartName string
		if installedAppVersion != nil {
			info.ChartUsed = &bean.ChartUsedDto{
				AppStoreAppName:    installedAppVersion.AppStoreApplicationVersion.Name,
				AppStoreAppVersion: installedAppVersion.AppStoreApplicationVersion.Version,
				ChartAvatar:        installedAppVersion.AppStoreApplicationVersion.Icon,
			}
			if installedAppVersion.AppStoreApplicationVersion.AppStore != nil {
				appStore := installedAppVersion.AppStoreApplicationVersion.AppStore
				if appStore.ChartRepoId != 0 && appStore.ChartRepo != nil {
					chartName = appStore.ChartRepo.Name
				} else if appStore.DockerArtifactStore != nil {
					chartName = appStore.DockerArtifactStore.Id
				}

				info.ChartUsed.AppStoreChartId = appStore.Id
				info.ChartUsed.AppStoreChartName = chartName
			}
		}
	} else {
		//app type not helm type, getting gitMaterials
		gitMaterials, err := impl.gitMaterialRepository.FindByAppId(appId)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting gitMaterials by appId", "err", err, "appId", appId)
			return nil, err
		}
		gitMaterialMetaDtos := make([]*bean.GitMaterialMetaDto, 0, len(gitMaterials))
		for _, gitMaterial := range gitMaterials {
			gitMaterialMetaDtos = append(gitMaterialMetaDtos, &bean.GitMaterialMetaDto{
				DisplayName:    gitMaterial.Name[strings.Index(gitMaterial.Name, "-")+1:],
				OriginalUrl:    gitMaterial.Url,
				RedirectionUrl: convertUrlToHttpsIfSshType(gitMaterial.Url),
			})
		}
		info.GitMaterials = gitMaterialMetaDtos
	}
	return info, nil
}

func convertUrlToHttpsIfSshType(url string) string {
	// Regular expression to match SSH URL patterns
	sshPattern := `^(git@|ssh:\/\/)([^:]+):(.+)$`

	// Compile the regular expression
	re := regexp.MustCompile(sshPattern)

	// Check if the input URL matches the SSH pattern, if not already a https one
	if !re.MatchString(url) {
		return url
	}
	// Replace the SSH parts with HTTPS parts
	httpsURL := re.ReplaceAllString(url, "https://$2/$3")
	return httpsURL
}

// getAppAndProjectForAppIdentifier, returns app db model for an app unique identifier or from display_name if both exists else it throws pg.ErrNoRows
func (impl AppCrudOperationServiceImpl) getAppAndProjectForAppIdentifier(appIdentifier *client.AppIdentifier) (*appRepository.App, error) {
	app := &appRepository.App{}
	var err error
	appNameUniqueIdentifier := appIdentifier.GetUniqueAppNameIdentifier()
	app, err = impl.appRepository.FindAppAndProjectByAppName(appNameUniqueIdentifier)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching app meta data by unique app identifier", "appNameUniqueIdentifier", appNameUniqueIdentifier, "err", err)
		return app, err
	}
	if util.IsErrNoRows(err) {
		//find app by display name if not found by unique identifier
		app, err = impl.appRepository.FindAppAndProjectByAppName(appIdentifier.ReleaseName)
		if err != nil {
			impl.logger.Errorw("error in fetching app meta data by display name", "displayName", appIdentifier.ReleaseName, "err", err)
			return app, err
		}
	}
	return app, nil
}

// updateAppNameToUniqueAppIdentifierInApp, migrates values of app_name col. in app table to unique identifier and also updates display_name with releaseName
// returns is requested external app is migrated or other app (linked to chart store) with same name is migrated(which is tracked via namespace).
func (impl AppCrudOperationServiceImpl) updateAppNameToUniqueAppIdentifierInApp(app *appRepository.App, appIdentifier *client.AppIdentifier) error {
	appNameUniqueIdentifier := appIdentifier.GetUniqueAppNameIdentifier()
	isLinked, installedApps, err := impl.installedAppDbService.IsExternalAppLinkedToChartStore(app.Id)
	if err != nil {
		impl.logger.Errorw("error in checking IsExternalAppLinkedToChartStore", "appId", app.Id, "err", err)
		return err
	}
	//if isLinked is true then installed_app found for this app then this app name is already linked to an installed app then
	//create new appEntry for all those installedApps and link installedApp.AppId to the newly created app.
	if isLinked {
		// if installed_apps are already present for that display_name then migrate the app_name to unique identifier with installedApp's ns and clusterId.
		// creating new entry for app all installedApps with uniqueAppNameIdentifier and display name
		err := impl.installedAppDbService.CreateNewAppEntryForAllInstalledApps(installedApps)
		if err != nil {
			impl.logger.Errorw("error in CreateNewAppEntryForAllInstalledApps", "appName", app.AppName, "err", err)
			//not returning from here as we have to migrate the app for requested ext-app and return the response for meta info
		}
	}
	// migrating the requested ext-app
	app.AppName = appNameUniqueIdentifier
	app.DisplayName = appIdentifier.ReleaseName
	app.UpdatedBy = bean2.SystemUserId
	app.UpdatedOn = time.Now()
	err = impl.appRepository.Update(app)
	if err != nil {
		impl.logger.Errorw("error in migrating displayName and appName to unique identifier", "appNameUniqueIdentifier", appNameUniqueIdentifier, "err", err)
		return err
	}
	return nil
}

func (impl AppCrudOperationServiceImpl) GetHelmAppMetaInfo(appId string) (*bean.AppMetaInfoDto, error) {

	// adding separate function for helm apps because for CLI helm apps, apps can be of form "1|clusterName|releaseName"
	// In this case app details can be fetched using app name / release Name.
	appIdSplitted := strings.Split(appId, "|")
	app := &appRepository.App{}
	var err error
	var displayName string
	impl.logger.Info("request payload, appId", appId)
	if len(appIdSplitted) > 1 {
		appIdDecoded, err := client.DecodeExternalAppAppId(appId)
		if err != nil {
			impl.logger.Errorw("error in decoding app id for external app", "appId", appId, "err", err)
			return nil, err
		}
		app, err = impl.getAppAndProjectForAppIdentifier(appIdDecoded)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("GetHelmAppMetaInfo, error in getAppAndProjectForAppIdentifier for external apps", "appIdentifier", appIdDecoded, "err", err)
			return nil, err
		}
		// if app.DisplayName is empty then that app_name is not yet migrated to app name unique identifier
		if app.Id > 0 && len(app.DisplayName) == 0 {
			err = impl.updateAppNameToUniqueAppIdentifierInApp(app, appIdDecoded)
			if err != nil {
				impl.logger.Errorw("GetHelmAppMetaInfo, error in migrating displayName and appName to unique identifier for external apps", "appIdentifier", appIdDecoded, "err", err)
				//not returning from here as we need to show helm app metadata even if migration of app_name fails, then migration can happen on project update
			}
		}
		if app.Id == 0 {
			app.AppName = appIdDecoded.ReleaseName
		}
		if util2.IsExternalChartStoreApp(app.DisplayName) {
			displayName = app.DisplayName
		}

	} else {
		installedAppIdInt, err := strconv.Atoi(appId)
		if err != nil {
			impl.logger.Errorw("error in converting appId to integer", "err", err)
			return nil, err
		}
		InstalledApp, err := impl.installedAppRepository.GetInstalledApp(installedAppIdInt)
		if err != nil {
			impl.logger.Errorw("service err, installedApp", "err", err)
			return nil, err
		}
		app.Id = InstalledApp.AppId
		app.AppName = InstalledApp.App.AppName
		app.TeamId = InstalledApp.App.TeamId
		app.Team.Name = InstalledApp.App.Team.Name
		app.CreatedBy = InstalledApp.App.CreatedBy
		app.Active = InstalledApp.App.Active
		if util2.IsExternalChartStoreApp(InstalledApp.App.DisplayName) {
			// in case of external apps, we will send display name as appName will be a unique identifier
			displayName = InstalledApp.App.DisplayName
		}
	}

	user, err := impl.userRepository.GetByIdIncludeDeleted(app.CreatedBy)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching user for app meta info", "error", err)
		return nil, err
	}
	userEmailId := ""
	if user != nil && user.Id > 0 {
		if user.Active {
			userEmailId = fmt.Sprintf(user.EmailId)
		} else {
			userEmailId = fmt.Sprintf("%s (inactive)", user.EmailId)
		}
	}
	info := &bean.AppMetaInfoDto{
		AppId:       app.Id,
		AppName:     app.AppName,
		ProjectId:   app.TeamId,
		ProjectName: app.Team.Name,
		CreatedBy:   userEmailId,
		CreatedOn:   app.CreatedOn,
		Active:      app.Active,
	}
	if util2.IsExternalChartStoreApp(displayName) {
		//special handling for ext-helm apps where name visible on UI is display name
		info.AppName = displayName
	}
	return info, nil
}

func (impl AppCrudOperationServiceImpl) getLabelsByAppIdForDeployment(appId int) (map[string]string, error) {
	labelsDto := make(map[string]string)
	labels, err := impl.appLabelRepository.FindAllByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting app labels by appId", "err", err, "appId", appId)
		return labelsDto, err
	}
	for _, label := range labels {
		labelKey := strings.TrimSpace(label.Key)
		labelValue := strings.TrimSpace(label.Value)

		if !label.Propagate {
			impl.logger.Warnw("Ignoring label to propagate to app level as propagation is false", "labelKey", labelKey, "labelValue", labelValue, "appId", appId)
			continue
		}

		// if labelKey or labelValue is empty then don't add in labels
		if len(labelKey) == 0 || len(labelValue) == 0 {
			impl.logger.Warnw("Ignoring label to propagate to app level", "labelKey", labelKey, "labelValue", labelValue, "appId", appId)
			continue
		}

		// if labelKey is not satisfying the label key criteria don't add in labels
		// label key must be a 'qualified name' (https://github.com/kubernetes/website/issues/17969)
		err = k8s.CheckIfValidLabel(labelKey, labelValue)
		if err != nil {
			impl.logger.Warnw("Ignoring label to propagate to app level", "err", err, "appId", appId)
			continue
		}

		labelsDto[labelKey] = labelValue
	}
	return labelsDto, nil
}

func (impl AppCrudOperationServiceImpl) getExtraAppLabelsToPropagate(appId int, appName, envName string) (map[string]string, error) {
	appMetaInfo, err := impl.appRepository.FindAppAndProjectByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in finding app and project by appId", "appId", appId, "err", err)
		return nil, err
	}
	return map[string]string{
		bean3.AppNameDevtronLabel:     appName,
		bean3.EnvNameDevtronLabel:     envName,
		bean3.ProjectNameDevtronLabel: appMetaInfo.Team.Name,
	}, nil
}

func (impl AppCrudOperationServiceImpl) GetAppLabelsForDeployment(appId int, appName, envName string) ([]byte, error) {
	appLabelJson := &bean.AppLabelsJsonForDeployment{}
	appLabelsMapFromDb, err := impl.getLabelsByAppIdForDeployment(appId)
	if err != nil {
		impl.logger.Errorw("error in getting app labels from db using appId", "appId", appId, "err", err)
		return nil, err
	}
	extraAppLabelsToPropagate, err := impl.getExtraAppLabelsToPropagate(appId, appName, envName)
	if err != nil {
		impl.logger.Errorw("error in getting extra app labels to propagate", "appName", appName, "envName", envName, "err", err)
		return nil, err
	}
	//when app labels are provided by the user and share the same label key names as those in the extraAppLabelsToPropagate map,
	//priority will be given to the user-provided label keys.
	mergedAppLabels := MergeChildMapToParentMap(appLabelsMapFromDb, extraAppLabelsToPropagate)

	appLabelJson.Labels = mergedAppLabels
	appLabelByte, err := json.Marshal(appLabelJson)
	if err != nil {
		impl.logger.Errorw("error in marshaling appLabels json", "err", err, "appLabelJson", appLabelJson)
		return nil, err
	}
	return appLabelByte, nil
}

func (impl AppCrudOperationServiceImpl) GetLabelsByAppId(appId int) (map[string]string, error) {
	labels, err := impl.appLabelRepository.FindAllByAppId(appId)
	if err != nil {
		if err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting app labels by appId", "err", err, "appId", appId)
			return nil, err
		} else {
			return nil, nil
		}
	}
	labelsDto := make(map[string]string)
	for _, label := range labels {
		labelsDto[label.Key] = label.Value
	}
	return labelsDto, nil
}

func (impl AppCrudOperationServiceImpl) GetAppMetaInfoByAppName(appName string) (*bean.AppMetaInfoDto, error) {
	app, err := impl.appRepository.FindAppAndProjectByAppName(appName)
	if err != nil {
		impl.logger.Errorw("error in fetching GetAppMetaInfoByAppName", "error", err)
		return nil, err
	}
	info := &bean.AppMetaInfoDto{
		AppId:       app.Id,
		AppName:     app.AppName,
		ProjectId:   app.TeamId,
		ProjectName: app.Team.Name,
		CreatedOn:   app.CreatedOn,
		Active:      app.Active,
	}
	return info, nil
}

func (impl AppCrudOperationServiceImpl) GetAppListByTeamIds(teamIds []int, appType string) ([]*TeamAppBean, error) {
	var appsRes []*TeamAppBean
	teamMap := make(map[int]*TeamAppBean)
	var err error
	if len(teamIds) == 0 {
		//no teamIds, getting all active teamIds
		teamIds, err = impl.teamRepository.FindAllActiveTeamIds()
		if err != nil {
			impl.logger.Errorw("error in getting all active team ids", "err", err)
			return nil, err
		}
	}
	apps, err := impl.appRepository.FindAppsByTeamIds(teamIds, appType)
	if err != nil {
		impl.logger.Errorw("error while fetching app", "err", err)
		return nil, err
	}
	for _, app := range apps {
		appName := app.AppName
		if util2.IsExternalChartStoreApp(app.DisplayName) {
			appName = app.DisplayName
		}
		if _, ok := teamMap[app.TeamId]; ok {
			teamMap[app.TeamId].AppList = append(teamMap[app.TeamId].AppList, &AppBean{Id: app.Id, Name: appName})
		} else {

			teamMap[app.TeamId] = &TeamAppBean{ProjectId: app.Team.Id, ProjectName: app.Team.Name}
			teamMap[app.TeamId].AppList = append(teamMap[app.TeamId].AppList, &AppBean{Id: app.Id, Name: appName})
		}
	}

	for _, v := range teamMap {
		if len(v.AppList) == 0 {
			v.AppList = make([]*AppBean, 0)
		}
		appsRes = append(appsRes, v)
	}

	if len(appsRes) == 0 {
		appsRes = make([]*TeamAppBean, 0)
	}

	return appsRes, err
}
