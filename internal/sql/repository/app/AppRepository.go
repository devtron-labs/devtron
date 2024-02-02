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

package app

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type App struct {
	tableName       struct{}       `sql:"app" pg:",discard_unknown_columns"`
	Id              int            `sql:"id,pk"`
	AppName         string         `sql:"app_name,notnull"` // same as app name
	DisplayName     string         `sql:"display_name"`
	Active          bool           `sql:"active, notnull"`
	TeamId          int            `sql:"team_id"`
	AppType         helper.AppType `sql:"app_type, notnull"`
	AppOfferingMode string         `sql:"app_offering_mode,notnull"`
	Description     string         `sql:"description"`
	Team            team.Team
	sql.AuditLog
}

type AppWithExtraQueryFields struct {
	App
	TotalCount int
}
type AppRepository interface {
	Save(pipelineGroup *App) error
	SaveWithTxn(pipelineGroup *App, tx *pg.Tx) error
	Update(app *App) error
	UpdateWithTxn(app *App, tx *pg.Tx) error
	SetDescription(id int, description string, userId int32) error
	FindActiveByName(appName string) (pipelineGroup *App, err error)
	FindJobByDisplayName(appName string) (pipelineGroup *App, err error)
	FindActiveListByName(appName string) ([]*App, error)
	FindById(id int) (pipelineGroup *App, err error)
	FindAppAndTeamByAppId(id int) (*App, error)
	FindActiveById(id int) (pipelineGroup *App, err error)
	FindAppsByTeamId(teamId int) ([]*App, error)
	FindAppsByTeamIds(teamId []int, appType string) ([]App, error)
	FindAppsByTeamName(teamName string) ([]App, error)
	FindAll() ([]*App, error)
	FindAppsByEnvironmentId(environmentId int) ([]App, error)
	FindAllActiveAppsWithTeam(appType helper.AppType) ([]*App, error)
	FindAllActiveAppsWithTeamWithTeamId(teamID int, appType helper.AppType) ([]*App, error)
	CheckAppExists(appNames []string) ([]*App, error)

	FindByIds(ids []*int) ([]*App, error)
	FetchAppsByFilterV2(appNameIncludes string, appNameExcludes string, environmentId int) ([]*App, error)
	FindAppAndProjectByAppId(appId int) (*App, error)
	FindAppAndProjectByAppName(appName string) (*App, error)
	GetConnection() *pg.DB
	FindAllMatchesByAppName(appName string, appType helper.AppType) ([]*App, error)
	FindIdsByTeamIdsAndTeamNames(teamIds []int, teamNames []string) ([]int, error)
	FindIdsByNames(appNames []string) ([]int, error)
	FindByNames(appNames []string) ([]*App, error)
	FetchAllActiveInstalledAppsWithAppIdAndName() ([]*App, error)
	FetchAllActiveDevtronAppsWithAppIdAndName() ([]*App, error)
	FindEnvironmentIdForInstalledApp(appId int) (int, error)
	FetchAppIdsWithFilter(jobListingFilter helper.AppListingFilter) ([]int, error)
	FindAllActiveAppsWithTeamByAppNameMatch(appNameMatch string, appType helper.AppType) ([]*App, error)
	FindAppAndProjectByIdsIn(ids []int) ([]*App, error)
	FindAppAndProjectByIdsOrderByTeam(ids []int) ([]*App, error)
	FetchAppIdsByDisplayNamesForJobs(names []string) (map[int]string, []int, error)
	GetActiveCiCdAppsCount(excludeAppIds []int) (int, error)
	FindAppsWithFilter(appNameLike, sortOrder string, limit, offset int, excludeAppIds []int) ([]AppWithExtraQueryFields, error)
}

const DevtronApp = "DevtronApp"
const DevtronChart = "DevtronChart"
const ExternalApp = "ExternalApp"

type AppRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewAppRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *AppRepositoryImpl {
	return &AppRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (repo AppRepositoryImpl) GetConnection() *pg.DB {
	return repo.dbConnection
}

func (repo AppRepositoryImpl) Save(pipelineGroup *App) error {
	err := repo.dbConnection.Insert(pipelineGroup)
	return err
}

func (repo AppRepositoryImpl) SaveWithTxn(pipelineGroup *App, tx *pg.Tx) error {
	err := tx.Insert(pipelineGroup)
	return err
}

func (repo AppRepositoryImpl) Update(app *App) error {
	_, err := repo.dbConnection.Model(app).WherePK().UpdateNotNull()
	return err
}

func (repo AppRepositoryImpl) UpdateWithTxn(app *App, tx *pg.Tx) error {
	err := tx.Update(app)
	return err
}

func (repo AppRepositoryImpl) SetDescription(id int, description string, userId int32) error {
	_, err := repo.dbConnection.Model((*App)(nil)).
		Set("description = ?", description).Set("updated_by = ?", userId).Set("updated_on = ?", time.Now()).
		Where("id = ?", id).Update()
	return err
}

func (repo AppRepositoryImpl) FindActiveByName(appName string) (*App, error) {
	pipelineGroup := &App{}
	err := repo.dbConnection.
		Model(pipelineGroup).
		Where("app_name = ?", appName).
		Where("active = ?", true).
		Order("id DESC").Limit(1).
		Select()
	// there is only single active app will be present in db with a same name.
	return pipelineGroup, err
}
func (repo AppRepositoryImpl) FindJobByDisplayName(appName string) (*App, error) {
	pipelineGroup := &App{}
	err := repo.dbConnection.
		Model(pipelineGroup).
		Where("display_name = ?", appName).
		Where("active = ?", true).
		Where("app_type = ?", helper.Job).
		Order("id DESC").Limit(1).
		Select()
	// there is only single active app will be present in db with a same name.
	return pipelineGroup, err
}

func (repo AppRepositoryImpl) FindActiveListByName(appName string) ([]*App, error) {
	var apps []*App
	err := repo.dbConnection.
		Model(&apps).
		Where("app_name = ?", appName).
		Where("active = ?", true).
		Order("id ASC").
		Select()
	// there is only single active app will be present in db with a same name. but check for concurrency
	return apps, err
}

func (repo AppRepositoryImpl) CheckAppExists(appNames []string) ([]*App, error) {
	var apps []*App
	err := repo.dbConnection.
		Model(&apps).
		Where("app_name in (?)", pg.In(appNames)).
		Where("active = ?", true).
		Select()
	return apps, err
}

func (repo AppRepositoryImpl) FindById(id int) (*App, error) {
	pipelineGroup := &App{}
	err := repo.dbConnection.Model(pipelineGroup).Where("id = ?", id).
		Where("active = ?", true).Select()
	return pipelineGroup, err
}

func (repo AppRepositoryImpl) FindAppAndTeamByAppId(id int) (*App, error) {
	pipelineGroup := &App{}
	err := repo.dbConnection.Model(pipelineGroup).
		Column("Team").
		Where("app.id = ?", id).
		Where("app.active = ?", true).
		Where("team.id = app.team_id").
		Select()

	return pipelineGroup, err
}

func (repo AppRepositoryImpl) FindActiveById(id int) (*App, error) {
	pipelineGroup := &App{}
	err := repo.dbConnection.
		Model(pipelineGroup).
		Where("id = ?", id).
		Where("active = ?", true).
		Select()
	return pipelineGroup, err
}

func (repo AppRepositoryImpl) FindAppsByTeamId(teamId int) ([]*App, error) {
	var apps []*App
	err := repo.dbConnection.Model(&apps).Where("team_id = ?", teamId).
		Where("active = ?", true).Select()
	return apps, err
}

func (repo AppRepositoryImpl) FindAppsByTeamIds(teamId []int, appType string) ([]App, error) {
	onlyDevtronCharts := 0
	if len(appType) > 0 && appType == DevtronChart {
		onlyDevtronCharts = 1
	}
	var apps []App
	err := repo.dbConnection.Model(&apps).Column("app.*", "Team").Where("team_id in (?)", pg.In(teamId)).
		Where("app.active=?", true).Where("app.app_type=?", onlyDevtronCharts).Select()
	return apps, err
}

func (repo AppRepositoryImpl) FindAppsByTeamName(teamName string) ([]App, error) {
	var apps []App
	err := repo.dbConnection.Model(&apps).Column("app.*").
		Join("inner join team t on t.id = app.team_id").
		Where("t.name = ?", teamName).Where("t.active = ?", true).
		Select()
	return apps, err
}

func (repo AppRepositoryImpl) FindAll() ([]*App, error) {
	var apps []*App
	err := repo.dbConnection.Model(&apps).Where("active = ?", true).Where("app_type = ?", 0).Select()
	return apps, err
}

func (repo AppRepositoryImpl) FindAppsByEnvironmentId(environmentId int) ([]App, error) {
	var apps []App
	err := repo.dbConnection.Model(&apps).ColumnExpr("DISTINCT app.*").
		Join("inner join pipeline p on p.app_id=app.id").Where("p.environment_id = ?", environmentId).Where("p.deleted = ?", false).
		Select()
	return apps, err
}

func (repo AppRepositoryImpl) FindAllActiveAppsWithTeam(appType helper.AppType) ([]*App, error) {
	var apps []*App
	err := repo.dbConnection.Model(&apps).Column("Team").
		Where("app.active = ?", true).Where("app.app_type = ?", appType).
		Select()
	return apps, err
}

func (repo AppRepositoryImpl) FindAllActiveAppsWithTeamWithTeamId(teamID int, appType helper.AppType) ([]*App, error) {
	var apps []*App
	err := repo.dbConnection.Model(&apps).Column("Team").
		Where("app.active = ?", true).
		Where("app.app_type = ?", appType).
		Where("app.team_id = ?", teamID).
		Select()
	return apps, err
}

func (repo AppRepositoryImpl) FindAllActiveAppsWithTeamByAppNameMatch(appNameMatch string, appType helper.AppType) ([]*App, error) {
	var apps []*App
	appNameLikeQuery := "app.app_name like '%" + appNameMatch + "%'"
	err := repo.dbConnection.Model(&apps).Column("Team").
		Where("app.active = ?", true).Where("app.app_type = ?", appType).Where(appNameLikeQuery).
		Select()
	return apps, err
}

func (repo AppRepositoryImpl) FindByIds(ids []*int) ([]*App, error) {
	var apps []*App
	err := repo.dbConnection.Model(&apps).Where("active = ?", true).Where("id in (?)", pg.In(ids)).Select()
	return apps, err
}

func (repo AppRepositoryImpl) FetchAppsByFilterV2(appNameIncludes string, appNameExcludes string, environmentId int) ([]*App, error) {
	var apps []*App
	var err error
	if environmentId > 0 && len(appNameExcludes) > 0 {
		err = repo.dbConnection.Model(&apps).ColumnExpr("DISTINCT app.*").
			Join("inner join pipeline p on p.app_id=app.id").
			Where("app.app_name like ?", ""+appNameIncludes+"%").Where("app.app_name not like ?", ""+appNameExcludes+"%").
			Where("app.active=?", true).Where("app_type=?", 0).
			Where("p.environment_id = ?", environmentId).Where("p.deleted = ?", false).
			Select()
	} else if environmentId > 0 && appNameExcludes == "" {
		err = repo.dbConnection.Model(&apps).ColumnExpr("DISTINCT app.*").
			Join("inner join pipeline p on p.app_id=app.id").
			Where("app.app_name like ?", ""+appNameIncludes+"%").
			Where("app.active=?", true).Where("app_type=?", 0).
			Where("p.environment_id = ?", environmentId).Where("p.deleted = ?", false).
			Select()
	} else if environmentId == 0 && len(appNameExcludes) > 0 {
		err = repo.dbConnection.Model(&apps).ColumnExpr("DISTINCT app.*").
			Where("app.app_name like ?", ""+appNameIncludes+"%").Where("app.app_name not like ?", ""+appNameExcludes+"%").
			Where("app.active=?", true).Where("app_type=?", 0).
			Select()
	} else if environmentId == 0 && appNameExcludes == "" {
		err = repo.dbConnection.Model(&apps).ColumnExpr("DISTINCT app.*").
			Where("app.app_name like ?", ""+appNameIncludes+"%").
			Where("app.active=?", true).Where("app_type=?", 0).
			Select()
	}
	return apps, err
}

func (repo AppRepositoryImpl) FindAppAndProjectByAppId(appId int) (*App, error) {
	app := &App{}
	err := repo.dbConnection.Model(app).Column("Team").
		Where("app.id = ?", appId).
		Where("app.active=?", true).
		Select()
	return app, err
}

func (repo AppRepositoryImpl) FindAppAndProjectByAppName(appName string) (*App, error) {
	app := &App{}
	err := repo.dbConnection.Model(app).Column("Team").
		Where("app.app_name = ?", appName).
		Where("app.active=?", true).
		Select()
	return app, err
}

func (repo AppRepositoryImpl) FindAllMatchesByAppName(appName string, appType helper.AppType) ([]*App, error) {
	var apps []*App
	var err error
	if appType == helper.Job {
		err = repo.dbConnection.Model(&apps).Where("display_name LIKE ?", "%"+appName+"%").Where("active = ?", true).Where("app_type = ?", appType).Select()
	} else {
		err = repo.dbConnection.Model(&apps).Where("app_name LIKE ?", "%"+appName+"%").Where("active = ?", true).Where("app_type = ?", appType).Select()
	}

	return apps, err
}

func (repo AppRepositoryImpl) FindIdsByTeamIdsAndTeamNames(teamIds []int, teamNames []string) ([]int, error) {
	var ids []int
	var err error
	if len(teamIds) == 0 && len(teamNames) == 0 {
		err = fmt.Errorf("invalid input arguments, no projectIds or projectNames to get apps")
		return nil, err
	}
	if len(teamIds) > 0 && len(teamNames) > 0 {
		query := `select app.id from app inner join team on team.id=app.team_id where team.active=? and app.active=?   
                 and (team.id in (?) or team.name in (?));`
		_, err = repo.dbConnection.Query(&ids, query, true, true, pg.In(teamIds), pg.In(teamNames))
	} else if len(teamIds) > 0 {
		query := "select id from app where team_id in (?) and active=?;"
		_, err = repo.dbConnection.Query(&ids, query, pg.In(teamIds), true)
	} else if len(teamNames) > 0 {
		query := "select app.id from app inner join team on team.id=app.team_id where team.name in (?) and team.active=? and app.active=?;"
		_, err = repo.dbConnection.Query(&ids, query, pg.In(teamNames), true, true)
	}
	if err != nil {
		repo.logger.Errorw("error in getting appIds by teamIds and teamNames", "err", err, "teamIds", teamIds, "teamNames", teamNames)
		return nil, err
	}
	return ids, err
}

func (repo AppRepositoryImpl) FindIdsByNames(appNames []string) ([]int, error) {
	var ids []int
	query := "select id from app where app_name in (?) and active=?;"
	_, err := repo.dbConnection.Query(&ids, query, pg.In(appNames), true)
	if err != nil {
		repo.logger.Errorw("error in getting appIds by names", "err", err, "names", appNames)
		return nil, err
	}
	return ids, err
}

func (repo AppRepositoryImpl) FindByNames(appNames []string) ([]*App, error) {
	var appNamesWithIds []*App
	err := repo.dbConnection.Model(&appNamesWithIds).
		Where("active=true").
		Where("app_name in (?)", pg.In(appNames)).
		Select()
	return appNamesWithIds, err
}

func (repo AppRepositoryImpl) FetchAllActiveInstalledAppsWithAppIdAndName() ([]*App, error) {
	repo.logger.Debug("reached at Fetch All Active Installed Apps With AppId And Name")
	var apps []*App

	err := repo.dbConnection.Model(&apps).
		Column("installed_apps.id", "app.app_name").
		Join("INNER JOIN installed_apps  on app.id = installed_apps.app_id").
		Where("app.active=true").
		Select()
	if err != nil && err != pg.ErrNoRows {
		repo.logger.Errorw("error while fetching installed apps With AppId And Name", "err", err)
		return apps, err
	}
	return apps, nil

}
func (repo AppRepositoryImpl) FetchAllActiveDevtronAppsWithAppIdAndName() ([]*App, error) {
	repo.logger.Debug("reached at Fetch All Active Devtron Apps With AppId And Name:")
	var apps []*App

	err := repo.dbConnection.Model(&apps).
		Column("id", "app_name").
		Where("app_type = ?", 0).
		Where("active", true).
		Select()
	if err != nil && err != pg.ErrNoRows {
		repo.logger.Errorw("error while fetching active Devtron apps With AppId And Name", "err", err)
		return apps, err
	}
	return apps, nil
}

func (repo AppRepositoryImpl) FindEnvironmentIdForInstalledApp(appId int) (int, error) {
	type envIdRes struct {
		envId int `json:"envId"`
	}
	res := envIdRes{}
	query := "select ia.environment_id " +
		"from installed_apps ia where ia.app_id = ?"
	_, err := repo.dbConnection.Query(&res, query, appId)
	return res.envId, err
}
func (repo AppRepositoryImpl) FetchAppIdsWithFilter(jobListingFilter helper.AppListingFilter) ([]int, error) {
	type AppId struct {
		Id int `json:"id"`
	}
	var jobIds []AppId
	whereCondition := " where active = true and app_type = 2 "
	if len(jobListingFilter.Teams) > 0 {
		whereCondition += " and team_id in (" + helper.GetCommaSepratedString(jobListingFilter.Teams) + ")"
	}
	if len(jobListingFilter.AppIds) > 0 {
		whereCondition += " and id in (" + helper.GetCommaSepratedString(jobListingFilter.AppIds) + ")"
	}

	if len(jobListingFilter.AppNameSearch) > 0 {
		whereCondition += " and display_name like '%" + jobListingFilter.AppNameSearch + "%' "
	}
	orderByCondition := " order by display_name "
	if jobListingFilter.SortOrder == "DESC" {
		orderByCondition += string(jobListingFilter.SortOrder)
	}
	query := "select id " + "from app " + whereCondition + orderByCondition

	_, err := repo.dbConnection.Query(&jobIds, query)
	appCounts := make([]int, 0)
	for _, id := range jobIds {
		appCounts = append(appCounts, id.Id)
	}
	return appCounts, err
}

func (repo AppRepositoryImpl) FindAppAndProjectByIdsIn(ids []int) ([]*App, error) {
	var apps []*App
	err := repo.dbConnection.Model(&apps).Column("app.*", "Team").Where("app.active = ?", true).Where("app.id in (?)", pg.In(ids)).Select()
	return apps, err
}
func (repo AppRepositoryImpl) FetchAppIdsByDisplayNamesForJobs(names []string) (map[int]string, []int, error) {
	type App struct {
		Id          int    `json:"id"`
		DisplayName string `json:"display_name"`
	}
	var jobIdName []App
	whereCondition := fmt.Sprintf(" where active = true and app_type = %v ", helper.Job)
	whereCondition += " and display_name in (" + helper.GetCommaSepratedStringWithComma(names) + ");"
	query := "select id, display_name from app " + whereCondition
	_, err := repo.dbConnection.Query(&jobIdName, query)
	appResp := make(map[int]string)
	jobIds := make([]int, 0)
	for _, id := range jobIdName {
		appResp[id.Id] = id.DisplayName
		jobIds = append(jobIds, id.Id)
	}
	return appResp, jobIds, err
}

func (repo AppRepositoryImpl) FindAppAndProjectByIdsOrderByTeam(ids []int) ([]*App, error) {
	var apps []*App
	if len(ids) == 0 {
		return apps, nil
	}
	err := repo.dbConnection.Model(&apps).
		Column("app.*", "Team").
		Where("app.active = ?", true).
		Where("app.id in (?)", pg.In(ids)).
		Order("app.team_id").
		Select()
	return apps, err
}

func (repo AppRepositoryImpl) GetActiveCiCdAppsCount(excludeAppIds []int) (int, error) {
	query := repo.dbConnection.Model(&App{}).
		Where("active=?", true).
		Where("app_type=?", helper.CustomApp)
	if len(excludeAppIds) > 0 {
		query = query.Where("id not in (?)", pg.In(excludeAppIds))
	}
	return query.Count()
}

func (repo AppRepositoryImpl) FindAppsWithFilter(appNameLike, sortOrder string, limit, offset int, excludeAppIds []int) ([]AppWithExtraQueryFields, error) {
	query := "SELECT id, app_name,COUNT(id) OVER() AS total_count " +
		" FROM app " +
		" WHERE active=true "
	if appNameLike != "" {
		query += " AND app_name LIKE '%" + appNameLike + "%' "
	}
	if sortOrder != "" {
		query += fmt.Sprintf(" ORDER BY app_name %s ", sortOrder)
	}
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d ", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d ", offset)
	}
	if len(excludeAppIds) > 0 {
		query += fmt.Sprintf(" AND id NOT IN (%s) ", helper.GetCommaSepratedString(excludeAppIds))
	}

	apps := make([]AppWithExtraQueryFields, 0)
	_, err := repo.dbConnection.Query(&apps, query)
	return apps, err
}
