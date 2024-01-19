package repository

import (
	"github.com/devtron-labs/devtron/pkg/infraConfig"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
)

const DEFAULT_PROFILE_NAME = "default"
const DEFAULT_PROFILE_EXISTS = "default profile exists"
const noPropertiesFound = "no properties found"

type InfraConfigRepository interface {
	CreateDefaultProfile(tx *pg.Tx, infraProfile *infraConfig.InfraProfile) error
	GetProfileByName(name string) (*infraConfig.InfraProfile, error)
	GetConfigurationsByProfileId(profileId int) ([]*infraConfig.InfraProfileConfiguration, error)
	CreateConfigurations(tx *pg.Tx, configurations []*infraConfig.InfraProfileConfiguration) error
	UpdateConfigurations(tx *pg.Tx, configurations []*infraConfig.InfraProfileConfiguration) error
	GetIdentifierCountForDefaultProfile(defaultProfileId int) (int, error)
	UpdateProfile(tx *pg.Tx, profileName string, profile *infraConfig.InfraProfile) error
	sql.TransactionWrapper
}

type InfraConfigRepositoryImpl struct {
	dbConnection *pg.DB
	*sql.TransactionUtilImpl
}

func NewInfraProfileRepositoryImpl(dbConnection *pg.DB) *InfraConfigRepositoryImpl {
	return &InfraConfigRepositoryImpl{
		dbConnection:        dbConnection,
		TransactionUtilImpl: sql.NewTransactionUtilImpl(dbConnection),
	}
}

// CreateDefaultProfile saves the default profile in the database only once in a lifetime.
// If the default profile already exists, it will not be saved again.
func (impl *InfraConfigRepositoryImpl) CreateDefaultProfile(tx *pg.Tx, infraProfile *infraConfig.InfraProfile) error {
	profile, err := impl.GetProfileByName(DEFAULT_PROFILE_NAME)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		return err
	}
	if profile != nil {
		return errors.New(DEFAULT_PROFILE_EXISTS)
	}
	err = tx.Insert(infraProfile)
	return err
}

func (impl *InfraConfigRepositoryImpl) GetProfileByName(name string) (*infraConfig.InfraProfile, error) {
	var infraProfile infraConfig.InfraProfile
	err := impl.dbConnection.Model(&infraProfile).
		Where("name = ?", name).
		Where("active = ?", true).
		Select()
	return &infraProfile, err
}

func (impl *InfraConfigRepositoryImpl) CreateConfigurations(tx *pg.Tx, configurations []*infraConfig.InfraProfileConfiguration) error {
	err := tx.Insert(&configurations)
	return err
}

func (impl *InfraConfigRepositoryImpl) UpdateConfigurations(tx *pg.Tx, configurations []*infraConfig.InfraProfileConfiguration) error {
	err := tx.Update(&configurations)
	return err
}

func (impl *InfraConfigRepositoryImpl) GetConfigurationsByProfileId(profileId int) ([]*infraConfig.InfraProfileConfiguration, error) {
	var configurations []*infraConfig.InfraProfileConfiguration
	err := impl.dbConnection.Model(&configurations).
		Where("infra_profile_id = ?", profileId).
		Where("active = ?", true).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, errors.New(noPropertiesFound)
	}
	return configurations, err
}

func (impl *InfraConfigRepositoryImpl) GetIdentifierCountForDefaultProfile(defaultProfileId int) (int, error) {
	query := " SELECT COUNT(DISTINCT app_id) " +
		" FROM resource_identifier_mapping " +
		" WHERE reference_type = ? AND reference_id IN ( " +
		" 	SELECT profile_id " +
		"   FROM infra_profile_configuration " +
		"   GROUP BY profile_id HAVING COUNT(profile_id) < ( " +
		" 	      SELECT COUNT(id) " +
		" 	      FROM infra_profile_configuration " +
		" 	      WHERE active=true AND profile_id=?) " +
		" ) AND active=true"

	count := 0
	_, err := impl.dbConnection.Query(&count, query, resourceQualifiers.InfraProfile, defaultProfileId)
	return count, err
}

func (impl *InfraConfigRepositoryImpl) UpdateProfile(tx *pg.Tx, profileName string, profile *infraConfig.InfraProfile) error {
	_, err := tx.Model(&infraConfig.InfraProfile{}).
		Set("name=?", profile.Name).
		Set("description=?", profile.Description).
		Set("updated_by=?", profile.UpdatedBy).
		Set("updated_on=?", profile.UpdatedOn).
		Where("name = ?", profileName).
		Update(profile)
	return err
}
