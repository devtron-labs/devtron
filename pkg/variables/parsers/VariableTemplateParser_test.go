package parsers

import (
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/variables/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

const JsonWithIntParam = `{"key1":[{"key2":"@{{container-port-number-new + 0}}"}]}`
const JsonWithIntParamResolvedTemplate = `{"key1":[{"key2":1800}]}`
const StringTemplate = "\"@{{Variable1}}\""
const StringTemplateResolved = "\"123\""
const StringTemplateWithIntParam = "- EXTERNAL_CI_ID: \"1\"\n  REPO_NAME_EXTERNAL_CI: @{{container-port-number-new}}\n  REGISTRY_URL_EXTERNAL_CI: docker.io/shivamnagar409\n- EXTERNAL_CI_ID: \"2\"\n  REPO_NAME_EXTERNAL_CI: test123\n  REGISTRY_URL_EXTERNAL_CI: docker.io/shivamnagar409\n"
const StringTemplateWithIntParamResolvedTemplate = "- EXTERNAL_CI_ID: \"1\"\n  REPO_NAME_EXTERNAL_CI: 1800\n  REGISTRY_URL_EXTERNAL_CI: docker.io/shivamnagar409\n- EXTERNAL_CI_ID: \"2\"\n  REPO_NAME_EXTERNAL_CI: test123\n  REGISTRY_URL_EXTERNAL_CI: docker.io/shivamnagar409\n"

func TestVariableTemplateParserImpl_ExtractVariables(t *testing.T) {
	logger, err := util.NewSugardLogger()
	assert.Nil(t, err)

	t.Run("extract variables", func(t *testing.T) {
		templateParser := NewVariableTemplateParserImpl(logger) // \"value\"
		sampleTemplate := `{"ConfigMaps":{"enabled":false,"maps":[]},"ConfigSecrets":{"enabled":false,"secrets":[]},"ContainerPort":[{"envoyPort":"@{{envoyPort + 0}}","idleTimeout":"@{{idleTimeoutVar / idleTimeoutDivVar}}s","name":"${1 + appName}","port":8080,"servicePort":80,"supportStreaming":false,"useHTTP2":false}],"EnvVariables":[],"EnvVariablesFromFieldPath":[{"fieldPath":"metadata.name","name":"POD_NAME"}],"GracePeriod":30,"LivenessProbe":{"Path":"","command":[],"failureThreshold":3,"httpHeaders":[],"initialDelaySeconds":20,"periodSeconds":10,"port":8080,"scheme":"","successThreshold":1,"tcp":false,"timeoutSeconds":5},"MaxSurge":1,"MaxUnavailable":0,"MinReadySeconds":60,"ReadinessProbe":{"Path":"","command":[],"failureThreshold":3,"httpHeaders":[],"initialDelaySeconds":20,"periodSeconds":10,"port":8080,"scheme":"","successThreshold":1,"tcp":false,"timeoutSeconds":5},"Spec":{"Affinity":{"Values":"nodes","key":""}},"ambassadorMapping":{"ambassadorId":"","cors":{},"enabled":false,"hostname":"devtron.example.com","labels":{},"prefix":"/","retryPolicy":{},"rewrite":"","tls":{"context":"","create":false,"hosts":[],"secretName":""}},"args":{"enabled":false,"value":["/bin/sh","-c","touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600"]},"autoPromotionSeconds":30,"autoscaling":{"MaxReplicas":2,"MinReplicas":1,"TargetCPUUtilizationPercentage":90,"TargetMemoryUtilizationPercentage":80,"annotations":{},"behavior":{},"enabled":false,"extraMetrics":[],"labels":{}},"command":{"enabled":false,"value":[],"workingDir":{}},"containerExtraSpecs":{},"containerSecurityContext":{},"containerSpec":{"lifecycle":{"enabled":false,"postStart":{"httpGet":{"host":"example.com","path":"/example","port":90}},"preStop":{"exec":{"command":["sleep","10"]}}}},"containers":[],"dbMigrationConfig":{"enabled":false},"envoyproxy":{"configMapName":"","image":"quay.io/devtron/envoy:v1.14.1","lifecycle":{},"resources":{"limits":{"cpu":"50m","memory":"50Mi"},"requests":{"cpu":"50m","memory":"50Mi"}}},"hostAliases":[],"image":{"pullPolicy":"IfNotPresent"},"imagePullSecrets":[],"ingress":{"annotations":{},"className":"","enabled":false,"hosts":[{"host":"chart-example1.local","pathType":"ImplementationSpecific","paths":["/example1"]},{"host":"chart-example2.local","pathType":"ImplementationSpecific","paths":["/example2","/example2/healthz"]}],"labels":{},"tls":[]},"ingressInternal":{"annotations":{},"className":"","enabled":false,"hosts":[{"host":"chart-example1.internal","pathType":"ImplementationSpecific","paths":["/example1"]},{"host":"chart-example2.internal","pathType":"ImplementationSpecific","paths":["/example2","/example2/healthz"]}],"tls":[]},"initContainers":[],"istio":{"enable":false,"gateway":{"annotations":{},"enabled":false,"host":"example.com","labels":{},"tls":{"enabled":false,"secretName":"secret-name"}},"virtualService":{"annotations":{},"enabled":false,"gateways":[],"hosts":[],"http":[{"corsPolicy":{},"headers":{},"match":[{"uri":{"prefix":"/v1"}},{"uri":{"prefix":"/v2"}}],"retries":{"attempts":2,"perTryTimeout":"3s"},"rewriteUri":"/","route":[{"destination":{"host":"service1","port":80}}],"timeout":"12s"},{"route":[{"destination":{"host":"service2"}}]}],"labels":{}}},"kedaAutoscaling":{"advanced":{},"authenticationRef":{},"cooldownPeriod":300,"enabled":false,"envSourceContainerName":"","fallback":{},"idleReplicaCount":0,"maxReplicaCount":2,"minReplicaCount":1,"pollingInterval":30,"triggerAuthentication":{"enabled":false,"name":"","spec":{}},"triggers":[]},"nodeSelector":{},"orchestrator.deploymant.algo":1,"pauseForSecondsBeforeSwitchActive":30,"podAnnotations":{},"podDisruptionBudget":{},"podExtraSpecs":{},"podLabels":{},"podSecurityContext":{},"prometheus":{"release":"monitoring"},"prometheusRule":{"additionalLabels":{},"enabled":false,"namespace":""},"rawYaml":[],"replicaCount":1,"resources":{"limits":{"cpu":"0.05","memory":"50Mi"},"requests":{"cpu":"0.01","memory":"10Mi"}},"rolloutAnnotations":{},"rolloutLabels":{},"secret":{"data":{},"enabled":false},"server":{"deployment":{"image":"","image_tag":"1-95af053"}},"service":{"annotations":{},"loadBalancerSourceRanges":[],"type":"ClusterIP"},"serviceAccount":{"annotations":{},"create":false,"name":""},"servicemonitor":{"additionalLabels":{}},"tolerations":[],"topologySpreadConstraints":[],"volumeMounts":[],"volumes":[],"waitForSecondsBeforeScalingDown":30}`
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
	templateParser := NewVariableTemplateParserImpl(logger)
	t.Run("parse template", func(t *testing.T) {
		scopedVariables := []*models.ScopedVariableData{{VariableName: "container-port-number-new", VariableValue: models.VariableValue{Value: "1800"}}}
		parserResponse := templateParser.ParseTemplate(VariableParserRequest{TemplateType: JsonVariableTemplate, Template: JsonWithIntParam, Variables: scopedVariables})
		parsedTemplate := parserResponse.ResolvedTemplate
		assert.Equal(t, JsonWithIntParamResolvedTemplate, parsedTemplate)
	})

	t.Run("parse stringify template", func(t *testing.T) {
		scopedVariables := []*models.ScopedVariableData{{VariableName: "Variable1", VariableValue: models.VariableValue{Value: "123"}}}
		parserResponse := templateParser.ParseTemplate(VariableParserRequest{TemplateType: StringVariableTemplate, Template: StringTemplate, Variables: scopedVariables})
		err = parserResponse.Error
		assert.Nil(t, err)
		assert.Equal(t, StringTemplateResolved, parserResponse.ResolvedTemplate)
	})

	t.Run("parse stringify template with int value", func(t *testing.T) {
		scopedVariables := []*models.ScopedVariableData{{VariableName: "container-port-number-new", VariableValue: models.VariableValue{Value: "1800"}}}
		parserResponse := templateParser.ParseTemplate(VariableParserRequest{TemplateType: StringVariableTemplate, Template: StringTemplateWithIntParam, Variables: scopedVariables})
		err = parserResponse.Error
		assert.Nil(t, err)
		assert.Equal(t, StringTemplateWithIntParamResolvedTemplate, parserResponse.ResolvedTemplate)
	})
}
