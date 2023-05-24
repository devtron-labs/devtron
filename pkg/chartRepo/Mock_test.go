package chartRepo

import (
	"context"
	"github.com/devtron-labs/devtron/internal/util"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	cluster2 "github.com/devtron-labs/devtron/pkg/cluster"
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

//----------
type ClusterServiceImplMock struct {
	mock.Mock
}

//func (impl ClusterServiceImplMock) Save(chartRepo *chartRepoRepository.ChartRepo, tx *pg.Tx) error {
//	panic("implement me")
//}

func (impl ClusterServiceImplMock) Save(parent context.Context, bean *cluster2.ClusterBean, userId int32) (*cluster2.ClusterBean, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) FindOne(clusterName string) (*cluster2.ClusterBean, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) FindOneActive(clusterName string) (*cluster2.ClusterBean, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) FindAll() ([]*cluster2.ClusterBean, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) FindAllActive() ([]cluster2.ClusterBean, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) DeleteFromDb(bean *cluster2.ClusterBean, userId int32) error {
	panic("implement me")
}

func (impl ClusterServiceImplMock) FindById(id int) (*cluster2.ClusterBean, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) FindByIds(id []int) ([]cluster2.ClusterBean, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) Update(ctx context.Context, bean *cluster2.ClusterBean, userId int32) (*cluster2.ClusterBean, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) Delete(bean *cluster2.ClusterBean, userId int32) error {
	panic("implement me")
}

func (impl ClusterServiceImplMock) FindAllForAutoComplete() ([]cluster2.ClusterBean, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) CreateGrafanaDataSource(clusterBean *cluster2.ClusterBean, env *repository.Environment) (int, error) {
	panic("implement me")
}
func (impl ClusterServiceImplMock) GetClusterConfig(cluster *cluster2.ClusterBean) (*util.ClusterConfig, error) {
	panic("implement me")
}
