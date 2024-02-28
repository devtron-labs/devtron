package resourceQualifiers

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"time"
)

type QualifierMappingService interface {
	CreateQualifierMappings(qualifierMappings []*QualifierMapping, tx *pg.Tx) ([]*QualifierMapping, error)
	GetQualifierMappingsForFilter(scope Scope) ([]*QualifierMapping, error)
	GetQualifierMappingsForFilterById(resourceId int) ([]*QualifierMapping, error)
	GetQualifierMappings(resourceType ResourceType, scope *Scope, resourceIds []int) ([]*QualifierMapping, error)
	GetQualifierMappingsByResourceType(resourceType ResourceType) ([]*QualifierMapping, error)
	GetActiveIdentifierCountPerResource(resourceType ResourceType, resourceIds []int, identifierKey int, identifierValueIntSpaceQuery string) ([]ResourceIdentifierCount, error)
	GetIdentifierIdsByResourceTypeAndIds(resourceType ResourceType, resourceIds []int, identifierKey int) ([]int, error)
	GetActiveMappingsCount(resourceType ResourceType, excludeIdentifiersQuery string, identifierKey int) (int, error)
	DeleteAllQualifierMappings(resourceType ResourceType, auditLog sql.AuditLog, tx *pg.Tx) error
	DeleteAllQualifierMappingsByResourceTypeAndId(resourceType ResourceType, resourceId int, auditLog sql.AuditLog, tx *pg.Tx) error
	DeleteByIdentifierKeyAndValue(resourceType ResourceType, identifierKey int, identifierValue int, auditLog sql.AuditLog, tx *pg.Tx) error
	DeleteAllByResourceTypeAndQualifierIds(resourceType ResourceType, resourceId int, qualifierIds []int, userId int32, tx *pg.Tx) error
	DeleteAllByIds(qualifierMappingIds []int, userId int32, tx *pg.Tx) error
	DeleteGivenQualifierMappingsByResourceType(resourceType ResourceType, identifierKey int, identifierValueInts []int, auditLog sql.AuditLog, tx *pg.Tx) error
	GetResourceIdsByIdentifier(resourceType ResourceType, identifierKey int, identifierId int) ([]int, error)
	GetQualifierMappingsWithIdentifierFilter(resourceType ResourceType, resourceId, identifierKey int, identifierValueStringLike, identifierValueSortOrder, excludeActiveIdentifiersQuery string, limit, offset int, needTotalCount bool) ([]*QualifierMappingWithExtraColumns, error)

	CreateMappingsForSelections(tx *pg.Tx, userId int32, resourceMappingSelections []*ResourceMappingSelection) ([]*ResourceMappingSelection, error)
	CreateMappings(tx *pg.Tx, userId int32, resourceType ResourceType, resourceIds []int, qualifierSelector QualifierSelector, scopes []*Scope) error
	GetResourceMappingsForScopes(resourceType ResourceType, qualifierSelector QualifierSelector, scopes []*Scope) ([]ResourceQualifierMappings, error)
	GetResourceMappingsForResources(resourceType ResourceType, resourceIds []int, qualifierSelector QualifierSelector) ([]ResourceQualifierMappings, error)
}

func (impl QualifierMappingServiceImpl) CreateMappingsForSelections(tx *pg.Tx, userId int32, resourceMappingSelections []*ResourceMappingSelection) ([]*ResourceMappingSelection, error) {

	resourceKeyMap := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()

	parentMappings := make([]*QualifierMapping, 0)
	childrenMappings := make([]*QualifierMapping, 0)
	parentMappingsMap := make(map[string]*QualifierMapping)

	mappingsToSelection := make(map[*QualifierMapping]*ResourceMappingSelection)
	for _, selection := range resourceMappingSelections {

		var parent *QualifierMapping
		children := make([]*QualifierMapping, 0)
		if selection.QualifierSelector.isCompound() {
			parent, children = GetQualifierMappingsForCompoundQualifier(selection, resourceKeyMap, userId)
			parentMappingsMap[parent.CompositeKey] = parent
		} else {
			intValue, stringValue := GetValuesFromScope(selection.QualifierSelector, selection.Scope)
			parent = selection.toResourceMapping(resourceKeyMap, intValue, stringValue, "", userId)
		}
		mappingsToSelection[parent] = selection
		parentMappings = append(parentMappings, parent)
		childrenMappings = append(childrenMappings, children...)
	}

	if len(parentMappings) > 0 {
		_, err := impl.qualifierMappingRepository.CreateQualifierMappings(parentMappings, tx)
		if err != nil {
			impl.logger.Errorw("error in getting parent mappings", "mappings", parentMappings, "err", err)
			return nil, err
		}
	}

	for _, mapping := range parentMappings {
		if selection, ok := mappingsToSelection[mapping]; ok {
			selection.Id = mapping.Id
		}
	}

	for _, childrenMapping := range childrenMappings {
		if parentScope, ok := parentMappingsMap[childrenMapping.CompositeKey]; ok {
			childrenMapping.ParentIdentifier = parentScope.Id
		}
	}

	if len(childrenMappings) > 0 {
		_, err := impl.qualifierMappingRepository.CreateQualifierMappings(childrenMappings, tx)
		if err != nil {
			impl.logger.Errorw("error in getting mappings", "err", err, "mappings", childrenMappings)
			return nil, err
		}
	}
	return lo.Values(mappingsToSelection), nil
}

func (impl QualifierMappingServiceImpl) CreateMappings(tx *pg.Tx, userId int32, resourceType ResourceType, resourceIds []int, qualifierSelector QualifierSelector, scopes []*Scope) error {
	mappings := make([]*ResourceMappingSelection, 0)
	for _, id := range resourceIds {
		for _, scope := range scopes {
			mapping := &ResourceMappingSelection{
				ResourceType:      resourceType,
				ResourceId:        id,
				QualifierSelector: qualifierSelector,
				Scope:             scope,
			}
			mappings = append(mappings, mapping)
		}
	}
	_, err := impl.CreateMappingsForSelections(tx, userId, mappings)
	return err
}

func (impl *QualifierMappingServiceImpl) filterAndGroupMappings(mappings []*QualifierMapping, selector QualifierSelector) [][]*QualifierMapping {

	numQualifiers := GetNumOfChildQualifiers(selector.toQualifier())
	parentIdToChildScopes := make(map[int][]*QualifierMapping)
	parentScopeIdToScope := make(map[int]*QualifierMapping, 0)
	parentScopeIds := make([]int, 0)
	for _, scope := range mappings {
		// is not parent so append it to the list in the map with key as its parent scopeID
		if scope.ParentIdentifier > 0 {
			parentIdToChildScopes[scope.ParentIdentifier] = append(parentIdToChildScopes[scope.ParentIdentifier], scope)
		} else {
			//is parent so collect IDs and put it in a map for easy retrieval
			parentScopeIds = append(parentScopeIds, scope.Id)
			parentScopeIdToScope[scope.Id] = scope
		}
	}

	for parentScopeId, _ := range parentIdToChildScopes {
		// this deletes the keys in the map where the key does not exist in the collected IDs for parent
		if !slices.Contains(parentScopeIds, parentScopeId) {
			delete(parentIdToChildScopes, parentScopeId)
		}
	}

	//selectedParentScopes :=  make([]*QualifierMapping,0)
	groupedMappings := make([][]*QualifierMapping, 0)
	for parentScopeId, childScopes := range parentIdToChildScopes {
		if len(childScopes) == numQualifiers {
			selectedParentScope := parentScopeIdToScope[parentScopeId]
			//selectedParentScopes = append(selectedParentScopes, selectedParentScope)

			mappingsGroup := []*QualifierMapping{selectedParentScope}
			mappingsGroup = append(mappingsGroup, childScopes...)
			groupedMappings = append(groupedMappings, mappingsGroup)
		}
	}
	return groupedMappings
}

func (impl QualifierMappingServiceImpl) getAppEnvScopeFromGroup(group []*QualifierMapping) *Scope {
	resourceKeyToName := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyIdNameMap()
	var appId, envId int
	var appName, envName string
	for _, mapping := range group {
		field := resourceKeyToName[mapping.IdentifierKey]
		switch field {
		case bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID:
			appId = mapping.IdentifierValueInt
			appName = mapping.IdentifierValueString
		case bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID:
			envId = mapping.IdentifierValueInt
			envName = mapping.IdentifierValueString
		}
	}
	return &Scope{
		AppId: appId,
		EnvId: envId,
		SystemMetadata: &SystemMetadata{
			EnvironmentName: envName,
			AppName:         appName,
		},
	}
}

func (impl QualifierMappingServiceImpl) getScopesForAppEnvSelector(mappingGroups [][]*QualifierMapping) map[int][]*Scope {

	resourceIdToScope := make(map[int][]*Scope)
	for _, group := range mappingGroups {
		scope := impl.getAppEnvScopeFromGroup(group)
		resourceId := group[0].ResourceId

		if _, ok := resourceIdToScope[resourceId]; ok {
			resourceIdToScope[resourceId] = append(resourceIdToScope[resourceId], scope)
		} else {
			resourceIdToScope[resourceId] = []*Scope{scope}
		}
	}
	return resourceIdToScope
}

func (impl QualifierMappingServiceImpl) GetResourceMappingsForScopes(resourceType ResourceType, qualifierSelector QualifierSelector, scopes []*Scope) ([]ResourceQualifierMappings, error) {
	if qualifierSelector != ApplicationEnvironmentSelector {
		return nil, fmt.Errorf("selector currently not implemented")
	}

	keyMap := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()

	valuesMap := make(map[Qualifier][][]int)
	appIds := make([]int, 0)
	envIds := make([]int, 0)
	for _, scope := range scopes {
		appIds = append(appIds, scope.AppId)
		envIds = append(envIds, scope.EnvId)
	}
	valuesMap[qualifierSelector.toQualifier()] = [][]int{appIds, envIds}
	mappings, err := impl.qualifierMappingRepository.GetQualifierMappingsForListOfQualifierValues(resourceType, nil, keyMap, []int{})
	if err != nil {
		return nil, err
	}

	return impl.processMappings(resourceType, mappings, qualifierSelector)
}
func (impl QualifierMappingServiceImpl) GetResourceMappingsForResources(resourceType ResourceType, resourceIds []int, qualifierSelector QualifierSelector) ([]ResourceQualifierMappings, error) {
	mappings, err := impl.qualifierMappingRepository.GetMappingsByResourceTypeAndIdsAndQualifierId(resourceType, resourceIds, int(qualifierSelector.toQualifier()))
	if err != nil {
		return nil, err
	}

	return impl.processMappings(resourceType, mappings, qualifierSelector)
}

func (impl QualifierMappingServiceImpl) processMappings(resourceType ResourceType, mappings []*QualifierMapping, qualifierSelector QualifierSelector) ([]ResourceQualifierMappings, error) {
	groups := impl.filterAndGroupMappings(mappings, qualifierSelector)
	if qualifierSelector != ApplicationEnvironmentSelector {
		return nil, fmt.Errorf("selector currently not implemented")
	}
	resourceIdToScopes := impl.getScopesForAppEnvSelector(groups)

	qualifierMappings := make([]ResourceQualifierMappings, 0)

	for resourceId, scopes := range resourceIdToScopes {
		for _, scope := range scopes {
			qualifierMappings = append(qualifierMappings, ResourceQualifierMappings{
				ResourceId:   resourceId,
				ResourceType: resourceType,
				Scope:        scope,
			})
		}
	}
	return qualifierMappings, nil
}

type QualifierMappingServiceImpl struct {
	logger                              *zap.SugaredLogger
	qualifierMappingRepository          QualifiersMappingRepository
	devtronResourceSearchableKeyService devtronResource.DevtronResourceSearchableKeyService
}

func NewQualifierMappingServiceImpl(logger *zap.SugaredLogger, qualifierMappingRepository QualifiersMappingRepository, devtronResourceSearchableKeyService devtronResource.DevtronResourceSearchableKeyService) (*QualifierMappingServiceImpl, error) {
	return &QualifierMappingServiceImpl{
		logger:                              logger,
		qualifierMappingRepository:          qualifierMappingRepository,
		devtronResourceSearchableKeyService: devtronResourceSearchableKeyService,
	}, nil
}

func (impl QualifierMappingServiceImpl) CreateQualifierMappings(qualifierMappings []*QualifierMapping, tx *pg.Tx) ([]*QualifierMapping, error) {
	return impl.qualifierMappingRepository.CreateQualifierMappings(qualifierMappings, tx)
}

func (impl QualifierMappingServiceImpl) GetQualifierMappings(resourceType ResourceType, scope *Scope, resourceIds []int) ([]*QualifierMapping, error) {
	searchableIdMap := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()
	return impl.qualifierMappingRepository.GetQualifierMappings(resourceType, scope, searchableIdMap, resourceIds)
}

func (impl QualifierMappingServiceImpl) GetQualifierMappingsForFilter(scope Scope) ([]*QualifierMapping, error) {
	searchableKeyNameIdMap := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()
	return impl.qualifierMappingRepository.GetQualifierMappingsForFilter(scope, searchableKeyNameIdMap)
}

func (impl QualifierMappingServiceImpl) GetQualifierMappingsForFilterById(resourceId int) ([]*QualifierMapping, error) {
	return impl.qualifierMappingRepository.GetQualifierMappingsForFilterById(resourceId)
}

func (impl QualifierMappingServiceImpl) GetQualifierMappingsByResourceType(resourceType ResourceType) ([]*QualifierMapping, error) {
	return impl.qualifierMappingRepository.GetQualifierMappingsByResourceType(resourceType)
}

func (impl QualifierMappingServiceImpl) DeleteAllQualifierMappings(resourceType ResourceType, auditLog sql.AuditLog, tx *pg.Tx) error {
	return impl.qualifierMappingRepository.DeleteAllQualifierMappings(resourceType, auditLog, tx)
}
func (impl QualifierMappingServiceImpl) DeleteAllQualifierMappingsByResourceTypeAndId(resourceType ResourceType, resourceId int, auditLog sql.AuditLog, tx *pg.Tx) error {
	return impl.qualifierMappingRepository.DeleteAllQualifierMappingsByResourceTypeAndId(resourceType, resourceId, auditLog, tx)
}

func (impl QualifierMappingServiceImpl) DeleteByIdentifierKeyAndValue(resourceType ResourceType, identifierKey int, identifierValue int, auditLog sql.AuditLog, tx *pg.Tx) error {
	return impl.qualifierMappingRepository.DeleteByResourceTypeIdentifierKeyAndValue(resourceType, identifierKey, identifierValue, auditLog, tx)
}

func (impl QualifierMappingServiceImpl) DeleteAllByResourceTypeAndQualifierIds(resourceType ResourceType, resourceId int, qualifierIds []int, userId int32, tx *pg.Tx) error {
	auditLog := sql.AuditLog{
		CreatedOn: time.Now(),
		CreatedBy: userId,
		UpdatedOn: time.Now(),
		UpdatedBy: userId,
	}
	return impl.qualifierMappingRepository.DeleteAllByResourceTypeAndQualifierId(resourceType, resourceId, qualifierIds, auditLog, tx)
}

func (impl QualifierMappingServiceImpl) DeleteAllByIds(qualifierMappingIds []int, userId int32, tx *pg.Tx) error {
	auditLog := sql.AuditLog{
		CreatedOn: time.Now(),
		CreatedBy: userId,
		UpdatedOn: time.Now(),
		UpdatedBy: userId,
	}
	return impl.qualifierMappingRepository.DeleteAllByIds(qualifierMappingIds, auditLog, tx)
}

func (impl QualifierMappingServiceImpl) DeleteGivenQualifierMappingsByResourceType(resourceType ResourceType, identifierKey int, identifierValueInts []int, auditLog sql.AuditLog, tx *pg.Tx) error {
	return impl.qualifierMappingRepository.DeleteGivenQualifierMappingsByResourceType(resourceType, identifierKey, identifierValueInts, auditLog, tx)
}

func (impl QualifierMappingServiceImpl) GetActiveIdentifierCountPerResource(resourceType ResourceType, resourceIds []int, identifierKey int, identifierValueIntSpaceQuery string) ([]ResourceIdentifierCount, error) {
	return impl.qualifierMappingRepository.GetActiveIdentifierCountPerResource(resourceType, resourceIds, identifierKey, identifierValueIntSpaceQuery)
}

func (impl QualifierMappingServiceImpl) GetIdentifierIdsByResourceTypeAndIds(resourceType ResourceType, resourceIds []int, identifierKey int) ([]int, error) {
	return impl.qualifierMappingRepository.GetIdentifierIdsByResourceTypeAndIds(resourceType, resourceIds, identifierKey)
}

func (impl QualifierMappingServiceImpl) GetActiveMappingsCount(resourceType ResourceType, excludeIdentifiersQuery string, identifierKey int) (int, error) {
	return impl.qualifierMappingRepository.GetActiveMappingsCount(resourceType, excludeIdentifiersQuery, identifierKey)
}

func (impl QualifierMappingServiceImpl) GetResourceIdsByIdentifier(resourceType ResourceType, identifierKey int, identifierId int) ([]int, error) {
	return impl.qualifierMappingRepository.GetResourceIdsByIdentifier(resourceType, identifierKey, identifierId)
}

func (impl QualifierMappingServiceImpl) GetQualifierMappingsWithIdentifierFilter(resourceType ResourceType, resourceId, identifierKey int, identifierValueStringLike, identifierValueSortOrder string, excludeActiveIdentifiersQuery string, limit, offset int, needTotalCount bool) ([]*QualifierMappingWithExtraColumns, error) {
	return impl.qualifierMappingRepository.GetQualifierMappingsWithIdentifierFilter(resourceType, resourceId, identifierKey, identifierValueStringLike, identifierValueSortOrder, excludeActiveIdentifiersQuery, limit, offset, needTotalCount)
}
