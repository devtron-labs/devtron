package devtron_integration_manager

type IntegrationManagerConfig struct {
	ModuleStatusCheckIntervalInMins int `env:"MODULE_STATUS_CHECK_INTERVAL_IN_MINs" envDefault:"1"`
}
