package argocdServer

import (
	"os"
	"testing"

	"github.com/devtron-labs/devtron/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestApplicationTargetConfiguration(t *testing.T) {
	for _, name := range []string{"ACD_NAMESPACE", "ARGO_APPLICATION_PROJECT"} {
		t.Setenv(name, "")
		require.NoError(t, os.Unsetenv(name))
	}
	defaults, err := GetACDDeploymentConfig()
	require.NoError(t, err)
	require.Equal(t, "devtroncd", defaults.ApplicationNamespace)
	require.Equal(t, "default", defaults.ApplicationProject)
	t.Setenv("ACD_NAMESPACE", "argo-control")
	t.Setenv("CD_NAMESPACE", "connection-only")
	t.Setenv("ARGO_APPLICATION_PROJECT", "applications")
	configured, err := GetACDDeploymentConfig()
	require.NoError(t, err)
	require.Equal(t, "argo-control", configured.ApplicationNamespace)
	require.Equal(t, "applications", configured.ApplicationProject)
	legacy, err := util.GetACDAuthConfig()
	require.NoError(t, err)
	require.Equal(t, legacy.ACDConfigMapNamespace, configured.ApplicationNamespace, "creation and existing repository/query consumers must use the same namespace")
	for _, test := range []struct{ namespace, project string }{
		{"invalid/namespace", "applications"}, {"argo-control", "invalid/project"},
	} {
		t.Setenv("ACD_NAMESPACE", test.namespace)
		t.Setenv("ARGO_APPLICATION_PROJECT", test.project)
		_, err := GetACDDeploymentConfig()
		require.Error(t, err)
	}
}
