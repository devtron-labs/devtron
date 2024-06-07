/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package user

import (
	"time"

	repository2 "github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type UserAudit struct {
	UserId    int32
	ClientIp  string
	CreatedOn time.Time
	UpdatedOn time.Time
}

type UserAuditService interface {
	Save(userAudit *UserAudit) error
	GetLatestByUserId(userId int32) (*UserAudit, error)
	GetLatestUser() (*UserAudit, error)
	Update(userAudit *UserAudit) error
}

type UserAuditServiceImpl struct {
	logger              *zap.SugaredLogger
	userAuditRepository repository2.UserAuditRepository
}

func NewUserAuditServiceImpl(logger *zap.SugaredLogger, userAuditRepository repository2.UserAuditRepository) *UserAuditServiceImpl {
	return &UserAuditServiceImpl{
		logger:              logger,
		userAuditRepository: userAuditRepository,
	}
}

func (impl UserAuditServiceImpl) Update(userAudit *UserAudit) error {
	userId := userAudit.UserId
	impl.logger.Infow("Update user audit", "userId", userId)
	userAuditDb := &repository2.UserAudit{
		UserId:   userId,
		ClientIp: userAudit.ClientIp,
	}
	err := impl.userAuditRepository.Update(userAuditDb)
	if err != nil {
		impl.logger.Errorw("error while updating user audit log", "userId", userId, "error", err)
		return err
	}
	return nil
}

func (impl UserAuditServiceImpl) Save(userAudit *UserAudit) error {
	userId := userAudit.UserId
	impl.logger.Infow("Saving user audit", "userId", userId)
	userAuditDb := &repository2.UserAudit{
		UserId:    userId,
		ClientIp:  userAudit.ClientIp,
		CreatedOn: userAudit.CreatedOn,
		UpdatedOn: userAudit.CreatedOn, // setting updated same as created first time
	}
	err := impl.userAuditRepository.Save(userAuditDb)
	if err != nil {
		impl.logger.Errorw("error while saving user audit log", "userId", userId, "error", err)
		return err
	}
	return nil
}

func (impl UserAuditServiceImpl) GetLatestByUserId(userId int32) (*UserAudit, error) {
	impl.logger.Infow("Getting latest user audit", "userId", userId)
	userAuditDb, err := impl.userAuditRepository.GetLatestByUserId(userId)

	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		} else {
			impl.logger.Errorw("error while getting latest audit log", "userId", userId, "error", err)
			return nil, err
		}
	}

	return &UserAudit{
		UserId:    userId,
		ClientIp:  userAuditDb.ClientIp,
		CreatedOn: userAuditDb.CreatedOn,
		UpdatedOn: userAuditDb.UpdatedOn,
	}, nil
}

func (impl UserAuditServiceImpl) GetLatestUser() (*UserAudit, error) {
	impl.logger.Info("Getting latest user audit")
	userAuditDb, err := impl.userAuditRepository.GetLatestUser()

	if err != nil {
		if err == pg.ErrNoRows {
			impl.logger.Errorw("no user audits found", "err", err)
		} else {
			impl.logger.Errorw("error while getting latest user audit log", "err", err)
		}
		return nil, err
	}

	return &UserAudit{
		UserId:    userAuditDb.UserId,
		ClientIp:  userAuditDb.ClientIp,
		CreatedOn: userAuditDb.CreatedOn,
		UpdatedOn: userAuditDb.UpdatedOn,
	}, nil
}
