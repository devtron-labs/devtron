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

	GetIdentifierList(lisFilter IdentifierListFilter, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) ([]*Identifier, error)
	// GetIdentifierCountForDefaultProfile(defaultProfileId int, identifierType int) (int, error)
	GetProfilesWhichContainsAllDefaultConfigurationKeys(defaultProfileId int) ([]int, error)
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
	// todo: hanlde pg no rows
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

func (impll *InfraConfigRepositoryImpl) GetProfilesWhichContainsAllDefaultConfigurationKeys(defaultProfileId int) ([]int, error) {
	query := " 	SELECT profile_id " +
		"   FROM infra_profile_configuration " +
		"   WHERE active = true " +
		"   GROUP BY profile_id HAVING COUNT(profile_id) = ( " +
		" 	      SELECT COUNT(id) " +
		" 	      FROM infra_profile_configuration " +
		" 	      WHERE active=true AND profile_id= ? " +
		"   ) "
	profileIds := make([]int, 0)
	_, err := impll.dbConnection.Query(&profileIds, query, defaultProfileId)
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

// todo: move this to service layer
func (impl *InfraConfigRepositoryImpl) GetIdentifierList(listFilter IdentifierListFilter, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) ([]*Identifier, error) {
	listFilter.IdentifierNameLike = "%" + listFilter.IdentifierNameLike + "%"
	identifierType := GetIdentifierKey(listFilter.IdentifierType, searchableKeyNameIdMap)
	// for empty profile name we have to get identifiers
	if listFilter.ProfileName == ALL_PROFILES {
		return impl.getIdentifiersListForMiscProfiles(listFilter, identifierType)
	}

	// for default profile
	if listFilter.ProfileName == DEFAULT_PROFILE_NAME {
		identifiers, err := impl.getIdentifiersListForDefaultProfile(listFilter, identifierType)
		return identifiers, err
	}

	// for any other profile
	identifiers, err := impl.getIdentifiersListForNonDefaultProfile(listFilter, identifierType)
	return identifiers, err

}

// todo: move this to qualifier mapping service
func (impl *InfraConfigRepositoryImpl) getIdentifiersListForNonDefaultProfile(listFilter IdentifierListFilter, identifierType int) ([]*Identifier, error) {
	query := "SELECT identifier_value_int AS id, identifier_value_string AS name, resource_id as profile_id, COUNT(id) OVER() AS total_identifier_count " +
		" FROM resource_qualifier_mapping "
	query += fmt.Sprintf(" WHERE resource_type = %d ", resourceQualifiers.InfraProfile)
	if listFilter.ProfileName != "" {
		query += fmt.Sprintf(" AND resource_id IN (SELECT id FROM infra_profile WHERE name = '%s')", listFilter.ProfileName)
	}

	query += fmt.Sprintf(" AND identifier_key = %d ", identifierType) +
		" AND active=true "

	if listFilter.IdentifierNameLike != "" {
		query += fmt.Sprintf(" AND identifier_value_string LIKE '%s' ", listFilter.IdentifierNameLike)
	}
	if listFilter.SortOrder != "" {
		query += fmt.Sprintf(" ORDER BY name %s ", listFilter.SortOrder)
	}
	if listFilter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d ", listFilter.Limit)
	}
	if listFilter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d ", listFilter.Offset)
	}

	var identifiers []*Identifier
	_, err := impl.dbConnection.Query(&identifiers, query)
	return identifiers, err
}

// todo: move this to app service
func (impl *InfraConfigRepositoryImpl) getIdentifiersListForMiscProfiles(listFilter IdentifierListFilter, identifierType int) ([]*Identifier, error) {
	// get apps first and then get their respective profile Ids
	// get apps using filters
	query := "SELECT id," +
		"app_name AS name," +
		" COUNT(id) OVER() AS total_identifier_count " +
		" FROM app " +
		" WHERE active=true "
	if listFilter.IdentifierNameLike != "" {
		query += fmt.Sprintf(" AND app_name LIKE '%s' ", listFilter.IdentifierNameLike)
	}
	if listFilter.SortOrder != "" {
		query += fmt.Sprintf(" ORDER BY name %s ", listFilter.SortOrder)
	}
	if listFilter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d ", listFilter.Limit)
	}
	if listFilter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d ", listFilter.Offset)
	}
	var identifiers []*Identifier
	_, err := impl.dbConnection.Query(&identifiers, query)
	if err != nil {
		return nil, err
	}
	// get profileIds for the above identifiers
	identifiers, err = impl.fillIdentifiersWithProfileId(identifierType, identifiers)
	return identifiers, err
}

// todo: split this into different db calls
func (impl *InfraConfigRepositoryImpl) getIdentifiersListForDefaultProfile(listFilter IdentifierListFilter, identifierType int) ([]*Identifier, error) {
	queryToGetAppIdsWhichDoesNotInheritDefaultConfigurations := " SELECT identifier_value_int " +
		" FROM resource_qualifier_mapping " +
		" WHERE resource_type = %d AND resource_id IN ( " +
		" 	SELECT profile_id " +
		"   FROM infra_profile_configuration " +
		"   GROUP BY profile_id HAVING COUNT(profile_id) = ( " +
		" 	      SELECT COUNT(ip.id) " +
		" 	      FROM infra_profile_configuration ipc" +
		"         INNER JOIN infra_profile ip ON ipc.profile_id = ip.id " +
		" 	      WHERE ip.active=true AND ipc.active=true AND ip.name='%s' ) " +
		" ) AND identifier_key = %d AND active=true"
	queryToGetAppIdsWhichDoesNotInheritDefaultConfigurations = fmt.Sprintf(queryToGetAppIdsWhichDoesNotInheritDefaultConfigurations, resourceQualifiers.InfraProfile, DEFAULT_PROFILE_NAME, identifierType)

	query := "SELECT id," +
		"app_name AS name," +
		"COUNT(id) OVER() AS total_identifier_count " +
		" FROM app " +
		" WHERE active=true " +
		" AND id NOT IN ( " + queryToGetAppIdsWhichDoesNotInheritDefaultConfigurations + " ) "
	if listFilter.IdentifierNameLike != "" {
		query += fmt.Sprintf(" AND app_name LIKE '%s' ", listFilter.IdentifierNameLike)
	}
	if listFilter.SortOrder != "" {
		query += fmt.Sprintf(" ORDER BY name %s ", listFilter.SortOrder)
	}
	if listFilter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d ", listFilter.Limit)
	}
	if listFilter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d ", listFilter.Offset)
	}

	var identifiers []*Identifier
	_, err := impl.dbConnection.Query(&identifiers, query)
	if err != nil {
		return nil, err
	}
	identifiers, err = impl.fillIdentifiersWithProfileId(identifierType, identifiers)
	return identifiers, err
}

// todo: move this to qualifier mapping service
func (impl *InfraConfigRepositoryImpl) fillIdentifiersWithProfileId(identifierType int, identifiers []*Identifier) ([]*Identifier, error) {
	// get profileIds for the above identifiers
	profileIdentifiersMappings := make([]*Identifier, 0)
	profileIdentifiersMappingsQuery := "SELECT identifier_value_int AS id,resource_id AS profile_id " +
		" FROM resource_qualifier_mapping " +
		" WHERE resource_type = ? " +
		" AND identifier_key = ? " +
		" AND active = true "
	_, err := impl.dbConnection.Query(&profileIdentifiersMappings, profileIdentifiersMappingsQuery, resourceQualifiers.InfraProfile, identifierType)
	if err != nil {
		return nil, err
	}
	identifiersMap := make(map[int]*Identifier)
	for _, identifier := range identifiers {
		identifiersMap[identifier.Id] = identifier
	}
	for _, profileIdentifierMapping := range profileIdentifiersMappings {
		identifier, ok := identifiersMap[profileIdentifierMapping.Id]
		if ok {
			identifier.ProfileId = profileIdentifierMapping.ProfileId
		}
	}
	return identifiers, nil
}
