package imageScanning

import (
	"github.com/devtron-labs/devtron/pkg/policyGoverance/security/imageScanning/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ScanToolMetadataService interface {
	MarkToolAsActive(toolName, version string, tx *pg.Tx) error
	MarkOtherToolsInActive(toolName string, tx *pg.Tx, version string) error
}

type ScanToolMetadataServiceImpl struct {
	logger                     *zap.SugaredLogger
	scanToolMetadataRepository repository.ScanToolMetadataRepository
}

func NewScanToolMetadataServiceImpl(logger *zap.SugaredLogger,
	scanToolMetadataRepository repository.ScanToolMetadataRepository) *ScanToolMetadataServiceImpl {
	return &ScanToolMetadataServiceImpl{
		logger:                     logger,
		scanToolMetadataRepository: scanToolMetadataRepository,
	}
}
func (impl *ScanToolMetadataServiceImpl) MarkToolAsActive(toolName, version string, tx *pg.Tx) error {
	return impl.scanToolMetadataRepository.MarkToolAsActive(toolName, version, tx)
}

func (impl *ScanToolMetadataServiceImpl) MarkOtherToolsInActive(toolName string, tx *pg.Tx, version string) error {
	return impl.scanToolMetadataRepository.MarkOtherToolsInActive(toolName, tx, version)
}
