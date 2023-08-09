package application

import (
	"context"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/devtron-labs/authenticator/client"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	s "github.com/devtron-labs/devtron/pkg/k8s"
	informer2 "github.com/devtron-labs/devtron/pkg/k8s/informer"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/terminal"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/k8s"
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
const testImage = "alpine:latest"
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

		//create ephemeral container
		err := k8sApplicationService.CreatePodEphemeralContainers(&req)
		testCreationSuccess(err, podName, req.BasicData.ContainerName, 1, k8sApplicationService, tt)

		//delete ephemeral container
		terminated, err := k8sApplicationService.TerminatePodEphemeralContainer(req)
		assert.Nil(tt, err)
		assert.Equal(tt, true, terminated)

		//fetch container list for the pod and check if the ephemeral container is terminated
		list, err := k8sApplicationService.GetPodContainersList(testClusterId, testNamespace, podName)
		assert.Nil(tt, err)
		assert.NotNil(tt, list)
		assert.Equal(tt, 0, len(list.EphemeralContainers))
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
	//restConfig, k8sRequest, err := getRestConfigAndK8sRequestObj()
	//k8sRequest.ResourceIdentifier.Name = podName
	//if err != nil {
	//	return err
	//}
	//_, err = k8sApplicationService.k8sClientService.DeleteResource(context.Background(), restConfig, k8sRequest)
	//return err
	return nil
}
func createTestPod(podName string, k8sApplicationService *K8sApplicationServiceImpl) error {
	//restConfig, k8sRequest, err := getRestConfigAndK8sRequestObj(nil)
	//if err != nil {
	//	return err
	//}
	//testPodJs1 := fmt.Sprintf(testPodJs, podName)
	//_, err = k8sApplicationService.k8sClientService.CreateResource(context.Background(), restConfig, k8sRequest, testPodJs1)
	//return err
	return nil
}

func getRestConfigAndK8sRequestObj(k8sCommonService s.K8sCommonService) (*rest.Config, *k8s.K8sRequestBean, error) {
	restConfig, err, _ := k8sCommonService.GetRestConfigByClusterId(context.Background(), testClusterId)
	if err != nil {
		return nil, nil, err
	}
	_, groupVersionKind, err := legacyscheme.Codecs.UniversalDeserializer().Decode([]byte(testPodJs), nil, nil)
	if err != nil {
		return restConfig, nil, err
	}

	k8sRequest := &k8s.K8sRequestBean{
		ResourceIdentifier: k8s.ResourceIdentifier{
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
	k8sUtil := k8s.NewK8sUtil(sugaredLogger, runtimeConfig)
	assert.Nil(t, err)
	db, _ := sql.NewDbConnection(config, sugaredLogger)
	ephemeralContainerRepository := repository.NewEphemeralContainersRepositoryImpl(db)
	clusterRepositoryImpl := repository.NewClusterRepositoryImpl(db, sugaredLogger)
	//Client Service has been removed. Please use application service or common service
	//k8sClientServiceImpl := application.NewK8sClientServiceImpl(sugaredLogger, clusterRepositoryImpl)
	v := informer2.NewGlobalMapClusterNamespace()
	k8sInformerFactoryImpl := informer2.NewK8sInformerFactoryImpl(sugaredLogger, v, runtimeConfig, k8sUtil)
	clusterServiceImpl := cluster.NewClusterServiceImpl(clusterRepositoryImpl, sugaredLogger, k8sUtil, k8sInformerFactoryImpl, nil, nil, nil)
	ephemeralContainerService := cluster.NewEphemeralContainerServiceImpl(ephemeralContainerRepository, sugaredLogger)
	terminalSessionHandlerImpl := terminal.NewTerminalSessionHandlerImpl(nil, clusterServiceImpl, sugaredLogger, k8sUtil, ephemeralContainerService)
	k8sApplicationService, _ := NewK8sApplicationServiceImpl(sugaredLogger, clusterServiceImpl, nil, nil, k8sUtil, nil, nil, nil, terminalSessionHandlerImpl, ephemeralContainerService, ephemeralContainerRepository)
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

func TestMatchRegex(t *testing.T) {
	cfg := &EphemeralContainerConfig{}
	env.Parse(cfg)
	ephemeralRegex := cfg.EphemeralServerVersionRegex
	type args struct {
		exp  string
		text string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Invalid regex",
			args: args{
				exp:  "**",
				text: "v1.23+",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Valid regex,text not matching with regex",
			args: args{
				exp:  ephemeralRegex,
				text: "v1.03+",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Valid regex,text not matching with regex",
			args: args{
				exp:  ephemeralRegex,
				text: "v1.22+",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Valid regex, text not matching with regex",
			args: args{
				exp:  ephemeralRegex,
				text: "v1.3",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Valid regex, text match with regex",
			args: args{
				exp:  ephemeralRegex,
				text: "v1.23+",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Valid regex, text match with regex",
			args: args{
				exp:  ephemeralRegex,
				text: "v1.26.6",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Valid regex, text match with regex",
			args: args{
				exp:  ephemeralRegex,
				text: "v1.26",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Valid regex, text match with regex",
			args: args{
				exp:  ephemeralRegex,
				text: "v1.30",
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := util2.MatchRegexExpression(tt.args.exp, tt.args.text)
			fmt.Println(err)
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchRegexExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MatchRegexExpression() got = %v, want %v", got, tt.want)
			}
		})
	}
}
