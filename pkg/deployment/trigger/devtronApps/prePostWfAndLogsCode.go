package devtronApps

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	cdWorkflow2 "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	"github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/adapter"
	bean3 "github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/constants"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	util2 "github.com/devtron-labs/devtron/pkg/pipeline/util"
	"k8s.io/client-go/rest"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (impl *HandlerServiceImpl) CancelStage(workflowRunnerId int, forceAbort bool, userId int32) (int, error) {
	workflowRunner, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(workflowRunnerId)
	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return 0, err
	}
	pipeline, err := impl.pipelineRepository.FindById(workflowRunner.CdWorkflow.PipelineId)
	if err != nil {
		impl.logger.Errorw("error while fetching cd pipeline", "err", err)
		return 0, err
	}

	env, err := impl.envRepository.FindById(pipeline.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("could not fetch stage env", "err", err)
		return 0, err
	}

	var clusterBean bean3.ClusterBean
	if env != nil && env.Cluster != nil {
		clusterBean = adapter.GetClusterBean(*env.Cluster)
	}
	clusterConfig := clusterBean.GetClusterConfig()
	var isExtCluster bool
	if workflowRunner.WorkflowType == types.PRE {
		isExtCluster = pipeline.RunPreStageInEnv
	} else if workflowRunner.WorkflowType == types.POST {
		isExtCluster = pipeline.RunPostStageInEnv
	}
	var restConfig *rest.Config
	if isExtCluster {
		restConfig, err = impl.K8sUtil.GetRestConfigByCluster(clusterConfig)
		if err != nil {
			impl.logger.Errorw("error in getting rest config by cluster id", "err", err)
			return 0, err
		}
	}
	// Terminate workflow
	cancelWfDtoRequest := &types.CancelWfRequestDto{
		ExecutorType: workflowRunner.ExecutorType,
		WorkflowName: workflowRunner.Name,
		Namespace:    workflowRunner.Namespace,
		RestConfig:   restConfig,
		IsExt:        isExtCluster,
		Environment:  nil,
	}
	err = impl.workflowService.TerminateWorkflow(cancelWfDtoRequest)
	if err != nil && forceAbort {
		impl.logger.Errorw("error in terminating workflow, with force abort flag as true", "workflowName", workflowRunner.Name, "err", err)
		cancelWfDtoRequest.WorkflowGenerateName = fmt.Sprintf("%d-%s", workflowRunnerId, workflowRunner.Name)
		err1 := impl.workflowService.TerminateDanglingWorkflows(cancelWfDtoRequest)
		if err1 != nil {
			impl.logger.Errorw("error in terminating dangling workflows", "cancelWfDtoRequest", cancelWfDtoRequest, "err", err)
			// ignoring error here in case of force abort, confirmed from product
		}
	} else if err != nil && strings.Contains(err.Error(), "cannot find workflow") {
		return 0, &util.ApiError{Code: "200", HttpStatusCode: http.StatusBadRequest, UserMessage: err.Error()}
	} else if err != nil {
		impl.logger.Error("cannot terminate wf runner", "err", err)
		return 0, err
	}
	if forceAbort {
		err = impl.handleForceAbortCaseForCdStage(workflowRunner, forceAbort)
		if err != nil {
			impl.logger.Errorw("error in handleForceAbortCaseForCdStage", "forceAbortFlag", forceAbort, "workflowRunner", workflowRunner, "err", err)
			return 0, err
		}
		return workflowRunner.Id, nil
	}
	if len(workflowRunner.ImagePathReservationIds) > 0 {
		err := impl.customTagService.DeactivateImagePathReservationByImageIds(workflowRunner.ImagePathReservationIds)
		if err != nil {
			impl.logger.Errorw("error in deactivating image path reservation ids", "err", err)
			return 0, err
		}
	}
	workflowRunner.Status = cdWorkflow2.WorkflowCancel
	workflowRunner.UpdatedOn = time.Now()
	workflowRunner.UpdatedBy = userId
	err = impl.cdWorkflowRunnerService.UpdateCdWorkflowRunnerWithStage(workflowRunner)
	if err != nil {
		impl.logger.Error("cannot update deleted workflow runner status, but wf deleted", "err", err)
		return 0, err
	}
	return workflowRunner.Id, nil
}

func (impl *HandlerServiceImpl) updateWorkflowRunnerForForceAbort(workflowRunner *pipelineConfig.CdWorkflowRunner) error {
	workflowRunner.Status = cdWorkflow2.WorkflowCancel
	workflowRunner.PodStatus = string(bean2.Failed)
	workflowRunner.Message = constants.FORCE_ABORT_MESSAGE_AFTER_STARTING_STAGE
	err := impl.cdWorkflowRunnerService.UpdateCdWorkflowRunnerWithStage(workflowRunner)
	if err != nil {
		impl.logger.Errorw("error in updating workflow status in cd workflow runner in force abort case", "err", err)
		return err
	}
	return nil
}

func (impl *HandlerServiceImpl) handleForceAbortCaseForCdStage(workflowRunner *pipelineConfig.CdWorkflowRunner, forceAbort bool) error {
	isWorkflowInNonTerminalStage := workflowRunner.Status == string(v1alpha1.NodePending) || workflowRunner.Status == string(v1alpha1.NodeRunning)
	if !isWorkflowInNonTerminalStage {
		if forceAbort {
			return impl.updateWorkflowRunnerForForceAbort(workflowRunner)
		} else {
			return &util.ApiError{Code: "200", HttpStatusCode: 400, UserMessage: "cannot cancel stage, stage not in progress"}
		}
	}
	//this arises when someone deletes the workflow in resource browser and wants to force abort a cd stage(pre/post)
	if workflowRunner.Status == string(v1alpha1.NodeRunning) && forceAbort {
		return impl.updateWorkflowRunnerForForceAbort(workflowRunner)
	}
	return nil
}

func (impl *HandlerServiceImpl) DownloadCdWorkflowArtifacts(buildId int) (*os.File, error) {
	wfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(buildId)
	if err != nil {
		impl.logger.Errorw("unable to fetch ciWorkflow", "err", err)
		return nil, err
	}
	useExternalBlobStorage := pipeline.IsExternalBlobStorageEnabled(wfr.IsExternalRun(), impl.config.UseBlobStorageConfigInCdWorkflow)
	if !wfr.BlobStorageEnabled {
		return nil, errors.New("logs-not-stored-in-repository")
	}

	cdConfigLogsBucket := impl.config.GetDefaultBuildLogsBucket()
	cdConfigCdCacheRegion := impl.config.GetDefaultCdLogsBucketRegion()

	item := strconv.Itoa(wfr.Id)
	awsS3BaseConfig := &blob_storage.AwsS3BaseConfig{
		AccessKey:         impl.config.BlobStorageS3AccessKey,
		Passkey:           impl.config.BlobStorageS3SecretKey,
		EndpointUrl:       impl.config.BlobStorageS3Endpoint,
		IsInSecure:        impl.config.BlobStorageS3EndpointInsecure,
		BucketName:        cdConfigLogsBucket,
		Region:            cdConfigCdCacheRegion,
		VersioningEnabled: impl.config.BlobStorageS3BucketVersioned,
	}
	azureBlobBaseConfig := &blob_storage.AzureBlobBaseConfig{
		Enabled:           impl.config.CloudProvider == types.BLOB_STORAGE_AZURE,
		AccountKey:        impl.config.AzureAccountKey,
		AccountName:       impl.config.AzureAccountName,
		BlobContainerName: impl.config.AzureBlobContainerCiLog,
	}
	gcpBlobBaseConfig := &blob_storage.GcpBlobBaseConfig{
		BucketName:             cdConfigLogsBucket,
		CredentialFileJsonData: impl.config.BlobStorageGcpCredentialJson,
	}
	cdArtifactLocationFormat := impl.config.GetArtifactLocationFormat()
	key := fmt.Sprintf(cdArtifactLocationFormat, wfr.CdWorkflow.Id, wfr.Id)
	if len(wfr.CdArtifactLocation) != 0 && util2.IsValidUrlSubPath(wfr.CdArtifactLocation) {
		key = wfr.CdArtifactLocation
	} else if util2.IsValidUrlSubPath(key) {
		impl.cdWorkflowRepository.MigrateCdArtifactLocation(wfr.Id, key)
	}
	baseLogLocationPathConfig := impl.config.BaseLogLocationPath
	blobStorageService := blob_storage.NewBlobStorageServiceImpl(nil)
	destinationKey := filepath.Clean(filepath.Join(baseLogLocationPathConfig, item))
	request := &blob_storage.BlobStorageRequest{
		StorageType:         impl.config.CloudProvider,
		SourceKey:           key,
		DestinationKey:      destinationKey,
		AzureBlobBaseConfig: azureBlobBaseConfig,
		AwsS3BaseConfig:     awsS3BaseConfig,
		GcpBlobBaseConfig:   gcpBlobBaseConfig,
	}
	if useExternalBlobStorage {
		clusterConfig, err := impl.clusterService.GetClusterConfigByClusterId(wfr.CdWorkflow.Pipeline.Environment.ClusterId)
		if err != nil {
			impl.logger.Errorw("GetClusterConfigByClusterId, error in fetching clusterConfig", "err", err, "clusterId", wfr.CdWorkflow.Pipeline.Environment.ClusterId)
			return nil, err
		}
		// fetch extClusterBlob cm and cs from k8s client, if they are present then read creds
		// from them else return.
		cmConfig, secretConfig, err := impl.blobConfigStorageService.FetchCmAndSecretBlobConfigFromExternalCluster(clusterConfig, wfr.Namespace)
		if err != nil {
			impl.logger.Errorw("error in fetching config map and secret from external cluster", "err", err, "clusterConfig", clusterConfig)
			return nil, err
		}
		request = pipeline.UpdateRequestWithExtClusterCmAndSecret(request, cmConfig, secretConfig)
	}
	_, numBytes, err := blobStorageService.Get(request)
	if err != nil {
		impl.logger.Errorw("error occurred while downloading file", "request", request, "error", err)
		return nil, errors.New("failed to download resource")
	}

	file, err := os.Open(destinationKey)
	if err != nil {
		impl.logger.Errorw("unable to open file", "file", item, "err", err)
		return nil, errors.New("unable to open file")
	}

	impl.logger.Infow("Downloaded ", "name", file.Name(), "bytes", numBytes)
	return file, nil
}

func (impl *HandlerServiceImpl) GetRunningWorkflowLogs(environmentId int, pipelineId int, wfrId int) (*bufio.Reader, func() error, error) {
	cdWorkflow, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(wfrId)
	if err != nil {
		impl.logger.Errorw("error on fetch wf runner", "err", err)
		return nil, nil, err
	}

	env, err := impl.envRepository.FindById(environmentId)
	if err != nil {
		impl.logger.Errorw("could not fetch stage env", "err", err)
		return nil, nil, err
	}

	pipeline, err := impl.pipelineRepository.FindById(cdWorkflow.CdWorkflow.PipelineId)
	if err != nil {
		impl.logger.Errorw("error while fetching cd pipeline", "err", err)
		return nil, nil, err
	}
	var clusterBean bean3.ClusterBean
	if env != nil && env.Cluster != nil {
		clusterBean = adapter.GetClusterBean(*env.Cluster)
	}
	clusterConfig := clusterBean.GetClusterConfig()
	var isExtCluster bool
	if cdWorkflow.WorkflowType == types.PRE {
		isExtCluster = pipeline.RunPreStageInEnv
	} else if cdWorkflow.WorkflowType == types.POST {
		isExtCluster = pipeline.RunPostStageInEnv
	}
	return impl.getWorkflowLogs(pipelineId, cdWorkflow, clusterConfig, isExtCluster)
}

func (impl *HandlerServiceImpl) getWorkflowLogs(pipelineId int, cdWorkflow *pipelineConfig.CdWorkflowRunner, clusterConfig *k8s.ClusterConfig, runStageInEnv bool) (*bufio.Reader, func() error, error) {
	cdLogRequest := types.BuildLogRequest{
		PodName:   cdWorkflow.PodName,
		Namespace: cdWorkflow.Namespace,
	}

	logStream, cleanUp, err := impl.ciLogService.FetchRunningWorkflowLogs(cdLogRequest, clusterConfig, runStageInEnv)
	if logStream == nil || err != nil {
		if !cdWorkflow.BlobStorageEnabled {
			return nil, nil, errors.New("logs-not-stored-in-repository")
		} else if string(v1alpha1.NodeSucceeded) == cdWorkflow.Status || string(v1alpha1.NodeError) == cdWorkflow.Status || string(v1alpha1.NodeFailed) == cdWorkflow.Status || cdWorkflow.Status == cdWorkflow2.WorkflowCancel {
			impl.logger.Debugw("pod is not live", "podName", cdWorkflow.PodName, "err", err)
			return impl.getLogsFromRepository(pipelineId, cdWorkflow, clusterConfig, runStageInEnv)
		}
		if err != nil {
			impl.logger.Errorw("err on fetch workflow logs", "err", err)
			return nil, nil, err
		} else if logStream == nil {
			return nil, cleanUp, fmt.Errorf("no logs found for pod %s", cdWorkflow.PodName)
		}
	}
	logReader := bufio.NewReader(logStream)
	return logReader, cleanUp, err
}

func (impl *HandlerServiceImpl) getLogsFromRepository(pipelineId int, cdWorkflow *pipelineConfig.CdWorkflowRunner, clusterConfig *k8s.ClusterConfig, isExt bool) (*bufio.Reader, func() error, error) {
	impl.logger.Debug("getting historic logs", "pipelineId", pipelineId)

	cdConfigLogsBucket := impl.config.GetDefaultBuildLogsBucket() // TODO -fixme
	cdConfigCdCacheRegion := impl.config.GetDefaultCdLogsBucketRegion()

	cdLogRequest := types.BuildLogRequest{
		PipelineId:    cdWorkflow.CdWorkflow.PipelineId,
		WorkflowId:    cdWorkflow.Id,
		PodName:       cdWorkflow.PodName,
		LogsFilePath:  cdWorkflow.LogLocation, // impl.ciCdConfig.CiDefaultBuildLogsKeyPrefix + "/" + cdWorkflow.Name + "/main.log", //TODO - fixme
		CloudProvider: impl.config.CloudProvider,
		AzureBlobConfig: &blob_storage.AzureBlobBaseConfig{
			Enabled:           impl.config.CloudProvider == types.BLOB_STORAGE_AZURE,
			AccountName:       impl.config.AzureAccountName,
			BlobContainerName: impl.config.AzureBlobContainerCiLog,
			AccountKey:        impl.config.AzureAccountKey,
		},
		AwsS3BaseConfig: &blob_storage.AwsS3BaseConfig{
			AccessKey:         impl.config.BlobStorageS3AccessKey,
			Passkey:           impl.config.BlobStorageS3SecretKey,
			EndpointUrl:       impl.config.BlobStorageS3Endpoint,
			IsInSecure:        impl.config.BlobStorageS3EndpointInsecure,
			BucketName:        cdConfigLogsBucket,
			Region:            cdConfigCdCacheRegion,
			VersioningEnabled: impl.config.BlobStorageS3BucketVersioned,
		},
		GcpBlobBaseConfig: &blob_storage.GcpBlobBaseConfig{
			BucketName:             cdConfigLogsBucket,
			CredentialFileJsonData: impl.config.BlobStorageGcpCredentialJson,
		},
	}
	useExternalBlobStorage := pipeline.IsExternalBlobStorageEnabled(isExt, impl.config.UseBlobStorageConfigInCdWorkflow)
	if useExternalBlobStorage {
		// fetch extClusterBlob cm and cs from k8s client, if they are present then read creds
		// from them else return.
		cmConfig, secretConfig, err := impl.blobConfigStorageService.FetchCmAndSecretBlobConfigFromExternalCluster(clusterConfig, cdWorkflow.Namespace)
		if err != nil {
			impl.logger.Errorw("error in fetching config map and secret from external cluster", "err", err, "clusterConfig", clusterConfig)
			return nil, nil, err
		}
		rq := &cdLogRequest
		rq.SetBuildLogRequest(cmConfig, secretConfig)
	}

	impl.logger.Debugw("s3 log req ", "pipelineId", pipelineId, "runnerId", cdWorkflow.Id)
	oldLogsStream, cleanUp, err := impl.ciLogService.FetchLogs(impl.config.BaseLogLocationPath, cdLogRequest)
	if err != nil {
		impl.logger.Errorw("err", err)
		return nil, nil, err
	}
	logReader := bufio.NewReader(oldLogsStream)
	return logReader, cleanUp, err
}
