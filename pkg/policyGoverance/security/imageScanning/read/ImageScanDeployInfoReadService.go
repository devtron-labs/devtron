package read

import (
	"github.com/devtron-labs/devtron/pkg/policyGoverance/security/imageScanning/repository"
	"go.uber.org/zap"
)

type ImageScanDeployInfoReadService interface {
	FetchByAppIdAndEnvId(appId int, envId int, objectType []string) (*repository.ImageScanDeployInfo, error)
}

type ImageScanDeployInfoReadServiceImpl struct {
	logger                        *zap.SugaredLogger
	imageScanDeployInfoRepository repository.ImageScanDeployInfoRepository
}

func NewImageScanDeployInfoReadService(logger *zap.SugaredLogger,
	imageScanDeployInfoRepository repository.ImageScanDeployInfoRepository) *ImageScanDeployInfoReadServiceImpl {
	return &ImageScanDeployInfoReadServiceImpl{
		logger:                        logger,
		imageScanDeployInfoRepository: imageScanDeployInfoRepository,
	}
}

func (impl *ImageScanDeployInfoReadServiceImpl) FetchByAppIdAndEnvId(appId int, envId int, objectType []string) (*repository.ImageScanDeployInfo, error) {
	return impl.imageScanDeployInfoRepository.FetchByAppIdAndEnvId(appId, envId, objectType)
}
