package draftAwareConfigService

import (
	"context"
	userBean "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	chartService "github.com/devtron-labs/devtron/pkg/chart"
	bean3 "github.com/devtron-labs/devtron/pkg/chart/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"go.uber.org/zap"
)

type DraftAwareConfigMapService interface {
	// below methods operate on cm creation and updation

	CMGlobalAddUpdate(ctx context.Context, configMapRequest *bean.ConfigDataRequest, userMetadata *userBean.UserMetadata) (*bean.ConfigDataRequest, error)
	CMEnvironmentAddUpdate(ctx context.Context, configMapRequest *bean.ConfigDataRequest, userMetadata *userBean.UserMetadata) (*bean.ConfigDataRequest, error)
	// below methods operate on cm deletion

	CMGlobalDelete(ctx context.Context, name string, deleteReq *bean.ConfigDataRequest, userMetadata *userBean.UserMetadata) (bool, error)
	CMEnvironmentDelete(ctx context.Context, name string, deleteReq *bean.ConfigDataRequest, userMetadata *userBean.UserMetadata) (bool, error)
}

type DraftAwareSecretService interface {
	// below methods operate on cm creation and updation

	CSGlobalAddUpdate(ctx context.Context, configMapRequest *bean.ConfigDataRequest, userMetadata *userBean.UserMetadata) (*bean.ConfigDataRequest, error)
	CSEnvironmentAddUpdate(ctx context.Context, configMapRequest *bean.ConfigDataRequest, userMetadata *userBean.UserMetadata) (*bean.ConfigDataRequest, error)
	// below methods operate on cm deletion

	CSGlobalDelete(ctx context.Context, name string, deleteReq *bean.ConfigDataRequest, userMetadata *userBean.UserMetadata) (bool, error)
	CSEnvironmentDelete(ctx context.Context, name string, deleteReq *bean.ConfigDataRequest, userMetadata *userBean.UserMetadata) (bool, error)
}

type DraftAwareDeploymentTemplateService interface {
	// below methods operate on deployment template

	// Create here is used for publishing base deployment template while saving dt for the first time.
	Create(ctx context.Context, templateRequest bean3.TemplateRequest, userMetadata *userBean.UserMetadata) (*bean3.TemplateRequest, error)
	// UpdateAppOverride here is used for updating base deployment template.
	UpdateAppOverride(ctx context.Context, templateRequest *bean3.TemplateRequest, token string, userMetadata *userBean.UserMetadata) (*bean3.TemplateRequest, error)
	// UpdateEnvironmentProperties here is used for updating and saving deployment template at env override level
	UpdateEnvironmentProperties(ctx context.Context, propertiesRequest *bean.EnvironmentProperties, token string, userMetadata *userBean.UserMetadata) (*bean.EnvironmentProperties, error)
	// ResetEnvironmentProperties method handles flow when a user deletes the deployment template env override.
	ResetEnvironmentProperties(ctx context.Context, propertiesRequest *bean.EnvironmentProperties, userMetadata *userBean.UserMetadata) (bool, error)
	// CreateEnvironmentPropertiesAndBaseIfNeeded is utilized when the deployment template chart version is updated and saved
	CreateEnvironmentPropertiesAndBaseIfNeeded(ctx context.Context, environmentProperties *bean.EnvironmentProperties, userMetadata *userBean.UserMetadata) (*bean.EnvironmentProperties, error)
}

type DraftAwareConfigService interface {
	DraftAwareConfigMapService
	DraftAwareSecretService
	DraftAwareDeploymentTemplateService
}
type DraftAwareConfigServiceImpl struct {
	logger                  *zap.SugaredLogger
	configMapService        pipeline.ConfigMapService
	chartService            chartService.ChartService
	propertiesConfigService pipeline.PropertiesConfigService
}

func NewDraftAwareResourceServiceImpl(logger *zap.SugaredLogger,
	configMapService pipeline.ConfigMapService,
	chartService chartService.ChartService,
	propertiesConfigService pipeline.PropertiesConfigService,
) *DraftAwareConfigServiceImpl {
	return &DraftAwareConfigServiceImpl{
		logger:                  logger,
		configMapService:        configMapService,
		chartService:            chartService,
		propertiesConfigService: propertiesConfigService,
	}
}

func (impl *DraftAwareConfigServiceImpl) CMGlobalAddUpdate(ctx context.Context, configMapRequest *bean.ConfigDataRequest, userMetadata *userBean.UserMetadata) (*bean.ConfigDataRequest, error) {
	resp, err := impl.configMapService.CMGlobalAddUpdate(configMapRequest)
	if err != nil {
		impl.logger.Errorw("error in CMGlobalAddUpdate", "configMapRequest", configMapRequest, "err", err)
		return nil, err
	}

	return resp, nil
}

func (impl *DraftAwareConfigServiceImpl) CMEnvironmentAddUpdate(ctx context.Context, configMapRequest *bean.ConfigDataRequest, userMetadata *userBean.UserMetadata) (*bean.ConfigDataRequest, error) {
	resp, err := impl.configMapService.CMEnvironmentAddUpdate(configMapRequest)
	if err != nil {
		impl.logger.Errorw("error in CMEnvironmentAddUpdate", "configMapRequest", configMapRequest, "err", err)
		return nil, err
	}

	return resp, nil
}

func (impl *DraftAwareConfigServiceImpl) CSGlobalAddUpdate(ctx context.Context, configMapRequest *bean.ConfigDataRequest, userMetadata *userBean.UserMetadata) (*bean.ConfigDataRequest, error) {
	resp, err := impl.configMapService.CSGlobalAddUpdate(configMapRequest)
	if err != nil {
		impl.logger.Errorw("error in CSGlobalAddUpdate", "err", err)
		return nil, err
	}

	return resp, nil
}

func (impl *DraftAwareConfigServiceImpl) CSEnvironmentAddUpdate(ctx context.Context, configMapRequest *bean.ConfigDataRequest, userMetadata *userBean.UserMetadata) (*bean.ConfigDataRequest, error) {
	resp, err := impl.configMapService.CSEnvironmentAddUpdate(configMapRequest)
	if err != nil {
		impl.logger.Errorw("error in CSGlobalAddUpdate", "err", err)
		return nil, err
	}

	return resp, nil

}

func (impl *DraftAwareConfigServiceImpl) CMGlobalDelete(ctx context.Context, name string, deleteReq *bean.ConfigDataRequest, userMetadata *userBean.UserMetadata) (bool, error) {
	resp, err := impl.configMapService.CMGlobalDelete(name, deleteReq.Id, deleteReq.UserId)
	if err != nil {
		impl.logger.Errorw("service err, CMGlobalDelete", "appId", deleteReq.AppId, "id", deleteReq.Id, "name", name, "err", err)
		return false, err
	}

	return resp, nil
}

func (impl *DraftAwareConfigServiceImpl) CMEnvironmentDelete(ctx context.Context, name string, deleteReq *bean.ConfigDataRequest, userMetadata *userBean.UserMetadata) (bool, error) {
	resp, err := impl.configMapService.CMEnvironmentDelete(name, deleteReq.Id, deleteReq.UserId)
	if err != nil {
		impl.logger.Errorw("service err, CMEnvironmentDelete", "appId", deleteReq.AppId, "envId", deleteReq.EnvironmentId, "id", deleteReq.Id, "err", err)
		return false, err
	}

	return resp, nil
}

func (impl *DraftAwareConfigServiceImpl) CSGlobalDelete(ctx context.Context, name string, deleteReq *bean.ConfigDataRequest, userMetadata *userBean.UserMetadata) (bool, error) {
	resp, err := impl.configMapService.CSGlobalDelete(name, deleteReq.Id, deleteReq.UserId)
	if err != nil {
		impl.logger.Errorw("service err, CSGlobalDelete", "appId", deleteReq.AppId, "id", deleteReq.Id, "name", name, "err", err)
		return false, err
	}

	return resp, nil
}

func (impl *DraftAwareConfigServiceImpl) CSEnvironmentDelete(ctx context.Context, name string, deleteReq *bean.ConfigDataRequest, userMetadata *userBean.UserMetadata) (bool, error) {
	resp, err := impl.configMapService.CSEnvironmentDelete(name, deleteReq.Id, deleteReq.UserId)
	if err != nil {
		impl.logger.Errorw("service err, CSEnvironmentDelete", "appId", deleteReq.AppId, "id", deleteReq.Id, "name", name, "err", err)
		return false, err
	}

	return resp, nil
}

func (impl *DraftAwareConfigServiceImpl) Create(ctx context.Context, templateRequest bean3.TemplateRequest, userMetadata *userBean.UserMetadata) (*bean3.TemplateRequest, error) {
	resp, err := impl.chartService.Create(templateRequest, ctx)
	if err != nil {
		impl.logger.Errorw("error in creating base deployment template", "appId", templateRequest.AppId, "err", err)
		return nil, err
	}

	return resp, nil
}

func (impl *DraftAwareConfigServiceImpl) UpdateAppOverride(ctx context.Context, templateRequest *bean3.TemplateRequest, token string, userMetadata *userBean.UserMetadata) (*bean3.TemplateRequest, error) {
	resp, err := impl.chartService.UpdateAppOverride(ctx, templateRequest)
	if err != nil {
		impl.logger.Errorw("error in updating base deployment template", "chartId", templateRequest.Id, "appId", templateRequest.AppId, "err", err)
		return nil, err
	}

	return resp, nil
}

func (impl *DraftAwareConfigServiceImpl) UpdateEnvironmentProperties(ctx context.Context, propertiesRequest *bean.EnvironmentProperties, token string, userMetadata *userBean.UserMetadata) (*bean.EnvironmentProperties, error) {
	resp, err := impl.propertiesConfigService.UpdateEnvironmentProperties(propertiesRequest.AppId, propertiesRequest, propertiesRequest.UserId)
	if err != nil {
		impl.logger.Errorw("error in creating/updating env level deployment template", "appId", propertiesRequest.AppId, "envId", propertiesRequest.EnvironmentId, "err", err)
		return nil, err
	}

	return resp, nil
}

func (impl *DraftAwareConfigServiceImpl) ResetEnvironmentProperties(ctx context.Context, propertiesRequest *bean.EnvironmentProperties, userMetadata *userBean.UserMetadata) (bool, error) {
	isSuccess, err := impl.propertiesConfigService.ResetEnvironmentProperties(propertiesRequest.Id, propertiesRequest.UserId)
	if err != nil {
		impl.logger.Errorw("service err, ResetEnvironmentProperties", "chartEnvConfigOverrideId", propertiesRequest.Id, "userId", propertiesRequest.UserId, "err", err)
		return false, err
	}

	return isSuccess, nil
}

func (impl *DraftAwareConfigServiceImpl) CreateEnvironmentPropertiesAndBaseIfNeeded(ctx context.Context, environmentProperties *bean.EnvironmentProperties, userMetadata *userBean.UserMetadata) (*bean.EnvironmentProperties, error) {
	resp, err := impl.propertiesConfigService.CreateEnvironmentPropertiesAndBaseIfNeeded(ctx, environmentProperties.AppId, environmentProperties)
	if err != nil {
		impl.logger.Errorw("error, CreateEnvironmentPropertiesAndBaseIfNeeded", "appId", environmentProperties.AppId, "req", environmentProperties, "err", err)
		return nil, err
	}

	return resp, nil
}
