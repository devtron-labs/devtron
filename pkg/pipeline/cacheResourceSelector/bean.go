package cacheResourceSelector

type ResourceStatus string

const (
	AvailableResourceStatus   ResourceStatus = "Available"
	UnAvailableResourceStatus ResourceStatus = "Unavailable"
)

type Config struct {
	CachePVCs           []string `env:"CACHE_PVCs" envDefault:"java-cache-pvc-2,node-cache-pvc-1,node-cache-pvc-2" envSeparator:","`
	PVCNameExpression   string   `env:"PVC_NAME_EXPRESSION" envDefault:"appLabels['devtron.ai/language'] == 'java' ? 'java-cache' : 'node-cache'"`
	MountPathExpression string   `env:"PVC_MOUNT_PATH_EXPRESSION" envDefault:"appLabels['devtron.ai/language'] == 'java' ? '/devtroncd/.m2' : '/devtroncd/node_modules'"`
}

const BuildPVCLabelKey1 = "devtron.ai/cache"
const BuildPVCLabelValue1 = "pvc"
const BuildPVCLabelKey2 = "devtron.ai/pvc-name"
const BuildWorkflowId = "devtron.ai/ciWorkflowId"

type CiCacheResource struct {
	PVCName   string
	MountPath string
}

func (resource *CiCacheResource) GetMap() map[string]string {
	dataMap := make(map[string]string)
	dataMap["PVCName"] = resource.PVCName
	dataMap["MountPath"] = resource.MountPath
	return dataMap
}

func (resource *CiCacheResource) GetResourceName() string {
	return resource.PVCName
}
