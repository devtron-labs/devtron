package clusterTerminalAccess

import (
	"context"
	"errors"
	"fmt"
	"github.com/devtron-labs/authenticator/client"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/k8s"
	"github.com/devtron-labs/devtron/pkg/k8s/informer"
	"github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs"
	repository10 "github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/terminal"
	repository3 "github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/stretchr/testify/assert"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"testing"
	"time"
)

func TestNewUserTerminalAccessServiceIT(t *testing.T) {
	//t.SkipNow()
	terminalAccessServiceImpl := initTerminalAccessService(t)
	t.Run("wrongImage", func(t *testing.T) {
		//t.SkipNow()
		baseImage := "devtron/randomimage"
		//baseImage = "trstringer/internal-kubectl"
		clusterId := 1
		userId := int32(2)
		request := &models.UserTerminalSessionRequest{
			UserId:    userId,
			ClusterId: clusterId,
			BaseImage: baseImage,
			ShellName: "sh",
			NodeName:  "demo-new",
		}
		time.Sleep(5 * time.Second)
		terminalSessionResponse, err := terminalAccessServiceImpl.StartTerminalSession(context.Background(), request)
		assert.Nil(t, err)
		fmt.Println(terminalSessionResponse)
		podManifest, err := terminalAccessServiceImpl.FetchPodManifest(context.Background(), terminalSessionResponse.TerminalAccessId)
		assert.Nil(t, err)
		fmt.Println("manifest", podManifest.Manifest)
		podEvents, err := terminalAccessServiceImpl.FetchPodEvents(context.Background(), terminalSessionResponse.TerminalAccessId)
		assert.Nil(t, err)
		fmt.Println(podEvents)
		sessionId, err := fetchSessionId(terminalAccessServiceImpl, terminalSessionResponse.TerminalAccessId)
		assert.Equal(t, err, errors.New("pod-terminated"))
		assert.Empty(t, sessionId)
		podManifest, err = terminalAccessServiceImpl.FetchPodManifest(context.Background(), terminalSessionResponse.TerminalAccessId)
		assert.Equal(t, err, errors.New("pod-terminated"))
		podEvents, err = terminalAccessServiceImpl.FetchPodEvents(context.Background(), terminalSessionResponse.TerminalAccessId)
		assert.Equal(t, err, errors.New("pod-terminated"))
	})

	t.Run("ClusterTerminalJourneyForSingleUserSingleSession", func(t *testing.T) {
		//t.SkipNow()
		baseImage := "trstringer/internal-kubectl:latest"
		updateTerminalSession := createAndUpdateSessionForUser(t, terminalAccessServiceImpl, 1, 2, baseImage)

		terminalAccessServiceImpl.StopTerminalSession(context.Background(), updateTerminalSession.TerminalAccessId)
		sessionId, _ := fetchSessionId(terminalAccessServiceImpl, updateTerminalSession.TerminalAccessId)
		fmt.Println("SessionId: ", sessionId)
		err := terminalAccessServiceImpl.DisconnectTerminalSession(context.Background(), updateTerminalSession.TerminalAccessId)
		assert.Nil(t, err)
		_, err = fetchSessionId(terminalAccessServiceImpl, updateTerminalSession.TerminalAccessId)
		assert.Equal(t, err, errors.New("pod-terminated"))
	})

	t.Run("ClusterTerminalJourneyForSingleUserMultiSession", func(t *testing.T) {
		//t.SkipNow()
		clusterId := 1
		userId := int32(2)
		baseImage := "trstringer/internal-kubectl:latest"
		terminalSession1 := createAndUpdateSessionForUser(t, terminalAccessServiceImpl, clusterId, userId, baseImage)
		terminalSession2 := createAndUpdateSessionForUser(t, terminalAccessServiceImpl, clusterId, userId, baseImage)
		terminalSession3 := createAndUpdateSessionForUser(t, terminalAccessServiceImpl, clusterId, userId, baseImage)

		terminalAccessServiceImpl.DisconnectAllSessionsForUser(context.Background(), userId)
		_, err := fetchSessionId(terminalAccessServiceImpl, terminalSession1.TerminalAccessId)
		assert.Nil(t, err)
		_, err = fetchSessionId(terminalAccessServiceImpl, terminalSession2.TerminalAccessId)
		assert.Nil(t, err)
		_, err = fetchSessionId(terminalAccessServiceImpl, terminalSession3.TerminalAccessId)
		assert.Nil(t, err)

		err = terminalAccessServiceImpl.DisconnectTerminalSession(context.Background(), terminalSession1.TerminalAccessId)
		assert.Nil(t, err)
		err = terminalAccessServiceImpl.DisconnectTerminalSession(context.Background(), terminalSession2.TerminalAccessId)
		assert.Nil(t, err)
		err = terminalAccessServiceImpl.DisconnectTerminalSession(context.Background(), terminalSession3.TerminalAccessId)
		assert.Nil(t, err)

		_, err = fetchSessionId(terminalAccessServiceImpl, terminalSession1.TerminalAccessId)
		assert.Equal(t, err, errors.New("pod-terminated"))
		_, err = fetchSessionId(terminalAccessServiceImpl, terminalSession2.TerminalAccessId)
		assert.Equal(t, err, errors.New("pod-terminated"))
		_, err = fetchSessionId(terminalAccessServiceImpl, terminalSession3.TerminalAccessId)
		assert.Equal(t, err, errors.New("pod-terminated"))
	})

	t.Run("convert to k8s structure", func(t *testing.T) {
		//t.SkipNow()
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
	defaultAuthPolicyRepositoryImpl := repository3.NewDefaultAuthPolicyRepositoryImpl(db, sugaredLogger)
	defaultAuthRoleRepositoryImpl := repository3.NewDefaultAuthRoleRepositoryImpl(db, sugaredLogger)
	userAuthRepositoryImpl := repository3.NewUserAuthRepositoryImpl(db, sugaredLogger, defaultAuthPolicyRepositoryImpl, defaultAuthRoleRepositoryImpl)
	userRepositoryImpl := repository3.NewUserRepositoryImpl(db, sugaredLogger)
	roleGroupRepositoryImpl := repository3.NewRoleGroupRepositoryImpl(db, sugaredLogger)
	clusterServiceImpl := cluster.NewClusterServiceImpl(clusterRepositoryImpl, sugaredLogger, nil, k8sInformerFactoryImpl, userAuthRepositoryImpl, userRepositoryImpl, roleGroupRepositoryImpl)
	//k8sClientServiceImpl := application2.NewK8sClientServiceImpl(sugaredLogger, clusterServiceImpl, nil)
	//clusterServiceImpl := cluster2.NewClusterServiceImplExtended(clusterRepositoryImpl, nil, nil, sugaredLogger, nil, nil, nil, nil, nil)
	k8sResourceHistoryRepositoryImpl := repository10.NewK8sResourceHistoryRepositoryImpl(db, sugaredLogger)
	appRepositoryImpl := app.NewAppRepositoryImpl(db, sugaredLogger)
	environmentRepositoryImpl := repository2.NewEnvironmentRepositoryImpl(db, sugaredLogger, nil)
	k8sResourceHistoryServiceImpl := kubernetesResourceAuditLogs.Newk8sResourceHistoryServiceImpl(k8sResourceHistoryRepositoryImpl, sugaredLogger, appRepositoryImpl, environmentRepositoryImpl)
	//k8sApplicationService := application.NewK8sApplicationServiceImpl(sugaredLogger, clusterServiceImpl, nil, nil, nil, nil, k8sResourceHistoryServiceImpl, nil)
	K8sCommonService := k8s.NewK8sCommonServiceImpl(sugaredLogger, nil, nil, k8sResourceHistoryServiceImpl, clusterServiceImpl)
	terminalSessionHandlerImpl := terminal.NewTerminalSessionHandlerImpl(nil, clusterServiceImpl, sugaredLogger)
	userTerminalSessionConfig, err := GetTerminalAccessConfig()
	assert.Nil(t, err)
	userTerminalSessionConfig.TerminalPodStatusSyncTimeInSecs = 30
	userTerminalSessionConfig.TerminalPodInActiveDurationInMins = 1
	terminalAccessServiceImpl, err := NewUserTerminalAccessServiceImpl(sugaredLogger, terminalAccessRepositoryImpl, userTerminalSessionConfig, K8sCommonService, terminalSessionHandlerImpl, nil, nil)
	assert.Nil(t, err)
	return terminalAccessServiceImpl
}

func createAndUpdateSessionForUser(t *testing.T, terminalAccessServiceImpl *UserTerminalAccessServiceImpl, clusterId int, userId int32, baseImage string) *models.UserTerminalSessionResponse {

	request := &models.UserTerminalSessionRequest{
		UserId:    userId,
		ClusterId: clusterId,
		BaseImage: baseImage,
		ShellName: "sh",
		NodeName:  "demo-new",
		Namespace: "default",
	}
	time.Sleep(5 * time.Second)
	startTerminalSession, err := terminalAccessServiceImpl.StartTerminalSession(context.Background(), request)
	assert.Nil(t, err)
	fmt.Println(startTerminalSession)
	sessionId, err := fetchSessionId(terminalAccessServiceImpl, startTerminalSession.TerminalAccessId)
	assert.Nil(t, err)
	fmt.Println("SessionId: ", sessionId)

	shellChangeResponse, err := terminalAccessServiceImpl.UpdateTerminalShellSession(context.Background(), &models.UserTerminalShellSessionRequest{TerminalAccessId: startTerminalSession.TerminalAccessId, ShellName: "bash"})
	assert.Nil(t, err)
	assert.Equal(t, shellChangeResponse.TerminalAccessId, startTerminalSession.TerminalAccessId)
	request.BaseImage = "nginx:latest"
	request.Id = startTerminalSession.TerminalAccessId
	updateTerminalSession, err := terminalAccessServiceImpl.UpdateTerminalSession(context.Background(), request)
	assert.Nil(t, err)
	assert.Equal(t, updateTerminalSession.UserId, request.UserId)
	fmt.Println("updated: ", updateTerminalSession)
	sessionId, err = fetchSessionId(terminalAccessServiceImpl, updateTerminalSession.TerminalAccessId)
	assert.Nil(t, err)
	fmt.Println("SessionId: ", sessionId)

	terminalShellSession, err := terminalAccessServiceImpl.UpdateTerminalShellSession(context.Background(), &models.UserTerminalShellSessionRequest{TerminalAccessId: updateTerminalSession.TerminalAccessId, ShellName: "bash"})
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
		fetchTerminalStatus, err := terminalAccessServiceImpl.FetchTerminalStatus(context.Background(), terminalAccessId, "default", "internal-kubectl", "sh")
		if err != nil {
			return sessionId, err
		}
		sessionId = fetchTerminalStatus.UserTerminalSessionId
		fmt.Println("sessionId: ", sessionId)
		time.Sleep(5 * time.Second)
	}
	return sessionId, nil
}
