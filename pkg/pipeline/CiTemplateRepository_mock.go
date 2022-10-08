package pipeline

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/stretchr/testify/mock"
)

type CiTemplateRepositoryMock struct {
	mock.Mock
}

func (impl CiTemplateRepositoryMock) Save(material *pipelineConfig.CiTemplate) error {
	//TODO implement me
	panic("implement me")
}

func (impl CiTemplateRepositoryMock) Update(material *pipelineConfig.CiTemplate) error {
	//TODO implement me
	panic("implement me")
}

func (impl CiTemplateRepositoryMock) FindByDockerRegistryId(dockerRegistryId string) (ciTemplates []*pipelineConfig.CiTemplate, err error) {
	//TODO implement me
	panic("implement me")
}

func (impl CiTemplateRepositoryMock) FindNumberOfAppsWithDockerConfigured(appIds []int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (impl *CiTemplateRepositoryMock) FindByAppId(appId int) (ciTemplate *pipelineConfig.CiTemplate, err error) {
	called := impl.Called(appId)
	ciTemplateInterface := called.Get(0)
	if ciTemplateInterface != nil {
		ciTemplate = ciTemplateInterface.(*pipelineConfig.CiTemplate)
	}
	errInterface := called.Get(1)
	if errInterface != nil {
		err = errInterface.(error)
	}
	return ciTemplate, err
}
