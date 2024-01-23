package repository

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/infraConfig"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
)

const DEFAULT_PROFILE_NAME = "default"
const DEFAULT_PROFILE_EXISTS = "default profile exists"
const noPropertiesFound = "no properties found"

type ProfileIdentifierCount struct {
	ProfileId       int
	IdentifierCount int
}

type InfraConfigRepository interface {
	GetProfileByName(name string) (*infraConfig.InfraProfile, error)
	GetProfileList(profileNameLike string) ([]*infraConfig.InfraProfile, error)

	GetConfigurationsByProfileId(profileIds []int) ([]*infraConfig.InfraProfileConfiguration, error)

	GetIdentifierCountForDefaultProfile(defaultProfileId int, identifierType int) (int, error)
	GetIdentifierCountForNonDefaultProfiles(profileIds []int, identifierType int) ([]ProfileIdentifierCount, error)

	CreateDefaultProfile(tx *pg.Tx, infraProfile *infraConfig.InfraProfile) error
	CreateProfile(tx *pg.Tx, infraProfile *infraConfig.InfraProfile) error
	CreateConfigurations(tx *pg.Tx, configurations []*infraConfig.InfraProfileConfiguration) error

	UpdateConfigurations(tx *pg.Tx, configurations []*infraConfig.InfraProfileConfiguration) error
	UpdateProfile(tx *pg.Tx, profileName string, profile *infraConfig.InfraProfile) error

	DeleteProfile(tx *pg.Tx, profileName string) error
	DeleteConfigurations(tx *pg.Tx, profileName string) error
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

func (impl *InfraConfigRepositoryImpl) CreateProfile(tx *pg.Tx, infraProfile *infraConfig.InfraProfile) error {
	err := tx.Insert(infraProfile)
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

func (impl *InfraConfigRepositoryImpl) GetProfileList(profileNameLike string) ([]*infraConfig.InfraProfile, error) {
	var infraProfiles []*infraConfig.InfraProfile
	query := impl.dbConnection.Model(&infraProfiles).
		Where("active = ?", true)
	if profileNameLike == "" {
		profileNameLike = "%" + profileNameLike + "%"
		query = query.Where("name LIKE ? OR name = ?", profileNameLike, DEFAULT_PROFILE_NAME)
	}
	err := query.Select()
	return infraProfiles, err
}

func (impl *InfraConfigRepositoryImpl) CreateConfigurations(tx *pg.Tx, configurations []*infraConfig.InfraProfileConfiguration) error {
	err := tx.Insert(&configurations)
	return err
}

func (impl *InfraConfigRepositoryImpl) UpdateConfigurations(tx *pg.Tx, configurations []*infraConfig.InfraProfileConfiguration) error {
	err := tx.Update(&configurations)
	return err
}

func (impl *InfraConfigRepositoryImpl) GetConfigurationsByProfileId(profileIds []int) ([]*infraConfig.InfraProfileConfiguration, error) {
	if len(profileIds) == 0 {
		return nil, errors.New("profileIds cannot be empty")
	}

	var configurations []*infraConfig.InfraProfileConfiguration
	err := impl.dbConnection.Model(&configurations).
		Where("infra_profile_id IN (?)", profileIds).
		Where("active = ?", true).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, errors.New(noPropertiesFound)
	}
	return configurations, err
}

func (impl *InfraConfigRepositoryImpl) GetIdentifierCountForDefaultProfile(defaultProfileId int, identifierKey int) (int, error) {
	queryToGetAppIdsWhichDoesntInheritDefaultConfigurations := " SELECT identifier_value_int " +
		" FROM resource_qualifier_mapping " +
		" WHERE reference_type = %d AND reference_id IN ( " +
		" 	SELECT profile_id " +
		"   FROM infra_profile_configuration " +
		"   GROUP BY profile_id HAVING COUNT(profile_id) = ( " +
		" 	      SELECT COUNT(id) " +
		" 	      FROM infra_profile_configuration " +
		" 	      WHERE active=true AND profile_id=%d ) " +
		" ) AND identifier_key = %d AND active=true"
	queryToGetAppIdsWhichDoesntInheritDefaultConfigurations = fmt.Sprintf(queryToGetAppIdsWhichDoesntInheritDefaultConfigurations, resourceQualifiers.InfraProfile, defaultProfileId, identifierKey)

	// exclude appIds which inherit default configurations
	query := " SELECT COUNT(id) " +
		" FROM app WHERE Id NOT IN (%s) and active=true"
	query = fmt.Sprintf(query, queryToGetAppIdsWhichDoesntInheritDefaultConfigurations)
	count := 0
	_, err := impl.dbConnection.Query(&count, query)
	return count, err
}

// GetIdentifierCountForNonDefaultProfiles returns the count of identifiers for the given profileIds and identifierType
// if resourceIds is empty, it will return the count of identifiers for all the profiles
func (impl *InfraConfigRepositoryImpl) GetIdentifierCountForNonDefaultProfiles(profileIds []int, identifierType int) ([]ProfileIdentifierCount, error) {
	query := " SELECT COUNT(DISTINCT identifier_id) as identifier_count, resource_id" +
		" FROM resource_qualifier_mapping " +
		" WHERE resource_type = ? AND identifier_type = ? AND active=true "
	if len(profileIds) > 0 {
		query += " AND resource_id IN (?) "
	}

	query += " GROUP BY resource_id"
	counts := make([]ProfileIdentifierCount, 0)
	_, err := impl.dbConnection.Query(&counts, query, resourceQualifiers.InfraProfile, identifierType)
	return counts, err
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

func (impl *InfraConfigRepositoryImpl) DeleteProfile(tx *pg.Tx, profileName string) error {
	_, err := tx.Model(&infraConfig.InfraProfile{}).
		Set("active=?", false).
		Where("name = ?", profileName).
		Update()
	return err
}

func (impl *InfraConfigRepositoryImpl) DeleteConfigurations(tx *pg.Tx, profileName string) error {
	_, err := tx.Model(&infraConfig.InfraProfileConfiguration{}).
		Set("active=?", false).
		Where("infra_profile_id IN (SELECT id FROM infra_profile WHERE name = ? AND active = true)", profileName).
		Update()
	return err
}
