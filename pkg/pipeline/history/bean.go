package history

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/util"
	"time"
)

type HistoryComponent string

const (
	DEPLOYMENT_TEMPLATE_TYPE_HISTORY_COMPONENT HistoryComponent = "DEPLOYMENT_TEMPLATE"
	CONFIGMAP_TYPE_HISTORY_COMPONENT           HistoryComponent = "CONFIGMAP"
	SECRET_TYPE_HISTORY_COMPONENT              HistoryComponent = "SECRET"
	PIPELINE_STRATEGY_TYPE_HISTORY_COMPONENT   HistoryComponent = "PIPELINE_STRATEGY"
)

type ComponentLevelHistoryDetailDto struct {
	ComponentName string            `json:"componentName"`
	HistoryConfig *HistoryDetailDto `json:"config"`
}

type AllDeploymentConfigurationDetail struct {
	DeploymentTemplateConfig *HistoryDetailDto                 `json:"deploymentTemplate"`
	ConfigMapConfig          []*ComponentLevelHistoryDetailDto `json:"configMap"`
	SecretConfig             []*ComponentLevelHistoryDetailDto `json:"secret"`
	StrategyConfig           *HistoryDetailDto                 `json:"pipelineStrategy"`
	WfrId                    int                               `json:"wfrId"`
}

type DeploymentConfigurationDto struct {
	Id                  int              `json:"id,omitempty"`
	Name                HistoryComponent `json:"name"`
	ChildComponentNames []string         `json:"childList,omitempty"`
}

type DeployedHistoryComponentMetadataDto struct {
	Id               int       `json:"id"`
	DeployedOn       time.Time `json:"deployedOn"`
	DeployedBy       string    `json:"deployedBy"` //emailId of user
	DeploymentStatus string    `json:"deploymentStatus"`
}

type HistoryDetailDto struct {
	//for deployment template
	TemplateName        string `json:"templateName,omitempty"`
	TemplateVersion     string `json:"templateVersion,omitempty"`
	IsAppMetricsEnabled *bool  `json:"isAppMetricsEnabled,omitempty"`
	//for pipeline strategy
	PipelineTriggerType pipelineConfig.TriggerType `json:"pipelineTriggerType,omitempty"`
	Strategy            string                     `json:"strategy,omitempty"`
	//for configmap and secret
	Type               string               `json:"type,omitempty"`
	External           *bool                `json:"external,omitempty"`
	MountPath          string               `json:"mountPath,omitempty"`
	ExternalSecretType string               `json:"externalType,omitempty"`
	RoleARN            string               `json:"roleARN,omitempty"`
	SubPath            *bool                `json:"subPath,omitempty"`
	FilePermission     string               `json:"filePermission,omitempty"`
	CodeEditorValue    *HistoryDetailConfig `json:"codeEditorValue"`
}

type HistoryDetailConfig struct {
	DisplayName      string            `json:"displayName"`
	Value            string            `json:"value"`
	VariableSnapshot map[string]string `json:"variableSnapshot"`
	ResolvedValue    string            `json:"resolvedValue"`
}

//history components(deployment template, configMaps, secrets, pipeline strategy) components below

type ConfigMapAndSecretHistoryDto struct {
	Id         int           `json:"id"`
	PipelineId int           `json:"pipelineId"`
	AppId      int           `json:"appId"`
	DataType   string        `json:"dataType,omitempty"`
	ConfigData []*ConfigData `json:"configData,omitempty"`
	Deployed   bool          `json:"deployed"`
	DeployedOn time.Time     `json:"deployedOn"`
	DeployedBy int32         `json:"deployedBy"`
	EmailId    string        `json:"emailId"`
}

type PrePostCdScriptHistoryDto struct {
	Id                   int                              `json:"id"`
	PipelineId           int                              `json:"pipelineId"`
	Script               string                           `json:"script"`
	Stage                string                           `json:"stage"`
	ConfigMapSecretNames PrePostStageConfigMapSecretNames `json:"configmapSecretNames"`
	ConfigMapData        []*ConfigData                    `json:"configmapData"`
	SecretData           []*ConfigData                    `json:"secretData"`
	TriggerType          string                           `json:"triggerType"`
	ExecInEnv            bool                             `json:"execInEnv"`
	Deployed             bool                             `json:"deployed"`
	DeployedOn           time.Time                        `json:"deployedOn"`
	DeployedBy           int32                            `json:"deployedBy"`
}

type PrePostStageConfigMapSecretNames struct {
	ConfigMaps []string `json:"configMaps"`
	Secrets    []string `json:"secrets"`
}

type DeploymentTemplateHistoryDto struct {
	Id                      int       `json:"id"`
	PipelineId              int       `json:"pipelineId"`
	AppId                   int       `json:"appId"`
	ImageDescriptorTemplate string    `json:"imageDescriptorTemplate,omitempty"`
	Template                string    `json:"template,omitempty"`
	TemplateName            string    `json:"templateName,omitempty"`
	TemplateVersion         string    `json:"templateVersion,omitempty"`
	IsAppMetricsEnabled     bool      `json:"isAppMetricsEnabled"`
	TargetEnvironment       int       `json:"targetEnvironment,omitempty"`
	Deployed                bool      `json:"deployed"`
	DeployedOn              time.Time `json:"deployedOn"`
	DeployedBy              int32     `json:"deployedBy"`
	EmailId                 string    `json:"emailId"`
	DeploymentStatus        string    `json:"deploymentStatus,omitempty"`
	WfrId                   int       `json:"wfrId,omitempty"`
	WorkflowType            string    `json:"workflowType,omitempty"`
}

type PipelineStrategyHistoryDto struct {
	Id         int       `json:"id"`
	PipelineId int       `json:"pipelineId"`
	Strategy   string    `json:"strategy,omitempty"`
	Config     string    `json:"config,omitempty"`
	Default    bool      `json:"default,omitempty"`
	Deployed   bool      `json:"deployed"`
	DeployedOn time.Time `json:"deployedOn"`
	DeployedBy int32     `json:"deployedBy"`
	EmailId    string    `json:"emailId"`
}

// duplicate structs below, because importing from pkg/pipeline was resulting in circular dependency

type ConfigList struct {
	ConfigData []*ConfigData `json:"maps"`
}

type SecretList struct {
	ConfigData []*ConfigData `json:"secrets"`
}

type ConfigData struct {
	Name                  string           `json:"name"`
	Type                  string           `json:"type"`
	External              bool             `json:"external"`
	MountPath             string           `json:"mountPath,omitempty"`
	Data                  json.RawMessage  `json:"data"`
	DefaultData           json.RawMessage  `json:"defaultData,omitempty"`
	DefaultMountPath      string           `json:"defaultMountPath,omitempty"`
	Global                bool             `json:"global"`
	ExternalSecretType    string           `json:"externalType"`
	ExternalSecret        []ExternalSecret `json:"secretData"`
	DefaultExternalSecret []ExternalSecret `json:"defaultSecretData,omitempty"`
	ESOSecretData         ESOSecretData    `json:"esoSecretData"`
	DefaultESOSecretData  ESOSecretData    `json:"defaultESOSecretData,omitempty"`
	RoleARN               string           `json:"roleARN"`
	SubPath               bool             `json:"subPath"`
	FilePermission        string           `json:"filePermission"`
}

type ExternalSecret struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Property string `json:"property,omitempty"`
	IsBinary bool   `json:"isBinary"`
}

type ESOSecretData struct {
	SecretStore     json.RawMessage `json:"secretStore,omitempty"`
	SecretStoreRef  json.RawMessage `json:"secretStoreRef,omitempty"`
	EsoData         []ESOData       `json:"esoData"`
	RefreshInterval string          `json:"refreshInterval,omitempty"`
}

type ESOData struct {
	SecretKey string `json:"secretKey"`
	Key       string `json:"key"`
	Property  string `json:"property,omitempty"`
}

// //TODO Aditya move to *history.ConfigData
//
//	func GetDecodedData(secretDataMap map[string]*ConfigData) (map[string]*ConfigData, error) {
//		//var marshal []byte
//		for name, configData := range secretDataMap {
//			//dataMap := make(map[string]string)
//			//
//			//err := json.Unmarshal(configData.Data, &dataMap)
//			//if err != nil {
//			//	return nil, err
//			//}
//			//for key, value := range dataMap {
//			//	decodedData, err := base64.StdEncoding.DecodeString(value)
//			//	//todo Aditya return err
//			//	if err != nil {
//			//		fmt.Println("Error decoding base64:", err)
//			//	}
//			//	dataMap[key] = string(decodedData)
//			//}
//			//marshal, err = json.Marshal(dataMap)
//			//if err != nil {
//			//	return nil, err
//			//}
//			marshal, err := util.GetDecodedAndEncodedData(configData.Data, util.DecodeSecret)
//			if err != nil {
//				return nil, err
//			}
//			configData.Data = marshal
//			secretDataMap[name] = configData
//
//		}
//		return secretDataMap, nil
//	}

//func (SecretList) GetTransformedDataForSecret(data string, mode util.SecretTransformMode) (string, error) {
//	secretsList := SecretList{}
//	err := json.Unmarshal([]byte(data), &secretsList)
//	if err != nil {
//		return "", err
//	}
//
//	for _, configData := range secretsList.ConfigData {
//		configData.Data, err = util.GetDecodedAndEncodedData(configData.Data, mode)
//		if err != nil {
//			return "", err
//		}
//	}
//
//	marshal, err := json.Marshal(secretsList)
//	if err != nil {
//		return "", err
//	}
//	return string(marshal), nil
//}

func (ConfigData) GetTransformedDataForSecret(data string, mode util.SecretTransformMode) (string, error) {
	secretDataMap := make(map[string]*ConfigData)
	err := json.Unmarshal([]byte(data), &secretDataMap)
	if err != nil {
		return "", err
	}

	for _, configData := range secretDataMap {
		data, err := util.GetDecodedAndEncodedData(configData.Data, mode)
		if err != nil {
			return "", err
		}
		configData.Data = data

	}

	resolvedTemplate, err := json.Marshal(secretDataMap)
	if err != nil {
		return "", err
	}
	return string(resolvedTemplate), nil
}

func (SecretList) GetTransformedDataForSecret(data string, mode util.SecretTransformMode) (string, error) {
	secretsList := SecretList{}
	err := json.Unmarshal([]byte(data), &secretsList)
	if err != nil {
		return "", err
	}

	for _, configData := range secretsList.ConfigData {
		configData.Data, err = util.GetDecodedAndEncodedData(configData.Data, mode)
		if err != nil {
			return "", err
		}
	}

	marshal, err := json.Marshal(secretsList)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}
