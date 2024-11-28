package read

import (
	"github.com/devtron-labs/devtron/pkg/policyGoverance/security/imageScanning/repository"
	"go.uber.org/zap"
)

type ImageScanResultReadService interface {
	FindByImageDigests(digest []string) ([]*repository.ImageScanExecutionResult, error)
}
type ImageScanResultReadServiceImpl struct {
	logger                    *zap.SugaredLogger
	ImageScanResultRepository repository.ImageScanResultRepository
}

func NewImageScanResultReadServiceImpl(logger *zap.SugaredLogger,
	ImageScanResultRepository repository.ImageScanResultRepository) *ImageScanResultReadServiceImpl {
	return &ImageScanResultReadServiceImpl{
		logger:                    logger,
		ImageScanResultRepository: ImageScanResultRepository,
	}
}

func (impl *ImageScanResultReadServiceImpl) FindByImageDigests(digest []string) ([]*repository.ImageScanExecutionResult, error) {
	return impl.ImageScanResultRepository.FindByImageDigests(digest)
}
