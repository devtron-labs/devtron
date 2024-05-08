package telemetry

type TelemetryConfig struct {
	// cloudProvider will be set only once at the startup and the value will be stored as cache
	// which will further reduce the api calls made to various IMDS
	cloudProvider string
}
