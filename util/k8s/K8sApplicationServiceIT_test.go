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
	errors2 "k8s.io/apimachinery/pkg/api/errors"
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
const testImage = "alpine"
const testImage2 = "quay.io/devtron/ubuntu-k8s-utils:latest"
const testPodJs = `{"apiVersion": "v1","kind": "Pod","metadata": {"name": "%s"},"spec": {"containers": [{"name": "nginx","image": "nginx","imagePullPolicy": "IfNotPresent"}]}}`
const testAdvancedManifest = `{"name":"%s","command":["sh"],"image":"%s","targetContainerName":"nginx","tty":true,"stdin":true}`

func TestEphemeralContainers(t *testing.T) {

	k8sApplicationService := initK8sApplicationService(t)

	t.Run("Create Ephemeral Container with valid Basic Data,container status will be running", func(tt *testing.T) {
		podName := testPodName + "-1"
		CreateAndDeletePod(podName, tt, k8sApplicationService)
		ephemeralContainerName := "debugger-basic-test"
		req := cluster.EphemeralContainerRequest{
			ClusterId:    testClusterId,
			Namespace:    testNamespace,
			PodName:      podName,
			UserId:       1,
			AdvancedData: nil,
			BasicData: &cluster.EphemeralContainerBasicData{
				ContainerName:       ephemeralContainerName,
				TargetContainerName: testContainer,
				Image:               testImage,
			},
		}
		time.Sleep(5 * time.Second)
		err := k8sApplicationService.CreatePodEphemeralContainers(&req)
		testCreationSuccess(err, podName, ephemeralContainerName, 1, k8sApplicationService, tt)
	})

	t.Run("Create Ephemeral Container with valid Advanced Data,container status will be running", func(tt *testing.T) {
		podName := "debugger-advanced-test" + "-2"
		CreateAndDeletePod(podName, tt, k8sApplicationService)
		ephemeralContainerName := "debugger-advanced-test"
		manifest := fmt.Sprintf(testAdvancedManifest, ephemeralContainerName, testImage2)

		req := cluster.EphemeralContainerRequest{
			ClusterId: testClusterId,
			Namespace: testNamespace,
			PodName:   podName,
			UserId:    1,
			BasicData: nil,
			AdvancedData: &cluster.EphemeralContainerAdvancedData{
				Manifest: manifest,
			},
		}
		time.Sleep(5 * time.Second)
		err := k8sApplicationService.CreatePodEphemeralContainers(&req)
		testCreationSuccess(err, podName, req.BasicData.ContainerName, 1, k8sApplicationService, tt)
	})

	t.Run("Create Ephemeral Container with inValid Data, container status will be terminated", func(tt *testing.T) {
		podName := testPodName + "-3"
		CreateAndDeletePod(podName, tt, k8sApplicationService)
		ephemeralContainerName := "debugger-basic-invalid-test"
		req := cluster.EphemeralContainerRequest{
			ClusterId:    testClusterId,
			Namespace:    testNamespace,
			PodName:      podName,
			UserId:       1,
			AdvancedData: nil,
			BasicData: &cluster.EphemeralContainerBasicData{
				ContainerName:       ephemeralContainerName,
				TargetContainerName: testContainer,
				Image:               "invalidImage",
			},
		}
		time.Sleep(5 * time.Second)
		err := k8sApplicationService.CreatePodEphemeralContainers(&req)
		testCreationSuccess(err, podName, req.BasicData.ContainerName, 0, k8sApplicationService, tt)
	})

	t.Run("Create Ephemeral Container with inValid Data, wrong pod name,error occurs with resource not found", func(tt *testing.T) {
		podName := testPodName + "-4"
		CreateAndDeletePod(podName, tt, k8sApplicationService)
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
		err := k8sApplicationService.CreatePodEphemeralContainers(&req)
		assert.NotNil(tt, err)
		assert.Equal(tt, true, errors2.IsNotFound(err))
	})

	t.Run("Create Ephemeral Container with advanced inValid Data, container status will be terminated", func(tt *testing.T) {
		podName := "debugger-advanced-test" + "-5"
		CreateAndDeletePod(podName, tt, k8sApplicationService)
		ephemeralContainerName := "debugger-advanced-invalid-test"
		manifest := fmt.Sprintf(testAdvancedManifest, ephemeralContainerName, "InvalidImage")
		req := cluster.EphemeralContainerRequest{
			ClusterId: testClusterId,
			Namespace: testNamespace,
			PodName:   podName,
			UserId:    1,
			BasicData: nil,
			AdvancedData: &cluster.EphemeralContainerAdvancedData{
				Manifest: manifest,
			},
		}
		time.Sleep(5 * time.Second)
		err := k8sApplicationService.CreatePodEphemeralContainers(&req)
		testCreationSuccess(err, podName, req.BasicData.ContainerName, 0, k8sApplicationService, tt)
	})

	t.Run("Create Ephemeral Container with advanced inValid Data, manifest with unsupported fields, container creation throws error", func(tt *testing.T) {
		podName := "debugger-advanced-test" + "-5"
		ephemeralContainerName := "debugger-advanced-invalid-test"
		CreateAndDeletePod(podName, tt, k8sApplicationService)
		manifest := fmt.Sprintf(`{"name":"%s","command":["sh"],"image":"%s","targetContainerName":"nginx","tty":true,"stdin":true, "lifecycle": {
          "preStop": {
            "exec": {
              "command": [
                "/bin/sh",
                "-c",
                " curl -X POST -H \"Content-Type: application/json\" -d '{\"eventType\": \"SIG_TERM\"}' localhost:8080/orchestrator/telemetry/summary"
              ]
            }
          }
        }}`, ephemeralContainerName, "InvalidImage")
		req := cluster.EphemeralContainerRequest{
			ClusterId: testClusterId,
			Namespace: testNamespace,
			PodName:   podName,
			UserId:    1,
			BasicData: nil,
			AdvancedData: &cluster.EphemeralContainerAdvancedData{
				Manifest: manifest,
			},
		}
		time.Sleep(5 * time.Second)
		err := k8sApplicationService.CreatePodEphemeralContainers(&req)
		assert.NotNil(tt, err)
		assert.Equal(tt, true, strings.Contains(err.Error(), fmt.Sprintf("Pod \"%s\" is invalid", podName)))
	})

	t.Run("Terminate Ephemeral Container with valid Data, container status will be terminated", func(tt *testing.T) {
		podName := testPodName + "-6"
		CreateAndDeletePod(podName, tt, k8sApplicationService)
		ephemeralContainerName := "debugger-termination-valid-payload-test"
		req := cluster.EphemeralContainerRequest{
			ClusterId:    testClusterId,
			Namespace:    testNamespace,
			PodName:      podName,
			UserId:       1,
			AdvancedData: nil,
			BasicData: &cluster.EphemeralContainerBasicData{
				ContainerName:       ephemeralContainerName,
				TargetContainerName: testContainer,
				Image:               testImage,
			},
		}
		time.Sleep(5 * time.Second)

		//create ephemeral container 1
		err := k8sApplicationService.CreatePodEphemeralContainers(&req)
		testCreationSuccess(err, podName, req.BasicData.ContainerName, 1, k8sApplicationService, tt)
		firstCreatedEphemeralContainerName := req.BasicData.ContainerName

		//create ephemeral container 2
		req.BasicData.TargetContainerName = ephemeralContainerName
		req.AdvancedData = nil
		err = k8sApplicationService.CreatePodEphemeralContainers(&req)
		testCreationSuccess(err, podName, req.BasicData.ContainerName, 2, k8sApplicationService, tt)
		//delete ephemeral container
		terminated, err := k8sApplicationService.TerminatePodEphemeralContainer(req)
		assert.Nil(tt, err)
		assert.Equal(tt, true, terminated)

		//fetch container list for the pod and check if the ephemeral container is terminated
		list, err := k8sApplicationService.GetPodContainersList(testClusterId, testNamespace, podName)
		assert.Nil(tt, err)
		assert.NotNil(tt, list)
		assert.Equal(tt, 1, len(list.EphemeralContainers))
		assert.NotEqual(tt, req.BasicData.ContainerName, list.EphemeralContainers[0])
		assert.Equal(tt, firstCreatedEphemeralContainerName, list.EphemeralContainers[0])
	})

	t.Run("Terminate Ephemeral Container with InValid Data,Invalid Pod Name payload, resource not found error", func(tt *testing.T) {
		podName := testPodName + "-7"
		CreateAndDeletePod(podName, tt, k8sApplicationService)
		ephemeralContainerName := "debugger-termination-invalid-payload-test"
		req := cluster.EphemeralContainerRequest{
			ClusterId:    testClusterId,
			Namespace:    testNamespace,
			PodName:      podName,
			UserId:       1,
			AdvancedData: nil,
			BasicData: &cluster.EphemeralContainerBasicData{
				ContainerName:       ephemeralContainerName,
				TargetContainerName: testContainer,
				Image:               testImage,
			},
		}
		time.Sleep(5 * time.Second)
		//create ephemeral container
		err := k8sApplicationService.CreatePodEphemeralContainers(&req)
		testCreationSuccess(err, podName, ephemeralContainerName, 1, k8sApplicationService, tt)

		//delete ephemeral container
		req.PodName = "InvalidPodName"
		terminated, err := k8sApplicationService.TerminatePodEphemeralContainer(req)
		assert.NotNil(tt, err)
		assert.Equal(tt, true, errors2.IsNotFound(err))
		assert.Equal(tt, false, terminated)
	})

}

func testCreationSuccess(err error, podName, ephemeralContainerName string, listLen int, k8sApplicationService *K8sApplicationServiceImpl, tt *testing.T) {
	assert.Nil(tt, err)
	time.Sleep(2 * time.Second)
	list, err := k8sApplicationService.GetPodContainersList(testClusterId, testNamespace, podName)
	assert.Nil(tt, err)
	assert.NotNil(tt, list)
	assert.Equal(tt, listLen, len(list.EphemeralContainers))
	if listLen > 0 {
		assert.Equal(tt, true, strings.Contains(list.EphemeralContainers[0], ephemeralContainerName))
	}
}

func deleteTestPod(podName string, k8sApplicationService *K8sApplicationServiceImpl) error {
	restConfig, k8sRequest, err := getRestConfigAndK8sRequestObj(k8sApplicationService)
	k8sRequest.ResourceIdentifier.Name = podName
	if err != nil {
		return err
	}
	_, err = k8sApplicationService.k8sClientService.DeleteResource(context.Background(), restConfig, k8sRequest)
	return err
}
func createTestPod(podName string, k8sApplicationService *K8sApplicationServiceImpl) error {
	restConfig, k8sRequest, err := getRestConfigAndK8sRequestObj(k8sApplicationService)
	if err != nil {
		return err
	}
	testPodJs1 := fmt.Sprintf(testPodJs, podName)
	_, err = k8sApplicationService.k8sClientService.CreateResource(context.Background(), restConfig, k8sRequest, testPodJs1)
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
	ephemeralContainerRepository := repository.NewEphemeralContainersRepositoryImpl(db)
	clusterRepositoryImpl := repository.NewClusterRepositoryImpl(db, sugaredLogger)
	k8sClientServiceImpl := application.NewK8sClientServiceImpl(sugaredLogger, clusterRepositoryImpl)
	v := informer.NewGlobalMapClusterNamespace()
	k8sInformerFactoryImpl := informer.NewK8sInformerFactoryImpl(sugaredLogger, v, runtimeConfig)
	clusterServiceImpl := cluster.NewClusterServiceImpl(clusterRepositoryImpl, sugaredLogger, nil, k8sInformerFactoryImpl, nil, nil, nil)
	ephemeralContainerService := cluster.NewEphemeralContainerServiceImpl(ephemeralContainerRepository, sugaredLogger)
	terminalSessionHandlerImpl := terminal.NewTerminalSessionHandlerImpl(nil, clusterServiceImpl, sugaredLogger, k8sUtil, ephemeralContainerService)
	k8sApplicationService := NewK8sApplicationServiceImpl(sugaredLogger, clusterServiceImpl, nil, k8sClientServiceImpl, nil, k8sUtil, nil, nil, terminalSessionHandlerImpl, ephemeralContainerService)
	return k8sApplicationService
}

func CreateAndDeletePod(podName string, t *testing.T, k8sApplicationService *K8sApplicationServiceImpl) {
	time.Sleep(5 * time.Second)
	err := createTestPod(podName, k8sApplicationService)
	if err != nil {
		assert.Fail(t, "error in creating test Pod ")
	}
	t.Cleanup(func() {
		fmt.Println("cleaning data ....")
		err = deleteTestPod(podName, k8sApplicationService)
		if err != nil {
			assert.Fail(t, "error in deleting test Pod after running testcases")
		}
		fmt.Println("data cleaned!")
	})
}
