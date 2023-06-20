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

package pipeline

import (
	"encoding/json"
	"errors"
	"github.com/caarlos0/env"
	repository "github.com/devtron-labs/devtron/internal/sql/repository/imageTagging"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strings"
	"time"
)

const TagsKey = "tags"
const CommentKey = "comment"
const DuplicateTagsInAppError = "cannot create duplicate tags in the same app"

type ImageTaggingServiceConfig struct {
	HideImageTaggingHardDelete bool `env:"HIDE_IMAGE_TAGGING_HARD_DELETE" envDefault:"false"`
}

type ImageTaggingResponseDTO struct {
	ImageReleaseTags           []*repository.ImageTag   `json:"imageReleaseTags"`
	AppReleaseTags             []string                 `json:"appReleaseTags"`
	ImageComment               *repository.ImageComment `json:"imageComment"`
	ProdEnvExists              bool                     `json:"tagsEditable"`
	HideImageTaggingHardDelete bool                     `json:"hideImageTaggingHardDelete"`
}

type ImageTaggingRequestDTO struct {
	CreateTags     []*repository.ImageTag  `json:"createTags"`
	SoftDeleteTags []*repository.ImageTag  `json:"softDeleteTags"`
	ImageComment   repository.ImageComment `json:"imageComment"`
	HardDeleteTags []*repository.ImageTag  `json:"hardDeleteTags"`
}

type ImageTaggingService interface {
	// GetTagsData returns the following fields in reponse Object
	//ImageReleaseTags -> this will get the tags of the artifact,
	//AppReleaseTags -> all the tags of the given appId,
	//imageComment -> comment of the given artifactId,
	// ProdEnvExists -> implies the existence of prod environment in any workflow of given ciPipelineId or its child ciPipeline's
	GetTagsData(ciPipelineId, appId, artifactId int) (*ImageTaggingResponseDTO, error)
	CreateOrUpdateImageTagging(ciPipelineId, appId, artifactId, userId int, imageTaggingRequest *ImageTaggingRequestDTO) (*ImageTaggingResponseDTO, error)
	GetProdEnvFromParentAndLinkedWorkflow(ciPipelineId int) (bool, error)
	GetProdEnvByCdPipelineId(pipelineId int) (bool, error)
	// ValidateImageTaggingRequest validates the requested payload
	ValidateImageTaggingRequest(imageTaggingRequest *ImageTaggingRequestDTO, appId, artifactId int) (bool, error)
	GetTagsByArtifactId(artifactId int) ([]*repository.ImageTag, error)
	// GetTaggingDataMapByAppId this will fetch a map of artifact vs []tags for given appId
	GetTagsDataMapByAppId(appId int) (map[int][]*repository.ImageTag, error)
	// GetTaggingDataMapByAppId this will fetch a map of artifact vs imageComment for given artifactIds
	GetImageCommentsDataMapByArtifactIds(artifactIds []int) (map[int]*repository.ImageComment, error)
	// GetUniqueTagsByAppId gets all the unique tag names for the given appId
	GetUniqueTagsByAppId(appId int) ([]string, error)
	GetImageTaggingServiceConfig() ImageTaggingServiceConfig
}

type ImageTaggingServiceImpl struct {
	imageTaggingRepo          repository.ImageTaggingRepository
	ciPipelineRepository      pipelineConfig.CiPipelineRepository
	cdPipelineRepository      pipelineConfig.PipelineRepository
	environmentRepository     repository2.EnvironmentRepository
	logger                    *zap.SugaredLogger
	imageTaggingServiceConfig *ImageTaggingServiceConfig
}

func NewImageTaggingServiceImpl(imageTaggingRepo repository.ImageTaggingRepository,
	ciPipelineRepository pipelineConfig.CiPipelineRepository,
	cdPipelineRepository pipelineConfig.PipelineRepository,
	environmentRepository repository2.EnvironmentRepository,
	logger *zap.SugaredLogger) *ImageTaggingServiceImpl {
	imageTaggingServiceConfig := &ImageTaggingServiceConfig{}
	err := env.Parse(imageTaggingServiceConfig)
	if err != nil {
		logger.Infow("error occurred while parsing ImageTaggingServiceConfig,so setting HIDE_IMAGE_TAGGING_HARD_DELETE to default value", "err", err)
	}
	return &ImageTaggingServiceImpl{
		imageTaggingRepo:          imageTaggingRepo,
		ciPipelineRepository:      ciPipelineRepository,
		cdPipelineRepository:      cdPipelineRepository,
		environmentRepository:     environmentRepository,
		logger:                    logger,
		imageTaggingServiceConfig: imageTaggingServiceConfig,
	}
}

func (impl ImageTaggingServiceImpl) GetImageTaggingServiceConfig() ImageTaggingServiceConfig {
	return *impl.imageTaggingServiceConfig
}

// GetTagsData returns the following fields in reponse Object
//ImageReleaseTags -> this will get the tags of the artifact,
//AppReleaseTags -> all the tags of the given appId,
//imageComment -> comment of the given artifactId,
// ProdEnvExists -> implies the existence of prod environment in any workflow of given ciPipelineId or its child ciPipeline's
func (impl ImageTaggingServiceImpl) GetTagsData(ciPipelineId, appId, artifactId int) (*ImageTaggingResponseDTO, error) {
	resp := &ImageTaggingResponseDTO{}
	imageComment, err := impl.imageTaggingRepo.GetImageComment(artifactId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching image comment using artifactId", "err", err, "artifactId", artifactId)
		return resp, err
	}
	appReleaseTags, err := impl.GetUniqueTagsByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching image tags using appId", "err", err, "appId", appId)
		return resp, err
	}
	imageReleaseTags, err := impl.GetTagsByArtifactId(artifactId)
	if err != nil {
		impl.logger.Errorw("error in fetching image tags using artifactId", "err", err, "artifactId", artifactId)
		return resp, err
	}
	prodEnvExists, err := impl.GetProdEnvFromParentAndLinkedWorkflow(ciPipelineId)
	if err != nil {
		impl.logger.Errorw("error in GetProdEnvFromParentAndLinkedWorkflow", "err", err, "ciPipelineId", ciPipelineId)
		return resp, err
	}
	resp.AppReleaseTags = appReleaseTags
	resp.ImageReleaseTags = imageReleaseTags
	resp.ImageComment = &imageComment
	resp.ProdEnvExists = prodEnvExists
	resp.HideImageTaggingHardDelete = impl.imageTaggingServiceConfig.HideImageTaggingHardDelete
	return resp, err
}

func (impl ImageTaggingServiceImpl) GetTagsByArtifactId(artifactId int) ([]*repository.ImageTag, error) {
	imageReleaseTags, err := impl.imageTaggingRepo.GetTagsByArtifactId(artifactId)
	if err != nil && err != pg.ErrNoRows {
		//log error
		impl.logger.Errorw("error in fetching image tags using artifactId", "err", err, "artifactId", artifactId)
		return imageReleaseTags, err
	}
	return imageReleaseTags, nil
}

// GetTaggingDataMapByAppId this will fetch a map of artifact vs []tags for given appId
func (impl ImageTaggingServiceImpl) GetTagsDataMapByAppId(appId int) (map[int][]*repository.ImageTag, error) {
	tags, err := impl.imageTaggingRepo.GetTagsByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error occurred in getting image tags by appId", "appId", appId, "err", err)
		return nil, err
	}
	result := make(map[int][]*repository.ImageTag)
	for _, tag := range tags {
		if _, ok := result[tag.ArtifactId]; !ok {
			result[tag.ArtifactId] = make([]*repository.ImageTag, 0)

		}
		result[tag.ArtifactId] = append(result[tag.ArtifactId], tag)
	}

	return result, nil

}

func (impl ImageTaggingServiceImpl) GetImageCommentsDataMapByArtifactIds(artifactIds []int) (map[int]*repository.ImageComment, error) {
	result := make(map[int]*repository.ImageComment)

	imageComments, err := impl.imageTaggingRepo.GetImageCommentsByArtifactIds(artifactIds)
	if err != nil && err != pg.ErrNoRows {
		//log error
		impl.logger.Errorw("error in fetching imageComments using appId", "artifactIds", artifactIds)
		return nil, err
	}

	//it may be possible that there are no tags for a artifact,but comment exists
	for _, comment := range imageComments {
		result[comment.ArtifactId] = comment
	}

	return result, nil
}

func (impl ImageTaggingServiceImpl) ValidateImageTaggingRequest(imageTaggingRequest *ImageTaggingRequestDTO, appId, artifactId int) (bool, error) {
	if imageTaggingRequest == nil {
		return false, errors.New("inValid payload")
	}
	//validate create tags
	for _, tags := range imageTaggingRequest.CreateTags {
		if tags.Id != 0 {
			return false, errors.New("bad request,create tags cannot contain id")
		}
		if (tags.AppId != 0 && tags.AppId != appId) || (tags.ArtifactId != 0 && tags.ArtifactId != artifactId) {
			return false, errors.New("bad request,appId or artifactId mismatch in one of the tag with the request")
		}
		err := tagNameValidation(tags.TagName)
		if err != nil {
			return false, err
		}
		tags.TagName = strings.ToLower(tags.TagName)
	}
	//validate update tags
	for _, tags := range imageTaggingRequest.SoftDeleteTags {
		if tags.Id == 0 {
			return false, errors.New("bad request,tags requested to delete should contain id")
		}
		if (tags.AppId != 0 && tags.AppId != appId) || (tags.ArtifactId != 0 && tags.ArtifactId != artifactId) {
			return false, errors.New("bad request,appId or artifactId mismatch in one of the tag with the request")
		}
		err := tagNameValidation(tags.TagName)
		if err != nil {
			return false, err
		}
		tags.TagName = strings.ToLower(tags.TagName)
	}

	for _, tags := range imageTaggingRequest.HardDeleteTags {
		if tags.Id == 0 {
			return false, errors.New("bad request,tags requested to delete should contain id")
		}
		if (tags.AppId != 0 && tags.AppId != appId) || (tags.ArtifactId != 0 && tags.ArtifactId != artifactId) {
			return false, errors.New("bad request,appId or artifactId mismatch in one of the tag with the request")
		}
		err := tagNameValidation(tags.TagName)
		if err != nil {
			return false, err
		}
		tags.TagName = strings.ToLower(tags.TagName)
	}
	if len(imageTaggingRequest.ImageComment.Comment) > 500 {
		return false, errors.New("bad request,comment has more than 500 characters")
	}
	return true, nil
}

func tagNameValidation(tag string) error {
	err := errors.New("tag name should be max of 128 characters long,tag name should not start with '.' and '-'")
	if len(tag) > 128 || len(tag) == 0 || tag[0] == '.' || tag[0] == '-' {
		return err
	}
	return nil
}
func (impl ImageTaggingServiceImpl) CreateOrUpdateImageTagging(ciPipelineId, appId, artifactId, userId int, imageTaggingRequest *ImageTaggingRequestDTO) (*ImageTaggingResponseDTO, error) {

	tx, err := impl.imageTaggingRepo.StartTx()
	defer func() {
		err = impl.imageTaggingRepo.RollbackTx(tx)
		if err != nil {
			impl.logger.Infow("error in rolling back transaction", "err", err, "ciPipelineId", ciPipelineId, "appId", appId, "artifactId", artifactId, "userId", userId, "imageTaggingRequest", imageTaggingRequest)
		}
	}()
	if err != nil {
		impl.logger.Errorw("error in creating transaction", "err", err)
		return nil, err
	}
	auditLogsList, err := impl.performTagOperationsAndGetAuditList(tx, appId, artifactId, userId, imageTaggingRequest)
	if err != nil {
		impl.logger.Errorw("error in performTagOperationsAndGetAuditList", "err", err, "appId", appId, "artifactId", artifactId, "userId", userId, "imageTaggingRequest", imageTaggingRequest)
		return nil, err
	}
	//save or update comment
	imageTaggingRequest.ImageComment.ArtifactId = artifactId
	imageTaggingRequest.ImageComment.UserId = userId
	imageCommentAudit, err := impl.getImageCommentAudit(imageTaggingRequest.ImageComment.Comment, userId, artifactId)
	if err != nil {
		return nil, err
	}

	//max one comment exists against an artifact currently,
	//fetch savedComment for this artifactId, then update or save accordingly
	savedComment, err := impl.imageTaggingRepo.GetImageComment(artifactId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting imageComment by artifactId", "err", err, "artifactId", artifactId)
		return nil, err
	}
	if savedComment.Id > 0 {
		//update only if the comment is different from saved comment
		if savedComment.Comment != imageTaggingRequest.ImageComment.Comment {
			imageTaggingRequest.ImageComment.Id = savedComment.Id
			err = impl.imageTaggingRepo.UpdateImageComment(tx, &imageTaggingRequest.ImageComment)
			if err != nil {
				impl.logger.Errorw("error in updating imageComment ", "err", err, "ImageComment", imageTaggingRequest.ImageComment)
				return nil, err
			}
			//set comment audit
			imageCommentAudit.Action = repository.ActionEdit
		}
	} else {
		err := impl.imageTaggingRepo.SaveImageComment(tx, &imageTaggingRequest.ImageComment)
		if err != nil {
			impl.logger.Errorw("error in saving imageComment ", "err", err, "ImageComment", imageTaggingRequest.ImageComment)
			return nil, err
		}
		//set comment audit
		imageCommentAudit.Action = repository.ActionSave
	}

	//add imageCommentAudit into the auditLogs list before saving audit
	auditLogsList = append(auditLogsList, imageCommentAudit)
	//save all the audts
	err = impl.imageTaggingRepo.SaveAuditLogsInBulk(tx, auditLogsList)
	if err != nil {
		impl.logger.Errorw("error in SaveAuditLogInBulk", "err", err, "auditLogsList", auditLogsList)
		return nil, err
	}
	//commit transaction
	err = impl.imageTaggingRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err, "ciPipelineId", ciPipelineId, "appId", appId, "artifactId", artifactId, "userId", userId, "imageTaggingRequest", imageTaggingRequest)
		return nil, err
	}
	return impl.GetTagsData(ciPipelineId, appId, artifactId)
}

func (impl ImageTaggingServiceImpl) performTagOperationsAndGetAuditList(tx *pg.Tx, appId, artifactId, userId int, imageTaggingRequest *ImageTaggingRequestDTO) ([]*repository.ImageTaggingAudit, error) {
	//first perform delete and then perform create operation.
	//case : user can delete existing tag and then create a new tag with same name, this is a valid request

	//soft delete tags
	softDeleteAuditTags := make([]string, len(imageTaggingRequest.SoftDeleteTags))
	for i, tag := range imageTaggingRequest.SoftDeleteTags {
		tag.AppId = appId
		tag.Deleted = true
		tag.ArtifactId = artifactId
		softDeleteAuditTags[i] = tag.TagName
	}
	//hard delete tags
	hardDeleteAuditTags := make([]string, len(imageTaggingRequest.HardDeleteTags))
	for i, tag := range imageTaggingRequest.HardDeleteTags {
		tag.AppId = appId
		tag.ArtifactId = artifactId
		hardDeleteAuditTags[i] = tag.TagName
	}
	//save release tags
	createAuditTags := make([]string, len(imageTaggingRequest.CreateTags))
	for i, tag := range imageTaggingRequest.CreateTags {
		tag.AppId = appId
		tag.ArtifactId = artifactId
		createAuditTags[i] = tag.TagName
	}

	var err error
	if len(imageTaggingRequest.SoftDeleteTags) > 0 {
		err := impl.imageTaggingRepo.UpdateReleaseTagInBulk(tx, imageTaggingRequest.SoftDeleteTags)
		if err != nil {
			impl.logger.Errorw("error in updating releaseTags in bulk", "err", err, "payLoad", imageTaggingRequest.SoftDeleteTags)
			return nil, err
		}
	}

	if len(imageTaggingRequest.HardDeleteTags) > 0 {
		err = impl.imageTaggingRepo.DeleteReleaseTagInBulk(tx, imageTaggingRequest.HardDeleteTags)
		if err != nil {
			impl.logger.Errorw("error in deleting releaseTag in bulk", "err", err, "releaseTags", imageTaggingRequest.HardDeleteTags)
			return nil, err
		}
	}

	if len(imageTaggingRequest.CreateTags) > 0 {
		err = impl.imageTaggingRepo.SaveReleaseTagsInBulk(tx, imageTaggingRequest.CreateTags)
		if err != nil {
			if strings.Contains(err.Error(), "release_tags_app_id_tag_name_key") {
				err = errors.New(DuplicateTagsInAppError)
			}
			impl.logger.Errorw("error in saving releaseTag", "err", err, "releaseTags", imageTaggingRequest.CreateTags)
			return nil, err
		}
	}
	//get tags audit list
	auditLogsList, err := impl.getImageTagAudits(softDeleteAuditTags, hardDeleteAuditTags, createAuditTags, userId, artifactId)
	if err != nil {
		impl.logger.Errorw("error in getImageTagAudits", "err", err)
		return nil, err
	}
	return auditLogsList, err
}
func (impl ImageTaggingServiceImpl) getImageTagAudits(softDeleteTags, hardDeleteTags, createTags []string, userId, artifactId int) ([]*repository.ImageTaggingAudit, error) {
	auditLogsList := make([]*repository.ImageTaggingAudit, 0)
	currentTime := time.Now()
	if len(softDeleteTags) > 0 {
		dataMap := make(map[string]interface{})
		dataMap[TagsKey] = softDeleteTags
		dataBytes, err := json.Marshal(&dataMap)
		if err != nil {
			impl.logger.Errorw("error in marshaling imageTagging data", "error", err, "data", dataMap)
			return auditLogsList, err
		}
		auditLog := &repository.ImageTaggingAudit{
			Data:       string(dataBytes),
			DataType:   repository.TagType,
			UpdatedBy:  userId,
			UpdatedOn:  currentTime,
			ArtifactId: artifactId,
			Action:     repository.ActionSoftDelete,
		}
		auditLogsList = append(auditLogsList, auditLog)
	}

	if len(hardDeleteTags) > 0 {
		dataMap := make(map[string]interface{})
		dataMap[TagsKey] = hardDeleteTags
		dataBytes, err := json.Marshal(&dataMap)
		if err != nil {
			impl.logger.Errorw("error in marshaling imageTagging data", "error", err, "data", dataMap)
			return auditLogsList, err
		}
		auditLog := &repository.ImageTaggingAudit{
			Data:       string(dataBytes),
			DataType:   repository.TagType,
			UpdatedBy:  userId,
			UpdatedOn:  currentTime,
			ArtifactId: artifactId,
			Action:     repository.ActionHardDelete,
		}
		auditLogsList = append(auditLogsList, auditLog)
	}

	if len(createTags) > 0 {
		dataMap := make(map[string]interface{})
		dataMap[TagsKey] = createTags
		dataBytes, err := json.Marshal(&dataMap)
		if err != nil {
			impl.logger.Errorw("error in marshaling imageTagging data", "error", err, "data", dataMap)
			return auditLogsList, err
		}
		auditLog := &repository.ImageTaggingAudit{
			Data:       string(dataBytes),
			DataType:   repository.TagType,
			UpdatedBy:  userId,
			UpdatedOn:  currentTime,
			ArtifactId: artifactId,
			Action:     repository.ActionSave,
		}
		auditLogsList = append(auditLogsList, auditLog)
	}

	return auditLogsList, nil

}

func (impl ImageTaggingServiceImpl) getImageCommentAudit(imageComment string, userId, artifactId int) (*repository.ImageTaggingAudit, error) {

	dataMap := make(map[string]string)
	dataMap[CommentKey] = imageComment
	dataBytes, err := json.Marshal(&dataMap)
	if err != nil {
		impl.logger.Errorw("error in marshaling imageTagging data", "error", err, "data", dataMap)
		return nil, err
	}
	auditLog := &repository.ImageTaggingAudit{
		Data:       string(dataBytes),
		DataType:   repository.CommentType,
		UpdatedBy:  userId,
		UpdatedOn:  time.Now(),
		ArtifactId: artifactId,
		//Action:     action,
	}

	return auditLog, nil
}

func (impl ImageTaggingServiceImpl) GetProdEnvFromParentAndLinkedWorkflow(ciPipelineId int) (bool, error) {
	prodEnvExists := false
	pipelines, err := impl.ciPipelineRepository.FindByParentCiPipelineId(ciPipelineId)
	if err != nil {
		//add log
		impl.logger.Errorw("error in getting all linked ciPipelineIds", "err", err, "ciPipelineId", ciPipelineId)
		return prodEnvExists, err
	}

	//get all the pipeline ids liked with the requested ciPipelineId
	pipelineIds := make([]int, len(pipelines)+1)
	pipelineIds[0] = ciPipelineId
	for i := 0; i < len(pipelines); i++ {
		pipelineIds[i+1] = pipelines[i].Id
	}

	envs, err := impl.environmentRepository.FindEnvLinkedWithCiPipelines(pipelineIds)
	if err != nil {
		//add log
		impl.logger.Errorw("error in getting envs using ciPipelineIds", "err", err, "ciPipelineIds", pipelineIds)
		return prodEnvExists, err
	}

	for _, env := range envs {
		//env is prod ,return true
		if env.Default {
			prodEnvExists = true
			break
		}
	}

	return prodEnvExists, nil

}

func (impl ImageTaggingServiceImpl) GetProdEnvByCdPipelineId(pipelineId int) (bool, error) {
	pipeline, err := impl.cdPipelineRepository.FindById(pipelineId)
	if err != nil {
		impl.logger.Errorw("error occurred in fetching cdPipeline with pipelineId", "err", err, "pipelineId", pipelineId)
		return false, err
	}
	if pipeline.Environment.Default {
		return true, nil
	}

	//CiPipelineId will be zero for external webhook ci
	if pipeline.CiPipelineId > 0 {
		return impl.GetProdEnvFromParentAndLinkedWorkflow(pipeline.CiPipelineId)
	}

	return false, nil

}

func (impl ImageTaggingServiceImpl) GetUniqueTagsByAppId(appId int) ([]string, error) {
	imageTags, err := impl.imageTaggingRepo.GetTagsByAppId(appId)
	if err != nil {
		impl.logger.Errorw("error in fetching image tags using appId", "err", err, "appId", appId)
		return nil, err
	}
	uniqueTags := make([]string, len(imageTags))
	for i, tag := range imageTags {
		uniqueTags[i] = tag.TagName
	}
	return uniqueTags, nil
}
