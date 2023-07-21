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

package genericNotes

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/genericNotes/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type GenericNoteService interface {
	Save(bean *repository.GenericNote, userId int32) (*repository.GenericNote, error)
	Update(bean *repository.GenericNote, userId int32) (*repository.GenericNote, error)
	GetGenericNotesForAppIds(appIds []int) (map[int]*bean.GenericNoteResponseBean, error)
}

type GenericNoteServiceImpl struct {
	genericNoteRepository     repository.GenericNoteRepository
	genericNoteHistoryService GenericNoteHistoryService
	userRepository            repository2.UserRepository
	logger                    *zap.SugaredLogger
}

func NewClusterNoteServiceImpl(genericNoteRepository repository.GenericNoteRepository, clusterNoteHistoryService GenericNoteHistoryService, userRepository repository2.UserRepository, logger *zap.SugaredLogger) *GenericNoteServiceImpl {
	genericNoteService := &GenericNoteServiceImpl{
		genericNoteRepository:     genericNoteRepository,
		genericNoteHistoryService: clusterNoteHistoryService,
		userRepository:            userRepository,
		logger:                    logger,
	}
	return genericNoteService
}

// TODO: This should return genericNote response bean
func (impl *GenericNoteServiceImpl) Save(bean *repository.GenericNote, userId int32) (*repository.GenericNote, error) {
	existingModel, err := impl.genericNoteRepository.FindByClusterId(bean.Identifier)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Error(err)
		return nil, err
	}
	if existingModel.Id > 0 {
		impl.logger.Errorw("error on fetching cluster, duplicate", "id", bean.Identifier)
		return nil, fmt.Errorf("cluster note already exists")
	}

	bean.CreatedBy = userId
	bean.UpdatedBy = userId
	bean.CreatedOn = time.Now()
	bean.UpdatedOn = time.Now()

	err = impl.genericNoteRepository.Save(bean)
	if err != nil {
		impl.logger.Errorw("error in saving cluster note in db", "err", err)
		return nil, err
	}

	// audit the existing description to cluster audit history
	clusterAudit := &GenericNoteHistoryBean{
		NoteId:      bean.Id,
		Description: bean.Description,
		CreatedOn:   bean.CreatedOn,
		CreatedBy:   bean.CreatedBy,
	}
	_, _ = impl.genericNoteHistoryService.Save(clusterAudit, userId)
	return bean, err
}

// TODO: this should return response bean
func (impl *GenericNoteServiceImpl) Update(bean *repository.GenericNote, userId int32) (*repository.GenericNote, error) {
	model, err := impl.genericNoteRepository.FindByClusterId(bean.Identifier)
	if err != nil {
		impl.logger.Error(err)
		return nil, err
	}
	if model.Id == 0 {
		impl.logger.Errorw("error on fetching cluster note, not found", "id", bean.Id)
		return nil, fmt.Errorf("cluster note not found")
	}
	// update the cluster description with new data
	model.Description = bean.Description
	model.UpdatedBy = userId
	model.UpdatedOn = time.Now()

	err = impl.genericNoteRepository.Update(model)
	if err != nil {
		return nil, err
	}

	// audit the existing description to cluster audit history
	clusterAudit := &GenericNoteHistoryBean{
		NoteId:      model.Id,
		Description: model.Description,
		CreatedOn:   model.CreatedOn,
		CreatedBy:   model.CreatedBy,
	}
	_, _ = impl.genericNoteHistoryService.Save(clusterAudit, userId)
	return model, err
}

func (impl *GenericNoteServiceImpl) GetGenericNotesForAppIds(appIds []int) (map[int]*bean.GenericNoteResponseBean, error) {
	appIdsToNoteMap := make(map[int]*bean.GenericNoteResponseBean)
	notes, err := impl.genericNoteRepository.GetGenericNotesForAppIds(appIds)
	if err != nil {
		return appIdsToNoteMap, err
	}
	usersMap := make(map[int32]string)
	userIds := make([]int32, 0, len(notes))
	for _, note := range notes {
		userIds = append(userIds, note.UpdatedBy)
	}

	users, err := impl.userRepository.GetByIds(userIds)
	if err != nil {
		return appIdsToNoteMap, err
	}

	for _, user := range users {
		usersMap[user.Id] = user.EmailId
	}

	for _, note := range notes {
		appIdsToNoteMap[note.Identifier] = &bean.GenericNoteResponseBean{
			Id:          note.Id,
			Description: note.Description,
			UpdatedBy:   usersMap[note.UpdatedBy],
			UpdatedOn:   note.UpdatedOn,
		}
	}

	notesNotFoundAppIds := make([]int, 0)
	for _, appId := range appIds {
		if _, ok := appIdsToNoteMap[appId]; !ok {
			notesNotFoundAppIds = append(notesNotFoundAppIds, appId)
		}
	}

	return appIdsToNoteMap, nil
}
