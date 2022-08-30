package devtron_integration_manager

type IntegrationManagerConfig struct {
	ModuleStatusCheckIntervalInMins int `env:"MODULE_STATUS_CHECK_INTERVAL_IN_MINs" envDefault:"1"`
}

type IntegrationModuleStatus = string

const (
	ModuleStatusNotInstalled  IntegrationModuleStatus = "notInstalled"
	ModuleStatusInstalled     IntegrationModuleStatus = "installed"
	ModuleStatusAvailable     IntegrationModuleStatus = "available"
	ModuleStatusInstalling    IntegrationModuleStatus = "installing"
	ModuleStatusInstallFailed IntegrationModuleStatus = "installFailed"
	ModuleStatusTimeout       IntegrationModuleStatus = "timeout"
)
