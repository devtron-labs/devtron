package imageScanning

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository"
	"go.uber.org/zap"
)

type ImageScanDeployInfoService interface {
	Save(model *repository.ImageScanDeployInfo) error
	Update(model *repository.ImageScanDeployInfo) error
}

type ImageScanDeployInfoServiceImpl struct {
	logger                        *zap.SugaredLogger
	imageScanDeployInfoRepository repository.ImageScanDeployInfoRepository
}

func NewImageScanDeployInfoService(logger *zap.SugaredLogger,
	imageScanDeployInfoRepository repository.ImageScanDeployInfoRepository) *ImageScanDeployInfoServiceImpl {
	return &ImageScanDeployInfoServiceImpl{
		logger:                        logger,
		imageScanDeployInfoRepository: imageScanDeployInfoRepository,
	}
}
func (impl *ImageScanDeployInfoServiceImpl) Save(model *repository.ImageScanDeployInfo) error {
	return impl.imageScanDeployInfoRepository.Save(model)
}

func (impl *ImageScanDeployInfoServiceImpl) Update(model *repository.ImageScanDeployInfo) error {
	return impl.imageScanDeployInfoRepository.Update(model)
}
