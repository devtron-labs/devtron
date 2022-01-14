package history

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/history"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type InstalledAppHistoryService interface {
	CreateInstalledAppHistory(installedAppVersionId int, values string, userId int32, tx *pg.Tx) (historyModel *history.InstalledAppHistory, err error)
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

func (impl InstalledAppHistoryServiceImpl) CreateInstalledAppHistory(installedAppVersionId int, values string, userId int32, tx *pg.Tx) (historyModel *history.InstalledAppHistory, err error) {
	historyModel = &history.InstalledAppHistory{
		InstalledAppVersionId: installedAppVersionId,
		Values:                values,
		DeployedBy:            userId,
		DeployedOn:            time.Now(),
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
	_, err = impl.installedAppHistoryRepository.CreateHistory(historyModel, tx)
	if err != nil {
		impl.logger.Errorw("error in creating history entry for installed app", "err", err, "history", historyModel)
		return nil, err
	}
	return historyModel, nil
}
