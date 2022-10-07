package k8s

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/mock"
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

type NewK8sApplicationServiceImplMock struct {
	mock.Mock
}

var testInput = make([]ResourceRequestAndGroupVersionKind, 0)
var ptr = 0

func (n NewK8sApplicationServiceImplMock) GetResource(restConfig *rest.Config, request *application.K8sRequestBean) (resp *application.ManifestResponse, err error) {

}

func (n NewK8sApplicationServiceImplMock) CreateResource(restConfig *rest.Config, request *application.K8sRequestBean, manifest string) (resp *application.ManifestResponse, err error) {
	//TODO implement me
	panic("implement me")
}

func (n NewK8sApplicationServiceImplMock) UpdateResource(restConfig *rest.Config, request *application.K8sRequestBean) (resp *application.ManifestResponse, err error) {
	//TODO implement me
	panic("implement me")
}

func (n NewK8sApplicationServiceImplMock) DeleteResource(restConfig *rest.Config, request *application.K8sRequestBean) (resp *application.ManifestResponse, err error) {
	//TODO implement me
	panic("implement me")
}

func (n NewK8sApplicationServiceImplMock) ListEvents(restConfig *rest.Config, request *application.K8sRequestBean) (*application.EventsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (n NewK8sApplicationServiceImplMock) GetPodLogs(restConfig *rest.Config, request *application.K8sRequestBean) (io.ReadCloser, error) {
	//TODO implement me
	panic("implement me")
}

var (
	k8sASI = NewK8sApplicationServiceImplMock{}
	impl   = NewK8sApplicationServiceImpl(
		nil, nil, nil, k8sASI, nil,
		nil, nil)
)

func Test_GetManifestsInBatch(t *testing.T) {
	n := 10
	kinds := []string{"Service", "Ingress", "Random", "Invalid"}
	ExpectedTestOutputs := make([]BatchResourceResponse, 0)
	for i := 0; i < n; i++ {
		idx := rand.Int31n(int32(len(kinds)))
		inp := generateTestResourceRequestAndGroupVersionKind(kinds[idx])
		testInput = append(testInput, inp)
	}
	for i := 0; i < n; i++ {
		man := generateTestManifest(testInput[i].Kind)
		bRR := BatchResourceResponse{
			ManifestResponse: &man,
			Err:              nil,
		}
		ExpectedTestOutputs = append(ExpectedTestOutputs, bRR)
	}

}
func generateTestResourceRequestAndGroupVersionKind(kind string) ResourceRequestAndGroupVersionKind {
	return ResourceRequestAndGroupVersionKind{
		Kind: kind,
	}
}

type test struct {
	inp application.ManifestResponse
	out Response
}

func Test_getUrls(t *testing.T) {
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
