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
	"github.com/devtron-labs/devtron/pkg/sql"
	bean2 "github.com/devtron-labs/devtron/pkg/user/bean"
	casbin2 "github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strings"
	"time"
)

type UserAuthRepository interface {
	CreateRole(role *RoleModel) (*RoleModel, error)
	CreateRoleWithTxn(userModel *RoleModel, tx *pg.Tx) (*RoleModel, error)
	GetRoleById(id int) (*RoleModel, error)
	GetRolesByIds(ids []int) ([]RoleModel, error)
	GetRoleByRoles(roles []string) ([]RoleModel, error)
	GetRolesByUserId(userId int32) ([]RoleModel, error)
	GetRolesByGroupId(userId int32) ([]*RoleModel, error)
	GetAllRole() ([]RoleModel, error)
	GetRolesByActionAndAccessType(action string, accessType string) ([]RoleModel, error)
	GetRoleByFilterForAllTypes(entity, team, app, env, act, accessType, cluster, namespace, group, kind, resource, action string, oldValues bool) (RoleModel, error)
	CreateUserRoleMapping(userRoleModel *UserRoleModel, tx *pg.Tx) (*UserRoleModel, error)
	GetUserRoleMappingByUserId(userId int32) ([]*UserRoleModel, error)
	DeleteUserRoleMapping(userRoleModel *UserRoleModel, tx *pg.Tx) (bool, error)
	DeleteUserRoleByRoleId(roleId int, tx *pg.Tx) error
	CreateDefaultPoliciesForAllTypes(team, entityName, env, entity, cluster, namespace, group, kind, resource, actionType, accessType string, UserId int32) (bool, error, []casbin2.Policy)
	CreateRoleForSuperAdminIfNotExists(tx *pg.Tx, UserId int32) (bool, error)
	SyncOrchestratorToCasbin(team string, entityName string, env string, tx *pg.Tx) (bool, error)
	UpdateTriggerPolicyForTerminalAccess() error
	GetRolesForEnvironment(envName, envIdentifier string) ([]*RoleModel, error)
	GetRolesForProject(teamName string) ([]*RoleModel, error)
	GetRolesForApp(appName string) ([]*RoleModel, error)
	GetRolesForChartGroup(chartGroupName string) ([]*RoleModel, error)
	DeleteRole(role *RoleModel, tx *pg.Tx) error
	//GetRoleByFilterForClusterEntity(cluster, namespace, group, kind, resource, action string) (RoleModel, error)
	GetRolesByUserIdAndEntityType(userId int32, entityType string) ([]*RoleModel, error)
	CreateRolesWithAccessTypeAndEntity(team, entityName, env, entity, cluster, namespace, group, kind, resource, actionType, accessType string, UserId int32, role string) (bool, error)
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
	Cluster     string   `sql:"cluster"`
	Namespace   string   `sql:"namespace"`
	Group       string   `sql:"group"`
	Kind        string   `sql:"kind"`
	Resource    string   `sql:"resource"`
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

	Cluster      string
	Namespace    string
	Group        string
	Kind         string
	Resource     string
	ClusterObj   string
	NamespaceObj string
	GroupObj     string
	KindObj      string
	ResourceObj  string
	Approver     bool
}

type ClusterRolePolicyDetails struct {
	Entity       string
	Cluster      string
	Namespace    string
	Group        string
	Kind         string
	Resource     string
	ClusterObj   string
	NamespaceObj string
	GroupObj     string
	KindObj      string
	ResourceObj  string
}

func (impl UserAuthRepositoryImpl) CreateRole(role *RoleModel) (*RoleModel, error) {
	err := impl.dbConnection.Insert(role)
	if err != nil {
		impl.Logger.Error("error in creating role", "err", err, "role", role)
		return role, err
	}
	return role, nil
}

func (impl UserAuthRepositoryImpl) CreateRoleWithTxn(userModel *RoleModel, tx *pg.Tx) (*RoleModel, error) {
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
func (impl UserAuthRepositoryImpl) GetRolesByIds(ids []int) ([]RoleModel, error) {
	var model []RoleModel
	err := impl.dbConnection.Model(&model).Where("id IN (?)", pg.In(ids)).Select()
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}
func (impl UserAuthRepositoryImpl) GetRoleByRoles(roles []string) ([]RoleModel, error) {
	var model []RoleModel
	err := impl.dbConnection.Model(&model).Where("role IN (?)", pg.In(roles)).Select()
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
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
	} else {
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

func (impl UserAuthRepositoryImpl) GetRoleByFilterForAllTypes(entity, team, app, env, act, accessType, cluster, namespace, group, kind, resource, action string, oldValues bool) (RoleModel, error) {
	var model RoleModel
	if entity == bean2.CLUSTER {

		query := "SELECT * FROM roles  WHERE entity = ? "
		var err error

		if len(cluster) > 0 {
			query += " and cluster='" + cluster + "' "
		} else {
			query += " and cluster IS NULL "
		}
		if len(namespace) > 0 {
			query += " and namespace='" + namespace + "' "
		} else {
			query += " and namespace IS NULL "
		}
		if len(group) > 0 {
			query += " and \"group\"='" + group + "' "
		} else {
			query += " and \"group\" IS NULL "
		}
		if len(kind) > 0 {
			query += " and kind='" + kind + "' "
		} else {
			query += " and kind IS NULL "
		}
		if len(resource) > 0 {
			query += " and resource='" + resource + "' "
		} else {
			query += " and resource IS NULL "
		}
		if len(action) > 0 {
			query += " and action='" + action + "' ;"
		} else {
			query += " and action IS NULL ;"
		}
		_, err = impl.dbConnection.Query(&model, query, bean.CLUSTER_ENTITIY)
		if err != nil {
			impl.Logger.Errorw("error in getting roles for clusterEntity", "err", err,
				bean2.CLUSTER, cluster, "namespace", namespace, "kind", kind, "group", group, "resource", resource)
			return model, err
		}
		return model, nil
	}

	EMPTY := ""
	/*if act == "admin" {
		act = "*"
	}*/

	var err error
	if entity == bean2.CHART_GROUP_TYPE && len(app) > 0 && act == "update" {
		query := "SELECT role.* FROM roles role WHERE role.entity = ? AND role.entity_name=? AND role.action=?"
		if len(accessType) == 0 {
			query = query + " and role.access_type is NULL"
		} else {
			query += " and role.access_type='" + accessType + "'"
		}
		_, err = impl.dbConnection.Query(&model, query, entity, app, act)
	} else if entity == bean2.CHART_GROUP_TYPE && app == "" {
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
			if oldValues {
				query = query + " and role.access_type is NULL"
			} else {
				query += " and role.access_type='" + accessType + "'"
			}

			_, err = impl.dbConnection.Query(&model, query, team, app, env, act)
		} else if len(team) > 0 && app == "" && len(env) > 0 && len(act) > 0 {

			query := "SELECT role.* FROM roles role WHERE role.team=? AND coalesce(role.entity_name,'')=? AND role.environment=? AND role.action=?"
			if oldValues {
				query = query + " and role.access_type is NULL"
			} else {
				query += " and role.access_type='" + accessType + "'"
			}
			_, err = impl.dbConnection.Query(&model, query, team, EMPTY, env, act)
		} else if len(team) > 0 && len(app) > 0 && env == "" && len(act) > 0 {
			//this is applicable for all environment of a team
			query := "SELECT role.* FROM roles role WHERE role.team = ? AND role.entity_name=? AND coalesce(role.environment,'')=? AND role.action=?"
			if oldValues {
				query = query + " and role.access_type is NULL"
			} else {
				query += " and role.access_type='" + accessType + "'"
			}

			_, err = impl.dbConnection.Query(&model, query, team, app, EMPTY, act)
		} else if len(team) > 0 && app == "" && env == "" && len(act) > 0 {
			//this is applicable for all environment of a team
			query := "SELECT role.* FROM roles role WHERE role.team = ? AND coalesce(role.entity_name,'')=? AND coalesce(role.environment,'')=? AND role.action=?"
			if oldValues {
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

func (impl UserAuthRepositoryImpl) DeleteUserRoleByRoleId(roleId int, tx *pg.Tx) error {
	var userRoleModel *UserRoleModel
	_, err := tx.Model(userRoleModel).
		Where("role_id = ?", roleId).Delete()
	if err != nil {
		impl.Logger.Error("err in deleting user role by role id", "err", err, "roleId", roleId)
		return err
	}
	return nil
}

func (impl UserAuthRepositoryImpl) CreateDefaultPoliciesForAllTypes(team, entityName, env, entity, cluster, namespace, group, kind, resource, actionType, accessType string, UserId int32) (bool, error, []casbin2.Policy) {
	//not using txn from parent caller because of conflicts in fetching of transactional save
	dbConnection := impl.dbConnection
	tx, err := dbConnection.Begin()
	var policiesToBeAdded []casbin2.Policy
	if err != nil {
		return false, err, policiesToBeAdded
	}
	// Rollback tx on error.
	defer tx.Rollback()

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

	clusterObj := cluster
	namespaceObj := namespace
	groupObj := group
	kindObj := kind
	resourceObj := resource

	if cluster == "" {
		clusterObj = "*"
	}
	if namespace == "" {
		namespaceObj = "*"
	}
	if group == "" {
		groupObj = "*"
	}
	if kind == "" {
		kindObj = "*"
	}
	if resource == "" {
		resourceObj = "*"
	}
	rolePolicyDetails := RolePolicyDetails{
		Team:         team,
		App:          entityName,
		Env:          env,
		TeamObj:      teamObj,
		EnvObj:       envObj,
		AppObj:       appObj,
		Entity:       entity,
		EntityName:   entityName,
		Cluster:      cluster,
		Namespace:    namespace,
		Group:        group,
		Kind:         kind,
		Resource:     resource,
		ClusterObj:   clusterObj,
		NamespaceObj: namespaceObj,
		GroupObj:     groupObj,
		KindObj:      kindObj,
		ResourceObj:  resourceObj,
	}

	//getting policies from db
	PoliciesDb, err := impl.defaultAuthPolicyRepository.GetPolicyByRoleTypeAndEntity(bean2.RoleType(actionType), accessType, entity)
	if err != nil {
		return false, err, policiesToBeAdded
	}
	//getting updated policies
	Policies, err := util.Tprintf(PoliciesDb, rolePolicyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated policies", "err", err, "roleType", bean2.RoleType(actionType), accessType)
		return false, err, policiesToBeAdded
	}
	//for START in Casbin Object Ends Here
	var policies bean.PolicyRequest
	err = json.Unmarshal([]byte(Policies), &policies)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err, policiesToBeAdded
	}
	impl.Logger.Debugw("add policy request", "policies", policies)
	policiesToBeAdded = append(policiesToBeAdded, policies.Data...)
	//Creating ROLES
	//getting roles from db
	roleDb, err := impl.defaultAuthRoleRepository.GetRoleByRoleTypeAndEntityType(bean2.RoleType(actionType), accessType, entity)
	if err != nil {
		return false, err, nil
	}
	role, err := util.Tprintf(roleDb, rolePolicyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated role", "err", err, "roleType", bean2.RoleType(actionType))
		return false, err, nil
	}
	//getting updated role
	var roleData bean.RoleData
	err = json.Unmarshal([]byte(role), &roleData)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err, nil
	}
	_, err = impl.createRole(&roleData, UserId)
	if err != nil && strings.Contains("duplicate key value violates unique constraint", err.Error()) {
		return false, err, nil
	}
	err = tx.Commit()
	if err != nil {
		return false, err, nil
	}
	return true, nil, policiesToBeAdded
}
func (impl UserAuthRepositoryImpl) CreateRolesWithAccessTypeAndEntity(team, entityName, env, entity, cluster, namespace, group, kind, resource, actionType, accessType string, UserId int32, role string) (bool, error) {
	roleData := bean.RoleData{
		Role:        role,
		Entity:      entity,
		Team:        team,
		EntityName:  entityName,
		Environment: env,
		Action:      actionType,
		AccessType:  accessType,
		Cluster:     cluster,
		Namespace:   namespace,
		Group:       group,
		Kind:        kind,
		Resource:    resource,
	}
	_, err := impl.createRole(&roleData, UserId)
	if err != nil && strings.Contains("duplicate key value violates unique constraint", err.Error()) {
		return false, err
	}
	return true, nil
}

func (impl UserAuthRepositoryImpl) CreateRoleForSuperAdminIfNotExists(tx *pg.Tx, UserId int32) (bool, error) {
	transaction, err := impl.dbConnection.Begin()
	if err != nil {
		return false, err
	}

	//Creating ROLES
	roleModel, err := impl.GetRoleByFilterForAllTypes("", "", "", "", bean2.SUPER_ADMIN, "", "", "", "", "", "", "", false)
	if err != nil && err != pg.ErrNoRows {
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
		_, err = impl.createRole(&roleManagerData, UserId)
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

func (impl UserAuthRepositoryImpl) createRole(roleData *bean.RoleData, UserId int32) (bool, error) {
	roleModel := &RoleModel{
		Role:        roleData.Role,
		Entity:      roleData.Entity,
		Team:        roleData.Team,
		EntityName:  roleData.EntityName,
		Environment: roleData.Environment,
		Action:      roleData.Action,
		AccessType:  roleData.AccessType,
		Cluster:     roleData.Cluster,
		Namespace:   roleData.Namespace,
		Group:       roleData.Group,
		Kind:        roleData.Kind,
		Resource:    roleData.Resource,
		AuditLog: sql.AuditLog{
			CreatedBy: UserId,
			CreatedOn: time.Now(),
			UpdatedBy: UserId,
			UpdatedOn: time.Now(),
		},
	}
	roleModel, err := impl.CreateRole(roleModel)
	if err != nil || roleModel == nil {
		return false, err
	}
	return true, nil
}

func (impl UserAuthRepositoryImpl) SyncOrchestratorToCasbin(team string, entityName string, env string, tx *pg.Tx) (bool, error) {

	//getting policies from db
	triggerPoliciesDb, err := impl.defaultAuthPolicyRepository.GetPolicyByRoleTypeAndEntity(bean2.TRIGGER_TYPE, bean2.DEVTRON_APP, bean2.ENTITY_APPS)
	if err != nil {
		return false, err
	}
	viewPoliciesDb, err := impl.defaultAuthPolicyRepository.GetPolicyByRoleTypeAndEntity(bean2.VIEW_TYPE, bean2.DEVTRON_APP, bean2.ENTITY_APPS)
	if err != nil {
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
		impl.Logger.Errorw("error in getting updated policies", "err", err, "roleType", bean2.TRIGGER_TYPE)
		return false, err
	}

	//getting updated view policies
	viewPolicies, err := util.Tprintf(viewPoliciesDb, policyDetails)
	if err != nil {
		impl.Logger.Errorw("error in getting updated policies", "err", err, "roleType", bean2.VIEW_TYPE)
		return false, err
	}

	//for START in Casbin Object Ends Here
	var policies []casbin2.Policy
	var policiesTrigger bean.PolicyRequest
	err = json.Unmarshal([]byte(triggerPolicies), &policiesTrigger)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	impl.Logger.Debugw("add policy request", "policies", policiesTrigger)
	policies = append(policies, policiesTrigger.Data...)
	var policiesView bean.PolicyRequest
	err = json.Unmarshal([]byte(viewPolicies), &policiesView)
	if err != nil {
		impl.Logger.Errorw("decode err", "err", err)
		return false, err
	}
	impl.Logger.Debugw("add policy request", "policies", policiesView)
	policies = append(policies, policiesView.Data...)
	casbin2.AddPolicy(policies)
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
	err = impl.UpdateDefaultPolicyByRoleType(newTriggerPolicy, bean2.TRIGGER_TYPE)
	if err != nil {
		impl.Logger.Errorw("error in updating default policy for trigger role", "err", err)
		return err
	}
	return nil
}

func (impl UserAuthRepositoryImpl) GetDefaultPolicyByRoleType(roleType bean2.RoleType) (policy string, err error) {
	policy, err = impl.defaultAuthPolicyRepository.GetPolicyByRoleTypeAndEntity(roleType, bean2.DEVTRON_APP, bean2.ENTITY_APPS)
	if err != nil {
		return "", err
	}
	return policy, nil
}

func (impl UserAuthRepositoryImpl) UpdateDefaultPolicyByRoleType(newPolicy string, roleType bean2.RoleType) (err error) {
	//getting all roles by role type
	roles, err := impl.GetRolesByActionAndAccessType(string(roleType), "")
	if err != nil {
		impl.Logger.Errorw("error in getting roles for trigger action", "err", err)
		return err
	}
	oldPolicy, err := impl.defaultAuthPolicyRepository.GetPolicyByRoleTypeAndEntity(roleType, bean2.DEVTRON_APP, bean2.ENTITY_APPS)
	if err != nil {
		return err
	}

	//updating new policy in db
	_, err = impl.defaultAuthPolicyRepository.UpdatePolicyByRoleType(newPolicy, roleType)
	if err != nil {
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
	//loading policy for safety
	casbin2.LoadPolicy()
	//updating all policies(for all roles) in casbin
	if len(addedPolicyFinal.Data) > 0 {
		casbin2.AddPolicy(addedPolicyFinal.Data)
	}
	if len(deletedPolicyFinal.Data) > 0 {
		casbin2.RemovePolicy(deletedPolicyFinal.Data)
	}
	//loading policy for syncing orchestrator to casbin with newly added policies
	casbin2.LoadPolicy()
	return nil
}

func (impl UserAuthRepositoryImpl) GetDiffBetweenPolicies(oldPolicy string, newPolicy string) (addedPolicies []casbin2.Policy, deletedPolicies []casbin2.Policy, err error) {
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

func (impl UserAuthRepositoryImpl) GetUpdatedAddedOrDeletedPolicies(policies []casbin2.Policy, rolePolicyDetails RolePolicyDetails) (bean.PolicyRequest, error) {
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

func (impl UserAuthRepositoryImpl) GetRolesForEnvironment(envName, envIdentifier string) ([]*RoleModel, error) {
	var roles []*RoleModel
	err := impl.dbConnection.Model(&roles).WhereOr("environment = ?", envName).
		WhereOr("environment = ?", envIdentifier).Select()
	if err != nil {
		impl.Logger.Errorw("error in getting roles for environment", "err", err, "envName", envName, "envIdentifier", envIdentifier)
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
	err := impl.dbConnection.Model(&roles).Where("entity is NULL").
		Where("entity_name = ?", appName).Select()
	if err != nil {
		impl.Logger.Errorw("error in getting roles for app", "err", err, "appName", appName)
		return nil, err
	}
	return roles, nil
}

func (impl UserAuthRepositoryImpl) GetRolesForChartGroup(chartGroupName string) ([]*RoleModel, error) {
	var roles []*RoleModel
	err := impl.dbConnection.Model(&roles).Where("entity = ?", bean2.CHART_GROUP_TYPE).
		Where("entity_name = ?", chartGroupName).Select()
	if err != nil {
		impl.Logger.Errorw("error in getting roles for chart group", "err", err, "chartGroupName", chartGroupName)
		return nil, err
	}
	return roles, nil
}

func (impl UserAuthRepositoryImpl) DeleteRole(role *RoleModel, tx *pg.Tx) error {
	err := tx.Delete(role)
	if err != nil {
		impl.Logger.Errorw("error in deleting role", "err", err, "role", role)
		return err
	}
	return nil
}

func (impl UserAuthRepositoryImpl) GetRolesByUserIdAndEntityType(userId int32, entityType string) ([]*RoleModel, error) {
	var models []*RoleModel
	err := impl.dbConnection.Model(&models).
		Column("role_model.*").
		Join("INNER JOIN user_roles ur on ur.role_id=role_model.id").
		Where("role_model.entity = ?", entityType).
		Where("ur.user_id = ?", userId).Select()
	if err != nil {
		impl.Logger.Error(err)
		return models, err
	}
	return models, nil
}
