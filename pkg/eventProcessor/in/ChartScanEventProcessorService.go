package in

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/common-lib-private/utils/k8s"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/common-lib/utils/k8sObjectsUtil"
	bean3 "github.com/devtron-labs/devtron/api/helm-app/bean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	security2 "github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"github.com/devtron-labs/devtron/pkg/generateManifest"
	"github.com/devtron-labs/devtron/pkg/security"
	"github.com/devtron-labs/devtron/pkg/sql"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const CHART_SCAN_WORKING_DIR_PATH = "/tmp/scan/charts"

type ChartScanEventProcessorImpl struct {
	logger                  *zap.SugaredLogger
	pubSubClient            *pubsub.PubSubClientServiceImpl
	helmAppService          service.HelmAppService
	K8sUtil                 *k8s.K8sUtilExtended
	helmAppClient           gRPC.HelmAppClient
	policyService           security.PolicyService
	imageScanDeployInfoRepo security2.ImageScanDeployInfoRepository
	imageScanHistoryRepo    security2.ImageScanHistoryRepository
	chartTemplateService    util.ChartTemplateService
}

func NewChartScanEventProcessorImpl(logger *zap.SugaredLogger,
	pubSubClient *pubsub.PubSubClientServiceImpl,
	helmAppService service.HelmAppService,
	helmAppClient gRPC.HelmAppClient,
	policyService security.PolicyService,
	imageScanDeployInfoRepo security2.ImageScanDeployInfoRepository,
	imageScanHistoryRepo security2.ImageScanHistoryRepository,
	K8sUtil *k8s.K8sUtilExtended,
	chartTemplateService util.ChartTemplateService,
) *ChartScanEventProcessorImpl {
	return &ChartScanEventProcessorImpl{
		logger:                  logger,
		pubSubClient:            pubSubClient,
		helmAppService:          helmAppService,
		policyService:           policyService,
		imageScanDeployInfoRepo: imageScanDeployInfoRepo,
		imageScanHistoryRepo:    imageScanHistoryRepo,
		helmAppClient:           helmAppClient,
		K8sUtil:                 K8sUtil,
		chartTemplateService:    chartTemplateService,
	}
}

func (impl *ChartScanEventProcessorImpl) SubscribeChartScanEvent() error {
	callback := func(msg *model.PubSubMsg) {
		request := &bean2.ChartScanEventBean{}
		err := json.Unmarshal([]byte(msg.Data), request)
		if err != nil {
			impl.logger.Error("Error while unmarshalling deployPayload json object", "error", err)
			return
		}

		impl.processScanEventForChartInstall(request)
	}

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		request := &bean2.ChartScanEventBean{}
		err := json.Unmarshal([]byte(msg.Data), &request)
		if err != nil {
			return "error while unmarshalling ChartScanEventBean json object", []interface{}{"error", err}
		}
		if payload := request.AppVersionDto; payload != nil {
			return "got message for CHART_SCAN_TOPIC", []interface{}{"installedAppVersionId", payload.InstalledAppVersionId, "installedAppVersionHistoryId", payload.InstalledAppVersionHistoryId}
		}

		if payload := request.DevtronAppDto; payload != nil {
			return "got message for CHART_SCAN_TOPIC", []interface{}{"cdWorkflowRunnerId", payload.CdWorkflowId, "chartName", payload.ChartName}
		}
		return "got message for CHART_SCAN_TOPIC", []interface{}{}
	}

	err := impl.pubSubClient.Subscribe(pubsub.CHART_SCAN_TOPIC, callback, loggerFunc)
	if err != nil {
		impl.logger.Error("err", err)
		return err
	}
	return nil
}

func (impl *ChartScanEventProcessorImpl) processScanEventForChartInstall(request *bean2.ChartScanEventBean) {
	appVersionDto := request.AppVersionDto
	var manifest string
	var chartBytes []byte
	var valuesYaml string
	ctx := context.Background()
	isHelmApp := appVersionDto != nil
	historyId := 0
	if isHelmApp {
		manifestRequest := impl.buildTemplateChartRequest(appVersionDto)
		resp, err := impl.helmAppService.TemplateChart(ctx, &manifestRequest)
		if err != nil {
			impl.logger.Errorw("error in generating manifest", "err", err, "request", manifestRequest)
			return
		}
		chartBytes = resp.ChartBytes
		valuesYaml = appVersionDto.ValuesOverrideYaml
		manifest = resp.GetManifest()
		historyId = appVersionDto.InstalledAppVersionHistoryId
	} else {
		devtronAppDto := request.DevtronAppDto
		chartBytes = devtronAppDto.ChartContent
		installReleaseReq, err := impl.buildInstallRequest(devtronAppDto)
		chartResponse, err := impl.helmAppClient.TemplateChart(ctx, installReleaseReq)
		if err != nil {
			impl.logger.Errorw("error in generating manifest", "err", err, "request", installReleaseReq)
			return
		}
		manifest = chartResponse.GeneratedManifest
		historyId = devtronAppDto.CdWorkflowId
		valuesYaml = devtronAppDto.ValuesYaml
	}
	dockerImages := k8sObjectsUtil.ExtractImageFromManifestYaml(manifest)

	for _, image := range dockerImages {
		impl.sendForScan(historyId, image, nil, "", isHelmApp)
	}
	impl.sendForScan(historyId, "", chartBytes, valuesYaml, isHelmApp)
}

func (impl *ChartScanEventProcessorImpl) buildInstallRequest(devtronAppDto *bean2.DevtronAppDto) (*gRPC.InstallReleaseRequest, error) {
	config, err := impl.helmAppService.GetClusterConf(bean3.DEFAULT_CLUSTER_ID)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "clusterId", 1, "err", err)
		return nil, err
	}
	k8sServerVersion, err := impl.K8sUtil.GetKubeVersion()
	if err != nil {
		impl.logger.Errorw("exception caught in getting k8sServerVersion", "err", err)
		return nil, err
	}

	// TODO: refactor this later
	var b bytes.Buffer
	writer := gzip.NewWriter(&b)
	_, err = writer.Write(devtronAppDto.ChartContent)
	if err != nil {
		impl.logger.Errorw("error on helm install custom while writing chartContent", "err", err)
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		impl.logger.Errorw("error on helm install custom while writing chartContent", "err", err)
		return nil, err
	}

	if _, err := os.Stat(CHART_SCAN_WORKING_DIR_PATH); os.IsNotExist(err) {
		err := os.MkdirAll(CHART_SCAN_WORKING_DIR_PATH, os.ModePerm)
		if err != nil {
			impl.logger.Errorw("err in creating dir", "err", err)
			return nil, err
		}
	}

	dir := impl.chartTemplateService.GetDir()
	referenceChartDir := filepath.Join(CHART_SCAN_WORKING_DIR_PATH, dir)
	referenceChartDir = fmt.Sprintf("%s.tgz", referenceChartDir)
	defer impl.chartTemplateService.CleanDir(referenceChartDir)
	err = ioutil.WriteFile(referenceChartDir, b.Bytes(), os.ModePerm)
	if err != nil {
		impl.logger.Errorw("error on helm install custom while writing chartContent", "err", err)
		return nil, err
	}

	chartBytes, err := os.ReadFile(referenceChartDir)
	if err != nil {
		fmt.Println("error in reading chartdata from the file ", " filePath : ", referenceChartDir, " err : ", err)
	}

	installReleaseReq := &gRPC.InstallReleaseRequest{
		ReleaseIdentifier: generateManifest.ReleaseIdentifier,
		K8SVersion:        k8sServerVersion.String(),
		ChartVersion:      devtronAppDto.ChartVersion,
		ChartName:         devtronAppDto.ChartName,
		ChartRepository:   generateManifest.ChartRepository,
		ChartContent: &gRPC.ChartContent{
			Content: chartBytes,
		},
		ValuesYaml: devtronAppDto.ValuesYaml,
	}
	installReleaseReq.ReleaseIdentifier.ClusterConfig = config
	return installReleaseReq, nil
}

func (impl *ChartScanEventProcessorImpl) buildImageScanHistoryObject(image string) *security2.ImageScanExecutionHistory {
	history := &security2.ImageScanExecutionHistory{
		Id:            0,
		Image:         image,
		ExecutionTime: time.Now(),
		ExecutedBy:    bean.SYSTEM_USER_ID,
	}
	return history
}

func (impl *ChartScanEventProcessorImpl) buildScanDeployInfoObject(historyIds []int, request *appStoreBean.InstallAppVersionDTO) *security2.ImageScanDeployInfo {
	scanDeployObject := &security2.ImageScanDeployInfo{
		ImageScanExecutionHistoryId: historyIds,
		ScanObjectMetaId:            request.InstalledAppVersionHistoryId,
		ObjectType:                  security2.ScanObjectType_CHART_HISTORY,
		EnvId:                       request.EnvironmentId,
		ClusterId:                   request.ClusterId,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: bean.SYSTEM_USER_ID,
			UpdatedOn: time.Now(),
			UpdatedBy: bean.SYSTEM_USER_ID,
		},
	}
	return scanDeployObject
}

func (impl *ChartScanEventProcessorImpl) buildTemplateChartRequest(request *appStoreBean.InstallAppVersionDTO) openapi2.TemplateChartRequest {
	envId := int32(request.EnvironmentId)
	clusterId := int32(request.ClusterId)
	namespace := request.Namespace
	appName := request.AppName
	iavId := int32(request.AppStoreVersion)

	manifestRequest := openapi2.TemplateChartRequest{
		EnvironmentId:                &envId,
		ClusterId:                    &clusterId,
		Namespace:                    &namespace,
		ReleaseName:                  &appName,
		AppStoreApplicationVersionId: &iavId,
		ValuesYaml:                   &request.ValuesOverrideYaml,
		ReturnChartBytes:             true,
	}
	return manifestRequest
}

func (impl *ChartScanEventProcessorImpl) getDockerImages(manifestRequest openapi2.TemplateChartRequest) ([]string, *openapi2.TemplateChartResponse, error) {
	ctx := context.Background()
	resp, err := impl.helmAppService.TemplateChart(ctx, &manifestRequest)
	if err != nil {
		impl.logger.Errorw("error in generating manifest", "err", err, "request", manifestRequest)
		return nil, nil, err
	}
	images := k8sObjectsUtil.ExtractImageFromManifestYaml(resp.GetManifest())
	return images, resp, err
}

func (impl *ChartScanEventProcessorImpl) sendForScan(historyId int, image string, chartBytes []byte, valuesYaml string, isHelmApp bool) {

	var err error
	var scanEvent *security.ScanEvent
	if len(image) > 0 {
		scanEvent = &security.ScanEvent{
			Image:         image,
			UserId:        bean.SYSTEM_USER_ID,
			SourceType:    security2.SourceTypeImage,
			SourceSubType: security2.SourceSubTypeManifest,
		}
	} else {
		scanEvent = &security.ScanEvent{
			UserId:        bean.SYSTEM_USER_ID,
			SourceType:    security2.SourceTypeCode,
			SourceSubType: security2.SourceSubTypeManifest,
			ManifestData: &security.ManifestData{
				ChartData:  chartBytes,
				ValuesYaml: []byte(valuesYaml),
			},
		}
	}
	if isHelmApp {
		scanEvent.ChartHistoryId = historyId
	} else {
		scanEvent.CdWorkflowId = historyId
	}
	err = impl.policyService.SendScanEventAsync(scanEvent)
	if err != nil {
		impl.logger.Errorw("error in sending image scan event", "err", err, "image", image)
	}
}
