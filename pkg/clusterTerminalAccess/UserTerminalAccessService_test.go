package clusterTerminalAccess

import (
	"errors"
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
	t.Run("ClusterTerminalJourneyForSingleUserSingleSession", func(t *testing.T) {
		terminalAccessServiceImpl := initTerminalAccessService(t)
		updateTerminalSession := createAndUpdateSessionForUser(t, terminalAccessServiceImpl, 1, 2)

		err := terminalAccessServiceImpl.StopTerminalSession(updateTerminalSession.TerminalAccessId)
		assert.Nil(t, err)
		sessionId, _ := fetchSessionId(terminalAccessServiceImpl, updateTerminalSession.TerminalAccessId)
		fmt.Println("SessionId: ", sessionId)
		err = terminalAccessServiceImpl.DisconnectTerminalSession(updateTerminalSession.TerminalAccessId)
		assert.Nil(t, err)
		_, err = fetchSessionId(terminalAccessServiceImpl, updateTerminalSession.TerminalAccessId)
		assert.Equal(t, err, errors.New("pod-terminated"))
	})

	t.Run("ClusterTerminalJourneyForSingleUserMultiSession", func(t *testing.T) {
		terminalAccessServiceImpl := initTerminalAccessService(t)
		clusterId := 1
		userId := int32(2)
		terminalSession1 := createAndUpdateSessionForUser(t, terminalAccessServiceImpl, clusterId, userId)
		terminalSession2 := createAndUpdateSessionForUser(t, terminalAccessServiceImpl, clusterId, userId)
		terminalSession3 := createAndUpdateSessionForUser(t, terminalAccessServiceImpl, clusterId, userId)

		terminalAccessServiceImpl.DisconnectAllSessionsForUser(userId)
		_, err := fetchSessionId(terminalAccessServiceImpl, terminalSession1.TerminalAccessId)
		assert.Nil(t, err)
		_, err = fetchSessionId(terminalAccessServiceImpl, terminalSession2.TerminalAccessId)
		assert.Nil(t, err)
		_, err = fetchSessionId(terminalAccessServiceImpl, terminalSession3.TerminalAccessId)
		assert.Nil(t, err)

		err = terminalAccessServiceImpl.DisconnectTerminalSession(terminalSession1.TerminalAccessId)
		assert.Nil(t, err)
		err = terminalAccessServiceImpl.DisconnectTerminalSession(terminalSession2.TerminalAccessId)
		assert.Nil(t, err)
		err = terminalAccessServiceImpl.DisconnectTerminalSession(terminalSession3.TerminalAccessId)
		assert.Nil(t, err)

		_, err = fetchSessionId(terminalAccessServiceImpl, terminalSession1.TerminalAccessId)
		assert.Equal(t, err, errors.New("pod-terminated"))
		_, err = fetchSessionId(terminalAccessServiceImpl, terminalSession2.TerminalAccessId)
		assert.Equal(t, err, errors.New("pod-terminated"))
		_, err = fetchSessionId(terminalAccessServiceImpl, terminalSession3.TerminalAccessId)
		assert.Equal(t, err, errors.New("pod-terminated"))
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

func initTerminalAccessService(t *testing.T) *UserTerminalAccessServiceImpl {
	sugaredLogger, _ := util.InitLogger()
	config, _ := sql.GetConfig()
	db, _ := sql.NewDbConnection(config, sugaredLogger)
	runtimeConfig, err := client.GetRuntimeConfig()
	assert.Nil(t, err)
	v := informer.NewGlobalMapClusterNamespace()
	k8sInformerFactoryImpl := informer.NewK8sInformerFactoryImpl(sugaredLogger, v, runtimeConfig)
	terminalAccessRepositoryImpl := repository.NewTerminalAccessRepositoryImpl(db, sugaredLogger)
	clusterRepositoryImpl := repository2.NewClusterRepositoryImpl(db, sugaredLogger)
	k8sClientServiceImpl := application2.NewK8sClientServiceImpl(sugaredLogger, clusterRepositoryImpl)
	clusterServiceImpl := cluster.NewClusterServiceImpl(clusterRepositoryImpl, sugaredLogger, nil, k8sInformerFactoryImpl)
	//clusterServiceImpl := cluster2.NewClusterServiceImplExtended(clusterRepositoryImpl, nil, nil, sugaredLogger, nil, nil, nil, nil, nil)
	k8sApplicationService := k8s.NewK8sApplicationServiceImpl(sugaredLogger, clusterServiceImpl, nil, k8sClientServiceImpl, nil, nil, nil)
	terminalSessionHandlerImpl := terminal.NewTerminalSessionHandlerImpl(nil, clusterServiceImpl, sugaredLogger)
	userTerminalSessionConfig, err := GetTerminalAccessConfig()
	assert.Nil(t, err)
	userTerminalSessionConfig.TerminalPodStatusSyncTimeInSecs = 60
	userTerminalSessionConfig.TerminalPodInActiveDurationInMins = 2
	terminalAccessServiceImpl, err := NewUserTerminalAccessServiceImpl(sugaredLogger, terminalAccessRepositoryImpl, userTerminalSessionConfig, k8sApplicationService, k8sClientServiceImpl, terminalSessionHandlerImpl)
	assert.Nil(t, err)
	return terminalAccessServiceImpl
}

func createAndUpdateSessionForUser(t *testing.T, terminalAccessServiceImpl *UserTerminalAccessServiceImpl, clusterId int, userId int32) *models.UserTerminalSessionResponse {

	request := &models.UserTerminalSessionRequest{
		UserId:    userId,
		ClusterId: clusterId,
		BaseImage: "trstringer/internal-kubectl:latest",
		ShellName: "sh",
		NodeName:  "demo-new",
	}
	time.Sleep(5 * time.Second)
	startTerminalSession, err := terminalAccessServiceImpl.StartTerminalSession(request)
	assert.Nil(t, err)
	fmt.Println(startTerminalSession)
	sessionId, err := fetchSessionId(terminalAccessServiceImpl, startTerminalSession.TerminalAccessId)
	assert.Nil(t, err)
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
	sessionId, err = fetchSessionId(terminalAccessServiceImpl, updateTerminalSession.TerminalAccessId)
	assert.Nil(t, err)
	fmt.Println("SessionId: ", sessionId)

	terminalShellSession, err := terminalAccessServiceImpl.UpdateTerminalShellSession(&models.UserTerminalShellSessionRequest{TerminalAccessId: updateTerminalSession.TerminalAccessId, ShellName: "bash"})
	assert.Nil(t, err)
	assert.Equal(t, terminalShellSession.TerminalAccessId, updateTerminalSession.TerminalAccessId)
	sessionId, err = fetchSessionId(terminalAccessServiceImpl, updateTerminalSession.TerminalAccessId)
	assert.Nil(t, err)
	fmt.Println("SessionId: ", sessionId)
	return updateTerminalSession
}

func fetchSessionId(terminalAccessServiceImpl *UserTerminalAccessServiceImpl, terminalAccessId int) (string, error) {
	sessionId := ""
	for sessionId == "" {
		fetchTerminalStatus, err := terminalAccessServiceImpl.FetchTerminalStatus(terminalAccessId)
		if err != nil {
			return sessionId, err
		}
		sessionId = fetchTerminalStatus.UserTerminalSessionId
		fmt.Println("sessionId: ", sessionId)
		time.Sleep(5 * time.Second)
	}
	return sessionId, nil
}
