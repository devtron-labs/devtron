package repository

import (
	"github.com/devtron-labs/devtron/pkg/infraProfiles"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
)

const DEFAULT_PROFILE_NAME = "default"
const DEFAULT_PROFILE_EXISTS = "default profile exists"

type InfraProfileRepository interface {
	CreateDefaultProfile(tx *pg.Tx, infraProfile *infraProfiles.InfraProfile) error
	GetDefaultProfile() (*infraProfiles.InfraProfile, error)
	CreateConfigurations(tx *pg.Tx, configurations []*infraProfiles.InfraProfileConfiguration) error
	sql.TransactionWrapper
}

type InfraProfileRepositoryImpl struct {
	dbConnection *pg.DB
	*sql.TransactionUtilImpl
}

func NewInfraProfileRepositoryImpl(dbConnection *pg.DB) *InfraProfileRepositoryImpl {
	return &InfraProfileRepositoryImpl{
		dbConnection:        dbConnection,
		TransactionUtilImpl: sql.NewTransactionUtilImpl(dbConnection),
	}
}

// CreateDefaultProfile saves the default profile in the database only once in a lifetime.
// If the default profile already exists, it will not be saved again.
func (impl InfraProfileRepositoryImpl) CreateDefaultProfile(tx *pg.Tx, infraProfile *infraProfiles.InfraProfile) error {
	profile, err := impl.GetDefaultProfile()
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		return err
	}
	if profile != nil {
		return errors.New(DEFAULT_PROFILE_EXISTS)
	}
	err = tx.Insert(infraProfile)
	return err
}

func (impl InfraProfileRepositoryImpl) GetDefaultProfile() (*infraProfiles.InfraProfile, error) {
	var infraProfile infraProfiles.InfraProfile
	err := impl.dbConnection.Model(&infraProfile).
		Where("name = ?", DEFAULT_PROFILE_NAME).
		Where("active = ?", true).
		Select()
	return &infraProfile, err
}

func (impl InfraProfileRepositoryImpl) CreateConfigurations(tx *pg.Tx, configurations []*infraProfiles.InfraProfileConfiguration) error {
	err := tx.Insert(&configurations)
	return err
}
