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

package globalTag

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/validation"
	"strings"
	"time"
)

type GlobalTagService interface {
	GetAllActiveTags() ([]*GlobalTagDto, error)
	GetActiveTagById(tagId int) (*GlobalTagDto, error)
	GetAllActiveTagsForProject(projectId int) ([]*GlobalTagDtoForProject, error)
	CreateTags(request *CreateGlobalTagsRequest, createdBy int32) error
	UpdateTags(request *UpdateGlobalTagsRequest, updatedBy int32) error
	DeleteTags(request *DeleteGlobalTagsRequest, deletedBy int32) error
	ValidateMandatoryLabelsForProject(projectId int, labels map[string]string) error
}

type GlobalTagServiceImpl struct {
	logger              *zap.SugaredLogger
	globalTagRepository GlobalTagRepository
}

func NewGlobalTagServiceImpl(logger *zap.SugaredLogger, globalTagRepository GlobalTagRepository) *GlobalTagServiceImpl {
	return &GlobalTagServiceImpl{
		logger:              logger,
		globalTagRepository: globalTagRepository,
	}
}

func (impl GlobalTagServiceImpl) GetAllActiveTags() ([]*GlobalTagDto, error) {
	impl.logger.Info("Getting all active global tags")

	// get from DB
	globalTagsFromDb, err := impl.globalTagRepository.FindAllActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while getting all active global tags from DB", "error", err)
		return nil, err
	}

	// convert to DTO
	globalTags := make([]*GlobalTagDto, 0)
	for _, globalTagFromDb := range globalTagsFromDb {
		globalTag := impl.convertGlobalTagDtoFromDBObject(globalTagFromDb)
		globalTags = append(globalTags, globalTag)
	}

	return globalTags, nil
}

func (impl GlobalTagServiceImpl) GetActiveTagById(tagId int) (*GlobalTagDto, error) {
	impl.logger.Infow("Getting active global tags", "tagId", tagId)

	// get from DB
	globalTagFromDb, err := impl.globalTagRepository.FindActiveById(tagId)
	if err != nil {
		impl.logger.Errorw("error while getting active global tag from DB", "tagId", tagId, "error", err)
		return nil, err
	}

	if globalTagFromDb == nil || globalTagFromDb.Id == 0 {
		errorMsg := fmt.Sprintf("Global tag not found for tagId - %d", tagId)
		return nil, errors.New(errorMsg)
	}

	// convert to DTO
	globalTag := impl.convertGlobalTagDtoFromDBObject(globalTagFromDb)

	return globalTag, nil
}

func (impl GlobalTagServiceImpl) GetAllActiveTagsForProject(projectId int) ([]*GlobalTagDtoForProject, error) {
	impl.logger.Infow("Getting all active global tags", "projectId", projectId)

	// get from DB
	globalTagsFromDb, err := impl.globalTagRepository.FindAllActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while getting all active global tags from DB", "error", err)
		return nil, err
	}

	// convert to DTO
	globalTags := make([]*GlobalTagDtoForProject, 0)
	for _, globalTagFromDb := range globalTagsFromDb {
		isMandatory := CheckIfTagIsMandatoryForProject(globalTagFromDb.MandatoryProjectIdsCsv, projectId)
		globalTag := &GlobalTagDtoForProject{
			Key:         globalTagFromDb.Key,
			Description: globalTagFromDb.Description,
			IsMandatory: isMandatory,
			Propagate:   globalTagFromDb.Propagate,
		}
		globalTags = append(globalTags, globalTag)
	}

	return globalTags, nil
}

func (impl GlobalTagServiceImpl) CreateTags(request *CreateGlobalTagsRequest, createdBy int32) error {
	impl.logger.Infow("Creating Global tags", "request", request, "createdBy", createdBy)

	tagKeysMap := make(map[string]bool)
	var globalTagsToSave []*GlobalTag
	// validations
	for _, tag := range request.Tags {
		key := strings.TrimSpace(tag.Key)

		// check if empty key
		if len(key) == 0 {
			return errors.New("Validation error - empty key found in the request")
		}

		// Check if array has same key or not - if same key found - return error
		if _, ok := tagKeysMap[key]; ok {
			errorMsg := fmt.Sprintf("Validation error - Duplicate tag - %s found in request", key)
			impl.logger.Errorw("Validation error while creating global tags. duplicate tag found", "tag", key)
			return errors.New(errorMsg)
		}

		// check kubernetes label key validation logic if key needs to propagate
		if tag.Propagate {
			errs := validation.IsQualifiedName(key)
			if len(errs) > 0 {
				errorMsg := fmt.Sprintf("Validation error - tag - %s is not satisfying the label key criteria", key)
				impl.logger.Errorw("error while checking if tag key valid", "errors", errs, "key", key)
				return errors.New(errorMsg)
			}
		}

		// Check if key exists with active true - if exists - return error
		exists, err := impl.globalTagRepository.CheckKeyExistsForAnyActiveTag(key)
		if err != nil {
			impl.logger.Errorw("error while checking if tag key exists in DB with active true", "error", err, "key", key)
			return err
		}
		if exists {
			errorMsg := fmt.Sprintf("Validation error - tag - %s already exists", key)
			impl.logger.Errorw("Validation error while creating global tags. tag already exists", "tag", key)
			return errors.New(errorMsg)
		}

		// insert in slice to save in DB
		globalTagsToSave = append(globalTagsToSave, &GlobalTag{
			Key:                    key,
			MandatoryProjectIdsCsv: tag.MandatoryProjectIdsCsv,
			Description:            tag.Description,
			Propagate:              tag.Propagate,
			Active:                 true,
			AuditLog:               sql.AuditLog{CreatedOn: time.Now(), CreatedBy: createdBy},
		})

		// set in map
		tagKeysMap[key] = true
	}

	// initiate TX
	dbConnection := impl.globalTagRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	err = impl.globalTagRepository.Save(globalTagsToSave, tx)
	if err != nil {
		impl.logger.Errorw("error while saving global tags", "error", err)
		return err
	}

	// commit TX
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (impl GlobalTagServiceImpl) UpdateTags(request *UpdateGlobalTagsRequest, updatedBy int32) error {
	impl.logger.Infow("Updating Global tags", "request", request, "updatedBy", updatedBy)

	// initiate TX
	dbConnection := impl.globalTagRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	// iterate -  get from DB and update in DB
	for _, tag := range request.Tags {
		key := strings.TrimSpace(tag.Key)

		// check if empty key
		if len(key) == 0 {
			return errors.New("Validation error - empty key found in the request")
		}

		tagId := tag.Id
		globalTagFromDb, err := impl.globalTagRepository.FindActiveById(tagId)
		if err != nil {
			impl.logger.Errorw("error while getting active global tag from DB", "error", err, "tagId", tagId)
			return err
		}

		// check kubernetes label key validation logic if key needs to propagate
		if tag.Propagate {
			key := globalTagFromDb.Key
			errs := validation.IsQualifiedName(key)
			if len(errs) > 0 {
				errorMsg := fmt.Sprintf("Validation error - tag - %s is not satisfying the label key criteria", key)
				impl.logger.Errorw("error while checking if tag key valid", "errors", errs, "key", key)
				return errors.New(errorMsg)
			}
		}

		// Check if key exists with active true - if exists - return error
		if globalTagFromDb.Key != key {
			exists, err := impl.globalTagRepository.CheckKeyExistsForAnyActiveTagExcludeTagId(key, tagId)
			if err != nil {
				impl.logger.Errorw("error while checking if tag key exists in DB with active true", "error", err, "key", key)
				return err
			}
			if exists {
				errorMsg := fmt.Sprintf("Validation error - tag - %s already exists", key)
				impl.logger.Errorw("Validation error while updating global tag. tag already exists", "tag", key)
				return errors.New(errorMsg)
			}
		}

		globalTagFromDb.Key = key
		globalTagFromDb.Description = tag.Description
		globalTagFromDb.MandatoryProjectIdsCsv = tag.MandatoryProjectIdsCsv
		globalTagFromDb.Propagate = tag.Propagate
		globalTagFromDb.UpdatedBy = updatedBy
		globalTagFromDb.UpdatedOn = time.Now()

		err = impl.globalTagRepository.Update(globalTagFromDb, tx)
		if err != nil {
			impl.logger.Errorw("error while updating global tag in DB", "error", err, "tagId", tagId)
			return err
		}
	}

	// commit TX
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (impl GlobalTagServiceImpl) DeleteTags(request *DeleteGlobalTagsRequest, deletedBy int32) error {
	impl.logger.Infow("Deleting Global tags", "request", request, "deletedBy", deletedBy)

	// get from DB
	var ids []int
	for _, id := range request.Ids {
		ids = append(ids, id)
	}
	globalTagsFromDb, err := impl.globalTagRepository.FindAllActiveByIds(ids)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error while getting active global tags from DB", "error", err, "ids", ids)
		return err
	}

	// initiate TX
	dbConnection := impl.globalTagRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	// iterate and mark inactive in DB
	for _, globalTagFromDb := range globalTagsFromDb {
		globalTagFromDb.Active = false
		globalTagFromDb.UpdatedOn = time.Now()
		globalTagFromDb.UpdatedBy = deletedBy
		err = impl.globalTagRepository.Update(globalTagFromDb, tx)
		if err != nil {
			impl.logger.Errorw("error while deleting global tag", "error", err, "id", globalTagFromDb.Id)
			return err
		}
	}

	// commit TX
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (impl GlobalTagServiceImpl) ValidateMandatoryLabelsForProject(projectId int, labels map[string]string) error {
	impl.logger.Infow("Validating labels", "projectId", projectId, "labels", labels)

	tags, err := impl.GetAllActiveTagsForProject(projectId)
	if err != nil {
		return err
	}
	err = CheckIfMandatoryLabelsProvided(labels, tags)
	if err != nil {
		impl.logger.Errorw("error in validating labels", "error", err)
		return err
	}
	return nil
}

func (impl GlobalTagServiceImpl) convertGlobalTagDtoFromDBObject(globalTagFromDb *GlobalTag) *GlobalTagDto {
	globalTag := &GlobalTagDto{
		Id:                     globalTagFromDb.Id,
		Key:                    globalTagFromDb.Key,
		Description:            globalTagFromDb.Description,
		MandatoryProjectIdsCsv: globalTagFromDb.MandatoryProjectIdsCsv,
		Propagate:              globalTagFromDb.Propagate,
		CreatedOnInMs:          globalTagFromDb.CreatedOn.UnixMilli(),
	}
	if !globalTagFromDb.UpdatedOn.IsZero() {
		globalTag.UpdatedOnInMs = globalTagFromDb.UpdatedOn.UnixMilli()
	}
	return globalTag
}
