package generateManifest

import (
	"context"
	client2 "github.com/devtron-labs/authenticator/client"
	"github.com/devtron-labs/devtron/api/bean"
	mocks4 "github.com/devtron-labs/devtron/api/helm-app/mocks"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	mocks3 "github.com/devtron-labs/devtron/internal/sql/repository/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	mocks6 "github.com/devtron-labs/devtron/internal/util/mocks"
	mocks2 "github.com/devtron-labs/devtron/pkg/app/mocks"
	"github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/chart/mocks"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	mocks5 "github.com/devtron-labs/devtron/pkg/chartRepo/repository/mocks"
	"github.com/devtron-labs/devtron/util/k8s"
	"reflect"
	"testing"
)

var K8sUtilObj *k8s.K8sUtil

func TestDeploymentTemplateServiceImpl_FetchDeploymentsWithChartRefs(t *testing.T) {
	defaultVersions := &chart.ChartRefResponse{
		ChartRefs: []chart.ChartRef{
			{
				Id:                    1,
				Version:               "v1.0.1",
				Name:                  "Deployment",
				Description:           "This is a deployment chart",
				UserUploaded:          false,
				IsAppMetricsSupported: false,
			},
			{
				Id:                    2,
				Version:               "v1.0.2",
				Name:                  "Deployment",
				Description:           "This is a deployment chart",
				UserUploaded:          false,
				IsAppMetricsSupported: false,
			},
			{
				Id:                    3,
				Version:               "v1.0.3",
				Name:                  "Deployment",
				Description:           "This is a deployment chart",
				UserUploaded:          false,
				IsAppMetricsSupported: false,
			},
		},
		LatestAppChartRef: 2,
		LatestEnvChartRef: 2,
	}
	publishedOnEnvs := []*bean.Environment{
		{
			ChartRefId:      2,
			EnvironmentId:   1,
			EnvironmentName: "devtron-demo",
		},
	}

	deployedOnEnv := []*repository.DeploymentTemplateComparisonMetadata{
		{
			ChartId:                  1,
			ChartVersion:             "4.18.1",
			EnvironmentId:            1,
			PipelineConfigOverrideId: 5,
			//StartedOn: 2023-08-26T16:36:55.732551Z,
			//FinishedOn: 2023-08-26T16:40:00.174576Z,
			Status: "Succeeded",
		},
		{
			ChartId:                  1,
			ChartVersion:             "4.18.1",
			EnvironmentId:            1,
			PipelineConfigOverrideId: 5,
			//StartedOn: 2023-08-26T16:36:55.732551Z,
			//FinishedOn: 2023-08-26T16:40:00.174576Z,
			Status: "Succeeded",
		},
		{
			ChartId:                  1,
			ChartVersion:             "4.18.1",
			EnvironmentId:            1,
			PipelineConfigOverrideId: 5,
			//StartedOn: 2023-08-26T16:36:55.732551Z,
			//FinishedOn: 2023-08-26T16:40:00.174576Z,
			Status: "Succeeded",
		},
	}

	deployedOnOtherEnvs := []*repository.DeploymentTemplateComparisonMetadata{
		{
			ChartId:                  1,
			ChartVersion:             "4.18.1",
			EnvironmentId:            2,
			PipelineConfigOverrideId: 9,
		},
	}

	type args struct {
		appId int
		envId int
	}
	tests := []struct {
		name    string
		args    args
		want    []*repository.DeploymentTemplateComparisonMetadata
		wantErr bool
	}{

		{
			name: "test for successfully fetching the list",
			args: args{
				appId: 1,
				envId: 1,
			},
			want: []*repository.DeploymentTemplateComparisonMetadata{
				{
					ChartId:                  1,
					ChartVersion:             "v1.0.1",
					ChartType:                "Deployment",
					EnvironmentId:            0,
					EnvironmentName:          "",
					PipelineConfigOverrideId: 0,
					StartedOn:                nil,
					FinishedOn:               nil,
					Status:                   "",
					Type:                     1,
				},
				{
					ChartId:                  2,
					ChartVersion:             "v1.0.2",
					ChartType:                "Deployment",
					EnvironmentId:            0,
					EnvironmentName:          "",
					PipelineConfigOverrideId: 0,
					StartedOn:                nil,
					FinishedOn:               nil,
					Status:                   "",
					Type:                     1,
				}, {
					ChartId:                  3,
					ChartVersion:             "v1.0.3",
					ChartType:                "Deployment",
					EnvironmentId:            0,
					EnvironmentName:          "",
					PipelineConfigOverrideId: 0,
					StartedOn:                nil,
					FinishedOn:               nil,
					Status:                   "",
					Type:                     1,
				}, {
					ChartId:                  2,
					ChartVersion:             "",
					ChartType:                "",
					EnvironmentId:            1,
					EnvironmentName:          "devtron-demo",
					PipelineConfigOverrideId: 0,
					StartedOn:                nil,
					FinishedOn:               nil,
					Status:                   "",
					Type:                     2,
				}, {
					ChartId:                  1,
					ChartVersion:             "4.18.1",
					ChartType:                "",
					EnvironmentId:            1,
					EnvironmentName:          "",
					PipelineConfigOverrideId: 5,
					StartedOn:                nil,
					FinishedOn:               nil,
					Status:                   "Succeeded",
					Type:                     3,
				}, {
					ChartId:                  1,
					ChartVersion:             "4.18.1",
					ChartType:                "",
					EnvironmentId:            1,
					EnvironmentName:          "",
					PipelineConfigOverrideId: 5,
					StartedOn:                nil,
					FinishedOn:               nil,
					Status:                   "Succeeded",
					Type:                     3,
				}, {
					ChartId:                  1,
					ChartVersion:             "4.18.1",
					ChartType:                "",
					EnvironmentId:            1,
					EnvironmentName:          "",
					PipelineConfigOverrideId: 5,
					StartedOn:                nil,
					FinishedOn:               nil,
					Status:                   "Succeeded",
					Type:                     3,
				}, {
					ChartId:                  1,
					ChartVersion:             "v1.0.1",
					ChartType:                "Deployment",
					EnvironmentId:            2,
					EnvironmentName:          "",
					PipelineConfigOverrideId: 9,
					StartedOn:                nil,
					FinishedOn:               nil,
					Status:                   "",
					Type:                     4,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			impl, chartService, appListingService, deploymentTemplateRepository, _ := InitEventSimpleFactoryImpl(t)
			chartService.On("ChartRefAutocompleteForAppOrEnv", tt.args.appId, 0).Return(defaultVersions, nil)
			appListingService.On("FetchMinDetailOtherEnvironment", tt.args.appId).Return(publishedOnEnvs, nil)
			deploymentTemplateRepository.On("FetchDeploymentHistoryWithChartRefs", tt.args.appId, tt.args.envId).Return(deployedOnEnv, nil)
			deploymentTemplateRepository.On("FetchLatestDeploymentWithChartRefs", tt.args.appId, tt.args.envId).Return(deployedOnOtherEnvs, nil)
			got, err := impl.FetchDeploymentsWithChartRefs(tt.args.appId, tt.args.envId)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchDeploymentsWithChartRefs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchDeploymentsWithChartRefs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeploymentTemplateServiceImpl_GetDeploymentTemplate(t *testing.T) {

	var myMap = make(map[string]interface{})
	myString := "{\\\"ContainerPort\\\":[{\\\"envoyPort\\\":6969,\\\"idleTimeout\\\":\\\"6969s\\\",\\\"name\\\":\\\"app\\\",\\\"port\\\":6969,\\\"servicePort\\\":69,\\\"supportStreaming\\\":false,\\\"useHTTP2\\\":false}],\\\"EnvVariables\\\":[],\\\"GracePeriod\\\":30,\\\"LivenessProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"scheme\\\":\\\"\\\",\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"MaxSurge\\\":1,\\\"MaxUnavailable\\\":0,\\\"MinReadySeconds\\\":60,\\\"ReadinessProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"scheme\\\":\\\"\\\",\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"Spec\\\":{\\\"Affinity\\\":{\\\"Key\\\":null,\\\"Values\\\":\\\"nodes\\\",\\\"key\\\":\\\"\\\"}},\\\"StartupProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"ambassadorMapping\\\":{\\\"ambassadorId\\\":\\\"\\\",\\\"cors\\\":{},\\\"enabled\\\":false,\\\"hostname\\\":\\\"devtron.example.com\\\",\\\"labels\\\":{},\\\"prefix\\\":\\\"/\\\",\\\"retryPolicy\\\":{},\\\"rewrite\\\":\\\"\\\",\\\"tls\\\":{\\\"context\\\":\\\"\\\",\\\"create\\\":false,\\\"hosts\\\":[],\\\"secretName\\\":\\\"\\\"}},\\\"args\\\":{\\\"enabled\\\":false,\\\"value\\\":[\\\"/bin/sh\\\",\\\"-c\\\",\\\"touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600\\\"]},\\\"autoscaling\\\":{\\\"MaxReplicas\\\":2,\\\"MinReplicas\\\":1,\\\"TargetCPUUtilizationPercentage\\\":90,\\\"TargetMemoryUtilizationPercentage\\\":69,\\\"annotations\\\":{},\\\"behavior\\\":{},\\\"enabled\\\":false,\\\"extraMetrics\\\":[],\\\"labels\\\":{}},\\\"command\\\":{\\\"enabled\\\":false,\\\"value\\\":[],\\\"workingDir\\\":{}},\\\"containerSecurityContext\\\":{},\\\"containerSpec\\\":{\\\"lifecycle\\\":{\\\"enabled\\\":false,\\\"postStart\\\":{\\\"httpGet\\\":{\\\"host\\\":\\\"example.com\\\",\\\"path\\\":\\\"/example\\\",\\\"port\\\":90}},\\\"preStop\\\":{\\\"exec\\\":{\\\"command\\\":[\\\"sleep\\\",\\\"10\\\"]}}}},\\\"containers\\\":[],\\\"dbMigrationConfig\\\":{\\\"enabled\\\":false},\\\"envoyproxy\\\":{\\\"configMapName\\\":\\\"\\\",\\\"image\\\":\\\"docker.io/envoyproxy/envoy:v1.16.0\\\",\\\"lifecycle\\\":{},\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"50m\\\",\\\"memory\\\":\\\"50Mi\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"50m\\\",\\\"memory\\\":\\\"50Mi\\\"}}},\\\"hostAliases\\\":[],\\\"image\\\":{\\\"pullPolicy\\\":\\\"IfNotPresent\\\"},\\\"imagePullSecrets\\\":[],\\\"ingress\\\":{\\\"annotations\\\":{},\\\"className\\\":\\\"\\\",\\\"enabled\\\":false,\\\"hosts\\\":[{\\\"host\\\":\\\"chart-example1.local\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example1\\\"]},{\\\"host\\\":\\\"chart-example2.local\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example2\\\",\\\"/example2/healthz\\\"]}],\\\"labels\\\":{},\\\"tls\\\":[]},\\\"ingressInternal\\\":{\\\"annotations\\\":{},\\\"className\\\":\\\"\\\",\\\"enabled\\\":false,\\\"hosts\\\":[{\\\"host\\\":\\\"chart-example1.internal\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example1\\\"]},{\\\"host\\\":\\\"chart-example2.internal\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example2\\\",\\\"/example2/healthz\\\"]}],\\\"tls\\\":[]},\\\"initContainers\\\":[],\\\"istio\\\":{\\\"authorizationPolicy\\\":{\\\"action\\\":null,\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"provider\\\":{},\\\"rules\\\":[]},\\\"destinationRule\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"subsets\\\":[],\\\"trafficPolicy\\\":{}},\\\"enable\\\":false,\\\"gateway\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"host\\\":\\\"example.com\\\",\\\"labels\\\":{},\\\"tls\\\":{\\\"enabled\\\":false,\\\"secretName\\\":\\\"secret-name\\\"}},\\\"peerAuthentication\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"mtls\\\":{\\\"mode\\\":null},\\\"portLevelMtls\\\":{},\\\"selector\\\":{\\\"enabled\\\":false}},\\\"requestAuthentication\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"jwtRules\\\":[],\\\"labels\\\":{},\\\"selector\\\":{\\\"enabled\\\":false}},\\\"virtualService\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"gateways\\\":[],\\\"hosts\\\":[],\\\"http\\\":[],\\\"labels\\\":{}}},\\\"kedaAutoscaling\\\":{\\\"advanced\\\":{},\\\"authenticationRef\\\":{},\\\"enabled\\\":false,\\\"envSourceContainerName\\\":\\\"\\\",\\\"maxReplicaCount\\\":2,\\\"minReplicaCount\\\":1,\\\"triggerAuthentication\\\":{\\\"enabled\\\":false,\\\"name\\\":\\\"\\\",\\\"spec\\\":{}},\\\"triggers\\\":[]},\\\"networkPolicy\\\":{\\\"annotations\\\":{},\\\"egress\\\":[],\\\"enabled\\\":false,\\\"ingress\\\":[],\\\"labels\\\":{},\\\"podSelector\\\":{\\\"matchExpressions\\\":[],\\\"matchLabels\\\":{}},\\\"policyTypes\\\":[]},\\\"pauseForSecondsBeforeSwitchActive\\\":30,\\\"podAnnotations\\\":{},\\\"podDisruptionBudget\\\":{},\\\"podLabels\\\":{},\\\"podSecurityContext\\\":{},\\\"prometheus\\\":{\\\"release\\\":\\\"monitoring\\\"},\\\"rawYaml\\\":[],\\\"replicaCount\\\":1,\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"0.05\\\",\\\"memory\\\":\\\"50Mi\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"0.01\\\",\\\"memory\\\":\\\"10Mi\\\"}},\\\"restartPolicy\\\":\\\"Always\\\",\\\"rolloutAnnotations\\\":{},\\\"rolloutLabels\\\":{},\\\"secret\\\":{\\\"data\\\":{},\\\"enabled\\\":false},\\\"server\\\":{\\\"deployment\\\":{\\\"image\\\":\\\"\\\",\\\"image_tag\\\":\\\"1-95af053\\\"}},\\\"service\\\":{\\\"annotations\\\":{},\\\"loadBalancerSourceRanges\\\":[],\\\"type\\\":\\\"ClusterIP\\\"},\\\"serviceAccount\\\":{\\\"annotations\\\":{},\\\"create\\\":false,\\\"name\\\":\\\"\\\"},\\\"servicemonitor\\\":{\\\"additionalLabels\\\":{}},\\\"tolerations\\\":[],\\\"topologySpreadConstraints\\\":[],\\\"volumeMounts\\\":[],\\\"volumes\\\":[],\\\"waitForSecondsBeforeScalingDown\\\":30,\\\"winterSoldier\\\":{\\\"action\\\":\\\"sleep\\\",\\\"annotation\\\":{},\\\"apiVersion\\\":\\\"pincher.devtron.ai/v1alpha1\\\",\\\"enabled\\\":false,\\\"fieldSelector\\\":[\\\"AfterTime(AddTime(ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'), '5m'), Now())\\\"],\\\"labels\\\":{},\\\"targetReplicas\\\":[],\\\"timeRangesWithZone\\\":{\\\"timeRanges\\\":[],\\\"timeZone\\\":\\\"Asia/Kolkata\\\"},\\\"type\\\":\\\"Rollout\\\"}}"
	chart := &chartRepoRepository.Chart{}
	chart.GlobalOverride = "{\\\"ContainerPort\\\":[{\\\"envoyPort\\\":6969,\\\"idleTimeout\\\":\\\"6969s\\\",\\\"name\\\":\\\"app\\\",\\\"port\\\":6969,\\\"servicePort\\\":69,\\\"supportStreaming\\\":false,\\\"useHTTP2\\\":false}],\\\"EnvVariables\\\":[],\\\"GracePeriod\\\":30,\\\"LivenessProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"scheme\\\":\\\"\\\",\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"MaxSurge\\\":1,\\\"MaxUnavailable\\\":0,\\\"MinReadySeconds\\\":60,\\\"ReadinessProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"scheme\\\":\\\"\\\",\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"Spec\\\":{\\\"Affinity\\\":{\\\"Key\\\":null,\\\"Values\\\":\\\"nodes\\\",\\\"key\\\":\\\"\\\"}},\\\"StartupProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"ambassadorMapping\\\":{\\\"ambassadorId\\\":\\\"\\\",\\\"cors\\\":{},\\\"enabled\\\":false,\\\"hostname\\\":\\\"devtron.example.com\\\",\\\"labels\\\":{},\\\"prefix\\\":\\\"/\\\",\\\"retryPolicy\\\":{},\\\"rewrite\\\":\\\"\\\",\\\"tls\\\":{\\\"context\\\":\\\"\\\",\\\"create\\\":false,\\\"hosts\\\":[],\\\"secretName\\\":\\\"\\\"}},\\\"args\\\":{\\\"enabled\\\":false,\\\"value\\\":[\\\"/bin/sh\\\",\\\"-c\\\",\\\"touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600\\\"]},\\\"autoscaling\\\":{\\\"MaxReplicas\\\":2,\\\"MinReplicas\\\":1,\\\"TargetCPUUtilizationPercentage\\\":90,\\\"TargetMemoryUtilizationPercentage\\\":69,\\\"annotations\\\":{},\\\"behavior\\\":{},\\\"enabled\\\":false,\\\"extraMetrics\\\":[],\\\"labels\\\":{}},\\\"command\\\":{\\\"enabled\\\":false,\\\"value\\\":[],\\\"workingDir\\\":{}},\\\"containerSecurityContext\\\":{},\\\"containerSpec\\\":{\\\"lifecycle\\\":{\\\"enabled\\\":false,\\\"postStart\\\":{\\\"httpGet\\\":{\\\"host\\\":\\\"example.com\\\",\\\"path\\\":\\\"/example\\\",\\\"port\\\":90}},\\\"preStop\\\":{\\\"exec\\\":{\\\"command\\\":[\\\"sleep\\\",\\\"10\\\"]}}}},\\\"containers\\\":[],\\\"dbMigrationConfig\\\":{\\\"enabled\\\":false},\\\"envoyproxy\\\":{\\\"configMapName\\\":\\\"\\\",\\\"image\\\":\\\"docker.io/envoyproxy/envoy:v1.16.0\\\",\\\"lifecycle\\\":{},\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"50m\\\",\\\"memory\\\":\\\"50Mi\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"50m\\\",\\\"memory\\\":\\\"50Mi\\\"}}},\\\"hostAliases\\\":[],\\\"image\\\":{\\\"pullPolicy\\\":\\\"IfNotPresent\\\"},\\\"imagePullSecrets\\\":[],\\\"ingress\\\":{\\\"annotations\\\":{},\\\"className\\\":\\\"\\\",\\\"enabled\\\":false,\\\"hosts\\\":[{\\\"host\\\":\\\"chart-example1.local\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example1\\\"]},{\\\"host\\\":\\\"chart-example2.local\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example2\\\",\\\"/example2/healthz\\\"]}],\\\"labels\\\":{},\\\"tls\\\":[]},\\\"ingressInternal\\\":{\\\"annotations\\\":{},\\\"className\\\":\\\"\\\",\\\"enabled\\\":false,\\\"hosts\\\":[{\\\"host\\\":\\\"chart-example1.internal\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example1\\\"]},{\\\"host\\\":\\\"chart-example2.internal\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example2\\\",\\\"/example2/healthz\\\"]}],\\\"tls\\\":[]},\\\"initContainers\\\":[],\\\"istio\\\":{\\\"authorizationPolicy\\\":{\\\"action\\\":null,\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"provider\\\":{},\\\"rules\\\":[]},\\\"destinationRule\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"subsets\\\":[],\\\"trafficPolicy\\\":{}},\\\"enable\\\":false,\\\"gateway\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"host\\\":\\\"example.com\\\",\\\"labels\\\":{},\\\"tls\\\":{\\\"enabled\\\":false,\\\"secretName\\\":\\\"secret-name\\\"}},\\\"peerAuthentication\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"mtls\\\":{\\\"mode\\\":null},\\\"portLevelMtls\\\":{},\\\"selector\\\":{\\\"enabled\\\":false}},\\\"requestAuthentication\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"jwtRules\\\":[],\\\"labels\\\":{},\\\"selector\\\":{\\\"enabled\\\":false}},\\\"virtualService\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"gateways\\\":[],\\\"hosts\\\":[],\\\"http\\\":[],\\\"labels\\\":{}}},\\\"kedaAutoscaling\\\":{\\\"advanced\\\":{},\\\"authenticationRef\\\":{},\\\"enabled\\\":false,\\\"envSourceContainerName\\\":\\\"\\\",\\\"maxReplicaCount\\\":2,\\\"minReplicaCount\\\":1,\\\"triggerAuthentication\\\":{\\\"enabled\\\":false,\\\"name\\\":\\\"\\\",\\\"spec\\\":{}},\\\"triggers\\\":[]},\\\"networkPolicy\\\":{\\\"annotations\\\":{},\\\"egress\\\":[],\\\"enabled\\\":false,\\\"ingress\\\":[],\\\"labels\\\":{},\\\"podSelector\\\":{\\\"matchExpressions\\\":[],\\\"matchLabels\\\":{}},\\\"policyTypes\\\":[]},\\\"pauseForSecondsBeforeSwitchActive\\\":30,\\\"podAnnotations\\\":{},\\\"podDisruptionBudget\\\":{},\\\"podLabels\\\":{},\\\"podSecurityContext\\\":{},\\\"prometheus\\\":{\\\"release\\\":\\\"monitoring\\\"},\\\"rawYaml\\\":[],\\\"replicaCount\\\":1,\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"0.05\\\",\\\"memory\\\":\\\"50Mi\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"0.01\\\",\\\"memory\\\":\\\"10Mi\\\"}},\\\"restartPolicy\\\":\\\"Always\\\",\\\"rolloutAnnotations\\\":{},\\\"rolloutLabels\\\":{},\\\"secret\\\":{\\\"data\\\":{},\\\"enabled\\\":false},\\\"server\\\":{\\\"deployment\\\":{\\\"image\\\":\\\"\\\",\\\"image_tag\\\":\\\"1-95af053\\\"}},\\\"service\\\":{\\\"annotations\\\":{},\\\"loadBalancerSourceRanges\\\":[],\\\"type\\\":\\\"ClusterIP\\\"},\\\"serviceAccount\\\":{\\\"annotations\\\":{},\\\"create\\\":false,\\\"name\\\":\\\"\\\"},\\\"servicemonitor\\\":{\\\"additionalLabels\\\":{}},\\\"tolerations\\\":[],\\\"topologySpreadConstraints\\\":[],\\\"volumeMounts\\\":[],\\\"volumes\\\":[],\\\"waitForSecondsBeforeScalingDown\\\":30,\\\"winterSoldier\\\":{\\\"action\\\":\\\"sleep\\\",\\\"annotation\\\":{},\\\"apiVersion\\\":\\\"pincher.devtron.ai/v1alpha1\\\",\\\"enabled\\\":false,\\\"fieldSelector\\\":[\\\"AfterTime(AddTime(ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'), '5m'), Now())\\\"],\\\"labels\\\":{},\\\"targetReplicas\\\":[],\\\"timeRangesWithZone\\\":{\\\"timeRanges\\\":[],\\\"timeZone\\\":\\\"Asia/Kolkata\\\"},\\\"type\\\":\\\"Rollout\\\"}}"

	type args struct {
		ctx     context.Context
		request DeploymentTemplateRequest
	}
	tests := []struct {
		name string

		args    args
		want    DeploymentTemplateResponse
		wantErr bool
	}{
		{
			name: "get values same as that of request",
			args: args{
				ctx: context.Background(),
				request: DeploymentTemplateRequest{
					Values:    "{\\\"ContainerPort\\\":[{\\\"envoyPort\\\":6969,\\\"idleTimeout\\\":\\\"6969s\\\",\\\"name\\\":\\\"app\\\",\\\"port\\\":6969,\\\"servicePort\\\":69,\\\"supportStreaming\\\":false,\\\"useHTTP2\\\":false}],\\\"EnvVariables\\\":[],\\\"GracePeriod\\\":30,\\\"LivenessProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"scheme\\\":\\\"\\\",\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"MaxSurge\\\":1,\\\"MaxUnavailable\\\":0,\\\"MinReadySeconds\\\":60,\\\"ReadinessProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"scheme\\\":\\\"\\\",\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"Spec\\\":{\\\"Affinity\\\":{\\\"Key\\\":null,\\\"Values\\\":\\\"nodes\\\",\\\"key\\\":\\\"\\\"}},\\\"StartupProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"ambassadorMapping\\\":{\\\"ambassadorId\\\":\\\"\\\",\\\"cors\\\":{},\\\"enabled\\\":false,\\\"hostname\\\":\\\"devtron.example.com\\\",\\\"labels\\\":{},\\\"prefix\\\":\\\"/\\\",\\\"retryPolicy\\\":{},\\\"rewrite\\\":\\\"\\\",\\\"tls\\\":{\\\"context\\\":\\\"\\\",\\\"create\\\":false,\\\"hosts\\\":[],\\\"secretName\\\":\\\"\\\"}},\\\"args\\\":{\\\"enabled\\\":false,\\\"value\\\":[\\\"/bin/sh\\\",\\\"-c\\\",\\\"touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600\\\"]},\\\"autoscaling\\\":{\\\"MaxReplicas\\\":2,\\\"MinReplicas\\\":1,\\\"TargetCPUUtilizationPercentage\\\":90,\\\"TargetMemoryUtilizationPercentage\\\":69,\\\"annotations\\\":{},\\\"behavior\\\":{},\\\"enabled\\\":false,\\\"extraMetrics\\\":[],\\\"labels\\\":{}},\\\"command\\\":{\\\"enabled\\\":false,\\\"value\\\":[],\\\"workingDir\\\":{}},\\\"containerSecurityContext\\\":{},\\\"containerSpec\\\":{\\\"lifecycle\\\":{\\\"enabled\\\":false,\\\"postStart\\\":{\\\"httpGet\\\":{\\\"host\\\":\\\"example.com\\\",\\\"path\\\":\\\"/example\\\",\\\"port\\\":90}},\\\"preStop\\\":{\\\"exec\\\":{\\\"command\\\":[\\\"sleep\\\",\\\"10\\\"]}}}},\\\"containers\\\":[],\\\"dbMigrationConfig\\\":{\\\"enabled\\\":false},\\\"envoyproxy\\\":{\\\"configMapName\\\":\\\"\\\",\\\"image\\\":\\\"docker.io/envoyproxy/envoy:v1.16.0\\\",\\\"lifecycle\\\":{},\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"50m\\\",\\\"memory\\\":\\\"50Mi\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"50m\\\",\\\"memory\\\":\\\"50Mi\\\"}}},\\\"hostAliases\\\":[],\\\"image\\\":{\\\"pullPolicy\\\":\\\"IfNotPresent\\\"},\\\"imagePullSecrets\\\":[],\\\"ingress\\\":{\\\"annotations\\\":{},\\\"className\\\":\\\"\\\",\\\"enabled\\\":false,\\\"hosts\\\":[{\\\"host\\\":\\\"chart-example1.local\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example1\\\"]},{\\\"host\\\":\\\"chart-example2.local\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example2\\\",\\\"/example2/healthz\\\"]}],\\\"labels\\\":{},\\\"tls\\\":[]},\\\"ingressInternal\\\":{\\\"annotations\\\":{},\\\"className\\\":\\\"\\\",\\\"enabled\\\":false,\\\"hosts\\\":[{\\\"host\\\":\\\"chart-example1.internal\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example1\\\"]},{\\\"host\\\":\\\"chart-example2.internal\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example2\\\",\\\"/example2/healthz\\\"]}],\\\"tls\\\":[]},\\\"initContainers\\\":[],\\\"istio\\\":{\\\"authorizationPolicy\\\":{\\\"action\\\":null,\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"provider\\\":{},\\\"rules\\\":[]},\\\"destinationRule\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"subsets\\\":[],\\\"trafficPolicy\\\":{}},\\\"enable\\\":false,\\\"gateway\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"host\\\":\\\"example.com\\\",\\\"labels\\\":{},\\\"tls\\\":{\\\"enabled\\\":false,\\\"secretName\\\":\\\"secret-name\\\"}},\\\"peerAuthentication\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"mtls\\\":{\\\"mode\\\":null},\\\"portLevelMtls\\\":{},\\\"selector\\\":{\\\"enabled\\\":false}},\\\"requestAuthentication\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"jwtRules\\\":[],\\\"labels\\\":{},\\\"selector\\\":{\\\"enabled\\\":false}},\\\"virtualService\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"gateways\\\":[],\\\"hosts\\\":[],\\\"http\\\":[],\\\"labels\\\":{}}},\\\"kedaAutoscaling\\\":{\\\"advanced\\\":{},\\\"authenticationRef\\\":{},\\\"enabled\\\":false,\\\"envSourceContainerName\\\":\\\"\\\",\\\"maxReplicaCount\\\":2,\\\"minReplicaCount\\\":1,\\\"triggerAuthentication\\\":{\\\"enabled\\\":false,\\\"name\\\":\\\"\\\",\\\"spec\\\":{}},\\\"triggers\\\":[]},\\\"networkPolicy\\\":{\\\"annotations\\\":{},\\\"egress\\\":[],\\\"enabled\\\":false,\\\"ingress\\\":[],\\\"labels\\\":{},\\\"podSelector\\\":{\\\"matchExpressions\\\":[],\\\"matchLabels\\\":{}},\\\"policyTypes\\\":[]},\\\"pauseForSecondsBeforeSwitchActive\\\":30,\\\"podAnnotations\\\":{},\\\"podDisruptionBudget\\\":{},\\\"podLabels\\\":{},\\\"podSecurityContext\\\":{},\\\"prometheus\\\":{\\\"release\\\":\\\"monitoring\\\"},\\\"rawYaml\\\":[],\\\"replicaCount\\\":1,\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"0.05\\\",\\\"memory\\\":\\\"50Mi\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"0.01\\\",\\\"memory\\\":\\\"10Mi\\\"}},\\\"restartPolicy\\\":\\\"Always\\\",\\\"rolloutAnnotations\\\":{},\\\"rolloutLabels\\\":{},\\\"secret\\\":{\\\"data\\\":{},\\\"enabled\\\":false},\\\"server\\\":{\\\"deployment\\\":{\\\"image\\\":\\\"\\\",\\\"image_tag\\\":\\\"1-95af053\\\"}},\\\"service\\\":{\\\"annotations\\\":{},\\\"loadBalancerSourceRanges\\\":[],\\\"type\\\":\\\"ClusterIP\\\"},\\\"serviceAccount\\\":{\\\"annotations\\\":{},\\\"create\\\":false,\\\"name\\\":\\\"\\\"},\\\"servicemonitor\\\":{\\\"additionalLabels\\\":{}},\\\"tolerations\\\":[],\\\"topologySpreadConstraints\\\":[],\\\"volumeMounts\\\":[],\\\"volumes\\\":[],\\\"waitForSecondsBeforeScalingDown\\\":30,\\\"winterSoldier\\\":{\\\"action\\\":\\\"sleep\\\",\\\"annotation\\\":{},\\\"apiVersion\\\":\\\"pincher.devtron.ai/v1alpha1\\\",\\\"enabled\\\":false,\\\"fieldSelector\\\":[\\\"AfterTime(AddTime(ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'), '5m'), Now())\\\"],\\\"labels\\\":{},\\\"targetReplicas\\\":[],\\\"timeRangesWithZone\\\":{\\\"timeRanges\\\":[],\\\"timeZone\\\":\\\"Asia/Kolkata\\\"},\\\"type\\\":\\\"Rollout\\\"}}",
					GetValues: true,
				},
			},
			wantErr: false,
			want: DeploymentTemplateResponse{
				Data: "{\\\"ContainerPort\\\":[{\\\"envoyPort\\\":6969,\\\"idleTimeout\\\":\\\"6969s\\\",\\\"name\\\":\\\"app\\\",\\\"port\\\":6969,\\\"servicePort\\\":69,\\\"supportStreaming\\\":false,\\\"useHTTP2\\\":false}],\\\"EnvVariables\\\":[],\\\"GracePeriod\\\":30,\\\"LivenessProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"scheme\\\":\\\"\\\",\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"MaxSurge\\\":1,\\\"MaxUnavailable\\\":0,\\\"MinReadySeconds\\\":60,\\\"ReadinessProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"scheme\\\":\\\"\\\",\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"Spec\\\":{\\\"Affinity\\\":{\\\"Key\\\":null,\\\"Values\\\":\\\"nodes\\\",\\\"key\\\":\\\"\\\"}},\\\"StartupProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"ambassadorMapping\\\":{\\\"ambassadorId\\\":\\\"\\\",\\\"cors\\\":{},\\\"enabled\\\":false,\\\"hostname\\\":\\\"devtron.example.com\\\",\\\"labels\\\":{},\\\"prefix\\\":\\\"/\\\",\\\"retryPolicy\\\":{},\\\"rewrite\\\":\\\"\\\",\\\"tls\\\":{\\\"context\\\":\\\"\\\",\\\"create\\\":false,\\\"hosts\\\":[],\\\"secretName\\\":\\\"\\\"}},\\\"args\\\":{\\\"enabled\\\":false,\\\"value\\\":[\\\"/bin/sh\\\",\\\"-c\\\",\\\"touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600\\\"]},\\\"autoscaling\\\":{\\\"MaxReplicas\\\":2,\\\"MinReplicas\\\":1,\\\"TargetCPUUtilizationPercentage\\\":90,\\\"TargetMemoryUtilizationPercentage\\\":69,\\\"annotations\\\":{},\\\"behavior\\\":{},\\\"enabled\\\":false,\\\"extraMetrics\\\":[],\\\"labels\\\":{}},\\\"command\\\":{\\\"enabled\\\":false,\\\"value\\\":[],\\\"workingDir\\\":{}},\\\"containerSecurityContext\\\":{},\\\"containerSpec\\\":{\\\"lifecycle\\\":{\\\"enabled\\\":false,\\\"postStart\\\":{\\\"httpGet\\\":{\\\"host\\\":\\\"example.com\\\",\\\"path\\\":\\\"/example\\\",\\\"port\\\":90}},\\\"preStop\\\":{\\\"exec\\\":{\\\"command\\\":[\\\"sleep\\\",\\\"10\\\"]}}}},\\\"containers\\\":[],\\\"dbMigrationConfig\\\":{\\\"enabled\\\":false},\\\"envoyproxy\\\":{\\\"configMapName\\\":\\\"\\\",\\\"image\\\":\\\"docker.io/envoyproxy/envoy:v1.16.0\\\",\\\"lifecycle\\\":{},\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"50m\\\",\\\"memory\\\":\\\"50Mi\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"50m\\\",\\\"memory\\\":\\\"50Mi\\\"}}},\\\"hostAliases\\\":[],\\\"image\\\":{\\\"pullPolicy\\\":\\\"IfNotPresent\\\"},\\\"imagePullSecrets\\\":[],\\\"ingress\\\":{\\\"annotations\\\":{},\\\"className\\\":\\\"\\\",\\\"enabled\\\":false,\\\"hosts\\\":[{\\\"host\\\":\\\"chart-example1.local\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example1\\\"]},{\\\"host\\\":\\\"chart-example2.local\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example2\\\",\\\"/example2/healthz\\\"]}],\\\"labels\\\":{},\\\"tls\\\":[]},\\\"ingressInternal\\\":{\\\"annotations\\\":{},\\\"className\\\":\\\"\\\",\\\"enabled\\\":false,\\\"hosts\\\":[{\\\"host\\\":\\\"chart-example1.internal\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example1\\\"]},{\\\"host\\\":\\\"chart-example2.internal\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example2\\\",\\\"/example2/healthz\\\"]}],\\\"tls\\\":[]},\\\"initContainers\\\":[],\\\"istio\\\":{\\\"authorizationPolicy\\\":{\\\"action\\\":null,\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"provider\\\":{},\\\"rules\\\":[]},\\\"destinationRule\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"subsets\\\":[],\\\"trafficPolicy\\\":{}},\\\"enable\\\":false,\\\"gateway\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"host\\\":\\\"example.com\\\",\\\"labels\\\":{},\\\"tls\\\":{\\\"enabled\\\":false,\\\"secretName\\\":\\\"secret-name\\\"}},\\\"peerAuthentication\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"mtls\\\":{\\\"mode\\\":null},\\\"portLevelMtls\\\":{},\\\"selector\\\":{\\\"enabled\\\":false}},\\\"requestAuthentication\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"jwtRules\\\":[],\\\"labels\\\":{},\\\"selector\\\":{\\\"enabled\\\":false}},\\\"virtualService\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"gateways\\\":[],\\\"hosts\\\":[],\\\"http\\\":[],\\\"labels\\\":{}}},\\\"kedaAutoscaling\\\":{\\\"advanced\\\":{},\\\"authenticationRef\\\":{},\\\"enabled\\\":false,\\\"envSourceContainerName\\\":\\\"\\\",\\\"maxReplicaCount\\\":2,\\\"minReplicaCount\\\":1,\\\"triggerAuthentication\\\":{\\\"enabled\\\":false,\\\"name\\\":\\\"\\\",\\\"spec\\\":{}},\\\"triggers\\\":[]},\\\"networkPolicy\\\":{\\\"annotations\\\":{},\\\"egress\\\":[],\\\"enabled\\\":false,\\\"ingress\\\":[],\\\"labels\\\":{},\\\"podSelector\\\":{\\\"matchExpressions\\\":[],\\\"matchLabels\\\":{}},\\\"policyTypes\\\":[]},\\\"pauseForSecondsBeforeSwitchActive\\\":30,\\\"podAnnotations\\\":{},\\\"podDisruptionBudget\\\":{},\\\"podLabels\\\":{},\\\"podSecurityContext\\\":{},\\\"prometheus\\\":{\\\"release\\\":\\\"monitoring\\\"},\\\"rawYaml\\\":[],\\\"replicaCount\\\":1,\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"0.05\\\",\\\"memory\\\":\\\"50Mi\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"0.01\\\",\\\"memory\\\":\\\"10Mi\\\"}},\\\"restartPolicy\\\":\\\"Always\\\",\\\"rolloutAnnotations\\\":{},\\\"rolloutLabels\\\":{},\\\"secret\\\":{\\\"data\\\":{},\\\"enabled\\\":false},\\\"server\\\":{\\\"deployment\\\":{\\\"image\\\":\\\"\\\",\\\"image_tag\\\":\\\"1-95af053\\\"}},\\\"service\\\":{\\\"annotations\\\":{},\\\"loadBalancerSourceRanges\\\":[],\\\"type\\\":\\\"ClusterIP\\\"},\\\"serviceAccount\\\":{\\\"annotations\\\":{},\\\"create\\\":false,\\\"name\\\":\\\"\\\"},\\\"servicemonitor\\\":{\\\"additionalLabels\\\":{}},\\\"tolerations\\\":[],\\\"topologySpreadConstraints\\\":[],\\\"volumeMounts\\\":[],\\\"volumes\\\":[],\\\"waitForSecondsBeforeScalingDown\\\":30,\\\"winterSoldier\\\":{\\\"action\\\":\\\"sleep\\\",\\\"annotation\\\":{},\\\"apiVersion\\\":\\\"pincher.devtron.ai/v1alpha1\\\",\\\"enabled\\\":false,\\\"fieldSelector\\\":[\\\"AfterTime(AddTime(ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'), '5m'), Now())\\\"],\\\"labels\\\":{},\\\"targetReplicas\\\":[],\\\"timeRangesWithZone\\\":{\\\"timeRanges\\\":[],\\\"timeZone\\\":\\\"Asia/Kolkata\\\"},\\\"type\\\":\\\"Rollout\\\"}}"},
		},
		{
			name: "get values for base charts",
			args: args{
				ctx: context.Background(),
				request: DeploymentTemplateRequest{
					Values:     "",
					GetValues:  true,
					Type:       1,
					ChartRefId: 1,
				},
			},
			wantErr: false,
			want: DeploymentTemplateResponse{
				Data: "{\\\"ContainerPort\\\":[{\\\"envoyPort\\\":6969,\\\"idleTimeout\\\":\\\"6969s\\\",\\\"name\\\":\\\"app\\\",\\\"port\\\":6969,\\\"servicePort\\\":69,\\\"supportStreaming\\\":false,\\\"useHTTP2\\\":false}],\\\"EnvVariables\\\":[],\\\"GracePeriod\\\":30,\\\"LivenessProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"scheme\\\":\\\"\\\",\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"MaxSurge\\\":1,\\\"MaxUnavailable\\\":0,\\\"MinReadySeconds\\\":60,\\\"ReadinessProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"scheme\\\":\\\"\\\",\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"Spec\\\":{\\\"Affinity\\\":{\\\"Key\\\":null,\\\"Values\\\":\\\"nodes\\\",\\\"key\\\":\\\"\\\"}},\\\"StartupProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"ambassadorMapping\\\":{\\\"ambassadorId\\\":\\\"\\\",\\\"cors\\\":{},\\\"enabled\\\":false,\\\"hostname\\\":\\\"devtron.example.com\\\",\\\"labels\\\":{},\\\"prefix\\\":\\\"/\\\",\\\"retryPolicy\\\":{},\\\"rewrite\\\":\\\"\\\",\\\"tls\\\":{\\\"context\\\":\\\"\\\",\\\"create\\\":false,\\\"hosts\\\":[],\\\"secretName\\\":\\\"\\\"}},\\\"args\\\":{\\\"enabled\\\":false,\\\"value\\\":[\\\"/bin/sh\\\",\\\"-c\\\",\\\"touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600\\\"]},\\\"autoscaling\\\":{\\\"MaxReplicas\\\":2,\\\"MinReplicas\\\":1,\\\"TargetCPUUtilizationPercentage\\\":90,\\\"TargetMemoryUtilizationPercentage\\\":69,\\\"annotations\\\":{},\\\"behavior\\\":{},\\\"enabled\\\":false,\\\"extraMetrics\\\":[],\\\"labels\\\":{}},\\\"command\\\":{\\\"enabled\\\":false,\\\"value\\\":[],\\\"workingDir\\\":{}},\\\"containerSecurityContext\\\":{},\\\"containerSpec\\\":{\\\"lifecycle\\\":{\\\"enabled\\\":false,\\\"postStart\\\":{\\\"httpGet\\\":{\\\"host\\\":\\\"example.com\\\",\\\"path\\\":\\\"/example\\\",\\\"port\\\":90}},\\\"preStop\\\":{\\\"exec\\\":{\\\"command\\\":[\\\"sleep\\\",\\\"10\\\"]}}}},\\\"containers\\\":[],\\\"dbMigrationConfig\\\":{\\\"enabled\\\":false},\\\"envoyproxy\\\":{\\\"configMapName\\\":\\\"\\\",\\\"image\\\":\\\"docker.io/envoyproxy/envoy:v1.16.0\\\",\\\"lifecycle\\\":{},\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"50m\\\",\\\"memory\\\":\\\"50Mi\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"50m\\\",\\\"memory\\\":\\\"50Mi\\\"}}},\\\"hostAliases\\\":[],\\\"image\\\":{\\\"pullPolicy\\\":\\\"IfNotPresent\\\"},\\\"imagePullSecrets\\\":[],\\\"ingress\\\":{\\\"annotations\\\":{},\\\"className\\\":\\\"\\\",\\\"enabled\\\":false,\\\"hosts\\\":[{\\\"host\\\":\\\"chart-example1.local\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example1\\\"]},{\\\"host\\\":\\\"chart-example2.local\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example2\\\",\\\"/example2/healthz\\\"]}],\\\"labels\\\":{},\\\"tls\\\":[]},\\\"ingressInternal\\\":{\\\"annotations\\\":{},\\\"className\\\":\\\"\\\",\\\"enabled\\\":false,\\\"hosts\\\":[{\\\"host\\\":\\\"chart-example1.internal\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example1\\\"]},{\\\"host\\\":\\\"chart-example2.internal\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example2\\\",\\\"/example2/healthz\\\"]}],\\\"tls\\\":[]},\\\"initContainers\\\":[],\\\"istio\\\":{\\\"authorizationPolicy\\\":{\\\"action\\\":null,\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"provider\\\":{},\\\"rules\\\":[]},\\\"destinationRule\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"subsets\\\":[],\\\"trafficPolicy\\\":{}},\\\"enable\\\":false,\\\"gateway\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"host\\\":\\\"example.com\\\",\\\"labels\\\":{},\\\"tls\\\":{\\\"enabled\\\":false,\\\"secretName\\\":\\\"secret-name\\\"}},\\\"peerAuthentication\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"mtls\\\":{\\\"mode\\\":null},\\\"portLevelMtls\\\":{},\\\"selector\\\":{\\\"enabled\\\":false}},\\\"requestAuthentication\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"jwtRules\\\":[],\\\"labels\\\":{},\\\"selector\\\":{\\\"enabled\\\":false}},\\\"virtualService\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"gateways\\\":[],\\\"hosts\\\":[],\\\"http\\\":[],\\\"labels\\\":{}}},\\\"kedaAutoscaling\\\":{\\\"advanced\\\":{},\\\"authenticationRef\\\":{},\\\"enabled\\\":false,\\\"envSourceContainerName\\\":\\\"\\\",\\\"maxReplicaCount\\\":2,\\\"minReplicaCount\\\":1,\\\"triggerAuthentication\\\":{\\\"enabled\\\":false,\\\"name\\\":\\\"\\\",\\\"spec\\\":{}},\\\"triggers\\\":[]},\\\"networkPolicy\\\":{\\\"annotations\\\":{},\\\"egress\\\":[],\\\"enabled\\\":false,\\\"ingress\\\":[],\\\"labels\\\":{},\\\"podSelector\\\":{\\\"matchExpressions\\\":[],\\\"matchLabels\\\":{}},\\\"policyTypes\\\":[]},\\\"pauseForSecondsBeforeSwitchActive\\\":30,\\\"podAnnotations\\\":{},\\\"podDisruptionBudget\\\":{},\\\"podLabels\\\":{},\\\"podSecurityContext\\\":{},\\\"prometheus\\\":{\\\"release\\\":\\\"monitoring\\\"},\\\"rawYaml\\\":[],\\\"replicaCount\\\":1,\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"0.05\\\",\\\"memory\\\":\\\"50Mi\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"0.01\\\",\\\"memory\\\":\\\"10Mi\\\"}},\\\"restartPolicy\\\":\\\"Always\\\",\\\"rolloutAnnotations\\\":{},\\\"rolloutLabels\\\":{},\\\"secret\\\":{\\\"data\\\":{},\\\"enabled\\\":false},\\\"server\\\":{\\\"deployment\\\":{\\\"image\\\":\\\"\\\",\\\"image_tag\\\":\\\"1-95af053\\\"}},\\\"service\\\":{\\\"annotations\\\":{},\\\"loadBalancerSourceRanges\\\":[],\\\"type\\\":\\\"ClusterIP\\\"},\\\"serviceAccount\\\":{\\\"annotations\\\":{},\\\"create\\\":false,\\\"name\\\":\\\"\\\"},\\\"servicemonitor\\\":{\\\"additionalLabels\\\":{}},\\\"tolerations\\\":[],\\\"topologySpreadConstraints\\\":[],\\\"volumeMounts\\\":[],\\\"volumes\\\":[],\\\"waitForSecondsBeforeScalingDown\\\":30,\\\"winterSoldier\\\":{\\\"action\\\":\\\"sleep\\\",\\\"annotation\\\":{},\\\"apiVersion\\\":\\\"pincher.devtron.ai/v1alpha1\\\",\\\"enabled\\\":false,\\\"fieldSelector\\\":[\\\"AfterTime(AddTime(ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'), '5m'), Now())\\\"],\\\"labels\\\":{},\\\"targetReplicas\\\":[],\\\"timeRangesWithZone\\\":{\\\"timeRanges\\\":[],\\\"timeZone\\\":\\\"Asia/Kolkata\\\"},\\\"type\\\":\\\"Rollout\\\"}}"},
		},
		{
			name: "get values for published on other envs",
			args: args{
				ctx: context.Background(),
				request: DeploymentTemplateRequest{
					Values:     "",
					GetValues:  true,
					Type:       2,
					ChartRefId: 1,
					AppId:      1,
				},
			},
			wantErr: false,
			want: DeploymentTemplateResponse{
				Data: "{\\\"ContainerPort\\\":[{\\\"envoyPort\\\":6969,\\\"idleTimeout\\\":\\\"6969s\\\",\\\"name\\\":\\\"app\\\",\\\"port\\\":6969,\\\"servicePort\\\":69,\\\"supportStreaming\\\":false,\\\"useHTTP2\\\":false}],\\\"EnvVariables\\\":[],\\\"GracePeriod\\\":30,\\\"LivenessProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"scheme\\\":\\\"\\\",\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"MaxSurge\\\":1,\\\"MaxUnavailable\\\":0,\\\"MinReadySeconds\\\":60,\\\"ReadinessProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"scheme\\\":\\\"\\\",\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"Spec\\\":{\\\"Affinity\\\":{\\\"Key\\\":null,\\\"Values\\\":\\\"nodes\\\",\\\"key\\\":\\\"\\\"}},\\\"StartupProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"ambassadorMapping\\\":{\\\"ambassadorId\\\":\\\"\\\",\\\"cors\\\":{},\\\"enabled\\\":false,\\\"hostname\\\":\\\"devtron.example.com\\\",\\\"labels\\\":{},\\\"prefix\\\":\\\"/\\\",\\\"retryPolicy\\\":{},\\\"rewrite\\\":\\\"\\\",\\\"tls\\\":{\\\"context\\\":\\\"\\\",\\\"create\\\":false,\\\"hosts\\\":[],\\\"secretName\\\":\\\"\\\"}},\\\"args\\\":{\\\"enabled\\\":false,\\\"value\\\":[\\\"/bin/sh\\\",\\\"-c\\\",\\\"touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600\\\"]},\\\"autoscaling\\\":{\\\"MaxReplicas\\\":2,\\\"MinReplicas\\\":1,\\\"TargetCPUUtilizationPercentage\\\":90,\\\"TargetMemoryUtilizationPercentage\\\":69,\\\"annotations\\\":{},\\\"behavior\\\":{},\\\"enabled\\\":false,\\\"extraMetrics\\\":[],\\\"labels\\\":{}},\\\"command\\\":{\\\"enabled\\\":false,\\\"value\\\":[],\\\"workingDir\\\":{}},\\\"containerSecurityContext\\\":{},\\\"containerSpec\\\":{\\\"lifecycle\\\":{\\\"enabled\\\":false,\\\"postStart\\\":{\\\"httpGet\\\":{\\\"host\\\":\\\"example.com\\\",\\\"path\\\":\\\"/example\\\",\\\"port\\\":90}},\\\"preStop\\\":{\\\"exec\\\":{\\\"command\\\":[\\\"sleep\\\",\\\"10\\\"]}}}},\\\"containers\\\":[],\\\"dbMigrationConfig\\\":{\\\"enabled\\\":false},\\\"envoyproxy\\\":{\\\"configMapName\\\":\\\"\\\",\\\"image\\\":\\\"docker.io/envoyproxy/envoy:v1.16.0\\\",\\\"lifecycle\\\":{},\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"50m\\\",\\\"memory\\\":\\\"50Mi\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"50m\\\",\\\"memory\\\":\\\"50Mi\\\"}}},\\\"hostAliases\\\":[],\\\"image\\\":{\\\"pullPolicy\\\":\\\"IfNotPresent\\\"},\\\"imagePullSecrets\\\":[],\\\"ingress\\\":{\\\"annotations\\\":{},\\\"className\\\":\\\"\\\",\\\"enabled\\\":false,\\\"hosts\\\":[{\\\"host\\\":\\\"chart-example1.local\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example1\\\"]},{\\\"host\\\":\\\"chart-example2.local\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example2\\\",\\\"/example2/healthz\\\"]}],\\\"labels\\\":{},\\\"tls\\\":[]},\\\"ingressInternal\\\":{\\\"annotations\\\":{},\\\"className\\\":\\\"\\\",\\\"enabled\\\":false,\\\"hosts\\\":[{\\\"host\\\":\\\"chart-example1.internal\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example1\\\"]},{\\\"host\\\":\\\"chart-example2.internal\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example2\\\",\\\"/example2/healthz\\\"]}],\\\"tls\\\":[]},\\\"initContainers\\\":[],\\\"istio\\\":{\\\"authorizationPolicy\\\":{\\\"action\\\":null,\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"provider\\\":{},\\\"rules\\\":[]},\\\"destinationRule\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"subsets\\\":[],\\\"trafficPolicy\\\":{}},\\\"enable\\\":false,\\\"gateway\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"host\\\":\\\"example.com\\\",\\\"labels\\\":{},\\\"tls\\\":{\\\"enabled\\\":false,\\\"secretName\\\":\\\"secret-name\\\"}},\\\"peerAuthentication\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"mtls\\\":{\\\"mode\\\":null},\\\"portLevelMtls\\\":{},\\\"selector\\\":{\\\"enabled\\\":false}},\\\"requestAuthentication\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"jwtRules\\\":[],\\\"labels\\\":{},\\\"selector\\\":{\\\"enabled\\\":false}},\\\"virtualService\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"gateways\\\":[],\\\"hosts\\\":[],\\\"http\\\":[],\\\"labels\\\":{}}},\\\"kedaAutoscaling\\\":{\\\"advanced\\\":{},\\\"authenticationRef\\\":{},\\\"enabled\\\":false,\\\"envSourceContainerName\\\":\\\"\\\",\\\"maxReplicaCount\\\":2,\\\"minReplicaCount\\\":1,\\\"triggerAuthentication\\\":{\\\"enabled\\\":false,\\\"name\\\":\\\"\\\",\\\"spec\\\":{}},\\\"triggers\\\":[]},\\\"networkPolicy\\\":{\\\"annotations\\\":{},\\\"egress\\\":[],\\\"enabled\\\":false,\\\"ingress\\\":[],\\\"labels\\\":{},\\\"podSelector\\\":{\\\"matchExpressions\\\":[],\\\"matchLabels\\\":{}},\\\"policyTypes\\\":[]},\\\"pauseForSecondsBeforeSwitchActive\\\":30,\\\"podAnnotations\\\":{},\\\"podDisruptionBudget\\\":{},\\\"podLabels\\\":{},\\\"podSecurityContext\\\":{},\\\"prometheus\\\":{\\\"release\\\":\\\"monitoring\\\"},\\\"rawYaml\\\":[],\\\"replicaCount\\\":1,\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"0.05\\\",\\\"memory\\\":\\\"50Mi\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"0.01\\\",\\\"memory\\\":\\\"10Mi\\\"}},\\\"restartPolicy\\\":\\\"Always\\\",\\\"rolloutAnnotations\\\":{},\\\"rolloutLabels\\\":{},\\\"secret\\\":{\\\"data\\\":{},\\\"enabled\\\":false},\\\"server\\\":{\\\"deployment\\\":{\\\"image\\\":\\\"\\\",\\\"image_tag\\\":\\\"1-95af053\\\"}},\\\"service\\\":{\\\"annotations\\\":{},\\\"loadBalancerSourceRanges\\\":[],\\\"type\\\":\\\"ClusterIP\\\"},\\\"serviceAccount\\\":{\\\"annotations\\\":{},\\\"create\\\":false,\\\"name\\\":\\\"\\\"},\\\"servicemonitor\\\":{\\\"additionalLabels\\\":{}},\\\"tolerations\\\":[],\\\"topologySpreadConstraints\\\":[],\\\"volumeMounts\\\":[],\\\"volumes\\\":[],\\\"waitForSecondsBeforeScalingDown\\\":30,\\\"winterSoldier\\\":{\\\"action\\\":\\\"sleep\\\",\\\"annotation\\\":{},\\\"apiVersion\\\":\\\"pincher.devtron.ai/v1alpha1\\\",\\\"enabled\\\":false,\\\"fieldSelector\\\":[\\\"AfterTime(AddTime(ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'), '5m'), Now())\\\"],\\\"labels\\\":{},\\\"targetReplicas\\\":[],\\\"timeRangesWithZone\\\":{\\\"timeRanges\\\":[],\\\"timeZone\\\":\\\"Asia/Kolkata\\\"},\\\"type\\\":\\\"Rollout\\\"}}"},
		},
		{
			name: "get values for deployed on envs",
			args: args{
				ctx: context.Background(),
				request: DeploymentTemplateRequest{
					Values:                   "",
					GetValues:                true,
					Type:                     3,
					ChartRefId:               1,
					AppId:                    1,
					PipelineConfigOverrideId: 1,
				},
			},
			wantErr: false,
			want: DeploymentTemplateResponse{
				Data: "{\\\"ContainerPort\\\":[{\\\"envoyPort\\\":6969,\\\"idleTimeout\\\":\\\"6969s\\\",\\\"name\\\":\\\"app\\\",\\\"port\\\":6969,\\\"servicePort\\\":69,\\\"supportStreaming\\\":false,\\\"useHTTP2\\\":false}],\\\"EnvVariables\\\":[],\\\"GracePeriod\\\":30,\\\"LivenessProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"scheme\\\":\\\"\\\",\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"MaxSurge\\\":1,\\\"MaxUnavailable\\\":0,\\\"MinReadySeconds\\\":60,\\\"ReadinessProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"scheme\\\":\\\"\\\",\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"Spec\\\":{\\\"Affinity\\\":{\\\"Key\\\":null,\\\"Values\\\":\\\"nodes\\\",\\\"key\\\":\\\"\\\"}},\\\"StartupProbe\\\":{\\\"Path\\\":\\\"\\\",\\\"command\\\":[],\\\"failureThreshold\\\":3,\\\"httpHeaders\\\":[],\\\"initialDelaySeconds\\\":20,\\\"periodSeconds\\\":10,\\\"port\\\":6969,\\\"successThreshold\\\":1,\\\"tcp\\\":false,\\\"timeoutSeconds\\\":5},\\\"ambassadorMapping\\\":{\\\"ambassadorId\\\":\\\"\\\",\\\"cors\\\":{},\\\"enabled\\\":false,\\\"hostname\\\":\\\"devtron.example.com\\\",\\\"labels\\\":{},\\\"prefix\\\":\\\"/\\\",\\\"retryPolicy\\\":{},\\\"rewrite\\\":\\\"\\\",\\\"tls\\\":{\\\"context\\\":\\\"\\\",\\\"create\\\":false,\\\"hosts\\\":[],\\\"secretName\\\":\\\"\\\"}},\\\"args\\\":{\\\"enabled\\\":false,\\\"value\\\":[\\\"/bin/sh\\\",\\\"-c\\\",\\\"touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600\\\"]},\\\"autoscaling\\\":{\\\"MaxReplicas\\\":2,\\\"MinReplicas\\\":1,\\\"TargetCPUUtilizationPercentage\\\":90,\\\"TargetMemoryUtilizationPercentage\\\":69,\\\"annotations\\\":{},\\\"behavior\\\":{},\\\"enabled\\\":false,\\\"extraMetrics\\\":[],\\\"labels\\\":{}},\\\"command\\\":{\\\"enabled\\\":false,\\\"value\\\":[],\\\"workingDir\\\":{}},\\\"containerSecurityContext\\\":{},\\\"containerSpec\\\":{\\\"lifecycle\\\":{\\\"enabled\\\":false,\\\"postStart\\\":{\\\"httpGet\\\":{\\\"host\\\":\\\"example.com\\\",\\\"path\\\":\\\"/example\\\",\\\"port\\\":90}},\\\"preStop\\\":{\\\"exec\\\":{\\\"command\\\":[\\\"sleep\\\",\\\"10\\\"]}}}},\\\"containers\\\":[],\\\"dbMigrationConfig\\\":{\\\"enabled\\\":false},\\\"envoyproxy\\\":{\\\"configMapName\\\":\\\"\\\",\\\"image\\\":\\\"docker.io/envoyproxy/envoy:v1.16.0\\\",\\\"lifecycle\\\":{},\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"50m\\\",\\\"memory\\\":\\\"50Mi\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"50m\\\",\\\"memory\\\":\\\"50Mi\\\"}}},\\\"hostAliases\\\":[],\\\"image\\\":{\\\"pullPolicy\\\":\\\"IfNotPresent\\\"},\\\"imagePullSecrets\\\":[],\\\"ingress\\\":{\\\"annotations\\\":{},\\\"className\\\":\\\"\\\",\\\"enabled\\\":false,\\\"hosts\\\":[{\\\"host\\\":\\\"chart-example1.local\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example1\\\"]},{\\\"host\\\":\\\"chart-example2.local\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example2\\\",\\\"/example2/healthz\\\"]}],\\\"labels\\\":{},\\\"tls\\\":[]},\\\"ingressInternal\\\":{\\\"annotations\\\":{},\\\"className\\\":\\\"\\\",\\\"enabled\\\":false,\\\"hosts\\\":[{\\\"host\\\":\\\"chart-example1.internal\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example1\\\"]},{\\\"host\\\":\\\"chart-example2.internal\\\",\\\"pathType\\\":\\\"ImplementationSpecific\\\",\\\"paths\\\":[\\\"/example2\\\",\\\"/example2/healthz\\\"]}],\\\"tls\\\":[]},\\\"initContainers\\\":[],\\\"istio\\\":{\\\"authorizationPolicy\\\":{\\\"action\\\":null,\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"provider\\\":{},\\\"rules\\\":[]},\\\"destinationRule\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"subsets\\\":[],\\\"trafficPolicy\\\":{}},\\\"enable\\\":false,\\\"gateway\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"host\\\":\\\"example.com\\\",\\\"labels\\\":{},\\\"tls\\\":{\\\"enabled\\\":false,\\\"secretName\\\":\\\"secret-name\\\"}},\\\"peerAuthentication\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"labels\\\":{},\\\"mtls\\\":{\\\"mode\\\":null},\\\"portLevelMtls\\\":{},\\\"selector\\\":{\\\"enabled\\\":false}},\\\"requestAuthentication\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"jwtRules\\\":[],\\\"labels\\\":{},\\\"selector\\\":{\\\"enabled\\\":false}},\\\"virtualService\\\":{\\\"annotations\\\":{},\\\"enabled\\\":false,\\\"gateways\\\":[],\\\"hosts\\\":[],\\\"http\\\":[],\\\"labels\\\":{}}},\\\"kedaAutoscaling\\\":{\\\"advanced\\\":{},\\\"authenticationRef\\\":{},\\\"enabled\\\":false,\\\"envSourceContainerName\\\":\\\"\\\",\\\"maxReplicaCount\\\":2,\\\"minReplicaCount\\\":1,\\\"triggerAuthentication\\\":{\\\"enabled\\\":false,\\\"name\\\":\\\"\\\",\\\"spec\\\":{}},\\\"triggers\\\":[]},\\\"networkPolicy\\\":{\\\"annotations\\\":{},\\\"egress\\\":[],\\\"enabled\\\":false,\\\"ingress\\\":[],\\\"labels\\\":{},\\\"podSelector\\\":{\\\"matchExpressions\\\":[],\\\"matchLabels\\\":{}},\\\"policyTypes\\\":[]},\\\"pauseForSecondsBeforeSwitchActive\\\":30,\\\"podAnnotations\\\":{},\\\"podDisruptionBudget\\\":{},\\\"podLabels\\\":{},\\\"podSecurityContext\\\":{},\\\"prometheus\\\":{\\\"release\\\":\\\"monitoring\\\"},\\\"rawYaml\\\":[],\\\"replicaCount\\\":1,\\\"resources\\\":{\\\"limits\\\":{\\\"cpu\\\":\\\"0.05\\\",\\\"memory\\\":\\\"50Mi\\\"},\\\"requests\\\":{\\\"cpu\\\":\\\"0.01\\\",\\\"memory\\\":\\\"10Mi\\\"}},\\\"restartPolicy\\\":\\\"Always\\\",\\\"rolloutAnnotations\\\":{},\\\"rolloutLabels\\\":{},\\\"secret\\\":{\\\"data\\\":{},\\\"enabled\\\":false},\\\"server\\\":{\\\"deployment\\\":{\\\"image\\\":\\\"\\\",\\\"image_tag\\\":\\\"1-95af053\\\"}},\\\"service\\\":{\\\"annotations\\\":{},\\\"loadBalancerSourceRanges\\\":[],\\\"type\\\":\\\"ClusterIP\\\"},\\\"serviceAccount\\\":{\\\"annotations\\\":{},\\\"create\\\":false,\\\"name\\\":\\\"\\\"},\\\"servicemonitor\\\":{\\\"additionalLabels\\\":{}},\\\"tolerations\\\":[],\\\"topologySpreadConstraints\\\":[],\\\"volumeMounts\\\":[],\\\"volumes\\\":[],\\\"waitForSecondsBeforeScalingDown\\\":30,\\\"winterSoldier\\\":{\\\"action\\\":\\\"sleep\\\",\\\"annotation\\\":{},\\\"apiVersion\\\":\\\"pincher.devtron.ai/v1alpha1\\\",\\\"enabled\\\":false,\\\"fieldSelector\\\":[\\\"AfterTime(AddTime(ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'), '5m'), Now())\\\"],\\\"labels\\\":{},\\\"targetReplicas\\\":[],\\\"timeRangesWithZone\\\":{\\\"timeRanges\\\":[],\\\"timeZone\\\":\\\"Asia/Kolkata\\\"},\\\"type\\\":\\\"Rollout\\\"}}"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl, chartService, _, deploymentTemplateRepository, chartRepository := InitEventSimpleFactoryImpl(t)
			if tt.name == "get values for base charts" {
				chartService.On("GetAppOverrideForDefaultTemplate", tt.args.request.ChartRefId).Return(myMap, myString, nil)
			}
			if tt.name == "get values for published on other envs" {
				chartRepository.On("FindLatestChartForAppByAppId", tt.args.request.AppId).Return(chart, nil)
			}

			if tt.name == "get values for deployed on envs" {
				deploymentTemplateRepository.On("FetchPipelineOverrideValues", tt.args.request.PipelineConfigOverrideId).Return(myString, nil)
			}

			_, err := impl.GetDeploymentTemplate(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDeploymentTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("GetDeploymentTemplate() got = %v, want %v", got, got)
			//}
		})
	}
}

func InitEventSimpleFactoryImpl(t *testing.T) (*DeploymentTemplateServiceImpl, *mocks.ChartService, *mocks2.AppListingService, *mocks3.DeploymentTemplateRepository, *mocks5.ChartRepository) {
	logger, _ := util.NewSugardLogger()
	chartService := mocks.NewChartService(t)
	appListingService := mocks2.NewAppListingService(t)
	appListingRepository := mocks3.NewAppListingRepository(t)
	deploymentTemplateRepository := mocks3.NewDeploymentTemplateRepository(t)
	helmAppService := mocks4.NewHelmAppService(t)
	chartRepository := mocks5.NewChartRepository(t)
	chartTemplateServiceImpl := mocks6.NewChartTemplateService(t)
	helmAppClient := mocks4.NewHelmAppClient(t)
	var k8sUtil *k8s.K8sUtil
	if K8sUtilObj != nil {
		k8sUtil = K8sUtilObj
	} else {
		config := &client2.RuntimeConfig{LocalDevMode: true}
		k8sUtil = k8s.NewK8sUtil(logger, config)
		K8sUtilObj = k8sUtil
	}
	impl := &DeploymentTemplateServiceImpl{
		Logger:                       logger,
		chartService:                 chartService,
		appListingService:            appListingService,
		appListingRepository:         appListingRepository,
		deploymentTemplateRepository: deploymentTemplateRepository,
		helmAppService:               helmAppService,
		chartRepository:              chartRepository,
		chartTemplateServiceImpl:     chartTemplateServiceImpl,
		K8sUtil:                      k8sUtil,
		helmAppClient:                helmAppClient,
	}
	return impl, chartService, appListingService, deploymentTemplateRepository, chartRepository
}
