package read

import (
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"go.uber.org/zap"
)

type HelmAppReadServiceImpl struct {
	logger         *zap.SugaredLogger
	clusterService cluster.ClusterService
}

type HelmAppReadService interface {
	GetClusterConf(clusterId int) (*gRPC.ClusterConfig, error)
}

func NewHelmAppReadServiceImpl(logger *zap.SugaredLogger,
	clusterService cluster.ClusterService,
) *HelmAppReadServiceImpl {
	return &HelmAppReadServiceImpl{
		clusterService: clusterService,
		logger:         logger,
	}
}

func (impl *HelmAppReadServiceImpl) GetClusterConf(clusterId int) (*gRPC.ClusterConfig, error) {
	cluster, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching cluster detail", "err", err)
		return nil, err
	}
	config := &gRPC.ClusterConfig{
		ApiServerUrl:          cluster.ServerUrl,
		Token:                 cluster.Config[commonBean.BearerToken],
		ClusterId:             int32(cluster.Id),
		ClusterName:           cluster.ClusterName,
		InsecureSkipTLSVerify: cluster.InsecureSkipTLSVerify,
	}
	if cluster.InsecureSkipTLSVerify == false {
		config.KeyData = cluster.Config[commonBean.TlsKey]
		config.CertData = cluster.Config[commonBean.CertData]
		config.CaData = cluster.Config[commonBean.CertificateAuthorityData]
	}
	return config, nil
}
