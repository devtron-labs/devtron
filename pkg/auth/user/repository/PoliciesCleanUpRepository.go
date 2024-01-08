package repository

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type PoliciesCleanUpRepository interface {
	GetConnection() (dbConnection *pg.DB)
	GetAllUserMappedRoles() ([]*UserRoleModel, error)
	GetAllRolesForActiveGroups(ids []int32) ([]*RoleGroupRoleMapping, error)
	DeleteAllRolesExceptActive([]int32) error
	UpdateDuplicateAndChangeRoleMappingForUsers(tx *pg.Tx) error
	UpdateDuplicateAndChangeRoleMappingForGroup(tx *pg.Tx) error
	DeleteDuplicateRolesForUsers(tx *pg.Tx) ([]*RoleModel, error)
	DeleteDuplicateRolesForGroups(tx *pg.Tx) ([]*RoleModel, error)
	DeleteDuplicateMappingForAllUsers(tx *pg.Tx) error
	DeleteDuplicateMappingForAllGroups(tx *pg.Tx) error
	GetAllUnMappedRoles(ids []int32) ([]*RoleModel, error)
	GetAllUnusedRolesForCasbinCleanUp() ([]string, error)
	DeleteRoleGroupRoleMappingforInactiveGroups(tx *pg.Tx) error
	DeleteRoleGroupRoleMappingforInactiveUsers(tx *pg.Tx) error
}

type PoliciesCleanUpRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewPoliciesCleanUpRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *PoliciesCleanUpRepositoryImpl {
	return &PoliciesCleanUpRepositoryImpl{dbConnection: dbConnection, Logger: logger}
}

func (impl *PoliciesCleanUpRepositoryImpl) GetConnection() (dbConnection *pg.DB) {
	return impl.dbConnection
}

func (impl *PoliciesCleanUpRepositoryImpl) GetAllUserMappedRoles() ([]*UserRoleModel, error) {
	var model []*UserRoleModel
	err := impl.dbConnection.Model(&model).Select()
	return model, err
}

func (impl *PoliciesCleanUpRepositoryImpl) GetAllRolesForActiveGroups(ids []int32) ([]*RoleGroupRoleMapping, error) {
	var model []*RoleGroupRoleMapping
	err := impl.dbConnection.Model(&model).Where("role_group_id in (?)", pg.In(ids)).Select()
	return model, err
}

func (impl *PoliciesCleanUpRepositoryImpl) UpdateDuplicateAndChangeRoleMappingForUsers(tx *pg.Tx) error {
	var model []*UserRoleModel

	//to check fully for only user _mapped role groups
	query := "update user_roles ip set role_id = t.new_map_id from ( select m1.id old_map_id, m2.new_map_id from roles m1 join ( select role, min(id) as new_map_id from roles group by role having count(*) >1 ) m2 on (m1.role) = (m2.role) and m1.id != m2.new_map_id  ) t where t.old_map_id = ip.role_id;"

	_, err := tx.Query(&model, query)
	if err != nil {
		return err
	}

	return nil

}
func (impl *PoliciesCleanUpRepositoryImpl) DeleteDuplicateMappingForAllUsers(tx *pg.Tx) error {
	var roleModels []*UserRoleModel
	query := `WITH cte AS
				(SELECT user_roles.id, user_roles.role_id,
					MIN(user_roles.role_id) OVER
					(PARTITION BY user_id, role) AS min_role_id
					FROM user_roles  JOIN roles ON user_roles.role_id = roles.id)
				DELETE FROM user_roles WHERE id IN (SELECT id FROM cte WHERE role_id <> min_role_id);`
	_, err := tx.Query(&roleModels, query)
	if err != nil {
		return err
	}
	return nil
}
func (impl *PoliciesCleanUpRepositoryImpl) DeleteDuplicateMappingForAllGroups(tx *pg.Tx) error {
	var roleModels []*RoleGroupRoleMapping
	query := `WITH cte AS
				(SELECT role_group_role_mapping.id, role_group_role_mapping.role_id,
					MIN(role_group_role_mapping.role_id) OVER
					(PARTITION BY role_group_id, role) AS min_role_id
					FROM role_group_role_mapping  JOIN roles ON role_group_role_mapping.role_id = roles.id)
				DELETE FROM role_group_role_mapping WHERE id IN (SELECT id FROM cte WHERE role_id <> min_role_id);`
	_, err := tx.Query(&roleModels, query)
	if err != nil {
		return err
	}
	return nil
}
func (impl *PoliciesCleanUpRepositoryImpl) UpdateDuplicateAndChangeRoleMappingForGroup(tx *pg.Tx) error {
	var models []*RoleGroupRoleMapping
	//to check fully for only active groups
	query := `update role_group_role_mapping ip set role_id = t.new_map_id from
 				(select m1.id old_map_id, m2.new_map_id from roles m1 join 
 					(select role, min(id) as new_map_id from roles group by role having count(*)>1) 
 				m2 on (m1.role) = (m2.role) and m1.id != m2.new_map_id) t where t.old_map_id = ip.role_id;`
	_, err := tx.Query(&models, query)
	if err != nil {
		return err
	}
	return nil
}
func (impl *PoliciesCleanUpRepositoryImpl) DeleteDuplicateRolesForUsers(tx *pg.Tx) ([]*RoleModel, error) {
	var roleModels []*RoleModel
	query := "delete from roles stm where not exists (select * from user_roles ip where ip.role_id = stm.id) RETURNING * ;"
	_, err := tx.Query(&roleModels, query)
	if err != nil {
		return roleModels, err
	}
	return roleModels, nil

}
func (impl *PoliciesCleanUpRepositoryImpl) DeleteDuplicateRolesForGroups(tx *pg.Tx) ([]*RoleModel, error) {
	var roleModels []*RoleModel
	query := "delete from roles stm where not exists (select * from role_group_role_mapping ip where ip.role_id = stm.id) RETURNING * ;"
	_, err := tx.Query(&roleModels, query)
	if err != nil {
		return roleModels, err
	}
	return roleModels, nil
}

func (impl *PoliciesCleanUpRepositoryImpl) DeleteAllRolesExceptActive(ids []int32) error {
	var roleModel *RoleModel
	//query := "delete from roles where id not in (?) ;"
	//_, err := impl.dbConnection.Query(&roleModels, query, pg.In(ids))
	_, err := impl.dbConnection.Model(roleModel).Where("id not in (?)", pg.In(ids)).Delete()
	if err != nil {
		impl.Logger.Error("err in deleting roles not in role ids", "err", err, "roleIds", ids)
		return err
	}
	return nil
}
func (impl *PoliciesCleanUpRepositoryImpl) GetAllUnusedRolesForCasbinCleanUp() ([]string, error) {
	var model []string
	query := `select role from roles where role not in 
				(select r.role  from roles r inner join user_roles ur on ur.role_id=r.id 
				union select r.role from roles r 
				inner join role_group_role_mapping rgrm  on rgrm.role_id=r.id);`
	_, err := impl.dbConnection.Query(&model, query)
	if err != nil {
		return model, err
	}

	return model, err
}
func (impl *PoliciesCleanUpRepositoryImpl) GetAllUnMappedRoles(ids []int32) ([]*RoleModel, error) {
	var roleModels []*RoleModel
	query := "select from roles where id not in (?) ;"
	_, err := impl.dbConnection.Query(&roleModels, query, pg.In(ids))
	if err != nil {
		return roleModels, err
	}
	return roleModels, nil
}
func (impl *PoliciesCleanUpRepositoryImpl) DeleteRoleGroupRoleMappingforInactiveUsers(tx *pg.Tx) error {
	var userRoleModels []*UserRoleModel
	query := "delete from user_roles where user_id in ( select id from users where active=false);"
	_, err := tx.Query(userRoleModels, query)
	if err != nil {
		return err
	}
	return nil
}
func (impl *PoliciesCleanUpRepositoryImpl) DeleteRoleGroupRoleMappingforInactiveGroups(tx *pg.Tx) error {
	var roleGroupsRoleModels []*RoleGroupRoleMapping
	query := "delete from role_group_role_mapping where role_group_id in ( select id from role_group where active=false);"
	_, err := tx.Query(roleGroupsRoleModels, query)
	if err != nil {
		return err
	}
	return nil
}
