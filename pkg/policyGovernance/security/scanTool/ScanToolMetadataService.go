package scanTool

import (
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/scanTool/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ScanToolMetadataService interface {
	MarkToolAsActive(toolName, version string, tx *pg.Tx) error
	MarkOtherToolsInActive(toolName string, tx *pg.Tx, version string) error
	ScanToolMetadataService_ent
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
