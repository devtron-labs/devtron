package resourceQualifiers

import (
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/read"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"time"
)

type QualifierMappingService interface {
	CreateQualifierMappings(qualifierMappings []*QualifierMapping, tx *pg.Tx) ([]*QualifierMapping, error)
	GetQualifierMappingsForFilter(scope Scope) ([]*QualifierMapping, error)
	GetQualifierMappingsByResourceId(resourceId int, resourceType ResourceType) ([]*QualifierMapping, error)
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

	DeleteResourceMappingsForScopes(tx *pg.Tx, userId int32, resourceType ResourceType, qualifierSelector QualifierSelector, scopes []*SelectionIdentifier) error
	CreateMappingsForSelections(tx *pg.Tx, userId int32, resourceMappingSelections []*ResourceMappingSelection) ([]*ResourceMappingSelection, error)
	CreateMappings(tx *pg.Tx, userId int32, resourceType ResourceType, resourceIds []int, qualifierSelector QualifierSelector, selectionIdentifiers []*SelectionIdentifier) error
	GetResourceMappingsForSelections(resourceType ResourceType, qualifierSelector QualifierSelector, selectionIdentifiers []*SelectionIdentifier) ([]ResourceQualifierMappings, error)
	GetResourceMappingsForResources(resourceType ResourceType, resourceIds []int, qualifierSelector QualifierSelector) ([]ResourceQualifierMappings, error)
	StartTx() (*pg.Tx, error)
	RollbackTx(tx *pg.Tx) error
	CommitTx(tx *pg.Tx) error
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
			intValue, stringValue := GetValuesFromSelectionIdentifier(selection.QualifierSelector, selection.SelectionIdentifier)
			parent = selection.toResourceMapping(selection.QualifierSelector, resourceKeyMap, intValue, stringValue, "", userId)
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
		if mapping, ok := parentMappingsMap[childrenMapping.CompositeKey]; ok {
			childrenMapping.ParentIdentifier = mapping.Id
		}
	}

	if len(childrenMappings) > 0 {
		_, err := impl.qualifierMappingRepository.CreateQualifierMappings(childrenMappings, tx)
		if err != nil {
			impl.logger.Errorw("error in getting mappings", "err", err, "mappings", childrenMappings)
			return nil, err
		}
	}

	return maps.Values(mappingsToSelection), nil
}

func (impl QualifierMappingServiceImpl) CreateMappings(tx *pg.Tx, userId int32, resourceType ResourceType, resourceIds []int, qualifierSelector QualifierSelector, selectionIdentifiers []*SelectionIdentifier) error {
	mappings := make([]*ResourceMappingSelection, 0)
	for _, id := range resourceIds {
		for _, selectionIdentifier := range selectionIdentifiers {
			mapping := &ResourceMappingSelection{
				ResourceType:        resourceType,
				ResourceId:          id,
				QualifierSelector:   qualifierSelector,
				SelectionIdentifier: selectionIdentifier,
			}
			mappings = append(mappings, mapping)
		}
	}
	_, err := impl.CreateMappingsForSelections(tx, userId, mappings)
	return err
}

func (impl *QualifierMappingServiceImpl) filterAndGroupMappings(mappings []*QualifierMapping, selector QualifierSelector, composites mapset.Set) [][]*QualifierMapping {

	numQualifiers := GetNumOfChildQualifiers(selector.toQualifier())
	parentIdToChildMappings := make(map[int][]*QualifierMapping)
	parentIdToMapping := make(map[int]*QualifierMapping, 0)
	parentMappingIds := make([]int, 0)
	for _, mapping := range mappings {
		// is not parent so append it to the list in the map with key as its parent ID
		if mapping.ParentIdentifier > 0 {
			parentIdToChildMappings[mapping.ParentIdentifier] = append(parentIdToChildMappings[mapping.ParentIdentifier], mapping)
		} else {
			// is parent so collect IDs and put it in a map for easy retrieval
			parentMappingIds = append(parentMappingIds, mapping.Id)
			parentIdToMapping[mapping.Id] = mapping
		}
	}

	for parentMappingId, _ := range parentIdToChildMappings {
		// this deletes the keys in the map where the key does not exist in the collected IDs for parent
		if !slices.Contains(parentMappingIds, parentMappingId) {
			delete(parentIdToChildMappings, parentMappingId)
		}
	}

	groupedMappings := make([][]*QualifierMapping, 0)
	for parentId, childMappings := range parentIdToChildMappings {
		if len(childMappings) == numQualifiers {
			selectedParentMapping := parentIdToMapping[parentId]
			composite := getCompositeString(selectedParentMapping.IdentifierValueInt, childMappings[0].IdentifierValueInt)
			if composites.Cardinality() > 0 && !composites.Contains(composite) {
				break
			}
			mappingsGroup := []*QualifierMapping{selectedParentMapping}
			mappingsGroup = append(mappingsGroup, childMappings...)
			groupedMappings = append(groupedMappings, mappingsGroup)
		}
	}
	return groupedMappings
}

func (impl QualifierMappingServiceImpl) getAppEnvIdentifierFromGroup(group []*QualifierMapping) *SelectionIdentifier {
	resourceKeyToName := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyIdNameMap()
	var appId, envId int
	var appName, envName string
	for _, mapping := range group {
		field := resourceKeyToName[mapping.IdentifierKey]
		switch field {
		case bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_APP_ID:
			appId, appName = mapping.GetIdValueAndName()
		case bean.DEVTRON_RESOURCE_SEARCHABLE_KEY_ENV_ID:
			envId, envName = mapping.GetIdValueAndName()
		}
	}
	return getSelectionIdentifierForAppEnv(appId, envId, getIdentifierNamesForAppEnv(envName, appName))
}

func (impl QualifierMappingServiceImpl) getSelectionIdentifierForAppEnvSelector(mappingGroups [][]*QualifierMapping) map[int][]*SelectionIdentifier {

	resourceIdToIdentifier := make(map[int][]*SelectionIdentifier)
	for _, group := range mappingGroups {
		identifier := impl.getAppEnvIdentifierFromGroup(group)
		resourceId := group[0].ResourceId

		if _, ok := resourceIdToIdentifier[resourceId]; ok {
			resourceIdToIdentifier[resourceId] = append(resourceIdToIdentifier[resourceId], identifier)
		} else {
			resourceIdToIdentifier[resourceId] = []*SelectionIdentifier{identifier}
		}
	}
	return resourceIdToIdentifier
}

func (impl QualifierMappingServiceImpl) DeleteResourceMappingsForScopes(tx *pg.Tx, userId int32, resourceType ResourceType, qualifierSelector QualifierSelector, scopes []*SelectionIdentifier) error {
	if qualifierSelector != ApplicationEnvironmentSelector {
		return fmt.Errorf("selector currently not implemented")
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
	mappings, err := impl.qualifierMappingRepository.GetQualifierMappingsForListOfQualifierValues(resourceType, valuesMap, keyMap, []int{})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error fetching resource mappings %v %v", resourceType, valuesMap))
	}
	groups := impl.filterAndGroupMappings(mappings, qualifierSelector, mapset.NewSet())
	mappingIds := make([]int, 0, len(mappings))
	for _, group := range groups {
		for _, mapping := range group {
			mappingIds = append(mappingIds, mapping.Id)
		}
	}
	return impl.DeleteAllByIds(mappingIds, userId, tx)

}

func (impl QualifierMappingServiceImpl) GetResourceMappingsForSelections(resourceType ResourceType, qualifierSelector QualifierSelector, selectionIdentifiers []*SelectionIdentifier) ([]ResourceQualifierMappings, error) {
	if qualifierSelector != ApplicationEnvironmentSelector {
		return nil, fmt.Errorf("selector currently not implemented")
	}

	keyMap := impl.devtronResourceSearchableKeyService.GetAllSearchableKeyNameIdMap()

	valuesMap := make(map[Qualifier][][]int)
	appIds := make([]int, 0)
	envIds := make([]int, 0)
	for _, selectionIdentifier := range selectionIdentifiers {
		appIds = append(appIds, selectionIdentifier.AppId)
		envIds = append(envIds, selectionIdentifier.EnvId)
	}
	valuesMap[qualifierSelector.toQualifier()] = [][]int{appIds, envIds}
	mappings, err := impl.qualifierMappingRepository.GetQualifierMappingsForListOfQualifierValues(resourceType, valuesMap, keyMap, []int{})
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("error fetching resource mappings %v %v", resourceType, valuesMap))
	}

	return impl.processMappings(resourceType, mappings, qualifierSelector, getCompositeStringsAppEnvSelection(selectionIdentifiers))
}
func (impl QualifierMappingServiceImpl) GetResourceMappingsForResources(resourceType ResourceType, resourceIds []int, qualifierSelector QualifierSelector) ([]ResourceQualifierMappings, error) {
	mappings, err := impl.qualifierMappingRepository.GetMappingsByResourceTypeAndIdsAndQualifierId(resourceType, resourceIds, int(qualifierSelector.toQualifier()))
	if err != nil {
		return nil, err
	}

	return impl.processMappings(resourceType, mappings, qualifierSelector, mapset.NewSet())
}

func (impl QualifierMappingServiceImpl) processMappings(resourceType ResourceType, mappings []*QualifierMapping, qualifierSelector QualifierSelector, composites mapset.Set) ([]ResourceQualifierMappings, error) {
	groups := impl.filterAndGroupMappings(mappings, qualifierSelector, composites)
	if qualifierSelector != ApplicationEnvironmentSelector {
		return nil, fmt.Errorf("selector currently not implemented")
	}
	resourceIdToIdentifiers := impl.getSelectionIdentifierForAppEnvSelector(groups)

	qualifierMappings := make([]ResourceQualifierMappings, 0)

	for resourceId, identifier := range resourceIdToIdentifiers {
		for _, identifier := range identifier {
			qualifierMappings = append(qualifierMappings, ResourceQualifierMappings{
				ResourceId:          resourceId,
				ResourceType:        resourceType,
				SelectionIdentifier: identifier,
			})
		}
	}
	return qualifierMappings, nil
}

type QualifierMappingServiceImpl struct {
	logger                              *zap.SugaredLogger
	qualifierMappingRepository          QualifiersMappingRepository
	devtronResourceSearchableKeyService read.DevtronResourceSearchableKeyService
}

func NewQualifierMappingServiceImpl(logger *zap.SugaredLogger, qualifierMappingRepository QualifiersMappingRepository, devtronResourceSearchableKeyService read.DevtronResourceSearchableKeyService) (*QualifierMappingServiceImpl, error) {
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

func (impl QualifierMappingServiceImpl) GetQualifierMappingsByResourceId(resourceId int, resourceType ResourceType) ([]*QualifierMapping, error) {
	return impl.qualifierMappingRepository.GetQualifierMappingsByResourceId(resourceId, resourceType)
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

func (impl QualifierMappingServiceImpl) RollbackTx(tx *pg.Tx) error {
	return impl.qualifierMappingRepository.RollbackTx(tx)
}

func (impl QualifierMappingServiceImpl) CommitTx(tx *pg.Tx) error {
	return impl.qualifierMappingRepository.CommitTx(tx)
}

func (impl QualifierMappingServiceImpl) StartTx() (*pg.Tx, error) {
	return impl.qualifierMappingRepository.StartTx()
}
