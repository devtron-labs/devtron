/*
 * Copyright (c) 2024. Devtron Inc.
 */

package appStoreDeploymentCommon

import (
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/adapter"
	"go.uber.org/zap"
	"io/ioutil"
	"path"
)

type AppStoreDeploymentCommonServiceEnterprise interface {
	AppStoreDeploymentCommonService
	// BuildChartWithValuesAndRequirementsConfig
	BuildChartWithValuesAndRequirementsConfig(appName, valuesString, requirementsString, chartName, chartVersion string) (chartBytesArr []byte, err error)
}

type AppStoreDeploymentCommonServiceEnterpriseImpl struct {
	logger               *zap.SugaredLogger
	chartTemplateService util.ChartTemplateService
	AppStoreDeploymentCommonService
}

func NewAppStoreDeploymentCommonServiceEnterpriseImpl(appStoreDeploymentCommonService AppStoreDeploymentCommonService,
	logger *zap.SugaredLogger, chartTemplateService util.ChartTemplateService) *AppStoreDeploymentCommonServiceEnterpriseImpl {
	return &AppStoreDeploymentCommonServiceEnterpriseImpl{
		AppStoreDeploymentCommonService: appStoreDeploymentCommonService,
		logger:                          logger,
		chartTemplateService:            chartTemplateService,
	}
}

func (impl AppStoreDeploymentCommonServiceEnterpriseImpl) BuildChartWithValuesAndRequirementsConfig(appName, valuesString, requirementsString, chartName, chartVersion string) (chartBytesArr []byte, err error) {

	chartBytesArr = make([]byte, 0)
	chartCreateRequest := adapter.ParseChartCreateRequest(appName, false)
	chartCreateResponse, err := impl.CreateChartProxyAndGetPath(chartCreateRequest)
	if err != nil {
		impl.logger.Errorw("error in building chart", "err", err)
	}

	valuesFilePath := path.Join(chartCreateResponse.BuiltChartPath, "values.yaml")
	err = ioutil.WriteFile(valuesFilePath, []byte(valuesString), 0600)
	if err != nil {
		return chartBytesArr, nil
	}

	requirementsFilePath := path.Join(chartCreateResponse.BuiltChartPath, "requirements.yaml")
	err = ioutil.WriteFile(requirementsFilePath, []byte(requirementsString), 0600)
	if err != nil {
		return chartBytesArr, nil
	}

	chartBytesArr, err = impl.chartTemplateService.LoadChartInBytes(chartCreateResponse.BuiltChartPath, true, chartName, chartVersion)
	if err != nil {
		impl.logger.Errorw("error in loading chart in bytes", "err", err)
		return chartBytesArr, nil
	}

	return chartBytesArr, err
}
