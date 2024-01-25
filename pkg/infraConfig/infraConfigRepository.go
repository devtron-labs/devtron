package infraConfig

import (
	repository1 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
)

type InfraConfigRepository interface {
	GetIdentifierCountForDefaultProfile() (int, error)
	GetProfileByName(name string) (*InfraProfile, error)
	GetConfigurationsByProfileName(profileName string) ([]*InfraProfileConfiguration, error)
	GetConfigurationsByProfileId(profileId int) ([]*InfraProfileConfiguration, error)

	CreateProfile(tx *pg.Tx, infraProfile *InfraProfile) error
	CreateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfiguration) error

	UpdateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfiguration) error
	UpdateProfile(tx *pg.Tx, profileName string, profile *InfraProfile) error
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

// CreateProfile saves the default profile in the database only once in a lifetime.
// If the default profile already exists, it will not be saved again.
func (impl *InfraConfigRepositoryImpl) CreateProfile(tx *pg.Tx, infraProfile *InfraProfile) error {
	err := tx.Insert(infraProfile)
	return err
}

func (impl *InfraConfigRepositoryImpl) GetProfileByName(name string) (*InfraProfile, error) {
	infraProfile := &InfraProfile{}
	err := impl.dbConnection.Model(infraProfile).
		Where("name = ?", name).
		Where("active = ?", true).
		Select()
	return infraProfile, err
}

func (impl *InfraConfigRepositoryImpl) CreateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfiguration) error {
	err := tx.Insert(&configurations)
	return err
}

func (impl *InfraConfigRepositoryImpl) UpdateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfiguration) error {
	err := tx.Update(&configurations)
	return err
}

func (impl *InfraConfigRepositoryImpl) GetConfigurationsByProfileName(profileName string) ([]*InfraProfileConfiguration, error) {
	var configurations []*InfraProfileConfiguration
	err := impl.dbConnection.Model(&configurations).
		Where("infra_profile_id IN (SELECT id FROM infra_profile WHERE name = ? AND active = true)", profileName).
		Where("active = ?", true).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, errors.New(NO_PROPERTIES_FOUND)
	}
	return configurations, err
}

func (impl *InfraConfigRepositoryImpl) GetConfigurationsByProfileId(profileId int) ([]*InfraProfileConfiguration, error) {
	var configurations []*InfraProfileConfiguration
	err := impl.dbConnection.Model(&configurations).
		Where("profile_id = ?", profileId).
		Where("active = ?", true).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, errors.New(NO_PROPERTIES_FOUND)
	}
	return configurations, err
}

func (impl *InfraConfigRepositoryImpl) GetIdentifierCountForDefaultProfile() (int, error) {

	count, err := impl.dbConnection.Model(&repository1.App{}).
		Where("active = ?", true).
		Count()
	return count, err
}

func (impl *InfraConfigRepositoryImpl) UpdateProfile(tx *pg.Tx, profileName string, profile *InfraProfile) error {
	_, err := tx.Model(profile).
		Set("description=?", profile.Description).
		Set("updated_by=?", profile.UpdatedBy).
		Set("updated_on=?", profile.UpdatedOn).
		Where("name = ?", profileName).
		Where("active = ?", true).
		Update()
	return err
}
