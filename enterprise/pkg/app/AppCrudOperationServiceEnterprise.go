/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package app

import (
	"github.com/devtron-labs/devtron/enterprise/pkg/globalTag"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	app2 "github.com/devtron-labs/devtron/pkg/app"
	repository2 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/EAMode"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/genericNotes"
	"github.com/devtron-labs/devtron/pkg/team"
	"go.uber.org/zap"
)

type AppCrudOperationServiceEnterpriseImpl struct {
	logger           *zap.SugaredLogger
	globalTagService globalTag.GlobalTagService
	appRepository    app.AppRepository
	*app2.AppCrudOperationServiceImpl
}

func NewAppCrudOperationServiceEnterpriseImpl(appLabelRepository pipelineConfig.AppLabelRepository,
	logger *zap.SugaredLogger, appRepository app.AppRepository, userRepository repository.UserRepository, installedAppRepository repository2.InstalledAppRepository,
	globalTagService globalTag.GlobalTagService, teamRepository team.TeamRepository, genericNoteService genericNotes.GenericNoteService, gitMaterialRepository pipelineConfig.MaterialRepository,
	installedAppDbService EAMode.InstalledAppDBService) *AppCrudOperationServiceEnterpriseImpl {
	return &AppCrudOperationServiceEnterpriseImpl{
		AppCrudOperationServiceImpl: app2.NewAppCrudOperationServiceImpl(appLabelRepository, logger, appRepository, userRepository, installedAppRepository, teamRepository, genericNoteService, gitMaterialRepository, installedAppDbService),
		logger:                      logger,
		globalTagService:            globalTagService,
		appRepository:               appRepository,
	}
}

func (impl *AppCrudOperationServiceEnterpriseImpl) UpdateApp(request *bean.CreateAppDTO) (*bean.CreateAppDTO, error) {
	// validate mandatory labels against project
	// if project is changed, then no need to validate mandatory tags against new project
	app, err := impl.appRepository.FindById(request.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching app", "error", err)
		return nil, err
	}
	teamId := request.TeamId
	if teamId == 0 {
		teamId = app.TeamId
	}
	if app.TeamId == teamId {
		labelsMap := make(map[string]string)
		for _, label := range request.AppLabels {
			labelsMap[label.Key] = label.Value
		}
		err := impl.globalTagService.ValidateMandatoryLabelsForProject(teamId, labelsMap)
		if err != nil {
			return nil, err
		}
	}
	// call forward
	return impl.AppCrudOperationServiceImpl.UpdateApp(request)
}
