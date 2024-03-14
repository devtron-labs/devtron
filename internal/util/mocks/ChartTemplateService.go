// Code generated by mockery v2.32.0. DO NOT EDIT.

package mocks

import (
	context "context"

	chart "k8s.io/helm/pkg/proto/hapi/chart"

	mock "github.com/stretchr/testify/mock"

	util "github.com/devtron-labs/devtron/internal/util"
)

// ChartTemplateService is an autogenerated mock type for the ChartTemplateService type
type ChartTemplateService struct {
	mock.Mock
}

// BuildChart provides a mock function with given fields: ctx, chartMetaData, referenceTemplatePath
func (_m *ChartTemplateService) BuildChart(ctx context.Context, chartMetaData *chart.Metadata, referenceTemplatePath string) (string, error) {
	ret := _m.Called(ctx, chartMetaData, referenceTemplatePath)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *chart.Metadata, string) (string, error)); ok {
		return rf(ctx, chartMetaData, referenceTemplatePath)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *chart.Metadata, string) string); ok {
		r0 = rf(ctx, chartMetaData, referenceTemplatePath)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *chart.Metadata, string) error); ok {
		r1 = rf(ctx, chartMetaData, referenceTemplatePath)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BuildChartProxyForHelmApps provides a mock function with given fields: chartCreateRequest
func (_m *ChartTemplateService) BuildChartProxyForHelmApps(chartCreateRequest *util.ChartCreateRequest) (*util.ChartCreateResponse, error) {
	ret := _m.Called(chartCreateRequest)

	var r0 *util.ChartCreateResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(*util.ChartCreateRequest) (*util.ChartCreateResponse, error)); ok {
		return rf(chartCreateRequest)
	}
	if rf, ok := ret.Get(0).(func(*util.ChartCreateRequest) *util.ChartCreateResponse); ok {
		r0 = rf(chartCreateRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*util.ChartCreateResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(*util.ChartCreateRequest) error); ok {
		r1 = rf(chartCreateRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CleanDir provides a mock function with given fields: dir
func (_m *ChartTemplateService) CleanDir(dir string) {
	_m.Called(dir)
}

// CreateZipFileForChart provides a mock function with given fields: _a0, outputChartPathDir
func (_m *ChartTemplateService) CreateZipFileForChart(_a0 *chart.Chart, outputChartPathDir string) ([]byte, error) {
	ret := _m.Called(_a0, outputChartPathDir)

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(*chart.Chart, string) ([]byte, error)); ok {
		return rf(_a0, outputChartPathDir)
	}
	if rf, ok := ret.Get(0).(func(*chart.Chart, string) []byte); ok {
		r0 = rf(_a0, outputChartPathDir)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(*chart.Chart, string) error); ok {
		r1 = rf(_a0, outputChartPathDir)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FetchValuesFromReferenceChart provides a mock function with given fields: chartMetaData, refChartLocation, templateName, userId, pipelineStrategyPath
func (_m *ChartTemplateService) FetchValuesFromReferenceChart(chartMetaData *chart.Metadata, refChartLocation string, templateName string, userId int32, pipelineStrategyPath string) (*util.ChartValues, error) {
	ret := _m.Called(chartMetaData, refChartLocation, templateName, userId, pipelineStrategyPath)

	var r0 *util.ChartValues
	var r1 error
	if rf, ok := ret.Get(0).(func(*chart.Metadata, string, string, int32, string) (*util.ChartValues, error)); ok {
		return rf(chartMetaData, refChartLocation, templateName, userId, pipelineStrategyPath)
	}
	if rf, ok := ret.Get(0).(func(*chart.Metadata, string, string, int32, string) *util.ChartValues); ok {
		r0 = rf(chartMetaData, refChartLocation, templateName, userId, pipelineStrategyPath)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*util.ChartValues)
		}
	}

	if rf, ok := ret.Get(1).(func(*chart.Metadata, string, string, int32, string) error); ok {
		r1 = rf(chartMetaData, refChartLocation, templateName, userId, pipelineStrategyPath)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByteArrayRefChart provides a mock function with given fields: chartMetaData, referenceTemplatePath
func (_m *ChartTemplateService) GetByteArrayRefChart(chartMetaData *chart.Metadata, referenceTemplatePath string) ([]byte, error) {
	ret := _m.Called(chartMetaData, referenceTemplatePath)

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(*chart.Metadata, string) ([]byte, error)); ok {
		return rf(chartMetaData, referenceTemplatePath)
	}
	if rf, ok := ret.Get(0).(func(*chart.Metadata, string) []byte); ok {
		r0 = rf(chartMetaData, referenceTemplatePath)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(*chart.Metadata, string) error); ok {
		r1 = rf(chartMetaData, referenceTemplatePath)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetChartVersion provides a mock function with given fields: location
func (_m *ChartTemplateService) GetChartVersion(location string) (string, error) {
	ret := _m.Called(location)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (string, error)); ok {
		return rf(location)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(location)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(location)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetDir provides a mock function with given fields:
func (_m *ChartTemplateService) GetDir() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// LoadChartFromDir provides a mock function with given fields: dir
func (_m *ChartTemplateService) LoadChartFromDir(dir string) (*chart.Chart, error) {
	ret := _m.Called(dir)

	var r0 *chart.Chart
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*chart.Chart, error)); ok {
		return rf(dir)
	}
	if rf, ok := ret.Get(0).(func(string) *chart.Chart); ok {
		r0 = rf(dir)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*chart.Chart)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(dir)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadChartInBytes provides a mock function with given fields: ChartPath, deleteChart
func (_m *ChartTemplateService) LoadChartInBytes(ChartPath string, deleteChart bool) ([]byte, error) {
	ret := _m.Called(ChartPath, deleteChart)

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(string, bool) ([]byte, error)); ok {
		return rf(ChartPath, deleteChart)
	}
	if rf, ok := ret.Get(0).(func(string, bool) []byte); ok {
		r0 = rf(ChartPath, deleteChart)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(string, bool) error); ok {
		r1 = rf(ChartPath, deleteChart)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PackageChart provides a mock function with given fields: tempReferenceTemplateDir, chartMetaData
func (_m *ChartTemplateService) PackageChart(tempReferenceTemplateDir string, chartMetaData *chart.Metadata) (*string, string, error) {
	ret := _m.Called(tempReferenceTemplateDir, chartMetaData)

	var r0 *string
	var r1 string
	var r2 error
	if rf, ok := ret.Get(0).(func(string, *chart.Metadata) (*string, string, error)); ok {
		return rf(tempReferenceTemplateDir, chartMetaData)
	}
	if rf, ok := ret.Get(0).(func(string, *chart.Metadata) *string); ok {
		r0 = rf(tempReferenceTemplateDir, chartMetaData)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
		}
	}

	if rf, ok := ret.Get(1).(func(string, *chart.Metadata) string); ok {
		r1 = rf(tempReferenceTemplateDir, chartMetaData)
	} else {
		r1 = ret.Get(1).(string)
	}

	if rf, ok := ret.Get(2).(func(string, *chart.Metadata) error); ok {
		r2 = rf(tempReferenceTemplateDir, chartMetaData)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// NewChartTemplateService creates a new instance of ChartTemplateService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewChartTemplateService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ChartTemplateService {
	mock := &ChartTemplateService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
