package read

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/repository"
	"go.uber.org/zap"
)

type ImageScanHistoryReadService interface {
	FindByImageAndDigest(imageDigest string, image string) (*repository.ImageScanExecutionHistory, error)
	FindByImage(image string) (*repository.ImageScanExecutionHistory, error)
}

type ImageScanHistoryReadServiceImpl struct {
	logger                     *zap.SugaredLogger
	imageScanHistoryRepository repository.ImageScanHistoryRepository
}

func NewImageScanHistoryReadService(logger *zap.SugaredLogger,
	imageScanHistoryRepository repository.ImageScanHistoryRepository) *ImageScanHistoryReadServiceImpl {
	return &ImageScanHistoryReadServiceImpl{
		logger:                     logger,
		imageScanHistoryRepository: imageScanHistoryRepository,
	}
}

func (impl *ImageScanHistoryReadServiceImpl) FindByImageAndDigest(imageDigest string, image string) (*repository.ImageScanExecutionHistory, error) {
	return impl.imageScanHistoryRepository.FindByImageAndDigest(imageDigest, image)
}

func (impl *ImageScanHistoryReadServiceImpl) FindByImage(image string) (*repository.ImageScanExecutionHistory, error) {
	return impl.imageScanHistoryRepository.FindByImage(image)
}
