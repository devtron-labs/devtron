package config

import (
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/go-pg/pg"
)

type configFactory[T any] interface {
	validate(platformConfigurations, defaultConfigurations []*v1.ConfigurationBean) error
	getConfigKeys() []v1.ConfigKeyStr
	getSupportedUnits() map[v1.ConfigKeyStr]map[string]v1.Unit

	getInfraConfigEntities(infraConfig *v1.InfraConfig, profileId int, platformName string) ([]*repository.InfraProfileConfigurationEntity, error)
	getValueFromString(valueString string) (T, int, error)
	isConfigActive(valueCount int, configActive bool) bool
	getValueFromBean(configurationBean *v1.ConfigurationBean) (T, error)
	formatTypedValueAsString(configValue any) (string, error)
	overrideInfraConfig(infraConfiguration *v1.InfraConfig, configurationBean *v1.ConfigurationBean) (*v1.InfraConfig, error)
	getAppliedConfiguration(key v1.ConfigKeyStr, infraConfiguration *v1.ConfigurationBean, defaultConfigurations []*v1.ConfigurationBean) (*v1.ConfigurationBean, error)
	handlePostCreateOperations(tx *pg.Tx, createdInfraConfig *repository.InfraProfileConfigurationEntity) error
	handlePostUpdateOperations(tx *pg.Tx, updatedInfraConfig *repository.InfraProfileConfigurationEntity) error
	handlePostDeleteOperations(tx *pg.Tx, deletedInfraConfig *repository.InfraProfileConfigurationEntity) error
	handleInfraConfigTriggerAudit(workflowId int, triggeredBy int32, infraConfig *v1.InfraConfig) error
	resolveScopeVariablesForAppliedConfiguration(scope resourceQualifiers.Scope, configuration *v1.ConfigurationBean) (*v1.ConfigurationBean, map[string]string, error)
}
