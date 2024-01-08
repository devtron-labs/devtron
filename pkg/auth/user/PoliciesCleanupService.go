package user

import (
	"fmt"

	"github.com/caarlos0/env/v6"
	casbin2 "github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/go-pg/pg"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type CleanUpPoliciesService interface {
	CleanUpPolicies() (bool, error)
}

type CleanUpPoliciesServiceImpl struct {
	userAuthRepository        repository.UserAuthRepository
	logger                    *zap.SugaredLogger
	userRepository            repository.UserRepository
	roleGroupRepository       repository.RoleGroupRepository
	cleanUpPoliciesRepository repository.PoliciesCleanUpRepository
}

func NewCleanUpPoliciesServiceImpl(userAuthRepository repository.UserAuthRepository,
	logger *zap.SugaredLogger, userRepository repository.UserRepository,
	roleGroupRepository repository.RoleGroupRepository,
	cleanUpPoliciesRepository repository.PoliciesCleanUpRepository) *CleanUpPoliciesServiceImpl {
	cleanUpPoliciesCronServiceImpl := &CleanUpPoliciesServiceImpl{
		logger:                    logger,
		userRepository:            userRepository,
		userAuthRepository:        userAuthRepository,
		roleGroupRepository:       roleGroupRepository,
		cleanUpPoliciesRepository: cleanUpPoliciesRepository,
	}
	cfg := &CleanUpPoliciesServiceConfig{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("failed to parse server cluster status config: " + err.Error())
	}
	if cfg.CleanUpPoliciesRun {
		cron := cron.New(cron.WithChain())
		cron.Start()

		_, err = cron.AddFunc(cfg.CleanUpPoliciesCronTime, cleanUpPoliciesCronServiceImpl.CleanUpPoliciesForCron)
		if err != nil {
			fmt.Println("error in adding cron function into CleanUpPolicies Cron")
			return nil
		}
	}
	return cleanUpPoliciesCronServiceImpl
}

type CleanUpPoliciesServiceConfig struct {
	CleanUpPoliciesRun      bool   `env:"CLEAN_UP_RBAC_POLICIES" envDefault:"false"`
	CleanUpPoliciesCronTime string `env:"CLEAN_UP_RBAC_POLICIES_CRON_TIME" envDefault:"0 0 * * *"`
}

func (impl *CleanUpPoliciesServiceImpl) CleanUpPoliciesForCron() {
	_, err := impl.CleanUpPolicies()
	if err != nil {
		impl.logger.Errorw("Failed in running cron for clean up rbac Policies")
	}
}
func (impl *CleanUpPoliciesServiceImpl) cleanUpDuplicateRolesFromOrchestrator(tx *pg.Tx) error {
	impl.logger.Infow("deleting duplicate mappings for all users from orchestrator")
	err := impl.cleanUpPoliciesRepository.DeleteDuplicateMappingForAllUsers(tx)
	if err != nil {
		impl.logger.Errorw("error in  deleting duplicate role mapping for users", "err", err)
		return err
	}
	impl.logger.Infow("deleted duplicate mappings for all users from orchestrator")
	impl.logger.Infow("deleting duplicate mappings for all groups from orchestrator")
	err = impl.cleanUpPoliciesRepository.DeleteDuplicateMappingForAllGroups(tx)
	if err != nil {
		impl.logger.Errorw("error in  deleting duplicate role mapping for groups", "err", err)
		return err
	}
	impl.logger.Infow("deleted duplicate mappings for all groups from orchestrator")
	impl.logger.Infow("updating duplicate role mappings for all users from orchestrator")
	err = impl.cleanUpPoliciesRepository.UpdateDuplicateAndChangeRoleMappingForUsers(tx)
	if err != nil {
		impl.logger.Errorw("error in updating roles for duplicate roles", "err", err)
		return err
	}
	impl.logger.Infow("updated duplicate role mappings for all users from orchestrator")
	impl.logger.Infow("updating duplicate role mappings for all groups from orchestrator")
	err = impl.cleanUpPoliciesRepository.UpdateDuplicateAndChangeRoleMappingForGroup(tx)
	if err != nil {
		impl.logger.Errorw("error in  updating roles for duplicate group roles", "err", err)
		return err
	}
	impl.logger.Infow("updated duplicate role mappings for all groups from orchestrator")
	impl.logger.Infow("deleting user roles mappings for all users from orchestrator")
	err = impl.cleanUpPoliciesRepository.DeleteRoleGroupRoleMappingforInactiveUsers(tx)
	if err != nil {
		impl.logger.Errorw("error in deleting user roles mappings for deleted users", "err", err)
		return err
	}
	impl.logger.Infow("deleted user roles mappings for all users  for inactive users")
	impl.logger.Infow("deleting role group role mappings for all groups from orchestrator")
	err = impl.cleanUpPoliciesRepository.DeleteRoleGroupRoleMappingforInactiveGroups(tx)
	if err != nil {
		impl.logger.Errorw("error in deleting role group role mappings for deleted rolegroups", "err", err)
		return err
	}
	impl.logger.Infow("deleted role group roleMapping for inactive groups")
	impl.logger.Infow("Committing transaction")
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in committing transaction", "err", err)
		return err
	}
	impl.logger.Infow("committed transaction")
	return nil
}

func (impl *CleanUpPoliciesServiceImpl) cleanUpUnusedRolesFromCasbin() error {
	impl.logger.Infow("Loading Policies from casbin")
	casbin2.LoadPolicy()
	impl.logger.Infow("Loaded Policies from casbin")
	impl.logger.Infow("getting all unused roles for casbin clean up")
	rolesToBeDeleted, err := impl.cleanUpPoliciesRepository.GetAllUnusedRolesForCasbinCleanUp()
	if err != nil {
		impl.logger.Errorw("error in getting unused roles for casbin clean Up", "err", err)
		return err
	}
	impl.logger.Infow("Got all unused roles for casbin clean up")
	impl.logger.Infow("Now Removing it from casbin through remove policies by roles", "count of policies to be deleted from casbin", len(rolesToBeDeleted))
	flag, err := casbin2.RemovePoliciesByRoles(rolesToBeDeleted)
	impl.logger.Infow("printing flag for casbin", "flag", flag)
	impl.logger.Infow("removed from casbin")
	if err != nil {
		impl.logger.Warnw("error in deleting casbin policy for role", "roles", rolesToBeDeleted)
		return err
	}
	impl.logger.Infow("Loading Policies from casbin")
	casbin2.LoadPolicy()
	impl.logger.Infow("Loaded Policies from casbin")
	return nil
}
func (impl *CleanUpPoliciesServiceImpl) cleanUpUnusedRolesFromOrchestrator() error {
	impl.logger.Infow("getting all user mapped roles from orchestrator")
	allUserMappedRoles, err := impl.cleanUpPoliciesRepository.GetAllUserMappedRoles()
	if err != nil {
		impl.logger.Errorw("error in getting all user mapped Roles", "err", err)
		return err
	}
	impl.logger.Infow("Got all user mapped roles, now appending role id in activeRoleIds")
	activeRoleIds := make([]int32, 0, len(allUserMappedRoles))
	for _, role := range allUserMappedRoles {
		activeRoleIds = append(activeRoleIds, int32(role.RoleId))
	}
	impl.logger.Infow("Getting all role Groups from orchestrator")
	activeGroups, err := impl.roleGroupRepository.GetAllRoleGroup()
	if err != nil {
		impl.logger.Errorw("error in getting active role Groups", "err", err)
		return err
	}
	impl.logger.Infow("Got all role Groups from orchestrator")
	impl.logger.Infow("Appending all active roleGroups Ids")
	activeGroupsIds := make([]int32, 0, len(activeGroups))
	for _, group := range activeGroups {
		activeGroupsIds = append(activeGroupsIds, group.Id)
	}
	impl.logger.Infow("Getting all roles for active groups")
	rolesForActiveGroups, err := impl.cleanUpPoliciesRepository.GetAllRolesForActiveGroups(activeGroupsIds)
	if err != nil {
		impl.logger.Errorw("error in getting roles for active groups", "err", err)
		return err
	}
	impl.logger.Infow("Got all roles for active groups")
	activeGroupsRoleIds := make([]int32, 0, len(rolesForActiveGroups))
	for _, group := range rolesForActiveGroups {
		activeGroupsRoleIds = append(activeGroupsRoleIds, int32(group.RoleId))
	}
	impl.logger.Infow("Appending both active role id for users and groups")
	idsToKeep := append(activeRoleIds, activeGroupsRoleIds...)

	impl.logger.Infow("deleting all roles except active from orchestrator")
	err = impl.cleanUpPoliciesRepository.DeleteAllRolesExceptActive(idsToKeep)
	if err != nil {
		impl.logger.Errorw("error in deleting all role groups except active")
		return err
	}
	impl.logger.Infow("deleted all roles except active from orchestrator")
	return nil
}

func (impl *CleanUpPoliciesServiceImpl) CleanUpPolicies() (bool, error) {
	dbConnection := impl.cleanUpPoliciesRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error in establishing connection", "err", err)
		return false, err
	}
	impl.logger.Infow("Made a transaction for clean policies in orchestrator")
	// Rollback tx on error.
	defer tx.Rollback()

	impl.logger.Infow("cleaning up duplicate roles from orchestrator")
	err = impl.cleanUpDuplicateRolesFromOrchestrator(tx)
	if err != nil {
		impl.logger.Errorw("error in deleting and updating in orchestrator", "err", err)
		return false, err
	}
	impl.logger.Infow("Cleaned up duplicate roles from orchestrator")
	impl.logger.Infow("Starting Cleaning up Unused roles from casbin now")
	err = impl.cleanUpUnusedRolesFromCasbin()
	if err != nil {
		impl.logger.Errorw("error in deleting unused roles from casbin", "err", err)
		return false, err
	}
	impl.logger.Infow("Cleaned up Unused roles from casbin now")
	impl.logger.Infow("Cleaning up Unused roles from orchestrator now")
	err = impl.cleanUpUnusedRolesFromOrchestrator()
	if err != nil {
		impl.logger.Errorw("error in deleting unused roles from orchestrator", "err", err)
		return false, err
	}
	impl.logger.Infow("Cleaned up Unused roles from orchestrator now")
	impl.logger.Infow("done cleaning up everything")
	return true, nil
}
