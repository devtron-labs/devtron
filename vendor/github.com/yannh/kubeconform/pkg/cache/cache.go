package cache

type Cache interface {
	Get(resourceKind, resourceAPIVersion, k8sVersion string) (interface{}, error)
	Set(resourceKind, resourceAPIVersion, k8sVersion string, schema interface{}) error
}
