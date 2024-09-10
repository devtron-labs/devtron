/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package resource

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/util"
	clientErrors "github.com/devtron-labs/devtron/pkg/errors"
	"github.com/go-pg/pg"
	"net/http"
	"regexp"
)

func (impl *InstalledAppResourceServiceImpl) FetchChartNotes(installedAppId int, envId int, token string, checkNotesAuth func(token string, appName string, envId int) bool) (string, error) {
	//check notes.txt in db
	installedApp, err := impl.installedAppRepository.FetchNotes(installedAppId)
	if err != nil && err != pg.ErrNoRows {
		return "", err
	}
	installedAppVerison, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppIdAndEnvId(installedAppId, envId)
	if err != nil {
		if err == pg.ErrNoRows {
			return "", &util.ApiError{HttpStatusCode: http.StatusBadRequest, Code: "400", UserMessage: "values are outdated. please fetch the latest version and try again", InternalMessage: err.Error()}
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
		notes, err := impl.findNotesForArgoApplication(installedAppId, envId)
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

func (impl *InstalledAppResourceServiceImpl) findNotesForArgoApplication(installedAppId, envId int) (string, error) {
	installedAppVerison, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppIdAndEnvId(installedAppId, envId)
	if err != nil {
		impl.logger.Errorw("error fetching installed  app version in installed app service", "err", err)
		return "", err
	}
	var notes string

	deploymentConfig, err := impl.deploymentConfigurationService.GetConfigForHelmApps(installedAppVerison.InstalledApp.AppId, installedAppVerison.InstalledApp.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in getiting deployment config db object by appId and envId", "appId", installedAppVerison.InstalledApp.AppId, "envId", installedAppVerison.InstalledApp.EnvironmentId, "err", err)
		return "", err
	}

	if util.IsAcdApp(deploymentConfig.DeploymentAppType) {
		appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installedAppVerison.AppStoreApplicationVersion.Id)
		if err != nil {
			impl.logger.Errorw("error fetching app store app version in installed app service", "err", err)
			return notes, err
		}
		k8sServerVersion, err := impl.K8sUtil.GetKubeVersion()
		if err != nil {
			impl.logger.Errorw("exception caught in getting k8sServerVersion", "err", err)
			return notes, err
		}

		//TODO: common function for below logic for install/update/notes
		var IsOCIRepo bool
		var registryCredential *gRPC.RegistryCredential
		var chartRepository *gRPC.ChartRepository
		dockerRegistryId := appStoreAppVersion.AppStore.DockerArtifactStoreId
		if dockerRegistryId != "" {
			ociRegistryConfigs, err := impl.OCIRegistryConfigRepository.FindByDockerRegistryId(dockerRegistryId)
			if err != nil {
				impl.logger.Errorw("error in fetching oci registry config", "err", err)
				return notes, err
			}
			var ociRegistryConfig *repository2.OCIRegistryConfig
			for _, config := range ociRegistryConfigs {
				if config.RepositoryAction == repository2.STORAGE_ACTION_TYPE_PULL || config.RepositoryAction == repository2.STORAGE_ACTION_TYPE_PULL_AND_PUSH {
					ociRegistryConfig = config
					break
				}
			}
			IsOCIRepo = true
			registryCredential = &gRPC.RegistryCredential{
				RegistryUrl:         appStoreAppVersion.AppStore.DockerArtifactStore.RegistryURL,
				Username:            appStoreAppVersion.AppStore.DockerArtifactStore.Username,
				Password:            appStoreAppVersion.AppStore.DockerArtifactStore.Password,
				AwsRegion:           appStoreAppVersion.AppStore.DockerArtifactStore.AWSRegion,
				AccessKey:           appStoreAppVersion.AppStore.DockerArtifactStore.AWSAccessKeyId,
				SecretKey:           appStoreAppVersion.AppStore.DockerArtifactStore.AWSSecretAccessKey,
				RegistryType:        string(appStoreAppVersion.AppStore.DockerArtifactStore.RegistryType),
				RepoName:            appStoreAppVersion.AppStore.Name,
				IsPublic:            ociRegistryConfig.IsPublic,
				Connection:          appStoreAppVersion.AppStore.DockerArtifactStore.Connection,
				RegistryName:        appStoreAppVersion.AppStore.DockerArtifactStoreId,
				RegistryCertificate: appStoreAppVersion.AppStore.DockerArtifactStore.Cert,
			}
		} else {
			chartRepository = &gRPC.ChartRepository{
				Name:                    appStoreAppVersion.AppStore.ChartRepo.Name,
				Url:                     appStoreAppVersion.AppStore.ChartRepo.Url,
				Username:                appStoreAppVersion.AppStore.ChartRepo.UserName,
				Password:                appStoreAppVersion.AppStore.ChartRepo.Password,
				AllowInsecureConnection: appStoreAppVersion.AppStore.ChartRepo.AllowInsecureConnection,
			}
		}

		installReleaseRequest := &gRPC.InstallReleaseRequest{
			ChartName:       appStoreAppVersion.Name,
			ChartVersion:    appStoreAppVersion.Version,
			ValuesYaml:      installedAppVerison.ValuesYaml,
			K8SVersion:      k8sServerVersion.String(),
			ChartRepository: chartRepository,
			ReleaseIdentifier: &gRPC.ReleaseIdentifier{
				ReleaseNamespace: installedAppVerison.InstalledApp.Environment.Namespace,
				ReleaseName:      installedAppVerison.InstalledApp.App.AppName,
			},
			IsOCIRepo:          IsOCIRepo,
			RegistryCredential: registryCredential,
		}

		clusterId := installedAppVerison.InstalledApp.Environment.ClusterId
		config, err := impl.helmAppService.GetClusterConf(clusterId)
		if err != nil {
			impl.logger.Errorw("error in fetching cluster detail", "clusterId", clusterId, "err", err)
			return "", err
		}
		installReleaseRequest.ReleaseIdentifier.ClusterConfig = config

		notes, err = impl.helmAppService.GetNotes(context.Background(), installReleaseRequest)
		if err != nil {
			impl.logger.Errorw("error in fetching notes", "err", err)
			apiError := clientErrors.ConvertToApiError(err)
			if apiError != nil {
				err = apiError
			}
			return notes, err
		}
		_, err = impl.updateNotesForInstalledApp(installedAppId, notes)
		if err != nil {
			impl.logger.Errorw("error in updating notes in db ", "err", err)
			return notes, err
		}
	}

	return notes, nil
}

// updateNotesForInstalledApp will update the notes in repository.InstalledApps table
func (impl *InstalledAppResourceServiceImpl) updateNotesForInstalledApp(installAppId int, notes string) (bool, error) {
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
