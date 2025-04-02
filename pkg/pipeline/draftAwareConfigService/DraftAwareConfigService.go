package draftAwareConfigService

import (
	"context"
	chartService "github.com/devtron-labs/devtron/pkg/chart"
	bean3 "github.com/devtron-labs/devtron/pkg/chart/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"go.uber.org/zap"
)

type DraftAwareResourceService interface {
	// below methods operate on cm cs creation and updation

	CMGlobalAddUpdate(ctx context.Context, configMapRequest *bean.ConfigDataRequest) (*bean.ConfigDataRequest, error)
	CMEnvironmentAddUpdate(ctx context.Context, configMapRequest *bean.ConfigDataRequest) (*bean.ConfigDataRequest, error)
	CSGlobalAddUpdate(ctx context.Context, configMapRequest *bean.ConfigDataRequest) (*bean.ConfigDataRequest, error)
	CSEnvironmentAddUpdate(ctx context.Context, configMapRequest *bean.ConfigDataRequest) (*bean.ConfigDataRequest, error)

	// below methods operate on cm cs deletion

	CMGlobalDelete(ctx context.Context, name string, deleteReq *bean.ConfigDataRequest) (bool, error)
	CMEnvironmentDelete(ctx context.Context, name string, deleteReq *bean.ConfigDataRequest) (bool, error)
	CSGlobalDelete(ctx context.Context, name string, deleteReq *bean.ConfigDataRequest) (bool, error)
	CSEnvironmentDelete(ctx context.Context, name string, deleteReq *bean.ConfigDataRequest) (bool, error)

	// below methods operate on deployment template

	// Create here is used for publishing base deployment template while saving dt for the first time.
	Create(ctx context.Context, templateRequest bean3.TemplateRequest) (*bean3.TemplateRequest, error)
	// UpdateAppOverride here is used for updating base deployment template.
	UpdateAppOverride(ctx context.Context, templateRequest *bean3.TemplateRequest, token string) (*bean3.TemplateRequest, error)
	// UpdateEnvironmentProperties here is used for updating and saving deployment template at env override level
	UpdateEnvironmentProperties(ctx context.Context, appId int, propertiesRequest *bean.EnvironmentProperties, token string) (*bean.EnvironmentProperties, error)
	// ResetEnvironmentProperties method handles flow when a user deletes the deployment template env override.
	ResetEnvironmentProperties(ctx context.Context, propertiesRequest *bean.EnvironmentProperties) (bool, error)
	// CreateEnvironmentPropertiesAndBaseIfNeeded is utilized when the deployment template chart version is updated and saved
	CreateEnvironmentPropertiesAndBaseIfNeeded(ctx context.Context, appId int, environmentProperties *bean.EnvironmentProperties) (*bean.EnvironmentProperties, error)
}
type DraftAwareResourceServiceImpl struct {
	logger                  *zap.SugaredLogger
	configMapService        pipeline.ConfigMapService
	chartService            chartService.ChartService
	propertiesConfigService pipeline.PropertiesConfigService
}

func NewDraftAwareResourceServiceImpl(logger *zap.SugaredLogger,
	configMapService pipeline.ConfigMapService,
	chartService chartService.ChartService,
	propertiesConfigService pipeline.PropertiesConfigService,
) *DraftAwareResourceServiceImpl {
	return &DraftAwareResourceServiceImpl{
		logger:                  logger,
		configMapService:        configMapService,
		chartService:            chartService,
		propertiesConfigService: propertiesConfigService,
	}
}

func (impl *DraftAwareResourceServiceImpl) CMGlobalAddUpdate(ctx context.Context, configMapRequest *bean.ConfigDataRequest) (*bean.ConfigDataRequest, error) {
	resp, err := impl.configMapService.CMGlobalAddUpdate(configMapRequest)
	if err != nil {
		impl.logger.Errorw("error in CMGlobalAddUpdate", "configMapRequest", configMapRequest, "err", err)
		return nil, err
	}
	var resourceName string
	if len(configMapRequest.ConfigData) > 0 && configMapRequest.ConfigData[0] != nil {
		resourceName = configMapRequest.ConfigData[0].Name
	}
	err = impl.performExpressEditActionsOnCmCsForExceptionUser(ctx, resourceName, configMapRequest)
	if err != nil {
		impl.logger.Errorw("error in performing express edit actions if user is exception", "err", err)
		return nil, err
	}
	return resp, nil
}

func (impl *DraftAwareResourceServiceImpl) CMEnvironmentAddUpdate(ctx context.Context, configMapRequest *bean.ConfigDataRequest) (*bean.ConfigDataRequest, error) {
	resp, err := impl.configMapService.CMEnvironmentAddUpdate(configMapRequest)
	if err != nil {
		impl.logger.Errorw("error in CMEnvironmentAddUpdate", "configMapRequest", configMapRequest, "err", err)
		return nil, err
	}
	var resourceName string
	if len(configMapRequest.ConfigData) > 0 && configMapRequest.ConfigData[0] != nil {
		resourceName = configMapRequest.ConfigData[0].Name
	}
	err = impl.performExpressEditActionsOnCmCsForExceptionUser(ctx, resourceName, configMapRequest)
	if err != nil {
		impl.logger.Errorw("error in performing express edit actions if user is exception", "err", err)
		return nil, err
	}
	return resp, nil
}

func (impl *DraftAwareResourceServiceImpl) CSGlobalAddUpdate(ctx context.Context, configMapRequest *bean.ConfigDataRequest) (*bean.ConfigDataRequest, error) {
	resp, err := impl.configMapService.CSGlobalAddUpdate(configMapRequest)
	if err != nil {
		impl.logger.Errorw("error in CSGlobalAddUpdate", "err", err)
		return nil, err
	}
	var resourceName string
	if len(configMapRequest.ConfigData) > 0 && configMapRequest.ConfigData[0] != nil {
		resourceName = configMapRequest.ConfigData[0].Name
	}
	err = impl.performExpressEditActionsOnCmCsForExceptionUser(ctx, resourceName, configMapRequest)
	if err != nil {
		impl.logger.Errorw("error in performing express edit actions if user is exception", "err", err)
		return nil, err
	}
	return resp, nil
}

func (impl *DraftAwareResourceServiceImpl) CSEnvironmentAddUpdate(ctx context.Context, configMapRequest *bean.ConfigDataRequest) (*bean.ConfigDataRequest, error) {
	resp, err := impl.configMapService.CSEnvironmentAddUpdate(configMapRequest)
	if err != nil {
		impl.logger.Errorw("error in CSGlobalAddUpdate", "err", err)
		return nil, err
	}
	var resourceName string
	if len(configMapRequest.ConfigData) > 0 && configMapRequest.ConfigData[0] != nil {
		resourceName = configMapRequest.ConfigData[0].Name
	}
	err = impl.performExpressEditActionsOnCmCsForExceptionUser(ctx, resourceName, configMapRequest)
	if err != nil {
		impl.logger.Errorw("error in performing express edit actions if user is exception", "err", err)
		return nil, err
	}
	return resp, nil

}

func (impl *DraftAwareResourceServiceImpl) CMGlobalDelete(ctx context.Context, name string, deleteReq *bean.ConfigDataRequest) (bool, error) {
	resp, err := impl.configMapService.CMGlobalDelete(name, deleteReq.Id, deleteReq.UserId)
	if err != nil {
		impl.logger.Errorw("service err, CMGlobalDelete", "appId", deleteReq.AppId, "id", deleteReq.Id, "name", name, "err", err)
		return false, err
	}
	err = impl.performExpressEditActionsOnCmCsForExceptionUser(ctx, name, deleteReq)
	if err != nil {
		impl.logger.Errorw("error in performing express edit actions if user is exception", "err", err)
		return false, err
	}
	return resp, nil
}

func (impl *DraftAwareResourceServiceImpl) CMEnvironmentDelete(ctx context.Context, name string, deleteReq *bean.ConfigDataRequest) (bool, error) {
	resp, err := impl.configMapService.CMEnvironmentDelete(name, deleteReq.Id, deleteReq.UserId)
	if err != nil {
		impl.logger.Errorw("service err, CMEnvironmentDelete", "appId", deleteReq.AppId, "envId", deleteReq.EnvironmentId, "id", deleteReq.Id, "err", err)
		return false, err
	}
	err = impl.performExpressEditActionsOnCmCsForExceptionUser(ctx, name, deleteReq)
	if err != nil {
		impl.logger.Errorw("error in performing express edit actions if user is exception", "err", err)
		return false, err
	}
	return resp, nil
}

func (impl *DraftAwareResourceServiceImpl) CSGlobalDelete(ctx context.Context, name string, deleteReq *bean.ConfigDataRequest) (bool, error) {
	resp, err := impl.configMapService.CSGlobalDelete(name, deleteReq.Id, deleteReq.UserId)
	if err != nil {
		impl.logger.Errorw("service err, CSGlobalDelete", "appId", deleteReq.AppId, "id", deleteReq.Id, "name", name, "err", err)
		return false, err
	}
	err = impl.performExpressEditActionsOnCmCsForExceptionUser(ctx, name, deleteReq)
	if err != nil {
		impl.logger.Errorw("error in performing express edit actions if user is exception", "err", err)
		return false, err
	}
	return resp, nil
}

func (impl *DraftAwareResourceServiceImpl) CSEnvironmentDelete(ctx context.Context, name string, deleteReq *bean.ConfigDataRequest) (bool, error) {
	resp, err := impl.configMapService.CSEnvironmentDelete(name, deleteReq.Id, deleteReq.UserId)
	if err != nil {
		impl.logger.Errorw("service err, CSEnvironmentDelete", "appId", deleteReq.AppId, "id", deleteReq.Id, "name", name, "err", err)
		return false, err
	}
	err = impl.performExpressEditActionsOnCmCsForExceptionUser(ctx, name, deleteReq)
	if err != nil {
		impl.logger.Errorw("error in performing express edit actions if user is exception", "err", err)
		return false, err
	}
	return resp, nil
}

func (impl *DraftAwareResourceServiceImpl) Create(ctx context.Context, templateRequest bean3.TemplateRequest) (*bean3.TemplateRequest, error) {
	resp, err := impl.chartService.Create(templateRequest, ctx)
	if err != nil {
		impl.logger.Errorw("error in creating base deployment template", "appId", templateRequest.AppId, "err", err)
		return nil, err
	}
	err = impl.performExpressEditActionsOnDeplTemplateForExceptionUser(ctx, templateRequest.AppId, -1, "")
	if err != nil {
		impl.logger.Errorw("error in performing express edit actions if user is exception", "err", err)
		return nil, err
	}
	return resp, nil
}

func (impl *DraftAwareResourceServiceImpl) UpdateAppOverride(ctx context.Context, templateRequest *bean3.TemplateRequest, token string) (*bean3.TemplateRequest, error) {
	resp, err := impl.chartService.UpdateAppOverride(ctx, templateRequest)
	if err != nil {
		impl.logger.Errorw("error in updating base deployment template", "chartId", templateRequest.Id, "appId", templateRequest.AppId, "err", err)
		return nil, err
	}
	err = impl.performExpressEditActionsOnDeplTemplateForExceptionUser(ctx, templateRequest.AppId, -1, "")
	if err != nil {
		impl.logger.Errorw("error in performing express edit actions if user is exception", "err", err)
		return nil, err
	}
	return resp, nil
}

func (impl *DraftAwareResourceServiceImpl) UpdateEnvironmentProperties(ctx context.Context, appId int, propertiesRequest *bean.EnvironmentProperties, token string) (*bean.EnvironmentProperties, error) {
	resp, err := impl.propertiesConfigService.UpdateEnvironmentProperties(appId, propertiesRequest, propertiesRequest.UserId)
	if err != nil {
		impl.logger.Errorw("error in creating/updating env level deployment template", "appId", appId, "envId", propertiesRequest.EnvironmentId, "err", err)
		return nil, err
	}
	err = impl.performExpressEditActionsOnDeplTemplateForExceptionUser(ctx, appId, propertiesRequest.EnvironmentId, "")
	if err != nil {
		impl.logger.Errorw("error in performing express edit actions if user is exception", "err", err)
		return nil, err
	}
	return resp, nil
}

func (impl *DraftAwareResourceServiceImpl) ResetEnvironmentProperties(ctx context.Context, propertiesRequest *bean.EnvironmentProperties) (bool, error) {
	isSuccess, err := impl.propertiesConfigService.ResetEnvironmentProperties(propertiesRequest.Id, propertiesRequest.UserId)
	if err != nil {
		impl.logger.Errorw("service err, ResetEnvironmentProperties", "chartEnvConfigOverrideId", propertiesRequest.Id, "userId", propertiesRequest.UserId, "err", err)
		return false, err
	}
	err = impl.performExpressEditActionsOnDeplTemplateForExceptionUser(ctx, propertiesRequest.AppId, propertiesRequest.EnvironmentId, "")
	if err != nil {
		impl.logger.Errorw("error in performing express edit actions if user is exception", "err", err)
		return false, err
	}
	return isSuccess, nil
}

func (impl *DraftAwareResourceServiceImpl) CreateEnvironmentPropertiesAndBaseIfNeeded(ctx context.Context, appId int, environmentProperties *bean.EnvironmentProperties) (*bean.EnvironmentProperties, error) {
	resp, err := impl.propertiesConfigService.CreateEnvironmentPropertiesAndBaseIfNeeded(ctx, appId, environmentProperties)
	if err != nil {
		impl.logger.Errorw("error, CreateEnvironmentPropertiesAndBaseIfNeeded", "appId", appId, "req", environmentProperties, "err", err)
		return nil, err
	}
	err = impl.performExpressEditActionsOnDeplTemplateForExceptionUser(ctx, environmentProperties.AppId, environmentProperties.EnvironmentId, "")
	if err != nil {
		impl.logger.Errorw("error in performing express edit actions if user is exception", "err", err)
		return nil, err
	}
	return resp, nil
}
