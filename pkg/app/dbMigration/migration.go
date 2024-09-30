package dbMigration

import (
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	repository2 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"go.uber.org/zap"
	"time"
)

type DbMigration interface {
	FixMultipleAppsForInstalledApp(appNameUniqueIdentifier string) (*appRepository.App, error)
}

type DbMigrationServiceImpl struct {
	logger                 *zap.SugaredLogger
	appRepository          appRepository.AppRepository
	installedAppRepository repository2.InstalledAppRepository
}

func NewDbMigrationServiceImpl(
	logger *zap.SugaredLogger, appRepository appRepository.AppRepository,
	installedAppRepository repository2.InstalledAppRepository,
) *DbMigrationServiceImpl {
	impl := &DbMigrationServiceImpl{
		logger:                 logger,
		appRepository:          appRepository,
		installedAppRepository: installedAppRepository,
	}
	return impl
}

func (impl DbMigrationServiceImpl) FixMultipleAppsForInstalledApp(appNameUniqueIdentifier string) (*appRepository.App, error) {
	installedApp, err := impl.installedAppRepository.GetInstalledAppByAppName(appNameUniqueIdentifier)
	if err != nil {
		impl.logger.Errorw("error in fetching installed app by unique identifier", "appNameUniqueIdentifier", appNameUniqueIdentifier, "err", err)
		return nil, err
	}
	validAppId := installedApp.AppId
	allActiveApps, err := impl.appRepository.FindAllActiveByName(appNameUniqueIdentifier)
	if err != nil {
		impl.logger.Errorw("error in fetching all active apps by name", "appName", appNameUniqueIdentifier, "err", err)
		return nil, err
	}
	var validApp *appRepository.App
	for _, activeApp := range allActiveApps {
		if activeApp.Id != validAppId {
			impl.logger.Info("duplicate entries found for app, marking app inactive ", "appName", appNameUniqueIdentifier)
			activeApp.Active = false
			activeApp.UpdatedOn = time.Now()
			activeApp.UpdatedBy = 1
			err := impl.appRepository.Update(activeApp)
			if err != nil {
				impl.logger.Errorw("error in marking app inactive", "name", activeApp.AppName, "err", err)
				return nil, err
			}
		} else {
			validApp = activeApp
		}
	}
	return validApp, nil
}
