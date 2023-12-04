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
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type GenericNoteService interface {
	Save(tx *pg.Tx, bean *repository.GenericNote, userId int32) (*bean.GenericNoteResponseBean, error)
	Update(bean *repository.GenericNote, userId int32) (*bean.GenericNoteResponseBean, error)
	GetGenericNotesForAppIds(appIds []int) (map[int]*bean.GenericNoteResponseBean, error)
}

type GenericNoteServiceImpl struct {
	genericNoteRepository     repository.GenericNoteRepository
	genericNoteHistoryService GenericNoteHistoryService
	userService               user.UserService
	logger                    *zap.SugaredLogger
}

func NewGenericNoteServiceImpl(genericNoteRepository repository.GenericNoteRepository, clusterNoteHistoryService GenericNoteHistoryService, userService user.UserService, logger *zap.SugaredLogger) *GenericNoteServiceImpl {
	genericNoteService := &GenericNoteServiceImpl{
		genericNoteRepository:     genericNoteRepository,
		genericNoteHistoryService: clusterNoteHistoryService,
		userService:               userService,
		logger:                    logger,
	}
	return genericNoteService
}

func (impl *GenericNoteServiceImpl) Save(tx *pg.Tx, req *repository.GenericNote, userId int32) (*bean.GenericNoteResponseBean, error) {
	existingModel, err := impl.genericNoteRepository.FindByIdentifier(req.Identifier, req.IdentifierType)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Error("error in finding generic note by identifier and identifier type", "err", err, "identifier", req.Identifier, "identifierType", req.IdentifierType)
		return nil, err
	}
	if existingModel.Id > 0 {
		impl.logger.Errorw("error on fetching cluster, duplicate", "id", req.Identifier)
		return nil, fmt.Errorf("cluster note already exists")
	}

	req.CreatedBy = userId
	req.UpdatedBy = userId
	req.CreatedOn = time.Now()
	req.UpdatedOn = time.Now()

	err = impl.genericNoteRepository.Save(tx, req)
	if err != nil {
		impl.logger.Errorw("error in saving cluster note in db", "err", err)
		return nil, err
	}

	// audit the existing description to cluster audit history
	clusterAudit := &GenericNoteHistoryBean{
		NoteId:      req.Id,
		Description: req.Description,
	}
	_, err = impl.genericNoteHistoryService.Save(tx, clusterAudit, userId)
	if err != nil {
		impl.logger.Errorw("error in saving generic note history", "err", err, "clusterAudit", clusterAudit)
		return nil, err
	}
	userEmailId, err := impl.userService.GetUserEmailById(req.UpdatedBy, false)
	if err != nil {
		impl.logger.Errorw("error in finding user by id", "userId", req.UpdatedBy, "err", err)
		return nil, err
	}

	return &bean.GenericNoteResponseBean{
		Id:          req.Id,
		Description: req.Description,
		UpdatedBy:   userEmailId,
		UpdatedOn:   req.UpdatedOn,
	}, err
}

func (impl *GenericNoteServiceImpl) Update(req *repository.GenericNote, userId int32) (*bean.GenericNoteResponseBean, error) {
	tx, err := impl.genericNoteRepository.StartTx()
	defer func() {
		err = impl.genericNoteRepository.RollbackTx(tx)
		if err != nil {
			impl.logger.Debugw("error in rolling back transaction", "err", err, "request", req, "userId", userId)
		}
	}()

	if err != nil {
		impl.logger.Errorw("error in starting db transaction", "err", err)
		return nil, err
	}

	model, err := impl.genericNoteRepository.FindByIdentifier(req.Identifier, req.IdentifierType)
	if err != nil {
		if err == pg.ErrNoRows {
			impl.logger.Debugw("id not found to update generic_note, saving new entry", "req", req, "userId", userId)
			res, err := impl.Save(tx, req, userId)
			if err == nil {
				err = impl.genericNoteRepository.CommitTx(tx)
				if err != nil {
					impl.logger.Errorw("error in committing db transaction", "err", err)
					return nil, err
				}
			}
			impl.logger.Errorw("error in saving cluster note in db", "err", err, "genericNoteReq", req)
			return res, err
		}
		impl.logger.Error("error in finding generic note by identifier and identifier type", "err", err, "identifier", req.Identifier, "identifierType", req.IdentifierType)
		return nil, err
	}

	// update the cluster description with new data
	model.Description = req.Description
	model.UpdatedBy = userId
	model.UpdatedOn = time.Now()

	err = impl.genericNoteRepository.Update(tx, model)
	if err != nil {
		impl.logger.Errorw("error occurred in genericNote update in transaction", "updateObject", model, "err", err)
		return nil, err
	}

	// audit the existing description to cluster audit history
	clusterAudit := &GenericNoteHistoryBean{
		NoteId:      model.Id,
		Description: model.Description,
	}
	_, err = impl.genericNoteHistoryService.Save(tx, clusterAudit, userId)
	if err != nil {
		impl.logger.Errorw("error in saving generic note history", "auditObject", clusterAudit)
		return nil, err
	}
	userEmailId, err := impl.userService.GetUserEmailById(model.UpdatedBy, false)
	if err != nil {
		impl.logger.Errorw("error in finding user by id", "userId", model.UpdatedBy, "err", err)
		return nil, err
	}

	err = impl.genericNoteRepository.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing db transaction in genericNote service", "err", err)
		return nil, err
	}
	return &bean.GenericNoteResponseBean{
		Id:          model.Id,
		Description: model.Description,
		UpdatedBy:   userEmailId,
		UpdatedOn:   model.UpdatedOn,
	}, err
}

func (impl *GenericNoteServiceImpl) GetGenericNotesForAppIds(appIds []int) (map[int]*bean.GenericNoteResponseBean, error) {
	appIdsToNoteMap := make(map[int]*bean.GenericNoteResponseBean)
	//get notes saved in generic note table
	notes, err := impl.genericNoteRepository.GetGenericNotesForAppIds(appIds)
	if err != nil {
		impl.logger.Errorw("error in getting generic notes for the given appIds from db", "err", err, "appIds", appIds)
		return appIdsToNoteMap, err
	}

	for _, note := range notes {
		appIdsToNoteMap[note.Identifier] = &bean.GenericNoteResponseBean{
			Id:          note.Id,
			Description: note.Description,
			UpdatedOn:   note.UpdatedOn,
		}
	}

	//filter the apps/jobs for which description is not found in generic note table
	notesNotFoundAppIds := make([]int, 0)
	for _, appId := range appIds {
		if _, ok := appIdsToNoteMap[appId]; !ok {
			notesNotFoundAppIds = append(notesNotFoundAppIds, appId)
		}
	}

	//get the description from the app table for the above notesNotFoundAppIds
	descriptions, err := impl.genericNoteRepository.GetDescriptionFromAppIds(notesNotFoundAppIds)
	if err != nil {
		impl.logger.Errorw("error in getting app.description for the given appIds from db", "err", err, "appIds", appIds)
		return appIdsToNoteMap, err
	}

	//get the users Email Ids for all the users
	usersMap := make(map[int32]string)
	userIds := make([]int32, 0, len(appIds))
	for _, note := range notes {
		userIds = append(userIds, note.UpdatedBy)
	}

	for _, desc := range descriptions {
		userIds = append(userIds, desc.UpdatedBy)
	}

	users, err := impl.userService.GetByIds(userIds)
	if err != nil {
		impl.logger.Errorw("error in finding users by userIds", "userIds", userIds, "err", err)
		return appIdsToNoteMap, err
	}

	for _, user := range users {
		usersMap[user.Id] = user.EmailId
	}

	//set the email ids in the response objects
	for _, desc := range descriptions {
		appIdsToNoteMap[desc.Identifier] = &bean.GenericNoteResponseBean{
			Id:          desc.Id,
			Description: desc.Description,
			UpdatedBy:   usersMap[desc.UpdatedBy],
			UpdatedOn:   desc.UpdatedOn,
		}
	}

	for _, note := range notes {
		appIdsToNoteMap[note.Identifier].UpdatedBy = usersMap[note.UpdatedBy]
	}

	return appIdsToNoteMap, nil
}
