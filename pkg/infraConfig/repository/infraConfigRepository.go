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

package repository

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	errors2 "github.com/devtron-labs/devtron/pkg/infraConfig/errors"
	unitsBean "github.com/devtron-labs/devtron/pkg/infraConfig/units/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/pkg/errors"
)

type ProfileIdentifierCount struct {
	ProfileId       int
	IdentifierCount int
}

type InfraProfileEntity struct {
	tableName        struct{}        `sql:"infra_profile" pg:",discard_unknown_columns"`
	Id               int             `sql:"id"`
	Name             string          `sql:"name"`
	Description      string          `sql:"description"`
	BuildxDriverType v1.BuildxDriver `sql:"buildx_driver_type,notnull"`
	Active           bool            `sql:"active"`
	sql.AuditLog
}

type ProfilePlatformMapping struct {
	tableName struct{} `sql:"profile_platform_mapping" pg:",discard_unknown_columns"`
	Id        int      `sql:"id"`
	ProfileId int      `sql:"profile_id"`
	Platform  string   `sql:"platform"`
	Active    bool     `sql:"active"`
	UniqueId  string   `sql:"-"`
	sql.AuditLog
}

func GetUniqueId(profileId int, platform string) string {
	return fmt.Sprintf("%d-%s", profileId, platform)
}

type InfraProfileConfigurationEntity struct {
	tableName struct{}     `sql:"infra_profile_configuration" pg:",discard_unknown_columns"`
	Id        int          `sql:"id,pk"`
	Key       v1.ConfigKey `sql:"key,notnull"`
	// Deprecated; use ValueString instead
	Value                    float64            `sql:"value"`
	ValueString              string             `sql:"value_string,notnull"`
	Unit                     unitsBean.UnitType `sql:"unit,notnull"`
	ProfilePlatformMappingId int                `sql:"profile_platform_mapping_id"`
	Active                   bool               `sql:"active,notnull"`
	UniqueId                 string             `sql:"-"`
	// Deprecated; use ProfilePlatformMappingId
	ProfileId int `sql:"profile_id"`

	ProfilePlatformMapping *ProfilePlatformMapping
	sql.AuditLog
}

type InfraConfigRepository interface {
	GetProfileByName(name string) (*InfraProfileEntity, error)
	CheckIfProfileExistsByName(name string) (bool, error)
	GetConfigurationsByProfileName(profileName string) ([]*InfraProfileConfigurationEntity, error)
	GetConfigurationsByProfileId(profileId int) ([]*InfraProfileConfigurationEntity, error)

	CreatePlatformProfileMapping(tx *pg.Tx, platformMapping []*ProfilePlatformMapping) error

	CreateProfile(tx *pg.Tx, infraProfile *InfraProfileEntity) error
	CreateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfigurationEntity) error

	UpdateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfigurationEntity) error
	UpdateProfile(tx *pg.Tx, profileName string, profile *InfraProfileEntity) error

	// GetProfileListByIds returns the list of profiles for the given profileIds
	// includeDefault is used to explicitly include the default profile in the list
	GetProfileListByIds(profileIds []int, includeDefault bool) ([]*InfraProfileEntity, error)
	GetConfigurationsByProfileIds(profileIds []int) ([]*InfraProfileConfigurationEntity, error)
	UpdatePlatformProfileMapping(tx *pg.Tx, platformMappings []*ProfilePlatformMapping) error
	GetPlatformListByProfileId(profileId int) ([]string, error)
	GetPlatformsByProfileName(profileName string) ([]*ProfilePlatformMapping, error)
	GetPlatformsByProfileIds(profileIds []int) ([]*ProfilePlatformMapping, error)
	GetPlatformsByProfileById(profileId int) ([]*ProfilePlatformMapping, error)
	sql.TransactionWrapper
	InfraConfigRepositoryEnt
}

type InfraConfigRepositoryImpl struct {
	dbConnection *pg.DB
	*sql.TransactionUtilImpl
}

func NewInfraProfileRepositoryImpl(dbConnection *pg.DB, TransactionUtilImpl *sql.TransactionUtilImpl) *InfraConfigRepositoryImpl {
	return &InfraConfigRepositoryImpl{
		dbConnection:        dbConnection,
		TransactionUtilImpl: TransactionUtilImpl,
	}
}

// CreateProfile saves the default profile in the database only once in a lifetime.
// If the default profile already exists, it will not be saved again.
func (impl *InfraConfigRepositoryImpl) CreateProfile(tx *pg.Tx, infraProfile *InfraProfileEntity) error {
	err := tx.Insert(infraProfile)
	return err
}

func (impl *InfraConfigRepositoryImpl) CreatePlatformProfileMapping(tx *pg.Tx, platformMapping []*ProfilePlatformMapping) error {
	err := tx.Insert(&platformMapping)
	return err
}

func (impl *InfraConfigRepositoryImpl) GetProfileByName(name string) (*InfraProfileEntity, error) {
	infraProfile := &InfraProfileEntity{}
	err := impl.dbConnection.Model(infraProfile).
		Where("name = ?", name).
		Where("active = ?", true).
		Select()
	return infraProfile, err
}

func (impl *InfraConfigRepositoryImpl) CheckIfProfileExistsByName(name string) (bool, error) {
	infraProfile := &InfraProfileEntity{}
	return impl.dbConnection.Model(infraProfile).Where("name = ?", name).Where("active =?", true).Exists()
}

func (impl *InfraConfigRepositoryImpl) CreateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfigurationEntity) error {
	if len(configurations) == 0 {
		return nil
	}
	err := tx.Insert(&configurations)
	return err
}

func (impl *InfraConfigRepositoryImpl) UpdateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfigurationEntity) error {
	_, err := tx.Model(&configurations).
		Column("value_string", "profile_platform_mapping_id", "unit", "active", "updated_by", "updated_on").
		Update()
	return err
}

func (impl *InfraConfigRepositoryImpl) GetConfigurationsByProfileName(profileName string) ([]*InfraProfileConfigurationEntity, error) {
	var configurations []*InfraProfileConfigurationEntity
	err := impl.dbConnection.Model(&configurations).
		Column("infra_profile_configuration_entity.*", "ProfilePlatformMapping").
		Join("INNER JOIN infra_profile").
		JoinOn("infra_profile_configuration_entity.profile_id = infra_profile.id").
		Where("infra_profile.name = ?", profileName).
		Where("infra_profile.active = ?", true).
		Where("infra_profile_configuration_entity.active = ?", true).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, errors2.NoPropertiesFoundError
	}
	return configurations, err
}

func (impl *InfraConfigRepositoryImpl) GetConfigurationsByProfileId(profileId int) ([]*InfraProfileConfigurationEntity, error) {
	var configurations []*InfraProfileConfigurationEntity
	err := impl.dbConnection.Model(&configurations).
		Column("infra_profile_configuration_entity.*", "ProfilePlatformMapping").
		Join("INNER JOIN infra_profile").
		JoinOn("infra_profile_configuration_entity.profile_id = infra_profile.id").
		Where("infra_profile.id = ?", profileId).
		Where("infra_profile.active = ?", true).
		Where("infra_profile_configuration_entity.active = ?", true).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, errors2.NoPropertiesFoundError
	}
	return configurations, err
}

func (impl *InfraConfigRepositoryImpl) GetConfigurationsByProfileIds(profileIds []int) ([]*InfraProfileConfigurationEntity, error) {
	if len(profileIds) == 0 {
		return nil, errors2.ProfileIdsRequired
	}
	var configurations []*InfraProfileConfigurationEntity
	err := impl.dbConnection.Model(&configurations).
		Column("infra_profile_configuration_entity.*", "ProfilePlatformMapping").
		Where("profile_platform_mapping.profile_id IN (?)", pg.In(profileIds)).
		Where("profile_platform_mapping.active = ?", true).
		Where("infra_profile_configuration_entity.active = ?", true).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, errors2.NoPropertiesFoundError
	}
	return configurations, err
}

func (impl *InfraConfigRepositoryImpl) UpdateProfile(tx *pg.Tx, profileName string, profile *InfraProfileEntity) error {
	_, err := tx.Model(profile).
		Set("name = ?", profile.Name).
		Set("description = ?", profile.Description).
		Set("buildx_driver_type = ?", profile.BuildxDriverType).
		Set("updated_by = ?", profile.UpdatedBy).
		Set("updated_on = ?", profile.UpdatedOn).
		Where("name = ?", profileName).
		Update()
	return err
}

func (impl *InfraConfigRepositoryImpl) GetProfileListByIds(profileIds []int, includeDefault bool) ([]*InfraProfileEntity, error) {
	var infraProfiles []*InfraProfileEntity
	err := impl.dbConnection.Model(&infraProfiles).
		Where("active = ?", true).
		WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			if len(profileIds) != 0 {
				q = q.WhereOr("id IN (?)", pg.In(profileIds))
			}
			if includeDefault {
				q = q.WhereOr("name = ?", v1.GLOBAL_PROFILE_NAME)
			}
			return q, nil
		}).Select()
	return infraProfiles, err
}

func (impl *InfraConfigRepositoryImpl) GetPlatformListByProfileId(profileId int) ([]string, error) {
	var platforms []string
	err := impl.dbConnection.Model(&ProfilePlatformMapping{}).
		Column("platform").
		Where("profile_id = ?", profileId).
		Where("active = ?", true).
		Select(&platforms)
	return platforms, err
}

func (impl *InfraConfigRepositoryImpl) GetPlatformsByProfileIds(profileIds []int) ([]*ProfilePlatformMapping, error) {
	var profilePlatformMappings []*ProfilePlatformMapping
	err := impl.dbConnection.Model(&profilePlatformMappings).
		Where("active = ?", true).
		WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			if len(profileIds) > 0 {
				q = q.WhereOr("profile_id IN (?)", pg.In(profileIds))
			}
			return q, nil
		}).Select()
	return profilePlatformMappings, err
}

func (impl *InfraConfigRepositoryImpl) GetPlatformsByProfileName(profileName string) ([]*ProfilePlatformMapping, error) {
	var profilePlatformMappings []*ProfilePlatformMapping
	err := impl.dbConnection.Model(&profilePlatformMappings).
		Column("profile_platform_mapping.*").
		Join("INNER JOIN infra_profile ip ON profile_platform_mapping.profile_id = ip.id").
		Where("ip.name = ?", profileName).
		Where("ip.active = ?", true).
		Where("profile_platform_mapping.active = ?", true).
		Select()
	return profilePlatformMappings, err
}

func (impl *InfraConfigRepositoryImpl) UpdatePlatformProfileMapping(tx *pg.Tx, platformMappings []*ProfilePlatformMapping) error {
	var err error
	for _, platformMapping := range platformMappings {
		_, err = tx.Model(platformMapping).
			Set("platform = ?", platformMapping.Platform).
			Set("active = ?", platformMapping.Active).
			Set("updated_by = ?", platformMapping.UpdatedBy).
			Set("updated_on = ?", platformMapping.UpdatedOn).
			Where("id = ?", platformMapping.Id).
			Update()
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl *InfraConfigRepositoryImpl) GetPlatformsByProfileById(profileId int) ([]*ProfilePlatformMapping, error) {
	var profilePlatformMappings []*ProfilePlatformMapping
	err := impl.dbConnection.Model(&profilePlatformMappings).
		Column("profile_platform_mapping.*").
		Join("INNER JOIN infra_profile ip ON profile_platform_mapping.profile_id = ip.id").
		Where("ip.id = ?", profileId).
		Where("ip.active = ?", true).
		Where("profile_platform_mapping.active = ?", true).
		Select()
	return profilePlatformMappings, err
}
