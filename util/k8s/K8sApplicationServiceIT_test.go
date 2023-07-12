package k8s

import (
	"github.com/devtron-labs/authenticator/client"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/terminal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPodContainersList(t *testing.T) {

	//testClusterId := 1
	k8sApplicationService := InitK8sApplicationService(t)
	t.Run("Create Ephemeral Container with valid Data,container status will be running", func(t *testing.T) {

	})

	t.Run("Create Ephemeral Container with inValid Data, container status will be terminated", func(t *testing.T) {

	})

	t.Run("Terminate Ephemeral Container with valid Data, container status will be terminated", func(t *testing.T) {

	})

	t.Run("Terminate Ephemeral Container with InValid Data,Invalid Pod Name payload, resource not found error", func(t *testing.T) {

	})

}

func InitK8sApplicationService(t *testing.T) *K8sApplicationServiceImpl {
	sugaredLogger, _ := util.InitLogger()
	config, _ := sql.GetConfig()
	runtimeConfig, err := client.GetRuntimeConfig()
	k8sUtil := util.NewK8sUtil(sugaredLogger, runtimeConfig)
	assert.Nil(t, err)
	db, _ := sql.NewDbConnection(config, sugaredLogger)
	clusterRepositoryImpl := repository.NewClusterRepositoryImpl(db, sugaredLogger)
	clusterServiceImpl := cluster.NewClusterServiceImpl(clusterRepositoryImpl, sugaredLogger, nil, nil, nil, nil, nil)
	terminalSessionHandlerImpl := terminal.NewTerminalSessionHandlerImpl(nil, clusterServiceImpl, sugaredLogger, k8sUtil)
	k8sApplicationService := NewK8sApplicationServiceImpl(sugaredLogger, clusterServiceImpl, nil, nil, nil, k8sUtil, nil, nil, terminalSessionHandlerImpl, nil)
	return k8sApplicationService
}
