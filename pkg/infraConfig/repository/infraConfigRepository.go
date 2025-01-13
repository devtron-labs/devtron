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
	infraBean "github.com/devtron-labs/devtron/pkg/infraConfig/bean"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"time"
)

type InfraProfileEntity struct {
	tableName        struct{}               `sql:"infra_profile" pg:",discard_unknown_columns"`
	Id               int                    `sql:"id"`
	Name             string                 `sql:"name"`
	Description      string                 `sql:"description"`
	BuildxDriverType infraBean.BuildxDriver `sql:"buildx_driver_type,notnull"`
	Active           bool                   `sql:"active"`
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
	tableName                struct{}            `sql:"infra_profile_configuration" pg:",discard_unknown_columns"`
	Id                       int                 `sql:"id,pk"`
	Key                      infraBean.ConfigKey `sql:"key,notnull"`
	Value                    float64             `sql:"value"`
	ValueString              string              `sql:"value_string,notnull"`
	Unit                     units.UnitSuffix    `sql:"unit,notnull"`
	ProfilePlatformMappingId int                 `sql:"profile_platform_mapping_id"`
	Active                   bool                `sql:"active,notnull"`
	UniqueId                 string              `sql:"-"`
	// Deprecated; use ProfilePlatformMappingId
	ProfileId int `sql:"profile_id"`

	ProfilePlatformMapping *ProfilePlatformMapping
	sql.AuditLog
}

type InfraConfigRepository interface {
	GetProfileByName(name string) (*InfraProfileEntity, error)
	GetConfigurationsByProfileName(profileName string) ([]*InfraProfileConfigurationEntity, error)
	GetConfigurationsByProfileIds(profileIds []int) ([]*InfraProfileConfigurationEntity, error)

	GetPlatformListByProfileName(profileName string) ([]string, error)
	CreatePlatformProfileMapping(tx *pg.Tx, platformMapping []*ProfilePlatformMapping) error

	CreateProfile(tx *pg.Tx, infraProfile *InfraProfileEntity) error
	CreateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfigurationEntity) error

	UpdateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfigurationEntity) error
	UpdateProfile(tx *pg.Tx, profileName string, profile *InfraProfileEntity) error
	sql.TransactionWrapper
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

func (impl *InfraConfigRepositoryImpl) GetProfileByName(name string) (*InfraProfileEntity, error) {
	infraProfile := &InfraProfileEntity{}
	err := impl.dbConnection.Model(infraProfile).
		Where("name = ?", name).
		Where("active = ?", true).
		Select()
	return infraProfile, err
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
		return nil, errors.New(infraBean.NO_PROPERTIES_FOUND)
	}
	return configurations, err
}

func (impl *InfraConfigRepositoryImpl) GetConfigurationsByProfileIds(profileIds []int) ([]*InfraProfileConfigurationEntity, error) {
	var configurations []*InfraProfileConfigurationEntity
	err := impl.dbConnection.Model(&configurations).
		Column("infra_profile_configuration_entity.*", "ProfilePlatformMapping").
		Where("profile_platform_mapping.profile_id IN (?)", pg.In(profileIds)).
		Where("profile_platform_mapping.active = ?", true).
		Where("infra_profile_configuration_entity.active = ?", true).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, errors.New(infraBean.NO_PROPERTIES_FOUND)
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
		Where("active = ?", true).
		Update()
	return err
}

func (impl *InfraConfigRepositoryImpl) UpdateBuildxDriverTypeInAllProfiles(tx *pg.Tx, buildxDriverType infraBean.BuildxDriver) error {
	_, err := tx.Model((*InfraProfileEntity)(nil)).
		Set("buildx_driver_type = ?", buildxDriverType).
		Set("updated_by = ?", 1).
		Set("updated_on = ?", time.Now()).
		Where("active = ?", true).
		Update()
	return err
}

func (impl *InfraConfigRepositoryImpl) GetPlatformListByProfileName(profileName string) ([]string, error) {
	var platforms []string
	err := impl.dbConnection.Model(&ProfilePlatformMapping{}).
		ColumnExpr("platform").
		Join("INNER JOIN infra_profile ip ON profile_platform_mapping.profile_id = ip.id").
		Where("ip.name = ?", profileName).
		Where("ip.active = ?", true).
		Where("profile_platform_mapping.active = ?", true).
		Select(&platforms)
	return platforms, err
}
func (impl *InfraConfigRepositoryImpl) CreatePlatformProfileMapping(tx *pg.Tx, platformMapping []*ProfilePlatformMapping) error {
	err := tx.Insert(&platformMapping)
	return err
}
