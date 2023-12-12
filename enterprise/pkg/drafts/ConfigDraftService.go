package drafts

import (
	"context"
	"encoding/json"
	"errors"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration"
	"github.com/devtron-labs/devtron/enterprise/pkg/protect"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/chart"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/user"
	"go.uber.org/zap"
	"k8s.io/utils/pointer"
	"time"
)

type ConfigDraftService interface {
	protect.ResourceProtectionUpdateListener
	CreateDraft(request ConfigDraftRequest) (*ConfigDraftResponse, error)
	AddDraftVersion(request ConfigDraftVersionRequest) (*ConfigDraftResponse, error)
	UpdateDraftState(draftId int, draftVersionId int, toUpdateDraftState DraftState, userId int32) (*DraftVersion, error)
	GetDraftVersionMetadata(draftId int) (*DraftVersionMetadataResponse, error) // would return version timestamp and user email id
	GetDraftComments(draftId int) (*DraftVersionCommentResponse, error)
	GetDrafts(appId int, envId int, resourceType DraftResourceType, userId int32) ([]AppConfigDraft, error)
	GetDraftById(draftId int, userId int32) (*ConfigDraftResponse, error) //  need to send ** in case of view only user for Secret data
	GetDraftByName(appId, envId int, resourceName string, resourceType DraftResourceType, userId int32) (*ConfigDraftResponse, error)
	ApproveDraft(draftId int, draftVersionId int, userId int32) (*DraftVersionResponse, error)
	DeleteComment(draftId int, draftCommentId int, userId int32) error
	GetDraftsCount(appId int, envIds []int) ([]*DraftCountResponse, error)
	EncryptCSData(draftCsData string) string
}

type ConfigDraftServiceImpl struct {
	logger                    *zap.SugaredLogger
	configDraftRepository     ConfigDraftRepository
	configMapService          pipeline.ConfigMapService
	chartService              chart.ChartService
	propertiesConfigService   pipeline.PropertiesConfigService
	resourceProtectionService protect.ResourceProtectionService
	userService               user.UserService
	appRepo                   app.AppRepository
	envRepository             repository2.EnvironmentRepository
	chartRepository           chartRepoRepository.ChartRepository
	lockedConfigService       lockConfiguration.LockConfigurationService
	envConfigRepo             chartConfig.EnvConfigOverrideRepository
	mergeUtil                 util.MergeUtil
}

func NewConfigDraftServiceImpl(logger *zap.SugaredLogger, configDraftRepository ConfigDraftRepository, configMapService pipeline.ConfigMapService, chartService chart.ChartService,
	propertiesConfigService pipeline.PropertiesConfigService, resourceProtectionService protect.ResourceProtectionService,
	userService user.UserService, appRepo app.AppRepository, envRepository repository2.EnvironmentRepository,
	chartRepository chartRepoRepository.ChartRepository,
	lockedConfigService lockConfiguration.LockConfigurationService,
	envConfigRepo chartConfig.EnvConfigOverrideRepository,
	mergeUtil util.MergeUtil) *ConfigDraftServiceImpl {
	draftServiceImpl := &ConfigDraftServiceImpl{
		logger:                    logger,
		configDraftRepository:     configDraftRepository,
		configMapService:          configMapService,
		chartService:              chartService,
		propertiesConfigService:   propertiesConfigService,
		resourceProtectionService: resourceProtectionService,
		userService:               userService,
		appRepo:                   appRepo,
		envRepository:             envRepository,
		chartRepository:           chartRepository,
		lockedConfigService:       lockedConfigService,
		envConfigRepo:             envConfigRepo,
		mergeUtil:                 mergeUtil,
	}
	resourceProtectionService.RegisterListener(draftServiceImpl)
	return draftServiceImpl
}

func (impl *ConfigDraftServiceImpl) OnStateChange(appId int, envId int, state protect.ProtectionState, userId int32) {
	impl.logger.Debugw("resource protection state change event received", "appId", appId, "envId", envId, "state", state)
	if state == protect.DisabledProtectionState {
		_ = impl.configDraftRepository.DiscardDrafts(appId, envId, userId)
	}
}

func (impl *ConfigDraftServiceImpl) CreateDraft(request ConfigDraftRequest) (*ConfigDraftResponse, error) {
	resourceType := request.Resource
	resourceAction := request.Action
	envId := request.EnvId
	appId := request.AppId
	protectionEnabled := impl.resourceProtectionService.ResourceProtectionEnabled(appId, envId)
	if !protectionEnabled {
		return nil, errors.New(ConfigProtectionDisabled)
	}
	validateResp, draftData, err := impl.validateDraftData(request.AppId, envId, resourceType, resourceAction, request.Data, request.UserId)
	if err != nil {
		return nil, err
	}
	if validateResp != nil {
		return &ConfigDraftResponse{LockValidateResponse: validateResp}, nil
	}
	//assign latest data
	request.Data = draftData
	return impl.configDraftRepository.CreateConfigDraft(request)
}

func (impl *ConfigDraftServiceImpl) AddDraftVersion(request ConfigDraftVersionRequest) (*ConfigDraftResponse, error) {
	draftId := request.DraftId
	latestDraftVersion, err := impl.configDraftRepository.GetLatestDraftVersion(draftId)
	if err != nil {
		return nil, err
	}
	draftDto := latestDraftVersion.Draft
	protectionEnabled := impl.resourceProtectionService.ResourceProtectionEnabled(draftDto.AppId, draftDto.EnvId)
	if !protectionEnabled {
		return nil, errors.New(ConfigProtectionDisabled)
	}
	lastDraftVersionId := request.LastDraftVersionId
	if latestDraftVersion.Id > lastDraftVersionId {
		return nil, errors.New(LastVersionOutdated)
	}

	currentTime := time.Now()
	if len(request.Data) > 0 {

		lockConfig, draftData, err := impl.validateDraftData(draftDto.AppId, draftDto.EnvId, draftDto.Resource, request.Action, request.Data, request.UserId)
		if err != nil {
			return nil, err
		}
		if lockConfig != nil {
			return &ConfigDraftResponse{DraftVersionId: lastDraftVersionId, LockValidateResponse: lockConfig}, nil
		}
		//assign latest draftData
		request.Data = draftData
		draftVersionDto := request.GetDraftVersionDto(currentTime)
		draftVersionId, err := impl.configDraftRepository.SaveDraftVersion(draftVersionDto)
		if err != nil {
			return nil, err
		}
		lastDraftVersionId = draftVersionId
	}

	if len(request.UserComment) > 0 {
		draftVersionCommentDto := request.GetDraftVersionComment(lastDraftVersionId, currentTime)
		err = impl.configDraftRepository.SaveDraftVersionComment(draftVersionCommentDto)
		if err != nil {
			return nil, err
		}
	}
	if proposed := request.ChangeProposed; proposed {
		err = impl.configDraftRepository.UpdateDraftState(draftId, AwaitApprovalDraftState, request.UserId)
		if err != nil {
			return nil, err
		}
	}
	return &ConfigDraftResponse{DraftVersionId: lastDraftVersionId}, nil
}

func (impl *ConfigDraftServiceImpl) UpdateDraftState(draftId int, draftVersionId int, toUpdateDraftState DraftState, userId int32) (*DraftVersion, error) {
	impl.logger.Infow("updating draft state", "draftId", draftId, "toUpdateDraftState", toUpdateDraftState, "userId", userId)
	// check app config draft is enabled or not ??
	latestDraftVersion, err := impl.validateDraftAction(draftId, draftVersionId, toUpdateDraftState, userId)
	if err != nil {
		return nil, err
	}
	err = impl.configDraftRepository.UpdateDraftState(draftId, toUpdateDraftState, userId)
	return latestDraftVersion, err
}

func (impl *ConfigDraftServiceImpl) validateDraftAction(draftId int, draftVersionId int, toUpdateDraftState DraftState, userId int32) (*DraftVersion, error) {
	latestDraftVersion, err := impl.configDraftRepository.GetLatestConfigDraft(draftId)
	if err != nil {
		return nil, err
	}
	if latestDraftVersion.Id != draftVersionId { // needed for current scope
		return nil, errors.New(LastVersionOutdated)
	}
	draftMetadataDto, err := impl.configDraftRepository.GetDraftMetadataById(draftId)
	if err != nil {
		return nil, err
	}
	draftCurrentState := draftMetadataDto.DraftState
	if draftCurrentState.IsTerminal() {
		impl.logger.Errorw("draft is already in terminal state", "draftId", draftId, "draftCurrentState", draftCurrentState)
		return nil, errors.New(DraftAlreadyInTerminalState)
	}
	if toUpdateDraftState == PublishedDraftState {
		if draftCurrentState != AwaitApprovalDraftState {
			impl.logger.Errorw("draft is not in await Approval state", "draftId", draftId, "draftCurrentState", draftCurrentState)
			return nil, errors.New(ApprovalRequestNotRaised)
		} else {
			contributedToDraft, err := impl.checkUserContributedToDraft(draftId, userId)
			if err != nil {
				return nil, err
			}
			if contributedToDraft {
				impl.logger.Errorw("user contributed to this draft", "draftId", draftId, "userId", userId)
				return nil, errors.New(UserContributedToDraft)
			}
		}
	}
	return latestDraftVersion, nil
}

func (impl *ConfigDraftServiceImpl) GetDraftVersionMetadata(draftId int) (*DraftVersionMetadataResponse, error) {
	draftVersionDtos, err := impl.configDraftRepository.GetDraftVersionsMetadata(draftId)
	if err != nil {
		return nil, err
	}
	var draftVersions []*DraftVersionMetadata
	for _, draftVersionDto := range draftVersionDtos {
		versionMetadata := draftVersionDto.ConvertToDraftVersionMetadata()
		draftVersions = append(draftVersions, versionMetadata)
	}
	err = impl.updateWithUserMetadata(draftVersions)
	if err != nil {
		return nil, errors.New("failed to fetch")
	}
	response := &DraftVersionMetadataResponse{}
	response.DraftId = draftId
	response.DraftVersions = draftVersions
	return response, nil
}

func (impl *ConfigDraftServiceImpl) GetDraftComments(draftId int) (*DraftVersionCommentResponse, error) {
	draftComments, err := impl.configDraftRepository.GetDraftVersionComments(draftId)
	if err != nil {
		return nil, err
	}
	var userIds []int32
	for _, draftComment := range draftComments {
		userIds = append(userIds, draftComment.CreatedBy)
	}
	userMetadataMap, err := impl.getUserMetadata(userIds)
	if err != nil {
		return nil, err
	}
	draftVersionVsComments := make(map[int][]UserCommentMetadata)
	for _, draftComment := range draftComments {
		draftVersionId := draftComment.DraftVersionId
		userComment := draftComment.ConvertToDraftVersionComment()
		if userInfo, found := userMetadataMap[userComment.UserId]; found {
			userComment.UserEmail = userInfo.EmailId
		}
		commentMetadataArray := draftVersionVsComments[draftVersionId]
		commentMetadataArray = append(commentMetadataArray, userComment)
		draftVersionVsComments[draftVersionId] = commentMetadataArray
	}
	var draftVersionComments []DraftVersionCommentBean
	for draftVersionId, userComments := range draftVersionVsComments {
		versionComment := DraftVersionCommentBean{
			DraftVersionId: draftVersionId,
			UserComments:   userComments,
		}
		draftVersionComments = append(draftVersionComments, versionComment)
	}
	response := &DraftVersionCommentResponse{
		DraftId:              draftId,
		DraftVersionComments: draftVersionComments,
	}
	return response, nil
}

func (impl *ConfigDraftServiceImpl) GetDrafts(appId int, envId int, resourceType DraftResourceType, userId int32) ([]AppConfigDraft, error) {
	draftMetadataDtos, err := impl.configDraftRepository.GetDraftMetadata(appId, envId, resourceType)
	if err != nil {
		return nil, err
	}
	var appConfigDrafts []AppConfigDraft
	for _, draftMetadataDto := range draftMetadataDtos {
		appConfigDraft := draftMetadataDto.ConvertToAppConfigDraft()
		appConfigDrafts = append(appConfigDrafts, appConfigDraft)
	}
	return appConfigDrafts, nil
}

func (impl *ConfigDraftServiceImpl) GetDraftByName(appId, envId int, resourceName string, resourceType DraftResourceType, userId int32) (*ConfigDraftResponse, error) {
	draftVersion, err := impl.configDraftRepository.GetLatestConfigDraftByName(appId, envId, resourceName, resourceType)
	if err != nil {
		return nil, err
	}
	draftResponse := draftVersion.ConvertToConfigDraft()
	err = impl.updateDraftResponse(draftResponse.DraftId, userId, draftResponse)
	if err != nil {
		return nil, err
	}
	return draftResponse, nil
}

func (impl *ConfigDraftServiceImpl) GetDraftById(draftId int, userId int32) (*ConfigDraftResponse, error) {
	configDraft, err := impl.configDraftRepository.GetLatestConfigDraft(draftId)
	if err != nil {
		return nil, err
	}
	draftResponse := configDraft.ConvertToConfigDraft()
	err = impl.updateDraftResponse(draftId, userId, draftResponse)
	if err != nil {
		return nil, err
	}
	return draftResponse, nil
}

func (impl *ConfigDraftServiceImpl) updateDraftResponse(draftId int, userId int32, draftResponse *ConfigDraftResponse) error {
	draftResponse.Approvers = impl.getApproversData(draftResponse.AppId, draftResponse.EnvId)
	userContributedToDraft, err := impl.checkUserContributedToDraft(draftId, userId)
	if err != nil {
		return err
	}
	draftResponse.CanApprove = pointer.BoolPtr(!userContributedToDraft)
	commentsCount, err := impl.configDraftRepository.GetDraftVersionCommentsCount(draftId)
	if err != nil {
		return err
	}
	draftResponse.CommentsCount = commentsCount
	return nil
}

func (impl ConfigDraftServiceImpl) DeleteComment(draftId int, draftCommentId int, userId int32) error {
	deletedCount, err := impl.configDraftRepository.DeleteComment(draftId, draftCommentId, userId)
	if err != nil {
		return err
	}
	if deletedCount == 0 {
		return errors.New(FailedToDeleteComment)
	}
	return nil
}

func (impl *ConfigDraftServiceImpl) ApproveDraft(draftId int, draftVersionId int, userId int32) (*DraftVersionResponse, error) {
	impl.logger.Infow("approving draft", "draftId", draftId, "draftVersionId", draftVersionId, "userId", userId)
	toUpdateDraftState := PublishedDraftState
	draftVersion, err := impl.validateDraftAction(draftId, draftVersionId, toUpdateDraftState, userId)
	if err != nil {
		return nil, err
	}
	draftData := draftVersion.Data
	draftsDto := draftVersion.Draft
	draftResourceType := draftsDto.Resource
	var draftVersionResponse *DraftVersionResponse
	if draftResourceType == CMDraftResource || draftResourceType == CSDraftResource {
		err = impl.handleCmCsData(draftResourceType, draftsDto, draftData, draftVersion.UserId, draftVersion.Action)
	} else {
		lockValidateResponse, err := impl.handleDeploymentTemplate(draftsDto.AppId, draftsDto.EnvId, draftData, draftVersion.UserId, draftVersion.Action)
		if err != nil {
			return nil, err
		}
		if lockValidateResponse != nil {
			draftVersionResponse.AllowedOverride = lockValidateResponse.AllowedOverride
			draftVersionResponse.LockedOverride = lockValidateResponse.LockedOverride
			draftVersionResponse.IsLockConfigError = lockValidateResponse.IsLockConfigError
		}
	}
	draftVersionResponse.DraftVersionId = draftVersionId
	if err != nil {
		return nil, err
	}
	err = impl.configDraftRepository.UpdateDraftState(draftId, toUpdateDraftState, userId)
	return draftVersionResponse, err
}

func (impl *ConfigDraftServiceImpl) handleCmCsData(draftResource DraftResourceType, draftDto *DraftDto, draftData string, userId int32, action ResourceAction) error {
	// if envId is -1 then it is base Configuration else Env level config
	appId := draftDto.AppId
	envId := draftDto.EnvId
	configDataRequest := &bean.ConfigDataRequest{}
	err := json.Unmarshal([]byte(draftData), configDataRequest)
	if err != nil {
		impl.logger.Errorw("error occurred while unmarshalling draftData of CM/CS", "appId", appId, "envId", envId, "err", err)
		return err
	}
	configDataRequest.UserId = userId // setting draftVersion userId
	isCm := draftResource == CMDraftResource
	if isCm {
		if envId == protect.BASE_CONFIG_ENV_ID {
			if action == DeleteResourceAction {
				_, err = impl.configMapService.CMGlobalDelete(draftDto.ResourceName, configDataRequest.Id, userId)
			} else {
				_, err = impl.configMapService.CMGlobalAddUpdate(configDataRequest)
			}
		} else {
			if action == DeleteResourceAction {
				_, err = impl.configMapService.CMEnvironmentDelete(draftDto.ResourceName, configDataRequest.Id, userId)
			} else {
				_, err = impl.configMapService.CMEnvironmentAddUpdate(configDataRequest)
			}
		}
	} else {
		if envId == protect.BASE_CONFIG_ENV_ID {
			if action == DeleteResourceAction {
				_, err = impl.configMapService.CSGlobalDelete(draftDto.ResourceName, configDataRequest.Id, userId)
			} else {
				_, err = impl.configMapService.CSGlobalAddUpdate(configDataRequest)
			}
		} else {
			if action == DeleteResourceAction {
				_, err = impl.configMapService.CSEnvironmentDelete(draftDto.ResourceName, configDataRequest.Id, userId)
			} else {
				_, err = impl.configMapService.CSEnvironmentAddUpdate(configDataRequest)
			}
		}
	}
	if err != nil {
		impl.logger.Errorw("error occurred while adding/updating/deleting config", "isCm", isCm, "action", action, "appId", appId, "envId", envId, "err", err)
	}
	return err
}

func (impl *ConfigDraftServiceImpl) EncryptCSData(draftCsData string) string {
	configDataRequest := &bean.ConfigDataRequest{}
	err := json.Unmarshal([]byte(draftCsData), configDataRequest)
	if err != nil {
		impl.logger.Errorw("error occurred while unmarshalling draftData of CS", "err", err)
		return draftCsData
	}
	configData := configDataRequest.ConfigData
	var configDataResponse []*bean.ConfigData
	for _, data := range configData {
		_ = impl.configMapService.EncryptCSData(data)
		configDataResponse = append(configDataResponse, data)
	}
	configDataRequest.ConfigData = configDataResponse
	encryptedCSData, err := json.Marshal(configDataRequest)
	if err != nil {
		impl.logger.Errorw("error occurred while marshalling config data request, so returning original data", "err", err)
		return draftCsData
	}
	return string(encryptedCSData)
}

func (impl *ConfigDraftServiceImpl) handleDeploymentTemplate(appId int, envId int, draftData string, userId int32, action ResourceAction) (*LockValidateResponse, error) {

	ctx := context.Background()
	var err error
	var lockValidateResp *LockValidateResponse
	if envId == protect.BASE_CONFIG_ENV_ID {
		lockValidateResp, err = impl.handleBaseDeploymentTemplate(appId, envId, draftData, userId, action, ctx)
		if err != nil {
			return nil, err
		}
	} else {
		lockValidateResp, err = impl.handleEnvLevelTemplate(appId, envId, draftData, userId, action, ctx)
		if err != nil {
			return nil, err
		}
	}
	return lockValidateResp, nil
}

func (impl *ConfigDraftServiceImpl) handleBaseDeploymentTemplate(appId int, envId int, draftData string, userId int32, action ResourceAction, ctx context.Context) (*LockValidateResponse, error) {
	templateRequest := chart.TemplateRequest{}
	var templateValidated bool
	err := json.Unmarshal([]byte(draftData), &templateRequest)
	if err != nil {
		impl.logger.Errorw("error occurred while unmarshalling draftData of deployment template", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}

	env, _ := impl.envRepository.FindById(envId)
	//VARIABLE_RESOLVE
	scope := resourceQualifiers.Scope{
		AppId:     appId,
		EnvId:     envId,
		ClusterId: env.ClusterId,
	}

	templateValidated, err = impl.chartService.DeploymentTemplateValidate(ctx, templateRequest.ValuesOverride, templateRequest.ChartRefId, scope)
	if err != nil {
		return nil, err
	}
	if !templateValidated {
		return nil, errors.New(TemplateOutdated)
	}
	templateRequest.UserId = userId
	var createResp *chart.TemplateResponse
	var lockValidateResp *LockValidateResponse
	if action == AddResourceAction {
		createResp, err = impl.chartService.Create(templateRequest, ctx)
	} else {
		createResp, err = impl.chartService.UpdateAppOverride(ctx, &templateRequest)
	}
	if createResp != nil {
		lockValidateResp = &LockValidateResponse{
			AllowedOverride:   createResp.AllowedOverride,
			LockedOverride:    createResp.LockedOverride,
			IsLockConfigError: createResp.IsLockConfigError,
		}
	}
	return lockValidateResp, err
}

func (impl *ConfigDraftServiceImpl) handleEnvLevelTemplate(appId int, envId int, draftData string, userId int32, action ResourceAction, ctx context.Context) (*LockValidateResponse, error) {
	envConfigProperties := &bean.EnvironmentProperties{}
	err := json.Unmarshal([]byte(draftData), envConfigProperties)
	if err != nil {
		impl.logger.Errorw("error occurred while unmarshalling draftData of env deployment template", "appId", appId, "envId", envId, "err", err)
		return nil, err
	}
	var updateResp *bean.EnvironmentUpdateResponse
	var lockValidateResp *LockValidateResponse
	if action == AddResourceAction || action == UpdateResourceAction {
		var templateValidated bool
		envConfigProperties.UserId = userId
		envConfigProperties.EnvironmentId = envId
		chartRefId := envConfigProperties.ChartRefId

		//VARIABLE_RESOLVE
		env, _ := impl.envRepository.FindById(envId)
		scope := resourceQualifiers.Scope{
			AppId:     appId,
			EnvId:     envId,
			ClusterId: env.ClusterId,
		}

		templateValidated, err = impl.chartService.DeploymentTemplateValidate(ctx, envConfigProperties.EnvOverrideValues, chartRefId, scope)
		if err != nil {
			return nil, err
		}
		if !templateValidated {
			return nil, errors.New(TemplateOutdated)
		}
		if action == AddResourceAction {
			//TODO code duplicated, needs refactoring
			updateResp, err = impl.createEnvLevelDeploymentTemplate(ctx, appId, envId, envConfigProperties, userId)
		} else {
			updateResp, err = impl.propertiesConfigService.UpdateEnvironmentProperties(appId, envConfigProperties, userId)
		}
		if err != nil {
			impl.logger.Errorw("service err, EnvConfigOverrideUpdate", "appId", appId, "envId", envId, "err", err, "payload", envConfigProperties)
		}
		if updateResp != nil {
			lockValidateResp = &LockValidateResponse{
				AllowedOverride:   updateResp.AllowedOverride,
				LockedOverride:    updateResp.LockedOverride,
				IsLockConfigError: updateResp.IsLockConfigError,
			}
		}
	} else {
		id := envConfigProperties.Id
		_, err = impl.propertiesConfigService.ResetEnvironmentProperties(id)
		if err != nil {
			impl.logger.Errorw("error occurred while deleting env level Deployment template", "id", id, "err", err)
		}
	}
	return lockValidateResp, err
}

func (impl *ConfigDraftServiceImpl) createEnvLevelDeploymentTemplate(ctx context.Context, appId int, envId int, envConfigProperties *bean.EnvironmentProperties, userId int32) (*bean.EnvironmentUpdateResponse, error) {
	createResp, err := impl.propertiesConfigService.CreateEnvironmentProperties(appId, envConfigProperties)
	if err != nil {
		if err.Error() == bean2.NOCHARTEXIST {
			err = impl.createMissingChart(ctx, appId, envId, envConfigProperties, userId)
			if err == nil {
				createResp, err = impl.propertiesConfigService.CreateEnvironmentProperties(appId, envConfigProperties)
			}
		}
	}
	return createResp, err
}

func (impl *ConfigDraftServiceImpl) createMissingChart(ctx context.Context, appId int, envId int, envConfigProperties *bean.EnvironmentProperties, userId int32) error {
	appMetrics := false
	if envConfigProperties.AppMetrics != nil {
		appMetrics = *envConfigProperties.AppMetrics
	}
	templateRequest := chart.TemplateRequest{
		AppId:               appId,
		ChartRefId:          envConfigProperties.ChartRefId,
		ValuesOverride:      []byte("{}"),
		UserId:              userId,
		IsAppMetricsEnabled: appMetrics,
	}
	_, err := impl.chartService.CreateChartFromEnvOverride(templateRequest, ctx)
	if err != nil {
		impl.logger.Errorw("service err, EnvConfigOverrideCreate from draft", "appId", appId, "envId", envId, "err", err, "payload", envConfigProperties)
	}
	return err
}

func (impl *ConfigDraftServiceImpl) updateWithUserMetadata(versions []*DraftVersionMetadata) error {
	var userIds []int32
	for _, versionMetadata := range versions {
		userIds = append(userIds, versionMetadata.UserId)
	}
	userIdVsUserInfoMap, err := impl.getUserMetadata(userIds)
	if err != nil {
		return err
	}
	for _, versionMetadata := range versions {
		if userInfo, found := userIdVsUserInfoMap[versionMetadata.UserId]; found {
			versionMetadata.UserEmail = userInfo.EmailId
		}
	}
	return nil
}

func (impl *ConfigDraftServiceImpl) getUserMetadata(userIds []int32) (map[int32]bean2.UserInfo, error) {
	userInfos, err := impl.userService.GetByIds(userIds)
	if err != nil {
		return nil, err
	}
	userIdVsUserInfoMap := make(map[int32]bean2.UserInfo, len(userIds))
	for _, userInfo := range userInfos {
		userIdVsUserInfoMap[userInfo.Id] = userInfo
	}
	return userIdVsUserInfoMap, nil
}

func (impl *ConfigDraftServiceImpl) getApproversData(appId int, envId int) []string {
	var approvers []string
	application, err := impl.appRepo.FindAppAndTeamByAppId(appId)
	if err != nil {
		return approvers
	}
	var appName = application.AppName
	var env *repository2.Environment
	envIdentifier := ""
	if envId > 0 {
		env, err = impl.envRepository.FindById(envId)
		if err != nil {
			return approvers
		}
		envIdentifier = env.EnvironmentIdentifier
	}
	approvers, err = impl.userService.GetConfigApprovalUsersByEnv(appName, envIdentifier, application.Team.Name)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching config approval emails, so sending empty approvers list", "err", err)
	}
	return approvers
}

func (impl *ConfigDraftServiceImpl) checkUserContributedToDraft(draftId int, userId int32) (bool, error) {
	versionsMetadata, err := impl.configDraftRepository.GetDraftVersionsMetadata(draftId)
	if err != nil {
		return false, err
	}
	for _, versionMetadata := range versionsMetadata {
		if versionMetadata.UserId == userId {
			return true, nil
		}
	}
	return false, nil
}

func (impl *ConfigDraftServiceImpl) GetDraftsCount(appId int, envIds []int) ([]*DraftCountResponse, error) {
	var draftCountResponse []*DraftCountResponse
	draftDtos, err := impl.configDraftRepository.GetDraftMetadataForAppAndEnv(appId, envIds)
	if err != nil {
		return draftCountResponse, err
	}
	draftCountMap := make(map[int]int, len(draftDtos))
	for _, draftDto := range draftDtos {
		envId := draftDto.EnvId
		count := draftCountMap[envId]
		count++
		draftCountMap[envId] = count
	}
	for envId, count := range draftCountMap {
		draftCountResponse = append(draftCountResponse, &DraftCountResponse{AppId: appId, EnvId: envId, DraftsCount: count})
	}
	return draftCountResponse, nil
}

func (impl *ConfigDraftServiceImpl) validateDraftData(appId int, envId int, resourceType DraftResourceType, action ResourceAction, draftData string, userId int32) (*LockValidateResponse, string, error) {
	if resourceType == CMDraftResource || resourceType == CSDraftResource {
		return nil, draftData, impl.validateCmCs(action, draftData)
	}
	return impl.validateDeploymentTemplate(appId, envId, action, draftData, userId)
}

func (impl *ConfigDraftServiceImpl) validateCmCs(resourceAction ResourceAction, draftData string) error {
	configDataRequest := &bean.ConfigDataRequest{}
	err := json.Unmarshal([]byte(draftData), configDataRequest)
	if err != nil {
		impl.logger.Errorw("error occurred while unmarshalling draftData of CM/CS", "err", err)
		return err
	}
	if resourceAction == AddResourceAction || resourceAction == UpdateResourceAction {
		configData := configDataRequest.ConfigData[0]
		_, err = impl.configMapService.ValidateConfigData(configData)
	} else {
		configId := configDataRequest.Id
		if configId == 0 {
			impl.logger.Errorw("error occurred while validating CM/CS ", "id", configId)
			err = errors.New("invalid config id")
		}
	}
	return err
}

func (impl *ConfigDraftServiceImpl) validateDeploymentTemplate(appId int, envId int, resourceAction ResourceAction, draftData string, userId int32) (*LockValidateResponse, string, error) {
	if envId == protect.BASE_CONFIG_ENV_ID {
		templateRequest := chart.TemplateRequest{}
		var templateValidated bool
		err := json.Unmarshal([]byte(draftData), &templateRequest)
		if err != nil {
			impl.logger.Errorw("error occurred while unmarshalling draftData of deployment template", "envId", envId, "err", err)
			return nil, draftData, err
		}
		currentLatestChart, err := impl.chartRepository.FindLatestChartForAppByAppId(templateRequest.AppId)
		if err != nil {
			return nil, draftData, err
		}

		if templateRequest.SaveEligibleChanges {
			eligible, err := impl.mergeUtil.JsonPatch([]byte(currentLatestChart.GlobalOverride), templateRequest.ValuesOverride)
			if err != nil {
				return nil, draftData, err
			}
			templateRequest.ValuesOverride = eligible
			templateByte, err := json.Marshal(templateRequest)
			if err != nil {
				return nil, draftData, err
			}
			draftData = string(templateByte)
		}

		isLockConfigError, lockedOverride, err := impl.lockedConfigService.HandleLockConfiguration(string(templateRequest.ValuesOverride), currentLatestChart.GlobalOverride, int(userId))
		if err != nil {
			return nil, draftData, err
		}
		if isLockConfigError {
			var jsonVal json.RawMessage
			_ = json.Unmarshal([]byte(lockedOverride), &jsonVal)
			return &LockValidateResponse{
				AllowedOverride:   nil,
				LockedOverride:    jsonVal,
				IsLockConfigError: true,
			}, draftData, nil
		}
		//VARIABLE_RESOLVE
		env, _ := impl.envRepository.FindById(envId)
		scope := resourceQualifiers.Scope{
			AppId:     templateRequest.AppId,
			EnvId:     envId,
			ClusterId: env.ClusterId,
		}
		templateValidated, err = impl.chartService.DeploymentTemplateValidate(context.Background(), templateRequest.ValuesOverride, templateRequest.ChartRefId, scope)
		if err != nil {
			return nil, draftData, err
		}
		if !templateValidated {
			return nil, draftData, errors.New(TemplateOutdated)
		}
	} else {
		envConfigProperties := &bean.EnvironmentProperties{}
		err := json.Unmarshal([]byte(draftData), envConfigProperties)
		if err != nil {
			impl.logger.Errorw("error occurred while unmarshalling draftData of env deployment template", "envId", envId, "err", err)
			return nil, draftData, err
		}
		if resourceAction == AddResourceAction || resourceAction == UpdateResourceAction {
			currentLatestChart, err := impl.envConfigRepo.FindLatestChartForAppByAppIdAndEnvId(appId, envId)
			if err != nil {
				return nil, draftData, err
			}
			if envConfigProperties.SaveEligibleChanges {
				eligible, err := impl.mergeUtil.JsonPatch([]byte(currentLatestChart.EnvOverrideValues), envConfigProperties.EnvOverrideValues)
				if err != nil {
					return nil, draftData, err
				}
				envConfigProperties.EnvOverrideValues = eligible
				envConfigByte, err := json.Marshal(envConfigProperties)
				if err != nil {
					return nil, draftData, err
				}
				draftData = string(envConfigByte)
			}
			isLockConfigError, lockedOverride, err := impl.lockedConfigService.HandleLockConfiguration(string(envConfigProperties.EnvOverrideValues), currentLatestChart.EnvOverrideValues, int(userId))
			if err != nil {
				return nil, draftData, err
			}
			if isLockConfigError {
				var jsonVal json.RawMessage
				_ = json.Unmarshal([]byte(lockedOverride), &jsonVal)
				return &LockValidateResponse{
					AllowedOverride:   nil,
					LockedOverride:    jsonVal,
					IsLockConfigError: true,
				}, draftData, nil
			}
			//VARIABLE_RESOLVE
			env, _ := impl.envRepository.FindById(envId)
			scope := resourceQualifiers.Scope{
				AppId:     appId,
				EnvId:     envId,
				ClusterId: env.ClusterId,
			}

			chartRefId := envConfigProperties.ChartRefId
			templateValidated, err := impl.chartService.DeploymentTemplateValidate(context.Background(), envConfigProperties.EnvOverrideValues, chartRefId, scope)
			if err != nil {
				return nil, draftData, err
			}
			if !templateValidated {
				return nil, draftData, errors.New(TemplateOutdated)
			}
		} else {
			id := envConfigProperties.Id
			if id == 0 {
				impl.logger.Errorw("error occurred while validating CM/CS ", "id", id)
				err = errors.New("invalid template ref id")
			}
		}
	}
	return nil, draftData, nil
}
