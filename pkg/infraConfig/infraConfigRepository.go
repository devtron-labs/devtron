package infraConfig

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/pkg/errors"
)

type ProfileIdentifierCount struct {
	ProfileId       int
	IdentifierCount int
}

type InfraConfigRepository interface {
	GetProfileIdByName(name string) (int, error)
	GetProfileByName(name string) (*InfraProfileEntity, error)
	GetProfileList(profileNameLike string) ([]*InfraProfileEntity, error)
	GetActiveProfileNames() ([]string, error)

	// GetProfileListByIds returns the list of profiles for the given profileIds
	// includeDefault is used to explicitly include the default profile in the list
	GetProfileListByIds(profileIds []int, includeDefault bool) ([]*InfraProfileEntity, error)
	GetConfigurationsByProfileName(profileName string) ([]*InfraProfileConfigurationEntity, error)
	GetConfigurationsByProfileIds(profileIds []int) ([]*InfraProfileConfigurationEntity, error)
	GetConfigurationsByScope(scope Scope, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) ([]*InfraProfileConfigurationEntity, error)

	// GetIdentifierCountForDefaultProfile(defaultProfileId int, identifierType int) (int, error)
	GetProfilesWhichContainsAllDefaultConfigurationKeysWithProfileId(defaultProfileId int) ([]int, error)
	GetProfilesWhichContainsAllDefaultConfigurationKeysUsingProfileName() ([]int, error)
	CreateProfile(tx *pg.Tx, infraProfile *InfraProfileEntity) error
	CreateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfigurationEntity) error

	UpdateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfigurationEntity) error
	UpdateProfile(tx *pg.Tx, profileName string, profile *InfraProfileEntity) error

	DeleteProfile(tx *pg.Tx, id int) error
	DeleteConfigurations(tx *pg.Tx, profileId int) error
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
func (impl *InfraConfigRepositoryImpl) CreateProfile(tx *pg.Tx, infraProfile *InfraProfileEntity) error {
	err := tx.Insert(infraProfile)
	return err
}

func (impl *InfraConfigRepositoryImpl) GetProfileIdByName(name string) (int, error) {
	var profileId int
	err := impl.dbConnection.Model(&InfraProfileEntity{}).
		Column("id").
		Where("name = ?", name).
		Where("active = ?", true).
		Select(&profileId)
	return profileId, err
}

func (impl *InfraConfigRepositoryImpl) GetProfileByName(name string) (*InfraProfileEntity, error) {
	infraProfile := &InfraProfileEntity{}
	err := impl.dbConnection.Model(infraProfile).
		Where("name = ?", name).
		Where("active = ?", true).
		Select()
	return infraProfile, err
}

func (impl *InfraConfigRepositoryImpl) GetProfileList(profileNameLike string) ([]*InfraProfileEntity, error) {
	var infraProfiles []*InfraProfileEntity
	query := impl.dbConnection.Model(&infraProfiles).
		Where("active = ?", true)
	if profileNameLike != "" {
		profileNameLike = "%" + profileNameLike + "%"
		query = query.Where("name LIKE ? OR name = ?", profileNameLike, DEFAULT_PROFILE_NAME)
	}
	query.Order("name ASC")
	err := query.Select()
	return infraProfiles, err
}

func (impl *InfraConfigRepositoryImpl) GetActiveProfileNames() ([]string, error) {
	var profileNames []string
	err := impl.dbConnection.Model((*InfraProfileEntity)(nil)).
		Column("name").
		Where("active = ?", true).
		Select(&profileNames)
	return profileNames, err
}

func (impl *InfraConfigRepositoryImpl) GetProfileListByIds(profileIds []int, includeDefault bool) ([]*InfraProfileEntity, error) {
	var infraProfiles []*InfraProfileEntity
	err := impl.dbConnection.Model(&infraProfiles).
		Where("active = ?", true).
		WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			if len(profileIds) > 0 {
				q = q.WhereOr("id IN (?)", pg.In(profileIds))
			}
			if includeDefault {
				q = q.WhereOr("name = ?", DEFAULT_PROFILE_NAME)
			}
			return q, nil
		}).Select()
	return infraProfiles, err
}

func (impl *InfraConfigRepositoryImpl) CreateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfigurationEntity) error {
	if len(configurations) == 0 {
		return nil
	}
	err := tx.Insert(&configurations)
	return err
}

func (impl *InfraConfigRepositoryImpl) UpdateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfigurationEntity) error {
	var err error
	for _, configuration := range configurations {
		_, err = tx.Model(configuration).
			Set("value = ?", configuration.Value).
			Set("unit = ?", configuration.Unit).
			Set("active = ?", configuration.Active).
			Set("updated_by = ?", configuration.UpdatedBy).
			Set("updated_on = ?", configuration.UpdatedOn).
			Where("id = ?", configuration.Id).
			Update()
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl *InfraConfigRepositoryImpl) GetConfigurationsByProfileName(profileName string) ([]*InfraProfileConfigurationEntity, error) {
	var configurations []*InfraProfileConfigurationEntity
	err := impl.dbConnection.Model(&configurations).
		Where("profile_id IN (SELECT id FROM infra_profile WHERE name = ? AND active = true)", profileName).
		Where("active = ?", true).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, NO_PROPERTIES_FOUND_ERROR
	}
	return configurations, err
}

func (impl *InfraConfigRepositoryImpl) GetConfigurationsByProfileIds(profileIds []int) ([]*InfraProfileConfigurationEntity, error) {
	// TODO Gireesh: use constants here
	if len(profileIds) == 0 {
		return nil, errors.New("profile ids cannot be empty")
	}

	var configurations []*InfraProfileConfigurationEntity
	err := impl.dbConnection.Model(&configurations).
		Where("profile_id IN (?)", pg.In(profileIds)).
		Where("active = ?", true).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, NO_PROPERTIES_FOUND_ERROR
	}
	return configurations, err
}

// todo: can use qualifierMapping service but need 2 db calls
func (impl *InfraConfigRepositoryImpl) GetConfigurationsByScope(scope Scope, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) ([]*InfraProfileConfigurationEntity, error) {
	var configurations []*InfraProfileConfigurationEntity
	getProfileIdByScopeQuery := "SELECT resource_id " +
		" FROM resource_qualifier_mapping " +
		fmt.Sprintf(" WHERE resource_type = %d AND identifier_key = %d AND identifier_value_int = %d AND active=true", resourceQualifiers.InfraProfile, GetIdentifierKey(APPLICATION, searchableKeyNameIdMap), scope.AppId)

	err := impl.dbConnection.Model(&configurations).
		Where("active = ?", true).
		Where("id IN (?)", getProfileIdByScopeQuery).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, NO_PROPERTIES_FOUND_ERROR
	}
	return configurations, err
}

func (impl *InfraConfigRepositoryImpl) GetProfilesWhichContainsAllDefaultConfigurationKeysWithProfileId(defaultProfileId int) ([]int, error) {
	query := " 	SELECT profile_id " +
		"   FROM infra_profile_configuration " +
		"   WHERE active = true " +
		"   GROUP BY profile_id HAVING COUNT(profile_id) = ( " +
		" 	      SELECT COUNT(id) " +
		" 	      FROM infra_profile_configuration " +
		" 	      WHERE active=true AND profile_id= ? " +
		"   ) "
	profileIds := make([]int, 0)
	_, err := impl.dbConnection.Query(&profileIds, query, defaultProfileId)
	return profileIds, err
}

func (impl *InfraConfigRepositoryImpl) GetProfilesWhichContainsAllDefaultConfigurationKeysUsingProfileName() ([]int, error) {
	query := " 	SELECT profile_id " +
		"   FROM infra_profile_configuration " +
		"   WHERE active = true " +
		"   GROUP BY profile_id HAVING COUNT(profile_id) = ( " +
		" 	      SELECT COUNT(ip.id) " +
		" 	      FROM infra_profile_configuration ipc" +
		"         INNER JOIN infra_profile ip ON ipc.profile_id = ip.id " +
		" 	      WHERE ip.active=true AND ipc.active=true AND ip.name= ? " +
		" ) "
	profileIds := make([]int, 0)
	_, err := impl.dbConnection.Query(&profileIds, query, DEFAULT_PROFILE_NAME)
	return profileIds, err
}

// func (impl *InfraConfigRepositoryImpl) GetIdentifierCountForDefaultProfile(defaultProfileId int, identifierKey int) (int, error) {
// 	queryToGetAppIdsWhichDoesntInheritDefaultConfigurations := " SELECT identifier_value_int " +
// 		" FROM resource_qualifier_mapping " +
// 		" WHERE resource_type = %d AND resource_id IN ( " +
// 		" 	SELECT profile_id " +
// 		"   FROM infra_profile_configuration " +
// 		"   GROUP BY profile_id HAVING COUNT(profile_id) = ( " +
// 		" 	      SELECT COUNT(id) " +
// 		" 	      FROM infra_profile_configuration " +
// 		" 	      WHERE active=true AND profile_id=%d ) " +
// 		" ) AND identifier_key = %d AND active=true"
// 	queryToGetAppIdsWhichDoesntInheritDefaultConfigurations = fmt.Sprintf(queryToGetAppIdsWhichDoesntInheritDefaultConfigurations, resourceQualifiers.InfraProfile, defaultProfileId, identifierKey)
//
// 	// exclude appIds which inherit default configurations
// 	query := " SELECT COUNT(id) " +
// 		" FROM app WHERE Id NOT IN (%s) and active=true"
// 	query = fmt.Sprintf(query, queryToGetAppIdsWhichDoesntInheritDefaultConfigurations)
// 	count := 0
// 	_, err := impl.dbConnection.Query(&count, query)
// 	return count, err
// }

func (impl *InfraConfigRepositoryImpl) UpdateProfile(tx *pg.Tx, profileName string, profile *InfraProfileEntity) error {
	_, err := tx.Model(profile).
		Set("name=?", profile.Name).
		Set("description=?", profile.Description).
		Set("updated_by=?", profile.UpdatedBy).
		Set("updated_on=?", profile.UpdatedOn).
		Where("name = ?", profileName).
		Update()
	return err
}

func (impl *InfraConfigRepositoryImpl) DeleteProfile(tx *pg.Tx, id int) error {
	_, err := tx.Model(&InfraProfileEntity{}).
		Set("active=?", false).
		Where("id = ?", id).
		Update()
	return err
}

func (impl *InfraConfigRepositoryImpl) DeleteConfigurations(tx *pg.Tx, profileId int) error {
	_, err := tx.Model(&InfraProfileConfigurationEntity{}).
		Set("active=?", false).
		Where("profile_id = ? ", profileId).
		Update()
	return err
}
