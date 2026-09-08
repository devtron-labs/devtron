package appStoreDeploymentCommon

import (
	"testing"
)

func TestAppStoreDeploymentCommonServiceHelper(t *testing.T) {
	t.Run("empty sources from manifest test", func(t *testing.T) {
		impl := &AppStoreDeploymentCommonServiceImpl{}
		sources, err := impl.getSourcesFromManifest("")
		if err != nil {
			t.Errorf("unexpected error on empty manifest: %v", err)
		}
		if len(sources) != 0 {
			t.Errorf("expected 0 sources, got %d", len(sources))
		}
	})
}
