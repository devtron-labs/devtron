package appstore

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/history"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type InstalledAppHistoryService interface {
	CreateInstalledAppHistory(installAppVersionReq *InstallAppVersionDTO, tx *pg.Tx) (historyModel *history.InstalledAppHistory, err error)
}

type InstalledAppHistoryServiceImpl struct {
	logger                        *zap.SugaredLogger
	installedAppHistoryRepository history.InstalledAppHistoryRepository
}

func NewInstalledAppHistoryServiceImpl(logger *zap.SugaredLogger, installedAppHistoryRepository history.InstalledAppHistoryRepository) *InstalledAppHistoryServiceImpl {
	return &InstalledAppHistoryServiceImpl{
		logger:                        logger,
		installedAppHistoryRepository: installedAppHistoryRepository,
	}
}

func (impl InstalledAppHistoryServiceImpl) CreateInstalledAppHistory(installAppVersionReq *InstallAppVersionDTO, tx *pg.Tx) (historyModel *history.InstalledAppHistory, err error) {
	historyModel = &history.InstalledAppHistory{
		InstalledAppVersionId: installAppVersionReq.InstalledAppVersionId,
		Values:                installAppVersionReq.ValuesOverrideYaml,
		DeployedBy:            installAppVersionReq.UserId,
		DeployedOn:            time.Now(),
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: installAppVersionReq.UserId,
			UpdatedOn: time.Now(),
			UpdatedBy: installAppVersionReq.UserId,
		},
	}
	_, err = impl.installedAppHistoryRepository.CreateHistory(historyModel, tx)
	if err != nil {
		impl.logger.Errorw("error in creating history entry for installed app", "err", err, "history", historyModel)
		return nil, err
	}
	return historyModel, nil
}
