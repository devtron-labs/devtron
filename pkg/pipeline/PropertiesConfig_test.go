package pipeline

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"github.com/devtron-labs/devtron/internal/util"
	"go.uber.org/zap"
	"reflect"
	"testing"
)


func TestPropertiesConfigServiceImpl_MergeValidation(t *testing.T) {
	type fields struct {
		logger                       *zap.SugaredLogger
		envConfigRepo                chartConfig.EnvConfigOverrideRepository
		chartRepo                    chartConfig.ChartRepository
		mergeUtil                    util.MergeUtil
		environmentRepository        cluster.EnvironmentRepository
		dbPipelineOrchestrator       DbPipelineOrchestrator
		application                  application.ServiceClient
		envLevelAppMetricsRepository repository.EnvLevelAppMetricsRepository
		appLevelMetricsRepository    repository.AppLevelMetricsRepository
	}
	type args struct {
		defaultTemplate   map[string]json.RawMessage
		EnvOverrideValues json.RawMessage
	}
	merged := json.RawMessage([]byte(`{"ContainerPort":[{"envoyPort":8799,"idleTimeout":"1800s","name":"app","port":8080,"servicePort":80,"supportStreaming":true,"useHTTP2":true}],"EnvVariables":[],"GracePeriod":42,"LivenessProbe":{"Path":"","command":[],"failureThreshold":3,"httpHeader":{"name":"","value":""},"initialDelaySeconds":20,"periodSeconds":10,"port":8080,"scheme":"","successThreshold":1,"tcp":false,"timeoutSeconds":5},"MaxSurge":1,"MaxUnavailable":0,"MinReadySeconds":60,"ReadinessProbe":{"Path":"","command":[],"failureThreshold":3,"httpHeader":{"name":"","value":""},"initialDelaySeconds":20,"periodSeconds":10,"port":8080,"scheme":"","successThreshold":1,"tcp":false,"timeoutSeconds":5},"Spec":{"Affinity":{"Values":"nodes","key":""}},"args":{"enabled":false,"value":["/bin/sh","-c","touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600"]},"autoscaling":{"MaxReplicas":2,"MinReplicas":1,"TargetCPUUtilizationPercentage":90,"TargetMemoryUtilizationPercentage":80,"enabled":false,"extraMetrics":[]},"command":{"enabled":false,"value":[]},"containers":[],"dbMigrationConfig":{"enabled":false},"envoyproxy":{"configMapName":"","image":"envoyproxy/envoy:v1.14.1","resources":{"limits":{"cpu":"50m","memory":"50Mi"},"requests":{"cpu":"50m","memory":"50Mi"}}},"image":{"pullPolicy":"IfNotPresent"},"imagePullSecrets":[],"ingress":{"annotations":{"kubernetes.io/ingress.class":"nginx","nginx.ingress.kubernetes.io/force-ssl-redirect":"false","nginx.ingress.kubernetes.io/ssl-redirect":"false"},"enabled":false,"hosts":[{"host":"chart-example1.local","paths":["/example1"]},{"host":"chart-example2.local","paths":["/example2","/example2/healthz"]}],"tls":[]},"ingressInternal":{"annotations":{},"enabled":false,"hosts":[{"host":"chart-example1.internal","paths":["/example1"]},{"host":"chart-example2.internal","paths":["/example2","/example2/healthz"]}],"tls":[]},"initContainers":[],"pauseForSecondsBeforeSwitchActive":30,"podAnnotations":{},"podLabels":{},"prometheus":{"release":"monitoring"},"rawYaml":[],"replicaCount":1,"resources":{"limits":{"cpu":"0.05","memory":"50Mi"},"requests":{"cpu":"0.01","memory":"10Mi"}},"secret":{"data":{},"enabled":false},"server":{"deployment":{"image":"","image_tag":"1-95af053"}},"service":{"annotations":{},"type":"ClusterIP"},"servicemonitor":{"additionalLabels":{}},"tolerations":[],"volumeMounts":[],"volumes":[],"waitForSecondsBeforeScalingDown":30}`))
	defaultTemplateByteValue := []byte(`{"defaultAppOverride":{"ContainerPort":[{"envoyPort":8799,"idleTimeout":"1800s","name":"app","port":8080,"servicePort":80,"supportStreaming":true,"useHTTP2":true}],"EnvVariables":[],"GracePeriod":30,"LivenessProbe":{"Path":"","command":[],"failureThreshold":3,"httpHeader":{"name":"","value":""},"initialDelaySeconds":20,"periodSeconds":10,"port":8080,"scheme":"","successThreshold":1,"tcp":false,"timeoutSeconds":5},"MaxSurge":1,"MaxUnavailable":0,"MinReadySeconds":60,"ReadinessProbe":{"Path":"","command":[],"failureThreshold":3,"httpHeader":{"name":"","value":""},"initialDelaySeconds":20,"periodSeconds":10,"port":8080,"scheme":"","successThreshold":1,"tcp":false,"timeoutSeconds":5},"Spec":{"Affinity":{"Key":null,"Values":"nodes","key":""}},"args":{"enabled":false,"value":["/bin/sh","-c","touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600"]},"autoscaling":{"MaxReplicas":2,"MinReplicas":1,"TargetCPUUtilizationPercentage":90,"TargetMemoryUtilizationPercentage":80,"enabled":false,"extraMetrics":[]},"command":{"enabled":false,"value":[]},"containers":[],"dbMigrationConfig":{"enabled":false},"envoyproxy":{"configMapName":"","image":"envoyproxy/envoy:v1.14.1","resources":{"limits":{"cpu":"50m","memory":"50Mi"},"requests":{"cpu":"50m","memory":"50Mi"}}},"image":{"pullPolicy":"IfNotPresent"},"imagePullSecrets":[],"ingress":{"annotations":{"kubernetes.io/ingress.class":"nginx","nginx.ingress.kubernetes.io/force-ssl-redirect":"false","nginx.ingress.kubernetes.io/ssl-redirect":"false"},"enabled":false,"hosts":[{"host":"chart-example1.local","paths":["/example1"]},{"host":"chart-example2.local","paths":["/example2","/example2/healthz"]}],"tls":[]},"ingressInternal":{"annotations":{},"enabled":false,"hosts":[{"host":"chart-example1.internal","paths":["/example1"]},{"host":"chart-example2.internal","paths":["/example2","/example2/healthz"]}],"tls":[]},"initContainers":[],"pauseForSecondsBeforeSwitchActive":30,"podAnnotations":{},"podLabels":{},"prometheus":{"release":"monitoring"},"rawYaml":[],"replicaCount":1,"resources":{"limits":{"cpu":"0.05","memory":"50Mi"},"requests":{"cpu":"0.01","memory":"10Mi"}},"secret":{"data":{},"enabled":false},"server":{"deployment":{"image":"","image_tag":"1-95af053"}},"service":{"annotations":{},"type":"ClusterIP"},"servicemonitor":{"additionalLabels":{}},"tolerations":[],"volumeMounts":[],"volumes":[],"waitForSecondsBeforeScalingDown":30}`)
	var defaultTemplate map[string]json.RawMessage
	if err := json.Unmarshal(defaultTemplateByteValue, &defaultTemplate); err != nil {
		return
	}
	EnvOverrideValues := json.RawMessage([]byte(`{
    "ContainerPort": [
      {
        "envoyPort": 8799,
        "idleTimeout": "1800s",
        "name": "app",
        "port": 8080,
        "servicePort": 80,
        "supportStreaming": true,
        "useHTTP2": true
      }
    ],
    "EnvVariables": [],
    "GracePeriod": 42,
    "LivenessProbe": {
      "Path": "",
      "command": [],
      "failureThreshold": 3,
      "httpHeader": {
        "name": "",
        "value": ""
      },
      "initialDelaySeconds": 20,
      "periodSeconds": 10,
      "port": 8080,
      "scheme": "",
      "successThreshold": 1,
      "tcp": false,
      "timeoutSeconds": 5
    },
    "MaxSurge": 1,
    "MaxUnavailable": 0,
    "MinReadySeconds": 60,
    "ReadinessProbe": {
      "Path": "",
      "command": [],
      "failureThreshold": 3,
      "httpHeader": {
        "name": "",
        "value": ""
      },
      "initialDelaySeconds": 20,
      "periodSeconds": 10,
      "port": 8080,
      "scheme": "",
      "successThreshold": 1,
      "tcp": false,
      "timeoutSeconds": 5
    },
    "Spec": {
      "Affinity": {
        "Key": null,
        "Values": "nodes",
        "key": ""
      }
    },
    "args": {
      "enabled": false,
      "value": [
        "/bin/sh",
        "-c",
        "touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600"
      ]
    },
    "autoscaling": {
      "MaxReplicas": 2,
      "MinReplicas": 1,
      "TargetCPUUtilizationPercentage": 90,
      "TargetMemoryUtilizationPercentage": 80,
      "enabled": false,
      "extraMetrics": []
    },
    "command": {
      "enabled": false,
      "value": []
    },
    "containers": [],
    "dbMigrationConfig": {
      "enabled": false
    },
    "envoyproxy": {
      "configMapName": "",
      "image": "envoyproxy/envoy:v1.14.1",
      "resources": {
        "limits": {
          "cpu": "50m",
          "memory": "50Mi"
        },
        "requests": {
          "cpu": "50m",
          "memory": "50Mi"
        }
      }
    },
    "image": {
      "pullPolicy": "IfNotPresent"
    },
    "imagePullSecrets": [],
    "ingress": {
      "annotations": {
        "kubernetes.io/ingress.class": "nginx",
        "nginx.ingress.kubernetes.io/force-ssl-redirect": "false",
        "nginx.ingress.kubernetes.io/ssl-redirect": "false"
      },
      "enabled": false,
      "hosts": [
        {
          "host": "chart-example1.local",
          "paths": [
            "/example1"
          ]
        },
        {
          "host": "chart-example2.local",
          "paths": [
            "/example2",
            "/example2/healthz"
          ]
        }
      ],
      "tls": []
    },
   
    "initContainers": [],
    "pauseForSecondsBeforeSwitchActive": 30,
    "prometheus": {
      "release": "monitoring"
    },
    "rawYaml": [],
    "replicaCount": 1,
    "resources": {
      "limits": {
        "cpu": "0.05",
        "memory": "50Mi"
      },
      "requests": {
        "cpu": "0.01",
        "memory": "10Mi"
      }
    },
    "secret": {
      "data": {},
      "enabled": false
    },
    "server": {
      "deployment": {
        "image": "",
        "image_tag": "1-95af053"
      }
    },
    "service": {
      "annotations": {},
      "type": "ClusterIP"
    },
    "servicemonitor": {
      "additionalLabels": {}
    },
    "tolerations": [],
    "volumeMounts": [],
    "volumes": [],
    "waitForSecondsBeforeScalingDown": 30
  }`))
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    json.RawMessage
		wantErr bool
	}{
		{
			name : "first_test",
			fields: fields{},
			args: args{
				defaultTemplate: defaultTemplate,
				EnvOverrideValues: EnvOverrideValues,
			},
			want: merged,
			wantErr: true,
		},
		
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := PropertiesConfigServiceImpl{
				logger:                       tt.fields.logger,
				envConfigRepo:                tt.fields.envConfigRepo,
				chartRepo:                    tt.fields.chartRepo,
				mergeUtil:                    tt.fields.mergeUtil,
				environmentRepository:        tt.fields.environmentRepository,
				dbPipelineOrchestrator:       tt.fields.dbPipelineOrchestrator,
				application:                  tt.fields.application,
				envLevelAppMetricsRepository: tt.fields.envLevelAppMetricsRepository,
				appLevelMetricsRepository:    tt.fields.appLevelMetricsRepository,
			}
			got, err := impl.MergeWithDefaultTemplate(tt.args.defaultTemplate, tt.args.EnvOverrideValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("MergeValidation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeValidation() got = %v, want %v", got, tt.want)
			}
		})
	}
}
