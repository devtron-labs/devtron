package service

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/go-pg/pg"
	"regexp"
)

func (impl *InstalledAppServiceImpl) FetchChartNotes(installedAppId int, envId int, token string, checkNotesAuth func(token string, appName string, envId int) bool) (string, error) {
	//check notes.txt in db
	installedApp, err := impl.installedAppRepository.FetchNotes(installedAppId)
	if err != nil && err != pg.ErrNoRows {
		return "", err
	}
	installedAppVerison, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppIdAndEnvId(installedAppId, envId)
	if err != nil {
		if err == pg.ErrNoRows {
			return "", fmt.Errorf("values are outdated. please fetch the latest version and try again")
		}
		impl.logger.Errorw("error fetching installed  app version in installed app service", "err", err)
		return "", err
	}
	chartVersion := installedAppVerison.AppStoreApplicationVersion.Version
	if err != nil {
		impl.logger.Errorw("error fetching chart  version in installed app service", "err", err)
		return "", err
	}
	re := regexp.MustCompile(`CHART VERSION: ([0-9]+\.[0-9]+\.[0-9]+)`)
	newStr := re.ReplaceAllString(installedApp.Notes, "CHART VERSION: "+chartVersion)
	installedApp.Notes = newStr
	appName := installedApp.App.AppName
	if err != nil {
		impl.logger.Errorw("error fetching notes from db", "err", err)
		return "", err
	}
	isValidAuth := checkNotesAuth(token, appName, envId)
	if !isValidAuth {
		impl.logger.Errorw("unauthorized user", "isValidAuth", isValidAuth)
		return "", fmt.Errorf("unauthorized user")
	}
	//if notes is not present in db then below call will happen
	if installedApp.Notes == "" {
		notes, _, err := impl.findNotesForArgoApplication(installedAppId, envId)
		if err != nil {
			impl.logger.Errorw("error fetching notes", "err", err)
			return "", err
		}
		if notes == "" {
			impl.logger.Errorw("error fetching notes", "err", err)
		}
		return notes, err
	}

	return installedApp.Notes, nil
}

func (impl *InstalledAppServiceImpl) findNotesForArgoApplication(installedAppId, envId int) (string, string, error) {
	installedAppVerison, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppIdAndEnvId(installedAppId, envId)
	if err != nil {
		impl.logger.Errorw("error fetching installed  app version in installed app service", "err", err)
		return "", "", err
	}
	var notes string
	appName := installedAppVerison.InstalledApp.App.AppName

	if util.IsAcdApp(installedAppVerison.InstalledApp.DeploymentAppType) {
		appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installedAppVerison.AppStoreApplicationVersion.Id)
		if err != nil {
			impl.logger.Errorw("error fetching app store app version in installed app service", "err", err)
			return notes, appName, err
		}
		k8sServerVersion, err := impl.K8sUtil.GetKubeVersion()
		if err != nil {
			impl.logger.Errorw("exception caught in getting k8sServerVersion", "err", err)
			return notes, appName, err
		}

		installReleaseRequest := &client.InstallReleaseRequest{
			ChartName:    appStoreAppVersion.Name,
			ChartVersion: appStoreAppVersion.Version,
			ValuesYaml:   installedAppVerison.ValuesYaml,
			K8SVersion:   k8sServerVersion.String(),
			ChartRepository: &client.ChartRepository{
				Name:     appStoreAppVersion.AppStore.ChartRepo.Name,
				Url:      appStoreAppVersion.AppStore.ChartRepo.Url,
				Username: appStoreAppVersion.AppStore.ChartRepo.UserName,
				Password: appStoreAppVersion.AppStore.ChartRepo.Password,
			},
			ReleaseIdentifier: &client.ReleaseIdentifier{
				ReleaseNamespace: installedAppVerison.InstalledApp.Environment.Namespace,
				ReleaseName:      installedAppVerison.InstalledApp.App.AppName,
			},
		}

		clusterId := installedAppVerison.InstalledApp.Environment.ClusterId
		config, err := impl.helmAppService.GetClusterConf(clusterId)
		if err != nil {
			impl.logger.Errorw("error in fetching cluster detail", "clusterId", clusterId, "err", err)
			return "", appName, err
		}
		installReleaseRequest.ReleaseIdentifier.ClusterConfig = config

		notes, err = impl.helmAppService.GetNotes(context.Background(), installReleaseRequest)
		if err != nil {
			impl.logger.Errorw("error in fetching notes", "err", err)
			return notes, appName, err
		}
		_, err = impl.updateNotesForInstalledApp(installedAppId, notes)
		if err != nil {
			impl.logger.Errorw("error in updating notes in db ", "err", err)
			return notes, appName, err
		}
	}

	return notes, appName, nil
}

// updateNotesForInstalledApp will update the notes in repository.InstalledApps table
func (impl *InstalledAppServiceImpl) updateNotesForInstalledApp(installAppId int, notes string) (bool, error) {
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return false, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	installedApp, err := impl.installedAppRepository.GetInstalledApp(installAppId)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return false, err
	}
	installedApp.Notes = notes
	_, err = impl.installedAppRepository.UpdateInstalledApp(installedApp, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error while commit db transaction to db", "error", err)
		return false, err
	}
	return true, nil
}
