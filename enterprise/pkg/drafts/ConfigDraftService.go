package drafts

import (
	"context"
	"encoding/json"
	"errors"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/enterprise/pkg/protect"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/chart"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/user"
	"go.uber.org/zap"
	"k8s.io/utils/pointer"
	"time"
)

type ConfigDraftService interface {
	protect.ResourceProtectionUpdateListener
	CreateDraft(request ConfigDraftRequest) (*ConfigDraftResponse, error)
	AddDraftVersion(request ConfigDraftVersionRequest) (int, error)
	UpdateDraftState(draftId int, draftVersionId int, toUpdateDraftState DraftState, userId int32) (*DraftVersion, error)
	GetDraftVersionMetadata(draftId int) (*DraftVersionMetadataResponse, error) // would return version timestamp and user email id
	GetDraftComments(draftId int) (*DraftVersionCommentResponse, error)
	GetDrafts(appId int, envId int, resourceType DraftResourceType, userId int32) ([]AppConfigDraft, error)
	GetDraftById(draftId int, userId int32) (*ConfigDraftResponse, error)                                   //  need to send ** in case of view only user for Secret data
	ApproveDraft(draftId int, draftVersionId int, userId int32) error
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
}

func NewConfigDraftServiceImpl(logger *zap.SugaredLogger, configDraftRepository ConfigDraftRepository, configMapService pipeline.ConfigMapService, chartService chart.ChartService,
	propertiesConfigService pipeline.PropertiesConfigService, resourceProtectionService protect.ResourceProtectionService,
	userService user.UserService, appRepo app.AppRepository, envRepository repository2.EnvironmentRepository) *ConfigDraftServiceImpl {
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
	return impl.configDraftRepository.CreateConfigDraft(request)
}

func (impl *ConfigDraftServiceImpl) AddDraftVersion(request ConfigDraftVersionRequest) (int, error) {
	draftId := request.DraftId
	latestDraftVersion, err := impl.configDraftRepository.GetLatestDraftVersionId(draftId)
	if err != nil {
		return 0, err
	}
	lastDraftVersionId := request.LastDraftVersionId
	if latestDraftVersion > lastDraftVersionId {
		return 0, errors.New(LastVersionOutdated)
	}

	currentTime := time.Now()
	if len(request.Data) > 0 {
		//TODO KB: if any edit is being made and draft is in Await_Approval state, ideally we should change state to Init unless changeProposed flag is true
		draftVersionDto := request.GetDraftVersionDto(currentTime)
		draftVersionId, err := impl.configDraftRepository.SaveDraftVersion(draftVersionDto)
		if err != nil {
			return 0, err
		}
		lastDraftVersionId = draftVersionId
	}

	if len(request.UserComment) > 0 {
		draftVersionCommentDto := request.GetDraftVersionComment(lastDraftVersionId, currentTime)
		err = impl.configDraftRepository.SaveDraftVersionComment(draftVersionCommentDto)
		if err != nil {
			return 0, err
		}
	}
	if proposed := request.ChangeProposed; proposed {
		err = impl.configDraftRepository.UpdateDraftState(draftId, AwaitApprovalDraftState, request.UserId)
		if err != nil {
			return 0, err
		}
	}
	return lastDraftVersionId, nil
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

func (impl *ConfigDraftServiceImpl) GetDraftById(draftId int, userId int32) (*ConfigDraftResponse, error) {
	configDraft, err := impl.configDraftRepository.GetLatestConfigDraft(draftId)
	if err != nil {
		return nil, err
	}
	draftResponse := configDraft.ConvertToConfigDraft()
	draftResponse.Approvers = impl.getApproversData(draftResponse.AppId, draftResponse.EnvId)
	userContributedToDraft, err := impl.checkUserContributedToDraft(draftId, userId)
	if err != nil {
		return nil, err
	}
	draftResponse.CanApprove = pointer.BoolPtr(!userContributedToDraft)
	commentsCount, err := impl.configDraftRepository.GetDraftVersionCommentsCount(draftId)
	if err != nil {
		return nil, err
	}
	draftResponse.CommentsCount = commentsCount
	return draftResponse, nil
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

func (impl *ConfigDraftServiceImpl) ApproveDraft(draftId int, draftVersionId int, userId int32) error {
	impl.logger.Infow("approving draft", "draftId", draftId, "draftVersionId", draftVersionId, "userId", userId)
	toUpdateDraftState := PublishedDraftState
	draftVersion, err := impl.validateDraftAction(draftId, draftVersionId, toUpdateDraftState, userId)
	if err != nil {
		return err
	}
	draftData := draftVersion.Data
	draftsDto := draftVersion.Draft
	draftResourceType := draftsDto.Resource
	if draftResourceType == CMDraftResource || draftResourceType == CSDraftResource {
		err = impl.handleCmCsData(draftResourceType, draftsDto, draftData, draftVersion.UserId, draftVersion.Action)
	} else {
		err = impl.handleDeploymentTemplate(draftsDto.AppId, draftsDto.EnvId, draftData, draftVersion.UserId, draftVersion.Action)
	}
	if err != nil {
		return err
	}
	err = impl.configDraftRepository.UpdateDraftState(draftId, toUpdateDraftState, userId)
	return err
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

func (impl *ConfigDraftServiceImpl) handleDeploymentTemplate(appId int, envId int, draftData string, userId int32, action ResourceAction) error {

	ctx := context.Background()
	var err error
	if envId == protect.BASE_CONFIG_ENV_ID {
		err = impl.handleBaseDeploymentTemplate(appId, envId, draftData, userId, ctx)
		if err != nil {
			return err
		}
	} else {
		err = impl.handleEnvLevelTemplate(appId, envId, draftData, userId, action, ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl *ConfigDraftServiceImpl) handleBaseDeploymentTemplate(appId int, envId int, draftData string, userId int32, ctx context.Context) error {
	templateRequest := &chart.TemplateRequest{}
	var templateValidated bool
	err := json.Unmarshal([]byte(draftData), templateRequest)
	if err != nil {
		impl.logger.Errorw("error occurred while unmarshalling draftData of deployment template", "appId", appId, "envId", envId, "err", err)
		return err
	}
	templateValidated, err = impl.chartService.DeploymentTemplateValidate(ctx, templateRequest.ValuesOverride, templateRequest.ChartRefId)
	if err != nil {
		return err
	}
	if !templateValidated {
		return errors.New("template-outdated")
	}
	templateRequest.UserId = userId
	_, err = impl.chartService.UpdateAppOverride(ctx, templateRequest)
	return err
}

func (impl *ConfigDraftServiceImpl) handleEnvLevelTemplate(appId int, envId int, draftData string, userId int32, action ResourceAction, ctx context.Context) error {
	envConfigProperties := &bean.EnvironmentProperties{}
	err := json.Unmarshal([]byte(draftData), envConfigProperties)
	if err != nil {
		impl.logger.Errorw("error occurred while unmarshalling draftData of env deployment template", "appId", appId, "envId", envId, "err", err)
		return err
	}
	if action == AddResourceAction || action == UpdateResourceAction {
		var templateValidated bool
		envConfigProperties.UserId = userId
		envConfigProperties.EnvironmentId = envId
		chartRefId := envConfigProperties.ChartRefId
		templateValidated, err = impl.chartService.DeploymentTemplateValidate(ctx, envConfigProperties.EnvOverrideValues, chartRefId)
		if err != nil {
			return err
		}
		if !templateValidated {
			return errors.New(TemplateOutdated)
		}
		if action == AddResourceAction {
			//TODO code duplicated, needs refactoring
			err = impl.createEnvLevelDeploymentTemplate(ctx, appId, envId, envConfigProperties, userId)
		} else {
			_, err = impl.propertiesConfigService.UpdateEnvironmentProperties(appId, envConfigProperties, userId)
		}
		if err != nil {
			impl.logger.Errorw("service err, EnvConfigOverrideUpdate", "appId", appId, "envId", envId, "err", err, "payload", envConfigProperties)
		}
	} else {
		id := envConfigProperties.Id
		_, err = impl.propertiesConfigService.ResetEnvironmentProperties(id)
		if err != nil {
			impl.logger.Errorw("error occurred while deleting env level Deployment template", "id", id, "err", err)
		}
	}
	return err
}

func (impl *ConfigDraftServiceImpl) createEnvLevelDeploymentTemplate(ctx context.Context, appId int, envId int, envConfigProperties *bean.EnvironmentProperties, userId int32) error {
	_, err := impl.propertiesConfigService.CreateEnvironmentProperties(appId, envConfigProperties)
	if err != nil {
		if err.Error() == bean2.NOCHARTEXIST {
			err = impl.createMissingChart(ctx, appId, envId, envConfigProperties, userId)
			if err == nil {
				_, err = impl.propertiesConfigService.CreateEnvironmentProperties(appId, envConfigProperties)
			}
		}
	}
	return err
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
	application, err := impl.appRepo.FindById(appId)
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
	approvers, err = impl.userService.GetConfigApprovalUsersByEnv(appName, envIdentifier)
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



