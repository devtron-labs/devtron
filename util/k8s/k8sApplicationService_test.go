package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/mock"
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"math/rand"
	"testing"
)

var manifest = `{
		"kind":"Service",
       "metadata":{
           "name":"test-service"
       },
		"spec": {
			"ingressClassName": "nginx",
			"rules": [
				{
					"host": "demo1.devtron.info",
					"http": {
						"paths": [
							{
								"backend": {
									"service": {
										"name": "devtron-service",
										"port": {
										"number": 80
										}
									}
								},
								"path": "/orchestrator",
								"pathType": "ImplementationSpecific"
							}
						]
					}
				},
				{
					"host": "demo1.devtron.info",
					"http": {
						"paths": [
							{
								"backend": {
									"service": {
										"name": "devtron-service",
										"port": {
										"number": 80
										}
									}
								},
								"path": "/dashboard",
								"pathType": "ImplementationSpecific"
							}
						]
					}
				}
			]
		},
		"status":{
			"loadBalancer":{
				"ingress":[
					{
						"hostname":"aws.ebs.23456",
						"ip":"111.222.333.444"
					}
				]
			}
		}
	  }`

type NewK8sClientServiceImplMock struct {
	mock.Mock
}
type NewClusterServiceMock struct {
	mock.Mock
}

func (n NewClusterServiceMock) Save(parent context.Context, bean *cluster.ClusterBean, userId int32) (*cluster.ClusterBean, error) {
	//TODO implement me
	panic("implement me")
}

func (n NewClusterServiceMock) FindOne(clusterName string) (*cluster.ClusterBean, error) {
	//TODO implement me
	panic("implement me")
}

func (n NewClusterServiceMock) FindOneActive(clusterName string) (*cluster.ClusterBean, error) {
	//TODO implement me
	panic("implement me")
}

func (n NewClusterServiceMock) FindAll() ([]*cluster.ClusterBean, error) {
	//TODO implement me
	panic("implement me")
}

func (n NewClusterServiceMock) FindAllActive() ([]cluster.ClusterBean, error) {
	//TODO implement me
	panic("implement me")
}

func (n NewClusterServiceMock) DeleteFromDb(bean *cluster.ClusterBean, userId int32) error {
	//TODO implement me
	panic("implement me")
}

func (n NewClusterServiceMock) FindById(id int) (*cluster.ClusterBean, error) {
	//TODO implement me
	return &cluster.ClusterBean{}, nil
}

func (n NewClusterServiceMock) FindByIds(id []int) ([]cluster.ClusterBean, error) {
	//TODO implement me
	panic("implement me")
}

func (n NewClusterServiceMock) Update(ctx context.Context, bean *cluster.ClusterBean, userId int32) (*cluster.ClusterBean, error) {
	//TODO implement me
	panic("implement me")
}

func (n NewClusterServiceMock) Delete(bean *cluster.ClusterBean, userId int32) error {
	//TODO implement me
	panic("implement me")
}

func (n NewClusterServiceMock) FindAllForAutoComplete() ([]cluster.ClusterBean, error) {
	//TODO implement me
	panic("implement me")
}

func (n NewClusterServiceMock) CreateGrafanaDataSource(clusterBean *cluster.ClusterBean, env *repository.Environment) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (n NewClusterServiceMock) GetClusterConfig(cluster *cluster.ClusterBean) (*util.ClusterConfig, error) {
	//TODO implement me
	panic("implement me")
}

func (n NewClusterServiceMock) GetK8sClient() (*v1.CoreV1Client, error) {
	//TODO implement me
	panic("implement me")
}

func (n NewK8sClientServiceImplMock) GetResource(restConfig *rest.Config, request *application.K8sRequestBean) (resp *application.ManifestResponse, err error) {
	kind := request.ResourceIdentifier.GroupVersionKind.Kind
	man := generateTestManifest(kind)
	return &man, nil
}

func (n NewK8sClientServiceImplMock) CreateResource(restConfig *rest.Config, request *application.K8sRequestBean, manifest string) (resp *application.ManifestResponse, err error) {
	//TODO implement me
	panic("implement me")
}

func (n NewK8sClientServiceImplMock) UpdateResource(restConfig *rest.Config, request *application.K8sRequestBean) (resp *application.ManifestResponse, err error) {
	//TODO implement me
	panic("implement me")
}

func (n NewK8sClientServiceImplMock) DeleteResource(restConfig *rest.Config, request *application.K8sRequestBean) (resp *application.ManifestResponse, err error) {
	//TODO implement me
	panic("implement me")
}

func (n NewK8sClientServiceImplMock) ListEvents(restConfig *rest.Config, request *application.K8sRequestBean) (*application.EventsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (n NewK8sClientServiceImplMock) GetPodLogs(restConfig *rest.Config, request *application.K8sRequestBean) (io.ReadCloser, error) {
	//TODO implement me
	panic("implement me")
}

func Test_GetManifestsInBatch(t *testing.T) {
	var (
		k8sCS          = NewK8sClientServiceImplMock{}
		clusterService = NewClusterServiceMock{}
		impl           = NewK8sApplicationServiceImpl(
			nil, clusterService, nil, k8sCS, nil,
			nil, nil, nil, nil, nil)
	)
	n := 10
	kinds := []string{"Service", "Ingress", "Random", "Invalid"}
	var testInput = make([]ResourceRequestBean, 0)
	expectedTestOutputs := make([]BatchResourceResponse, 0)
	for i := 0; i < n; i++ {
		idx := rand.Int31n(int32(len(kinds)))
		inp := generateTestResourceRequest(kinds[idx])
		testInput = append(testInput, inp)
	}
	for i := 0; i < n; i++ {
		man := generateTestManifest(testInput[i].K8sRequest.ResourceIdentifier.GroupVersionKind.Kind)
		bRR := BatchResourceResponse{
			ManifestResponse: &man,
			Err:              nil,
		}
		expectedTestOutputs = append(expectedTestOutputs, bRR)
	}

	t.Run(fmt.Sprint("test1"), func(t *testing.T) {
		resultOutput := impl.GetHostUrlsByBatch(testInput)
		//check if all the output manifests are expected
		for j, _ := range resultOutput {
			if !cmp.Equal(resultOutput[j], expectedTestOutputs[j]) {
				t.Errorf("expected %+v but got %+v", expectedTestOutputs[j].ManifestResponse, resultOutput[j].ManifestResponse)
				break
			}
		}

	})

}
func generateTestResourceRequest(kind string) ResourceRequestBean {
	return ResourceRequestBean{
		AppIdentifier: &client.AppIdentifier{},
		K8sRequest: &application.K8sRequestBean{
			ResourceIdentifier: application.ResourceIdentifier{
				GroupVersionKind: schema.GroupVersionKind{
					Kind: kind,
				},
			},
		},
	}
}

type test struct {
	inp application.ManifestResponse
	out Response
}

func Test_getUrls(t *testing.T) {
	impl := NewK8sApplicationServiceImpl(
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil)
	tests := make([]test, 3)
	tests[0] = test{
		inp: generateTestManifest("Service"),
		out: Response{
			Kind:     "Service",
			Name:     "test-service",
			PointsTo: "aws.ebs.23456",
			Urls:     make([]string, 0),
		},
	}
	tests[1] = test{
		inp: generateTestManifest("Ingress"),
		out: Response{
			Kind:     "Ingress",
			Name:     "test-service",
			PointsTo: "aws.ebs.23456",
			Urls:     []string{"demo1.devtron.info/orchestrator", "demo1.devtron.info/dashboard"},
		},
	}
	tests[2] = test{
		inp: generateTestManifest("Invalid"),
		out: Response{
			Kind:     "",
			Name:     "",
			PointsTo: "",
			Urls:     make([]string, 0),
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint("testcase:", i), func(t *testing.T) {
			resultGot := impl.getUrls(&tt.inp)
			if !cmp.Equal(resultGot, tt.out) {
				t.Errorf("expected %s but got %s", tt.out, resultGot)
			}
		})
	}
}

func generateTestManifest(kind string) application.ManifestResponse {
	return application.ManifestResponse{
		Manifest: unstructured.Unstructured{
			Object: getObj(kind),
		},
	}
}

func getObj(kind string) map[string]interface{} {
	var obj map[string]interface{}
	if (kind != "Service") && (kind != "Ingress") {
		fmt.Println(kind)
		manifest = `{"invalid":{}}`
	}
	err := json.Unmarshal([]byte(manifest), &obj)
	if err != nil {
		fmt.Print("error in marshaling : ", err)
		return nil
	}
	if _, ok := obj["kind"]; ok {
		obj["kind"] = kind
	}
	return obj
}
