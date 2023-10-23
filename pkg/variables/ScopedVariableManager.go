package variables

import (
	"encoding/json"
	mapset "github.com/deckarep/golang-set"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/variables/models"
	"github.com/devtron-labs/devtron/pkg/variables/parsers"
	"github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/devtron-labs/devtron/pkg/variables/utils"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type ScopedVariableManager interface {
	GetScopedVariables(scope resourceQualifiers.Scope, varNames []string, unmaskSensitiveData bool) (scopedVariableDataObj []*models.ScopedVariableData, err error)
	GetVariableMapForUsedVariables(scopedVariables []*models.ScopedVariableData, usedVars []string) map[string]string
	GetEntityToVariableMapping(entity []repository.Entity) (map[repository.Entity][]string, error)
	SaveVariableHistoriesForTrigger(variableHistories []*repository.VariableSnapshotHistoryBean, userId int32) error

	GetVariableSnapshot(reference []repository.HistoryReference) (map[repository.HistoryReference]*repository.VariableSnapshotHistoryBean, error)
	GetResolvedTemplateWithSnapshot(template string, reference repository.HistoryReference) (string, map[string]string, error)
	ParseTemplateWithScopedVariables(template string, scopedVariables []*models.ScopedVariableData) (string, error)
	ExtractVariablesAndResolveTemplateAppService(scope resourceQualifiers.Scope, template string, entity repository.Entity) (string, map[string]string, error)

	GetMappedVariablesAndResolveTemplate(template string, scope resourceQualifiers.Scope, entity repository.Entity, isSuperAdmin bool) (string, map[string]string, error)
	GetMappedVariablesAndResolveTemplateBatch(template string, entities []repository.Entity, scope resourceQualifiers.Scope) (string, map[string]string, error)
	ExtractVariablesAndResolveTemplate(scope resourceQualifiers.Scope, template string, templateType parsers.VariableTemplateType, isSuperAdmin bool, maskUnknownVariable bool) (string, map[string]string, error)
	ExtractAndMapVariables(template string, entityId int, entityType repository.EntityType, userId int32, tx *pg.Tx) error
	GetVariableSnapshotAndResolveTemplate(template string, reference repository.HistoryReference, isSuperAdmin bool) (map[string]string, string, error)
	RemoveMappedVariables(entityId int, entityType repository.EntityType, userId int32, tx *pg.Tx) error
}

func (impl ScopedVariableManagerImpl) SaveVariableHistoriesForTrigger(variableHistories []*repository.VariableSnapshotHistoryBean, userId int32) error {
	return impl.variableSnapshotHistoryService.SaveVariableHistoriesForTrigger(variableHistories, userId)
}

type ScopedVariableManagerImpl struct {
	logger                         *zap.SugaredLogger
	scopedVariableService          ScopedVariableService
	variableEntityMappingService   VariableEntityMappingService
	variableSnapshotHistoryService VariableSnapshotHistoryService
	variableTemplateParser         parsers.VariableTemplateParser
}

func NewScopedVariableManagerImpl(logger *zap.SugaredLogger,
	scopedVariableService ScopedVariableService,
	variableEntityMappingService VariableEntityMappingService,
	variableSnapshotHistoryService VariableSnapshotHistoryService,
	variableTemplateParser parsers.VariableTemplateParser,
) (*ScopedVariableManagerImpl, error) {

	scopedVariableManagerImpl := &ScopedVariableManagerImpl{
		logger:                         logger,
		scopedVariableService:          scopedVariableService,
		variableEntityMappingService:   variableEntityMappingService,
		variableSnapshotHistoryService: variableSnapshotHistoryService,
		variableTemplateParser:         variableTemplateParser,
	}

	return scopedVariableManagerImpl, nil
}

func (impl ScopedVariableManagerImpl) GetMappedVariablesAndResolveTemplate(template string, scope resourceQualifiers.Scope, entity repository.Entity, isSuperAdmin bool) (string, map[string]string, error) {

	variableMap := make(map[string]string)
	entityToVariables, err := impl.variableEntityMappingService.GetAllMappingsForEntities([]repository.Entity{entity})
	if err != nil {
		return template, variableMap, err
	}
	if vars, ok := entityToVariables[entity]; !ok || len(vars) == 0 {
		return template, variableMap, nil
	}

	// pre-populating variable map with variable so that the variables which don't have any resolved data
	// is saved in snapshot
	for _, variable := range entityToVariables[entity] {
		variableMap[variable] = impl.scopedVariableService.GetFormattedVariableForName(variable)
	}

	//scopedVariables := make([]*models.ScopedVariableData, 0)
	//if _, ok := entityToVariables[entity]; ok && len(entityToVariables[entity]) > 0 {
	scopedVariables, err := impl.scopedVariableService.GetScopedVariables(scope, entityToVariables[entity], isSuperAdmin)
	if err != nil {
		return template, variableMap, err
	}
	//}

	for _, variable := range scopedVariables {
		variableMap[variable.VariableName] = variable.VariableValue.StringValue()
	}

	if len(variableMap) == 0 {
		return template, variableMap, nil
	}

	parserRequest := parsers.VariableParserRequest{Template: template, Variables: scopedVariables, TemplateType: parsers.JsonVariableTemplate}
	parserResponse := impl.variableTemplateParser.ParseTemplate(parserRequest)
	err = parserResponse.Error
	if err != nil {
		return template, variableMap, err
	}
	resolvedTemplate := parserResponse.ResolvedTemplate

	return resolvedTemplate, variableMap, nil
}

func (impl ScopedVariableManagerImpl) ExtractVariablesAndResolveTemplate(scope resourceQualifiers.Scope, template string, templateType parsers.VariableTemplateType, isSuperAdmin bool, maskUnknownVariable bool) (string, map[string]string, error) {
	//Todo Subhashish manager layer
	variableSnapshot := make(map[string]string)
	usedVariables, err := impl.variableTemplateParser.ExtractVariables(template, templateType)
	if err != nil {
		return template, variableSnapshot, err
	}

	if len(usedVariables) == 0 {
		return template, variableSnapshot, err
	}

	scopedVariables, err := impl.scopedVariableService.GetScopedVariables(scope, usedVariables, isSuperAdmin)
	if err != nil {
		return template, variableSnapshot, err
	}

	for _, variable := range scopedVariables {
		variableSnapshot[variable.VariableName] = variable.VariableValue.StringValue()
	}

	if maskUnknownVariable {
		for _, variable := range usedVariables {
			if _, ok := variableSnapshot[variable]; !ok {
				scopedVariables = append(scopedVariables, &models.ScopedVariableData{
					VariableName:  variable,
					VariableValue: &models.VariableValue{Value: models.UndefinedValue},
				})
			}
		}
	}

	parserRequest := parsers.VariableParserRequest{Template: template, Variables: scopedVariables, TemplateType: templateType, IgnoreUnknownVariables: true}
	parserResponse := impl.variableTemplateParser.ParseTemplate(parserRequest)
	err = parserResponse.Error
	if err != nil {
		return template, variableSnapshot, err
	}
	resolvedTemplate := parserResponse.ResolvedTemplate
	return resolvedTemplate, variableSnapshot, nil
}

func (impl ScopedVariableManagerImpl) ExtractAndMapVariables(template string, entityId int, entityType repository.EntityType, userId int32, tx *pg.Tx) error {
	usedVariables, err := impl.variableTemplateParser.ExtractVariables(template, parsers.JsonVariableTemplate)
	if err != nil {
		return err
	}
	err = impl.variableEntityMappingService.UpdateVariablesForEntity(usedVariables, repository.Entity{
		EntityType: entityType,
		EntityId:   entityId,
	}, userId, tx)
	if err != nil {
		return err
	}
	return nil
}

func (impl ScopedVariableManagerImpl) GetVariableSnapshotAndResolveTemplate(template string, reference repository.HistoryReference, isSuperAdmin bool) (map[string]string, string, error) {
	variableSnapshotMap := make(map[string]string)
	references, err := impl.variableSnapshotHistoryService.GetVariableHistoryForReferences([]repository.HistoryReference{reference})
	if err != nil {
		return variableSnapshotMap, template, err
	}

	if _, ok := references[reference]; ok {
		err = json.Unmarshal(references[reference].VariableSnapshot, &variableSnapshotMap)
		if err != nil {
			return variableSnapshotMap, template, err
		}
	}

	if len(variableSnapshotMap) == 0 {
		return variableSnapshotMap, template, err
	}

	varNames := make([]string, 0)
	for varName, _ := range variableSnapshotMap {
		varNames = append(varNames, varName)
	}
	varNameToIsSensitive, err := impl.scopedVariableService.CheckForSensitiveVariables(varNames)
	if err != nil {
		return variableSnapshotMap, template, err
	}

	scopedVariableData := parsers.GetScopedVarData(variableSnapshotMap, varNameToIsSensitive, isSuperAdmin)
	request := parsers.VariableParserRequest{Template: template, TemplateType: parsers.JsonVariableTemplate, Variables: scopedVariableData, IgnoreUnknownVariables: true}
	parserResponse := impl.variableTemplateParser.ParseTemplate(request)
	err = parserResponse.Error
	if err != nil {
		return variableSnapshotMap, template, err
	}
	resolvedTemplate := parserResponse.ResolvedTemplate

	return variableSnapshotMap, resolvedTemplate, nil
}

func (impl ScopedVariableManagerImpl) RemoveMappedVariables(entityId int, entityType repository.EntityType, userId int32, tx *pg.Tx) error {

	err := impl.variableEntityMappingService.DeleteMappingsForEntities([]repository.Entity{{
		EntityType: entityType,
		EntityId:   entityId,
	}}, userId, tx)
	if err != nil {
		return err
	}
	return nil
}

func (impl ScopedVariableManagerImpl) GetMappedVariablesAndResolveTemplateBatch(template string, entities []repository.Entity, scope resourceQualifiers.Scope) (string, map[string]string, error) {

	//entities := make([]repository.Entity, 0)
	//for _, stageId := range pipelineStageIds {
	//	entities = append(entities, repository.Entity{
	//		EntityType: repository.EntityTypePipelineStage,
	//		EntityId:   stageId,
	//	})
	//}
	variableMap := make(map[string]string)
	mappingsForEntities, err := impl.variableEntityMappingService.GetAllMappingsForEntities(entities)
	if err != nil {
		impl.logger.Errorw("Error in fetching mapped variables in request", "error", err)
		return template, variableMap, err
	}

	//early exit if no variables found
	if len(mappingsForEntities) == 0 {
		return template, variableMap, nil
	}

	// collecting all unique variable names in a stage
	varNamesSet := mapset.NewSet()
	for _, variableNames := range mappingsForEntities {
		for _, variableName := range variableNames {
			varNamesSet.Add(variableName)
		}
	}
	varNames := utils.ToStringArray(varNamesSet.ToSlice())

	scopedVariables, err := impl.scopedVariableService.GetScopedVariables(scope, varNames, true)
	if err != nil {
		return template, variableMap, err
	}

	variableSnapshot := make(map[string]string)
	for _, variable := range scopedVariables {
		variableSnapshot[variable.VariableName] = variable.VariableValue.StringValue()
	}

	//responseJson, err := json.Marshal(unresolvedResponse)
	//if err != nil {
	//	impl.logger.Errorw("Error in marshaling stage", "error", err, "unresolvedResponse", unresolvedResponse)
	//	return nil, err
	//}
	parserResponse := impl.variableTemplateParser.ParseTemplate(parsers.VariableParserRequest{
		TemplateType:           parsers.StringVariableTemplate,
		Template:               template,
		Variables:              scopedVariables,
		IgnoreUnknownVariables: true,
	})
	err = parserResponse.Error
	if err != nil {
		impl.logger.Errorw("Error in parsing stage", "error", err, "template", template, "vars", scopedVariables)
		return template, variableMap, err
	}
	//resolvedResponse := &bean.PrePostAndRefPluginStepsResponse{}
	//err = json.Unmarshal([]byte(parserResponse.ResolvedTemplate), resolvedResponse)
	//if err != nil {
	//	impl.logger.Errorw("Error in unmarshalling stage", "error", err)
	//
	//	return template, err
	//}
	//resolvedResponse.VariableSnapshot = variableSnapshot
	return parserResponse.ResolvedTemplate, variableSnapshot, nil
}

func (impl ScopedVariableManagerImpl) ExtractVariablesAndResolveTemplateAppService(scope resourceQualifiers.Scope, template string, entity repository.Entity) (string, map[string]string, error) {

	variableMap := make(map[string]string)
	entities := []repository.Entity{entity}
	entityToVariables, err := impl.GetEntityToVariableMapping(entities)
	if err != nil {
		return template, variableMap, err
	}
	if vars, ok := entityToVariables[entity]; !ok || len(vars) == 0 {
		return template, variableMap, nil
	}

	// pre-populating variable map with variable so that the variables which don't have any resolved data
	// is saved in snapshot
	for _, variable := range entityToVariables[entity] {
		variableMap[variable] = impl.scopedVariableService.GetFormattedVariableForName(variable)
	}

	scopedVariables, err := impl.scopedVariableService.GetScopedVariables(scope, entityToVariables[entity], true)
	if err != nil {
		return template, variableMap, err
	}

	for _, variable := range scopedVariables {
		variableMap[variable.VariableName] = variable.VariableValue.StringValue()
	}

	parserRequest := parsers.VariableParserRequest{Template: template, Variables: scopedVariables, TemplateType: parsers.JsonVariableTemplate}
	parserResponse := impl.variableTemplateParser.ParseTemplate(parserRequest)
	err = parserResponse.Error
	if err != nil {
		return template, variableMap, err
	}
	resolvedTemplate := parserResponse.ResolvedTemplate
	return resolvedTemplate, variableMap, nil
}

func (impl ScopedVariableManagerImpl) GetEntityToVariableMapping(entity []repository.Entity) (map[repository.Entity][]string, error) {
	entityToVariables, err := impl.variableEntityMappingService.GetAllMappingsForEntities(entity)
	return entityToVariables, err
}

func (impl ScopedVariableManagerImpl) GetScopedVariables(scope resourceQualifiers.Scope, varNames []string, unmaskSensitiveData bool) (scopedVariableDataObj []*models.ScopedVariableData, err error) {
	return impl.scopedVariableService.GetScopedVariables(scope, varNames, unmaskSensitiveData)
}

func (impl ScopedVariableManagerImpl) GetVariableMapForUsedVariables(scopedVariables []*models.ScopedVariableData, usedVars []string) map[string]string {
	variableMap := make(map[string]string)
	for _, variable := range scopedVariables {
		if slices.Contains(usedVars, variable.VariableName) {
			variableMap[variable.VariableName] = variable.VariableValue.StringValue()
		}
	}
	return variableMap
}

func (impl ScopedVariableManagerImpl) ParseTemplateWithScopedVariables(template string, scopedVariables []*models.ScopedVariableData) (string, error) {

	parserRequest := parsers.VariableParserRequest{Template: template, Variables: scopedVariables, TemplateType: parsers.JsonVariableTemplate}
	parserResponse := impl.variableTemplateParser.ParseTemplate(parserRequest)
	err := parserResponse.Error
	if err != nil {
		return template, err
	}

	resolvedTemplate := parserResponse.ResolvedTemplate
	return resolvedTemplate, nil
}

func (impl ScopedVariableManagerImpl) GetResolvedTemplateWithSnapshot(template string, reference repository.HistoryReference) (string, map[string]string, error) {

	variableSnapshotMap := make(map[string]string)
	references := []repository.HistoryReference{reference}
	variableSnapshot, err := impl.GetVariableSnapshot(references)
	if err != nil {
		return template, variableSnapshotMap, err
	}

	if _, ok := variableSnapshot[reference]; !ok {
		return template, variableSnapshotMap, nil
	}

	err = json.Unmarshal(variableSnapshot[reference].VariableSnapshot, &variableSnapshotMap)
	if err != nil {
		return template, variableSnapshotMap, err
	}

	if len(variableSnapshotMap) == 0 {
		return template, variableSnapshotMap, nil
	}
	scopedVariableData := parsers.GetScopedVarData(variableSnapshotMap, make(map[string]bool), true)
	request := parsers.VariableParserRequest{Template: template, TemplateType: parsers.JsonVariableTemplate, Variables: scopedVariableData}
	parserResponse := impl.variableTemplateParser.ParseTemplate(request)
	err = parserResponse.Error
	if err != nil {
		return template, variableSnapshotMap, err
	}
	resolvedTemplate := parserResponse.ResolvedTemplate
	return resolvedTemplate, variableSnapshotMap, nil
}

func (impl ScopedVariableManagerImpl) GetVariableSnapshot(reference []repository.HistoryReference) (map[repository.HistoryReference]*repository.VariableSnapshotHistoryBean, error) {
	variableSnapshot, err := impl.variableSnapshotHistoryService.GetVariableHistoryForReferences(reference)
	return variableSnapshot, err
}
