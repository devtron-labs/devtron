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

/*
	@description: user authentication and authorization
*/
package repository

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strings"
)

const (
	PROJECT_TYPE     = "team"
	ENV_TYPE         = "environment"
	APP_TYPE         = "app"
	CHART_GROUP_TYPE = "chart-group"
)

type UserAuthRepository interface {
	CreateRole(userModel *RoleModel, tx *pg.Tx) (*RoleModel, error)
	GetRoleById(id int) (*RoleModel, error)
	GetRolesByUserId(userId int32) ([]RoleModel, error)
	GetRolesByGroupId(userId int32) ([]*RoleModel, error)
	GetAllRole() ([]RoleModel, error)
	GetRolesByActionAndAccessType(action string, accessType string) ([]RoleModel, error)
	GetRoleByFilter(entity string, team string, app string, env string, act string, accessType string) (RoleModel, error)
	CreateUserRoleMapping(userRoleModel *UserRoleModel, tx *pg.Tx) (*UserRoleModel, error)
	GetUserRoleMappingByUserId(userId int32) ([]*UserRoleModel, error)
	DeleteUserRoleMapping(userRoleModel *UserRoleModel, tx *pg.Tx) (bool, error)

	CreateDefaultPolicies(team string, entityName string, env string, tx *pg.Tx) (bool, error)
	CreateDefaultHelmPolicies(team string, entityName string, env string, tx *pg.Tx) (bool, error)
	CreateDefaultPoliciesForGlobalEntity(entity string, entityName string, action string, tx *pg.Tx) (bool, error)
	CreateRoleForSuperAdminIfNotExists(tx *pg.Tx) (bool, error)
	SyncOrchestratorToCasbin(team string, entityName string, env string, tx *pg.Tx) (bool, error)
	UpdateTriggerPolicyForTerminalAccess() error
	GetRolesForEnvironment(envName string) ([]*RoleModel, error)
	GetRolesForProject(teamName string) ([]*RoleModel, error)
	GetRolesForApp(appName string) ([]*RoleModel, error)
	GetRolesForChartGroup(chartGroupName string) ([]*RoleModel, error)
	DeleteRoles(roles []*RoleModel, tx *pg.Tx) error
}

type UserAuthRepositoryImpl struct {
	dbConnection                *pg.DB
	Logger                      *zap.SugaredLogger
	defaultAuthPolicyRepository DefaultAuthPolicyRepository
	defaultAuthRoleRepository   DefaultAuthRoleRepository
}

func NewUserAuthRepositoryImpl(dbConnection *pg.DB, Logger *zap.SugaredLogger,
	defaultAuthPolicyRepository DefaultAuthPolicyRepository,
	defaultAuthRoleRepository DefaultAuthRoleRepository) *UserAuthRepositoryImpl {
	return &UserAuthRepositoryImpl{
		dbConnection:                dbConnection,
		Logger:                      Logger,
		defaultAuthPolicyRepository: defaultAuthPolicyRepository,
		defaultAuthRoleRepository:   defaultAuthRoleRepository,
	}
}

type RoleModel struct {
	TableName   struct{} `sql:"roles" pg:",discard_unknown_columns"`
	Id          int      `sql:"id,pk"`
	Role        string   `sql:"role,notnull"`
	Entity      string   `sql:"entity"`
	Team        string   `sql:"team"`
	EntityName  string   `sql:"entity_name"`
	Environment string   `sql:"environment"`
	Action      string   `sql:"action"`
	AccessType  string   `sql:"access_type"`
	sql.AuditLog
}

type RolePolicyDetails struct {
	Team       string
	Env        string
	App        string
	TeamObj    string
	EnvObj     string
	AppObj     string
	Entity     string
	EntityName string
}

func (impl UserAuthRepositoryImpl) CreateRole(userModel *RoleModel, tx *pg.Tx) (*RoleModel, error) {
	err := tx.Insert(userModel)
	if err != nil {
		impl.Logger.Error(err)
		return userModel, err
	}
	return userModel, nil
}
func (impl UserAuthRepositoryImpl) GetRoleById(id int) (*RoleModel, error) {
	var model RoleModel
	err := impl.dbConnection.Model(&model).Where("id = ?", id).Select()
	if err != nil {
		impl.Logger.Error(err)
		return &model, err
	}
	return &model, nil
}
func (impl UserAuthRepositoryImpl) GetRolesByUserId(userId int32) ([]RoleModel, error) {
	var models []RoleModel
	err := impl.dbConnection.Model(&models).
		Column("role_model.*").
		Join("INNER JOIN user_roles ur on ur.role_id=role_model.id").
		Where("ur.user_id = ?", userId).Select()
	if err != nil {
		impl.Logger.Error(err)
		return models, err
	}
	return models, nil
}
func (impl UserAuthRepositoryImpl) GetRolesByGroupId(roleGroupId int32) ([]*RoleModel, error) {
	var models []*RoleModel
	err := impl.dbConnection.Model(&models).
		Column("role_model.*").
		Join("INNER JOIN role_group_role_mapping rgrm on rgrm.role_id=role_model.id").
		Join("INNER JOIN role_group rg on rg.id=rgrm.role_group_id").
		Where("rg.id = ?", roleGroupId).Select()
	if err != nil {
		impl.Logger.Error(err)
		return models, err
	}
	return models, nil
}
func (impl UserAuthRepositoryImpl) GetRole(role string) (*RoleModel, error) {
	var model RoleModel
	err := impl.dbConnection.Model(&model).Where("role = ?", role).Select()
	if err != nil {
		impl.Logger.Error(err)
		return &model, err
	}
	return &model, nil
}
func (impl UserAuthRepositoryImpl) GetAllRole() ([]RoleModel, error) {
	var models []RoleModel
	err := impl.dbConnection.Model(&models).Select()
	if err != nil {
		impl.Logger.Error(err)
		return models, err
	}
	return models, nil
}

func (impl UserAuthRepositoryImpl) GetRolesByActionAndAccessType(action string, accessType string) ([]RoleModel, error) {
	var models []RoleModel
	var err error
	if accessType == "" {
		err = impl.dbConnection.Model(&models).Where("action = ?", action).
			Where("access_type is NULL").
			Select()
	} else{
		err = impl.dbConnection.Model(&models).Where("action = ?", action).
			Where("access_type = ?", accessType).
			Select()
	}
	if err != nil {
		impl.Logger.Error("err in getting role by action", "err", err, "action", action, "accessType", accessType)
		return models, err
	}
	return models, nil
}

func (impl UserAuthRepositoryImpl) GetRoleByFilter(entity string, team string, app string, env string, act string, accessType string) (RoleModel, error) {
	var model RoleModel
	EMPTY := ""
	/*if act == "admin" {
		act = "*"
	}*/

	var err error
	if len(entity) > 0 && len(app) > 0 && act == "update" {
		query := "SELECT role.* FROM roles role WHERE role.entity = ? AND role.entity_name=? AND role.action=?"
		if len(accessType) == 0 {
			query = query + " and role.access_type is NULL"
		} else {
			query += " and role.access_type='" + accessType + "'"
		}
		_, err = impl.dbConnection.Query(&model, query, entity, app, act)
	} else if len(entity) > 0 && app == "" {
		query := "SELECT role.* FROM roles role WHERE role.entity = ? AND role.action=?"
		if len(accessType) == 0 {
			query = query + " and role.access_type is NULL"
		} else {
			query += " and role.access_type='" + accessType + "'"
		}
		_, err = impl.dbConnection.Query(&model, query, entity, act)
	} else {
		if len(team) > 0 && len(app) > 0 && len(env) > 0 && len(act) > 0 {
			query := "SELECT role.* FROM roles role WHERE role.team = ? AND role.entity_name=? AND role.environment=? AND role.action=?"
			if len(accessType) == 0 {
				query = query + " and role.access_type is NULL"
			} else {
				query += " and role.access_type='" + accessType + "'"
			}
			_, err = impl.dbConnection.Query(&model, query, team, app, env, act)
		} else if len(team) > 0 && app == "" && len(env) > 0 && len(act) > 0 {
			query := "SELECT role.* FROM roles role WHERE role.team=? AND coalesce(role.entity_name,'')=? AND role.environment=? AND role.action=?"
			if len(accessType) == 0 {
				query = query + " and role.access_type is NULL"
			} else {
				query += " and role.access_type='" + accessType + "'"
			}
			_, err = impl.dbConnection.Query(&model, query, team, EMPTY, env, act)
		} else if len(team) > 0 && len(app) > 0 && env == "" && len(act) > 0 {
			//this is applicable for all environment of a team
			query := "SELECT role.* FROM roles role WHERE role.team = ? AND role.entity_name=? AND coalesce(role.environment,'')=? AND role.action=?"
			if len(accessType) == 0 {
				query = query + " and role.access_type is NULL"
			} else {
				query += " and role.access_type='" + accessType + "'"
			}
			_, err = impl.dbConnection.Query(&model, query, team, app, EMPTY, act)
		} else if len(team) > 0 && app == "" && env == "" && len(act) > 0 {
			//this is applicable for all environment of a team
			query := "SELECT role.* FROM roles role WHERE role.team = ? AND coalesce(role.entity_name,'')=? AND coalesce(role.environment,'')=? AND role.action=?"
			if len(accessType) == 0 {
				query = query + " and role.access_type is NULL"
			} else {
				query += " and role.access_type='" + accessType + "'"
			}
			_, err = impl.dbConnection.Query(&model, query, team, EMPTY, EMPTY, act)
		} else if team == "" && app == "" && env == "" && len(act) > 0 {
			//this is applicable for super admin, all env, all team, all app
			query := "SELECT role.* FROM roles role WHERE coalesce(role.team,'') = ? AND coalesce(role.entity_name,'')=? AND coalesce(role.environment,'')=? AND role.action=?"
			if len(accessType) == 0 {
				query = query + " and role.access_type is NULL"
			} else {
				query += " and role.access_type='" + accessType + "'"
			}
			_, err = impl.dbConnection.Query(&model, query, EMPTY, EMPTY, EMPTY, act)
		} else if team == "" && app == "" && env == "" && act == "" {
			return model, nil
		} else {
			return model, nil
		}
	}

	if err != nil {
		impl.Logger.Errorw("exception while fetching roles", "err", err)
		return model, err
	}
	return model, nil
}

func (impl UserAuthRepositoryImpl) CreateUserRoleMapping(userRoleModel *UserRoleModel, tx *pg.Tx) (*UserRoleModel, error) {
	err := tx.Insert(userRoleModel)
	if err != nil {
		impl.Logger.Error(err)
		return userRoleModel, err
	}

	return userRoleModel, nil
}
func (impl UserAuthRepositoryImpl) GetUserRoleMappingByUserId(userId int32) ([]*UserRoleModel, error) {
	var userRoleModels []*UserRoleModel
	err := impl.dbConnection.Model(&userRoleModels).Where("user_id = ?", userId).Select()
	if err != nil {
		impl.Logger.Error(err)
		return userRoleModels, err
	}
	return userRoleModels, nil
}
func (impl UserAuthRepositoryImpl) DeleteUserRoleMapping(userRoleModel *UserRoleModel, tx *pg.Tx) (bool, error) {
	err := tx.Delete(userRoleModel)
	if err != nil {
		impl.Logger.Error(err)
		return false, err
	}
	return true, nil
}

func (impl UserAuthRepositoryImpl) CreateDefaultPolicies(team string, entityName string, env string, tx *pg.Tx) (bool, error) {
	transaction, err := impl.dbConnection.Begin()
	if err != nil {
		return false, err
	}

	//getting policies from db
	managerPoliciesDb, err := impl.defaultAuthPolicyRepository.GetPolicyByRoleType(MANAGER_TYPE)
	if err != nil {
		impl.Logger.Errorw("error in getting default policy by roleType", "err", err, "roleType", MANAGER_TYPE)
		return false, err
	}
	adminPoliciesDb, err := impl.defaultAuthPolicyRepository.GetPolicyByRoleType(ADMIN_TYPE)
	if err != nil {
		impl.Logger.Errorw("error in getting default policy by roleType", "err", err, "roleType", ADMIN_TYPE)
		return false, err
	}
	triggerPoliciesDb, err := impl.defaultAuthPolicyRepository.GetPolicyByRoleType(TRIGGER_TYPE)
	if err != nil {
		impl.Logger.Errorw("error in getting default policy by roleType", "err", err, "roleType", TRIGGER_TYPE)
		return false, err
	}
	viewPoliciesDb, err := impl.defaultAuthPolicyRepository.GetPolicyByRoleType(VIEW_TYPE)
	if err != nil {
		impl.Logger.Errorw("error in getting default policy by roleType", "err", err, "roleType", VIEW_TYPE)
		return false, err
	}

	//for START in Casbin Object
	teamObj := team
	envObj := env
	appObj := entityName
	if teamObj == "" {
		teamObj = "*"
	}
	if envObj == "" {
		envObj = "*"
	}
	if appObj == "" {
		appObj = "*"
	}

	rolePolicyDetails := RolePolicyDetails{
		Team:    team,
		App:     entityName,
		Env:     env,
		TeamObj: teamObj,
		EnvObj:  envObj,
		AppObj:  appObj,
	}

	//getting updated manager policies
	managerPolicies, err := util.Tprintf(managerPoliciesDb, rolePolicyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated policies", "err", err, "roleType", MANAGER_TYPE)
		return false, err
	}

	//getting updated admin policies
	adminPolicies, err := util.Tprintf(adminPoliciesDb, rolePolicyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated policies", "err", err, "roleType", ADMIN_TYPE)
		return false, err
	}

	//getting updated trigger policies
	triggerPolicies, err := util.Tprintf(triggerPoliciesDb, rolePolicyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated policies", "err", err, "roleType", TRIGGER_TYPE)
		return false, err
	}

	//getting updated view policies
	viewPolicies, err := util.Tprintf(viewPoliciesDb, rolePolicyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated policies", "err", err, "roleType", VIEW_TYPE)
		return false, err
	}

	//for START in Casbin Object Ends Here

	var policiesManager bean.PolicyRequest
	err = json.Unmarshal([]byte(managerPolicies), &policiesManager)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	impl.Logger.Debugw("add policy request", "policies", policiesManager)
	casbin.AddPolicy(policiesManager.Data)

	var policiesAdmin bean.PolicyRequest
	err = json.Unmarshal([]byte(adminPolicies), &policiesAdmin)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}

	impl.Logger.Debugw("add policy request", "policies", policiesAdmin)
	casbin.AddPolicy(policiesAdmin.Data)

	var policiesTrigger bean.PolicyRequest
	err = json.Unmarshal([]byte(triggerPolicies), &policiesTrigger)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	impl.Logger.Debugw("add policy request", "policies", policiesTrigger)
	casbin.AddPolicy(policiesTrigger.Data)

	var policiesView bean.PolicyRequest
	err = json.Unmarshal([]byte(viewPolicies), &policiesView)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	impl.Logger.Debugw("add policy request", "policies", policiesView)
	casbin.AddPolicy(policiesView.Data)

	//Creating ROLES
	//getting roles from db
	roleManagerDb, err := impl.defaultAuthRoleRepository.GetRoleByRoleType(MANAGER_TYPE)
	if err != nil {
		impl.Logger.Errorw("error in getting default role by roleType", "err", err, "roleType", MANAGER_TYPE)
		return false, err
	}
	roleAdminDb, err := impl.defaultAuthRoleRepository.GetRoleByRoleType(ADMIN_TYPE)
	if err != nil {
		impl.Logger.Errorw("error in getting default role by roleType", "err", err, "roleType", ADMIN_TYPE)
		return false, err
	}
	roleTriggerDb, err := impl.defaultAuthRoleRepository.GetRoleByRoleType(TRIGGER_TYPE)
	if err != nil {
		impl.Logger.Errorw("error in getting default role by roleType", "err", err, "roleType", TRIGGER_TYPE)
		return false, err
	}
	roleViewDb, err := impl.defaultAuthRoleRepository.GetRoleByRoleType(VIEW_TYPE)
	if err != nil {
		impl.Logger.Errorw("error in getting default role by roleType", "err", err, "roleType", VIEW_TYPE)
		return false, err
	}

	//getting updated manager role
	roleManager, err := util.Tprintf(roleManagerDb, rolePolicyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated role", "err", err, "roleType", MANAGER_TYPE)
		return false, err
	}

	//getting updated admin role
	roleAdmin, err := util.Tprintf(roleAdminDb, rolePolicyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated role", "err", err, "roleType", ADMIN_TYPE)
		return false, err
	}

	//getting updated trigger role
	roleTrigger, err := util.Tprintf(roleTriggerDb, rolePolicyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated role", "err", err, "roleType", TRIGGER_TYPE)
		return false, err
	}

	//getting updated view role
	roleView, err := util.Tprintf(roleViewDb, rolePolicyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated role", "err", err, "roleType", VIEW_TYPE)
		return false, err
	}

	var roleManagerData bean.RoleData
	err = json.Unmarshal([]byte(roleManager), &roleManagerData)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	_, err = impl.createRole(&roleManagerData, transaction)
	if err != nil && strings.Contains("duplicate key value violates unique constraint", err.Error()) {
		return false, err
	}

	var roleAdminData bean.RoleData
	err = json.Unmarshal([]byte(roleAdmin), &roleAdminData)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	_, err = impl.createRole(&roleAdminData, transaction)
	if err != nil && strings.Contains("duplicate key value violates unique constraint", err.Error()) {
		return false, err
	}

	var roleTriggerData bean.RoleData
	err = json.Unmarshal([]byte(roleTrigger), &roleTriggerData)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	_, err = impl.createRole(&roleTriggerData, transaction)
	if err != nil && strings.Contains("duplicate key value violates unique constraint", err.Error()) {
		return false, err
	}

	var roleViewData bean.RoleData
	err = json.Unmarshal([]byte(roleView), &roleViewData)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	_, err = impl.createRole(&roleViewData, transaction)
	if err != nil && strings.Contains("duplicate key value violates unique constraint", err.Error()) {
		return false, err
	}

	err = transaction.Commit()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (impl UserAuthRepositoryImpl) CreateDefaultHelmPolicies(team string, entityName string, env string, tx *pg.Tx) (bool, error) {
	transaction, err := impl.dbConnection.Begin()
	if err != nil {
		return false, err
	}
	adminPolicies := "{\r\n    \"data\": [\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"helm-app:admin_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"helm-app\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"<TEAM_OBJ>/<ENV_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"helm-app:admin_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"team\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"helm-app:admin_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"global-environment\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<ENV_OBJ>\"\r\n        }\r\n    ]\r\n}"
	editPolicies := "{\r\n    \"data\": [\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"helm-app:edit_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"helm-app\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>/<ENV_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"helm-app:edit_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"helm-app\",\r\n            \"act\": \"update\",\r\n            \"obj\": \"<TEAM_OBJ>/<ENV_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"helm-app:edit_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"global-environment\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<ENV_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"helm-app:edit_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"team\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>\"\r\n        }\r\n    ]\r\n}"
	viewPolicies := "{\r\n    \"data\": [\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"helm-app:view_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"helm-app\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>/<ENV_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"helm-app:view_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"global-environment\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<ENV_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"helm-app:view_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"team\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>\"\r\n        }\r\n    ]\r\n}"

	adminPolicies = strings.ReplaceAll(adminPolicies, "<TEAM>", team)
	adminPolicies = strings.ReplaceAll(adminPolicies, "<ENV>", env)
	adminPolicies = strings.ReplaceAll(adminPolicies, "<APP>", entityName)

	editPolicies = strings.ReplaceAll(editPolicies, "<TEAM>", team)
	editPolicies = strings.ReplaceAll(editPolicies, "<ENV>", env)
	editPolicies = strings.ReplaceAll(editPolicies, "<APP>", entityName)

	viewPolicies = strings.ReplaceAll(viewPolicies, "<TEAM>", team)
	viewPolicies = strings.ReplaceAll(viewPolicies, "<ENV>", env)
	viewPolicies = strings.ReplaceAll(viewPolicies, "<APP>", entityName)

	//for START in Casbin Object
	teamObj := team
	envObj := env
	appObj := entityName
	if teamObj == "" {
		teamObj = "*"
	}
	if envObj == "" {
		envObj = "*"
	}
	if appObj == "" {
		appObj = "*"
	}
	adminPolicies = strings.ReplaceAll(adminPolicies, "<TEAM_OBJ>", teamObj)
	adminPolicies = strings.ReplaceAll(adminPolicies, "<ENV_OBJ>", envObj)
	adminPolicies = strings.ReplaceAll(adminPolicies, "<APP_OBJ>", appObj)

	editPolicies = strings.ReplaceAll(editPolicies, "<TEAM_OBJ>", teamObj)
	editPolicies = strings.ReplaceAll(editPolicies, "<ENV_OBJ>", envObj)
	editPolicies = strings.ReplaceAll(editPolicies, "<APP_OBJ>", appObj)

	viewPolicies = strings.ReplaceAll(viewPolicies, "<TEAM_OBJ>", teamObj)
	viewPolicies = strings.ReplaceAll(viewPolicies, "<ENV_OBJ>", envObj)
	viewPolicies = strings.ReplaceAll(viewPolicies, "<APP_OBJ>", appObj)
	//for START in Casbin Object Ends Here

	var policiesAdmin bean.PolicyRequest
	err = json.Unmarshal([]byte(adminPolicies), &policiesAdmin)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	impl.Logger.Debugw("add policy request", "policies", policiesAdmin)
	casbin.AddPolicy(policiesAdmin.Data)

	var policiesEdit bean.PolicyRequest
	err = json.Unmarshal([]byte(editPolicies), &policiesEdit)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	impl.Logger.Debugw("add policy request", "policies", policiesEdit)
	casbin.AddPolicy(policiesEdit.Data)

	var policiesView bean.PolicyRequest
	err = json.Unmarshal([]byte(viewPolicies), &policiesView)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	impl.Logger.Debugw("add policy request", "policies", policiesView)
	casbin.AddPolicy(policiesView.Data)

	//Creating ROLES
	roleAdmin := "{\n    \"role\": \"helm-app:admin_<TEAM>_<ENV>_<APP>\",\n    \"casbinSubjects\": [\n        \"helm-app:admin_<TEAM>_<ENV>_<APP>\"\n    ],\n    \"team\": \"<TEAM>\",\n    \"entityName\": \"<APP>\",\n    \"environment\": \"<ENV>\",\n    \"action\": \"admin\",\n    \"accessType\": \"helm-app\"\n}"
	roleEdit := "{\n    \"role\": \"helm-app:edit_<TEAM>_<ENV>_<APP>\",\n    \"casbinSubjects\": [\n        \"helm-app:edit_<TEAM>_<ENV>_<APP>\"\n    ],\n    \"team\": \"<TEAM>\",\n    \"entityName\": \"<APP>\",\n    \"environment\": \"<ENV>\",\n    \"action\": \"edit\",\n    \"accessType\": \"helm-app\"\n}"
	roleView := "{\n    \"role\": \"helm-app:view_<TEAM>_<ENV>_<APP>\",\n    \"casbinSubjects\": [\n        \"helm-app:view_<TEAM>_<ENV>_<APP>\"\n    ],\n    \"team\": \"<TEAM>\",\n    \"entityName\": \"<APP>\",\n    \"environment\": \"<ENV>\",\n    \"action\": \"view\",\n    \"accessType\": \"helm-app\"\n}"

	roleAdmin = strings.ReplaceAll(roleAdmin, "<TEAM>", team)
	roleAdmin = strings.ReplaceAll(roleAdmin, "<ENV>", env)
	roleAdmin = strings.ReplaceAll(roleAdmin, "<APP>", entityName)

	roleEdit = strings.ReplaceAll(roleEdit, "<TEAM>", team)
	roleEdit = strings.ReplaceAll(roleEdit, "<ENV>", env)
	roleEdit = strings.ReplaceAll(roleEdit, "<APP>", entityName)

	roleView = strings.ReplaceAll(roleView, "<TEAM>", team)
	roleView = strings.ReplaceAll(roleView, "<ENV>", env)
	roleView = strings.ReplaceAll(roleView, "<APP>", entityName)

	var roleAdminData bean.RoleData
	err = json.Unmarshal([]byte(roleAdmin), &roleAdminData)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	_, err = impl.createRole(&roleAdminData, transaction)
	if err != nil && strings.Contains("duplicate key value violates unique constraint", err.Error()) {
		return false, err
	}

	var roleEditData bean.RoleData
	err = json.Unmarshal([]byte(roleEdit), &roleEditData)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	_, err = impl.createRole(&roleEditData, transaction)
	if err != nil && strings.Contains("duplicate key value violates unique constraint", err.Error()) {
		return false, err
	}

	var roleViewData bean.RoleData
	err = json.Unmarshal([]byte(roleView), &roleViewData)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	_, err = impl.createRole(&roleViewData, transaction)
	if err != nil && strings.Contains("duplicate key value violates unique constraint", err.Error()) {
		return false, err
	}

	err = transaction.Commit()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (impl UserAuthRepositoryImpl) CreateDefaultPoliciesForGlobalEntity(entity string, entityName string, action string, tx *pg.Tx) (bool, error) {
	transaction, err := impl.dbConnection.Begin()
	if err != nil {
		return false, err
	}
	// Rollback tx on error.
	defer transaction.Rollback()

	//getting policies from db
	entityAllPolicyDb, err := impl.defaultAuthPolicyRepository.GetPolicyByRoleType(ENTITY_ALL_TYPE)
	if err != nil {
		impl.Logger.Errorw("error in getting default policy by roleType", "err", err, "roleType", ENTITY_ALL_TYPE)
		return false, err
	}
	entityViewPolicyDb, err := impl.defaultAuthPolicyRepository.GetPolicyByRoleType(ENTITY_VIEW_TYPE)
	if err != nil {
		impl.Logger.Errorw("error in getting default policy by roleType", "err", err, "roleType", ENTITY_VIEW_TYPE)
		return false, err
	}

	policyDetails := RolePolicyDetails{
		Entity:     entity,
		EntityName: entityName,
	}

	//getting updated entityAll policies
	entityAllPolicy, err := util.Tprintf(entityAllPolicyDb, policyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated policies", "err", err, "roleType", ENTITY_ALL_TYPE)
		return false, err
	}

	//getting updated entityView policies
	entityViewPolicy, err := util.Tprintf(entityViewPolicyDb, policyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated policies", "err", err, "roleType", ENTITY_VIEW_TYPE)
		return false, err
	}

	//for START in Casbin Object Ends Here
	var policiesAdmin bean.PolicyRequest
	err = json.Unmarshal([]byte(entityAllPolicy), &policiesAdmin)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	impl.Logger.Debugw("add policy request", "policies", policiesAdmin)
	casbin.AddPolicy(policiesAdmin.Data)

	var policiesView bean.PolicyRequest
	err = json.Unmarshal([]byte(entityViewPolicy), &policiesView)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	impl.Logger.Debugw("add policy request", "policies", policiesView)
	casbin.AddPolicy(policiesView.Data)

	//getting policy from db
	entitySpecificPolicyDb, err := impl.defaultAuthPolicyRepository.GetPolicyByRoleType(ENTITY_SPECIFIC_TYPE)
	if err != nil {
		impl.Logger.Errorw("error in getting default policy by roleType", "err", err, "roleType", ENTITY_SPECIFIC_TYPE)
		return false, err
	}

	//getting updated entitySpecific policies
	entitySpecificPolicy, err := util.Tprintf(entitySpecificPolicyDb, policyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated policies", "err", err, "roleType", ENTITY_SPECIFIC_TYPE)
		return false, err
	}

	var policiesSpecific bean.PolicyRequest
	err = json.Unmarshal([]byte(entitySpecificPolicy), &policiesSpecific)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	impl.Logger.Debugw("add policy request", "policies", policiesSpecific)
	casbin.AddPolicy(policiesSpecific.Data)
	//CASBIN ENDS

	//Creating ROLES

	//getting role from db
	entitySpecificAdminDb, err := impl.defaultAuthRoleRepository.GetRoleByRoleType(ENTITY_SPECIFIC_ADMIN_TYPE)
	if err != nil {
		impl.Logger.Errorw("error in getting default policy by roleType", "err", err, "roleType", ENTITY_SPECIFIC_ADMIN_TYPE)
		return false, err
	}

	//getting updated role
	roleAdmin, err := util.Tprintf(entitySpecificAdminDb, policyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated policies", "err", err, "roleType", ENTITY_SPECIFIC_ADMIN_TYPE)
		return false, err
	}

	//getting role from db
	entitySpecificViewDb, err := impl.defaultAuthRoleRepository.GetRoleByRoleType(ENTITY_SPECIFIC_VIEW_TYPE)
	if err != nil {
		impl.Logger.Errorw("error in getting default policy by roleType", "err", err, "roleType", ENTITY_SPECIFIC_VIEW_TYPE)
		return false, err
	}

	//getting updated role
	roleView, err := util.Tprintf(entitySpecificViewDb, policyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated policies", "err", err, "roleType", ENTITY_SPECIFIC_VIEW_TYPE)
		return false, err
	}

	var roleAdminData bean.RoleData
	err = json.Unmarshal([]byte(roleAdmin), &roleAdminData)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	_, err = impl.GetRole(roleAdminData.Role)
	if err != nil || err == pg.ErrNoRows {
		_, err = impl.createRole(&roleAdminData, transaction)
		if err != nil && strings.Contains("duplicate key value violates unique constraint", err.Error()) {
			return false, err
		}
	}

	var roleViewData bean.RoleData
	err = json.Unmarshal([]byte(roleView), &roleViewData)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	_, err = impl.GetRole(roleViewData.Role)
	if err != nil || err == pg.ErrNoRows {
		_, err = impl.createRole(&roleViewData, transaction)
		if err != nil && strings.Contains("duplicate key value violates unique constraint", err.Error()) {
			return false, err
		}
	}

	//getting role from db
	roleSpecificDb, err := impl.defaultAuthRoleRepository.GetRoleByRoleType(ROLE_SPECIFIC_TYPE)
	if err != nil {
		impl.Logger.Errorw("error in getting default policy by roleType", "err", err, "roleType", ROLE_SPECIFIC_TYPE)
		return false, err
	}

	//getting updated role
	roleSpecific, err := util.Tprintf(roleSpecificDb, policyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated policies", "err", err, "roleType", ROLE_SPECIFIC_TYPE)
		return false, err
	}

	var roleSpecificData bean.RoleData
	err = json.Unmarshal([]byte(roleSpecific), &roleSpecificData)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	_, err = impl.GetRole(roleSpecificData.Role)
	if err != nil && err == pg.ErrNoRows {
		_, err = impl.createRole(&roleSpecificData, transaction)
		if err != nil && strings.Contains("duplicate key value violates unique constraint", err.Error()) {
			return false, err
		}
	}

	err = transaction.Commit()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (impl UserAuthRepositoryImpl) CreateRoleForSuperAdminIfNotExists(tx *pg.Tx) (bool, error) {
	transaction, err := impl.dbConnection.Begin()
	if err != nil {
		return false, err
	}

	//Creating ROLES
	roleModel, err := impl.GetRoleByFilter("", "", "", "", "super-admin", "")
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("Error in fetching role by filter", "err", err)
		return false, err
	}
	if roleModel.Id == 0 || err == pg.ErrNoRows {
		roleManager := "{\r\n    \"role\": \"role:super-admin___\",\r\n    \"casbinSubjects\": [\r\n        \"role:super-admin___\"\r\n    ],\r\n    \"team\": \"\",\r\n    \"entityName\": \"\",\r\n    \"environment\": \"\",\r\n    \"action\": \"super-admin\"\r\n}"

		var roleManagerData bean.RoleData
		err = json.Unmarshal([]byte(roleManager), &roleManagerData)
		if err != nil {
			impl.Logger.Errorw("decode err", "err", err)
			return false, err
		}
		_, err = impl.createRole(&roleManagerData, transaction)
		if err != nil && strings.Contains("duplicate key value violates unique constraint", err.Error()) {
			return false, err
		}
	}
	err = transaction.Commit()
	if err != nil {
		return false, err
	}
	return true, nil
}

func (impl UserAuthRepositoryImpl) createRole(roleData *bean.RoleData, tx *pg.Tx) (bool, error) {
	roleModel := &RoleModel{
		Role:        roleData.Role,
		Entity:      roleData.Entity,
		Team:        roleData.Team,
		EntityName:  roleData.EntityName,
		Environment: roleData.Environment,
		Action:      roleData.Action,
		AccessType:  roleData.AccessType,
	}
	roleModel, err := impl.CreateRole(roleModel, tx)
	if err != nil || roleModel == nil {
		return false, err
	}
	return true, nil
}

func (impl UserAuthRepositoryImpl) SyncOrchestratorToCasbin(team string, entityName string, env string, tx *pg.Tx) (bool, error) {

	//getting policies from db
	triggerPoliciesDb, err := impl.defaultAuthPolicyRepository.GetPolicyByRoleType(TRIGGER_TYPE)
	if err != nil {
		impl.Logger.Errorw("error in getting default policy by roleType", "err", err, "roleType", TRIGGER_TYPE)
		return false, err
	}
	viewPoliciesDb, err := impl.defaultAuthPolicyRepository.GetPolicyByRoleType(VIEW_TYPE)
	if err != nil {
		impl.Logger.Errorw("error in getting default policy by roleType", "err", err, "roleType", VIEW_TYPE)
		return false, err
	}

	//for START in Casbin Object
	teamObj := team
	envObj := env
	appObj := entityName
	if teamObj == "" {
		teamObj = "*"
	}
	if envObj == "" {
		envObj = "*"
	}
	if appObj == "" {
		appObj = "*"
	}

	policyDetails := RolePolicyDetails{
		Team:    team,
		App:     entityName,
		Env:     env,
		TeamObj: teamObj,
		EnvObj:  envObj,
		AppObj:  appObj,
	}

	//getting updated trigger policies
	triggerPolicies, err := util.Tprintf(triggerPoliciesDb, policyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated policies", "err", err, "roleType", TRIGGER_TYPE)
		return false, err
	}

	//getting updated view policies
	viewPolicies, err := util.Tprintf(viewPoliciesDb, policyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated policies", "err", err, "roleType", VIEW_TYPE)
		return false, err
	}

	//for START in Casbin Object Ends Here

	var policiesTrigger bean.PolicyRequest
	err = json.Unmarshal([]byte(triggerPolicies), &policiesTrigger)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	impl.Logger.Debugw("add policy request", "policies", policiesTrigger)
	casbin.AddPolicy(policiesTrigger.Data)

	var policiesView bean.PolicyRequest
	err = json.Unmarshal([]byte(viewPolicies), &policiesView)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	impl.Logger.Debugw("add policy request", "policies", policiesView)
	casbin.AddPolicy(policiesView.Data)

	return true, nil
}

func (impl UserAuthRepositoryImpl) UpdateTriggerPolicyForTerminalAccess() (err error) {
	newTriggerPolicy := `{
    "data": [
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "applications",
            "act": "get",
            "obj": "{{.TeamObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "applications",
            "act": "trigger",
            "obj": "{{.TeamObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "environment",
            "act": "trigger",
            "obj": "{{.EnvObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "environment",
            "act": "get",
            "obj": "{{.EnvObj}}/{{.AppObj}}"
        },
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "global-environment",
            "act": "get",
            "obj": "{{.EnvObj}}"
        },
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "team",
            "act": "get",
            "obj": "{{.TeamObj}}"
        },
        {
            "type": "p",
            "sub": "role:trigger_{{.Team}}_{{.Env}}_{{.App}}",
            "res": "terminal",
            "act": "exec",
            "obj": "{{.TeamObj}}/{{.EnvObj}}/{{.AppObj}}"
        }
    ]
}`
	err = impl.UpdateDefaultPolicyByRoleType(newTriggerPolicy, TRIGGER_TYPE)
	if err != nil {
		impl.Logger.Errorw("error in updating default policy for trigger role", "err", err)
		return err
	}
	return nil
}

func (impl UserAuthRepositoryImpl) GetDefaultPolicyByRoleType(roleType RoleType) (policy string, err error) {
	policy, err = impl.defaultAuthPolicyRepository.GetPolicyByRoleType(roleType)
	if err != nil {
		impl.Logger.Errorw("error in getting default policy by role type", "err", err, "roleType", roleType)
		return "", err
	}
	return policy, nil
}

func (impl UserAuthRepositoryImpl) UpdateDefaultPolicyByRoleType(newPolicy string, roleType RoleType) (err error) {
	//getting all roles by role type
	roles, err := impl.GetRolesByActionAndAccessType(string(roleType), "")
	if err != nil {
		impl.Logger.Errorw("error in getting roles for trigger action", "err", err)
		return err
	}
	oldPolicy, err := impl.defaultAuthPolicyRepository.GetPolicyByRoleType(roleType)
	if err != nil {
		impl.Logger.Errorw("error in getting default policy by roleType", "err", err, "roleType", roleType)
		return err
	}

	//updating new policy in db
	_, err = impl.defaultAuthPolicyRepository.UpdatePolicyByRoleType(newPolicy, roleType)
	if err != nil {
		impl.Logger.Errorw("error in updating default policy by roleType", "err", err, "roleType", roleType)
		return err
	}

	//getting diff between new and old policy(policies deleted/added)
	addedPolicies, deletedPolicies, err := impl.GetDiffBetweenPolicies(oldPolicy, newPolicy)
	if err != nil {
		impl.Logger.Errorw("error in getting diff between old and new policy", "err", err)
		return err
	}
	var addedPolicyFinal bean.PolicyRequest
	var deletedPolicyFinal bean.PolicyRequest
	for _, role := range roles {
		teamObj := role.Team
		envObj := role.Environment
		appObj := role.EntityName
		if teamObj == "" {
			teamObj = "*"
		}
		if envObj == "" {
			envObj = "*"
		}
		if appObj == "" {
			appObj = "*"
		}

		rolePolicyDetails := RolePolicyDetails{
			Team:    role.Team,
			Env:     role.Environment,
			App:     role.EntityName,
			TeamObj: teamObj,
			EnvObj:  envObj,
			AppObj:  appObj,
		}
		if len(addedPolicies) > 0 {
			addedPolicyReq, err := impl.GetUpdatedAddedOrDeletedPolicies(addedPolicies, rolePolicyDetails)
			if err != nil {
				impl.Logger.Errorw("error in getting updated added policies", "err", err)
				return err
			}
			addedPolicyFinal.Data = append(addedPolicyFinal.Data, addedPolicyReq.Data...)
		}
		if len(deletedPolicies) > 0 {
			deletedPolicyReq, err := impl.GetUpdatedAddedOrDeletedPolicies(deletedPolicies, rolePolicyDetails)
			if err != nil {
				impl.Logger.Errorw("error in getting updated deleted policies", "err", err)
				return err
			}
			deletedPolicyFinal.Data = append(deletedPolicyFinal.Data, deletedPolicyReq.Data...)
		}
	}
	//updating all policies(for all roles) in casbin
	if len(addedPolicyFinal.Data) > 0 {
		casbin.AddPolicy(addedPolicyFinal.Data)
	}
	if len(deletedPolicyFinal.Data) > 0 {
		casbin.RemovePolicy(deletedPolicyFinal.Data)
	}
	return nil
}

func (impl UserAuthRepositoryImpl) GetDiffBetweenPolicies(oldPolicy string, newPolicy string) (addedPolicies []casbin.Policy, deletedPolicies []casbin.Policy, err error) {
	var oldPolicyObj bean.PolicyRequest
	err = json.Unmarshal([]byte(oldPolicy), &oldPolicyObj)
	if err != nil {
		impl.Logger.Errorw("error in un-marshaling old policy", "err", err)
		return addedPolicies, deletedPolicies, err
	}

	var newPolicyObj bean.PolicyRequest
	err = json.Unmarshal([]byte(newPolicy), &newPolicyObj)
	if err != nil {
		impl.Logger.Errorw("error in un-marshaling new policy", "err", err)
		return addedPolicies, deletedPolicies, err
	}

	oldPolicyMap := make(map[string]bool)
	for _, oldPolicyData := range oldPolicyObj.Data {
		//converting all fields of data to a string
		data := fmt.Sprintf("type:%s,sub:%s,res:%s,act:%s,obj:%s", oldPolicyData.Type, oldPolicyData.Sub, oldPolicyData.Res, oldPolicyData.Act, oldPolicyData.Obj)
		//creating entry for data, keeping false because if present in new policy
		//then will be set to true and will not be included in deletedPolicies
		oldPolicyMap[data] = false
	}

	for _, newPolicyData := range newPolicyObj.Data {
		//converting all fields of data to a string
		data := fmt.Sprintf("type:%s,sub:%s,res:%s,act:%s,obj:%s", newPolicyData.Type, newPolicyData.Sub, newPolicyData.Res, newPolicyData.Act, newPolicyData.Obj)

		if _, ok := oldPolicyMap[data]; !ok {
			//data not present in old policy, to be included in addedPolicies
			addedPolicies = append(addedPolicies, newPolicyData)
		} else {
			//data present in old policy; set old policy to true, so it does not get included in deletedPolicies
			oldPolicyMap[data] = true
		}
	}

	//check oldPolicies for updating deletedPolicies
	for _, oldPolicyData := range oldPolicyObj.Data {
		data := fmt.Sprintf("type:%s,sub:%s,res:%s,act:%s,obj:%s", oldPolicyData.Type, oldPolicyData.Sub, oldPolicyData.Res, oldPolicyData.Act, oldPolicyData.Obj)
		if presentInNew := oldPolicyMap[data]; !presentInNew {
			//data not present in old policy, to be included in addedPolicies
			deletedPolicies = append(deletedPolicies, oldPolicyData)
		}
	}

	return addedPolicies, deletedPolicies, nil
}

func (impl UserAuthRepositoryImpl) GetUpdatedAddedOrDeletedPolicies(policies []casbin.Policy, rolePolicyDetails RolePolicyDetails) (bean.PolicyRequest, error) {
	var policyResp bean.PolicyRequest
	var policyReq bean.PolicyRequest
	policyReq.Data = policies
	policy, err := json.Marshal(policyReq)
	if err != nil {
		impl.Logger.Errorw("error in marshaling policy", "err", err)
		return policyResp, err
	}
	//getting updated policy
	updatedPolicy, err := util.Tprintf(string(policy), rolePolicyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated policy", "err", err)
		return policyResp, err
	}

	err = json.Unmarshal([]byte(updatedPolicy), &policyResp)
	if err != nil {
		impl.Logger.Errorw("error in un-marshaling policy", "err", err)
		return policyResp, err
	}
	return policyResp, nil
}

func (impl UserAuthRepositoryImpl) GetRolesForEnvironment(envName string) ([]*RoleModel, error) {
	var roles []*RoleModel
	err := impl.dbConnection.Model(&roles).Where("environment = ?", envName).Select()
	if err != nil {
		impl.Logger.Errorw("error in getting roles for environment", "err", err, "envName", envName)
		return nil, err
	}
	return roles, nil
}

func (impl UserAuthRepositoryImpl) GetRolesForProject(teamName string) ([]*RoleModel, error) {
	var roles []*RoleModel
	err := impl.dbConnection.Model(&roles).Where("team = ?", teamName).Select()
	if err != nil {
		impl.Logger.Errorw("error in getting roles for team", "err", err, "teamName", teamName)
		return nil, err
	}
	return roles, nil
}

func (impl UserAuthRepositoryImpl) GetRolesForApp(appName string) ([]*RoleModel, error) {
	var roles []*RoleModel
	err := impl.dbConnection.Model(&roles).Where("role not like ?", fmt.Sprintf("role:"+CHART_GROUP_TYPE+"%")).
		Where("entity_name = ?", appName).Select()
	if err != nil {
		impl.Logger.Errorw("error in getting roles for app", "err", err, "appName", appName)
		return nil, err
	}
	return roles, nil
}

func (impl UserAuthRepositoryImpl) GetRolesForChartGroup(chartGroupName string) ([]*RoleModel, error) {
	var roles []*RoleModel
	err := impl.dbConnection.Model(&roles).Where("role like ?", fmt.Sprintf("role:"+CHART_GROUP_TYPE+"%")).
		Where("entity_name = ?", chartGroupName).Select()
	if err != nil {
		impl.Logger.Errorw("error in getting roles for chart group", "err", err, "chartGroupName", chartGroupName)
		return nil, err
	}
	return roles, nil
}

func (impl UserAuthRepositoryImpl) DeleteRoles(roles []*RoleModel, tx *pg.Tx) error {
	err := tx.Delete(&roles)
	if err != nil {
		impl.Logger.Errorw("error in deleting roles", "err", err)
		return err
	}
	return nil
}
