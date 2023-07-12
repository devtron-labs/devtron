package k8s

import (
	"context"
	"fmt"
	"github.com/devtron-labs/authenticator/client"
	"github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/devtron-labs/devtron/client/k8s/informer"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/terminal"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"strings"
	"testing"
	"time"
)

const testClusterId = 1
const testNamespace = "default"
const testPodName = "nginx-test"
const testContainer = "nginx"
const testImage = "nginx"
const testPodJs = `{"apiVersion": "v1","kind": "Pod","metadata": {"name": "nginx-test"},"spec": {"containers": [{"name": "nginx","image": "nginx","imagePullPolicy": "IfNotPresent"}]}}`

func TestGetPodContainersList(t *testing.T) {

	k8sApplicationService := initK8sApplicationService(t)

	t.Run("Create Ephemeral Container with valid Basic Data,container status will be running", func(tt *testing.T) {
		CreateAndDeletePod(tt, k8sApplicationService)
		ephemeralContainerName := "debugger-basic-test"
		req := cluster.EphemeralContainerRequest{
			ClusterId:    testClusterId,
			Namespace:    testNamespace,
			PodName:      testPodName,
			UserId:       1,
			AdvancedData: nil,
			BasicData: &cluster.EphemeralContainerBasicData{
				ContainerName:       ephemeralContainerName,
				TargetContainerName: testContainer,
				Image:               testImage,
			},
		}
		time.Sleep(5 * time.Second)
		err := k8sApplicationService.CreatePodEphemeralContainers(req)
		assert.Nil(tt, err)
		time.Sleep(5 * time.Second)
		list, err := k8sApplicationService.GetPodContainersList(testClusterId, testNamespace, testPodName)
		assert.Nil(tt, err)
		assert.NotNil(tt, list)
		assert.Equal(tt, 1, len(list.EphemeralContainers))
		assert.Equal(tt, true, strings.Contains(list.EphemeralContainers[0], ephemeralContainerName))
	})

	t.Run("Create Ephemeral Container with valid Advanced Data,container status will be running", func(tt *testing.T) {
		CreateAndDeletePod(tt, k8sApplicationService)
		manifest := `{"name":"debugger-advanced-test","command":["sh"],"image":"quay.io/devtron/ubuntu-k8s-utils:latest","targetContainer":"nginx","tty":true,"stdin":true}`
		ephemeralContainerName := "debugger-advanced-test"
		req := cluster.EphemeralContainerRequest{
			ClusterId: testClusterId,
			Namespace: testNamespace,
			PodName:   testPodName,
			UserId:    1,
			BasicData: nil,
			AdvancedData: &cluster.EphemeralContainerAdvancedData{
				Manifest: manifest,
			},
		}
		time.Sleep(5 * time.Second)
		err := k8sApplicationService.CreatePodEphemeralContainers(req)
		assert.Nil(tt, err)
		time.Sleep(5 * time.Second)
		list, err := k8sApplicationService.GetPodContainersList(testClusterId, testNamespace, testPodName)
		assert.Nil(tt, err)
		assert.NotNil(tt, list)
		assert.Equal(tt, 1, len(list.EphemeralContainers))
		assert.Equal(tt, true, strings.Contains(list.EphemeralContainers[0], ephemeralContainerName))
	})

	t.Run("Create Ephemeral Container with inValid Data, container status will be terminated", func(tt *testing.T) {
		CreateAndDeletePod(tt, k8sApplicationService)
		ephemeralContainerName := "debugger-basic-invalid-test"
		req := cluster.EphemeralContainerRequest{
			ClusterId:    testClusterId,
			Namespace:    testNamespace,
			PodName:      testPodName,
			UserId:       1,
			AdvancedData: nil,
			BasicData: &cluster.EphemeralContainerBasicData{
				ContainerName:       ephemeralContainerName,
				TargetContainerName: testContainer,
				Image:               "invalidImage",
			},
		}
		time.Sleep(5 * time.Second)
		err := k8sApplicationService.CreatePodEphemeralContainers(req)
		assert.Nil(tt, err)
		time.Sleep(5 * time.Second)
		list, err := k8sApplicationService.GetPodContainersList(testClusterId, testNamespace, testPodName)
		assert.Nil(tt, err)
		assert.NotNil(tt, list)
		assert.Equal(tt, 0, len(list.EphemeralContainers))
	})

	t.Run("Create Ephemeral Container with inValid Data, wrong pod name,error occurs with resource not found", func(tt *testing.T) {
		CreateAndDeletePod(tt, k8sApplicationService)
		ephemeralContainerName := "debugger-basic-invalid-test"
		req := cluster.EphemeralContainerRequest{
			ClusterId:    testClusterId,
			Namespace:    testNamespace,
			PodName:      "invalidPodName",
			UserId:       1,
			AdvancedData: nil,
			BasicData: &cluster.EphemeralContainerBasicData{
				ContainerName:       ephemeralContainerName,
				TargetContainerName: testContainer,
				Image:               testImage,
			},
		}
		time.Sleep(5 * time.Second)
		err := k8sApplicationService.CreatePodEphemeralContainers(req)
		assert.NotNil(tt, err)
	})

	t.Run("Create Ephemeral Container with inValid Data, container status will be terminated", func(t *testing.T) {

	})

	t.Run("Terminate Ephemeral Container with valid Data, container status will be terminated", func(t *testing.T) {

	})

	t.Run("Terminate Ephemeral Container with InValid Data,Invalid Pod Name payload, resource not found error", func(t *testing.T) {

	})

}

func deleteTestPod(k8sApplicationService *K8sApplicationServiceImpl) error {
	restConfig, k8sRequest, err := getRestConfigAndK8sRequestObj(k8sApplicationService)
	k8sRequest.ResourceIdentifier.Name = testPodName
	if err != nil {
		return err
	}
	_, err = k8sApplicationService.k8sClientService.DeleteResource(context.Background(), restConfig, k8sRequest)
	return err
}
func createTestPod(k8sApplicationService *K8sApplicationServiceImpl) error {
	restConfig, k8sRequest, err := getRestConfigAndK8sRequestObj(k8sApplicationService)
	if err != nil {
		return err
	}
	_, err = k8sApplicationService.k8sClientService.CreateResource(context.Background(), restConfig, k8sRequest, testPodJs)
	return err
}

func getRestConfigAndK8sRequestObj(k8sApplicationService *K8sApplicationServiceImpl) (*rest.Config, *application.K8sRequestBean, error) {
	restConfig, err := k8sApplicationService.GetRestConfigByClusterId(context.Background(), testClusterId)
	if err != nil {
		return nil, nil, err
	}
	_, groupVersionKind, err := legacyscheme.Codecs.UniversalDeserializer().Decode([]byte(testPodJs), nil, nil)
	if err != nil {
		return restConfig, nil, err
	}

	k8sRequest := &application.K8sRequestBean{
		ResourceIdentifier: application.ResourceIdentifier{
			Namespace: testNamespace,
			GroupVersionKind: schema.GroupVersionKind{
				Group:   groupVersionKind.Group,
				Version: groupVersionKind.Version,
				Kind:    groupVersionKind.Kind,
			},
		},
	}
	return restConfig, k8sRequest, nil
}
func initK8sApplicationService(t *testing.T) *K8sApplicationServiceImpl {
	sugaredLogger, _ := util.InitLogger()
	config, _ := sql.GetConfig()
	runtimeConfig, err := client.GetRuntimeConfig()
	k8sUtil := util.NewK8sUtil(sugaredLogger, runtimeConfig)
	assert.Nil(t, err)
	db, _ := sql.NewDbConnection(config, sugaredLogger)
	clusterRepositoryImpl := repository.NewClusterRepositoryImpl(db, sugaredLogger)
	k8sClientServiceImpl := application.NewK8sClientServiceImpl(sugaredLogger, clusterRepositoryImpl)
	v := informer.NewGlobalMapClusterNamespace()
	k8sInformerFactoryImpl := informer.NewK8sInformerFactoryImpl(sugaredLogger, v, runtimeConfig)
	clusterServiceImpl := cluster.NewClusterServiceImpl(clusterRepositoryImpl, sugaredLogger, nil, k8sInformerFactoryImpl, nil, nil, nil)
	terminalSessionHandlerImpl := terminal.NewTerminalSessionHandlerImpl(nil, clusterServiceImpl, sugaredLogger, k8sUtil)
	k8sApplicationService := NewK8sApplicationServiceImpl(sugaredLogger, clusterServiceImpl, nil, k8sClientServiceImpl, nil, k8sUtil, nil, nil, terminalSessionHandlerImpl, nil)
	return k8sApplicationService
}

func CreateAndDeletePod(t *testing.T, k8sApplicationService *K8sApplicationServiceImpl) {
	err := createTestPod(k8sApplicationService)
	if err != nil {
		assert.Fail(t, "error in creating test Pod ")
	}
	t.Cleanup(func() {
		fmt.Println("cleaning data ....")
		err = deleteTestPod(k8sApplicationService)
		if err != nil {
			assert.Fail(t, "error in deleting test Pod after running testcases")
		}
	})
}
