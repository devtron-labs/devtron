package cacheResourceSelector

type ResourceStatus string

const (
	AvailableResourceStatus   ResourceStatus = "Available"
	UnAvailableResourceStatus ResourceStatus = "Unavailable"
)

type Config struct {
	CachePVCs           []string `env:"CACHE_PVCs"`
	PVCNameExpression   string   `env:"PVC_NAME_EXPRESSION"`
	MountPathExpression string   `env:"PVC_MOUNT_PATH_EXPRESSION"`
}

const BuildPVCLabel = "devtron.ai/build/pvc"
