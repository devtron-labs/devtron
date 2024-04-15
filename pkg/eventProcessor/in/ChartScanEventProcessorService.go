package in

import (
	"context"
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/common-lib/utils/k8sObjectsUtil"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	security2 "github.com/devtron-labs/devtron/internal/sql/repository/security"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/security"
	"github.com/devtron-labs/devtron/pkg/sql"
	"go.uber.org/zap"
	"time"
)

type ChartScanEventProcessorImpl struct {
	logger                  *zap.SugaredLogger
	pubSubClient            *pubsub.PubSubClientServiceImpl
	helmAppService          service.HelmAppService
	policyService           security.PolicyService
	imageScanDeployInfoRepo security2.ImageScanDeployInfoRepository
	imageScanHistoryRepo    security2.ImageScanHistoryRepository
}

func NewChartScanEventProcessorImpl(logger *zap.SugaredLogger,
	pubSubClient *pubsub.PubSubClientServiceImpl,
	helmAppService service.HelmAppService,
	policyService security.PolicyService,
	imageScanDeployInfoRepo security2.ImageScanDeployInfoRepository,
	imageScanHistoryRepo security2.ImageScanHistoryRepository,
) *ChartScanEventProcessorImpl {
	return &ChartScanEventProcessorImpl{
		logger:                  logger,
		pubSubClient:            pubSubClient,
		helmAppService:          helmAppService,
		policyService:           policyService,
		imageScanDeployInfoRepo: imageScanDeployInfoRepo,
		imageScanHistoryRepo:    imageScanHistoryRepo,
	}
}

func (impl *ChartScanEventProcessorImpl) SubscribeChartScanEvent() error {
	callback := func(msg *model.PubSubMsg) {
		request := &appStoreBean.InstallAppVersionDTO{}
		err := json.Unmarshal([]byte(msg.Data), &request)
		if err != nil {
			impl.logger.Error("Error while unmarshalling deployPayload json object", "error", err)
			return
		}

		impl.processScanEventForChartInstall(request)
	}

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		payload := &appStoreBean.InstallAppVersionDTO{}
		err := json.Unmarshal([]byte(msg.Data), &payload)
		if err != nil {
			return "error while unmarshalling InstallAppVersionDTO json object", []interface{}{"error", err}
		}
		return "got message for CHART_SCAN_TOPIC", []interface{}{"installedAppVersionId", payload.InstalledAppVersionId, "installedAppVersionHistoryId", payload.InstalledAppVersionHistoryId}
	}

	err := impl.pubSubClient.Subscribe(pubsub.CHART_SCAN_TOPIC, callback, loggerFunc)
	if err != nil {
		impl.logger.Error("err", err)
		return err
	}
	return nil
}

func (impl *ChartScanEventProcessorImpl) processScanEventForChartInstall(request *appStoreBean.InstallAppVersionDTO) {
	manifestRequest := impl.buildTemplateChartRequest(request)
	dockerImages, err := impl.getDockerImages(manifestRequest)
	if err != nil {
		impl.logger.Error("Error on fetching docker images from generated manifest", "error", err, "manifestRequest", manifestRequest)
		return
	}

	historyIds := make([]int, 0)
	for _, image := range dockerImages {
		history := impl.buildImageScanHistoryObject(image)
		err := impl.imageScanHistoryRepo.Save(history)
		if err != nil {
			impl.logger.Error("Error on saving ImageScanExecutionHistory", "error", err, "history", history)
			continue
		}
		impl.sendForScan(history.Id, image)
		historyIds = append(historyIds, history.Id)
	}
	scanDeployObject := impl.buildScanDeployInfoObject(historyIds, request)
	err = impl.imageScanDeployInfoRepo.Save(scanDeployObject)
	if err != nil {
		impl.logger.Error("Error on saving ImageScanDeployInfo", "error", err, "err")
		return
	}
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
	}
	return manifestRequest
}

func (impl *ChartScanEventProcessorImpl) getDockerImages(manifestRequest openapi2.TemplateChartRequest) ([]string, error) {
	ctx := context.Background()
	resp, err := impl.helmAppService.TemplateChart(ctx, &manifestRequest)
	if err != nil {
		impl.logger.Errorw("error in generating manifest", "err", err, "request", manifestRequest)
		return nil, err
	}
	images, err := k8sObjectsUtil.ExtractImageFromManifestYaml(resp.GetManifest())
	return images, err
}

func (impl *ChartScanEventProcessorImpl) sendForScan(historyId int, image string) {
	err := impl.policyService.SendEventToClairUtilityAsync(&security.ScanEvent{
		Image:         image,
		UserId:        bean.SYSTEM_USER_ID,
		ScanHistoryId: historyId,
	})
	if err != nil {
		impl.logger.Errorw("error in sending image scan event", "err", err, "image", image)
	}
}
