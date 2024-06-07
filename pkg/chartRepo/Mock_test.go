/*
 * Copyright (c) 2024. Devtron Inc.
 */

package chartRepo

import (
	"context"
	"github.com/devtron-labs/common-lib-private/utils/k8s"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/mock"
)

type ChartRepoRepositoryImplMock struct {
	mock.Mock
}

func (impl ChartRepoRepositoryImplMock) Save(chartRepo *chartRepoRepository.ChartRepo, tx *pg.Tx) error {
	panic("implement me")
}
func (impl ChartRepoRepositoryImplMock) Update(chartRepo *chartRepoRepository.ChartRepo, tx *pg.Tx) error {
	panic("implement me")
}
func (impl ChartRepoRepositoryImplMock) GetDefault() (*chartRepoRepository.ChartRepo, error) {
	panic("implement me")
}
func (impl ChartRepoRepositoryImplMock) FindById(id int) (*chartRepoRepository.ChartRepo, error) {
	panic("implement me")
}
func (impl ChartRepoRepositoryImplMock) FindAll() ([]*chartRepoRepository.ChartRepo, error) {
	panic("implement me")
}
func (impl ChartRepoRepositoryImplMock) GetConnection() *pg.DB {
	panic("implement me")
}
func (impl ChartRepoRepositoryImplMock) MarkChartRepoDeleted(chartRepo *chartRepoRepository.ChartRepo, tx *pg.Tx) error {
	panic("implement me")
}

// ----------
type ClusterServiceImplMock struct {
	mock.Mock
}

//func (impl ClusterServiceImplMock) Save(chartRepo *chartRepoRepository.ChartRepo, tx *pg.Tx) error {
//	panic("implement me")
//}

func (impl ClusterServiceImplMock) Save(parent context.Context, bean *bean.ClusterBean, userId int32) (*bean.ClusterBean, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) FindOne(clusterName string) (*bean.ClusterBean, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) FindOneActive(clusterName string) (*bean.ClusterBean, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) FindAll() ([]*bean.ClusterBean, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) FindAllActive() ([]bean.ClusterBean, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) DeleteFromDb(bean *bean.ClusterBean, userId int32) error {
	panic("implement me")
}

func (impl ClusterServiceImplMock) FindById(id int) (*bean.ClusterBean, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) FindByIds(id []int) ([]bean.ClusterBean, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) Update(ctx context.Context, bean *bean.ClusterBean, userId int32) (*bean.ClusterBean, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) Delete(bean *bean.ClusterBean, userId int32) error {
	panic("implement me")
}

func (impl ClusterServiceImplMock) FindAllForAutoComplete() ([]bean.ClusterBean, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) CreateGrafanaDataSource(clusterBean *bean.ClusterBean, env *repository.Environment) (int, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) GetClusterConfig(cluster *bean.ClusterBean) (*k8s.ClusterConfig, error) {
	panic("implement me")
}
