package user

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
	"time"
)

type HelmUserService interface {
	CreateHelmUser(userInfo *bean.UserInfo) ([]*bean.UserInfo, error)
	UpdateHelmUser(userInfo *bean.UserInfo) (*bean.UserInfo, error)
}

type HelmUserServiceImpl struct {
	logger         *zap.SugaredLogger
	userRepository HelmUserRepository
}

func NewHelmUserServiceImpl(logger *zap.SugaredLogger,
	userRepository HelmUserRepository) *HelmUserServiceImpl {
	serviceImpl := &HelmUserServiceImpl{
		logger:         logger,
		userRepository: userRepository,
	}
	cStore = sessions.NewCookieStore(randKey())
	return serviceImpl
}

func (impl HelmUserServiceImpl) CreateHelmUser(userInfo *bean.UserInfo) ([]*bean.UserInfo, error) {
	model := &HelmUserModel{}
	model.UpdatedOn = time.Now()
	model.UpdatedBy = userInfo.UserId
	model.Active = true
	model, err := impl.userRepository.CreateHelmUser(model, nil)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	return nil, nil
}

func (impl HelmUserServiceImpl) UpdateHelmUser(userInfo *bean.UserInfo) (*bean.UserInfo, error) {
	model := &HelmUserModel{}
	model.UpdatedOn = time.Now()
	model.UpdatedBy = userInfo.UserId
	model.Active = true
	model, err := impl.userRepository.UpdateHelmUser(model, nil)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	return nil, nil
}
