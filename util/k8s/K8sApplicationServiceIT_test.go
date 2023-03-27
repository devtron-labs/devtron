package k8s

import (
	"context"
	"fmt"
	client2 "github.com/devtron-labs/authenticator/client"
	"github.com/devtron-labs/devtron/api/connector"
	mocks1 "github.com/devtron-labs/devtron/api/helm-app/mocks"
	"github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/cluster/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"testing"
)

const default_namespace = "default"

func getK8sApplicationService(t *testing.T) *K8sApplicationServiceImpl {
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
	return NewK8sApplicationServiceImpl(logger, &cluster.ClusterServiceImpl{}, *pump, *k8sClient, helmAppService, k8sUtil, acdAuthConfig, kubernetesResourceAuditLogs.K8sResourceHistoryServiceImpl{})
}

func TestGetPodLogs(t *testing.T) {
	k8sApplicationService := getK8sApplicationService(t)
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
	testManifestYaml := `{"apiVersion": "v1","kind": "Pod","metadata": null,"name": "%s","labels": null,"env": "test","spec": null,"containers": [{"name": "%s"}],"image": "nginx","imagePullPolicy": "IfNotPresent"}`

	containerName := "nginx"
	podName := "nginx"

	testManifestYaml = fmt.Sprintf(testManifestYaml, podName, containerName)
	err = yaml.Unmarshal([]byte(testManifestYaml), &testManifest)
	if err != nil {
		t.Fail()
		return
	}
	success, err := k8sApplicationService.applyResourceFromManifest(context.Background(), testManifest, restConfig, default_namespace)
	assert.Equal(t, true, success)
	if err != nil || !success {
		t.Fail()
		return
	}

	t.Run("", func(tt *testing.T) {
		request := &ResourceRequestBean{}
		logs, err := k8sApplicationService.GetPodLogs(context.Background(), request)
		assert.Nil(tt, err)
		assert.NotNil(tt, logs)
		err = logs.Close()
		assert.Nil(tt, err)
	})

	t.Run("", func(tt *testing.T) {
		request := &ResourceRequestBean{}
		logs, err := k8sApplicationService.GetPodLogs(context.Background(), request)
		assert.Nil(tt, err)
		assert.NotNil(tt, logs)
		err = logs.Close()
		assert.Nil(tt, err)
	})

	t.Run("", func(tt *testing.T) {
		request := &ResourceRequestBean{}
		logs, err := k8sApplicationService.GetPodLogs(context.Background(), request)
		assert.Nil(tt, err)
		assert.NotNil(tt, logs)
		err = logs.Close()
		assert.Nil(tt, err)
	})

	t.Run("", func(tt *testing.T) {
		request := &ResourceRequestBean{}
		logs, err := k8sApplicationService.GetPodLogs(context.Background(), request)
		assert.Nil(tt, err)
		assert.NotNil(tt, logs)
		err = logs.Close()
		assert.Nil(tt, err)
	})

	t.Run("", func(tt *testing.T) {
		request := &ResourceRequestBean{}
		logs, err := k8sApplicationService.GetPodLogs(context.Background(), request)
		assert.Nil(tt, err)
		assert.NotNil(tt, logs)
		err = logs.Close()
		assert.Nil(tt, err)
	})
}
