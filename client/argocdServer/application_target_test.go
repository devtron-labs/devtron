package argocdServer

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplicationTargetConfiguration(t *testing.T) {
	for _, name := range []string{"CD_NAMESPACE", "ARGO_APPLICATION_PROJECT"} {
		t.Setenv(name, "")
		require.NoError(t, os.Unsetenv(name))
	}
	defaults, err := GetACDDeploymentConfig()
	require.NoError(t, err)
	require.Equal(t, "devtroncd", defaults.ApplicationNamespace)
	require.Equal(t, "default", defaults.ApplicationProject)
	t.Setenv("ACD_NAMESPACE", "installation")
	t.Setenv("CD_NAMESPACE", "argo-control")
	t.Setenv("ARGO_APPLICATION_PROJECT", "applications")
	configured, err := GetACDDeploymentConfig()
	require.NoError(t, err)
	require.Equal(t, "argo-control", configured.ApplicationNamespace)
	require.Equal(t, "applications", configured.ApplicationProject)
	for _, test := range []struct{ namespace, project string }{
		{"invalid/namespace", "applications"}, {"argo-control", "invalid/project"},
	} {
		t.Setenv("CD_NAMESPACE", test.namespace)
		t.Setenv("ARGO_APPLICATION_PROJECT", test.project)
		_, err := GetACDDeploymentConfig()
		require.Error(t, err)
	}
}
