/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bean

type InstalledAppMin struct {
	// Installed App details
	Id     int
	Active bool
	// Deprecated; currently in use for backward compatibility
	GitOpsRepoName string
	// Deprecated; use deployment_config table instead GitOpsRepoName has been migrated to GitOpsRepoUrl; Make sure to migrate from GitOpsRepoName for future flows
	GitOpsRepoUrl      string
	IsCustomRepository bool
	// Deprecated; use deployment_config table instead
	DeploymentAppType          string
	DeploymentAppDeleteRequest bool
	EnvironmentId              int
	AppId                      int
}

type InstalledAppWithAppDetails struct {
	*InstalledAppMin
	// Extra App details
	AppName         string
	AppOfferingMode string
	TeamId          int
}

type InstalledAppWithEnvDetails struct {
	*InstalledAppWithAppDetails
	// Extra Environment details
	EnvironmentName       string
	EnvironmentIdentifier string
	Namespace             string
	ClusterId             int
}

type InstalledAppDeleteRequest struct {
	InstalledAppId  int
	AppName         string
	AppId           int
	EnvironmentId   int
	AppOfferingMode string
	ClusterId       int
	Namespace       string
}

type InstalledAppWithEnvAndClusterDetails struct {
	*InstalledAppWithEnvDetails
	// Extra Cluster details
	ClusterName string
}

func (i *InstalledAppWithEnvAndClusterDetails) GetInstalledAppMin() *InstalledAppMin {
	if i == nil {
		return nil
	}
	return i.InstalledAppMin
}

func (i *InstalledAppWithEnvAndClusterDetails) GetInstalledAppWithAppDetails() *InstalledAppWithAppDetails {
	if i == nil {
		return nil
	}
	return i.InstalledAppWithAppDetails
}

func (i *InstalledAppWithEnvAndClusterDetails) GetInstalledAppWithEnvDetails() *InstalledAppWithEnvDetails {
	if i == nil {
		return nil
	}
	return i.InstalledAppWithEnvDetails
}

type InstalledAppVersionMin struct {
	// Installed App Version details
	Id                           int
	InstalledAppId               int
	AppStoreApplicationVersionId int
	ValuesYaml                   string
	Active                       bool
	ReferenceValueId             int
	ReferenceValueKind           string
}

type InstalledAppVersionWithAppStoreDetails struct {
	*InstalledAppVersionMin
	// Extra App Store Version details
	AppStoreVersion string
}
