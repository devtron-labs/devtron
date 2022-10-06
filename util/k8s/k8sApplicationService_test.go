package k8s

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/client/k8s/application"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

type test struct {
	inp application.ManifestResponse
	out Response
}

func Test_getUrls(t *testing.T) {
	impl := NewK8sApplicationServiceImpl(
		nil, nil, nil, nil, nil,
		nil, nil)

	tests := make([]test, 3)
	tests[0] = test{
		inp: application.ManifestResponse{
			Manifest: unstructured.Unstructured{
				Object: getObj("Service"),
			},
		},
		out: Response{
			Kind:     "Service",
			Name:     "test-service",
			PointsTo: "aws.ebs.23456",
			Urls:     make([]string, 0),
		},
	}
	tests[1] = test{
		inp: application.ManifestResponse{
			Manifest: unstructured.Unstructured{
				Object: getObj("Ingress"),
			},
		},
		out: Response{
			Kind:     "Ingress",
			Name:     "test-service",
			PointsTo: "aws.ebs.23456",
			Urls:     []string{"demo1.devtron.info/orchestrator", "demo1.devtron.info/dashboard"},
		},
	}
	tests[2] = test{
		inp: application.ManifestResponse{
			Manifest: unstructured.Unstructured{
				Object: getObj("Invalid"),
			},
		},
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

func getObj(kind string) map[string]interface{} {
	var obj map[string]interface{}
	if kind == "Invalid" {
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
