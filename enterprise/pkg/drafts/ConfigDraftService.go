package drafts

import (
	"errors"
	"go.uber.org/zap"
)

type ConfigDraftService interface {
	CreateDraft(request ConfigDraftRequest) (*ConfigDraftResponse, error)
	AddDraftVersion(request ConfigDraftVersionRequest) (int, error)
	UpdateDraftState(draftId int, draftState DraftState, userId int32) error
	GetDraftVersionMetadata(draftId int) error // would return version timestamp and user email id
	GetDraftComments(draftId int) error
	GetDrafts(appIds []int, envId int, resourceType DraftResourceType) error // need to take care of secret data
	GetDraftById(draftId int)                                                //  need to send ** in case of view only user for Secret data
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
	latestDraftVersion, err := impl.configDraftRepository.GetLatestDraftVersion(request.DraftId)
	if err != nil {
		return 0, err
	}
	lastDraftVersionId := request.LastDraftVersionId
	if latestDraftVersion > lastDraftVersionId {
		return 0, errors.New("last-version-outdated")
	}

	if len(request.Data) > 0 {
		draftVersionDto := DraftVersionDto{}
		draftVersionDto.DraftMetadataId = request.DraftId
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
		draftVersionCommentDto.DraftMetadataId = request.DraftId
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

func (impl ConfigDraftServiceImpl) GetDraftVersionMetadata(draftId int) error {
	panic("implement me")
}

func (impl ConfigDraftServiceImpl) GetDraftComments(draftId int) error {
	panic("implement me")
}

func (impl ConfigDraftServiceImpl) GetDrafts(appIds []int, envId int, resourceType DraftResourceType) error {
	panic("implement me")
}

func (impl ConfigDraftServiceImpl) GetDraftById(draftId int) {
	panic("implement me")
}
