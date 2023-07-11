package k8s

import (
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster/mocks"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPodContainersList(t *testing.T) {
	sugaredLogger, _ := util.InitLogger()
	cfg, err := sql.GetConfig()
	dbConn, err := sql.NewDbConnection(cfg, sugaredLogger)
	if err != nil {
		assert.Fail(t, "error in loading config")
	}
	clusterRepo := repository.NewClusterRepositoryImpl(dbConn, sugaredLogger)
	testClusterId := 1
	clusterBean, err := clusterRepo.FindById(testClusterId)
	assert.Nil(t, err)
	clusterCfg := clusterBean.GetClusterConfig()

	mockedClusterService := mocks.NewClusterService(t)
	mockedClusterService.On("FindById", testClusterId).Return(clusterBean, nil)
	mockedClusterService.On("GetClusterConfig", clusterBean).Return(clusterCfg, nil)

}
