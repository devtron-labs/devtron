/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	v1 "github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
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

func (impl *InfraConfigServiceImpl) getInfraProfileIdsByScope(scope *v1.Scope) ([]int, error) {
	// First check if there's a pipeline-level infra profile
	if scope.PipelineId > 0 {
		pipelineInfraProfileId, err := impl.getPipelineInfraProfileId(scope.PipelineId)
		if err != nil {
			impl.logger.Errorw("error in fetching pipeline-level infra profile", "pipelineId", scope.PipelineId, "error", err)
			return make([]int, 0), err
		}
		if pipelineInfraProfileId != nil {
			impl.logger.Debugw("found pipeline-level infra profile", "pipelineId", scope.PipelineId, "infraProfileId", *pipelineInfraProfileId)
			return []int{*pipelineInfraProfileId}, nil
		}
	}
	
	// For OSS, user can't create infra profiles so no need to fetch infra profiles
	// Falls back to global profile
	return make([]int, 0), nil
}

func (impl *InfraConfigServiceImpl) getPipelineInfraProfileId(pipelineId int) (*int, error) {
	// Query the ci_pipeline table directly for the infra_profile_id
	var infraProfileId *int
	query := `SELECT infra_profile_id FROM ci_pipeline WHERE id = ? AND deleted = false`
	_, err := impl.infraProfileRepo.GetDbConnection().Query(&infraProfileId, query, pipelineId)
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return nil, nil // Pipeline not found, fall back to app-level profile
		}
		return nil, err
	}
	return infraProfileId, nil
}
