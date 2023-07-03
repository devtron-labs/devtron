package drafts

import (
	"errors"
	"go.uber.org/zap"
)

type ConfigDraftService interface {
	CreateDraft(request ConfigDraftRequest) (*ConfigDraftResponse, error)
	AddDraftVersion(request ConfigDraftVersionRequest) (int, error)
	UpdateDraftState(draftId int, draftState DraftState, userId int32) error
	GetDraftVersionMetadata(draftId int) (*DraftVersionMetadataResponse, error) // would return version timestamp and user email id
	GetDraftComments(draftId int) (*DraftVersionCommentResponse, error)
	GetDrafts(appId int, envId int, resourceType DraftResourceType) ([]AppConfigDraft, error) // need to take care of secret data
	GetDraftById(draftId int) (*ConfigDraftResponse, error)                                   //  need to send ** in case of view only user for Secret data
}

type ConfigDraftServiceImpl struct {
	logger                *zap.SugaredLogger
	configDraftRepository ConfigDraftRepository
}

func NewConfigDraftServiceImpl(logger *zap.SugaredLogger, configDraftRepository ConfigDraftRepository) *ConfigDraftServiceImpl {
	return &ConfigDraftServiceImpl{
		logger:                logger,
		configDraftRepository: configDraftRepository,
	}
}

func (impl ConfigDraftServiceImpl) CreateDraft(request ConfigDraftRequest) (*ConfigDraftResponse, error) {
	return impl.configDraftRepository.CreateConfigDraft(request)
}

func (impl ConfigDraftServiceImpl) AddDraftVersion(request ConfigDraftVersionRequest) (int, error) {
	latestDraftVersion, err := impl.configDraftRepository.GetLatestDraftVersionId(request.DraftId)
	if err != nil {
		return 0, err
	}
	lastDraftVersionId := request.LastDraftVersionId
	if latestDraftVersion > lastDraftVersionId {
		return 0, errors.New("last-version-outdated")
	}

	if len(request.Data) > 0 {
		draftVersionDto := DraftVersionDto{}
		draftVersionDto.DraftId = request.DraftId
		draftVersionDto.Data = request.Data
		draftVersionDto.Action = request.Action
		draftVersionDto.UserId = request.UserId
		draftVersionId, err := impl.configDraftRepository.SaveDraftVersion(draftVersionDto)
		if err != nil {
			return 0, err
		}
		lastDraftVersionId = draftVersionId
	}

	if len(request.UserComment) > 0 {
		draftVersionCommentDto := DraftVersionCommentDto{}
		draftVersionCommentDto.DraftId = request.DraftId
		draftVersionCommentDto.DraftVersionId = lastDraftVersionId
		draftVersionCommentDto.Comment = request.UserComment
		draftVersionCommentDto.UserId = request.UserId
		err := impl.configDraftRepository.SaveDraftVersionComment(draftVersionCommentDto)
		if err != nil {
			return 0, err
		}
	}
	return lastDraftVersionId, nil
}

func (impl ConfigDraftServiceImpl) UpdateDraftState(draftId int, toUpdateDraftState DraftState, userId int32) error {
	impl.logger.Infow("updating draft state", "draftId", draftId, "toUpdateDraftState", toUpdateDraftState, "userId", userId)
	// check app config draft is enabled or not ??
	draftMetadataDto, err := impl.configDraftRepository.GetDraftMetadataById(draftId)
	if err != nil {
		return err
	}
	draftCurrentState := draftMetadataDto.DraftState
	if draftCurrentState.IsTerminal() {
		impl.logger.Errorw("draft is already in terminal state", "draftId", draftId, "draftCurrentState", draftCurrentState)
		return errors.New("already-in-terminal-state")
	}
	if toUpdateDraftState == PublishedDraftState && draftCurrentState != AwaitApprovalDraftState {
		impl.logger.Errorw("draft is not in await Approval state", "draftId", draftId, "draftCurrentState", draftCurrentState)
		return errors.New("approval-request-not-raised")
	}
	return impl.configDraftRepository.UpdateDraftState(draftId, toUpdateDraftState, userId)
}

func (impl ConfigDraftServiceImpl) GetDraftVersionMetadata(draftId int) (*DraftVersionMetadataResponse, error) {
	draftVersionDtos, err := impl.configDraftRepository.GetDraftVersionsMetadata(draftId)
	if err != nil {
		return nil, err
	}
	var draftVersions []DraftVersionMetadata
	for _, draftVersionDto := range draftVersionDtos {
		versionMetadata := draftVersionDto.ConvertToDraftVersionMetadata()
		draftVersions = append(draftVersions, versionMetadata)
	}

	//TODO need to set email id of user
	response := &DraftVersionMetadataResponse{}
	response.DraftId = draftId
	response.DraftVersions = draftVersions
	return response, nil
}

func (impl ConfigDraftServiceImpl) GetDraftComments(draftId int) (*DraftVersionCommentResponse, error) {
	draftComments, err := impl.configDraftRepository.GetDraftVersionComments(draftId)
	if err != nil {
		return nil, err
	}
	var draftVersionVsComments map[int][]UserCommentMetadata
	for _, draftComment := range draftComments {
		draftVersionId := draftComment.DraftVersionId
		userComment := draftComment.ConvertToDraftVersionComment() //TODO email id is not set
		commentMetadataArray, ok := draftVersionVsComments[draftVersionId]
		commentMetadataArray = append(commentMetadataArray, userComment)
		if !ok {
			draftVersionVsComments[draftVersionId] = commentMetadataArray
		}
	}
	var draftVersionComments []DraftVersionComment
	for draftVersionId, userComments := range draftVersionVsComments {
		versionComment := DraftVersionComment{
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

func (impl ConfigDraftServiceImpl) GetDrafts(appId int, envId int, resourceType DraftResourceType) ([]AppConfigDraft, error) {
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

func (impl ConfigDraftServiceImpl) GetDraftById(draftId int) (*ConfigDraftResponse, error) {
	configDraft, err := impl.configDraftRepository.GetLatestConfigDraft(draftId)
	if err != nil {
		return nil, err
	}
	draftResponse := configDraft.ConvertToConfigDraft()
	return &draftResponse, nil
}
