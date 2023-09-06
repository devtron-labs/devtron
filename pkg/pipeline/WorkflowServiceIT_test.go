package pipeline

import (
	"fmt"
	"github.com/devtron-labs/authenticator/client"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
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
	"github.com/devtron-labs/devtron/util/k8s"
	"github.com/stretchr/testify/assert"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"net/url"
	"testing"
)

var pipelineId = 0

var cmManifest = "{\"kind\":\"ConfigMap\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"cm-20-ci\",\"creationTimestamp\":null,\"ownerReferences\":[{\"apiVersion\":\"argoproj.io/v1alpha1\",\"kind\":\"Workflow\",\"name\":\"{{workflow.name}}\",\"uid\":\"{{workflow.uid}}\",\"blockOwnerDeletion\":true}]},\"data\":{\"key\":\"value\"}}"
var secManifest = "{\"kind\":\"Secret\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"devtron-secret-20-ci\",\"creationTimestamp\":null,\"ownerReferences\":[{\"apiVersion\":\"argoproj.io/v1alpha1\",\"kind\":\"Workflow\",\"name\":\"{{workflow.name}}\",\"uid\":\"{{workflow.uid}}\",\"blockOwnerDeletion\":true}]},\"data\":{\"skey\":\"c3ZhbHVl\"},\"type\":\"Opaque\"}"
var cmManifest2 = "{\"kind\":\"ConfigMap\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"cm1-20-ci\",\"creationTimestamp\":null,\"ownerReferences\":[{\"apiVersion\":\"argoproj.io/v1alpha1\",\"kind\":\"Workflow\",\"name\":\"{{workflow.name}}\",\"uid\":\"{{workflow.uid}}\",\"blockOwnerDeletion\":true}]},\"data\":{\"key\":\"value\"}}"
var secManifest2 = "{\"kind\":\"Secret\",\"apiVersion\":\"v1\",\"metadata\":{\"name\":\"devtron-secret1-20-ci\",\"creationTimestamp\":null,\"ownerReferences\":[{\"apiVersion\":\"argoproj.io/v1alpha1\",\"kind\":\"Workflow\",\"name\":\"{{workflow.name}}\",\"uid\":\"{{workflow.uid}}\",\"blockOwnerDeletion\":true}]},\"data\":{\"skey1\":\"c3ZhbHVlMQ==\"},\"type\":\"Opaque\"}"

func getWorkflowServiceImpl(t *testing.T) *WorkflowServiceImpl {
	logger, dbConnection := getDbConnAndLoggerService(t)
	ciCdConfig, _ := GetCiCdConfig()
	newGlobalCMCSRepositoryImpl := repository.NewGlobalCMCSRepositoryImpl(logger, dbConnection)
	globalCMCSServiceImpl := NewGlobalCMCSServiceImpl(logger, newGlobalCMCSRepositoryImpl)
	newEnvConfigOverrideRepository := chartConfig.NewEnvConfigOverrideRepository(dbConnection)
	newConfigMapRepositoryImpl := chartConfig.NewConfigMapRepositoryImpl(logger, dbConnection)
	newChartRepository := chartRepoRepository.NewChartRepository(dbConnection)
	newCommonServiceImpl := commonService.NewCommonServiceImpl(logger, newChartRepository, newEnvConfigOverrideRepository, nil, nil, nil, nil, nil, nil, nil)
	mergeUtil := util.MergeUtil{Logger: logger}
	appService := app.NewAppService(nil, nil, &mergeUtil, logger, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, newConfigMapRepositoryImpl, nil, nil, nil, nil, nil, newCommonServiceImpl, nil, nil, nil, nil, nil, nil, nil, nil, "", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	runTimeConfig, _ := client.GetRuntimeConfig()
	k8sUtil := k8s.NewK8sUtil(logger, runTimeConfig)
	clusterRepositoryImpl := repository3.NewClusterRepositoryImpl(dbConnection, logger)
	v := informer.NewGlobalMapClusterNamespace()
	k8sInformerFactoryImpl := informer.NewK8sInformerFactoryImpl(logger, v, runTimeConfig, k8sUtil)
	clusterService := cluster.NewClusterServiceImpl(clusterRepositoryImpl, logger, k8sUtil, k8sInformerFactoryImpl, nil, nil, nil)
	k8sCommonServiceImpl := k8s2.NewK8sCommonServiceImpl(logger, k8sUtil, clusterService)
	appStatusRepositoryImpl := appStatus.NewAppStatusRepositoryImpl(dbConnection, logger)
	environmentRepositoryImpl := repository3.NewEnvironmentRepositoryImpl(dbConnection, logger, appStatusRepositoryImpl)
	argoWorkflowExecutorImpl := NewArgoWorkflowExecutorImpl(logger)
	workflowServiceImpl, _ := NewWorkflowServiceImpl(logger, environmentRepositoryImpl, ciCdConfig, appService, globalCMCSServiceImpl, argoWorkflowExecutorImpl, k8sUtil, nil, k8sCommonServiceImpl)
	return workflowServiceImpl
}

func TestWorkflowServiceImpl_SubmitWorkflow(t *testing.T) {
	t.SkipNow()
	workflowServiceImpl := getWorkflowServiceImpl(t)

	t.Run("Verify submit workflow with S3 archive logs", func(t *testing.T) {

		workflowRequest := WorkflowRequest{
			WorkflowNamePrefix: "1-ci",
			PipelineName:       "ci-1-sslm",
			PipelineId:         2,
			DockerImageTag:     "680028fe-1-2",
			DockerRegistryId:   "ashexp",
			DockerRegistryType: "docker-hub",
			DockerRegistryURL:  "docker.io",
			DockerConnection:   "",
			DockerCert:         "",
			DockerRepository:   "ashexp/test",
			CheckoutPath:       "",
			DockerUsername:     "ashexp",
			DockerPassword:     "dckr_pat_0t9-gLzcn_8fPdf10QNzzaOYNog",
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
			CiImage:                  "686244538589.dkr.ecr.us-east-2.amazonaws.com/cirunner:47",
			Namespace:                "devtron-ci",
			WorkflowId:               2,
			TriggeredBy:              2,
			CacheLimit:               5000000000,
			BeforeDockerBuildScripts: nil,
			AfterDockerBuildScripts:  nil,
			CiArtifactLocation:       "s3://devtron-pro-ci-logs/arsenal-v1/ci-artifacts/2/2.zip",
			CiArtifactBucket:         "devtron-pro-ci-logs",
			CiArtifactFileName:       "arsenal-v1/ci-artifacts/2/2.zip",
			CiArtifactRegion:         "",
			ScanEnabled:              true,
			CloudProvider:            "S3",
			BlobStorageConfigured:    true,
			BlobStorageS3Config: &blob_storage.BlobStorageS3Config{
				AccessKey:                  "devtrontestblobstorage",
				Passkey:                    "",
				EndpointUrl:                "http://devtron-minio.devtroncd:9000",
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
			BlobStorageLogsKey:         "arsenal-v1/2-ci-1-sslm-1",
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
			OrchestratorHost:          "http://devtroncd-orchestrator-service-prod.devtroncd/webhook/msg/nats",
			OrchestratorToken:         "",
			IsExtRun:                  false,
			ImageRetryCount:           0,
			ImageRetryInterval:        5,
			Type:                      bean2.CI_WORKFLOW_PIPELINE_TYPE,
			WorkflowExecutor:          pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,
		}

		data, _ := workflowServiceImpl.SubmitWorkflow(&workflowRequest)
		createdWf := data.Items[0].Object
		fmt.Println(createdWf)
		verifyMetadata(t, workflowRequest, createdWf)

		spec := createdWf["spec"].(interface{}).(map[string]interface{})

		verifySpec(t, workflowServiceImpl, spec, bean2.CI_WORKFLOW_NAME)
		assert.Equal(t, workflowServiceImpl.ciCdConfig.CiNodeLabelSelector, spec["nodeSelector"])

		assert.Equal(t, 1, len(spec["templates"].(interface{}).([]interface{})))
		var template map[string]interface{}
		if spec["templates"].(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})["name"] == bean2.CI_WORKFLOW_NAME {
			template = spec["templates"].(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})
		} else {
			template = spec["templates"].(interface{}).([]interface{})[1].(interface{}).(map[string]interface{})
		}
		verifyTemplateSpecContainerPort(t, template)

		verifyTemplateSpecContainerEnv(t, template, workflowServiceImpl)

		verifyResourceLimitAndRequest(t, template, workflowServiceImpl)

		verifyS3BlobStorage(t, template, workflowServiceImpl, workflowRequest)

		verifyToleration(t, workflowServiceImpl, createdWf)
	})

	t.Run("Verify submit workflow without archive logs", func(t *testing.T) {

		workflowRequest := WorkflowRequest{
			WorkflowNamePrefix: "1-ci",
			PipelineName:       "ci-1-sslm",
			PipelineId:         1,
			DockerImageTag:     "680028fe-1-2",
			DockerRegistryId:   "ashexp",
			DockerRegistryType: "docker-hub",
			DockerRegistryURL:  "docker.io",
			DockerConnection:   "",
			DockerCert:         "",
			DockerRepository:   "ashexp/test",
			CheckoutPath:       "",
			DockerUsername:     "ashexp",
			DockerPassword:     "dckr_pat_0t9-gLzcn_8fPdf10QNzzaOYNog",
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
			CiImage:                    "686244538589.dkr.ecr.us-east-2.amazonaws.com/cirunner:47",
			Namespace:                  "devtron-ci",
			WorkflowId:                 2,
			TriggeredBy:                2,
			CacheLimit:                 5000000000,
			BeforeDockerBuildScripts:   nil,
			AfterDockerBuildScripts:    nil,
			CiArtifactLocation:         "s3://devtron-pro-ci-logs/arsenal-v1/ci-artifacts/2/2.zip",
			CiArtifactBucket:           "devtron-pro-ci-logs",
			CiArtifactFileName:         "arsenal-v1/ci-artifacts/2/2.zip",
			CiArtifactRegion:           "",
			ScanEnabled:                true,
			CloudProvider:              "S3",
			BlobStorageConfigured:      false,
			BlobStorageS3Config:        nil,
			AzureBlobConfig:            nil,
			GcpBlobConfig:              nil,
			BlobStorageLogsKey:         "arsenal-v1/2-ci-1-sslm-1",
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
			OrchestratorHost:          "http://devtroncd-orchestrator-service-prod.devtroncd/webhook/msg/nats",
			OrchestratorToken:         "",
			IsExtRun:                  false,
			ImageRetryCount:           0,
			ImageRetryInterval:        5,
			Type:                      bean2.CI_WORKFLOW_PIPELINE_TYPE,
			WorkflowExecutor:          pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,
		}

		data, _ := workflowServiceImpl.SubmitWorkflow(&workflowRequest)
		createdWf := data.Items[0].Object
		fmt.Println(createdWf)
		verifyMetadata(t, workflowRequest, createdWf)

		spec := createdWf["spec"].(interface{}).(map[string]interface{})
		verifySpec(t, workflowServiceImpl, spec, bean2.CI_WORKFLOW_NAME)
		assert.Equal(t, workflowServiceImpl.ciCdConfig.CiNodeLabelSelector, spec["nodeSelector"].(interface{}))

		assert.Equal(t, 1, len(spec["templates"].(interface{}).([]interface{})))

		var template map[string]interface{}
		if spec["templates"].(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})["name"] == bean2.CI_WORKFLOW_NAME {
			template = spec["templates"].(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})
		} else {
			template = spec["templates"].(interface{}).([]interface{})[1].(interface{}).(map[string]interface{})
		}

		verifyTemplateSpecContainerPort(t, template)

		assert.Equal(t, false, template["archiveLocation"].(interface{}).(map[string]interface{})["archiveLogs"].(interface{}).(bool))

		verifyTemplateSpecContainerEnv(t, template, workflowServiceImpl)

		verifyResourceLimitAndRequest(t, template, workflowServiceImpl)

		verifyToleration(t, workflowServiceImpl, createdWf)
	})

	t.Run("Verify submit workflow for Jobs", func(t *testing.T) {

		workflowRequest := WorkflowRequest{
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
			CiImage:                  "quay.io/devtron/ci-runner:8a9f8b8c-138-15021",
			Namespace:                "devtron-ci",
			WorkflowId:               20,
			TriggeredBy:              2,
			CacheLimit:               5000000000,
			BeforeDockerBuildScripts: nil,
			AfterDockerBuildScripts:  nil,
			CiArtifactLocation:       "devtron/ci-artifacts/20/20.zip",
			CiArtifactBucket:         "",
			CiArtifactFileName:       "devtron/ci-artifacts/20/20.zip",
			CiArtifactRegion:         "",
			ScanEnabled:              false,
			CloudProvider:            "AZURE",
			BlobStorageConfigured:    true,
			BlobStorageS3Config: &blob_storage.BlobStorageS3Config{
				AccessKey:                  "devtrontestblobstorage",
				Passkey:                    "",
				EndpointUrl:                "http://devtron-minio.devtroncd:9000",
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
				AccountName:           "devtrontestblobstorage",
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
			OrchestratorHost:          "http://devtroncd-orchestrator-service-prod.devtroncd/webhook/msg/nats",
			OrchestratorToken:         "",
			IsExtRun:                  false,
			ImageRetryCount:           0,
			ImageRetryInterval:        5,
			Type:                      bean2.JOB_WORKFLOW_PIPELINE_TYPE,
			WorkflowExecutor:          pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,
		}

		data, _ := workflowServiceImpl.SubmitWorkflow(&workflowRequest)
		createdWf := data.Items[0].Object
		fmt.Println(createdWf)

		verifyMetadata(t, workflowRequest, createdWf)

		spec := createdWf["spec"].(interface{}).(map[string]interface{})
		assert.Equal(t, workflowServiceImpl.ciCdConfig.CiNodeLabelSelector, spec["nodeSelector"].(interface{}))
		verifySpec(t, workflowServiceImpl, spec, bean2.CI_WORKFLOW_WITH_STAGES)

		assert.Equal(t, 6, len(spec["templates"].(interface{}).([]interface{})))

		for _, template := range spec["templates"].(interface{}).([]interface{}) {
			if template.(interface{}).(map[string]interface{})["name"] == bean2.CI_WORKFLOW_NAME {

				verifyTemplateSpecContainerPort(t, template.(interface{}).(map[string]interface{}))

				//	verifyTemplateSpecContainerEnvFrom(t, template.(interface{}).(map[string]interface{}))

				verifyTemplateSpecContainerEnv(t, template.(interface{}).(map[string]interface{}), workflowServiceImpl)

				verifyResourceLimitAndRequest(t, template.(interface{}).(map[string]interface{}), workflowServiceImpl)

				verifyTemplateSpecContainerVolumeMounts(t, template.(interface{}).(map[string]interface{}), 4)

				verifyS3BlobStorage(t, template.(interface{}).(map[string]interface{}), workflowServiceImpl, workflowRequest)

			}
			if template.(interface{}).(map[string]interface{})["name"] == bean2.CI_WORKFLOW_WITH_STAGES {

				verifyTemplateSteps(t, template.(interface{}).(map[string]interface{})["steps"].(interface{}).([]interface{}))

			}
			if template.(interface{}).(map[string]interface{})["name"] == "cm-0" {

				verifyTemplateResource(t, template.(interface{}).(map[string]interface{})["resource"].(interface{}).(map[string]interface{}), "create", cmManifest)

			}
			if template.(interface{}).(map[string]interface{})["name"] == "cm-1" {

				verifyTemplateResource(t, template.(interface{}).(map[string]interface{})["resource"].(interface{}).(map[string]interface{}), "create", cmManifest2)

			}
			if template.(interface{}).(map[string]interface{})["name"] == "sec-0" {

				verifyTemplateResource(t, template.(interface{}).(map[string]interface{})["resource"].(interface{}).(map[string]interface{}), "create", secManifest)

			}
			if template.(interface{}).(map[string]interface{})["name"] == "sec-1" {

				verifyTemplateResource(t, template.(interface{}).(map[string]interface{})["resource"].(interface{}).(map[string]interface{}), "create", secManifest2)

			}
		}

		assert.Equal(t, 4, len(spec["volumes"].(interface{}).([]interface{})))

		verifyToleration(t, workflowServiceImpl, createdWf)
	})

	t.Run("Verify submit workflow for External Jobs", func(t *testing.T) {

		workflowRequest := WorkflowRequest{
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
			CiImage:                  "quay.io/devtron/ci-runner:8a9f8b8c-138-15021",
			Namespace:                "2-devtron",
			WorkflowId:               5,
			TriggeredBy:              2,
			CacheLimit:               5000000000,
			BeforeDockerBuildScripts: nil,
			AfterDockerBuildScripts:  nil,
			CiArtifactLocation:       "devtron/ci-artifacts/20/20.zip",
			CiArtifactBucket:         "",
			CiArtifactFileName:       "devtron/ci-artifacts/20/20.zip",
			CiArtifactRegion:         "",
			ScanEnabled:              false,
			CloudProvider:            "AZURE",
			BlobStorageConfigured:    true,
			BlobStorageS3Config: &blob_storage.BlobStorageS3Config{
				AccessKey:                  "devtrontestblobstorage",
				Passkey:                    "",
				EndpointUrl:                "http://devtron-minio.devtroncd:9000",
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
				AccountName:           "devtrontestblobstorage",
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
			OrchestratorHost:          "http://devtroncd-orchestrator-service-prod.devtroncd/webhook/msg/nats",
			OrchestratorToken:         "",
			IsExtRun:                  true,
			ImageRetryCount:           0,
			ImageRetryInterval:        5,
			Type:                      bean2.JOB_WORKFLOW_PIPELINE_TYPE,
			WorkflowExecutor:          pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,

			Env: &repository3.Environment{
				Id:        3,
				Name:      "2-devtron",
				ClusterId: 2,
				Cluster: &repository3.Cluster{
					Id:                 2,
					ClusterName:        "in_cluster",
					ServerUrl:          "https://172.173.222.240:16443",
					PrometheusEndpoint: "",
					Active:             true,
					CdArgoSetup:        true,
					Config: map[string]string{
						"bearer_token": "SGFSZzFoMitKR1ZNUzZzSXdod2tSMEYycmprdit3eVlaRXNxTVl5UHIybz0K",
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
		}

		data, _ := workflowServiceImpl.SubmitWorkflow(&workflowRequest)
		createdWf := data.Items[0].Object
		fmt.Println(createdWf)

		verifyMetadata(t, workflowRequest, createdWf)

		spec := createdWf["spec"].(interface{}).(map[string]interface{})
		verifySpec(t, workflowServiceImpl, spec, bean2.CI_WORKFLOW_WITH_STAGES)

		assert.Equal(t, 10, len(spec["templates"].(interface{}).([]interface{})))

		for _, template := range spec["templates"].(interface{}).([]interface{}) {
			if template.(interface{}).(map[string]interface{})["name"].(interface{}).(string) == bean2.CI_WORKFLOW_NAME {

				verifyTemplateSpecContainerPort(t, template.(interface{}).(map[string]interface{}))

				//verifyTemplateSpecContainerEnvFrom(t, template, 8)

				verifyTemplateSpecContainerEnv(t, template.(interface{}).(map[string]interface{}), workflowServiceImpl)

				verifyResourceLimitAndRequest(t, template.(interface{}).(map[string]interface{}), workflowServiceImpl)

				verifyTemplateSpecContainerVolumeMounts(t, template.(interface{}).(map[string]interface{}), 8)

				verifyS3BlobStorage(t, template.(interface{}).(map[string]interface{}), workflowServiceImpl, workflowRequest)

			}
			if template.(interface{}).(map[string]interface{})["name"].(interface{}).(string) == bean2.CI_WORKFLOW_WITH_STAGES {

				verifyTemplateSteps(t, template.(interface{}).(map[string]interface{})["steps"].(interface{}).([]interface{}))

			}
			if template.(interface{}).(map[string]interface{})["name"].(interface{}).(string) == "cm-0" {

				verifyTemplateResource(t, template.(interface{}).(map[string]interface{})["resource"].(interface{}).(map[string]interface{}), "create", cmManifest)

			}
			if template.(interface{}).(map[string]interface{})["name"].(interface{}).(string) == "cm-1" {

				verifyTemplateResource(t, template.(interface{}).(map[string]interface{})["resource"].(interface{}).(map[string]interface{}), "create", cmManifest2)

			}
			if template.(interface{}).(map[string]interface{})["name"].(interface{}).(string) == "cm-3" {

				verifyTemplateResource(t, template.(interface{}).(map[string]interface{})["resource"].(interface{}).(map[string]interface{}), "create", "")

			}
			if template.(interface{}).(map[string]interface{})["name"].(interface{}).(string) == "cm-2" {

				verifyTemplateResource(t, template.(interface{}).(map[string]interface{})["resource"].(interface{}).(map[string]interface{}), "create", "")

			}
			if template.(interface{}).(map[string]interface{})["name"].(interface{}).(string) == "sec-2" {

				verifyTemplateResource(t, template.(interface{}).(map[string]interface{})["resource"].(interface{}).(map[string]interface{}), "create", "")

			}
			if template.(interface{}).(map[string]interface{})["name"].(interface{}).(string) == "sec-3" {

				verifyTemplateResource(t, template.(interface{}).(map[string]interface{})["resource"].(interface{}).(map[string]interface{}), "create", "")

			}
			if template.(interface{}).(map[string]interface{})["name"].(interface{}).(string) == "sec-1" {

				verifyTemplateResource(t, template.(interface{}).(map[string]interface{})["resource"].(interface{}).(map[string]interface{}), "create", secManifest2)

			}
			if template.(interface{}).(map[string]interface{})["name"].(interface{}).(string) == "sec-0" {

				verifyTemplateResource(t, template.(interface{}).(map[string]interface{})["resource"].(interface{}).(map[string]interface{}), "create", secManifest)

			}
		}

		assert.Equal(t, 8, len(spec["volumes"].(interface{}).([]interface{})))

		verifyToleration(t, workflowServiceImpl, createdWf)
	})

}

func verifyTemplateSteps(t *testing.T, steps []interface{}) {
	count := 0
	for _, step := range steps {
		if step.(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})["template"].(interface{}).(string) == "cm-0" {
			assert.Equal(t, "create-env-cm-0", step.(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})["name"].(interface{}).(string))
			count++
		}
		if step.(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})["template"].(interface{}).(string) == "cm-1" {
			assert.Equal(t, "create-env-cm-1", step.(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})["name"].(interface{}).(string))
			count++

		}
		if step.(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})["template"].(interface{}).(string) == "sec-0" {
			assert.Equal(t, "create-env-sec-0", step.(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})["name"].(interface{}).(string))
			count++
		}
		if step.(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})["template"].(interface{}).(string) == "sec-1" {
			assert.Equal(t, "create-env-sec-1", step.(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})["name"].(interface{}).(string))
			count++
		}
		if step.(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})["template"].(interface{}).(string) == "ci" {
			assert.Equal(t, "run-wf", step.(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})["name"].(interface{}).(string))
			count++
		}
	}
	assert.Equal(t, count, len(steps))

}

func verifyTemplateResource(t *testing.T, resource map[string]interface{}, action string, manifest string) {
	assert.Equal(t, action, resource["action"])
	//assert.Equal(t, manifest, resource["manifest"])
}

func verifyMetadata(t *testing.T, workflowRequest WorkflowRequest, createdWf map[string]interface{}) {
	assert.Equal(t, map[string]string{"devtron.ai/workflow-purpose": "ci"}, createdWf["metadata"].(interface{}).(map[string]interface{})["labels"])
	assert.Equal(t, workflowRequest.WorkflowNamePrefix+"-", createdWf["metadata"].(interface{}).(map[string]interface{})["generateName"])
	assert.Equal(t, workflowRequest.Namespace, createdWf["metadata"].(interface{}).(map[string]interface{})["namespace"])
}

func verifySpec(t *testing.T, workflowServiceImpl *WorkflowServiceImpl, spec map[string]interface{}, stageName string) {
	assert.Equal(t, stageName, spec["entrypoint"].(interface{}).(string))
	assert.Equal(t, workflowServiceImpl.ciCdConfig.CiWorkflowServiceAccount, spec["serviceAccountName"])
}

func verifyTemplateSpecContainerPort(t *testing.T, template map[string]interface{}) {
	assert.Equal(t, 1, len(template["container"].(interface{}).(map[string]interface{})["ports"].(interface{}).([]interface{})))
	assert.Equal(t, "app-data", template["container"].(interface{}).(map[string]interface{})["ports"].(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})["name"])
	assert.Equal(t, int32(9102), template["container"].(interface{}).(map[string]interface{})["ports"].(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})["containerPort"])
}

func verifyTemplateSpecContainerEnv(t *testing.T, template map[string]interface{}, workflowServiceImpl *WorkflowServiceImpl) {
	assert.Equal(t, 3, len(template["container"].(interface{}).(map[string]interface{})["env"].(interface{}).([]interface{})))
	for _, env := range template["container"].(interface{}).(map[string]interface{})["env"].(interface{}).([]interface{}) {
		if env.(interface{}).(map[string]interface{})["name"] == "IMAGE_SCANNER_ENDPOINT" {
			assert.Equal(t, workflowServiceImpl.ciCdConfig.ImageScannerEndpoint, env.(interface{}).(map[string]interface{})["value"])
		}
	}
}

//func verifyTemplateSpecContainerEnvFrom(t *testing.T, template map[string]interface{}) {
//	count := 0
//	for _, envFrom := range template["container"].(interface{}).(map[string]interface{})["env"].(interface{}).([]interface{}) {
//		if envFrom.(interface{}).(map[string]interface{})["secretRef"] != nil {
//			assert.True(t, envFrom.SecretRef.LocalObjectReference.Name != "")
//			count++
//		}
//		if envFrom.ConfigMapRef != nil {
//			assert.True(t, envFrom.ConfigMapRef.LocalObjectReference.Name != "")
//			count++
//		}
//	}
//	assert.Equal(t, 4, count)
//}

func verifyTemplateSpecContainerVolumeMounts(t *testing.T, template map[string]interface{}, count int) {
	assert.Equal(t, count, len(template["container"].(interface{}).(map[string]interface{})["volumeMounts"].(interface{}).([]interface{})))
	for _, volumeMount := range template["container"].(interface{}).(map[string]interface{})["volumeMounts"].(interface{}).([]interface{}) {
		assert.True(t, volumeMount.(interface{}).(map[string]interface{})["Name"] != "")
		assert.True(t, volumeMount.(interface{}).(map[string]interface{})["mountPath"] != "")
	}
}

func verifyResourceLimitAndRequest(t *testing.T, template map[string]interface{}, workflowServiceImpl *WorkflowServiceImpl) {

	resourceLimit := template["container"].(interface{}).(map[string]interface{})["resources"].(interface{}).(map[string]interface{})["limits"].(interface{}).(map[string]interface{})
	resourceRequest := template["container"].(interface{}).(map[string]interface{})["resources"].(interface{}).(map[string]interface{})["requests"].(interface{}).(map[string]interface{})

	//assert.Equal(t, resourceLimit["cpu"], resource.MustParse(workflowServiceImpl.ciConfig.LimitCpu))
	assert.Equal(t, resourceLimit["memory"], resource.MustParse(workflowServiceImpl.ciCdConfig.CiLimitMem))
	//assert.Equal(t, resourceRequest["cpu"], resource.MustParse(workflowServiceImpl.ciConfig.ReqCpu))
	assert.Equal(t, resourceRequest["memory"], resource.MustParse(workflowServiceImpl.ciCdConfig.CiReqMem))
}

func verifyS3BlobStorage(t *testing.T, template map[string]interface{}, workflowServiceImpl *WorkflowServiceImpl, workflowRequest WorkflowRequest) {

	assert.Equal(t, true, template["archiveLocation"].(interface{}).(map[string]interface{})["archiveLogs"].(interface{}).(bool))
	assert.Equal(t, workflowServiceImpl.ciCdConfig.CiDefaultBuildLogsKeyPrefix+"/"+workflowRequest.WorkflowNamePrefix, template["archiveLocation"].(interface{}).(map[string]interface{})["s3"].(interface{}).(map[string]interface{})["key"].(interface{}).(string))
	assert.Equal(t, "accessKey", template["archiveLocation"].(interface{}).(map[string]interface{})["s3"].(interface{}).(map[string]interface{})["accessKeySecret"].(interface{}).(map[string]interface{})["key"].(interface{}).(string))
	assert.Equal(t, "workflow-minio-cred", template["archiveLocation"].(interface{}).(map[string]interface{})["s3"].(interface{}).(map[string]interface{})["accessKeySecret"].(interface{}).(map[string]interface{})["name"].(interface{}).(string))
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

	assert.Equal(t, s3CompatibleEndpointUrl, template["archiveLocation"].(interface{}).(map[string]interface{})["s3"].(interface{}).(map[string]interface{})["endpoint"].(interface{}).(string))
	assert.Equal(t, blobStorageS3Config.CiLogBucketName, template["archiveLocation"].(interface{}).(map[string]interface{})["s3"].(interface{}).(map[string]interface{})["bucket"].(interface{}).(string))
	assert.Equal(t, blobStorageS3Config.CiLogRegion, template["archiveLocation"].(interface{}).(map[string]interface{})["s3"].(interface{}).(map[string]interface{})["region"].(interface{}).(string))
	assert.Equal(t, blobStorageS3Config.IsInSecure, template["archiveLocation"].(interface{}).(map[string]interface{})["s3"].(interface{}).(map[string]interface{})["insecure"].(interface{}).(bool))
	assert.Equal(t, "secretKey", template["archiveLocation"].(interface{}).(map[string]interface{})["s3"].(interface{}).(map[string]interface{})["secretKeySecret"].(interface{}).(map[string]interface{})["key"].(interface{}).(string))
	assert.Equal(t, "workflow-minio-cred", template["archiveLocation"].(interface{}).(map[string]interface{})["s3"].(interface{}).(map[string]interface{})["secretKeySecret"].(interface{}).(map[string]interface{})["name"].(interface{}).(string))
}

func verifyToleration(t *testing.T, workflowServiceImpl *WorkflowServiceImpl, createdWf map[string]interface{}) {
	assert.Equal(t, 1, len(createdWf["spec"].(interface{}).(map[string]interface{})["tolerations"].(interface{}).([]interface{})))
	assert.Equal(t, workflowServiceImpl.ciCdConfig.CiTaintKey, createdWf["spec"].(interface{}).(map[string]interface{})["tolerations"].(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})["key"])
	assert.Equal(t, v12.TolerationOpEqual, createdWf["spec"].(interface{}).(map[string]interface{})["tolerations"].(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})["operator"])
	assert.Equal(t, workflowServiceImpl.ciCdConfig.CiTaintValue, createdWf["spec"].(interface{}).(map[string]interface{})["tolerations"].(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})["value"])
	assert.Equal(t, v12.TaintEffectNoSchedule, createdWf["spec"].(interface{}).(map[string]interface{})["tolerations"].(interface{}).([]interface{})[0].(interface{}).(map[string]interface{})["effect"])
}
