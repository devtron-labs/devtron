package service

import (
	v1 "github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/go-pg/pg"
)

type InfraConfigServiceEnt interface {
}

func (impl *InfraConfigServiceImpl) isMigrationRequired() (bool, error) {
	return true, nil
}

func (impl *InfraConfigServiceImpl) markMigrationComplete(tx *pg.Tx) error {
	return nil
}

func (impl *InfraConfigServiceImpl) resolveScopeVariablesForAppliedProfile(scope resourceQualifiers.Scope, appliedProfileConfig *v1.ProfileBeanDto) (*v1.ProfileBeanDto, map[string]map[string]string, error) {
	return appliedProfileConfig, nil, nil
}

func (impl *InfraConfigServiceImpl) updateBuildxDriverTypeInExistingProfiles(tx *pg.Tx) error {
	return nil
}

func (impl *InfraConfigServiceImpl) getCreatableK8sDriverConfigs(profileId int, envConfigs []*repository.InfraProfileConfigurationEntity) ([]*repository.InfraProfileConfigurationEntity, error) {
	return make([]*repository.InfraProfileConfigurationEntity, 0), nil
}

func (impl *InfraConfigServiceImpl) getDefaultBuildxDriverType() v1.BuildxDriver {
	return v1.BuildxDockerContainerDriver
}

func (impl *InfraConfigServiceImpl) getInfraProfilesByScope(scope *v1.Scope, includeDefault bool) ([]*repository.InfraProfileEntity, []int, error) {
	// for OSS, user can't create infra profiles so no need to fetch infra profiles
	profileIds := make([]int, 0)
	infraProfilesEntities, err := impl.infraProfileRepo.GetProfileListByIds(profileIds, includeDefault)
	if err != nil {
		impl.logger.Errorw("error in fetching profile entities by ids", "scope", scope, "profileIds", profileIds, "error", err)
		return nil, profileIds, err
	}
	return infraProfilesEntities, profileIds, err
}
