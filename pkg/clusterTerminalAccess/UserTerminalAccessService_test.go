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
	"k8s.io/kubernetes/pkg/api/legacyscheme"
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
		clusterId := 1
		request := &models.UserTerminalSessionRequest{
			UserId:    2,
			ClusterId: clusterId,
			BaseImage: "trstringer/internal-kubectl:latest",
			ShellName: "sh",
			NodeName:  "demo-new",
		}
		time.Sleep(5 * time.Second)
		startTerminalSession, err := terminalAccessServiceImpl.StartTerminalSession(request)
		assert.Nil(t, err)
		fmt.Println(startTerminalSession)
		sessionId := fetchSessionId(t, terminalAccessServiceImpl, startTerminalSession.TerminalAccessId)
		fmt.Println("SessionId: ", sessionId)

		shellChangeResponse, err := terminalAccessServiceImpl.UpdateTerminalShellSession(&models.UserTerminalShellSessionRequest{TerminalAccessId: startTerminalSession.TerminalAccessId, ShellName: "bash"})
		assert.Nil(t, err)
		assert.Equal(t, shellChangeResponse.TerminalAccessId, startTerminalSession.TerminalAccessId)
		request.BaseImage = "nginx:latest"
		request.Id = startTerminalSession.TerminalAccessId
		updateTerminalSession, err := terminalAccessServiceImpl.UpdateTerminalSession(request)
		assert.Nil(t, err)
		assert.Equal(t, updateTerminalSession.UserId, request.UserId)
		fmt.Println("updated: ", updateTerminalSession)

		sessionId = fetchSessionId(t, terminalAccessServiceImpl, updateTerminalSession.TerminalAccessId)
		fmt.Println("SessionId: ", sessionId)

		err = terminalAccessServiceImpl.StopTerminalSession(updateTerminalSession.TerminalAccessId)
		assert.Nil(t, err)
		sessionId = fetchSessionId(t, terminalAccessServiceImpl, updateTerminalSession.TerminalAccessId)
		fmt.Println("SessionId: ", sessionId)
		err = terminalAccessServiceImpl.DisconnectTerminalSession(updateTerminalSession.TerminalAccessId)
		assert.Nil(t, err)
		sessionId = fetchSessionId(t, terminalAccessServiceImpl, updateTerminalSession.TerminalAccessId)
		fmt.Println("SessionId: ", sessionId)
	})

	t.Run("convert to k8s structure", func(t *testing.T) {
		podJson := "{\"apiVersion\":\"rbac.authorization.k8s.io/v1\",\"kind\":\"ClusterRoleBinding\",\"metadata\":{\"name\":\"${pod_name}-crb\"},\"subjects\":[{\"kind\":\"ServiceAccount\",\"name\":\"${pod_name}-sa\",\"namespace\":\"${default_namespace}\"}],\"roleRef\":{\"kind\":\"ClusterRole\",\"name\":\"cluster-admin\",\"apiGroup\":\"rbac.authorization.k8s.io\"}}"
		_, groupVersionKind, err := legacyscheme.Codecs.UniversalDeserializer().Decode([]byte(podJson), nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, groupVersionKind.Group, "rbac.authorization.k8s.io")
		assert.Equal(t, groupVersionKind.Version, "v1")
		assert.Equal(t, groupVersionKind.Kind, "ClusterRoleBinding")

		podJson = "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"name\":\"${pod_name}\"},\"spec\":{\"serviceAccountName\":\"${pod_name}-sa\",\"nodeSelector\":{\"kubernetes.io/hostname\":\"${node_name}\"},\"containers\":[{\"name\":\"internal-kubectl\",\"image\":\"${base_image}\",\"command\":[\"/bin/bash\",\"-c\",\"--\"],\"args\":[\"while true; do sleep 30; done;\"]}]}}"
		_, groupVersionKind, err = legacyscheme.Codecs.UniversalDeserializer().Decode([]byte(podJson), nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, groupVersionKind.Group, "")
		assert.Equal(t, groupVersionKind.Version, "v1")
		assert.Equal(t, groupVersionKind.Kind, "Pod")
	})
}

func fetchSessionId(t *testing.T, terminalAccessServiceImpl *UserTerminalAccessServiceImpl, terminalAccessId int) string {
	sessionId := ""
	for sessionId == "" {
		fetchTerminalStatus, err := terminalAccessServiceImpl.FetchTerminalStatus(terminalAccessId)
		assert.Nil(t, err)
		sessionId = fetchTerminalStatus.UserTerminalSessionId
		fmt.Println("sessionId: ", sessionId)
		time.Sleep(5 * time.Second)
	}
	return sessionId
}
