package util

import (
	"github.com/devtron-labs/devtron/pkg/k8s/application/bean"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func TestParseK8sProxyURL(t *testing.T) {
	//Test case input
	urls := []string{"/api/v1/namespaces/mynamespace/pods", "/api", "/api/v1/namespaces/default/pods", "/api/v1/namespaces/default/pods/my-pod", "/api/v1/namespaces/devtroncd/services/postgresql-postgresql",
		"/apis/apps/v1/namespaces/default/deployments/frontend",
		"/api/v1/nodes/my-node",
		"/api/v1/namespaces/default/secrets/my-secret",
		"/apis/batch/v1/namespaces/my-namespace/job",
	}
	// Test case expected output
	namespaces := []string{"mynamespace", bean.ALL, "default", "default", "devtroncd", "default", bean.ALL, "default", "my-namespace"}
	resourceNames := []string{bean.ALL, bean.ALL, bean.ALL, "my-pod", "postgresql-postgresql", "frontend", bean.ALL, "my-secret", bean.ALL}
	GVKs := []schema.GroupVersionKind{
		{Group: bean.K8sEmpty, Version: "v1", Kind: "pod"},
		{Group: bean.K8sEmpty, Version: bean.ALL, Kind: bean.ALL},
		{Group: bean.K8sEmpty, Version: "v1", Kind: "pod"},
		{Group: bean.K8sEmpty, Version: "v1", Kind: "pod"},
		{Group: bean.K8sEmpty, Version: "v1", Kind: "service"},
		{Group: "apps", Version: "v1", Kind: "deployment"},
		{Group: bean.ALL, Version: bean.ALL, Kind: bean.ALL},
		{Group: bean.K8sEmpty, Version: bean.V1, Kind: "secret"},
		{Group: "batch", Version: bean.V1, Kind: "job"},
	}
	//result
	for idx, url := range urls {
		t.Run(url, func(t *testing.T) {
			namespace, gvk, resourceName := ParseK8sProxyURL(urls[idx])
			if namespace != namespaces[idx] {
				t.Errorf("namespace mismatch. Expected %s, got %s", namespaces[idx], namespace)
			}
			if gvk != GVKs[idx] {
				t.Errorf("gvk mismatch. Expected %s, got %s", GVKs[idx], gvk)
			}
			if resourceName != resourceNames[idx] {
				t.Errorf("resourceName mismatch. Expected %s, got %s", resourceNames[idx], resourceName)
			}
		})
	}
}
