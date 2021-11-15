package pipeline
//
//import (
//	"encoding/json"
//	"fmt"
//	"github.com/devtron-labs/devtron/client/argocdServer/repository"
//	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
//	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
//	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
//	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
//	"github.com/devtron-labs/devtron/internal/util"
//	"go.uber.org/zap"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
//	"k8s.io/apimachinery/pkg/runtime/schema"
//	"net/http"
//	"reflect"
//	"testing"
//)
//const RequestChartRefId = 12
//var templateRequest = `{
//    "code": 200,
//    "status": "OK",
//    "result": {
//        "globalConfig": {
//            "id": 195,
//            "appId": 454,
//            "refChartTemplate": "reference-chart_3-9-0",
//            "refChartTemplateVersion": "3.9.0",
//            "chartRepositoryId": 1,
//			"valuesOverride": nil,
//            "defaultAppOverride": {
//                "ContainerPort": [
//                    {
//                        "envoyPort": 8799,
//                        "idleTimeout": "1800s",
//                        "name": "app",
//                        "port": 8080,
//                        "servicePort": 80,
//                        "supportStreaming": true,
//                        "useHTTP2": true
//                    }
//                ],
//                "EnvVariables": [],
//                "GracePeriod": 31,
//                "LivenessProbe": {
//                    "Path": "",
//                    "command": [],
//                    "failureThreshold": 3,
//                    "httpHeader": {
//                        "name": "",
//                        "value": ""
//                    },
//                    "initialDelaySeconds": 20,
//                    "periodSeconds": 10,
//                    "port": 8080,
//                    "scheme": "",
//                    "successThreshold": 1,
//                    "tcp": false,
//                    "timeoutSeconds": 5
//                },
//                "MaxSurge": 1,
//                "MaxUnavailable": 0,
//                "MinReadySeconds": 60,
//                "ReadinessProbe": {
//                    "Path": "",
//                    "command": [],
//                    "failureThreshold": 3,
//                    "httpHeader": {
//                        "name": "",
//                        "value": ""
//                    },
//                    "initialDelaySeconds": 20,
//                    "periodSeconds": 10,
//                    "port": 8080,
//                    "scheme": "",
//                    "successThreshold": 1,
//                    "tcp": false,
//                    "timeoutSeconds": 5
//                },
//                "Spec": {
//                    "Affinity": {
//                        "Key": null,
//                        "Values": "nodes",
//                        "key": ""
//                    }
//                },
//                "args": {
//                    "enabled": false,
//                    "value": [
//                        "/bin/sh",
//                        "-c",
//                        "touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600"
//                    ]
//                },
//                "autoscaling": {
//                    "MaxReplicas": 2,
//                    "MinReplicas": 1,
//                    "TargetCPUUtilizationPercentage": 90,
//                    "TargetMemoryUtilizationPercentage": 80,
//                    "enabled": false,
//                    "extraMetrics": []
//                },
//                "command": {
//                    "enabled": false,
//                    "value": []
//                },
//                "containers": [],
//                "dbMigrationConfig": {
//                    "enabled": false
//                },
//                "envoyproxy": {
//                    "configMapName": "",
//                    "image": "envoyproxy/envoy:v1.14.1",
//                    "resources": {
//                        "limits": {
//                            "cpu": "0.05",
//                            "memory": "50Mi"
//                        },
//                        "requests": {
//                            "cpu": "40m",
//                            "memory": "50Mi"
//                        }
//                    }
//                },
//                "image": {
//                    "pullPolicy": "IfNotPresent"
//                },
//                "imagePullSecrets": [],
//                "ingress": {
//                    "annotations": {
//                        "kubernetes.io/ingress.class": "nginx",
//                        "nginx.ingress.kubernetes.io/force-ssl-redirect": "false",
//                        "nginx.ingress.kubernetes.io/ssl-redirect": "false"
//                    },
//                    "enabled": false,
//                    "hosts": [
//                        {
//                            "host": "chart-example1.local",
//                            "paths": [
//                                "/example1"
//                            ]
//                        },
//                        {
//                            "host": "chart-example2.local",
//                            "paths": [
//                                "/example2",
//                                "/example2/healthz"
//                            ]
//                        }
//                    ],
//                    "tls": []
//                },
//                "ingressInternal": {
//                    "annotations": {},
//                    "enabled": false,
//                    "hosts": [
//                        {
//                            "host": "chart-example1.internal",
//                            "paths": [
//                                "/example1"
//                            ]
//                        },
//                        {
//                            "host": "chart-example2.internal",
//                            "paths": [
//                                "/example2",
//                                "/example2/healthz"
//                            ]
//                        }
//                    ],
//                    "tls": []
//                },
//                "initContainers": [],
//                "pauseForSecondsBeforeSwitchActive": 30,
//                "prometheus": {
//                    "release": "monitoring"
//                },
//                "rawYaml": [],
//                "replicaCount": 1,
//                "resources": {
//                    "limits": {
//                        "cpu": "0.05",
//                        "memory": "50Mi"
//                    },
//                    "requests": {
//                        "cpu": "0.04",
//                        "memory": "10Mi"
//                    }
//                },
//                "secret": {
//                    "data": {},
//                    "enabled": false
//                },
//                "server": {
//                    "deployment": {
//                        "image": "",
//                        "image_tag": "1-95af053"
//                    }
//                },
//                "service": {
//                    "annotations": {},
//                    "type": "ClusterIP"
//                },
//                "servicemonitor": {
//                    "additionalLabels": {}
//                },
//                "tolerations": [],
//                "volumeMounts": [],
//                "volumes": [],
//                "waitForSecondsBeforeScalingDown": 30
//            },
//            "chartRefId": 10,
//            "latest": true,
//            "isAppMetricsEnabled": false
//        }
//    }
//}`
//
//const messages=`{
//    "code": 200,
//    "status": "OK",
//    "result": {
//        "globalConfig": {
//            "id": 195,
//            "appId": 454,
//            "refChartTemplate": "reference-chart_3-9-0",
//            "refChartTemplateVersion": "3.9.0",
//            "chartRepositoryId": 1,
//            "defaultAppOverride": {
//                "ContainerPort": [
//                    {
//                        "envoyPort": 8799,
//                        "idleTimeout": "1800s",
//                        "name": "app",
//                        "port": 8080,
//                        "servicePort": 80,
//                        "supportStreaming": true,
//                        "useHTTP2": true
//                    }
//                ],
//                "EnvVariables": [],
//                "GracePeriod": 31,
//                "LivenessProbe": {
//                    "Path": "",
//                    "command": [],
//                    "failureThreshold": 3,
//                    "httpHeader": {
//                        "name": "",
//                        "value": ""
//                    },
//                    "initialDelaySeconds": 20,
//                    "periodSeconds": 10,
//                    "port": 8080,
//                    "scheme": "",
//                    "successThreshold": 1,
//                    "tcp": false,
//                    "timeoutSeconds": 5
//                },
//                "MaxSurge": 1,
//                "MaxUnavailable": 0,
//                "MinReadySeconds": 60,
//                "ReadinessProbe": {
//                    "Path": "",
//                    "command": [],
//                    "failureThreshold": 3,
//                    "httpHeader": {
//                        "name": "",
//                        "value": ""
//                    },
//                    "initialDelaySeconds": 20,
//                    "periodSeconds": 10,
//                    "port": 8080,
//                    "scheme": "",
//                    "successThreshold": 1,
//                    "tcp": false,
//                    "timeoutSeconds": 5
//                },
//                "Spec": {
//                    "Affinity": {
//                        "Key": null,
//                        "Values": "nodes",
//                        "key": ""
//                    }
//                },
//                "args": {
//                    "enabled": false,
//                    "value": [
//                        "/bin/sh",
//                        "-c",
//                        "touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600"
//                    ]
//                },
//                "autoscaling": {
//                    "MaxReplicas": 2,
//                    "MinReplicas": 1,
//                    "TargetCPUUtilizationPercentage": 90,
//                    "TargetMemoryUtilizationPercentage": 80,
//                    "enabled": false,
//                    "extraMetrics": []
//                },
//                "command": {
//                    "enabled": false,
//                    "value": []
//                },
//                "containers": [],
//                "dbMigrationConfig": {
//                    "enabled": false
//                },
//                "envoyproxy": {
//                    "configMapName": "",
//                    "image": "envoyproxy/envoy:v1.14.1",
//                    "resources": {
//                        "limits": {
//                            "cpu": "0.05",
//                            "memory": "50Mi"
//                        },
//                        "requests": {
//                            "cpu": "40m",
//                            "memory": "50Mi"
//                        }
//                    }
//                },
//                "image": {
//                    "pullPolicy": "IfNotPresent"
//                },
//                "imagePullSecrets": [],
//                "ingress": {
//                    "annotations": {
//                        "kubernetes.io/ingress.class": "nginx",
//                        "nginx.ingress.kubernetes.io/force-ssl-redirect": "false",
//                        "nginx.ingress.kubernetes.io/ssl-redirect": "false"
//                    },
//                    "enabled": false,
//                    "hosts": [
//                        {
//                            "host": "chart-example1.local",
//                            "paths": [
//                                "/example1"
//                            ]
//                        },
//                        {
//                            "host": "chart-example2.local",
//                            "paths": [
//                                "/example2",
//                                "/example2/healthz"
//                            ]
//                        }
//                    ],
//                    "tls": []
//                },
//                "ingressInternal": {
//                    "annotations": {},
//                    "enabled": false,
//                    "hosts": [
//                        {
//                            "host": "chart-example1.internal",
//                            "paths": [
//                                "/example1"
//                            ]
//                        },
//                        {
//                            "host": "chart-example2.internal",
//                            "paths": [
//                                "/example2",
//                                "/example2/healthz"
//                            ]
//                        }
//                    ],
//                    "tls": []
//                },
//                "initContainers": [],
//                "pauseForSecondsBeforeSwitchActive": 30,
//                "prometheus": {
//                    "release": "monitoring"
//                },
//                "rawYaml": [],
//                "replicaCount": 1,
//                "resources": {
//                    "limits": {
//                        "cpu": "0.05",
//                        "memory": "50Mi"
//                    },
//                    "requests": {
//                        "cpu": "0.04",
//                        "memory": "10Mi"
//                    }
//                },
//                "secret": {
//                    "data": {},
//                    "enabled": false
//                },
//                "server": {
//                    "deployment": {
//                        "image": "",
//                        "image_tag": "1-95af053"
//                    }
//                },
//                "service": {
//                    "annotations": {},
//                    "type": "ClusterIP"
//                },
//                "servicemonitor": {
//                    "additionalLabels": {}
//                },
//                "tolerations": [],
//                "volumeMounts": [],
//                "volumes": [],
//                "waitForSecondsBeforeScalingDown": 30
//            },
//            "chartRefId": 10,
//            "latest": true,
//            "isAppMetricsEnabled": false
//        }
//    }
//}`
//func TestChartServiceImpl_DefaultTemplateWithSavedTemplateData(t *testing.T) {
//	type fields struct {
//		chartRepository           chartConfig.ChartRepository
//		logger                    *zap.SugaredLogger
//		repoRepository            chartConfig.ChartRepoRepository
//		chartTemplateService      util.ChartTemplateService
//		pipelineGroupRepository   pipelineConfig.AppRepository
//		mergeUtil                 util.MergeUtil
//		repositoryService         repository.ServiceClient
//		refChartDir               RefChartDir
//		defaultChart              DefaultChart
//		chartRefRepository        chartConfig.ChartRefRepository
//		envOverrideRepository     chartConfig.EnvConfigOverrideRepository
//		pipelineConfigRepository  chartConfig.PipelineConfigRepository
//		configMapRepository       chartConfig.ConfigMapRepository
//		environmentRepository     cluster.EnvironmentRepository
//		pipelineRepository        pipelineConfig.PipelineRepository
//		appLevelMetricsRepository repository2.AppLevelMetricsRepository
//		client                    *http.Client
//	}
//	type args struct {
//		RequestChartRefId int
//		templateRequest   string
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    string
//		wantErr bool
//	}{
//		{
//			name:   "base listing case",
//			fields: fields{
//				chartRepository:           tt.fields.chartRepository,
//				logger:                    tt.fields.logger,
//				repoRepository:            tt.fields.repoRepository,
//				chartTemplateService:      tt.fields.chartTemplateService,
//				pipelineGroupRepository:   tt.fields.pipelineGroupRepository,
//				mergeUtil:                 tt.fields.mergeUtil,
//				repositoryService:         tt.fields.repositoryService,
//				refChartDir:               tt.fields.refChartDir,
//				defaultChart:              tt.fields.defaultChart,
//				chartRefRepository:        tt.fields.chartRefRepository,
//				envOverrideRepository:     tt.fields.envOverrideRepository,
//				pipelineConfigRepository:  tt.fields.pipelineConfigRepository,
//				configMapRepository:       tt.fields.configMapRepository,
//				environmentRepository:     tt.fields.environmentRepository,
//				pipelineRepository:        tt.fields.pipelineRepository,
//				appLevelMetricsRepository: tt.fields.appLevelMetricsRepository,
//				client:                    tt.fields.client,
//			},
//			args: args{
//				RequestChartRefId: RequestChartRefId,
//				templateRequest: templateRequest,
//			},
//			want: messages,
//			wantErr: false,
//
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			impl := ChartServiceImpl{
//				chartRepository:           tt.fields.chartRepository,
//				logger:                    tt.fields.logger,
//				repoRepository:            tt.fields.repoRepository,
//				chartTemplateService:      tt.fields.chartTemplateService,
//				pipelineGroupRepository:   tt.fields.pipelineGroupRepository,
//				mergeUtil:                 tt.fields.mergeUtil,
//				repositoryService:         tt.fields.repositoryService,
//				refChartDir:               tt.fields.refChartDir,
//				defaultChart:              tt.fields.defaultChart,
//				chartRefRepository:        tt.fields.chartRefRepository,
//				envOverrideRepository:     tt.fields.envOverrideRepository,
//				pipelineConfigRepository:  tt.fields.pipelineConfigRepository,
//				configMapRepository:       tt.fields.configMapRepository,
//				environmentRepository:     tt.fields.environmentRepository,
//				pipelineRepository:        tt.fields.pipelineRepository,
//				appLevelMetricsRepository: tt.fields.appLevelMetricsRepository,
//				client:                    tt.fields.client,
//			}
//			got, err := impl.DefaultTemplateWithSavedTemplateData(tt.args.RequestChartRefId, tt.args.templateRequest)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("DefaultTemplateWithSavedTemplateData() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("DefaultTemplateWithSavedTemplateData() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
