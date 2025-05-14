/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package gitHost

import (
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/build/git/gitHost/bean"
	"github.com/devtron-labs/devtron/pkg/build/git/gitHost/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"time"
)

type GitHostConfig interface {
	Create(request *bean2.GitHostRequest) (int, error)
}

type GitHostConfigImpl struct {
	logger      *zap.SugaredLogger
	gitHostRepo repository.GitHostRepository
}

func NewGitHostConfigImpl(gitHostRepo repository.GitHostRepository, logger *zap.SugaredLogger) *GitHostConfigImpl {
	return &GitHostConfigImpl{
		logger:      logger,
		gitHostRepo: gitHostRepo,
	}
}

// Create in DB
func (impl GitHostConfigImpl) Create(request *bean2.GitHostRequest) (int, error) {
	impl.logger.Debugw("get git host create request", "req", request)
	exist, err := impl.gitHostRepo.Exists(request.Name)
	if err != nil {
		impl.logger.Errorw("error in fetching git host ", "name", request.Name, "err", err)
		err = &util.ApiError{
			InternalMessage: "git host creation failed, error in fetching by name",
			UserMessage:     "git host creation failed, error in fetching by name",
		}
		return 0, err
	}
	if exist {
		impl.logger.Warnw("git host already exists", "name", request.Name)
		err = &util.ApiError{
			Code:            constants.GitHostCreateFailedAlreadyExists,
			InternalMessage: "git host already exists",
			UserMessage:     "git host already exists",
		}
		return 0, errors.NewAlreadyExists(err, request.Name)
	}
	gitHost := &repository.GitHost{
		Name:        GetUniqueGitHostName(request.Name),
		DisplayName: request.Name,
		Active:      request.Active,
		AuditLog:    sql.AuditLog{CreatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: request.UserId},
	}
	err = impl.gitHostRepo.Save(gitHost)
	if err != nil {
		impl.logger.Errorw("error in saving git host in db", "data", gitHost, "err", err)
		err = &util.ApiError{
			Code:            constants.GitHostCreateFailedInDb,
			InternalMessage: "git host failed to create in db",
			UserMessage:     "git host failed to create in db",
		}
		return 0, err
	}
	return gitHost.Id, nil
}

func GetUniqueGitHostName(displayName string) string {
	return displayName + "_" + util2.GetRandomStringOfGivenLength(6)
}
