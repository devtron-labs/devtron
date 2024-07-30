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

package infraConfig

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
)

type InfraConfigRepository interface {
	GetProfileByName(name string) (*InfraProfileEntity, error)
	GetConfigurationsByProfileName(profileName string) ([]*InfraProfileConfigurationEntity, error)
	GetConfigurationsByProfileId(profileId int) ([]*InfraProfileConfigurationEntity, error)

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
	err := tx.Insert(&configurations)
	return err
}

func (impl *InfraConfigRepositoryImpl) UpdateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfigurationEntity) error {
	var err error
	for _, configuration := range configurations {
		_, err = tx.Model(configuration).
			Set("value = ?", configuration.Value).
			Set("unit = ?", configuration.Unit).
			Set("updated_by = ?", configuration.UpdatedBy).
			Set("updated_on = ?", configuration.UpdatedOn).
			Where("id = ?", configuration.Id).
			Update()
	}
	return err
}

func (impl *InfraConfigRepositoryImpl) GetConfigurationsByProfileName(profileName string) ([]*InfraProfileConfigurationEntity, error) {
	var configurations []*InfraProfileConfigurationEntity
	err := impl.dbConnection.Model(&configurations).
		Where("profile_id IN (SELECT id FROM infra_profile WHERE name = ? AND active = true)", profileName).
		Where("active = ?", true).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, errors.New(NO_PROPERTIES_FOUND)
	}
	return configurations, err
}

func (impl *InfraConfigRepositoryImpl) GetConfigurationsByProfileId(profileId int) ([]*InfraProfileConfigurationEntity, error) {
	var configurations []*InfraProfileConfigurationEntity
	err := impl.dbConnection.Model(&configurations).
		Where("profile_id = ?", profileId).
		Where("active = ?", true).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, errors.New(NO_PROPERTIES_FOUND)
	}
	return configurations, err
}

func (impl *InfraConfigRepositoryImpl) UpdateProfile(tx *pg.Tx, profileName string, profile *InfraProfileEntity) error {
	_, err := tx.Model(profile).
		Set("description=?", profile.Description).
		Set("updated_by=?", profile.UpdatedBy).
		Set("updated_on=?", profile.UpdatedOn).
		Where("name = ?", profileName).
		Where("active = ?", true).
		Update()
	return err
}
