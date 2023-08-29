package pipeline

import (
	"fmt"
	"github.com/devtron-labs/authenticator/client"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"github.com/devtron-labs/devtron/internal/sql/repository"
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
	"testing"
)

var pipelineId = 0

func getWorkflowServiceImpl(t *testing.T) *WorkflowServiceImpl {
	logger, dbConnection := getDbConnAndLoggerService(t)
	ciConfig, _ := GetCiConfig()
	newGlobalCMCSRepositoryImpl := repository.NewGlobalCMCSRepositoryImpl(logger, dbConnection)
	globalCMCSServiceImpl := NewGlobalCMCSServiceImpl(logger, newGlobalCMCSRepositoryImpl)
	newEnvConfigOverrideRepository := chartConfig.NewEnvConfigOverrideRepository(dbConnection)
	newConfigMapRepositoryImpl := chartConfig.NewConfigMapRepositoryImpl(logger, db)
	newChartRepository := chartRepoRepository.NewChartRepository(db)
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
	workflowServiceImpl, _ := NewWorkflowServiceImpl(logger, ciConfig, globalCMCSServiceImpl, appService, newConfigMapRepositoryImpl, k8sUtil, k8sCommonServiceImpl)
	return workflowServiceImpl
}

func TestWorkflowServiceImpl_SubmitWorkflow(t *testing.T) {
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
		CiProjectDetails: []CiProjectDetails{{
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
			GitOptions: GitOptions{
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
		ContainerResources:       ContainerResources{},
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
		BlobStorageConfigured:    false,
		BlobStorageS3Config: &blob_storage.BlobStorageS3Config{
			AccessKey:                  "",
			Passkey:                    "",
			EndpointUrl:                "",
			IsInSecure:                 false,
			CiLogBucketName:            "devtron-pro-ci-logs",
			CiLogRegion:                "us-east-2",
			CiLogBucketVersioning:      true,
			CiCacheBucketName:          "ci-caching",
			CiCacheRegion:              "us-east-2",
			CiCacheBucketVersioning:    true,
			CiArtifactBucketName:       "devtron-pro-ci-logs",
			CiArtifactRegion:           "us-east-2",
			CiArtifactBucketVersioning: true,
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
		AppId:                     1,
		EnvironmentId:             0,
		OrchestratorHost:          "http://devtroncd-orchestrator-service-prod.devtroncd/webhook/msg/nats",
		OrchestratorToken:         "",
		IsExtRun:                  false,
		ImageRetryCount:           0,
		ImageRetryInterval:        5,
	}
	workflowServiceImpl := getWorkflowServiceImpl(t)
	createdWf, _ := workflowServiceImpl.SubmitWorkflow(&workflowRequest, nil, nil, false)

	fmt.Println(createdWf)

}
