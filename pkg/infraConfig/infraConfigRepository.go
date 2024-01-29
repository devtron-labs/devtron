package infraConfig

import (
	"fmt"
	helper2 "github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/pkg/errors"
	"time"
)

type ProfileIdentifierCount struct {
	ProfileId       int
	IdentifierCount int
}

type InfraConfigRepository interface {
	GetProfileIdByName(name string) (int, error)
	GetProfileByName(name string) (*InfraProfile, error)
	GetProfileList(profileNameLike string) ([]*InfraProfile, error)

	// GetProfileListByIds returns the list of profiles for the given profileIds
	// includeDefault is used to explicitly include the default profile in the list
	GetProfileListByIds(profileIds []int, includeDefault bool) ([]*InfraProfile, error)

	GetConfigurationsByProfileName(profileName string) ([]*InfraProfileConfiguration, error)
	GetConfigurationsByProfileId(profileIds []int) ([]*InfraProfileConfiguration, error)
	GetConfigurationsByScope(scope Scope, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) ([]*InfraProfileConfiguration, error)

	GetIdentifierList(lisFilter IdentifierListFilter, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) ([]*Identifier, error)
	GetIdentifierCountForDefaultProfile(defaultProfileId int, identifierType int) (int, error)
	GetIdentifierCountForNonDefaultProfiles(profileIds []int, identifierType int) ([]ProfileIdentifierCount, error)

	CreateProfile(tx *pg.Tx, infraProfile *InfraProfile) error
	CreateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfiguration) error

	UpdateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfiguration) error
	UpdateProfile(tx *pg.Tx, profileName string, profile *InfraProfile) error

	DeleteProfileIdentifierMappingsByIds(tx *pg.Tx, userId int32, identifierIds []int, identifierType IdentifierType, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) error
	DeleteProfile(tx *pg.Tx, profileName string) error
	DeleteConfigurations(tx *pg.Tx, profileName string) error
	DeleteProfileIdentifierMappings(tx *pg.Tx, profileName string) error
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

func (impl *InfraConfigRepositoryImpl) GetProfileIdByName(name string) (int, error) {
	var profileId int
	err := impl.dbConnection.Model(&InfraProfile{}).
		Column("id").
		Where("name = ?", name).
		Where("active = ?", true).
		Select(&profileId)
	return profileId, err
}

func (impl *InfraConfigRepositoryImpl) GetProfileByName(name string) (*InfraProfile, error) {
	infraProfile := &InfraProfile{}
	err := impl.dbConnection.Model(infraProfile).
		Where("name = ?", name).
		Where("active = ?", true).
		Select()
	return infraProfile, err
}

func (impl *InfraConfigRepositoryImpl) GetProfileList(profileNameLike string) ([]*InfraProfile, error) {
	var infraProfiles []*InfraProfile
	query := impl.dbConnection.Model(&infraProfiles).
		Where("active = ?", true)
	if profileNameLike == "" {
		profileNameLike = "%" + profileNameLike + "%"
		query = query.Where("name LIKE ? OR name = ?", profileNameLike, DEFAULT_PROFILE_NAME)
	}
	query.Order("name ASC")
	err := query.Select()
	return infraProfiles, err
}

func (impl *InfraConfigRepositoryImpl) GetProfileListByIds(profileIds []int, includeDefault bool) ([]*InfraProfile, error) {
	var infraProfiles []*InfraProfile
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

func (impl *InfraConfigRepositoryImpl) CreateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfiguration) error {
	if len(configurations) == 0 {
		return nil
	}
	err := tx.Insert(&configurations)
	return err
}

func (impl *InfraConfigRepositoryImpl) UpdateConfigurations(tx *pg.Tx, configurations []*InfraProfileConfiguration) error {
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

func (impl *InfraConfigRepositoryImpl) GetConfigurationsByProfileName(profileName string) ([]*InfraProfileConfiguration, error) {
	var configurations []*InfraProfileConfiguration
	err := impl.dbConnection.Model(&configurations).
		Where("profile_id IN (SELECT id FROM infra_profile WHERE name = ? AND active = true)", profileName).
		Where("active = ?", true).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, errors.New(NO_PROPERTIES_FOUND)
	}
	return configurations, err
}

func (impl *InfraConfigRepositoryImpl) GetConfigurationsByProfileId(profileIds []int) ([]*InfraProfileConfiguration, error) {
	if len(profileIds) == 0 {
		return nil, errors.New("profileIds cannot be empty")
	}

	var configurations []*InfraProfileConfiguration
	err := impl.dbConnection.Model(&configurations).
		Where("profile_id IN (?)", pg.In(profileIds)).
		Where("active = ?", true).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, errors.New(NO_PROPERTIES_FOUND)
	}
	return configurations, err
}

func (impl *InfraConfigRepositoryImpl) GetConfigurationsByScope(scope Scope, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) ([]*InfraProfileConfiguration, error) {
	var configurations []*InfraProfileConfiguration
	getProfileIdByScopeQuery := "SELECT resource_id " +
		" FROM resource_qualifier_mapping " +
		fmt.Sprintf(" WHERE resource_type = %d AND identifier_key = %d AND identifier_value_int = %d AND active=true", resourceQualifiers.InfraProfile, GetIdentifierKey(APPLICATION, searchableKeyNameIdMap), scope.AppId)

	err := impl.dbConnection.Model(&configurations).
		Where("active = ?", true).
		Where("id IN (?)", getProfileIdByScopeQuery).
		Select()
	if errors.Is(err, pg.ErrNoRows) {
		return nil, errors.New(NO_PROPERTIES_FOUND)
	}
	return configurations, err
}

func (impl *InfraConfigRepositoryImpl) GetIdentifierCountForDefaultProfile(defaultProfileId int, identifierKey int) (int, error) {
	queryToGetAppIdsWhichDoesntInheritDefaultConfigurations := " SELECT identifier_value_int " +
		" FROM resource_qualifier_mapping " +
		" WHERE resource_type = %d AND resource_id IN ( " +
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
	query := " SELECT COUNT(DISTINCT identifier_value_int) as identifier_count, resource_id" +
		" FROM resource_qualifier_mapping " +
		" WHERE resource_type = ? AND identifier_key = ? AND active=true "
	if len(profileIds) > 0 {
		query += fmt.Sprintf(" AND resource_id IN (%s) ", helper2.GetCommaSepratedStringWithComma(profileIds))
	}

	query += " GROUP BY resource_id"
	counts := make([]ProfileIdentifierCount, 0)
	_, err := impl.dbConnection.Query(&counts, query, resourceQualifiers.InfraProfile, identifierType)
	return counts, err
}

func (impl *InfraConfigRepositoryImpl) UpdateProfile(tx *pg.Tx, profileName string, profile *InfraProfile) error {
	_, err := tx.Model(profile).
		Set("name=?", profile.Name).
		Set("description=?", profile.Description).
		Set("updated_by=?", profile.UpdatedBy).
		Set("updated_on=?", profile.UpdatedOn).
		Where("name = ?", profileName).
		Update()
	return err
}

func (impl *InfraConfigRepositoryImpl) DeleteProfile(tx *pg.Tx, profileName string) error {
	_, err := tx.Model(&InfraProfile{}).
		Set("active=?", false).
		Where("name = ?", profileName).
		Update()
	return err
}

func (impl *InfraConfigRepositoryImpl) DeleteConfigurations(tx *pg.Tx, profileName string) error {
	_, err := tx.Model(&InfraProfileConfiguration{}).
		Set("active=?", false).
		Where("profile_id IN (SELECT id FROM infra_profile WHERE name = ? AND active = true)", profileName).
		Update()
	return err
}

func (impl *InfraConfigRepositoryImpl) GetIdentifierList(listFilter IdentifierListFilter, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) ([]*Identifier, error) {
	listFilter.IdentifierNameLike = "%" + listFilter.IdentifierNameLike + "%"
	identifierType := GetIdentifierKey(listFilter.IdentifierType, searchableKeyNameIdMap)
	totalOverriddenCountQuery := "SELECT COUNT(id) " +
		" FROM resource_qualifier_mapping " +
		fmt.Sprintf(" WHERE resource_type = %d ", resourceQualifiers.InfraProfile) +
		" AND active=true "

	// for empty profile name we have to get identifiers
	if listFilter.ProfileName == ALL_PROFILES {
		return impl.getIdentifiersListForMiscProfiles(listFilter, totalOverriddenCountQuery, identifierType)
	}

	// for default profile
	if listFilter.ProfileName == DEFAULT_PROFILE_NAME {
		identifiers, err := impl.getIdentifiersListForDefaultProfile(listFilter, identifierType, totalOverriddenCountQuery)
		return identifiers, err
	}

	// for any other profile
	finalQuery := "SELECT identifier_value_int AS id, identifier_value_string AS name, resource_id as profile_id, COUNT(id) OVER() AS total_identifier_count, (" + totalOverriddenCountQuery + ") AS overridden_identifier_count " +
		" FROM resource_qualifier_mapping "
	finalQuery += fmt.Sprintf(" WHERE resource_type = %d ", resourceQualifiers.InfraProfile)
	if listFilter.ProfileName != "" {
		finalQuery += fmt.Sprintf(" AND resource_id IN (SELECT id FROM infra_profile WHERE name = '%s')", listFilter.ProfileName)
	}

	filterQuery := "SELECT id" +
		" FROM app " +
		" WHERE active=true " +
		" AND app_name LIKE ? " +
		" ORDER BY name ? " +
		" LIMIT ? " +
		" OFFSET ? "

	finalQuery += " AND identifier_key = ? " +
		" AND id IN (" + filterQuery + ") " +
		" AND active=true"

	var identifiers []*Identifier
	_, err := impl.dbConnection.Query(&identifiers, finalQuery,
		identifierType,
		listFilter.IdentifierNameLike,
		listFilter.SortOrder,
		listFilter.Limit,
		listFilter.Offset)
	return identifiers, err

}

func (impl *InfraConfigRepositoryImpl) getIdentifiersListForMiscProfiles(listFilter IdentifierListFilter, totalOverriddenCountQuery string, identifierType int) ([]*Identifier, error) {
	// get apps first and then get their respective profile Ids
	// get apps using filters
	query := "SELECT id," +
		"app_name AS name," +
		" COUNT(id) OVER() AS total_identifier_count,(" + totalOverriddenCountQuery + ") AS overridden_identifier_count " +
		" FROM app " +
		" WHERE active=true " +
		" AND app_name LIKE ? " +
		" ORDER BY name ? " +
		" LIMIT ? " +
		" OFFSET ? "
	var identifiers []*Identifier
	_, err := impl.dbConnection.Query(&identifiers, query,
		listFilter.IdentifierNameLike,
		listFilter.SortOrder,
		listFilter.Limit,
		listFilter.Offset)
	if err != nil {
		return nil, err
	}
	// get profileIds for the above identifiers
	profileIdentifiersMappings := make([]*Identifier, 0)
	profileIdentifiersMappingsQuery := "SELECT identifier_value_int AS id,resource_id AS profile_id " +
		" FROM resource_qualifier_mapping " +
		" WHERE resource_type = ? " +
		" AND identifier_key = ? " +
		" AND active = true "
	_, err = impl.dbConnection.Query(&profileIdentifiersMappings, profileIdentifiersMappingsQuery, resourceQualifiers.InfraProfile, identifierType)
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

func (impl *InfraConfigRepositoryImpl) getIdentifiersListForDefaultProfile(listFilter IdentifierListFilter, identifierType int, totalOverriddenCountQuery string) ([]*Identifier, error) {
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
		"(SELECT id FROM infra_profile WHERE name = ?) AS profile_id, COUNT(id) OVER() AS total_identifier_count,(" + totalOverriddenCountQuery + ") AS overridden_identifier_count " +
		" FROM app " +
		" WHERE active=true " +
		" AND app_name LIKE ? " +
		" AND id NOT IN ( " + queryToGetAppIdsWhichDoesNotInheritDefaultConfigurations + " ) " +
		" ORDER BY name ? " +
		" LIMIT ? " +
		" OFFSET ? "

	var identifiers []*Identifier
	_, err := impl.dbConnection.Query(&identifiers, query,
		DEFAULT_PROFILE_NAME,
		listFilter.IdentifierNameLike,
		listFilter.SortOrder,
		listFilter.Limit,
		listFilter.Offset)
	return identifiers, err
}

func (impl *InfraConfigRepositoryImpl) DeleteProfileIdentifierMappings(tx *pg.Tx, profileName string) error {
	_, err := tx.Model(&resourceQualifiers.QualifierMapping{}).
		Set("active=?", false).
		Where("resource_id IN (SELECT id FROM infra_profile WHERE name = ? AND active = true)", profileName).
		Where("resource_type=?", resourceQualifiers.InfraProfile).
		Where("active=?", true).
		Update()
	return err
}

func (impl *InfraConfigRepositoryImpl) DeleteProfileIdentifierMappingsByIds(tx *pg.Tx, userId int32, identifierIds []int, identifierType IdentifierType, searchableKeyNameIdMap map[bean.DevtronResourceSearchableKeyName]int) error {
	_, err := tx.Model(&resourceQualifiers.QualifierMapping{}).
		Set("active=?", false).
		Set("updated_by=?").
		Set("updated_on=?", time.Now()).
		Where("resource_type=?", resourceQualifiers.InfraProfile).
		Where("active=?", true).
		Where("identifier_value_int IN (?)", pg.In(identifierIds)).
		Where("identifier_key=?", GetIdentifierKey(identifierType, searchableKeyNameIdMap)).
		Update()
	return err
}
