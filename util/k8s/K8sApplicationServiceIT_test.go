package k8s

import (
	"context"
	"fmt"
	client2 "github.com/devtron-labs/authenticator/client"
	"github.com/devtron-labs/devtron/api/connector"
	client "github.com/devtron-labs/devtron/api/helm-app"
	mocks1 "github.com/devtron-labs/devtron/api/helm-app/mocks"
	"github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	mocks2 "github.com/devtron-labs/devtron/pkg/cluster/mocks"
	"github.com/devtron-labs/devtron/pkg/cluster/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
	"time"
)

func getK8sApplicationService(clusterServiceMocked cluster.ClusterService, t *testing.T) *K8sApplicationServiceImpl {
	logger, err := util.NewSugardLogger()
	if err != nil {
		return nil
	}
	pump := connector.NewPumpImpl(logger)
	mockedClusterRepository := mocks.NewClusterRepository(t)
	k8sClient := application.NewK8sClientServiceImpl(logger, mockedClusterRepository)
	helmAppService := mocks1.NewHelmAppService(t)
	runtimeConfig, err := client2.GetRuntimeConfig()
	runtimeConfig.LocalDevMode = true
	if err != nil {
		return nil
	}
	acdAuthConfig, err := util2.GetACDAuthConfig()
	if err != nil {
		return nil
	}
	k8sUtil := util.NewK8sUtil(logger, runtimeConfig)
	return NewK8sApplicationServiceImpl(logger, clusterServiceMocked, *pump, *k8sClient, helmAppService, k8sUtil, acdAuthConfig, kubernetesResourceAuditLogs.K8sResourceHistoryServiceImpl{})
}

func TestGetPodLogs(t *testing.T) {

	testContainerName := "nginx"
	testPodName := "nginx"
	testClusterId := 1

	clusterServiceMocked := mocks2.NewClusterService(t)
	k8sApplicationService := getK8sApplicationService(clusterServiceMocked, t)
	if k8sApplicationService == nil {
		t.Fail()
		return
	}
	restConfig, err := k8sApplicationService.K8sUtil.GetK8sClusterRestConfig()
	if err != nil {
		t.Fail()
		return
	}
	var testManifest unstructured.Unstructured
	testManifestYaml := `{"apiVersion": "v1","kind": "Pod","metadata": {"name": "%s","labels": {"env": "test"}},"spec": {"containers": [{"name": "%s","image": "nginx","imagePullPolicy": "IfNotPresent"}]}}`
	testManifestYaml = fmt.Sprintf(testManifestYaml, testPodName, testContainerName)
	manifestMap := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(testManifestYaml), &manifestMap)
	testManifest.SetUnstructuredContent(manifestMap)
	if err != nil {
		t.Fail()
		return
	}
	success, err := k8sApplicationService.applyResourceFromManifest(context.Background(), testManifest, restConfig, DEFAULT_NAMESPACE)
	assert.Equal(t, true, success)
	if err != nil || !success {
		if err != nil {
			k8sApplicationService.logger.Errorw("err : ", err.Error())
		}
		t.Fail()
		return
	}

	clusterServiceMocked.On("FindById", testClusterId).Return(cluster.ClusterBean{
		ClusterName: DEFAULT_CLUSTER,
		Config:      make(map[string]string),
	})

	request := &ResourceRequestBean{
		AppIdentifier: &client.AppIdentifier{
			ClusterId: testClusterId,
			Namespace: DEFAULT_NAMESPACE,
		},
		K8sRequest: &application.K8sRequestBean{
			ResourceIdentifier: application.ResourceIdentifier{
				Name:      testPodName,
				Namespace: DEFAULT_NAMESPACE,
				GroupVersionKind: schema.GroupVersionKind{
					Version: "v1",
					Kind:    "Pod",
				},
			},
		},
	}

	t.Run("", func(tt *testing.T) {
		var tailLine int64 = 2
		request.K8sRequest.PodLogsRequest = application.PodLogsRequest{
			ContainerName:     testContainerName,
			PreviousContainer: false,
			TailLines:         &tailLine,
			Follow:            true,
		}
		logs, err := k8sApplicationService.GetPodLogs(context.Background(), request)
		assert.Nil(tt, err)
		assert.NotNil(tt, logs)
		err = logs.Close()
		assert.Nil(tt, err)
	})

	t.Run("", func(tt *testing.T) {
		var tailLine int64 = 2
		request.K8sRequest.PodLogsRequest = application.PodLogsRequest{
			ContainerName:     testContainerName,
			PreviousContainer: false,
			TailLines:         &tailLine,
			Follow:            false,
		}
		logs, err := k8sApplicationService.GetPodLogs(context.Background(), request)
		assert.Nil(tt, err)
		assert.NotNil(tt, logs)
		err = logs.Close()
		assert.Nil(tt, err)
	})

	t.Run("", func(tt *testing.T) {
		var tailLine int64 = 2
		request.K8sRequest.PodLogsRequest = application.PodLogsRequest{
			ContainerName:     testContainerName,
			PreviousContainer: true,
			TailLines:         &tailLine,
			Follow:            false,
		}
		logs, err := k8sApplicationService.GetPodLogs(context.Background(), request)
		assert.Nil(tt, err)
		assert.NotNil(tt, logs)
		err = logs.Close()
		assert.Nil(tt, err)
	})

	t.Run("", func(tt *testing.T) {
		var sinceSeconds int64 = 100
		request.K8sRequest.PodLogsRequest = application.PodLogsRequest{
			ContainerName:     testContainerName,
			PreviousContainer: true,
			TailLines:         nil,
			Follow:            false,
			SinceTime:         nil,
			SinceSeconds:      &sinceSeconds,
		}
		logs, err := k8sApplicationService.GetPodLogs(context.Background(), request)
		assert.Nil(tt, err)
		assert.NotNil(tt, logs)
		err = logs.Close()
		assert.Nil(tt, err)
	})

	t.Run("", func(tt *testing.T) {
		var sinceSeconds int64 = 100
		timeNow := time.Now()
		sinceTime := metav1.Time{Time: timeNow.Add(time.Duration(-1 * sinceSeconds))}
		request.K8sRequest.PodLogsRequest = application.PodLogsRequest{
			ContainerName:     testContainerName,
			PreviousContainer: true,
			TailLines:         nil,
			Follow:            false,
			SinceTime:         &sinceTime,
			SinceSeconds:      &sinceSeconds,
		}
		logs, err := k8sApplicationService.GetPodLogs(context.Background(), request)
		assert.Nil(tt, err)
		assert.NotNil(tt, logs)
		err = logs.Close()
		assert.Nil(tt, err)
	})

	t.Run("", func(tt *testing.T) {
		var tailLine int64 = 1
		var sinceSeconds int64 = 3600
		timeNow := time.Now()
		sinceTime := metav1.Time{Time: timeNow.Add(time.Duration(-1 * sinceSeconds))}
		request.K8sRequest.PodLogsRequest = application.PodLogsRequest{
			ContainerName:     testContainerName,
			PreviousContainer: true,
			TailLines:         &tailLine,
			Follow:            false,
			SinceTime:         &sinceTime,
			SinceSeconds:      &sinceSeconds,
		}
		//should only get 1 line in logs since tailLines is set
		logs, err := k8sApplicationService.GetPodLogs(context.Background(), request)
		assert.Nil(tt, err)
		assert.NotNil(tt, logs)
		err = logs.Close()
		assert.Nil(tt, err)
	})

	t.Run("", func(tt *testing.T) {
		request.K8sRequest.PodLogsRequest = application.PodLogsRequest{
			ContainerName:     testContainerName,
			PreviousContainer: true,
			TailLines:         nil,
			Follow:            false,
			SinceTime:         nil,
			SinceSeconds:      nil,
		}
		//should get all the logs since the creation of container
		logs, err := k8sApplicationService.GetPodLogs(context.Background(), request)
		assert.Nil(tt, err)
		assert.NotNil(tt, logs)
		err = logs.Close()
		assert.Nil(tt, err)
	})
}
