/*
 * Copyright (c) 2020 Devtron Labs
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
 *
 */

package appStore

import (
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreRepository "github.com/devtron-labs/devtron/pkg/appStore/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	repository2 "github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ChartGroupServiceImpl struct {
	chartGroupEntriesRepository     appStoreRepository.ChartGroupEntriesRepository
	chartGroupRepository            appStoreRepository.ChartGroupReposotory
	Logger                          *zap.SugaredLogger
	chartGroupDeploymentRepository  appStoreRepository.ChartGroupDeploymentRepository
	installedAppRepository          appStoreRepository.InstalledAppRepository
	appStoreVersionValuesRepository appStoreRepository.AppStoreVersionValuesRepository
	userAuthService                 user.UserAuthService
}

func NewChartGroupServiceImpl(chartGroupEntriesRepository appStoreRepository.ChartGroupEntriesRepository,
	chartGroupRepository appStoreRepository.ChartGroupReposotory,
	Logger *zap.SugaredLogger, chartGroupDeploymentRepository appStoreRepository.ChartGroupDeploymentRepository,
	installedAppRepository appStoreRepository.InstalledAppRepository, appStoreVersionValuesRepository appStoreRepository.AppStoreVersionValuesRepository, userAuthService user.UserAuthService) *ChartGroupServiceImpl {
	return &ChartGroupServiceImpl{
		chartGroupEntriesRepository:     chartGroupEntriesRepository,
		chartGroupRepository:            chartGroupRepository,
		Logger:                          Logger,
		chartGroupDeploymentRepository:  chartGroupDeploymentRepository,
		installedAppRepository:          installedAppRepository,
		appStoreVersionValuesRepository: appStoreVersionValuesRepository,
		userAuthService:                 userAuthService,
	}
}

type ChartGroupService interface {
	CreateChartGroup(req *ChartGroupBean) (*ChartGroupBean, error)
	UpdateChartGroup(req *ChartGroupBean) (*ChartGroupBean, error)
	SaveChartGroupEntries(req *ChartGroupBean) (*ChartGroupBean, error)
	GetChartGroupWithChartMetaData(chartGroupId int) (*ChartGroupBean, error)
	ChartGroupList(max int) (*ChartGroupList, error)
	GetChartGroupWithInstallationDetail(chartGroupId int) (*ChartGroupBean, error)
	ChartGroupListMin(max int) ([]*ChartGroupBean, error)
	DeleteChartGroup(req *ChartGroupBean) error
}

type ChartGroupList struct {
	Groups []*ChartGroupBean `json:"groups,omitempty"`
}
type ChartGroupBean struct {
	Name               string                 `json:"name,omitempty" validate:"name-component,max=200"`
	Description        string                 `json:"description,omitempty"`
	Id                 int                    `json:"id,omitempty"`
	ChartGroupEntries  []*ChartGroupEntryBean `json:"chartGroupEntries,omitempty"`
	InstalledChartData []*InstalledChartData  `json:"installedChartData,omitempty"`
	UserId             int32                  `json:"-"`
}

type ChartGroupEntryBean struct {
	Id                           int            `json:"id,omitempty"`
	AppStoreValuesVersionId      int            `json:"appStoreValuesVersionId,omitempty"` //AppStoreVersionValuesId
	AppStoreValuesVersionName    string         `json:"appStoreValuesVersionName,omitempty"`
	AppStoreValuesChartVersion   string         `json:"appStoreValuesChartVersion,omitempty"`   //chart version corresponding to values
	AppStoreApplicationVersionId int            `json:"appStoreApplicationVersionId,omitempty"` //AppStoreApplicationVersionId
	ChartMetaData                *ChartMetaData `json:"chartMetaData,omitempty"`
	ReferenceType                string         `json:"referenceType, omitempty"`
}

type ChartMetaData struct {
	ChartName                  string `json:"chartName,omitempty"`
	ChartRepoName              string `json:"chartRepoName,omitempty"`
	Icon                       string `json:"icon,omitempty"`
	AppStoreId                 int    `json:"appStoreId"`
	AppStoreApplicationVersion string `json:"appStoreApplicationVersion"`
	EnvironmentName            string `json:"environmentName,omitempty"`
	EnvironmentId              int    `json:"environmentId,omitempty"` //FIXME REMOVE THIS ATTRIBUTE AFTER REMOVING ENVORONMENTID FROM GETINSTALLEDAPPCALL
	IsChartRepoActive          bool   `json:"isChartRepoActive"`
}

type InstalledChartData struct {
	InstallationTime time.Time         `json:"installationTime,omitempty"`
	InstalledCharts  []*InstalledChart `json:"installedCharts,omitempty"`
}

type InstalledChart struct {
	ChartMetaData
	InstalledAppId int `json:"installedAppId,omitempty"`
}

func (impl *ChartGroupServiceImpl) CreateChartGroup(req *ChartGroupBean) (*ChartGroupBean, error) {
	impl.Logger.Debugw("chart group create request", "req", req)
	chartGrouModel := &appStoreRepository.ChartGroup{
		Name:        req.Name,
		Description: req.Description,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: req.UserId,
			UpdatedOn: time.Now(),
			UpdatedBy: req.UserId,
		},
	}
	group, err := impl.chartGroupRepository.Save(chartGrouModel)
	if err != nil {
		impl.Logger.Errorw("error in creating chart group", "req", chartGrouModel, "err", err)
		return nil, err
	}
	req.Id = group.Id
	return req, nil
}

func (impl *ChartGroupServiceImpl) UpdateChartGroup(req *ChartGroupBean) (*ChartGroupBean, error) {
	impl.Logger.Debugw("chart group update request", "req", req)
	chartGrouModel := &appStoreRepository.ChartGroup{
		Name:        req.Name,
		Description: req.Description,
		Id:          req.Id,
		AuditLog: sql.AuditLog{
			UpdatedOn: time.Now(),
			UpdatedBy: req.UserId,
		},
	}
	group, err := impl.chartGroupRepository.Update(chartGrouModel)

	if err != nil {
		impl.Logger.Errorw("error in update chart group", "req", chartGrouModel, "err", err)
		return nil, err
	}
	req.Id = group.Id
	return req, nil
}

func (impl *ChartGroupServiceImpl) SaveChartGroupEntries(req *ChartGroupBean) (*ChartGroupBean, error) {
	group, err := impl.chartGroupRepository.FindByIdWithEntries(req.Id)
	if err != nil {
		impl.Logger.Errorw("error in fetching chart group", "id", req.Id, "err", err)
		return nil, err
	}
	var newEntries []*ChartGroupEntryBean
	oldEntriesMap := make(map[int]*ChartGroupEntryBean)
	for _, entryBean := range req.ChartGroupEntries {
		if entryBean.Id != 0 {
			oldEntriesMap[entryBean.Id] = entryBean
			//update
		} else {
			//create
			newEntries = append(newEntries, entryBean)
		}
	}
	var updateEntries []*appStoreRepository.ChartGroupEntry
	for _, existingEntry := range group.ChartGroupEntries {
		if entry, ok := oldEntriesMap[existingEntry.Id]; ok {
			//update
			existingEntry.AppStoreApplicationVersionId = entry.AppStoreApplicationVersionId
			existingEntry.AppStoreValuesVersionId = entry.AppStoreValuesVersionId
		} else {
			//delete
			existingEntry.Deleted = true
		}
		existingEntry.UpdatedBy = req.UserId
		existingEntry.UpdatedOn = time.Now()
		updateEntries = append(updateEntries, existingEntry)
	}

	var createEntries []*appStoreRepository.ChartGroupEntry
	for _, entryBean := range newEntries {
		entry := &appStoreRepository.ChartGroupEntry{
			AppStoreValuesVersionId:      entryBean.AppStoreValuesVersionId,
			AppStoreApplicationVersionId: entryBean.AppStoreApplicationVersionId,
			ChartGroupId:                 group.Id,
			Deleted:                      false,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: req.UserId,
				UpdatedOn: time.Now(),
				UpdatedBy: req.UserId,
			},
		}
		createEntries = append(createEntries, entry)
	}
	finalEntries, err := impl.chartGroupEntriesRepository.SaveAndUpdateInTransaction(createEntries, updateEntries)
	if err != nil {
		impl.Logger.Errorw("error in adding entries", "err", err)
		return nil, err
	}
	impl.Logger.Debugw("all entries,", "entry", finalEntries)
	return impl.GetChartGroupWithChartMetaData(req.Id)
}

func (impl *ChartGroupServiceImpl) GetChartGroupWithChartMetaData(chartGroupId int) (*ChartGroupBean, error) {
	chartGroup, err := impl.chartGroupRepository.FindById(chartGroupId)
	if err != nil {
		return nil, err
	}
	chartGroupEntries, err := impl.chartGroupEntriesRepository.FindEntriesWithChartMetaByChartGroupId([]int{chartGroupId})
	if err != nil {
		return nil, err
	}
	chartGroupRes := &ChartGroupBean{
		Name:        chartGroup.Name,
		Description: chartGroup.Description,
		Id:          chartGroup.Id,
	}
	for _, chartGroupEntry := range chartGroupEntries {
		entry := impl.charterEntryAdopter(chartGroupEntry)
		chartGroupRes.ChartGroupEntries = append(chartGroupRes.ChartGroupEntries, entry)
	}
	return chartGroupRes, err
}

func (impl *ChartGroupServiceImpl) charterEntryAdopter(chartGroupEntry *appStoreRepository.ChartGroupEntry) *ChartGroupEntryBean {

	var referenceType string
	var valueVersionName string
	var appStoreValuesChartVersion string
	if chartGroupEntry.AppStoreValuesVersionId == 0 {
		referenceType = appStoreBean.REFERENCE_TYPE_DEFAULT
		appStoreValuesChartVersion = chartGroupEntry.AppStoreApplicationVersion.Version
	} else {
		referenceType = appStoreBean.REFERENCE_TYPE_TEMPLATE
		valueVersionName = chartGroupEntry.AppStoreValuesVersion.Name
		//FIXME: orm join not working.  to quick fix it
		valuesVersion, err := impl.appStoreVersionValuesRepository.FindById(chartGroupEntry.AppStoreValuesVersionId)
		if err != nil {
			return nil
		} else {
			appStoreValuesChartVersion = valuesVersion.AppStoreApplicationVersion.Version
		}

		//appStoreValuesChartVersion = chartGroupEntry.AppStoreValuesVersion.AppStoreApplicationVersion.Version
	}
	entry := &ChartGroupEntryBean{
		Id:                           chartGroupEntry.Id,
		AppStoreValuesVersionId:      chartGroupEntry.AppStoreValuesVersionId,
		AppStoreApplicationVersionId: chartGroupEntry.AppStoreApplicationVersionId,
		ReferenceType:                referenceType,
		AppStoreValuesVersionName:    valueVersionName,
		AppStoreValuesChartVersion:   appStoreValuesChartVersion,
		ChartMetaData: &ChartMetaData{
			ChartName:                  chartGroupEntry.AppStoreApplicationVersion.Name,
			ChartRepoName:              chartGroupEntry.AppStoreApplicationVersion.AppStore.ChartRepo.Name,
			Icon:                       chartGroupEntry.AppStoreApplicationVersion.Icon,
			AppStoreId:                 chartGroupEntry.AppStoreApplicationVersion.AppStoreId,
			AppStoreApplicationVersion: chartGroupEntry.AppStoreApplicationVersion.Version,
			IsChartRepoActive:          chartGroupEntry.AppStoreApplicationVersion.AppStore.ChartRepo.Active,
		},
	}
	return entry
}

func (impl *ChartGroupServiceImpl) ChartGroupList(max int) (*ChartGroupList, error) {
	groups, err := impl.chartGroupRepository.GetAll(max)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return nil, nil
	}
	var groupIds []int
	groupMap := make(map[int]*ChartGroupBean)
	for _, group := range groups {
		chartGroupRes := &ChartGroupBean{
			Name:        group.Name,
			Description: group.Description,
			Id:          group.Id,
		}
		groupMap[group.Id] = chartGroupRes
		groupIds = append(groupIds, group.Id)
	}
	groupEntries, err := impl.chartGroupEntriesRepository.FindEntriesWithChartMetaByChartGroupId(groupIds)
	if err != nil {
		return nil, err
	}
	for _, groupentry := range groupEntries {
		entry := impl.charterEntryAdopter(groupentry)
		entries := groupMap[groupentry.ChartGroupId].ChartGroupEntries
		entries = append(entries, entry)
		groupMap[groupentry.ChartGroupId].ChartGroupEntries = entries
	}
	var chartGroups []*ChartGroupBean
	for _, v := range groupMap {
		chartGroups = append(chartGroups, v)
	}
	if len(chartGroups) == 0 {
		chartGroups = make([]*ChartGroupBean, 0)
	}
	return &ChartGroupList{Groups: chartGroups}, nil
}

func (impl *ChartGroupServiceImpl) GetChartGroupWithInstallationDetail(chartGroupId int) (*ChartGroupBean, error) {
	chartGroupBean, err := impl.GetChartGroupWithChartMetaData(chartGroupId)
	if err != nil {
		return nil, err
	}
	deployments, err := impl.chartGroupDeploymentRepository.FindByChartGroupId(chartGroupId)
	if err != nil {
		impl.Logger.Errorw("error in finding deployment", "chartGroupId", chartGroupId, "err", err)
		return nil, err
	}
	groupDeploymentMap := make(map[string][]*appStoreRepository.ChartGroupDeployment)
	for _, deployment := range deployments {
		groupDeploymentMap[deployment.GroupInstallationId] = append(groupDeploymentMap[deployment.GroupInstallationId], deployment)
	}
	for _, groupDeployments := range groupDeploymentMap {
		installedChartData := &InstalledChartData{}
		//installedChartData.InstallationTime
		for _, deployment := range groupDeployments {
			installedChartData.InstallationTime = deployment.CreatedOn
			versions, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppIdMeta(deployment.InstalledAppId)
			if err != nil {
				return nil, err
			}
			version := versions[0]
			installedChart := &InstalledChart{
				ChartMetaData: ChartMetaData{
					ChartName:         version.InstalledApp.App.AppName,
					ChartRepoName:     version.AppStoreApplicationVersion.AppStore.ChartRepo.Name,
					Icon:              version.AppStoreApplicationVersion.Icon,
					AppStoreId:        version.AppStoreApplicationVersion.AppStoreId,
					EnvironmentName:   version.InstalledApp.Environment.Name,
					EnvironmentId:     version.InstalledApp.EnvironmentId,
					IsChartRepoActive: version.AppStoreApplicationVersion.AppStore.ChartRepo.Active,
				},
				InstalledAppId: version.InstalledAppId,
			}
			installedChartData.InstalledCharts = append(installedChartData.InstalledCharts, installedChart)
		}

		chartGroupBean.InstalledChartData = append(chartGroupBean.InstalledChartData, installedChartData)
	}
	/*	for _, deployment := range deployments {
		versions, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppIdMeta(deployment.InstalledAppId)
		if err != nil {
			return nil, err
		}
		version := versions[0]
		installedChartData := &InstalledChartData{
			InstallationTime: version.CreatedOn,
			InstalledCharts: []*InstalledChart{&InstalledChart{
				ChartMetaData: ChartMetaData{
					ChartName:       version.InstalledApp.App.AppName,
					ChartRepoName:   version.AppStoreApplicationVersion.AppStore.ChartRepo.Name,
					Icon:            version.AppStoreApplicationVersion.Icon,
					AppStoreId:      version.AppStoreApplicationVersion.AppStoreId,
					EnvironmentName: version.InstalledApp.Environment.Name,
					EnvironmentId:   version.InstalledApp.EnvironmentId,
				},
				InstalledAppId: version.InstalledAppId,
			}},
		}
		chartGroupBean.InstalledChartData = append(chartGroupBean.InstalledChartData, installedChartData)
	}*/
	return chartGroupBean, nil
}

func (impl *ChartGroupServiceImpl) ChartGroupListMin(max int) ([]*ChartGroupBean, error) {
	var chartGroupList []*ChartGroupBean
	groups, err := impl.chartGroupRepository.GetAll(max)
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		chartGroupRes := &ChartGroupBean{
			Name: group.Name,
			Id:   group.Id,
		}
		chartGroupList = append(chartGroupList, chartGroupRes)
	}
	if len(chartGroupList) == 0 {
		chartGroupList = make([]*ChartGroupBean, 0)
	}
	return chartGroupList, nil
}

func (impl *ChartGroupServiceImpl) DeleteChartGroup(req *ChartGroupBean) error {
	//finding existing
	existingChartGroup, err := impl.chartGroupRepository.FindById(req.Id)
	if err != nil {
		impl.Logger.Errorw("No matching entry found for delete.", "err", err, "id", req.Id)
		return err
	}
	//finding chart mappings by group id
	chartGroupMappings, err := impl.chartGroupEntriesRepository.FindEntriesWithChartMetaByChartGroupId([]int{req.Id})
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in getting chart group entries, DeleteChartGroup", "err", err, "chartGroupId", req.Id)
		return err
	}
	var chartGroupMappingIds []int
	for _, chartGroupMapping := range chartGroupMappings {
		chartGroupMappingIds = append(chartGroupMappingIds, chartGroupMapping.Id)
	}
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.Logger.Errorw("error in establishing connection", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	//deleting chart mapping in group
	if len(chartGroupMappingIds) > 0 {
		_, err = impl.chartGroupEntriesRepository.MarkChartGroupEntriesDeleted(chartGroupMappingIds, tx)
		if err != nil {
			impl.Logger.Errorw("error in deleting chart group mappings", "err", err)
			return err
		}
	}
	//deleting chart group
	err = impl.chartGroupRepository.MarkChartGroupDeleted(existingChartGroup.Id, tx)
	if err != nil {
		impl.Logger.Errorw("error in deleting chart group", "err", err, "chartGroupId", existingChartGroup.Id)
		return err
	}
	//deleting auth roles entries for this chart group
	err = impl.userAuthService.DeleteRoles(repository2.CHART_GROUP_TYPE, req.Name, tx, "")
	if err != nil {
		impl.Logger.Errorw("error in deleting auth roles", "err", err)
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
