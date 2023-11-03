package pipeline

import (
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/devtron-labs/authenticator/client"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/appStatus"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/cluster"
	repository3 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/commonService"
	k8s2 "github.com/devtron-labs/devtron/pkg/k8s"
	"github.com/devtron-labs/devtron/pkg/k8s/informer"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/stretchr/testify/assert"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"net/url"
	"reflect"
	"testing"
)

var pipelineId = 0

var cmManifest = "{\"kind\":\"ConfigMap\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"cm-5-ci\",\"creationTimestamp\":null,\"ownerReferences\":[{\"apiVersion\":\"argoproj.io/v1alpha1\",\"kind\":\"Workflow\",\"name\":\"{{workflow.name}}\",\"uid\":\"{{workflow.uid}}\",\"blockOwnerDeletion\":true}]},\"data\":{\"key\":\"value\"}}"
var secManifest = "{\"kind\":\"Secret\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"secret-5-ci\",\"creationTimestamp\":null,\"ownerReferences\":[{\"apiVersion\":\"argoproj.io/v1alpha1\",\"kind\":\"Workflow\",\"name\":\"{{workflow.name}}\",\"uid\":\"{{workflow.uid}}\",\"blockOwnerDeletion\":true}]},\"data\":{\"skey\":\"c3ZhbHVl\"},\"type\":\"Opaque\"}"
var cmManifest2 = "{\"kind\":\"ConfigMap\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"cm1-5-ci\",\"creationTimestamp\":null,\"ownerReferences\":[{\"apiVersion\":\"argoproj.io/v1alpha1\",\"kind\":\"Workflow\",\"name\":\"{{workflow.name}}\",\"uid\":\"{{workflow.uid}}\",\"blockOwnerDeletion\":true}]},\"data\":{\"key1\":\"value1\"}}"
var secManifest2 = "{\"kind\":\"Secret\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"secret1-5-ci\",\"creationTimestamp\":null,\"ownerReferences\":[{\"apiVersion\":\"argoproj.io/v1alpha1\",\"kind\":\"Workflow\",\"name\":\"{{workflow.name}}\",\"uid\":\"{{workflow.uid}}\",\"blockOwnerDeletion\":true}]},\"data\":{\"skey1\":\"c3ZhbHVlMQ==\"},\"type\":\"Opaque\"}"
var secManifest3 = "{\"kind\":\"Secret\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"secret5-5-ci\",\"creationTimestamp\":null,\"ownerReferences\":[{\"apiVersion\":\"argoproj.io/v1alpha1\",\"kind\":\"Workflow\",\"name\":\"{{workflow.name}}\",\"uid\":\"{{workflow.uid}}\",\"blockOwnerDeletion\":true}]},\"data\":{\"skey5\":\"c3ZhbHVlNQ==\"},\"type\":\"Opaque\"}"
var secManifest4 = "{\"kind\":\"Secret\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"secret4-5-ci\",\"creationTimestamp\":null,\"ownerReferences\":[{\"apiVersion\":\"argoproj.io/v1alpha1\",\"kind\":\"Workflow\",\"name\":\"{{workflow.name}}\",\"uid\":\"{{workflow.uid}}\",\"blockOwnerDeletion\":true}]},\"data\":{\"skey4\":\"c3ZhbHVlNA==\"},\"type\":\"Opaque\"}"
var cmManifest3 = "{\"kind\":\"ConfigMap\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"cm5-5-ci\",\"creationTimestamp\":null,\"ownerReferences\":[{\"apiVersion\":\"argoproj.io/v1alpha1\",\"kind\":\"Workflow\",\"name\":\"{{workflow.name}}\",\"uid\":\"{{workflow.uid}}\",\"blockOwnerDeletion\":true}]},\"data\":{\"key5\":\"value5\"}}"
var cmManifest4 = "{\"kind\":\"ConfigMap\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"cm4-5-ci\",\"creationTimestamp\":null,\"ownerReferences\":[{\"apiVersion\":\"argoproj.io/v1alpha1\",\"kind\":\"Workflow\",\"name\":\"{{workflow.name}}\",\"uid\":\"{{workflow.uid}}\",\"blockOwnerDeletion\":true}]},\"data\":{\"key4\":\"value4\"}}"

func getWorkflowServiceImpl(t *testing.T) *WorkflowServiceImpl {
	logger, dbConnection := getDbConnAndLoggerService(t)
	ciCdConfig, _ := types.GetCiCdConfig()
	newGlobalCMCSRepositoryImpl := repository.NewGlobalCMCSRepositoryImpl(logger, dbConnection)
	globalCMCSServiceImpl := NewGlobalCMCSServiceImpl(logger, newGlobalCMCSRepositoryImpl)
	newEnvConfigOverrideRepository := chartConfig.NewEnvConfigOverrideRepository(dbConnection)
	newConfigMapRepositoryImpl := chartConfig.NewConfigMapRepositoryImpl(logger, dbConnection)
	newChartRepository := chartRepoRepository.NewChartRepository(dbConnection)
	newCommonServiceImpl := commonService.NewCommonServiceImpl(logger, newChartRepository, newEnvConfigOverrideRepository, nil, nil, nil, nil, nil, nil, nil)
	mergeUtil := util.MergeUtil{Logger: logger}
	appService := app.NewAppService(nil, nil, &mergeUtil, logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, newConfigMapRepositoryImpl, nil, nil, nil, nil, nil, newCommonServiceImpl, nil, nil, nil, nil, nil, nil, nil, nil, "", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	runTimeConfig, _ := client.GetRuntimeConfig()
	k8sUtil := k8s.NewK8sUtil(logger, runTimeConfig)
	clusterRepositoryImpl := repository3.NewClusterRepositoryImpl(dbConnection, logger)
	v := informer.NewGlobalMapClusterNamespace()
	k8sInformerFactoryImpl := informer.NewK8sInformerFactoryImpl(logger, v, runTimeConfig, k8sUtil)
	clusterService := cluster.NewClusterServiceImpl(clusterRepositoryImpl, logger, k8sUtil, k8sInformerFactoryImpl, nil, nil, nil)
	k8sCommonServiceImpl := k8s2.NewK8sCommonServiceImpl(logger, k8sUtil, clusterService)
	appStatusRepositoryImpl := appStatus.NewAppStatusRepositoryImpl(dbConnection, logger)
	environmentRepositoryImpl := repository3.NewEnvironmentRepositoryImpl(dbConnection, logger, appStatusRepositoryImpl)
	argoWorkflowExecutorImpl := executors.NewArgoWorkflowExecutorImpl(logger)
	workflowServiceImpl, _ := NewWorkflowServiceImpl(logger, environmentRepositoryImpl, ciCdConfig, appService, globalCMCSServiceImpl, argoWorkflowExecutorImpl, k8sUtil, nil, k8sCommonServiceImpl)
	return workflowServiceImpl
}
func TestWorkflowServiceImpl_SubmitWorkflow(t *testing.T) {
	workflowServiceImpl := getWorkflowServiceImpl(t)

	t.Run("Verify submit workflow with S3 archive logs", func(t *testing.T) {

		workflowRequest := types.WorkflowRequest{
			WorkflowNamePrefix: "1-ci",
			PipelineName:       "ci-1-sslm",
			PipelineId:         2,
			DockerImageTag:     "680028fe-1-2",
			DockerRegistryId:   "",
			DockerRegistryType: "docker-hub",
			DockerRegistryURL:  "docker.io",
			DockerConnection:   "",
			DockerCert:         "",
			DockerRepository:   "",
			CheckoutPath:       "",
			DockerUsername:     "",
			DockerPassword:     "",
			AwsRegion:          "",
			AccessKey:          "",
			SecretKey:          "",
			CiCacheLocation:    "ci-caching",
			CiCacheRegion:      "us-east-2",
			CiCacheFileName:    "ci-1-sslm-1.tar.gz",
			CiProjectDetails: []bean2.CiProjectDetails{{
				GitRepository:   "https://github.com/pawan-59/test",
				MaterialName:    "1-test",
				CheckoutPath:    "./",
				FetchSubmodules: false,
				CommitHash:      "680028fef6172f9528b2f29119f1713e4a1ae1a4",
				GitTag:          "",
				CommitTime:      "2023-05-17T12:04:40+05:30",
				Type:            "SOURCE_TYPE_BRANCH_FIXED",
				Message:         "Update script.js",
				Author:          "Pawan Kumar <85476803+pawan-59@users.noreply.github.com>",
				GitOptions: bean2.GitOptions{
					UserName:      "",
					Password:      "",
					SshPrivateKey: "",
					AccessToken:   "",
					AuthMode:      "ANONYMOUS",
				},
				SourceType:  "SOURCE_TYPE_BRANCH_FIXED",
				SourceValue: "main",
				WebhookData: pipelineConfig.WebhookData{},
			}},
			ContainerResources:       bean2.ContainerResources{},
			ActiveDeadlineSeconds:    3600,
			CiImage:                  "",
			Namespace:                "devtron-ci",
			WorkflowId:               2,
			TriggeredBy:              2,
			CacheLimit:               5000000000,
			BeforeDockerBuildScripts: nil,
			AfterDockerBuildScripts:  nil,
			CiArtifactLocation:       "",
			CiArtifactBucket:         "",
			CiArtifactFileName:       "",
			CiArtifactRegion:         "",
			ScanEnabled:              true,
			CloudProvider:            "S3",
			BlobStorageConfigured:    true,
			BlobStorageS3Config: &blob_storage.BlobStorageS3Config{
				AccessKey:                  "",
				Passkey:                    "",
				EndpointUrl:                "",
				IsInSecure:                 true,
				CiLogBucketName:            "ci-log-container",
				CiLogRegion:                "us-east-2",
				CiLogBucketVersioning:      true,
				CiCacheBucketName:          "",
				CiCacheRegion:              "",
				CiCacheBucketVersioning:    false,
				CiArtifactBucketName:       "",
				CiArtifactRegion:           "",
				CiArtifactBucketVersioning: false,
			},
			AzureBlobConfig:            nil,
			GcpBlobConfig:              nil,
			BlobStorageLogsKey:         "",
			InAppLoggingEnabled:        false,
			DefaultAddressPoolBaseCidr: "",
			DefaultAddressPoolSize:     0,
			PreCiSteps:                 nil,
			PostCiSteps:                nil,
			RefPlugins:                 nil,
			AppName:                    "app",
			TriggerByAuthor:            "admin",
			CiBuildConfig: &bean2.CiBuildConfigBean{
				Id:                        1,
				GitMaterialId:             0,
				BuildContextGitMaterialId: 1,
				UseRootBuildContext:       true,
				CiBuildType:               "self-dockerfile-build",
				DockerBuildConfig: &bean2.DockerBuildConfig{
					DockerfilePath:         "Dockerfile",
					DockerfileContent:      "",
					Args:                   nil,
					TargetPlatform:         "",
					Language:               "",
					LanguageFramework:      "",
					DockerBuildOptions:     nil,
					BuildContext:           "",
					UseBuildx:              false,
					BuildxK8sDriverOptions: nil,
				},
				BuildPackConfig: nil,
			},
			CiBuildDockerMtuValue:     -1,
			IgnoreDockerCachePush:     false,
			IgnoreDockerCachePull:     false,
			CacheInvalidate:           false,
			IsPvcMounted:              false,
			ExtraEnvironmentVariables: nil,
			EnableBuildContext:        false,
			AppId:                     12,
			EnvironmentId:             0,
			OrchestratorHost:          "",
			OrchestratorToken:         "",
			IsExtRun:                  false,
			ImageRetryCount:           0,
			ImageRetryInterval:        5,
			Type:                      bean2.CI_WORKFLOW_PIPELINE_TYPE,
			WorkflowExecutor:          pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,
		}

		data, _ := workflowServiceImpl.SubmitWorkflow(&workflowRequest)

		createdWf := &v1alpha1.Workflow{}
		obj := data.Items[0].Object

		runtime.DefaultUnstructuredConverter.FromUnstructured(obj, createdWf)

		verifyMetadata(t, workflowRequest, createdWf)

		verifySpec(t, workflowServiceImpl, createdWf, bean2.CI_WORKFLOW_NAME)
		assert.Equal(t, reflect.ValueOf(reflect.ValueOf(workflowServiceImpl.ciCdConfig.CiNodeLabelSelector)), reflect.ValueOf(reflect.ValueOf(createdWf.Spec.NodeSelector)))

		assert.Equal(t, 1, len(createdWf.Spec.Templates))
		template := createdWf.Spec.Templates[0]

		verifyTemplateSpecContainerPort(t, template)

		verifyTemplateSpecContainerEnv(t, template, workflowServiceImpl)

		verifyResourceLimitAndRequest(t, template, workflowServiceImpl)

		verifyS3BlobStorage(t, template, workflowServiceImpl, workflowRequest)

		verifyToleration(t, workflowServiceImpl, createdWf)
	})

	t.Run("Verify submit workflow without archive logs", func(t *testing.T) {

		workflowRequest := types.WorkflowRequest{
			WorkflowNamePrefix: "1-ci",
			PipelineName:       "ci-1-sslm",
			PipelineId:         2,
			DockerImageTag:     "680028fe-1-2",
			DockerRegistryId:   "",
			DockerRegistryType: "docker-hub",
			DockerRegistryURL:  "docker.io",
			DockerConnection:   "",
			DockerCert:         "",
			DockerRepository:   "",
			CheckoutPath:       "",
			DockerUsername:     "",
			DockerPassword:     "",
			AwsRegion:          "",
			AccessKey:          "",
			SecretKey:          "",
			CiCacheLocation:    "ci-caching",
			CiCacheRegion:      "us-east-2",
			CiCacheFileName:    "ci-1-sslm-1.tar.gz",
			CiProjectDetails: []bean2.CiProjectDetails{{
				GitRepository:   "https://github.com/pawan-59/test",
				MaterialName:    "1-test",
				CheckoutPath:    "./",
				FetchSubmodules: false,
				CommitHash:      "680028fef6172f9528b2f29119f1713e4a1ae1a4",
				GitTag:          "",
				CommitTime:      "2023-05-17T12:04:40+05:30",
				Type:            "SOURCE_TYPE_BRANCH_FIXED",
				Message:         "Update script.js",
				Author:          "Pawan Kumar <85476803+pawan-59@users.noreply.github.com>",
				GitOptions: bean2.GitOptions{
					UserName:      "",
					Password:      "",
					SshPrivateKey: "",
					AccessToken:   "",
					AuthMode:      "ANONYMOUS",
				},
				SourceType:  "SOURCE_TYPE_BRANCH_FIXED",
				SourceValue: "main",
				WebhookData: pipelineConfig.WebhookData{},
			}},
			ContainerResources:         bean2.ContainerResources{},
			ActiveDeadlineSeconds:      3600,
			CiImage:                    "",
			Namespace:                  "devtron-ci",
			WorkflowId:                 2,
			TriggeredBy:                2,
			CacheLimit:                 5000000000,
			BeforeDockerBuildScripts:   nil,
			AfterDockerBuildScripts:    nil,
			CiArtifactLocation:         "",
			CiArtifactBucket:           "",
			CiArtifactFileName:         "",
			CiArtifactRegion:           "",
			ScanEnabled:                true,
			CloudProvider:              "S3",
			BlobStorageConfigured:      false,
			BlobStorageS3Config:        nil,
			AzureBlobConfig:            nil,
			GcpBlobConfig:              nil,
			BlobStorageLogsKey:         "",
			InAppLoggingEnabled:        false,
			DefaultAddressPoolBaseCidr: "",
			DefaultAddressPoolSize:     0,
			PreCiSteps:                 nil,
			PostCiSteps:                nil,
			RefPlugins:                 nil,
			AppName:                    "app",
			TriggerByAuthor:            "admin",
			CiBuildConfig: &bean2.CiBuildConfigBean{
				Id:                        1,
				GitMaterialId:             0,
				BuildContextGitMaterialId: 1,
				UseRootBuildContext:       true,
				CiBuildType:               "self-dockerfile-build",
				DockerBuildConfig: &bean2.DockerBuildConfig{
					DockerfilePath:         "Dockerfile",
					DockerfileContent:      "",
					Args:                   nil,
					TargetPlatform:         "",
					Language:               "",
					LanguageFramework:      "",
					DockerBuildOptions:     nil,
					BuildContext:           "",
					UseBuildx:              false,
					BuildxK8sDriverOptions: nil,
				},
				BuildPackConfig: nil,
			},
			CiBuildDockerMtuValue:     -1,
			IgnoreDockerCachePush:     false,
			IgnoreDockerCachePull:     false,
			CacheInvalidate:           false,
			IsPvcMounted:              false,
			ExtraEnvironmentVariables: nil,
			EnableBuildContext:        false,
			AppId:                     2,
			EnvironmentId:             0,
			OrchestratorHost:          "",
			OrchestratorToken:         "",
			IsExtRun:                  false,
			ImageRetryCount:           0,
			ImageRetryInterval:        5,
			Type:                      bean2.CI_WORKFLOW_PIPELINE_TYPE,
			WorkflowExecutor:          pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,
		}

		data, _ := workflowServiceImpl.SubmitWorkflow(&workflowRequest)

		createdWf := &v1alpha1.Workflow{}
		obj := data.Items[0].Object

		runtime.DefaultUnstructuredConverter.FromUnstructured(obj, createdWf)
		verifyMetadata(t, workflowRequest, createdWf)

		verifySpec(t, workflowServiceImpl, createdWf, bean2.CI_WORKFLOW_NAME)
		assert.Equal(t, workflowServiceImpl.ciCdConfig.CiNodeLabelSelector, createdWf.Spec.NodeSelector)

		assert.Equal(t, 1, len(createdWf.Spec.Templates))

		template := v1alpha1.Template{}

		if createdWf.Spec.Templates[0].Name == bean2.CI_WORKFLOW_NAME {
			template = createdWf.Spec.Templates[0]
		} else {
			template = createdWf.Spec.Templates[1]
		}

		verifyTemplateSpecContainerPort(t, template)

		assert.Equal(t, false, *template.ArchiveLocation.ArchiveLogs)

		verifyTemplateSpecContainerEnv(t, template, workflowServiceImpl)

		verifyResourceLimitAndRequest(t, template, workflowServiceImpl)

		verifyToleration(t, workflowServiceImpl, createdWf)
	})

	t.Run("Verify submit workflow for Jobs", func(t *testing.T) {

		workflowRequest := types.WorkflowRequest{
			WorkflowNamePrefix: "20-pipeline-2",
			PipelineName:       "pipeline",
			PipelineId:         1,
			DockerImageTag:     "",
			DockerRegistryId:   "",
			DockerRegistryType: "",
			DockerRegistryURL:  "",
			DockerConnection:   "",
			DockerCert:         "",
			DockerRepository:   "",
			CheckoutPath:       "",
			DockerUsername:     "",
			DockerPassword:     "",
			AwsRegion:          "",
			AccessKey:          "",
			SecretKey:          "",
			CiCacheLocation:    "",
			CiCacheRegion:      "",
			CiCacheFileName:    "pipeline-2.tar.gz",
			CiProjectDetails: []bean2.CiProjectDetails{{
				GitRepository:   "https://github.com/pawan-59/test",
				MaterialName:    "1-test",
				CheckoutPath:    "./",
				FetchSubmodules: false,
				CommitHash:      "680028fef6172f9528b2f29119f1713e4a1ae1a4",
				GitTag:          "",
				CommitTime:      "2023-05-17T12:04:40+05:30",
				Type:            "SOURCE_TYPE_BRANCH_FIXED",
				Message:         "Update script.js",
				Author:          "Pawan Kumar <85476803+pawan-59@users.noreply.github.com>",
				GitOptions: bean2.GitOptions{
					UserName:      "",
					Password:      "",
					SshPrivateKey: "",
					AccessToken:   "",
					AuthMode:      "ANONYMOUS",
				},
				SourceType:  "SOURCE_TYPE_BRANCH_FIXED",
				SourceValue: "main",
				WebhookData: pipelineConfig.WebhookData{},
			}},
			ContainerResources:       bean2.ContainerResources{},
			ActiveDeadlineSeconds:    3600,
			CiImage:                  "",
			Namespace:                "devtron-ci",
			WorkflowId:               20,
			TriggeredBy:              2,
			CacheLimit:               5000000000,
			BeforeDockerBuildScripts: nil,
			AfterDockerBuildScripts:  nil,
			CiArtifactLocation:       "",
			CiArtifactBucket:         "",
			CiArtifactFileName:       "",
			CiArtifactRegion:         "",
			ScanEnabled:              false,
			CloudProvider:            "AZURE",
			BlobStorageConfigured:    true,
			BlobStorageS3Config: &blob_storage.BlobStorageS3Config{
				AccessKey:                  "",
				Passkey:                    "",
				EndpointUrl:                "",
				IsInSecure:                 true,
				CiLogBucketName:            "",
				CiLogRegion:                "us-east-2",
				CiLogBucketVersioning:      true,
				CiCacheBucketName:          "",
				CiCacheRegion:              "",
				CiCacheBucketVersioning:    false,
				CiArtifactBucketName:       "",
				CiArtifactRegion:           "",
				CiArtifactBucketVersioning: false,
			},
			AzureBlobConfig: &blob_storage.AzureBlobConfig{
				Enabled:               true,
				AccountName:           "",
				BlobContainerCiLog:    "",
				BlobContainerCiCache:  "",
				BlobContainerArtifact: "",
				AccountKey:            "",
			},
			GcpBlobConfig:              nil,
			BlobStorageLogsKey:         "",
			InAppLoggingEnabled:        false,
			DefaultAddressPoolBaseCidr: "",
			DefaultAddressPoolSize:     0,
			PreCiSteps: []*bean2.StepObject{
				{
					Name:                     "Task 1",
					Index:                    1,
					StepType:                 "INLINE",
					ExecutorType:             "SHELL",
					RefPluginId:              0,
					Script:                   "#!/bin/sh \nset -eo pipefail \n#set -v  ## uncomment this to debug the script \n\necho \"Hello\"",
					InputVars:                nil,
					ExposedPorts:             nil,
					OutputVars:               nil,
					TriggerSkipConditions:    nil,
					SuccessFailureConditions: nil,
					DockerImage:              "",
					Command:                  "",
				},
			},
			PostCiSteps:     nil,
			RefPlugins:      nil,
			AppName:         "job/f1851uikJ",
			TriggerByAuthor: "admin",
			CiBuildConfig: &bean2.CiBuildConfigBean{
				Id:                        2,
				GitMaterialId:             0,
				BuildContextGitMaterialId: 0,
				UseRootBuildContext:       false,
				CiBuildType:               "skip-build",
				DockerBuildConfig:         nil,
				BuildPackConfig:           nil,
			},
			CiBuildDockerMtuValue:     -1,
			IgnoreDockerCachePush:     false,
			IgnoreDockerCachePull:     false,
			CacheInvalidate:           false,
			IsPvcMounted:              false,
			ExtraEnvironmentVariables: nil,
			EnableBuildContext:        true,
			AppId:                     1,
			EnvironmentId:             0,
			OrchestratorHost:          "",
			OrchestratorToken:         "",
			IsExtRun:                  false,
			ImageRetryCount:           0,
			ImageRetryInterval:        5,
			Type:                      bean2.JOB_WORKFLOW_PIPELINE_TYPE,
			WorkflowExecutor:          pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,
		}

		data, _ := workflowServiceImpl.SubmitWorkflow(&workflowRequest)

		createdWf := &v1alpha1.Workflow{}
		obj := data.Items[0].Object

		runtime.DefaultUnstructuredConverter.FromUnstructured(obj, createdWf)
		verifyMetadata(t, workflowRequest, createdWf)

		verifySpec(t, workflowServiceImpl, createdWf, bean2.CI_WORKFLOW_WITH_STAGES)
		assert.Equal(t, workflowServiceImpl.ciCdConfig.CiNodeLabelSelector, createdWf.Spec.NodeSelector)

		assert.Equal(t, 6, len(createdWf.Spec.Templates))

		for _, template := range createdWf.Spec.Templates {
			if template.Name == bean2.CI_WORKFLOW_NAME {

				verifyTemplateSpecContainerPort(t, template)

				verifyTemplateSpecContainerEnvFrom(t, template, 4)

				verifyTemplateSpecContainerEnv(t, template, workflowServiceImpl)

				verifyResourceLimitAndRequest(t, template, workflowServiceImpl)

				verifyTemplateSpecContainerVolumeMounts(t, template, 4)

				verifyS3BlobStorage(t, template, workflowServiceImpl, workflowRequest)

			}
			if template.Name == bean2.CI_WORKFLOW_WITH_STAGES {

				verifyTemplateSteps(t, template.Steps)

			}
			if template.Name == "cm-0" {

				verifyTemplateResource(t, template.Resource, "create", cmManifest)

			}
			if template.Name == "cm-1" {

				verifyTemplateResource(t, template.Resource, "create", cmManifest2)

			}
			if template.Name == "sec-0" {

				verifyTemplateResource(t, template.Resource, "create", secManifest)

			}
			if template.Name == "sec-1" {

				verifyTemplateResource(t, template.Resource, "create", secManifest2)

			}
		}

		assert.Equal(t, 4, len(createdWf.Spec.Volumes))

		verifyToleration(t, workflowServiceImpl, createdWf)
	})

	t.Run("Verify submit workflow for External Jobs", func(t *testing.T) {

		workflowRequest := types.WorkflowRequest{
			WorkflowNamePrefix: "7-pipeline-1",
			PipelineName:       "pipeline",
			PipelineId:         1,
			DockerImageTag:     "",
			DockerRegistryId:   "",
			DockerRegistryType: "",
			DockerRegistryURL:  "",
			DockerConnection:   "",
			DockerCert:         "",
			DockerRepository:   "",
			CheckoutPath:       "",
			DockerUsername:     "",
			DockerPassword:     "",
			AwsRegion:          "",
			AccessKey:          "",
			SecretKey:          "",
			CiCacheLocation:    "",
			CiCacheRegion:      "",
			CiCacheFileName:    "pipeline-1.tar.gz",
			CiProjectDetails: []bean2.CiProjectDetails{{
				GitRepository:   "https://github.com/pawan-59/test",
				MaterialName:    "1-test",
				CheckoutPath:    "./",
				FetchSubmodules: false,
				CommitHash:      "680028fef6172f9528b2f29119f1713e4a1ae1a4",
				GitTag:          "",
				CommitTime:      "2023-05-17T12:04:40+05:30",
				Type:            "SOURCE_TYPE_BRANCH_FIXED",
				Message:         "Update script.js",
				Author:          "Pawan Kumar <85476803+pawan-59@users.noreply.github.com>",
				GitOptions: bean2.GitOptions{
					UserName:      "",
					Password:      "",
					SshPrivateKey: "",
					AccessToken:   "",
					AuthMode:      "ANONYMOUS",
				},
				SourceType:  "SOURCE_TYPE_BRANCH_FIXED",
				SourceValue: "main",
				WebhookData: pipelineConfig.WebhookData{},
			}},
			ContainerResources:       bean2.ContainerResources{},
			ActiveDeadlineSeconds:    3600,
			CiImage:                  "",
			Namespace:                "2-devtron",
			WorkflowId:               5,
			TriggeredBy:              2,
			CacheLimit:               5000000000,
			BeforeDockerBuildScripts: nil,
			AfterDockerBuildScripts:  nil,
			CiArtifactLocation:       "",
			CiArtifactBucket:         "",
			CiArtifactFileName:       "",
			CiArtifactRegion:         "",
			ScanEnabled:              false,
			CloudProvider:            "AZURE",
			BlobStorageConfigured:    true,
			BlobStorageS3Config: &blob_storage.BlobStorageS3Config{
				AccessKey:                  "",
				Passkey:                    "",
				EndpointUrl:                "",
				IsInSecure:                 true,
				CiLogBucketName:            "ci-log-container",
				CiLogRegion:                "us-east-2",
				CiLogBucketVersioning:      true,
				CiCacheBucketName:          "",
				CiCacheRegion:              "",
				CiCacheBucketVersioning:    false,
				CiArtifactBucketName:       "",
				CiArtifactRegion:           "",
				CiArtifactBucketVersioning: false,
			},
			AzureBlobConfig: &blob_storage.AzureBlobConfig{
				Enabled:               true,
				AccountName:           "",
				BlobContainerCiLog:    "ci-log-container",
				BlobContainerCiCache:  "ci-cache-container",
				BlobContainerArtifact: "ci-log-container",
				AccountKey:            "",
			},
			GcpBlobConfig:              nil,
			BlobStorageLogsKey:         "",
			InAppLoggingEnabled:        false,
			DefaultAddressPoolBaseCidr: "",
			DefaultAddressPoolSize:     0,
			PreCiSteps: []*bean2.StepObject{
				{
					Name:                     "Task 1",
					Index:                    1,
					StepType:                 "INLINE",
					ExecutorType:             "SHELL",
					RefPluginId:              0,
					Script:                   "#!/bin/sh \nset -eo pipefail \n#set -v  ## uncomment this to debug the script \n\necho \"Hello\"",
					InputVars:                nil,
					ExposedPorts:             nil,
					OutputVars:               nil,
					TriggerSkipConditions:    nil,
					SuccessFailureConditions: nil,
					DockerImage:              "",
					Command:                  "",
				},
			},
			PostCiSteps:     nil,
			RefPlugins:      nil,
			AppName:         "",
			TriggerByAuthor: "admin",
			CiBuildConfig: &bean2.CiBuildConfigBean{
				Id:                        2,
				GitMaterialId:             0,
				BuildContextGitMaterialId: 0,
				UseRootBuildContext:       false,
				CiBuildType:               "skip-build",
				DockerBuildConfig:         nil,
				BuildPackConfig:           nil,
			},
			CiBuildDockerMtuValue:     -1,
			IgnoreDockerCachePush:     false,
			IgnoreDockerCachePull:     false,
			CacheInvalidate:           false,
			IsPvcMounted:              false,
			ExtraEnvironmentVariables: nil,
			EnableBuildContext:        true,
			AppId:                     1,
			EnvironmentId:             0,
			OrchestratorHost:          "",
			OrchestratorToken:         "",
			IsExtRun:                  false,
			ImageRetryCount:           0,
			ImageRetryInterval:        5,
			Env: &repository3.Environment{
				Id:        3,
				Name:      "2-devtron",
				ClusterId: 2,
				Cluster: &repository3.Cluster{
					Id:                 2,
					ClusterName:        "in_cluster",
					ServerUrl:          "",
					PrometheusEndpoint: "",
					Active:             true,
					CdArgoSetup:        true,
					Config: map[string]string{
						"bearer_token": "",
					},
					PUserName:              "",
					PPassword:              "",
					PTlsClientCert:         "",
					PTlsClientKey:          "",
					AgentInstallationStage: 0,
					K8sVersion:             "v1.26.8",
					ErrorInConnecting:      "",
					IsVirtualCluster:       false,
					InsecureSkipTlsVerify:  true,
				},
				Active:                true,
				Default:               false,
				GrafanaDatasourceId:   0,
				Namespace:             "2-devtron",
				EnvironmentIdentifier: "in_cluster__2-devtron",
				Description:           "",
				IsVirtualEnvironment:  false,
			},
			Type:             bean2.JOB_WORKFLOW_PIPELINE_TYPE,
			WorkflowExecutor: pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,
		}

		data, _ := workflowServiceImpl.SubmitWorkflow(&workflowRequest)

		createdWf := &v1alpha1.Workflow{}
		obj := data.Items[0].Object

		runtime.DefaultUnstructuredConverter.FromUnstructured(obj, createdWf)
		verifyMetadata(t, workflowRequest, createdWf)

		verifySpec(t, workflowServiceImpl, createdWf, bean2.CI_WORKFLOW_WITH_STAGES)

		assert.Equal(t, 10, len(createdWf.Spec.Templates))

		for _, template := range createdWf.Spec.Templates {
			if template.Name == bean2.CI_WORKFLOW_NAME {

				verifyTemplateSpecContainerPort(t, template)

				verifyTemplateSpecContainerEnvFrom(t, template, 8)

				verifyTemplateSpecContainerEnv(t, template, workflowServiceImpl)

				verifyResourceLimitAndRequest(t, template, workflowServiceImpl)

				verifyTemplateSpecContainerVolumeMounts(t, template, 8)

				verifyS3BlobStorage(t, template, workflowServiceImpl, workflowRequest)

			}
			if template.Name == bean2.CI_WORKFLOW_WITH_STAGES {

				verifyTemplateSteps(t, template.Steps)

			}
			if template.Name == "cm-0" {

				verifyTemplateResource(t, template.Resource, "create", cmManifest)

			}
			if template.Name == "cm-1" {

				verifyTemplateResource(t, template.Resource, "create", cmManifest2)

			}
			if template.Name == "cm-3" {

				verifyTemplateResource(t, template.Resource, "create", cmManifest3)

			}
			if template.Name == "cm-2" {

				verifyTemplateResource(t, template.Resource, "create", cmManifest4)

			}
			if template.Name == "sec-2" {

				verifyTemplateResource(t, template.Resource, "create", secManifest4)

			}
			if template.Name == "sec-3" {

				verifyTemplateResource(t, template.Resource, "create", secManifest3)

			}
			if template.Name == "sec-1" {

				verifyTemplateResource(t, template.Resource, "create", secManifest2)

			}
			if template.Name == "sec-0" {

				verifyTemplateResource(t, template.Resource, "create", secManifest)

			}
		}

		assert.Equal(t, 8, len(createdWf.Spec.Volumes))

		verifyToleration(t, workflowServiceImpl, createdWf)
	})

}

func verifyTemplateSteps(t *testing.T, steps []v1alpha1.ParallelSteps) {
	count := 0
	for _, step := range steps {
		if step.Steps[0].Template == "cm-0" {
			assert.Equal(t, "create-env-cm-0", step.Steps[0].Name)
			count++
		}
		if step.Steps[0].Template == "cm-1" {
			assert.Equal(t, "create-env-cm-1", step.Steps[0].Name)
			count++

		}
		if step.Steps[0].Template == "cm-2" {
			assert.Equal(t, "create-env-cm-2", step.Steps[0].Name)
			count++
		}
		if step.Steps[0].Template == "cm-3" {
			assert.Equal(t, "create-env-cm-3", step.Steps[0].Name)
			count++

		}
		if step.Steps[0].Template == "sec-0" {
			assert.Equal(t, "create-env-sec-0", step.Steps[0].Name)
			count++
		}
		if step.Steps[0].Template == "sec-1" {
			assert.Equal(t, "create-env-sec-1", step.Steps[0].Name)
			count++
		}
		if step.Steps[0].Template == "sec-2" {
			assert.Equal(t, "create-env-sec-2", step.Steps[0].Name)
			count++
		}
		if step.Steps[0].Template == "sec-3" {
			assert.Equal(t, "create-env-sec-3", step.Steps[0].Name)
			count++
		}
		if step.Steps[0].Template == "ci" {
			assert.Equal(t, "run-wf", step.Steps[0].Name)
			count++
		}
	}
	assert.Equal(t, count, len(steps))

}

func verifyTemplateResource(t *testing.T, resource *v1alpha1.ResourceTemplate, action string, manifest string) {
	assert.Equal(t, action, resource.Action)
	//assert.Equal(t, manifest, resource.Manifest)
}

func verifyMetadata(t *testing.T, workflowRequest types.WorkflowRequest, createdWf *v1alpha1.Workflow) {
	assert.Equal(t, map[string]string{"devtron.ai/workflow-purpose": "ci"}, createdWf.ObjectMeta.Labels)
	assert.Equal(t, workflowRequest.WorkflowNamePrefix+"-", createdWf.ObjectMeta.GenerateName)
	assert.Equal(t, workflowRequest.Namespace, createdWf.ObjectMeta.Namespace)
}

func verifySpec(t *testing.T, workflowServiceImpl *WorkflowServiceImpl, createdWf *v1alpha1.Workflow, stageName string) {
	assert.Equal(t, stageName, createdWf.Spec.Entrypoint)
	assert.Equal(t, workflowServiceImpl.ciCdConfig.CiWorkflowServiceAccount, createdWf.Spec.ServiceAccountName)
}

func verifyTemplateSpecContainerPort(t *testing.T, template v1alpha1.Template) {
	assert.Equal(t, 1, len(template.Container.Ports))
	assert.Equal(t, "app-data", template.Container.Ports[0].Name)
	assert.Equal(t, int32(9102), template.Container.Ports[0].ContainerPort)
}

func verifyTemplateSpecContainerEnv(t *testing.T, template v1alpha1.Template, workflowServiceImpl *WorkflowServiceImpl) {
	assert.Equal(t, 3, len(template.Container.Env))
	for _, env := range template.Container.Env {
		if env.Name == "IMAGE_SCANNER_ENDPOINT" {
			assert.Equal(t, workflowServiceImpl.ciCdConfig.ImageScannerEndpoint, env.Value)
		}
	}
}

func verifyTemplateSpecContainerEnvFrom(t *testing.T, template v1alpha1.Template, sum int) {
	count := 0
	for _, envFrom := range template.Container.EnvFrom {
		if envFrom.SecretRef != nil {
			assert.True(t, envFrom.SecretRef.LocalObjectReference.Name != "")
			count++
		}
		if envFrom.ConfigMapRef != nil {
			assert.True(t, envFrom.ConfigMapRef.LocalObjectReference.Name != "")
			count++
		}
	}
	assert.Equal(t, sum, count)
}

func verifyTemplateSpecContainerVolumeMounts(t *testing.T, template v1alpha1.Template, count int) {
	assert.Equal(t, count, len(template.Container.VolumeMounts))
	for _, volumeMount := range template.Container.VolumeMounts {
		assert.True(t, volumeMount.Name != "")
		assert.True(t, volumeMount.MountPath != "")
	}
}

func verifyResourceLimitAndRequest(t *testing.T, template v1alpha1.Template, workflowServiceImpl *WorkflowServiceImpl) {

	resourceLimit := template.Container.Resources.Limits
	resourceRequest := template.Container.Resources.Requests

	//assert.Equal(t, resourceLimit["cpu"], resource.MustParse(workflowServiceImpl.ciConfig.LimitCpu))
	assert.Equal(t, resourceLimit["memory"], resource.MustParse(workflowServiceImpl.ciCdConfig.CiLimitMem))
	//assert.Equal(t, resourceRequest["cpu"], resource.MustParse(workflowServiceImpl.ciConfig.ReqCpu))
	assert.Equal(t, resourceRequest["memory"], resource.MustParse(workflowServiceImpl.ciCdConfig.CiReqMem))
}

func verifyS3BlobStorage(t *testing.T, template v1alpha1.Template, workflowServiceImpl *WorkflowServiceImpl, workflowRequest types.WorkflowRequest) {

	assert.Equal(t, true, *template.ArchiveLocation.ArchiveLogs)
	assert.Equal(t, workflowServiceImpl.ciCdConfig.CiDefaultBuildLogsKeyPrefix+"/"+workflowRequest.WorkflowNamePrefix, template.ArchiveLocation.S3.Key)
	assert.Equal(t, "accessKey", template.ArchiveLocation.S3.S3Bucket.AccessKeySecret.Key)
	assert.Equal(t, "workflow-minio-cred", template.ArchiveLocation.S3.S3Bucket.AccessKeySecret.LocalObjectReference.Name)
	blobStorageS3Config := workflowRequest.BlobStorageS3Config
	s3CompatibleEndpointUrl := blobStorageS3Config.EndpointUrl
	if s3CompatibleEndpointUrl == "" {
		s3CompatibleEndpointUrl = "s3.amazonaws.com"
	} else {
		parsedUrl, err := url.Parse(s3CompatibleEndpointUrl)
		if err != nil {
			assert.Error(t, err)
		} else {
			s3CompatibleEndpointUrl = parsedUrl.Host
		}
	}

	assert.Equal(t, s3CompatibleEndpointUrl, template.ArchiveLocation.S3.S3Bucket.Endpoint)
	assert.Equal(t, blobStorageS3Config.CiLogBucketName, template.ArchiveLocation.S3.S3Bucket.Bucket)
	assert.Equal(t, blobStorageS3Config.CiLogRegion, template.ArchiveLocation.S3.S3Bucket.Region)
	assert.Equal(t, blobStorageS3Config.IsInSecure, *template.ArchiveLocation.S3.S3Bucket.Insecure)
	assert.Equal(t, "secretKey", template.ArchiveLocation.S3.S3Bucket.SecretKeySecret.Key)
	assert.Equal(t, "workflow-minio-cred", template.ArchiveLocation.S3.S3Bucket.SecretKeySecret.LocalObjectReference.Name)
}

func verifyToleration(t *testing.T, workflowServiceImpl *WorkflowServiceImpl, createdWf *v1alpha1.Workflow) {
	assert.Equal(t, 1, len(createdWf.Spec.Tolerations))
	assert.Equal(t, workflowServiceImpl.ciCdConfig.CiTaintKey, createdWf.Spec.Tolerations[0].Key)
	assert.Equal(t, v12.TolerationOpEqual, createdWf.Spec.Tolerations[0].Operator)
	assert.Equal(t, workflowServiceImpl.ciCdConfig.CiTaintValue, createdWf.Spec.Tolerations[0].Value)
	assert.Equal(t, v12.TaintEffectNoSchedule, createdWf.Spec.Tolerations[0].Effect)
}
