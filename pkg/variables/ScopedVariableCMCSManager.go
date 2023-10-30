package variables

import (
	"context"
	"encoding/json"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	models2 "github.com/devtron-labs/devtron/pkg/variables/models"
	"github.com/devtron-labs/devtron/pkg/variables/parsers"
	repository1 "github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/devtron-labs/devtron/pkg/variables/utils"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
)

type ScopedVariableCMCSManager interface {
	ScopedVariableManager
	GetResolvedCMCSHistoryDtos(ctx context.Context, configType repository.ConfigType, configList bean2.ConfigList, history *repository.ConfigmapAndSecretHistory, secretList bean2.SecretList) (map[string]bean2.ConfigData, map[string]string, error)
	ResolveCMCSHistoryDto(ctx context.Context, configType repository.ConfigType, configList bean2.ConfigList, history *repository.ConfigmapAndSecretHistory, componentName string, secretList bean2.SecretList) (map[string]string, string, error)

	CreateVariableMappingsForCMApp(model *chartConfig.ConfigMapAppModel) error
	CreateVariableMappingsForCMEnv(model *chartConfig.ConfigMapEnvModel) error
	CreateVariableMappingsForSecretApp(model *chartConfig.ConfigMapAppModel) error
	CreateVariableMappingsForSecretEnv(model *chartConfig.ConfigMapEnvModel) error

	ResolveCMCSTrigger(cType bean.DeploymentConfigurationType, scope resourceQualifiers.Scope, configMapAppId int, configMapEnvId int, configMapByte []byte, secretDataByte []byte, configMapHistoryId int, secretHistoryId int) (string, string, map[string]string, map[string]string, error)
	ResolveCMCS(
		scope resourceQualifiers.Scope, configAppLevelId int,
		configEnvLevelId int,
		mergedConfigMap map[string]*bean2.ConfigData,
		mergedSecret map[string]*bean2.ConfigData) (map[string]*bean2.ConfigData, map[string]*bean2.ConfigData, map[string]string, map[string]string, error)

	ResolveForPrePostStageTrigger(scope resourceQualifiers.Scope, configResponse bean.ConfigMapJson, secretResponse bean.ConfigSecretJson, cmAppId int, cmEnvId int) (*bean.ConfigMapJson, *bean.ConfigSecretJson, error)
}

type ScopedVariableCMCSManagerImpl struct {
	ScopedVariableManagerImpl
}

func NewScopedVariableCMCSManagerImpl(logger *zap.SugaredLogger,
	scopedVariableService ScopedVariableService,
	variableEntityMappingService VariableEntityMappingService,
	variableSnapshotHistoryService VariableSnapshotHistoryService,
	variableTemplateParser parsers.VariableTemplateParser,
) (*ScopedVariableCMCSManagerImpl, error) {

	scopedVariableManagerImpl := ScopedVariableManagerImpl{
		logger:                         logger,
		scopedVariableService:          scopedVariableService,
		variableEntityMappingService:   variableEntityMappingService,
		variableSnapshotHistoryService: variableSnapshotHistoryService,
		variableTemplateParser:         variableTemplateParser,
	}
	scopedVariableCMCSManagerImpl := &ScopedVariableCMCSManagerImpl{
		ScopedVariableManagerImpl: scopedVariableManagerImpl,
	}

	return scopedVariableCMCSManagerImpl, nil
}

func (impl *ScopedVariableCMCSManagerImpl) ResolveCMCSHistoryDto(ctx context.Context, configType repository.ConfigType, configList bean2.ConfigList, history *repository.ConfigmapAndSecretHistory, componentName string, secretList bean2.SecretList) (map[string]string, string, error) {
	var variableSnapshotMap map[string]string
	var cMCSData map[string]bean2.ConfigData
	var err error
	if configType == repository.CONFIGMAP_TYPE {
		cMCSData, variableSnapshotMap, err = impl.ResolveCMHistoryDto(ctx, configList, history)
	} else if configType == repository.SECRET_TYPE {
		cMCSData, variableSnapshotMap, err = impl.ResolveSecretHistoryDto(ctx, secretList, history)
	}
	if err != nil {
		return nil, "", err
	}

	return variableSnapshotMap, string(cMCSData[componentName].Data), nil
}

func (impl *ScopedVariableCMCSManagerImpl) ResolveSecretHistoryDto(ctx context.Context, secretList bean2.SecretList, history *repository.ConfigmapAndSecretHistory) (map[string]bean2.ConfigData, map[string]string, error) {
	cMCSData := make(map[string]bean2.ConfigData, 0)
	secretListJson, err := json.Marshal(secretList)
	reference := repository1.HistoryReference{
		HistoryReferenceId:   history.Id,
		HistoryReferenceType: repository1.HistoryReferenceTypeSecret,
	}
	data, err := secretList.GetTransformedDataForSecret(string(secretListJson), util.DecodeSecret)
	if err != nil {
		return cMCSData, nil, err
	}
	isSuperAdmin, err := util.GetIsSuperAdminFromContext(ctx)

	variableSnapshotMap, resolvedTemplate, err := impl.GetVariableSnapshotAndResolveTemplate(data, reference, isSuperAdmin, false)
	if err != nil {
		return cMCSData, nil, err
	}
	resolvedTemplate, err = secretList.GetTransformedDataForSecret(resolvedTemplate, util.EncodeSecret)
	if err != nil {
		return cMCSData, nil, err
	}

	resolvedSecretList := bean2.SecretList{}
	err = json.Unmarshal([]byte(resolvedTemplate), &resolvedSecretList)
	if err != nil {
		return cMCSData, nil, err
	}
	for i, _ := range resolvedSecretList.ConfigData {
		cMCSData[resolvedSecretList.ConfigData[i].Name] = *resolvedSecretList.ConfigData[i]
	}
	return cMCSData, variableSnapshotMap, nil
}

func (impl *ScopedVariableCMCSManagerImpl) ResolveCMHistoryDto(ctx context.Context, configList bean2.ConfigList, history *repository.ConfigmapAndSecretHistory) (map[string]bean2.ConfigData, map[string]string, error) {
	cMCSData := make(map[string]bean2.ConfigData, 0)
	configListJson, err := json.Marshal(configList)
	reference := repository1.HistoryReference{
		HistoryReferenceId:   history.Id,
		HistoryReferenceType: repository1.HistoryReferenceTypeConfigMap,
	}
	isSuperAdmin, err := util.GetIsSuperAdminFromContext(ctx)
	if err != nil {
		return cMCSData, nil, err
	}
	variableSnapshotMap, resolvedTemplate, err := impl.GetVariableSnapshotAndResolveTemplate(string(configListJson), reference, isSuperAdmin, true)
	if err != nil {
		return cMCSData, nil, err
	}

	resolvedConfigList := bean2.ConfigList{}
	err = json.Unmarshal([]byte(resolvedTemplate), &resolvedConfigList)
	if err != nil {
		return cMCSData, nil, err
	}
	for i, _ := range resolvedConfigList.ConfigData {
		cMCSData[resolvedConfigList.ConfigData[i].Name] = *resolvedConfigList.ConfigData[i]
	}
	return cMCSData, variableSnapshotMap, nil
}

func (impl *ScopedVariableCMCSManagerImpl) GetResolvedCMCSHistoryDtos(ctx context.Context, configType repository.ConfigType, configList bean2.ConfigList, history *repository.ConfigmapAndSecretHistory, secretList bean2.SecretList) (map[string]bean2.ConfigData, map[string]string, error) {
	resolvedData := make(map[string]bean2.ConfigData, 0)
	var variableSnapshotMap map[string]string
	var err error
	if configType == repository.SECRET_TYPE {
		resolvedData, variableSnapshotMap, err = impl.ResolveCMHistoryDto(ctx, configList, history)
		if err != nil {
			return nil, nil, err
		}
	} else if configType == repository.CONFIGMAP_TYPE {
		resolvedData, variableSnapshotMap, err = impl.ResolveSecretHistoryDto(ctx, secretList, history)
		if err != nil {
			return nil, nil, err
		}
	}
	return resolvedData, variableSnapshotMap, nil
}

func (impl *ScopedVariableCMCSManagerImpl) CreateVariableMappingsForCMEnv(model *chartConfig.ConfigMapEnvModel) error {
	return impl.extractAndMapVariables(model.ConfigMapData, model.Id, repository1.EntityTypeConfigMapEnvLevel, model.UpdatedBy)
}
func (impl *ScopedVariableCMCSManagerImpl) CreateVariableMappingsForCMApp(model *chartConfig.ConfigMapAppModel) error {
	return impl.extractAndMapVariables(model.ConfigMapData, model.Id, repository1.EntityTypeConfigMapAppLevel, model.UpdatedBy)
}
func (impl *ScopedVariableCMCSManagerImpl) CreateVariableMappingsForSecretEnv(model *chartConfig.ConfigMapEnvModel) error {
	//VARIABLE_MAPPING_UPDATE
	sl := bean2.SecretList{}
	data, err := sl.GetTransformedDataForSecret(model.SecretData, util.DecodeSecret)
	if err != nil {
		return err
	}
	return impl.extractAndMapVariables(data, model.Id, repository1.EntityTypeSecretEnvLevel, model.UpdatedBy)
}
func (impl *ScopedVariableCMCSManagerImpl) CreateVariableMappingsForSecretApp(model *chartConfig.ConfigMapAppModel) error {
	//VARIABLE_MAPPING_UPDATE
	sl := bean2.SecretList{}
	data, err := sl.GetTransformedDataForSecret(model.SecretData, util.DecodeSecret)
	if err != nil {
		return err
	}
	return impl.extractAndMapVariables(data, model.Id, repository1.EntityTypeSecretAppLevel, model.UpdatedBy)
}

func (impl *ScopedVariableCMCSManagerImpl) extractAndMapVariables(template string, entityId int, entityType repository1.EntityType, userId int32) error {
	return impl.ExtractAndMapVariables(template, entityId, entityType, userId, nil)
}

func GetResolvedCMCSList(resolvedCS string, resolvedCM string) (map[string]*bean2.ConfigData, map[string]*bean2.ConfigData, error) {
	resolvedSecretList := map[string]*bean2.ConfigData{}
	err := json.Unmarshal([]byte(resolvedCS), &resolvedSecretList)
	if err != nil {
		return nil, nil, err
	}
	resolvedConfigList := map[string]*bean2.ConfigData{}
	err = json.Unmarshal([]byte(resolvedCM), &resolvedConfigList)
	if err != nil {
		return nil, nil, err
	}
	return resolvedSecretList, resolvedConfigList, nil
}

func (impl *ScopedVariableCMCSManagerImpl) ResolveCMCS(
	scope resourceQualifiers.Scope, configAppLevelId int,
	configEnvLevelId int,
	mergedConfigMap map[string]*bean2.ConfigData,
	mergedSecret map[string]*bean2.ConfigData) (map[string]*bean2.ConfigData, map[string]*bean2.ConfigData, map[string]string, map[string]string, error) {

	varNamesCM, varNamesCS, scopedVariables, err := impl.getScopedAndCollectVarNames(scope, configAppLevelId, configEnvLevelId)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	var resolvedTemplateCM, encodedSecretData string
	var variableMapCM, variableMapCS map[string]string

	mergedConfigMapJson, err := json.Marshal(mergedConfigMap)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	parserRequest := parsers.CreateParserRequest(string(mergedConfigMapJson), parsers.JsonVariableTemplate, scopedVariables, true)
	resolvedTemplateCM, err = impl.ParseTemplateWithScopedVariables(parserRequest)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	variableMapCM = parsers.GetVariableMapForUsedVariables(scopedVariables, varNamesCM)

	configData := bean2.ConfigData{}
	mergedSecretJson, err := json.Marshal(mergedSecret)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	decodedSecrets, err := configData.GetTransformedDataForSecretData(string(mergedSecretJson), util.DecodeSecret)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	parserRequest = parsers.CreateParserRequest(decodedSecrets, parsers.JsonVariableTemplate, scopedVariables, true)
	resolvedTemplateCS, err := impl.ParseTemplateWithScopedVariables(parserRequest)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	variableMapCS = parsers.GetVariableMapForUsedVariables(scopedVariables, varNamesCS)
	encodedSecretData, err = configData.GetTransformedDataForSecretData(resolvedTemplateCS, util.EncodeSecret)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	resolvedSecretList, resolvedConfigList, err := GetResolvedCMCSList(resolvedTemplateCM, encodedSecretData)

	return resolvedConfigList, resolvedSecretList, variableMapCM, variableMapCS, nil
}

func (impl *ScopedVariableCMCSManagerImpl) getScopedAndCollectVarNames(scope resourceQualifiers.Scope, configMapAppId int, configMapEnvId int) ([]string, []string, []*models2.ScopedVariableData, error) {
	varNamesCM := make([]string, 0)
	varNamesCS := make([]string, 0)
	entitiesForCM := util.GetBeans(
		repository1.GetEntity(configMapAppId, repository1.EntityTypeConfigMapAppLevel),
		repository1.GetEntity(configMapEnvId, repository1.EntityTypeConfigMapEnvLevel),
	)
	entitiesForCS := util.GetBeans(
		repository1.GetEntity(configMapAppId, repository1.EntityTypeSecretAppLevel),
		repository1.GetEntity(configMapEnvId, repository1.EntityTypeSecretEnvLevel),
	)

	entityToVariables, err := impl.GetEntityToVariableMapping(append(entitiesForCS, entitiesForCM...))
	if err != nil {
		return varNamesCM, varNamesCS, nil, err
	}
	varNamesCM = repository1.CollectVariables(entityToVariables, entitiesForCM)
	varNamesCS = repository1.CollectVariables(entityToVariables, entitiesForCS)
	usedVariablesInCMCS := utils.FilterDuplicatesInStringArray(append(varNamesCM, varNamesCS...))
	scopedVariables, err := impl.GetScopedVariables(scope, usedVariablesInCMCS, true)
	return varNamesCM, varNamesCS, scopedVariables, nil
}

func (impl *ScopedVariableCMCSManagerImpl) ResolvedVariableForLastSaved(scope resourceQualifiers.Scope, configMapAppId int, configMapEnvId int, configMapByte []byte, secretDataByte []byte) (string, string, map[string]string, map[string]string, error) {
	var resolvedCS, resolvedCM string
	var variableSnapshotForCM, variableSnapshotForCS map[string]string
	varNamesCM, varNamesCS, scopedVariables, err := impl.getScopedAndCollectVarNames(scope, configMapAppId, configMapEnvId)
	if err != nil {
		return string(configMapByte), string(secretDataByte), variableSnapshotForCM, variableSnapshotForCS, err
	}

	if configMapByte != nil && len(varNamesCM) > 0 {
		parserRequest := parsers.CreateParserRequest(string(configMapByte), parsers.JsonVariableTemplate, scopedVariables, false)
		resolvedCM, err = impl.ParseTemplateWithScopedVariables(parserRequest)
		if err != nil {
			return resolvedCM, string(secretDataByte), variableSnapshotForCM, variableSnapshotForCS, err
		}
		variableSnapshotForCM = parsers.GetVariableMapForUsedVariables(scopedVariables, varNamesCM)
	}

	if secretDataByte != nil && len(varNamesCS) > 0 {
		ab := bean.ConfigSecretRootJson{}
		data, err := ab.GetTransformedDataForSecretData(string(secretDataByte), util.DecodeSecret)
		if err != nil {
			return resolvedCM, string(secretDataByte), variableSnapshotForCM, variableSnapshotForCS, err
		}
		parserRequest := parsers.CreateParserRequest(data, parsers.JsonVariableTemplate, scopedVariables, false)
		resolvedCSDecoded, err := impl.ParseTemplateWithScopedVariables(parserRequest)
		variableSnapshotForCS = parsers.GetVariableMapForUsedVariables(scopedVariables, varNamesCS)
		resolvedCS, err = ab.GetTransformedDataForSecretData(resolvedCSDecoded, util.EncodeSecret)
		if err != nil {
			return resolvedCM, resolvedCM, variableSnapshotForCM, variableSnapshotForCS, err
		}
	}
	resolvedCMs, resolvedCSs, ok := resolvedCMCS(resolvedCM, resolvedCS, secretDataByte, configMapByte)
	if ok {
		return resolvedCMs, resolvedCSs, variableSnapshotForCM, variableSnapshotForCS, nil
	}

	return resolvedCM, resolvedCS, variableSnapshotForCM, variableSnapshotForCS, nil
}

func resolvedCMCS(resolvedCM string, resolvedCS string, secretDataByte []byte, configMapByte []byte) (string, string, bool) {
	if resolvedCM != "" && resolvedCS == "" {
		return resolvedCM, string(secretDataByte), true
	} else if resolvedCM == "" && resolvedCS != "" {
		return string(configMapByte), resolvedCS, true
	} else if resolvedCM == "" && resolvedCS == "" {
		return string(configMapByte), string(secretDataByte), true
	}
	return "", "", false
}

func (impl *ScopedVariableCMCSManagerImpl) ResolvedVariableForSpecificType(configMapHistoryId int, secretHistoryId int, configMapByte []byte, secretDataByte []byte) (string, string, map[string]string, map[string]string, error) {

	reference := repository1.HistoryReference{
		HistoryReferenceId:   configMapHistoryId,
		HistoryReferenceType: repository1.HistoryReferenceTypeConfigMap,
	}

	variableMapCM, resolvedTemplateCM, err := impl.GetVariableSnapshotAndResolveTemplate(string(configMapByte), reference, true, false)
	if err != nil {
		return "", "", nil, nil, err
	}
	reference = repository1.HistoryReference{
		HistoryReferenceId:   secretHistoryId,
		HistoryReferenceType: repository1.HistoryReferenceTypeSecret,
	}
	ab := bean.ConfigSecretRootJson{}
	data, err := ab.GetTransformedDataForSecretData(string(secretDataByte), util.DecodeSecret)
	if err != nil {
		return "", "", nil, nil, err
	}
	variableMapCS, resolvedTemplateCS, err := impl.GetVariableSnapshotAndResolveTemplate(data, reference, true, false)
	encodedSecretData, err := ab.GetTransformedDataForSecretData(resolvedTemplateCS, util.EncodeSecret)
	if err != nil {
		return "", "", nil, nil, err
	}
	return resolvedTemplateCM, encodedSecretData, variableMapCM, variableMapCS, nil
}

func (impl *ScopedVariableCMCSManagerImpl) ResolveCMCSTrigger(cType bean.DeploymentConfigurationType, scope resourceQualifiers.Scope, configMapAppId int, configMapEnvId int, configMapByte []byte, secretDataByte []byte, configMapHistoryId int, secretHistoryId int) (string, string, map[string]string, map[string]string, error) {
	var resolvedCM, resolvedCS string
	var cmSnapshot, csSnapshot map[string]string
	var err error
	if cType == bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED {
		resolvedCM, resolvedCS, cmSnapshot, csSnapshot, err = impl.ResolvedVariableForLastSaved(scope, configMapAppId, configMapEnvId, configMapByte, secretDataByte)
	}
	if cType == bean.DEPLOYMENT_CONFIG_TYPE_SPECIFIC_TRIGGER {
		resolvedCM, resolvedCS, cmSnapshot, csSnapshot, err = impl.ResolvedVariableForSpecificType(configMapHistoryId, secretHistoryId, configMapByte, secretDataByte)
	}
	if err != nil {
		return "", "", nil, nil, err
	}
	return resolvedCM, resolvedCS, cmSnapshot, csSnapshot, nil
}

func (impl *ScopedVariableCMCSManagerImpl) ResolveForPrePostStageTrigger(scope resourceQualifiers.Scope, configResponse bean.ConfigMapJson, secretResponse bean.ConfigSecretJson, cmAppId int, cmEnvId int) (*bean.ConfigMapJson, *bean.ConfigSecretJson, error) {

	configResponseR := bean.ConfigMapRootJson{ConfigMapJson: configResponse}
	secretResponseR := bean.ConfigSecretRootJson{ConfigSecretJson: secretResponse}
	configMapByte, err := json.Marshal(configResponseR)
	if err != nil {
		return nil, nil, err
	}
	secretDataByte, err := json.Marshal(secretResponseR)
	if err != nil {
		return nil, nil, err

	}

	resolvedCM, resolvedCS, _, _, err := impl.ResolvedVariableForLastSaved(scope, cmAppId, cmEnvId, configMapByte, secretDataByte)

	var configResponseResolved bean.ConfigMapRootJson
	var secretResponseResolved bean.ConfigSecretRootJson
	err = json.Unmarshal([]byte(resolvedCM), &configResponseResolved)
	if err != nil {
		return nil, nil, err
	}
	err = json.Unmarshal([]byte(resolvedCS), &secretResponseResolved)
	if err != nil {
		return nil, nil, err
	}

	return &configResponseResolved.ConfigMapJson, &secretResponseResolved.ConfigSecretJson, nil
}
