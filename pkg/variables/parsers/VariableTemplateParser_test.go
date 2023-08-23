package parsers

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVariableTemplateParserImpl_ExtractVariables(t *testing.T) {
	logger, err := util.NewSugardLogger()
	assert.Nil(t, err)

	t.Run("extract variables", func(t *testing.T) {
		templateParser := NewVariableTemplateParserImpl(logger)
		sampleTemplate := `{"ConfigMaps":{"enabled":false,"maps":[]},"ConfigSecrets":{"enabled":false,"secrets":[]},"ContainerPort":[{"envoyPort":@{{envoyPort}},"idleTimeout":"@{{idleTimeoutVar / idleTimeoutDivVar}}s","name":"${1 + appName}","port":8080,"servicePort":80,"supportStreaming":false,"useHTTP2":false}],"EnvVariables":[],"EnvVariablesFromFieldPath":[{"fieldPath":"metadata.name","name":"POD_NAME"}],"GracePeriod":30,"LivenessProbe":{"Path":"","command":[],"failureThreshold":3,"httpHeaders":[],"initialDelaySeconds":20,"periodSeconds":10,"port":8080,"scheme":"","successThreshold":1,"tcp":false,"timeoutSeconds":5},"MaxSurge":1,"MaxUnavailable":0,"MinReadySeconds":60,"ReadinessProbe":{"Path":"","command":[],"failureThreshold":3,"httpHeaders":[],"initialDelaySeconds":20,"periodSeconds":10,"port":8080,"scheme":"","successThreshold":1,"tcp":false,"timeoutSeconds":5},"Spec":{"Affinity":{"Values":"nodes","key":""}},"ambassadorMapping":{"ambassadorId":"","cors":{},"enabled":false,"hostname":"devtron.example.com","labels":{},"prefix":"/","retryPolicy":{},"rewrite":"","tls":{"context":"","create":false,"hosts":[],"secretName":""}},"args":{"enabled":false,"value":["/bin/sh","-c","touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600"]},"autoPromotionSeconds":30,"autoscaling":{"MaxReplicas":2,"MinReplicas":1,"TargetCPUUtilizationPercentage":90,"TargetMemoryUtilizationPercentage":80,"annotations":{},"behavior":{},"enabled":false,"extraMetrics":[],"labels":{}},"command":{"enabled":false,"value":[],"workingDir":{}},"containerExtraSpecs":{},"containerSecurityContext":{},"containerSpec":{"lifecycle":{"enabled":false,"postStart":{"httpGet":{"host":"example.com","path":"/example","port":90}},"preStop":{"exec":{"command":["sleep","10"]}}}},"containers":[],"dbMigrationConfig":{"enabled":false},"envoyproxy":{"configMapName":"","image":"quay.io/devtron/envoy:v1.14.1","lifecycle":{},"resources":{"limits":{"cpu":"50m","memory":"50Mi"},"requests":{"cpu":"50m","memory":"50Mi"}}},"hostAliases":[],"image":{"pullPolicy":"IfNotPresent"},"imagePullSecrets":[],"ingress":{"annotations":{},"className":"","enabled":false,"hosts":[{"host":"chart-example1.local","pathType":"ImplementationSpecific","paths":["/example1"]},{"host":"chart-example2.local","pathType":"ImplementationSpecific","paths":["/example2","/example2/healthz"]}],"labels":{},"tls":[]},"ingressInternal":{"annotations":{},"className":"","enabled":false,"hosts":[{"host":"chart-example1.internal","pathType":"ImplementationSpecific","paths":["/example1"]},{"host":"chart-example2.internal","pathType":"ImplementationSpecific","paths":["/example2","/example2/healthz"]}],"tls":[]},"initContainers":[],"istio":{"enable":false,"gateway":{"annotations":{},"enabled":false,"host":"example.com","labels":{},"tls":{"enabled":false,"secretName":"secret-name"}},"virtualService":{"annotations":{},"enabled":false,"gateways":[],"hosts":[],"http":[{"corsPolicy":{},"headers":{},"match":[{"uri":{"prefix":"/v1"}},{"uri":{"prefix":"/v2"}}],"retries":{"attempts":2,"perTryTimeout":"3s"},"rewriteUri":"/","route":[{"destination":{"host":"service1","port":80}}],"timeout":"12s"},{"route":[{"destination":{"host":"service2"}}]}],"labels":{}}},"kedaAutoscaling":{"advanced":{},"authenticationRef":{},"cooldownPeriod":300,"enabled":false,"envSourceContainerName":"","fallback":{},"idleReplicaCount":0,"maxReplicaCount":2,"minReplicaCount":1,"pollingInterval":30,"triggerAuthentication":{"enabled":false,"name":"","spec":{}},"triggers":[]},"nodeSelector":{},"orchestrator.deploymant.algo":1,"pauseForSecondsBeforeSwitchActive":30,"podAnnotations":{},"podDisruptionBudget":{},"podExtraSpecs":{},"podLabels":{},"podSecurityContext":{},"prometheus":{"release":"monitoring"},"prometheusRule":{"additionalLabels":{},"enabled":false,"namespace":""},"rawYaml":[],"replicaCount":1,"resources":{"limits":{"cpu":"0.05","memory":"50Mi"},"requests":{"cpu":"0.01","memory":"10Mi"}},"rolloutAnnotations":{},"rolloutLabels":{},"secret":{"data":{},"enabled":false},"server":{"deployment":{"image":"","image_tag":"1-95af053"}},"service":{"annotations":{},"loadBalancerSourceRanges":[],"type":"ClusterIP"},"serviceAccount":{"annotations":{},"create":false,"name":""},"servicemonitor":{"additionalLabels":{}},"tolerations":[],"topologySpreadConstraints":[],"volumeMounts":[],"volumes":[],"waitForSecondsBeforeScalingDown":30}`
		variables, err := templateParser.ExtractVariables(sampleTemplate)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(variables))
		assert.Equal(t, "envoyPort", variables[0])
		assert.Equal(t, "idleTimeoutVar", variables[1])
		assert.Equal(t, "idleTimeoutDivVar", variables[2])
	})
}

func TestVariableTemplateParserImpl_ParseTemplate(t *testing.T) {
	logger, err := util.NewSugardLogger()
	assert.Nil(t, err)
	t.Run("parse template", func(t *testing.T) {
		templateParser := NewVariableTemplateParserImpl(logger)
		sampleTemplate := `{"ConfigMaps":{"enabled":false,"maps":[]},"ConfigSecrets":{"enabled":false,"secrets":[]},"ContainerPort":[{"envoyPort":@{{envoyPort}},"idleTimeout":"@{{idleTimeoutVar / idleTimeoutDivVar}}s","name":"@{{1 + appNameVar}}","port":8080,"servicePort":@{{1 + idleTimeoutDivVar}}s,"supportStreaming":false,"useHTTP2":false}],"EnvVariables":[],"EnvVariablesFromFieldPath":[{"fieldPath":"metadata.name","name":"POD_NAME"}],"GracePeriod":30,"LivenessProbe":{"Path":"","command":[],"failureThreshold":3,"httpHeaders":[],"initialDelaySeconds":20,"periodSeconds":10,"port":8080,"scheme":"","successThreshold":1,"tcp":false,"timeoutSeconds":5},"MaxSurge":1,"MaxUnavailable":0,"MinReadySeconds":60,"ReadinessProbe":{"Path":"","command":[],"failureThreshold":3,"httpHeaders":[],"initialDelaySeconds":20,"periodSeconds":10,"port":8080,"scheme":"","successThreshold":1,"tcp":false,"timeoutSeconds":5},"Spec":{"Affinity":{"Values":"nodes","key":""}},"ambassadorMapping":{"ambassadorId":"","cors":{},"enabled":false,"hostname":"devtron.example.com","labels":{},"prefix":"/","retryPolicy":{},"rewrite":"","tls":{"context":"","create":false,"hosts":[],"secretName":""}},"args":{"enabled":false,"value":["/bin/sh","-c","touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600"]},"autoPromotionSeconds":30,"autoscaling":{"MaxReplicas":2,"MinReplicas":1,"TargetCPUUtilizationPercentage":90,"TargetMemoryUtilizationPercentage":80,"annotations":{},"behavior":{},"enabled":false,"extraMetrics":[],"labels":{}},"command":{"enabled":false,"value":[],"workingDir":{}},"containerExtraSpecs":{},"containerSecurityContext":{},"containerSpec":{"lifecycle":{"enabled":false,"postStart":{"httpGet":{"host":"example.com","path":"/example","port":90}},"preStop":{"exec":{"command":["sleep","10"]}}}},"containers":[],"dbMigrationConfig":{"enabled":false},"envoyproxy":{"configMapName":"","image":"quay.io/devtron/envoy:v1.14.1","lifecycle":{},"resources":{"limits":{"cpu":"50m","memory":"50Mi"},"requests":{"cpu":"50m","memory":"50Mi"}}},"hostAliases":[],"image":{"pullPolicy":"IfNotPresent"},"imagePullSecrets":[],"ingress":{"annotations":{},"className":"","enabled":false,"hosts":[{"host":"chart-example1.local","pathType":"ImplementationSpecific","paths":["/example1"]},{"host":"chart-example2.local","pathType":"ImplementationSpecific","paths":["/example2","/example2/healthz"]}],"labels":{},"tls":[]},"ingressInternal":{"annotations":{},"className":"","enabled":false,"hosts":[{"host":"chart-example1.internal","pathType":"ImplementationSpecific","paths":["/example1"]},{"host":"chart-example2.internal","pathType":"ImplementationSpecific","paths":["/example2","/example2/healthz"]}],"tls":[]},"initContainers":[],"istio":{"enable":false,"gateway":{"annotations":{},"enabled":false,"host":"example.com","labels":{},"tls":{"enabled":false,"secretName":"secret-name"}},"virtualService":{"annotations":{},"enabled":false,"gateways":[],"hosts":[],"http":[{"corsPolicy":{},"headers":{},"match":[{"uri":{"prefix":"/v1"}},{"uri":{"prefix":"/v2"}}],"retries":{"attempts":2,"perTryTimeout":"3s"},"rewriteUri":"/","route":[{"destination":{"host":"service1","port":80}}],"timeout":"12s"},{"route":[{"destination":{"host":"service2"}}]}],"labels":{}}},"kedaAutoscaling":{"advanced":{},"authenticationRef":{},"cooldownPeriod":300,"enabled":false,"envSourceContainerName":"","fallback":{},"idleReplicaCount":0,"maxReplicaCount":2,"minReplicaCount":1,"pollingInterval":30,"triggerAuthentication":{"enabled":false,"name":"","spec":{}},"triggers":[]},"nodeSelector":{},"orchestrator.deploymant.algo":1,"pauseForSecondsBeforeSwitchActive":30,"podAnnotations":{},"podDisruptionBudget":{},"podExtraSpecs":{},"podLabels":{},"podSecurityContext":{},"prometheus":{"release":"monitoring"},"prometheusRule":{"additionalLabels":{},"enabled":false,"namespace":""},"rawYaml":[],"replicaCount":1,"resources":{"limits":{"cpu":"0.05","memory":"50Mi"},"requests":{"cpu":"0.01","memory":"10Mi"}},"rolloutAnnotations":{},"rolloutLabels":{},"secret":{"data":{},"enabled":false},"server":{"deployment":{"image":"","image_tag":"1-95af053"}},"service":{"annotations":{},"loadBalancerSourceRanges":[],"type":"ClusterIP"},"serviceAccount":{"annotations":{},"create":false,"name":""},"servicemonitor":{"additionalLabels":{}},"tolerations":[],"topologySpreadConstraints":[],"volumeMounts":[],"volumes":[],"waitForSecondsBeforeScalingDown":30}`
		varValueMap := make(map[string]string, 0)
		//varValueMap["envoyPort"] = "1800"
		varValueMap["idleTimeoutVar"] = "600"
		varValueMap["idleTimeoutDivVar"] = "10"
		parsedTemplate := templateParser.ParseTemplate(sampleTemplate, varValueMap)
		fmt.Println("parsed template", parsedTemplate)
	})
}

func TestVariableTemplateParserImpl_ParseTemplateWithAddition(t *testing.T) {
	logger, err := util.NewSugardLogger()
	assert.Nil(t, err)
	t.Run("parse template", func(t *testing.T) {
		templateParser := NewVariableTemplateParserImpl(logger)
		sampleTemplate := TestData
		varValueMap := make(map[string]string, 0)
		varValueMap["container-port-number-new"] = "1800"
		parsedTemplate := templateParser.ParseTemplate(sampleTemplate, varValueMap)
		fmt.Println("parsed template", parsedTemplate)
	})
}

const TestData = `{
"ContainerPort": [
{
"envoyPort": 9000,
"idleTimeout": "1800s",
"name": "subs-app",
"port": "@{{container-port-number-new+1}}",
"servicePort": 80,
"supportStreaming": false,
"useHTTP2": false
}
],
"EnvVariables": [],
"GracePeriod": 30,
"LivenessProbe": {
"Path": "",
"command": [],
"failureThreshold": 3,
"httpHeaders": [],
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
"httpHeaders": [],
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
"StartupProbe": {
"Path": "",
"command": [],
"failureThreshold": 3,
"httpHeaders": [],
"initialDelaySeconds": 20,
"periodSeconds": 10,
"port": 8080,
"successThreshold": 1,
"tcp": false,
"timeoutSeconds": 5
},
"ambassadorMapping": {
"ambassadorId": "",
"cors": {},
"enabled": false,
"hostname": "devtron.example.com",
"labels": {},
"prefix": "/",
"retryPolicy": {},
"rewrite": "",
"tls": {
"context": "",
"create": false,
"hosts": [],
"secretName": ""
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
"annotations": {},
"behavior": {},
"enabled": false,
"extraMetrics": [],
"labels": {}
},
"command": {
"enabled": false,
"value": [],
"workingDir": {}
},
"containerSecurityContext": {},
"containerSpec": {
"lifecycle": {
"enabled": false,
"postStart": {
"httpGet": {
"host": "example.com",
"path": "/example",
"port": 90
}
},
"preStop": {
"exec": {
"command": [
"sleep",
"10"
]
}
}
}
},
"containers": [],
"dbMigrationConfig": {
"enabled": false
},
"envoyproxy": {
"configMapName": "",
"image": "docker.io/envoyproxy/envoy:v1.16.0",
"lifecycle": {},
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
"flaggerCanary": {
"addOtherGateways": [],
"addOtherHosts": [],
"analysis": {
"interval": "15s",
"maxWeight": 50,
"stepWeight": 5,
"threshold": 5
},
"annotations": {},
"appProtocol": "http",
"corsPolicy": null,
"createIstioGateway": {
"annotations": {},
"enabled": false,
"host": null,
"labels": {},
"tls": {
"enabled": false,
"secretName": null
}
},
"enabled": false,
"gatewayRefs": null,
"headers": null,
"labels": {},
"loadtest": {
"enabled": true,
"url": "http://flagger-loadtester.istio-system/"
},
"match": [
{
"uri": {
"prefix": "/"
}
}
],
"portDiscovery": true,
"retries": null,
"rewriteUri": "/",
"serviceport": 8080,
"targetPort": 8080,
"thresholds": {
"latency": 500,
"successRate": 90
},
"timeout": null
},
"hostAliases": [],
"image": {
"pullPolicy": "IfNotPresent"
},
"imagePullSecrets": [],
"ingress": {
"annotations": {},
"className": "",
"enabled": false,
"hosts": [
{
"host": "chart-example1.local",
"pathType": "ImplementationSpecific",
"paths": [
"/example1"
]
}
],
"labels": {},
"tls": []
},
"ingressInternal": {
"annotations": {},
"className": "",
"enabled": false,
"hosts": [
{
"host": "chart-example1.internal",
"pathType": "ImplementationSpecific",
"paths": [
"/example1"
]
},
{
"host": "chart-example2.internal",
"pathType": "ImplementationSpecific",
"paths": [
"/example2",
"/example2/healthz"
]
}
],
"tls": []
},
"initContainers": [],
"istio": {
"authorizationPolicy": {
"action": null,
"annotations": {},
"enabled": false,
"labels": {},
"provider": {},
"rules": []
},
"destinationRule": {
"annotations": {},
"enabled": false,
"labels": {},
"subsets": [],
"trafficPolicy": {}
},
"enable": false,
"gateway": {
"annotations": {},
"enabled": false,
"host": "example.com",
"labels": {},
"tls": {
"enabled": false,
"secretName": "example-secret"
}
},
"peerAuthentication": {
"annotations": {},
"enabled": false,
"labels": {},
"mtls": {
"mode": null
},
"portLevelMtls": {},
"selector": {
"enabled": false
}
},
"requestAuthentication": {
"annotations": {},
"enabled": false,
"jwtRules": [],
"labels": {},
"selector": {
"enabled": false
}
},
"virtualService": {
"annotations": {},
"enabled": false,
"gateways": [],
"hosts": [],
"http": [],
"labels": {}
}
},
"kedaAutoscaling": {
"advanced": {},
"authenticationRef": {},
"enabled": false,
"envSourceContainerName": "",
"maxReplicaCount": 2,
"minReplicaCount": 1,
"triggerAuthentication": {
"enabled": false,
"name": "",
"spec": {}
},
"triggers": []
},
"networkPolicy": {
"annotations": {},
"egress": [],
"enabled": false,
"ingress": [],
"labels": {},
"podSelector": {
"matchExpressions": [],
"matchLabels": {}
},
"policyTypes": []
},
"pauseForSecondsBeforeSwitchActive": 30,
"podAnnotations": {},
"podLabels": {},
"podSecurityContext": {},
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
"restartPolicy": "Always",
"rolloutAnnotations": {},
"rolloutLabels": {},
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
"loadBalancerSourceRanges": [],
"type": "ClusterIP"
},
"serviceAccount": {
"annotations": {},
"create": false,
"name": ""
},
"servicemonitor": {
"additionalLabels": {}
},
"tolerations": [],
"topologySpreadConstraints": [],
"volumeMounts": [],
"volumes": [],
"waitForSecondsBeforeScalingDown": 30,
"winterSoldier": {
"action": "sleep",
"annotation": {},
"apiVersion": "pincher.devtron.ai/v1alpha1",
"enabled": false,
"fieldSelector": [
"AfterTime(AddTime(ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'), '5m'), Now())"
],
"labels": {},
"targetReplicas": [],
"timeRangesWithZone": {
"timeRanges": [],
"timeZone": "Asia/Kolkata"
},
"type": "Deployment"
}
}`
