package integrationTest

import (
	"context"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/models"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	mocks3 "github.com/devtron-labs/devtron/internal/sql/repository/chartConfig/mocks"
	mocks5 "github.com/devtron-labs/devtron/internal/sql/repository/mocks"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	mocks2 "github.com/devtron-labs/devtron/pkg/chartRepo/repository/mocks"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	mocks4 "github.com/devtron-labs/devtron/pkg/cluster/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDeploymentTemplateHistoryService(t *testing.T) {

	t.Run("GetEnvOverrideWhenDeploymentConfigTypeIsSpecificTrigger", func(t *testing.T) {

		mockedDeploymentTemplateHistoryRepository := mocks.NewDeploymentTemplateHistoryRepository(t)

		deploymentTemplateHistoryResponse := &repository.DeploymentTemplateHistory{
			Id:                      1,
			PipelineId:              1,
			AppId:                   1,
			ImageDescriptorTemplate: "{\"server\":{\"deployment\":{\"image_tag\":\"{{.Tag}}\",\"image\":\"{{.Name}}\"}},\"pipelineName\": \"{{.PipelineName}}\",\"releaseVersion\":\"{{.ReleaseVersion}}\",\"deploymentType\": \"{{.Deploymen\ntType}}\", \"app\": \"{{.App}}\", \"env\": \"{{.Env}}\", \"appMetrics\": {{.AppMetrics}}}",
			Template:                "{\"ContainerPort\":[{\"envoyPort\":8799,\"idleTimeout\":\"1800s\",\"name\":\"app\",\"port\":8080,\"servicePort\":80,\"supportStreaming\":false,\"useHTTP2\":false}],\"EnvVariables\":[],\"GracePeriod\":\n30,\"LivenessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"M\naxSurge\":1,\"MaxUnavailable\":0,\"MinReadySeconds\":60,\"ReadinessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"succe\nssThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"Spec\":{\"Affinity\":{\"Key\":null,\"Values\":\"nodes\",\"key\":\"\"}},\"ambassadorMapping\":{\"ambassadorId\":\"\",\"cors\":{},\"enabled\":false,\"hostname\":\"devtron.example.com\",\n\"labels\":{},\"prefix\":\"/\",\"retryPolicy\":{},\"rewrite\":\"\",\"tls\":{\"context\":\"\",\"create\":false,\"hosts\":[],\"secretName\":\"\"}},\"args\":{\"enabled\":false,\"value\":[\"/bin/sh\",\"-c\",\"touch /tmp/healthy; sleep 30; rm -rf\n /tmp/healthy; sleep 600\"]},\"autoscaling\":{\"MaxReplicas\":2,\"MinReplicas\":1,\"TargetCPUUtilizationPercentage\":90,\"TargetMemoryUtilizationPercentage\":80,\"annotations\":{},\"behavior\":{},\"enabled\":false,\"extraM\netrics\":[],\"labels\":{}},\"command\":{\"enabled\":false,\"value\":[],\"workingDir\":{}},\"containerSecurityContext\":{},\"containerSpec\":{\"lifecycle\":{\"enabled\":false,\"postStart\":{\"httpGet\":{\"host\":\"example.com\",\"pat\nh\":\"/example\",\"port\":90}},\"preStop\":{\"exec\":{\"command\":[\"sleep\",\"10\"]}}}},\"containers\":[],\"dbMigrationConfig\":{\"enabled\":false},\"envoyproxy\":{\"configMapName\":\"\",\"image\":\"quay.io/devtron/envoy:v1.14.1\",\"li\nfecycle\":{},\"resources\":{\"limits\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"},\"requests\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"}}},\"hostAliases\":[],\"image\":{\"pullPolicy\":\"IfNotPresent\"},\"imagePullSecrets\":[],\"ingress\":{\"annotati\nons\":{},\"className\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.local\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example1\"]}],\"labels\":{},\"tls\":[]},\"ingressInternal\":{\"annotations\":{},\"classN\name\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example1\"]},{\"host\":\"chart-example2.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":\n[\"/example2\",\"/example2/healthz\"]}],\"tls\":[]},\"initContainers\":[],\"istio\":{\"enable\":false,\"gateway\":{\"annotations\":{},\"enabled\":false,\"host\":\"example.com\",\"labels\":{},\"tls\":{\"enabled\":false,\"secretName\":\"\nsecret-name\"}},\"virtualService\":{\"annotations\":{},\"enabled\":false,\"gateways\":[],\"hosts\":[],\"http\":[{\"corsPolicy\":{},\"headers\":{},\"match\":[{\"uri\":{\"prefix\":\"/v1\"}},{\"uri\":{\"prefix\":\"/v2\"}}],\"retries\":{\"att\nempts\":2,\"perTryTimeout\":\"3s\"},\"rewriteUri\":\"/\",\"route\":[{\"destination\":{\"host\":\"service1\",\"port\":80}}],\"timeout\":\"12s\"},{\"route\":[{\"destination\":{\"host\":\"service2\"}}]}],\"labels\":{}}},\"kedaAutoscaling\":{\"\nadvanced\":{},\"authenticationRef\":{},\"enabled\":false,\"envSourceContainerName\":\"\",\"maxReplicaCount\":2,\"minReplicaCount\":1,\"triggerAuthentication\":{\"enabled\":false,\"name\":\"\",\"spec\":{}},\"triggers\":[]},\"pauseF\norSecondsBeforeSwitchActive\":30,\"podAnnotations\":{},\"podLabels\":{},\"podSecurityContext\":{},\"prometheus\":{\"release\":\"monitoring\"},\"rawYaml\":[],\"replicaCount\":1,\"resources\":{\"limits\":{\"cpu\":\"0.05\",\"memory\":\n\"50Mi\"},\"requests\":{\"cpu\":\"0.01\",\"memory\":\"10Mi\"}},\"rolloutAnnotations\":{},\"rolloutLabels\":{},\"secret\":{\"data\":{},\"enabled\":false},\"server\":{\"deployment\":{\"image\":\"\",\"image_tag\":\"1-95af053\"}},\"service\":{\"\nannotations\":{},\"loadBalancerSourceRanges\":[],\"type\":\"ClusterIP\"},\"serviceAccount\":{\"annotations\":{},\"create\":false,\"name\":\"\"},\"servicemonitor\":{\"additionalLabels\":{}},\"tolerations\":[],\"topologySpreadCons\ntraints\":[],\"volumeMounts\":[],\"volumes\":[],\"waitForSecondsBeforeScalingDown\":30}\n",
			TargetEnvironment:       1,
			TemplateName:            "Rollout Deployment",
			TemplateVersion:         "4.17.0",
			IsAppMetricsEnabled:     false,
			Deployed:                true,
			DeployedOn:              time.Now(),
			DeployedBy:              1,
			AuditLog:                sql.AuditLog{},
		}
		mockedDeploymentTemplateHistoryRepository.On("GetHistoryByPipelineIdAndWfrId", 1, 1).
			Return(deploymentTemplateHistoryResponse, nil)

		mockedChartRefRepository := mocks2.NewChartRefRepository(t)
		mockedChartRefRepository.On("FindByVersionAndName", "", "4.17.0").
			Return(&chartRepoRepository.ChartRef{
				Id:                     29,
				Location:               "reference-chart_4-17-0",
				Version:                "4.17.0",
				Active:                 true,
				Default:                true,
				Name:                   "",
				ChartData:              nil,
				ChartDescription:       "",
				UserUploaded:           false,
				IsAppMetricsSupported:  true,
				DeploymentStrategyPath: "pipeline-values.yaml",
				JsonPathForStrategy:    "",
				AuditLog:               sql.AuditLog{},
			}, nil)

		mockedEnvConfigOverrideRepository := mocks3.NewEnvConfigOverrideRepository(t)
		mockedEnvConfigOverrideRepository.On("GetByAppIdEnvIdAndChartRefId", 1, 1, 29).
			Return(&chartConfig.EnvConfigOverride{
				Id:                1,
				ChartId:           29,
				TargetEnvironment: 1,
				EnvOverrideValues: "",
				Status:            models.CHARTSTATUS_SUCCESS,
				ManualReviewed:    false,
				Active:            true,
				Namespace:         "default",
				Latest:            true,
				Previous:          false,
				IsOverride:        true,
				IsBasicViewLocked: false,
				CurrentViewEditor: "",
				AuditLog:          sql.AuditLog{},
			}, nil)

		mockedEnvironmentRepository := mocks4.NewEnvironmentRepository(t)
		mockedEnvironmentRepository.On("FindById", 1).
			Return(&repository2.Environment{
				Id:                    1,
				Name:                  "devtron-demo",
				ClusterId:             1,
				Active:                true,
				Default:               false,
				GrafanaDatasourceId:   0,
				Namespace:             "default",
				EnvironmentIdentifier: "devtron-demo",
				Description:           "",
				IsVirtualEnvironment:  false,
				AuditLog:              sql.AuditLog{},
			}, nil)

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		appServiceImpl := app.NewAppService(mockedEnvConfigOverrideRepository, nil, nil, sugaredLogger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, mockedEnvironmentRepository, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", mockedChartRefRepository, nil, nil, nil, nil, nil, nil, nil, mockedDeploymentTemplateHistoryRepository, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		overrideRequest := &bean.ValuesOverrideRequest{
			PipelineId:                            1,
			AppId:                                 1,
			CiArtifactId:                          1,
			AdditionalOverride:                    nil,
			TargetDbVersion:                       0,
			ForceTrigger:                          false,
			DeploymentTemplate:                    "",
			DeploymentWithConfig:                  bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER,
			WfrIdForDeploymentWithSpecificTrigger: 1,
			CdWorkflowType:                        "deploy",
			WfrId:                                 1,
			CdWorkflowId:                          1,
			UserId:                                1,
			DeploymentType:                        models.DEPLOYMENTTYPE_DEPLOY,
			EnvId:                                 1,
			EnvName:                               "",
			ClusterId:                             0,
			AppName:                               "1",
			PipelineName:                          "test",
			DeploymentAppType:                     "", //should work independent of deployment type
		}

		envOverride, err := appServiceImpl.GetEnvOverrideByTriggerType(overrideRequest, time.Now(), context.Background())
		assert.Nil(t, err)
		assert.NotNil(t, envOverride.Environment)
		assert.Equal(t, envOverride.EnvOverrideValues, deploymentTemplateHistoryResponse.Template)
		assert.Equal(t, envOverride.IsOverride, true)

	})

	t.Run("GetEnvOverrideWhenConfigTypeIsLastSavedAndIsFirstTrigger", func(t *testing.T) {

		overrideRequest := &bean.ValuesOverrideRequest{
			PipelineId:                            1,
			AppId:                                 1,
			CiArtifactId:                          1,
			AdditionalOverride:                    nil,
			TargetDbVersion:                       0,
			ForceTrigger:                          false,
			DeploymentTemplate:                    "",
			DeploymentWithConfig:                  bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED,
			WfrIdForDeploymentWithSpecificTrigger: 1,
			CdWorkflowType:                        "deploy",
			WfrId:                                 1,
			CdWorkflowId:                          1,
			UserId:                                1,
			DeploymentType:                        models.DEPLOYMENTTYPE_DEPLOY,
			EnvId:                                 1,
			EnvName:                               "",
			ClusterId:                             0,
			AppName:                               "1",
			PipelineName:                          "test",
			DeploymentAppType:                     "", //should work independent of deployment type
		}

		mockedEnvConfigOverrideRepository := mocks3.NewEnvConfigOverrideRepository(t)
		mockedEnvConfigOverrideRepository.On("ActiveEnvConfigOverride", 1, 1).
			Return(&chartConfig.EnvConfigOverride{}, nil)

		mockedChartOutput := &chartRepoRepository.Chart{
			Id:                      1,
			AppId:                   1,
			ChartRepoId:             1,
			ChartName:               "test-app",
			ChartVersion:            "4.17.1",
			ChartRepo:               "default-chartmuseum",
			ChartRepoUrl:            "http://devtron-chartmuseum.devtroncd:8080/",
			Values:                  "{\"ConfigMaps\":{\"enabled\":false,\"maps\":[]},\"ConfigSecrets\":{\"enabled\":false,\"secrets\":[]},\"ContainerPort\":[{\"envoyPort\":8799,\"idleTimeout\":\"1800s\",\"name\":\"app\",\"port\":8080,\"servicePort\":80,\"supportStreaming\":false,\"useHTTP2\":false}],\"EnvVariables\":[],\"EnvVariablesFromFieldPath\":[{\"fieldPath\":\"metadata.name\",\"name\":\"POD_NAME\"}],\"GracePeriod\":30,\"LivenessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"MaxSurge\":1,\"MaxUnavailable\":0,\"MinReadySeconds\":60,\"ReadinessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"Spec\":{\"Affinity\":{\"Values\":\"nodes\",\"key\":\"\"}},\"ambassadorMapping\":{\"ambassadorId\":\"\",\"cors\":{},\"enabled\":false,\"hostname\":\"devtron.example.com\",\"labels\":{},\"prefix\":\"/\",\"retryPolicy\":{},\"rewrite\":\"\",\"tls\":{\"context\":\"\",\"create\":false,\"hosts\":[],\"secretName\":\"\"}},\"args\":{\"enabled\":false,\"value\":[\"/bin/sh\",\"-c\",\"touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600\"]},\"autoPromotionSeconds\":30,\"autoscaling\":{\"MaxReplicas\":2,\"MinReplicas\":1,\"TargetCPUUtilizationPercentage\":90,\"TargetMemoryUtilizationPercentage\":80,\"annotations\":{},\"behavior\":{},\"enabled\":false,\"extraMetrics\":[],\"labels\":{}},\"command\":{\"enabled\":false,\"value\":[],\"workingDir\":{}},\"containerExtraSpecs\":{},\"containerSecurityContext\":{},\"containerSpec\":{\"lifecycle\":{\"enabled\":false,\"postStart\":{\"httpGet\":{\"host\":\"example.com\",\"path\":\"/example\",\"port\":90}},\"preStop\":{\"exec\":{\"command\":[\"sleep\",\"10\"]}}}},\"containers\":[],\"dbMigrationConfig\":{\"enabled\":false},\"envoyproxy\":{\"configMapName\":\"\",\"image\":\"quay.io/devtron/envoy:v1.14.1\",\"lifecycle\":{},\"resources\":{\"limits\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"},\"requests\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"}}},\"hostAliases\":[],\"image\":{\"pullPolicy\":\"IfNotPresent\"},\"imagePullSecrets\":[],\"ingress\":{\"annotations\":{},\"className\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.local\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example1\"]}],\"labels\":{},\"tls\":[]},\"ingressInternal\":{\"annotations\":{},\"className\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example1\"]},{\"host\":\"chart-example2.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example2\",\"/example2/healthz\"]}],\"tls\":[]},\"initContainers\":[],\"istio\":{\"enable\":false,\"gateway\":{\"annotations\":{},\"enabled\":false,\"host\":\"example.com\",\"labels\":{},\"tls\":{\"enabled\":false,\"secretName\":\"secret-name\"}},\"virtualService\":{\"annotations\":{},\"enabled\":false,\"gateways\":[],\"hosts\":[],\"http\":[{\"corsPolicy\":{},\"headers\":{},\"match\":[{\"uri\":{\"prefix\":\"/v1\"}},{\"uri\":{\"prefix\":\"/v2\"}}],\"retries\":{\"attempts\":2,\"perTryTimeout\":\"3s\"},\"rewriteUri\":\"/\",\"route\":[{\"destination\":{\"host\":\"service1\",\"port\":80}}],\"timeout\":\"12s\"},{\"route\":[{\"destination\":{\"host\":\"service2\"}}]}],\"labels\":{}}},\"kedaAutoscaling\":{\"advanced\":{},\"authenticationRef\":{},\"cooldownPeriod\":300,\"enabled\":false,\"envSourceContainerName\":\"\",\"fallback\":{},\"idleReplicaCount\":0,\"maxReplicaCount\":2,\"minReplicaCount\":1,\"pollingInterval\":30,\"triggerAuthentication\":{\"enabled\":false,\"name\":\"\",\"spec\":{}},\"triggers\":[]},\"nodeSelector\":{},\"orchestrator.deploymant.algo\":1,\"pauseForSecondsBeforeSwitchActive\":30,\"podAnnotations\":{},\"podDisruptionBudget\":{},\"podExtraSpecs\":{},\"podLabels\":{},\"podSecurityContext\":{},\"prometheus\":{\"release\":\"monitoring\"},\"prometheusRule\":{\"additionalLabels\":{},\"enabled\":false,\"namespace\":\"\"},\"rawYaml\":[],\"replicaCount\":1,\"resources\":{\"limits\":{\"cpu\":\"0.05\",\"memory\":\"50Mi\"},\"requests\":{\"cpu\":\"0.01\",\"memory\":\"10Mi\"}},\"rolloutAnnotations\":{},\"rolloutLabels\":{},\"secret\":{\"data\":{},\"enabled\":false},\"server\":{\"deployment\":{\"image\":\"\",\"image_tag\":\"1-95af053\"}},\"service\":{\"annotations\":{},\"loadBalancerSourceRanges\":[],\"type\":\"ClusterIP\"},\"serviceAccount\":{\"annotations\":{},\"create\":false,\"name\":\"\"},\"servicemonitor\":{\"additionalLabels\":{}},\"tolerations\":[],\"topologySpreadConstraints\":[],\"volumeMounts\":[],\"volumes\":[],\"waitForSecondsBeforeScalingDown\":30}",
			GlobalOverride:          "{\"ContainerPort\":[{\"envoyPort\":8799,\"idleTimeout\":\"1800s\",\"name\":\"app\",\"port\":8080,\"servicePort\":80,\"supportStreaming\":false,\"useHTTP2\":false}],\"EnvVariables\":[],\"GracePeriod\":30,\"LivenessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"MaxSurge\":1,\"MaxUnavailable\":0,\"MinReadySeconds\":60,\"ReadinessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"Spec\":{\"Affinity\":{\"Key\":null,\"Values\":\"nodes\",\"key\":\"\"}},\"ambassadorMapping\":{\"ambassadorId\":\"\",\"cors\":{},\"enabled\":false,\"hostname\":\"devtron.example.com\",\"labels\":{},\"prefix\":\"/\",\"retryPolicy\":{},\"rewrite\":\"\",\"tls\":{\"context\":\"\",\"create\":false,\"hosts\":[],\"secretName\":\"\"}},\"args\":{\"enabled\":false,\"value\":[\"/bin/sh\",\"-c\",\"touch /tmp/healthy; sleep 30; rm -rf /tmp/healthy; sleep 600\"]},\"autoscaling\":{\"MaxReplicas\":2,\"MinReplicas\":1,\"TargetCPUUtilizationPercentage\":90,\"TargetMemoryUtilizationPercentage\":80,\"annotations\":{},\"behavior\":{},\"enabled\":false,\"extraMetrics\":[],\"labels\":{}},\"command\":{\"enabled\":false,\"value\":[],\"workingDir\":{}},\"containerSecurityContext\":{},\"containerSpec\":{\"lifecycle\":{\"enabled\":false,\"postStart\":{\"httpGet\":{\"host\":\"example.com\",\"path\":\"/example\",\"port\":90}},\"preStop\":{\"exec\":{\"command\":[\"sleep\",\"10\"]}}}},\"containers\":[],\"dbMigrationConfig\":{\"enabled\":false},\"envoyproxy\":{\"configMapName\":\"\",\"image\":\"quay.io/devtron/envoy:v1.14.1\",\"lifecycle\":{},\"resources\":{\"limits\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"},\"requests\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"}}},\"hostAliases\":[],\"image\":{\"pullPolicy\":\"IfNotPresent\"},\"imagePullSecrets\":[],\"ingress\":{\"annotations\":{},\"className\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.local\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example1\"]}],\"labels\":{},\"tls\":[]},\"ingressInternal\":{\"annotations\":{},\"className\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example1\"]},{\"host\":\"chart-example2.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example2\",\"/example2/healthz\"]}],\"tls\":[]},\"initContainers\":[],\"istio\":{\"enable\":false,\"gateway\":{\"annotations\":{},\"enabled\":false,\"host\":\"example.com\",\"labels\":{},\"tls\":{\"enabled\":false,\"secretName\":\"secret-name\"}},\"virtualService\":{\"annotations\":{},\"enabled\":false,\"gateways\":[],\"hosts\":[],\"http\":[{\"corsPolicy\":{},\"headers\":{},\"match\":[{\"uri\":{\"prefix\":\"/v1\"}},{\"uri\":{\"prefix\":\"/v2\"}}],\"retries\":{\"attempts\":2,\"perTryTimeout\":\"3s\"},\"rewriteUri\":\"/\",\"route\":[{\"destination\":{\"host\":\"service1\",\"port\":80}}],\"timeout\":\"12s\"},{\"route\":[{\"destination\":{\"host\":\"service2\"}}]}],\"labels\":{}}},\"kedaAutoscaling\":{\"advanced\":{},\"authenticationRef\":{},\"enabled\":false,\"envSourceContainerName\":\"\",\"maxReplicaCount\":2,\"minReplicaCount\":1,\"triggerAuthentication\":{\"enabled\":false,\"name\":\"\",\"spec\":{}},\"triggers\":[]},\"pauseForSecondsBeforeSwitchActive\":30,\"podAnnotations\":{},\"podLabels\":{},\"podSecurityContext\":{},\"prometheus\":{\"release\":\"monitoring\"},\"rawYaml\":[],\"replicaCount\":1,\"resources\":{\"limits\":{\"cpu\":\"0.05\",\"memory\":\"50Mi\"},\"requests\":{\"cpu\":\"0.01\",\"memory\":\"10Mi\"}},\"rolloutAnnotations\":{},\"rolloutLabels\":{},\"secret\":{\"data\":{},\"enabled\":false},\"server\":{\"deployment\":{\"image\":\"\",\"image_tag\":\"1-95af053\"}},\"service\":{\"annotations\":{},\"loadBalancerSourceRanges\":[],\"type\":\"ClusterIP\"},\"serviceAccount\":{\"annotations\":{},\"create\":false,\"name\":\"\"},\"servicemonitor\":{\"additionalLabels\":{}},\"tolerations\":[],\"topologySpreadConstraints\":[],\"volumeMounts\":[],\"volumes\":[],\"waitForSecondsBeforeScalingDown\":30}",
			ReleaseOverride:         "{\"autoPromotionSeconds\":30,\"dbMigrationConfig\":{\"enabled\":false},\"orchestrator.deploymant.algo\":1,\"pauseForSecondsBeforeSwitchActive\":0,\"server\":{\"deployment\":{\"enabled\":false,\"image\":\"IMAGE_REPO\",\"image_tag\":\"IMAGE_TAG\"}},\"waitForSecondsBeforeScalingDown\":0}",
			PipelineOverride:        "{\"deployment\":{\"strategy\":{\"blueGreen\":{\"autoPromotionEnabled\":false,\"autoPromotionSeconds\":30,\"previewReplicaCount\":1,\"scaleDownDelaySeconds\":30},\"canary\":{\"maxSurge\":\"25%\",\"maxUnavailable\":1,\"steps\":[{\"setWeight\":25},{\"pause\":{\"duration\":15}},{\"setWeight\":50},{\"pause\":{\"duration\":15}},{\"setWeight\":75},{\"pause\":{\"duration\":15}}]},\"recreate\":{},\"rolling\":{\"maxSurge\":\"25%\",\"maxUnavailable\":1}}}}",
			Status:                  0,
			Active:                  false,
			GitRepoUrl:              "",
			ChartLocation:           "reference-chart_4-17-0/4.17.1",
			ReferenceTemplate:       "reference-chart_4-17-0",
			ImageDescriptorTemplate: "{\"server\":{\"deployment\":{\"image_tag\":\"{{.Tag}}\",\"image\":\"{{.Name}}\"}},\"pipelineName\": \"{{.PipelineName}}\",\"releaseVersion\":\"{{.ReleaseVersion}}\",\"deploymentType\": \"{{.DeploymentType}}\", \"app\": \"{{.App}}\", \"env\": \"{{.Env}}\", \"appMetrics\": {{.AppMetrics}}}",
			ChartRefId:              29,
			Latest:                  true,
			Previous:                false,
			ReferenceChart:          nil,
			IsBasicViewLocked:       false,
			CurrentViewEditor:       "",
			AuditLog:                sql.AuditLog{},
		}

		mockedChartRepository := mocks2.NewChartRepository(t)
		mockedChartRepository.On("FindLatestChartForAppByAppId", 1).
			Return(mockedChartOutput, nil)

		mockedEnvConfigOverrideRepository.On("FindChartByAppIdAndEnvIdAndChartRefId", 1, 1, 29).
			Return(nil, nil)

		mockedEnvironmentRepository := mocks4.NewEnvironmentRepository(t)
		mockedEnvironmentRepository.On("FindById", 1).
			Return(&repository2.Environment{
				Id:                    1,
				Name:                  "devtron-demo",
				ClusterId:             1,
				Active:                true,
				Default:               false,
				GrafanaDatasourceId:   0,
				Namespace:             "default",
				EnvironmentIdentifier: "devtron-demo",
				Description:           "",
				IsVirtualEnvironment:  false,
				AuditLog:              sql.AuditLog{},
			}, nil)

		triggeredAt := time.Now()
		mockedEnvConfigOverrideRepository.On("Save", &chartConfig.EnvConfigOverride{
			Active:            true,
			ManualReviewed:    true,
			Status:            models.CHARTSTATUS_SUCCESS,
			TargetEnvironment: overrideRequest.EnvId,
			ChartId:           mockedChartOutput.Id,
			AuditLog:          sql.AuditLog{UpdatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId},
			Namespace:         "default",
			IsOverride:        false,
			EnvOverrideValues: "{}",
			Latest:            false,
			IsBasicViewLocked: mockedChartOutput.IsBasicViewLocked,
			CurrentViewEditor: mockedChartOutput.CurrentViewEditor,
		}).Return(nil)

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		appServiceImpl := app.NewAppService(mockedEnvConfigOverrideRepository,
			nil, nil, sugaredLogger,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, mockedEnvironmentRepository, nil,
			nil, nil, nil,
			mockedChartRepository, nil, nil,
			nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, "", nil,
			nil, nil, nil, nil,
			nil, nil, nil,
			nil, nil,
			nil, nil,
			nil, nil,
			nil, nil, nil,
			nil, nil,
			nil, nil, nil)

		envOverride, err := appServiceImpl.GetEnvOverrideByTriggerType(overrideRequest, triggeredAt, context.Background())
		assert.Nil(t, err)
		assert.NotNil(t, envOverride.Environment)
		assert.Equal(t, envOverride.Chart, mockedChartOutput)
		assert.Equal(t, envOverride.ChartId, mockedChartOutput.Id)
		assert.Equal(t, envOverride.IsOverride, false)

	})

	t.Run("GetAppLevelMetricsWhenDeploymentConfigTypeIsSpecificTriggerAndAppMetricsDisabled", func(t *testing.T) {
		mockedDeploymentTemplateHistoryRepository := mocks.NewDeploymentTemplateHistoryRepository(t)

		deploymentTemplateHistoryResponse := &repository.DeploymentTemplateHistory{
			Id:                      1,
			PipelineId:              1,
			AppId:                   1,
			ImageDescriptorTemplate: "{\"server\":{\"deployment\":{\"image_tag\":\"{{.Tag}}\",\"image\":\"{{.Name}}\"}},\"pipelineName\": \"{{.PipelineName}}\",\"releaseVersion\":\"{{.ReleaseVersion}}\",\"deploymentType\": \"{{.Deploymen\ntType}}\", \"app\": \"{{.App}}\", \"env\": \"{{.Env}}\", \"appMetrics\": {{.AppMetrics}}}",
			Template:                "{\"ContainerPort\":[{\"envoyPort\":8799,\"idleTimeout\":\"1800s\",\"name\":\"app\",\"port\":8080,\"servicePort\":80,\"supportStreaming\":false,\"useHTTP2\":false}],\"EnvVariables\":[],\"GracePeriod\":\n30,\"LivenessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"M\naxSurge\":1,\"MaxUnavailable\":0,\"MinReadySeconds\":60,\"ReadinessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"succe\nssThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"Spec\":{\"Affinity\":{\"Key\":null,\"Values\":\"nodes\",\"key\":\"\"}},\"ambassadorMapping\":{\"ambassadorId\":\"\",\"cors\":{},\"enabled\":false,\"hostname\":\"devtron.example.com\",\n\"labels\":{},\"prefix\":\"/\",\"retryPolicy\":{},\"rewrite\":\"\",\"tls\":{\"context\":\"\",\"create\":false,\"hosts\":[],\"secretName\":\"\"}},\"args\":{\"enabled\":false,\"value\":[\"/bin/sh\",\"-c\",\"touch /tmp/healthy; sleep 30; rm -rf\n /tmp/healthy; sleep 600\"]},\"autoscaling\":{\"MaxReplicas\":2,\"MinReplicas\":1,\"TargetCPUUtilizationPercentage\":90,\"TargetMemoryUtilizationPercentage\":80,\"annotations\":{},\"behavior\":{},\"enabled\":false,\"extraM\netrics\":[],\"labels\":{}},\"command\":{\"enabled\":false,\"value\":[],\"workingDir\":{}},\"containerSecurityContext\":{},\"containerSpec\":{\"lifecycle\":{\"enabled\":false,\"postStart\":{\"httpGet\":{\"host\":\"example.com\",\"pat\nh\":\"/example\",\"port\":90}},\"preStop\":{\"exec\":{\"command\":[\"sleep\",\"10\"]}}}},\"containers\":[],\"dbMigrationConfig\":{\"enabled\":false},\"envoyproxy\":{\"configMapName\":\"\",\"image\":\"quay.io/devtron/envoy:v1.14.1\",\"li\nfecycle\":{},\"resources\":{\"limits\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"},\"requests\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"}}},\"hostAliases\":[],\"image\":{\"pullPolicy\":\"IfNotPresent\"},\"imagePullSecrets\":[],\"ingress\":{\"annotati\nons\":{},\"className\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.local\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example1\"]}],\"labels\":{},\"tls\":[]},\"ingressInternal\":{\"annotations\":{},\"classN\name\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example1\"]},{\"host\":\"chart-example2.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":\n[\"/example2\",\"/example2/healthz\"]}],\"tls\":[]},\"initContainers\":[],\"istio\":{\"enable\":false,\"gateway\":{\"annotations\":{},\"enabled\":false,\"host\":\"example.com\",\"labels\":{},\"tls\":{\"enabled\":false,\"secretName\":\"\nsecret-name\"}},\"virtualService\":{\"annotations\":{},\"enabled\":false,\"gateways\":[],\"hosts\":[],\"http\":[{\"corsPolicy\":{},\"headers\":{},\"match\":[{\"uri\":{\"prefix\":\"/v1\"}},{\"uri\":{\"prefix\":\"/v2\"}}],\"retries\":{\"att\nempts\":2,\"perTryTimeout\":\"3s\"},\"rewriteUri\":\"/\",\"route\":[{\"destination\":{\"host\":\"service1\",\"port\":80}}],\"timeout\":\"12s\"},{\"route\":[{\"destination\":{\"host\":\"service2\"}}]}],\"labels\":{}}},\"kedaAutoscaling\":{\"\nadvanced\":{},\"authenticationRef\":{},\"enabled\":false,\"envSourceContainerName\":\"\",\"maxReplicaCount\":2,\"minReplicaCount\":1,\"triggerAuthentication\":{\"enabled\":false,\"name\":\"\",\"spec\":{}},\"triggers\":[]},\"pauseF\norSecondsBeforeSwitchActive\":30,\"podAnnotations\":{},\"podLabels\":{},\"podSecurityContext\":{},\"prometheus\":{\"release\":\"monitoring\"},\"rawYaml\":[],\"replicaCount\":1,\"resources\":{\"limits\":{\"cpu\":\"0.05\",\"memory\":\n\"50Mi\"},\"requests\":{\"cpu\":\"0.01\",\"memory\":\"10Mi\"}},\"rolloutAnnotations\":{},\"rolloutLabels\":{},\"secret\":{\"data\":{},\"enabled\":false},\"server\":{\"deployment\":{\"image\":\"\",\"image_tag\":\"1-95af053\"}},\"service\":{\"\nannotations\":{},\"loadBalancerSourceRanges\":[],\"type\":\"ClusterIP\"},\"serviceAccount\":{\"annotations\":{},\"create\":false,\"name\":\"\"},\"servicemonitor\":{\"additionalLabels\":{}},\"tolerations\":[],\"topologySpreadCons\ntraints\":[],\"volumeMounts\":[],\"volumes\":[],\"waitForSecondsBeforeScalingDown\":30}\n",
			TargetEnvironment:       1,
			TemplateName:            "Rollout Deployment",
			TemplateVersion:         "4.17.0",
			IsAppMetricsEnabled:     false,
			Deployed:                true,
			DeployedOn:              time.Now(),
			DeployedBy:              1,
			AuditLog:                sql.AuditLog{},
		}
		mockedDeploymentTemplateHistoryRepository.On("GetHistoryByPipelineIdAndWfrId", 1, 1).
			Return(deploymentTemplateHistoryResponse, nil)

		overrideRequest := &bean.ValuesOverrideRequest{
			PipelineId:                            1,
			AppId:                                 1,
			CiArtifactId:                          1,
			AdditionalOverride:                    nil,
			TargetDbVersion:                       0,
			ForceTrigger:                          false,
			DeploymentTemplate:                    "",
			DeploymentWithConfig:                  bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER,
			WfrIdForDeploymentWithSpecificTrigger: 1,
			CdWorkflowType:                        "deploy",
			WfrId:                                 1,
			CdWorkflowId:                          1,
			UserId:                                1,
			DeploymentType:                        models.DEPLOYMENTTYPE_DEPLOY,
			EnvId:                                 1,
			EnvName:                               "",
			ClusterId:                             0,
			AppName:                               "1",
			PipelineName:                          "test",
			DeploymentAppType:                     "", //should work independent of deployment type
		}

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		appServiceImpl := app.NewAppService(nil,
			nil, nil, sugaredLogger,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, "", nil,
			nil, nil, nil, nil,
			nil, nil, nil,
			mockedDeploymentTemplateHistoryRepository, nil,
			nil, nil,
			nil, nil,
			nil, nil, nil,
			nil, nil,
			nil, nil, nil)

		isAppMetricsEnabled, err := appServiceImpl.GetAppMetricsByTriggerType(overrideRequest, context.Background())
		assert.Nil(t, err)
		assert.Equal(t, isAppMetricsEnabled, false)

	})

	t.Run("GetAppLevelMetricsWhenDeploymentConfigTypeIsSpecificTriggerAndAppMetricsEnabled", func(t *testing.T) {
		mockedDeploymentTemplateHistoryRepository := mocks.NewDeploymentTemplateHistoryRepository(t)

		deploymentTemplateHistoryResponse := &repository.DeploymentTemplateHistory{
			Id:                      1,
			PipelineId:              1,
			AppId:                   1,
			ImageDescriptorTemplate: "{\"server\":{\"deployment\":{\"image_tag\":\"{{.Tag}}\",\"image\":\"{{.Name}}\"}},\"pipelineName\": \"{{.PipelineName}}\",\"releaseVersion\":\"{{.ReleaseVersion}}\",\"deploymentType\": \"{{.Deploymen\ntType}}\", \"app\": \"{{.App}}\", \"env\": \"{{.Env}}\", \"appMetrics\": {{.AppMetrics}}}",
			Template:                "{\"ContainerPort\":[{\"envoyPort\":8799,\"idleTimeout\":\"1800s\",\"name\":\"app\",\"port\":8080,\"servicePort\":80,\"supportStreaming\":false,\"useHTTP2\":false}],\"EnvVariables\":[],\"GracePeriod\":\n30,\"LivenessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"successThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"M\naxSurge\":1,\"MaxUnavailable\":0,\"MinReadySeconds\":60,\"ReadinessProbe\":{\"Path\":\"\",\"command\":[],\"failureThreshold\":3,\"httpHeaders\":[],\"initialDelaySeconds\":20,\"periodSeconds\":10,\"port\":8080,\"scheme\":\"\",\"succe\nssThreshold\":1,\"tcp\":false,\"timeoutSeconds\":5},\"Spec\":{\"Affinity\":{\"Key\":null,\"Values\":\"nodes\",\"key\":\"\"}},\"ambassadorMapping\":{\"ambassadorId\":\"\",\"cors\":{},\"enabled\":false,\"hostname\":\"devtron.example.com\",\n\"labels\":{},\"prefix\":\"/\",\"retryPolicy\":{},\"rewrite\":\"\",\"tls\":{\"context\":\"\",\"create\":false,\"hosts\":[],\"secretName\":\"\"}},\"args\":{\"enabled\":false,\"value\":[\"/bin/sh\",\"-c\",\"touch /tmp/healthy; sleep 30; rm -rf\n /tmp/healthy; sleep 600\"]},\"autoscaling\":{\"MaxReplicas\":2,\"MinReplicas\":1,\"TargetCPUUtilizationPercentage\":90,\"TargetMemoryUtilizationPercentage\":80,\"annotations\":{},\"behavior\":{},\"enabled\":false,\"extraM\netrics\":[],\"labels\":{}},\"command\":{\"enabled\":false,\"value\":[],\"workingDir\":{}},\"containerSecurityContext\":{},\"containerSpec\":{\"lifecycle\":{\"enabled\":false,\"postStart\":{\"httpGet\":{\"host\":\"example.com\",\"pat\nh\":\"/example\",\"port\":90}},\"preStop\":{\"exec\":{\"command\":[\"sleep\",\"10\"]}}}},\"containers\":[],\"dbMigrationConfig\":{\"enabled\":false},\"envoyproxy\":{\"configMapName\":\"\",\"image\":\"quay.io/devtron/envoy:v1.14.1\",\"li\nfecycle\":{},\"resources\":{\"limits\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"},\"requests\":{\"cpu\":\"50m\",\"memory\":\"50Mi\"}}},\"hostAliases\":[],\"image\":{\"pullPolicy\":\"IfNotPresent\"},\"imagePullSecrets\":[],\"ingress\":{\"annotati\nons\":{},\"className\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.local\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example1\"]}],\"labels\":{},\"tls\":[]},\"ingressInternal\":{\"annotations\":{},\"classN\name\":\"\",\"enabled\":false,\"hosts\":[{\"host\":\"chart-example1.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":[\"/example1\"]},{\"host\":\"chart-example2.internal\",\"pathType\":\"ImplementationSpecific\",\"paths\":\n[\"/example2\",\"/example2/healthz\"]}],\"tls\":[]},\"initContainers\":[],\"istio\":{\"enable\":false,\"gateway\":{\"annotations\":{},\"enabled\":false,\"host\":\"example.com\",\"labels\":{},\"tls\":{\"enabled\":false,\"secretName\":\"\nsecret-name\"}},\"virtualService\":{\"annotations\":{},\"enabled\":false,\"gateways\":[],\"hosts\":[],\"http\":[{\"corsPolicy\":{},\"headers\":{},\"match\":[{\"uri\":{\"prefix\":\"/v1\"}},{\"uri\":{\"prefix\":\"/v2\"}}],\"retries\":{\"att\nempts\":2,\"perTryTimeout\":\"3s\"},\"rewriteUri\":\"/\",\"route\":[{\"destination\":{\"host\":\"service1\",\"port\":80}}],\"timeout\":\"12s\"},{\"route\":[{\"destination\":{\"host\":\"service2\"}}]}],\"labels\":{}}},\"kedaAutoscaling\":{\"\nadvanced\":{},\"authenticationRef\":{},\"enabled\":false,\"envSourceContainerName\":\"\",\"maxReplicaCount\":2,\"minReplicaCount\":1,\"triggerAuthentication\":{\"enabled\":false,\"name\":\"\",\"spec\":{}},\"triggers\":[]},\"pauseF\norSecondsBeforeSwitchActive\":30,\"podAnnotations\":{},\"podLabels\":{},\"podSecurityContext\":{},\"prometheus\":{\"release\":\"monitoring\"},\"rawYaml\":[],\"replicaCount\":1,\"resources\":{\"limits\":{\"cpu\":\"0.05\",\"memory\":\n\"50Mi\"},\"requests\":{\"cpu\":\"0.01\",\"memory\":\"10Mi\"}},\"rolloutAnnotations\":{},\"rolloutLabels\":{},\"secret\":{\"data\":{},\"enabled\":false},\"server\":{\"deployment\":{\"image\":\"\",\"image_tag\":\"1-95af053\"}},\"service\":{\"\nannotations\":{},\"loadBalancerSourceRanges\":[],\"type\":\"ClusterIP\"},\"serviceAccount\":{\"annotations\":{},\"create\":false,\"name\":\"\"},\"servicemonitor\":{\"additionalLabels\":{}},\"tolerations\":[],\"topologySpreadCons\ntraints\":[],\"volumeMounts\":[],\"volumes\":[],\"waitForSecondsBeforeScalingDown\":30}\n",
			TargetEnvironment:       1,
			TemplateName:            "Rollout Deployment",
			TemplateVersion:         "4.17.0",
			IsAppMetricsEnabled:     true,
			Deployed:                true,
			DeployedOn:              time.Now(),
			DeployedBy:              1,
			AuditLog:                sql.AuditLog{},
		}
		mockedDeploymentTemplateHistoryRepository.On("GetHistoryByPipelineIdAndWfrId", 1, 1).
			Return(deploymentTemplateHistoryResponse, nil)

		overrideRequest := &bean.ValuesOverrideRequest{
			PipelineId:                            1,
			AppId:                                 1,
			CiArtifactId:                          1,
			AdditionalOverride:                    nil,
			TargetDbVersion:                       0,
			ForceTrigger:                          false,
			DeploymentTemplate:                    "",
			DeploymentWithConfig:                  bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER,
			WfrIdForDeploymentWithSpecificTrigger: 1,
			CdWorkflowType:                        "deploy",
			WfrId:                                 1,
			CdWorkflowId:                          1,
			UserId:                                1,
			DeploymentType:                        models.DEPLOYMENTTYPE_DEPLOY,
			EnvId:                                 1,
			EnvName:                               "",
			ClusterId:                             0,
			AppName:                               "1",
			PipelineName:                          "test",
			DeploymentAppType:                     "", //should work independent of deployment type
		}

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		appServiceImpl := app.NewAppService(nil,
			nil, nil, sugaredLogger,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, "", nil,
			nil, nil, nil, nil,
			nil, nil, nil,
			mockedDeploymentTemplateHistoryRepository, nil,
			nil, nil,
			nil, nil,
			nil, nil, nil,
			nil, nil,
			nil, nil, nil)

		isAppMetricsEnabled, err := appServiceImpl.GetAppMetricsByTriggerType(overrideRequest, context.Background())
		assert.Nil(t, err)
		assert.Equal(t, isAppMetricsEnabled, true)

	})

	t.Run("GetAppLevelMetricsWhenDeploymentConfigTypeIsDeploymentConfigLastSavedAndEnvOverride", func(t *testing.T) {

		overrideRequest := &bean.ValuesOverrideRequest{
			PipelineId:                            1,
			AppId:                                 1,
			CiArtifactId:                          1,
			AdditionalOverride:                    nil,
			TargetDbVersion:                       0,
			ForceTrigger:                          false,
			DeploymentTemplate:                    "",
			DeploymentWithConfig:                  bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED,
			WfrIdForDeploymentWithSpecificTrigger: 1,
			CdWorkflowType:                        "deploy",
			WfrId:                                 1,
			CdWorkflowId:                          1,
			UserId:                                1,
			DeploymentType:                        models.DEPLOYMENTTYPE_DEPLOY,
			EnvId:                                 1,
			EnvName:                               "",
			ClusterId:                             0,
			AppName:                               "1",
			PipelineName:                          "test",
			DeploymentAppType:                     "", //should work independent of deployment type
		}

		mockedAppLevelMetricsRepository := mocks5.NewAppLevelMetricsRepository(t)

		appLevelMetricsDBObject := &repository3.AppLevelMetrics{
			Id:           1,
			AppId:        1,
			AppMetrics:   false,
			InfraMetrics: false,
			AuditLog:     sql.AuditLog{},
		}
		mockedAppLevelMetricsRepository.On("FindByAppId", 1).
			Return(appLevelMetricsDBObject, nil)

		mockedEnvLevelMetricsRepository := mocks5.NewEnvLevelAppMetricsRepository(t)

		appMetrics := true

		mockedEnvLevelMetricsDBObject := &repository3.EnvLevelAppMetrics{
			Id:           1,
			AppId:        1,
			EnvId:        1,
			AppMetrics:   &appMetrics,
			InfraMetrics: nil,
			AuditLog:     sql.AuditLog{},
		}

		mockedEnvLevelMetricsRepository.On("FindByAppIdAndEnvId", 1, 1).
			Return(mockedEnvLevelMetricsDBObject, nil)

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		appServiceImpl := app.NewAppService(nil,
			nil, nil, sugaredLogger,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, mockedAppLevelMetricsRepository, mockedEnvLevelMetricsRepository,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, "", nil,
			nil, nil, nil, nil,
			nil, nil, nil,
			nil, nil,
			nil, nil,
			nil, nil,
			nil, nil, nil,
			nil, nil,
			nil, nil, nil)

		isAppMetricsEnabled, err := appServiceImpl.GetAppMetricsByTriggerType(overrideRequest, context.Background())
		assert.Nil(t, err)
		assert.Equal(t, isAppMetricsEnabled, true)

	})

	t.Run("GetAppLevelMetricsWhenDeploymentConfigTypeIsDeploymentConfigLastSavedAndEnvNotOverride", func(t *testing.T) {

		overrideRequest := &bean.ValuesOverrideRequest{
			PipelineId:                            1,
			AppId:                                 1,
			CiArtifactId:                          1,
			AdditionalOverride:                    nil,
			TargetDbVersion:                       0,
			ForceTrigger:                          false,
			DeploymentTemplate:                    "",
			DeploymentWithConfig:                  bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED,
			WfrIdForDeploymentWithSpecificTrigger: 1,
			CdWorkflowType:                        "deploy",
			WfrId:                                 1,
			CdWorkflowId:                          1,
			UserId:                                1,
			DeploymentType:                        models.DEPLOYMENTTYPE_DEPLOY,
			EnvId:                                 1,
			EnvName:                               "",
			ClusterId:                             0,
			AppName:                               "1",
			PipelineName:                          "test",
			DeploymentAppType:                     "", //should work independent of deployment type
		}

		mockedAppLevelMetricsRepository := mocks5.NewAppLevelMetricsRepository(t)

		appLevelMetricsDBObject := &repository3.AppLevelMetrics{
			Id:           1,
			AppId:        1,
			AppMetrics:   false,
			InfraMetrics: false,
			AuditLog:     sql.AuditLog{},
		}
		mockedAppLevelMetricsRepository.On("FindByAppId", 1).
			Return(appLevelMetricsDBObject, nil)

		mockedEnvLevelMetricsRepository := mocks5.NewEnvLevelAppMetricsRepository(t)

		mockedEnvLevelMetricsDBObject := &repository3.EnvLevelAppMetrics{
			Id:           0,
			AppId:        0,
			EnvId:        0,
			AppMetrics:   nil,
			InfraMetrics: nil,
			AuditLog:     sql.AuditLog{},
		}

		mockedEnvLevelMetricsRepository.On("FindByAppIdAndEnvId", 1, 1).
			Return(mockedEnvLevelMetricsDBObject, nil)

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		appServiceImpl := app.NewAppService(nil,
			nil, nil, sugaredLogger,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, mockedAppLevelMetricsRepository, mockedEnvLevelMetricsRepository,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, "", nil,
			nil, nil, nil, nil,
			nil, nil, nil,
			nil, nil,
			nil, nil,
			nil, nil,
			nil, nil, nil,
			nil, nil,
			nil, nil, nil)

		isAppMetricsEnabled, err := appServiceImpl.GetAppMetricsByTriggerType(overrideRequest, context.Background())
		assert.Nil(t, err)
		assert.Equal(t, isAppMetricsEnabled, false)

	})

	t.Run("GetPipelineStrategyWhenDeploymentConfigTypeIsDeploymentWithSpecificTrigger", func(t *testing.T) {

		mockedStrategyRepository := mocks.NewPipelineStrategyHistoryRepository(t)

		pipelineStrategyHistoryRepository := &repository.PipelineStrategyHistory{
			TableName:           struct{}{},
			Id:                  1,
			PipelineId:          1,
			Strategy:            "ROLLING",
			Config:              "{\"deployment\":{\"strategy\":{\"rolling\":{\"maxSurge\":\"25%\",\"maxUnavailable\":1}}}}",
			Default:             false,
			Deployed:            false,
			DeployedOn:          time.Now(),
			DeployedBy:          1,
			PipelineTriggerType: "AUTOMATIC",
			AuditLog:            sql.AuditLog{},
		}

		mockedStrategyRepository.On("GetHistoryByPipelineIdAndWfrId", 1, 1).
			Return(pipelineStrategyHistoryRepository, nil)

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)

		appServiceImpl := app.NewAppService(nil,
			nil, nil, sugaredLogger,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, "", nil,
			nil, nil, nil, nil,
			nil, nil, mockedStrategyRepository,
			nil, nil,
			nil, nil,
			nil, nil,
			nil, nil, nil,
			nil, nil,
			nil, nil, nil)

		overrideRequest := &bean.ValuesOverrideRequest{
			PipelineId:                            1,
			AppId:                                 1,
			CiArtifactId:                          1,
			AdditionalOverride:                    nil,
			TargetDbVersion:                       0,
			ForceTrigger:                          false,
			DeploymentTemplate:                    "",
			DeploymentWithConfig:                  bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER,
			WfrIdForDeploymentWithSpecificTrigger: 1,
			CdWorkflowType:                        "deploy",
			WfrId:                                 1,
			CdWorkflowId:                          1,
			UserId:                                1,
			DeploymentType:                        models.DEPLOYMENTTYPE_DEPLOY,
			EnvId:                                 1,
			EnvName:                               "",
			ClusterId:                             0,
			AppName:                               "1",
			PipelineName:                          "test",
			DeploymentAppType:                     "", //should work independent of deployment type
		}

		strategy, err := appServiceImpl.GetDeploymentStrategyByTriggerType(overrideRequest, context.Background())

		assert.Nil(t, err)
		assert.Equal(t, strategy.Config, pipelineStrategyHistoryRepository.Config)
		assert.Equal(t, strategy.Strategy, pipelineStrategyHistoryRepository.Strategy)
		assert.Equal(t, strategy.PipelineId, pipelineStrategyHistoryRepository.PipelineId)

	})

	t.Run("GetPipelineStrategyWhenDeploymentConfigLastSavedIfForceTriggerTrue", func(t *testing.T) {

		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)
		mockedPipelineConfigRepository := mocks3.NewPipelineConfigRepository(t)

		overrideRequest := &bean.ValuesOverrideRequest{
			PipelineId:                            1,
			AppId:                                 1,
			CiArtifactId:                          1,
			AdditionalOverride:                    nil,
			TargetDbVersion:                       0,
			ForceTrigger:                          false,
			DeploymentWithConfig:                  bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED,
			WfrIdForDeploymentWithSpecificTrigger: 1,
			CdWorkflowType:                        "deploy",
			WfrId:                                 1,
			CdWorkflowId:                          1,
			UserId:                                1,
			DeploymentType:                        models.DEPLOYMENTTYPE_DEPLOY,
			EnvId:                                 1,
			EnvName:                               "",
			ClusterId:                             0,
			AppName:                               "1",
			PipelineName:                          "test",
			DeploymentAppType:                     "", //should work independent of deployment type
			DeploymentTemplate:                    "ROLLING",
		}

		pipelineStrategy := &chartConfig.PipelineStrategy{
			Id:         1,
			PipelineId: 1,
			Strategy:   "ROLLING",
			Config:     "{\"deployment\":{\"strategy\":{\"rolling\":{\"maxSurge\":\"25%\",\"maxUnavailable\":1}}}}",
			Default:    true,
			Deleted:    false,
			AuditLog:   sql.AuditLog{},
		}

		//mockedPipelineConfigRepository.On("GetDefaultStrategyByPipelineId", 1).
		//	Return(pipelineStrategy, nil)

		mockedPipelineConfigRepository.On("FindByStrategyAndPipelineId", chartRepoRepository.DEPLOYMENT_STRATEGY_ROLLING, 1).
			Return(pipelineStrategy, nil)

		appServiceImpl := app.NewAppService(nil,
			nil, nil, sugaredLogger,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, mockedPipelineConfigRepository,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, "", nil,
			nil, nil, nil, nil,
			nil, nil, nil,
			nil, nil,
			nil, nil,
			nil, nil,
			nil, nil, nil,
			nil, nil,
			nil, nil, nil)

		strategy, err := appServiceImpl.GetDeploymentStrategyByTriggerType(overrideRequest, context.Background())

		assert.Nil(t, err)
		assert.Equal(t, string(strategy.Strategy), overrideRequest.DeploymentTemplate)
	})

}
