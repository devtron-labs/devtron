package bean

type AppDetail struct {
	Metadata                 *AppMetadata                    `json:"metadata,notnull" validate:"required"`
	GitMaterials             []*GitMaterial                  `json:"gitMaterials,notnull" validate:"required"`
	GlobalDeploymentTemplate *DeploymentTemplate             `json:"globalDeploymentTemplate,notnull" validate:"required"`
	GlobalConfigMaps         []*ConfigMap                    `json:"globalConfigMaps"`
	GlobalSecrets            []*Secret                       `json:"globalSecrets"`
	EnvironmentOverrides     map[string]*EnvironmentOverride `json:"environmentOverride"`
}

type AppMetadata struct {
	AppName     string      `json:"appName" validate:"required"`
	ProjectName string      `json:"projectName" validate:"required"`
	Labels      []*AppLabel `json:"labels"`
}

type AppLabel struct {
	Key   string `json:"key,notnull" validate:"required"`
	Value string `json:"value,notnull" validate:"required"`
}

type GitMaterial struct {
	GitAccountName  string `json:"gitAccountName,notnull" validate:"required"`
	GitUrl          string `json:"gitUrl,notnull" validate:"required"`
	CheckoutPath    string `json:"checkoutPath,notnull" validate:"required"`
	FetchSubmodules bool   `json:"fetchSubmodules"`
}

type DeploymentTemplate struct {
	ChartRefId     int                    `json:"chartRefId,notnull" validate:"required"`
	Template       map[string]interface{} `json:"template,notnull" validate:"required"`
	ShowAppMetrics bool                   `json:"showAppMetrics"`
}

type ConfigMap struct {
	Name                  string                                `json:"name,notnull" validate:"required"`
	IsExternal            bool                                  `json:"isExternal"`
	UsageType             string                                `json:"usageType,omitempty" validate:"oneof=environment volume"`
	Data                  map[string]interface{}                `json:"data"`
	DataVolumeUsageConfig *ConfigMapSecretDataVolumeUsageConfig `json:"dataVolumeUsageConfig"`
}

type Secret struct {
	Name                  string                                `json:"name,notnull" validate:"required"`
	IsExternal            bool                                  `json:"isExternal"`
	ExternalType          string                                `json:"externalType,omitempty"`
	UsageType             string                                `json:"usageType,omitempty" validate:"oneof=environment volume"`
	Data                  map[string]interface{}                `json:"data"`
	DataVolumeUsageConfig *ConfigMapSecretDataVolumeUsageConfig `json:"dataVolumeUsageConfig"`
	RoleArn               string                                `json:"roleArn"`
	ExternalSecretData    []*ExternalSecret                     `json:"externalSecretData"`
}

type ConfigMapSecretDataVolumeUsageConfig struct {
	MountPath      string `json:"mountPath"`
	SubPath        bool   `json:"subPath"`
	FilePermission string `json:"filePermission"`
}

type ExternalSecret struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Property string `json:"property,omitempty"`
	IsBinary bool   `json:"isBinary"`
}

type EnvironmentOverride struct {
	DeploymentTemplate *DeploymentTemplate `json:"deploymentTemplate"`
	ConfigMaps         []*ConfigMap        `json:"configMaps"`
	Secrets            []*Secret           `json:"secrets"`
}

type AppWorkflows struct{

}