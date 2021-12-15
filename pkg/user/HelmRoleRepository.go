package user

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strings"
)

type HelmUserRoleRepository interface {
	CreateRole(userModel *HelmRoleModel, tx *pg.Tx) (*HelmRoleModel, error)
	GetRoleById(id int) (*HelmRoleModel, error)
	GetRolesByUserId(userId int32) ([]HelmRoleModel, error)
	GetRolesByGroupId(userId int32) ([]*HelmRoleModel, error)
	GetAllRole() ([]HelmRoleModel, error)
	GetRoleByFilter(entity string, team string, app string, env string, act string) (HelmRoleModel, error)
	CreateUserRoleMapping(userRoleModel *HelmUserRoleModel, tx *pg.Tx) (*HelmUserRoleModel, error)
	GetUserRoleMappingByUserId(userId int32) ([]*HelmUserRoleModel, error)
	DeleteUserRoleMapping(userRoleModel *HelmUserRoleModel, tx *pg.Tx) (bool, error)

	CreateDefaultPolicies(team string, entityName string, env string, tx *pg.Tx) (bool, error)
	CreateDefaultPoliciesForGlobalEntity(entity string, entityName string, action string, tx *pg.Tx) (bool, error)
	CreateUpdateDefaultPoliciesForSuperAdmin(tx *pg.Tx) (bool, error)
	SyncOrchestratorToCasbin(team string, entityName string, env string, tx *pg.Tx) (bool, error)
}

type HelmUserRoleRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewHelmUserRoleRepositoryImpl(dbConnection *pg.DB, Logger *zap.SugaredLogger) *HelmUserRoleRepositoryImpl {
	return &HelmUserRoleRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

type HelmRoleModel struct {
	TableName   struct{} `sql:"roles" pg:",discard_unknown_columns"`
	Id          int      `sql:"id,pk"`
	Role        string   `sql:"role,notnull"`
	Entity      string   `sql:"entity"`
	Team        string   `sql:"team"`
	EntityName  string   `sql:"entity_name"`
	Environment string   `sql:"environment"`
	Action      string   `sql:"action"`
	sql.AuditLog
}

type HelmUserRoleModel struct {
	TableName struct{} `sql:"user_roles"`
	Id        int      `sql:"id,pk"`
	UserId    int32    `sql:"user_id,notnull"`
	RoleId    int      `sql:"role_id,notnull"`
	User      HelmUserModel
	sql.AuditLog
}

func (impl HelmUserRoleRepositoryImpl) CreateRole(userModel *HelmRoleModel, tx *pg.Tx) (*HelmRoleModel, error) {
	err := tx.Insert(userModel)
	if err != nil {
		impl.Logger.Error(err)
		return userModel, err
	}
	return userModel, nil
}
func (impl HelmUserRoleRepositoryImpl) GetRoleById(id int) (*HelmRoleModel, error) {
	var model HelmRoleModel
	err := impl.dbConnection.Model(&model).Where("id = ?", id).Select()
	if err != nil {
		impl.Logger.Error(err)
		return &model, err
	}
	return &model, nil
}
func (impl HelmUserRoleRepositoryImpl) GetRolesByUserId(userId int32) ([]HelmRoleModel, error) {
	var models []HelmRoleModel
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
func (impl HelmUserRoleRepositoryImpl) GetRolesByGroupId(roleGroupId int32) ([]*HelmRoleModel, error) {
	var models []*HelmRoleModel
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

func (impl HelmUserRoleRepositoryImpl) GetRole(role string) (*HelmRoleModel, error) {
	var model HelmRoleModel
	err := impl.dbConnection.Model(&model).Where("role = ?", role).Select()
	if err != nil {
		impl.Logger.Error(err)
		return &model, err
	}
	return &model, nil
}
func (impl HelmUserRoleRepositoryImpl) GetAllRole() ([]HelmRoleModel, error) {
	var models []HelmRoleModel
	err := impl.dbConnection.Model(&models).Select()
	if err != nil {
		impl.Logger.Error(err)
		return models, err
	}
	return models, nil
}
func (impl HelmUserRoleRepositoryImpl) GetRoleByFilter(entity string, team string, app string, env string, act string) (HelmRoleModel, error) {
	var models HelmRoleModel
	EMPTY := ""
	var err error
	if len(entity) > 0 && len(app) > 0 && act == "update" {
		query := "SELECT role.* FROM roles role WHERE role.entity = ? AND role.entity_name=? AND role.action=?"
		_, err = impl.dbConnection.Query(&models, query, entity, app, act)
	} else if len(entity) > 0 && app == "" {
		query := "SELECT role.* FROM roles role WHERE role.entity = ? AND role.action=?"
		_, err = impl.dbConnection.Query(&models, query, entity, act)
	} else {
		if len(team) > 0 && len(app) > 0 && len(env) > 0 && len(act) > 0 {
			query := "SELECT role.* FROM roles role WHERE role.team = ? AND role.entity_name=? AND role.environment=? AND role.action=?"
			_, err = impl.dbConnection.Query(&models, query, team, app, env, act)
		} else if len(team) > 0 && app == "" && len(env) > 0 && len(act) > 0 {
			query := "SELECT role.* FROM roles role WHERE role.team=? AND coalesce(role.entity_name,'')=? AND role.environment=? AND role.action=?"
			_, err = impl.dbConnection.Query(&models, query, team, EMPTY, env, act)
		} else if len(team) > 0 && len(app) > 0 && env == "" && len(act) > 0 {
			//this is applicable for all environment of a team
			query := "SELECT role.* FROM roles role WHERE role.team = ? AND role.entity_name=? AND coalesce(role.environment,'')=? AND role.action=?"
			_, err = impl.dbConnection.Query(&models, query, team, app, EMPTY, act)
		} else if len(team) > 0 && app == "" && env == "" && len(act) > 0 {
			//this is applicable for all environment of a team
			query := "SELECT role.* FROM roles role WHERE role.team = ? AND coalesce(role.entity_name,'')=? AND coalesce(role.environment,'')=? AND role.action=?"
			_, err = impl.dbConnection.Query(&models, query, team, EMPTY, EMPTY, act)
		} else if team == "" && app == "" && env == "" && len(act) > 0 {
			//this is applicable for super admin, all env, all team, all app
			query := "SELECT role.* FROM roles role WHERE coalesce(role.team,'') = ? AND coalesce(role.entity_name,'')=? AND coalesce(role.environment,'')=? AND role.action=?"
			_, err = impl.dbConnection.Query(&models, query, EMPTY, EMPTY, EMPTY, act)
		} else if team == "" && app == "" && env == "" && act == "" {
			return models, nil
		} else {
			return models, nil
		}
	}

	if err != nil {
		impl.Logger.Errorw("exception while fetching roles", "err", err)
		return models, err
	}
	return models, nil
}

func (impl HelmUserRoleRepositoryImpl) CreateUserRoleMapping(userRoleModel *HelmUserRoleModel, tx *pg.Tx) (*HelmUserRoleModel, error) {
	err := tx.Insert(userRoleModel)
	if err != nil {
		impl.Logger.Error(err)
		return userRoleModel, err
	}

	return userRoleModel, nil
}
func (impl HelmUserRoleRepositoryImpl) GetUserRoleMappingByUserId(userId int32) ([]*HelmUserRoleModel, error) {
	var userRoleModels []*HelmUserRoleModel
	err := impl.dbConnection.Model(&userRoleModels).Where("user_id = ?", userId).Select()
	if err != nil {
		impl.Logger.Error(err)
		return userRoleModels, err
	}
	return userRoleModels, nil
}
func (impl HelmUserRoleRepositoryImpl) DeleteUserRoleMapping(userRoleModel *HelmUserRoleModel, tx *pg.Tx) (bool, error) {
	err := tx.Delete(userRoleModel)
	if err != nil {
		impl.Logger.Error(err)
		return false, err
	}
	return true, nil
}

func (impl HelmUserRoleRepositoryImpl) CreateDefaultPolicies(team string, entityName string, env string, tx *pg.Tx) (bool, error) {
	transaction, err := impl.dbConnection.Begin()
	if err != nil {
		return false, err
	}

	managerPolicies := "{\r\n    \"data\": [\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:manager_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"applications\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"<TEAM_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:manager_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"environment\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"<ENV_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:manager_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"team\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"<TEAM_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:manager_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"user\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"<TEAM_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:manager_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"notification\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"<TEAM_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:manager_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"global-environment\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"<ENV_OBJ>\"\r\n        }\r\n    ]\r\n}"
	adminPolicies := "{\r\n    \"data\": [\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:admin_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"applications\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"<TEAM_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:admin_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"environment\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"<ENV_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:admin_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"team\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:admin_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"global-environment\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<ENV_OBJ>\"\r\n        }\r\n    ]\r\n}"
	triggerPolicies := "{\r\n    \"data\": [\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"applications\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"applications\",\r\n            \"act\": \"trigger\",\r\n            \"obj\": \"<TEAM_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"environment\",\r\n            \"act\": \"trigger\",\r\n            \"obj\": \"<ENV_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"environment\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<ENV_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"global-environment\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<ENV_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"team\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>\"\r\n        }\r\n    ]\r\n}"
	viewPolicies := "{\r\n    \"data\": [\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:view_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"applications\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:view_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"environment\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<ENV_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:view_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"global-environment\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<ENV_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:view_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"team\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>\"\r\n        }\r\n    ]\r\n}"

	managerPolicies = strings.ReplaceAll(managerPolicies, "<TEAM>", team)
	managerPolicies = strings.ReplaceAll(managerPolicies, "<ENV>", env)
	managerPolicies = strings.ReplaceAll(managerPolicies, "<APP>", entityName)

	adminPolicies = strings.ReplaceAll(adminPolicies, "<TEAM>", team)
	adminPolicies = strings.ReplaceAll(adminPolicies, "<ENV>", env)
	adminPolicies = strings.ReplaceAll(adminPolicies, "<APP>", entityName)

	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<TEAM>", team)
	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<ENV>", env)
	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<APP>", entityName)

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
	managerPolicies = strings.ReplaceAll(managerPolicies, "<TEAM_OBJ>", teamObj)
	managerPolicies = strings.ReplaceAll(managerPolicies, "<ENV_OBJ>", envObj)
	managerPolicies = strings.ReplaceAll(managerPolicies, "<APP_OBJ>", appObj)

	adminPolicies = strings.ReplaceAll(adminPolicies, "<TEAM_OBJ>", teamObj)
	adminPolicies = strings.ReplaceAll(adminPolicies, "<ENV_OBJ>", envObj)
	adminPolicies = strings.ReplaceAll(adminPolicies, "<APP_OBJ>", appObj)

	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<TEAM_OBJ>", teamObj)
	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<ENV_OBJ>", envObj)
	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<APP_OBJ>", appObj)

	viewPolicies = strings.ReplaceAll(viewPolicies, "<TEAM_OBJ>", teamObj)
	viewPolicies = strings.ReplaceAll(viewPolicies, "<ENV_OBJ>", envObj)
	viewPolicies = strings.ReplaceAll(viewPolicies, "<APP_OBJ>", appObj)
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
	roleManager := "{\r\n    \"role\": \"role:manager_<TEAM>_<ENV>_<APP>\",\r\n    \"casbinSubjects\": [\r\n        \"role:manager_<TEAM>_<ENV>_<APP>\"\r\n    ],\r\n    \"team\": \"<TEAM>\",\r\n    \"entityName\": \"<APP>\",\r\n    \"environment\": \"<ENV>\",\r\n    \"action\": \"manager\"\r\n}"
	roleAdmin := "{\n    \"role\": \"role:admin_<TEAM>_<ENV>_<APP>\",\n    \"casbinSubjects\": [\n        \"role:admin_<TEAM>_<ENV>_<APP>\"\n    ],\n    \"team\": \"<TEAM>\",\n    \"entityName\": \"<APP>\",\n    \"environment\": \"<ENV>\",\n    \"action\": \"admin\"\n}"
	roleTrigger := "{\n    \"role\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\n    \"casbinSubjects\": [\n        \"role:trigger_<TEAM>_<ENV>_<APP>\"\n    ],\n    \"team\": \"<TEAM>\",\n    \"entityName\": \"<APP>\",\n    \"environment\": \"<ENV>\",\n    \"action\": \"trigger\"\n}"
	roleView := "{\n    \"role\": \"role:view_<TEAM>_<ENV>_<APP>\",\n    \"casbinSubjects\": [\n        \"role:view_<TEAM>_<ENV>_<APP>\"\n    ],\n    \"team\": \"<TEAM>\",\n    \"entityName\": \"<APP>\",\n    \"environment\": \"<ENV>\",\n    \"action\": \"view\"\n}"
	roleManager = strings.ReplaceAll(roleManager, "<TEAM>", team)
	roleManager = strings.ReplaceAll(roleManager, "<ENV>", env)
	roleManager = strings.ReplaceAll(roleManager, "<APP>", entityName)

	roleAdmin = strings.ReplaceAll(roleAdmin, "<TEAM>", team)
	roleAdmin = strings.ReplaceAll(roleAdmin, "<ENV>", env)
	roleAdmin = strings.ReplaceAll(roleAdmin, "<APP>", entityName)

	roleTrigger = strings.ReplaceAll(roleTrigger, "<TEAM>", team)
	roleTrigger = strings.ReplaceAll(roleTrigger, "<ENV>", env)
	roleTrigger = strings.ReplaceAll(roleTrigger, "<APP>", entityName)

	roleView = strings.ReplaceAll(roleView, "<TEAM>", team)
	roleView = strings.ReplaceAll(roleView, "<ENV>", env)
	roleView = strings.ReplaceAll(roleView, "<APP>", entityName)

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

func (impl HelmUserRoleRepositoryImpl) CreateDefaultPoliciesForGlobalEntity(entity string, entityName string, action string, tx *pg.Tx) (bool, error) {
	transaction, err := impl.dbConnection.Begin()
	if err != nil {
		return false, err
	}
	// Rollback tx on error.
	defer transaction.Rollback()
	entityAllPolicy := "{\r\n    \"data\": [\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:<ENTITY>_admin\",\r\n            \"res\": \"<ENTITY>\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"*\"\r\n        }\r\n    ]\r\n}"
	entityViewPolicy := "{\r\n    \"data\": [\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:<ENTITY>_view\",\r\n            \"res\": \"<ENTITY>\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"*\"\r\n        }\r\n    ]\r\n}"

	entityAllPolicy = strings.ReplaceAll(entityAllPolicy, "<ENTITY>", entity)
	entityViewPolicy = strings.ReplaceAll(entityViewPolicy, "<ENTITY>", entity)

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

	entitySpecificPolicy := "{\r\n    \"data\": [\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:<ENTITY>_<ENTITY_NAME>_specific\",\r\n            \"res\": \"<ENTITY>\",\r\n            \"act\": \"update\",\r\n            \"obj\": \"<ENTITY_NAME>\"\r\n        },\r\n       {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:<ENTITY>_<ENTITY_NAME>_specific\",\r\n            \"res\": \"<ENTITY>\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<ENTITY_NAME>\"\r\n        }\r\n    ]\r\n}"
	entitySpecificPolicy = strings.ReplaceAll(entitySpecificPolicy, "<ENTITY>", entity)
	entitySpecificPolicy = strings.ReplaceAll(entitySpecificPolicy, "<ENTITY_NAME>", entityName)

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
	roleAdmin := "{\r\n    \"role\": \"role:<ENTITY>_admin\",\r\n    \"casbinSubjects\": [\r\n        \"role:<ENTITY>_admin\"\r\n    ],\r\n    \"entity\": \"<ENTITY>\",\r\n    \"team\": \"\",\r\n    \"application\": \"\",\r\n    \"environment\": \"\",\r\n    \"action\": \"admin\"\r\n}"
	roleAdmin = strings.ReplaceAll(roleAdmin, "<ENTITY>", entity)
	roleView := "{\r\n    \"role\": \"role:<ENTITY>_view\",\r\n    \"casbinSubjects\": [\r\n        \"role:<ENTITY>_view\"\r\n    ],\r\n    \"entity\": \"<ENTITY>\",\r\n    \"team\": \"\",\r\n    \"application\": \"\",\r\n    \"environment\": \"\",\r\n    \"action\": \"view\"\r\n}"
	roleView = strings.ReplaceAll(roleView, "<ENTITY>", entity)

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

	roleSpecific := "{\r\n    \"role\": \"role:<ENTITY>_<ENTITY_NAME>_specific\",\r\n    \"casbinSubjects\": [\r\n        \"role:<ENTITY>_<ENTITY_NAME>_specific\"\r\n    ],\r\n    \"entity\": \"<ENTITY>\",\r\n    \"team\": \"\",\r\n    \"entityName\": \"<ENTITY_NAME>\",\r\n    \"environment\": \"\",\r\n    \"action\": \"update\"\r\n}"
	roleSpecific = strings.ReplaceAll(roleSpecific, "<ENTITY>", entity)
	roleSpecific = strings.ReplaceAll(roleSpecific, "<ENTITY_NAME>", entityName)

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

func (impl HelmUserRoleRepositoryImpl) CreateUpdateDefaultPoliciesForSuperAdmin(tx *pg.Tx) (bool, error) {
	transaction, err := impl.dbConnection.Begin()
	if err != nil {
		return false, err
	}

	managerPolicies := "{\r\n    \"data\": [\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:super-admin___\",\r\n            \"res\": \"cluster\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"*\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:super-admin___\",\r\n            \"res\": \"git\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"*\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:super-admin___\",\r\n            \"res\": \"admin\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"*\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:super-admin___\",\r\n            \"res\": \"migrate\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"*\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:super-admin___\",\r\n            \"res\": \"applications\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"*\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:super-admin___\",\r\n            \"res\": \"environment\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"*\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:super-admin___\",\r\n            \"res\": \"team\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"*\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:super-admin___\",\r\n            \"res\": \"user\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"*\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:super-admin___\",\r\n            \"res\": \"notification\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"*\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:super-admin___\",\r\n            \"res\": \"global-environment\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"*\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:super-admin___\",\r\n            \"res\": \"chart-group\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"*\"\r\n        }\r\n    ]\r\n}"

	var policiesManager bean.PolicyRequest
	err = json.Unmarshal([]byte(managerPolicies), &policiesManager)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	impl.Logger.Debugw("add policy request", "policies", policiesManager)
	casbin.AddPolicy(policiesManager.Data)

	//Creating ROLES
	roleModel, err := impl.GetRoleByFilter("", "", "", "", "super-admin")
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

func (impl HelmUserRoleRepositoryImpl) createRole(roleData *bean.RoleData, tx *pg.Tx) (bool, error) {
	roleModel := &HelmRoleModel{
		Role:        roleData.Role,
		Entity:      roleData.Entity,
		Team:        roleData.Team,
		EntityName:  roleData.EntityName,
		Environment: roleData.Environment,
		Action:      roleData.Action,
	}
	roleModel, err := impl.CreateRole(roleModel, tx)
	if err != nil || roleModel == nil {
		return false, err
	}
	return true, nil
}

func (impl HelmUserRoleRepositoryImpl) SyncOrchestratorToCasbin(team string, entityName string, env string, tx *pg.Tx) (bool, error) {

	triggerPolicies := "{\r\n    \"data\": [\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"applications\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"applications\",\r\n            \"act\": \"trigger\",\r\n            \"obj\": \"<TEAM_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"environment\",\r\n            \"act\": \"trigger\",\r\n            \"obj\": \"<ENV_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"environment\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<ENV_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"global-environment\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<ENV_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"team\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>\"\r\n        }\r\n    ]\r\n}"
	viewPolicies := "{\r\n    \"data\": [\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:view_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"applications\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:view_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"environment\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<ENV_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:view_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"global-environment\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<ENV_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:view_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"team\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>\"\r\n        }\r\n    ]\r\n}"

	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<TEAM>", team)
	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<ENV>", env)
	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<APP>", entityName)

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

	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<TEAM_OBJ>", teamObj)
	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<ENV_OBJ>", envObj)
	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<APP_OBJ>", appObj)

	viewPolicies = strings.ReplaceAll(viewPolicies, "<TEAM_OBJ>", teamObj)
	viewPolicies = strings.ReplaceAll(viewPolicies, "<ENV_OBJ>", envObj)
	viewPolicies = strings.ReplaceAll(viewPolicies, "<APP_OBJ>", appObj)
	//for START in Casbin Object Ends Here

	var policiesTrigger bean.PolicyRequest
	err := json.Unmarshal([]byte(triggerPolicies), &policiesTrigger)
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
