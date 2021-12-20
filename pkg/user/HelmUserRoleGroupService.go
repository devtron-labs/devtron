package user

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	casbin2 "github.com/devtron-labs/devtron/pkg/user/casbin"
	bean2 "github.com/devtron-labs/devtron/pkg/user/dto"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
	"strings"
	"time"
)

type HelmUserRoleGroupService interface {
	CreateRoleGroup(request *bean2.HelmRoleGroupDto) (*bean2.HelmRoleGroupDto, error)
	UpdateRoleGroup(request *bean2.HelmRoleGroupDto) (*bean2.HelmRoleGroupDto, error)

	FetchRoleGroupsById(id int32) (*bean2.HelmRoleGroupDto, error)
	FetchRoleGroups() ([]*bean2.HelmRoleGroupDto, error)
}

type HelmUserRoleGroupServiceImpl struct {
	userAuthRepository  HelmUserRoleRepository
	logger              *zap.SugaredLogger
	userRepository      HelmUserRepository
	roleGroupRepository HelmUserRoleGroupRepository
}

func NewHelmUserRoleGroupServiceImpl(userAuthRepository HelmUserRoleRepository,
	logger *zap.SugaredLogger, userRepository HelmUserRepository,
	roleGroupRepository HelmUserRoleGroupRepository) *HelmUserRoleGroupServiceImpl {
	serviceImpl := &HelmUserRoleGroupServiceImpl{
		userAuthRepository:  userAuthRepository,
		logger:              logger,
		userRepository:      userRepository,
		roleGroupRepository: roleGroupRepository,
	}
	cStore = sessions.NewCookieStore(randKey())
	return serviceImpl
}

func (impl HelmUserRoleGroupServiceImpl) CreateRoleGroup(request *bean2.HelmRoleGroupDto) (*bean2.HelmRoleGroupDto, error) {
	dbConnection := impl.roleGroupRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	if request.Id > 0 {
		_, err := impl.roleGroupRepository.GetRoleGroupById(request.Id)
		if err != nil {
			impl.logger.Errorw("error while fetching user from db", "error", err)
			return nil, err
		}
	} else {

		//create new user in our db on d basis of info got from google api or hex. assign a basic role
		model := &HelmRoleGroup{
			Name:        request.Name,
			Description: request.Description,
		}
		rgName := strings.ToLower(request.Name)
		object := "group:" + strings.ReplaceAll(rgName, " ", "_")
		model.CasbinName = object
		model.CreatedBy = request.UserId
		model.UpdatedBy = request.UserId
		model.CreatedOn = time.Now()
		model.UpdatedOn = time.Now()
		model.Active = true
		model, err := impl.roleGroupRepository.CreateRoleGroup(model, tx)
		request.Id = model.Id
		if err != nil {
			impl.logger.Errorw("error in creating new user group", "error", err)
			err = &util.ApiError{
				Code:            constants.UserCreateDBFailed,
				InternalMessage: "failed to create new user in db",
				UserMessage:     fmt.Sprintf("requested by"),
			}
			return request, err
		}
		model.Id = model.Id

		//Starts Role and Mapping
		var policies []casbin2.Policy
		for _, roleFilter := range request.RoleFilters {
			if roleFilter.EntityName == "" {
				roleFilter.EntityName = "NONE"
			}
			if roleFilter.Environment == "" {
				roleFilter.Environment = "NONE"
			}
			entityNames := strings.Split(roleFilter.EntityName, ",")
			environments := strings.Split(roleFilter.Environment, ",")
			for _, environment := range environments {
				for _, entityName := range entityNames {
					if entityName == "NONE" {
						entityName = ""
					}
					if environment == "NONE" {
						environment = ""
					}
					roleModel, err := impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action)
					if err != nil {
						impl.logger.Errorw("Error in fetching role by filter", "user", request)
						return nil, err
					}
					if roleModel.Id == 0 {
						impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
						//userInfo.Status = "role not fount for any given filter: " + roleFilter.Team + "," + roleFilter.Environment + "," + roleFilter.Application + "," + roleFilter.Action

						//TODO - create roles from here
						if len(roleFilter.Team) > 0 && len(roleFilter.Environment) > 0 {
							flag, err := impl.userAuthRepository.CreateDefaultPolicies(roleFilter.Team, entityName, environment, tx)
							if err != nil || flag == false {
								return nil, err
							}
							roleModel, err = impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action)
							if err != nil {
								impl.logger.Errorw("Error in fetching role by filter", "user", request)
								return nil, err
							}
							if roleModel.Id == 0 {
								impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
								request.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
								continue
							}
						} else if len(roleFilter.Entity) > 0 {
							flag, err := impl.userAuthRepository.CreateDefaultPoliciesForGlobalEntity(roleFilter.Entity, entityName, roleFilter.Action, tx)
							if err != nil || flag == false {
								return nil, err
							}
							roleModel, err = impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action)
							if err != nil {
								impl.logger.Errorw("Error in fetching role by filter", "user", request)
								return nil, err
							}
							if roleModel.Id == 0 {
								impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
								request.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
								continue
							}
						} else {
							continue
						}
					}

					if roleModel.Id > 0 {
						roleGroupMappingModel := &HelmRoleGroupRoleMapping{RoleGroupId: model.Id, RoleId: roleModel.Id}
						roleGroupMappingModel.CreatedBy = request.UserId
						roleGroupMappingModel.UpdatedBy = request.UserId
						roleGroupMappingModel.CreatedOn = time.Now()
						roleGroupMappingModel.UpdatedOn = time.Now()
						roleGroupMappingModel, err = impl.roleGroupRepository.CreateRoleGroupRoleMapping(roleGroupMappingModel, tx)
						if err != nil {
							return nil, err
						}
						policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(model.CasbinName), Obj: casbin2.Object(roleModel.Role)})
					}
				}
			}
		}

		if len(policies) > 0 {
			pRes := casbin2.AddPolicy(policies)
			println(pRes)
		}
		//Ends
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (impl HelmUserRoleGroupServiceImpl) UpdateRoleGroup(request *bean2.HelmRoleGroupDto) (*bean2.HelmRoleGroupDto, error) {
	dbConnection := impl.roleGroupRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	roleGroup, err := impl.roleGroupRepository.GetRoleGroupById(request.Id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	//policyGroup.Name = request.Name
	roleGroup.Description = request.Description
	roleGroup.UpdatedOn = time.Now()
	roleGroup.UpdatedBy = request.UserId
	roleGroup.Active = true
	roleGroup, err = impl.roleGroupRepository.UpdateRoleGroup(roleGroup, tx)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}

	roleGroupMappingModels, err := impl.roleGroupRepository.GetRoleGroupRoleMappingByRoleGroupId(roleGroup.Id)
	if err != nil {
		return nil, err
	}

	// DELETE PROCESS STARTS
	existingRoleIds := make(map[int]*HelmRoleGroupRoleMapping)
	remainingExistingRoleIds := make(map[int]*HelmRoleGroupRoleMapping)
	for _, item := range roleGroupMappingModels {
		existingRoleIds[item.RoleId] = item
		remainingExistingRoleIds[item.RoleId] = item
	}

	// Filter out removed items in current request
	//var policies []casbin2.Policy
	for _, roleFilter := range request.RoleFilters {
		if roleFilter.EntityName == "" {
			roleFilter.EntityName = "NONE"
		}
		if roleFilter.Environment == "" {
			roleFilter.Environment = "NONE"
		}
		entityNames := strings.Split(roleFilter.EntityName, ",")
		environments := strings.Split(roleFilter.Environment, ",")
		for _, environment := range environments {
			for _, entityName := range entityNames {
				if entityName == "NONE" {
					entityName = ""
				}
				if environment == "NONE" {
					environment = ""
				}
				roleModel, err := impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action)
				if err != nil {
					impl.logger.Errorw("Error in fetching roles by filter", "user", request)
					return nil, err
				}
				if roleModel.Id == 0 {
					impl.logger.Warnw("no role found for given filter", "filter", roleFilter)
					request.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
					continue
				}
				//roleModel := roleModels[0]
				if _, ok := existingRoleIds[roleModel.Id]; ok {
					delete(remainingExistingRoleIds, roleModel.Id)
				}
			}
		}
	}

	//delete remaining Ids from casbin role mapping table in orchestrator and casbin policy db
	// which are existing but not provided in this request
	var policiesRemove []casbin2.Policy
	for _, model := range remainingExistingRoleIds {
		_, err := impl.roleGroupRepository.DeleteRoleGroupRoleMapping(model, tx)
		if err != nil {
			return nil, err
		}
		role, err := impl.userAuthRepository.GetRoleById(model.RoleId)
		if err != nil {
			return nil, err
		}
		policyGroup, err := impl.roleGroupRepository.GetRoleGroupById(model.RoleGroupId)
		if err != nil {
			return nil, err
		}
		policiesRemove = append(policiesRemove, casbin2.Policy{Type: "g", Sub: casbin2.Subject(policyGroup.CasbinName), Obj: casbin2.Object(role.Role)})
	}
	if len(policiesRemove) > 0 {
		pRes := casbin2.RemovePolicy(policiesRemove)
		println(pRes)
	}
	// DELETE PROCESS ENDS

	//Adding New Policies
	var policies []casbin2.Policy
	for _, roleFilter := range request.RoleFilters {
		if roleFilter.EntityName == "" {
			roleFilter.EntityName = "NONE"
		}
		if roleFilter.Environment == "" {
			roleFilter.Environment = "NONE"
		}
		entityNames := strings.Split(roleFilter.EntityName, ",")
		environments := strings.Split(roleFilter.Environment, ",")
		for _, environment := range environments {
			for _, entityName := range entityNames {
				if entityName == "NONE" {
					entityName = ""
				}
				if environment == "NONE" {
					environment = ""
				}
				roleModel, err := impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action)
				if err != nil {
					impl.logger.Errorw("Error in fetching role by filter", "user", request)
					return nil, err
				}
				if roleModel.Id == 0 {
					impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
					request.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action

					//TODO - create roles from here
					if len(roleFilter.Team) > 0 && len(roleFilter.Environment) > 0 {
						flag, err := impl.userAuthRepository.CreateDefaultPolicies(roleFilter.Team, entityName, environment, tx)
						if err != nil || flag == false {
							return nil, err
						}
						roleModel, err = impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action)
						if err != nil {
							impl.logger.Errorw("Error in fetching role by filter", "user", request)
							return nil, err
						}
						if roleModel.Id == 0 {
							impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
							request.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
							continue
						}
					} else if len(roleFilter.Entity) > 0 {
						flag, err := impl.userAuthRepository.CreateDefaultPoliciesForGlobalEntity(roleFilter.Entity, entityName, roleFilter.Action, tx)
						if err != nil || flag == false {
							return nil, err
						}
						roleModel, err = impl.userAuthRepository.GetRoleByFilter(roleFilter.Entity, roleFilter.Team, entityName, environment, roleFilter.Action)
						if err != nil {
							impl.logger.Errorw("Error in fetching role by filter", "user", request)
							return nil, err
						}
						if roleModel.Id == 0 {
							impl.logger.Debugw("no role found for given filter", "filter", roleFilter)
							request.Status = "role not fount for any given filter: " + roleFilter.Team + "," + environment + "," + entityName + "," + roleFilter.Action
							continue
						}
					} else {
						continue
					}
				}

				if roleModel.Id > 0 {
					if _, ok := existingRoleIds[roleModel.Id]; ok {
						//Adding policies which is removed
						policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(roleGroup.CasbinName), Obj: casbin2.Object(roleModel.Role)})
					} else {
						//new role ids in new array, add it
						roleGroupMappingModel := &HelmRoleGroupRoleMapping{RoleGroupId: request.Id, RoleId: roleModel.Id}
						roleGroupMappingModel.CreatedBy = request.UserId
						roleGroupMappingModel.UpdatedBy = request.UserId
						roleGroupMappingModel.CreatedOn = time.Now()
						roleGroupMappingModel.UpdatedOn = time.Now()
						roleGroupMappingModel, err = impl.roleGroupRepository.CreateRoleGroupRoleMapping(roleGroupMappingModel, tx)
						if err != nil {
							return nil, err
						}
						policies = append(policies, casbin2.Policy{Type: "g", Sub: casbin2.Subject(roleGroup.CasbinName), Obj: casbin2.Object(roleModel.Role)})
					}
				}
			}
		}
	}

	//updating in casbin
	if len(policies) > 0 {
		casbin2.AddPolicy(policies)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return request, nil
}

func (impl HelmUserRoleGroupServiceImpl) FetchRoleGroupsById(id int32) (*bean2.HelmRoleGroupDto, error) {
	roleGroup, err := impl.roleGroupRepository.GetRoleGroupById(id)
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}

	roles, err := impl.userAuthRepository.GetRolesByGroupId(roleGroup.Id)
	if err != nil {
		impl.logger.Errorw("No Roles Found for user", "roleGroupId", roleGroup.Id)
	}
	var roleFilters []bean2.RoleFilter
	roleFilterMap := make(map[string]*bean2.RoleFilter)
	for _, role := range roles {
		key := ""
		if len(role.Team) > 0 {
			key = fmt.Sprintf("%s_%s", role.Team, role.Action)
		} else if len(role.Entity) > 0 {
			key = fmt.Sprintf("%s_%s", role.Entity, role.Action)
		}
		if _, ok := roleFilterMap[key]; ok {
			envArr := strings.Split(roleFilterMap[key].Environment, ",")
			if containsArr(envArr, AllEnvironment) {
				roleFilterMap[key].Environment = AllEnvironment
			} else if !containsArr(envArr, role.Environment) {
				roleFilterMap[key].Environment = fmt.Sprintf("%s,%s", roleFilterMap[key].Environment, role.Environment)
			}
			entityArr := strings.Split(roleFilterMap[key].EntityName, ",")
			if !containsArr(entityArr, role.EntityName) {
				roleFilterMap[key].EntityName = fmt.Sprintf("%s,%s", roleFilterMap[key].EntityName, role.EntityName)
			}
		} else {
			roleFilterMap[key] = &bean2.RoleFilter{
				Entity:      role.Entity,
				Team:        role.Team,
				Environment: role.Environment,
				EntityName:  role.EntityName,
				Action:      role.Action,
			}
		}
	}
	for _, v := range roleFilterMap {
		roleFilters = append(roleFilters, *v)
	}
	if len(roleFilters) == 0 {
		roleFilters = make([]bean2.RoleFilter, 0)
	}
	bean := &bean2.HelmRoleGroupDto{
		Id:          roleGroup.Id,
		Name:        roleGroup.Name,
		Description: roleGroup.Description,
		RoleFilters: roleFilters,
	}
	return bean, nil
}

func (impl HelmUserRoleGroupServiceImpl) FetchRoleGroups() ([]*bean2.HelmRoleGroupDto, error) {
	roleGroup, err := impl.roleGroupRepository.GetAllRoleGroup()
	if err != nil {
		impl.logger.Errorw("error while fetching user from db", "error", err)
		return nil, err
	}
	var list []*bean2.HelmRoleGroupDto
	for _, item := range roleGroup {
		bean := &bean2.HelmRoleGroupDto{
			Id:          item.Id,
			Name:        item.Name,
			Description: item.Description,
			RoleFilters: make([]bean2.RoleFilter, 0),
		}
		list = append(list, bean)
	}

	if len(list) == 0 {
		list = make([]*bean2.HelmRoleGroupDto, 0)
	}
	return list, nil
}
