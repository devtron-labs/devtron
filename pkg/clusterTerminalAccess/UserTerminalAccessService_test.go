package clusterTerminalAccess

import (
	"fmt"
	"github.com/devtron-labs/authenticator/client"
	application2 "github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/devtron-labs/devtron/client/k8s/informer"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/terminal"
	"github.com/devtron-labs/devtron/util/k8s"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewUserTerminalAccessService(t *testing.T) {
	t.SkipNow()
	t.Run("applyTemplates", func(t *testing.T) {
		sugaredLogger, _ := util.InitLogger()
		config, _ := sql.GetConfig()
		db, _ := sql.NewDbConnection(config, sugaredLogger)
		runtimeConfig, err := client.GetRuntimeConfig()
		v := informer.NewGlobalMapClusterNamespace()
		k8sInformerFactoryImpl := informer.NewK8sInformerFactoryImpl(sugaredLogger, v, runtimeConfig)
		terminalAccessRepositoryImpl := repository.NewTerminalAccessRepositoryImpl(db, sugaredLogger)
		clusterRepositoryImpl := repository2.NewClusterRepositoryImpl(db, sugaredLogger)
		k8sClientServiceImpl := application2.NewK8sClientServiceImpl(sugaredLogger, clusterRepositoryImpl)
		clusterServiceImpl := cluster.NewClusterServiceImpl(clusterRepositoryImpl, sugaredLogger, nil, k8sInformerFactoryImpl)
		//clusterServiceImpl := cluster2.NewClusterServiceImplExtended(clusterRepositoryImpl, nil, nil, sugaredLogger, nil, nil, nil, nil, nil)
		k8sApplicationService := k8s.NewK8sApplicationServiceImpl(sugaredLogger, clusterServiceImpl, nil, k8sClientServiceImpl, nil, nil, nil)
		terminalSessionHandlerImpl := terminal.NewTerminalSessionHandlerImpl(nil, clusterServiceImpl, sugaredLogger)
		terminalAccessServiceImpl, _ := NewUserTerminalAccessServiceImpl(sugaredLogger, terminalAccessRepositoryImpl, k8sApplicationService, k8sClientServiceImpl, terminalSessionHandlerImpl)
		clusterId := 2
		request := &models.UserTerminalSessionRequest{
			UserId:    2,
			ClusterId: clusterId,
			BaseImage: "trstringer/internal-kubectl:latest",
			ShellName: "sh",
		}
		startTerminalSession, err := terminalAccessServiceImpl.StartTerminalSession(request)
		if err != nil {
			return
		}
		fmt.Println(startTerminalSession)
		sessionId := ""
		for sessionId == "" {
			fetchTerminalStatus, err := terminalAccessServiceImpl.FetchTerminalStatus(startTerminalSession.TerminalAccessId)
			assert.Nil(t, err)
			sessionId = fetchTerminalStatus.UserTerminalSessionId
			fmt.Println("sessionId: ", sessionId)
			time.Sleep(1 * time.Second)
		}
		fmt.Println("SessionId: ", sessionId)

		err = terminalAccessServiceImpl.StopTerminalSession(startTerminalSession.TerminalAccessId)
		if err != nil {
			fmt.Println(err)
		}
		for true {
			fmt.Println("looping")
			time.Sleep(5 * time.Second)
		}

	})
}
