package telemetry

import (
	"context"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	bean3 "github.com/devtron-labs/devtron/pkg/attributes/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/go-pg/pg"
	"github.com/tidwall/gjson"
	"k8s.io/apimachinery/pkg/version"
)

// EA-mode compatible telemetry queries - no imports needed for current methods

func (impl *TelemetryEventClientImpl) getHelmAppCount() int {
	if impl.installedAppReadService == nil {
		impl.logger.Warnw("installedAppReadService not available for helm app count")
		return -1
	}
	count, err := impl.installedAppReadService.GetActiveInstalledAppCount()
	if err != nil {
		impl.logger.Errorw("error getting helm app count", "err", err)
		return -1
	}
	return count
}

// EA-mode compatible telemetry methods

func (impl *TelemetryEventClientImpl) getClusterCounts() (physicalCount int, isolatedCount int) {
	clusters, err := impl.clusterService.FindAllActive()
	if err != nil {
		impl.logger.Errorw("error getting cluster counts", "err", err)
		return -1, -1
	}

	physicalCount = 0
	isolatedCount = 0

	for _, cluster := range clusters {
		if cluster.IsVirtualCluster {
			isolatedCount++
		} else {
			physicalCount++
		}
	}

	return physicalCount, isolatedCount
}

// Note: FULL-mode specific methods like getDevtronAppCount, getJobCount, etc.
// are now implemented in TelemetryEventClientImplExtended in telemetryQueriesExtended.go

func (impl *TelemetryEventClientImpl) getActiveUsersLast30Days() int {
	if impl.userAuditService == nil {
		impl.logger.Warnw("userAuditService not available for active users count")
		return -1
	}

	count, err := impl.userAuditService.GetActiveUsersCountInLast30Days()
	if err != nil {
		impl.logger.Errorw("error getting active users count in last 30 days", "err", err)
		return -1
	}

	impl.logger.Debugw("counted active users in last 30 days", "count", count)
	return count
}

func (impl *TelemetryEventClientImpl) GetSummaryDetailsForTelemetry() (cluster []bean.ClusterBean, user []bean2.UserInfo,
	k8sServerVersion *version.Info, hostURL bool, ssoSetup bool, HelmAppAccessCount string, ChartStoreVisitCount string,
	SkippedOnboarding bool, HelmAppUpdateCounter string, helmChartSuccessfulDeploymentCount int, ExternalHelmAppClusterCount map[int32]int) {

	discoveryClient, err := impl.K8sUtil.GetK8sDiscoveryClientInCluster()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}
	k8sServerVersion, err = discoveryClient.ServerVersion()
	if err != nil {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}

	users, err := impl.userService.GetAll()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summery event", "err", err)
		return
	}

	clusters, err := impl.clusterService.FindAllActive()

	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("exception caught inside telemetry summary event", "err", err)
		return
	}

	hostURL = false

	attribute, err := impl.attributeRepo.FindByKey(bean3.HostUrlKey)
	if err == nil && attribute.Id > 0 {
		hostURL = true
	}

	attribute, err = impl.attributeRepo.FindByKey("HelmAppAccessCounter")

	if err == nil {
		HelmAppAccessCount = attribute.Value
	}

	attribute, err = impl.attributeRepo.FindByKey("ChartStoreVisitCount")

	if err == nil {
		ChartStoreVisitCount = attribute.Value
	}

	attribute, err = impl.attributeRepo.FindByKey("HelmAppUpdateCounter")

	if err == nil {
		HelmAppUpdateCounter = attribute.Value
	}

	helmChartSuccessfulDeploymentCount, err = impl.installedAppReadService.GetDeploymentSuccessfulStatusCountForTelemetry()

	//externalHelmCount := make(map[int32]int)
	ExternalHelmAppClusterCount = make(map[int32]int)

	for _, clusterDetail := range clusters {
		req := &gRPC.AppListRequest{}
		config := &gRPC.ClusterConfig{
			ApiServerUrl:          clusterDetail.ServerUrl,
			Token:                 clusterDetail.Config[commonBean.BearerToken],
			ClusterId:             int32(clusterDetail.Id),
			ClusterName:           clusterDetail.ClusterName,
			InsecureSkipTLSVerify: clusterDetail.InsecureSkipTLSVerify,
		}

		if clusterDetail.InsecureSkipTLSVerify == false {
			config.KeyData = clusterDetail.Config[commonBean.TlsKey]
			config.CertData = clusterDetail.Config[commonBean.CertData]
			config.CaData = clusterDetail.Config[commonBean.CertificateAuthorityData]
		}
		req.Clusters = append(req.Clusters, config)
		applicationStream, err := impl.helmAppClient.ListApplication(context.Background(), req)
		if err == nil {
			clusterList, err1 := applicationStream.Recv()
			if err1 != nil {
				impl.logger.Errorw("error in list helm applications streams recv", "err", err)
			}
			if err1 != nil && clusterList != nil && !clusterList.Errored {
				ExternalHelmAppClusterCount[clusterList.ClusterId] = len(clusterList.DeployedAppDetail)
			}
		} else {
			impl.logger.Errorw("error while fetching list application from kubelink", "err", err)
		}
	}

	//getting userData from emailId
	userData, err := impl.userAttributesRepository.GetUserDataByEmailId(AdminEmailIdConst)

	SkippedOnboardingValue := gjson.Get(userData, SkippedOnboardingConst).Str

	if SkippedOnboardingValue == "true" {
		SkippedOnboarding = true
	} else {
		SkippedOnboarding = false
	}

	ssoSetup = false

	ssoConfig, err := impl.ssoLoginService.GetAll()
	if err == nil && len(ssoConfig) > 0 {
		ssoSetup = true
	}

	return clusters, users, k8sServerVersion, hostURL, ssoSetup, HelmAppAccessCount, ChartStoreVisitCount, SkippedOnboarding, HelmAppUpdateCounter, helmChartSuccessfulDeploymentCount, ExternalHelmAppClusterCount
}
