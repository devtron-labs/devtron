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
)

type ScopedVariableManager interface {

	// pass throughs
	GetScopedVariables(scope resourceQualifiers.Scope, varNames []string, unmaskSensitiveData bool) (scopedVariableDataObj []*models.ScopedVariableData, err error)
	GetEntityToVariableMapping(entity []repository.Entity) (map[repository.Entity][]string, error)
	SaveVariableHistoriesForTrigger(variableHistories []*repository.VariableSnapshotHistoryBean, userId int32) error
	RemoveMappedVariables(entityId int, entityType repository.EntityType, userId int32, tx *pg.Tx) error
	ParseTemplateWithScopedVariables(request parsers.VariableParserRequest) (string, error)

	// variable mapping
	ExtractAndMapVariables(template string, entityId int, entityType repository.EntityType, userId int32, tx *pg.Tx) error

	// template resolvers
	ExtractVariablesAndResolveTemplate(scope resourceQualifiers.Scope, template string, templateType parsers.VariableTemplateType, unmaskSensitiveData bool, maskUnknownVariable bool) (string, map[string]string, error)
	GetMappedVariablesAndResolveTemplate(template string, scope resourceQualifiers.Scope, entity repository.Entity, unmaskSensitiveData bool) (string, map[string]string, error)
	GetMappedVariablesAndResolveTemplateBatch(template string, scope resourceQualifiers.Scope, entities []repository.Entity) (string, map[string]string, error)
	GetVariableSnapshotAndResolveTemplate(template string, templateType parsers.VariableTemplateType, reference repository.HistoryReference, isSuperAdmin bool, ignoreUnknown bool) (map[string]string, string, error)
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

	scopedVariables, err := impl.scopedVariableService.GetScopedVariables(scope, entityToVariables[entity], isSuperAdmin)
	if err != nil {
		return template, variableMap, err
	}

	for _, variable := range scopedVariables {
		variableMap[variable.VariableName] = variable.VariableValue.StringValue()
	}

	if len(variableMap) == 0 {
		return template, variableMap, nil
	}

	parserRequest := parsers.VariableParserRequest{Template: template, Variables: scopedVariables, TemplateType: parsers.JsonVariableTemplate}

	resolvedTemplate, err := impl.ParseTemplateWithScopedVariables(parserRequest)
	if err != nil {
		return template, variableMap, err
	}

	return resolvedTemplate, variableMap, nil
}

func (impl ScopedVariableManagerImpl) ExtractVariablesAndResolveTemplate(scope resourceQualifiers.Scope, template string, templateType parsers.VariableTemplateType, isSuperAdmin bool, maskUnknownVariable bool) (string, map[string]string, error) {

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
	resolvedTemplate, err := impl.ParseTemplateWithScopedVariables(parserRequest)
	if err != nil {
		return template, variableSnapshot, err
	}

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

func (impl ScopedVariableManagerImpl) GetVariableSnapshotAndResolveTemplate(template string, templateType parsers.VariableTemplateType, reference repository.HistoryReference, isSuperAdmin bool, ignoreUnknown bool) (map[string]string, string, error) {
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
	request := parsers.VariableParserRequest{Template: template, TemplateType: templateType, Variables: scopedVariableData, IgnoreUnknownVariables: ignoreUnknown}

	resolvedTemplate, err := impl.ParseTemplateWithScopedVariables(request)
	if err != nil {
		return variableSnapshotMap, resolvedTemplate, err
	}

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

func (impl ScopedVariableManagerImpl) GetMappedVariablesAndResolveTemplateBatch(template string, scope resourceQualifiers.Scope, entities []repository.Entity) (string, map[string]string, error) {

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

	request := parsers.VariableParserRequest{
		TemplateType:           parsers.StringVariableTemplate,
		Template:               template,
		Variables:              scopedVariables,
		IgnoreUnknownVariables: true,
	}
	resolvedTemplate, err := impl.ParseTemplateWithScopedVariables(request)
	if err != nil {
		impl.logger.Errorw("Error in parsing stage", "error", err, "template", template, "vars", scopedVariables)
		return template, variableMap, err
	}

	return resolvedTemplate, variableSnapshot, nil
}

func (impl ScopedVariableManagerImpl) GetEntityToVariableMapping(entity []repository.Entity) (map[repository.Entity][]string, error) {
	entityToVariables, err := impl.variableEntityMappingService.GetAllMappingsForEntities(entity)
	return entityToVariables, err
}

func (impl ScopedVariableManagerImpl) GetScopedVariables(scope resourceQualifiers.Scope, varNames []string, unmaskSensitiveData bool) (scopedVariableDataObj []*models.ScopedVariableData, err error) {
	return impl.scopedVariableService.GetScopedVariables(scope, varNames, unmaskSensitiveData)
}

func (impl ScopedVariableManagerImpl) ParseTemplateWithScopedVariables(request parsers.VariableParserRequest) (string, error) {

	parserResponse := impl.variableTemplateParser.ParseTemplate(request)
	err := parserResponse.Error
	if err != nil {
		return request.Template, err
	}

	resolvedTemplate := parserResponse.ResolvedTemplate
	return resolvedTemplate, nil
}
