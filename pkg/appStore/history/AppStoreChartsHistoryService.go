package history

import (
	"github.com/devtron-labs/devtron/pkg/appStore/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type AppStoreChartsHistoryService interface {
	CreateAppStoreChartsHistory(installedAppsId int, values string, userId int32, tx *pg.Tx) (historyModel *repository.AppStoreChartsHistory, err error)
}

type AppStoreChartsHistoryServiceImpl struct {
	logger                          *zap.SugaredLogger
	appStoreChartsHistoryRepository repository.AppStoreChartsHistoryRepository
}

func NewAppStoreChartsHistoryServiceImpl(logger *zap.SugaredLogger, appStoreChartsHistoryRepository repository.AppStoreChartsHistoryRepository) *AppStoreChartsHistoryServiceImpl {
	return &AppStoreChartsHistoryServiceImpl{
		logger:                          logger,
		appStoreChartsHistoryRepository: appStoreChartsHistoryRepository,
	}
}

func (impl AppStoreChartsHistoryServiceImpl) CreateAppStoreChartsHistory(installedAppsId int, values string, userId int32, tx *pg.Tx) (historyModel *repository.AppStoreChartsHistory, err error) {
	historyModel = &repository.AppStoreChartsHistory{
		InstalledAppsId: installedAppsId,
		Values:          values,
		DeployedBy:      userId,
		DeployedOn:      time.Now(),
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
			UpdatedOn: time.Now(),
			UpdatedBy: userId,
		},
	}
	if tx != nil {
		_, err = impl.appStoreChartsHistoryRepository.CreateHistoryWithTxn(historyModel, tx)
	} else {
		_, err = impl.appStoreChartsHistoryRepository.CreateHistory(historyModel)
	}
	if err != nil {
		impl.logger.Errorw("error in creating history entry for app store charts", "err", err, "history", historyModel)
		return nil, err
	}
	return historyModel, nil
}
