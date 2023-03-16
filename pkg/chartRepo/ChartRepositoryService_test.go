package chartRepo

import (
	"testing"

	"github.com/devtron-labs/devtron/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type ChartRepositoryServiceMock struct {
	mock.Mock
}

func TestChartRepositoryServiceImpl_ValidateChartDetails(t *testing.T) {
	sugaredLogger, _ := util.NewSugardLogger()
	impl := &ChartRepositoryServiceImpl{
		logger:         sugaredLogger,
		repoRepository: new(ChartRepoRepositoryImplMock),
		K8sUtil:        nil,
		clusterService: new(ClusterServiceImplMock),
		aCDAuthConfig:  nil,
		client:         nil,
	}

	type args struct {
		FileName     string
		chartVersion string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test file format",
			args: struct {
				FileName     string
				chartVersion string
			}{FileName: "test.tar.gz", chartVersion: "1.0.0"},
			want: "test_1-0-0",
		},
		{
			name: "Test file format",
			args: struct {
				FileName     string
				chartVersion string
			}{FileName: "test.pdf", chartVersion: "1.0.0"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := impl.ValidateChartDetails(tt.args.FileName, tt.args.chartVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateChartDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateChartDetails() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChartRepositoryServiceRemoveRepoData(t *testing.T) {
	sugaredLogger, err := util.NewSugardLogger()
	assert.Nil(t, err)
	
    impl := &ChartRepositoryServiceImpl{
        logger:  sugaredLogger,
    }
    t.Run("Invalid JSON byte array", func(t *testing.T) {
        data := map[string]string{
            "helm.repositories": "invalid json",
        }
        _, err := impl.removeRepoData(data, "test-repo")
        if err == nil {
            t.Errorf("Expected an error but got nil")
        }
    })

    t.Run("Error in unmarshal helmRepositories", func(t *testing.T) {
        data := map[string]string{
            "helm.repositories": "- name: test-1\n  type: helm\n  url: https://localhost\n- name: test-repo\n  type: helm\n  url: https://localhost\n- name: test-2\n  type: helm\n  url: https://localhost\n",
        }
        _, err := impl.removeRepoData(data, "test-repo")
        if err == nil {
            t.Errorf("Expected an error but got nil")
        }
    })

    t.Run("Error in unmarshal repositories", func(t *testing.T) {
        data := map[string]string{
            "helm.repositories": "null\n",
            "repositories": "invalid json",
        }
        _, err := impl.removeRepoData(data, "test-repo")
        if err == nil {
            t.Errorf("Expected an error but got nil")
        }
    })

    t.Run("Repository is found and removed from repositories", func(t *testing.T) {
        data := map[string]string{
            "helm.repositories": "null\n",
			"repositories": "- name: test-1\n  type: helm\n  url: https://localhost\n- name: test-repo\n  type: helm\n  url: https://localhost\n- name: test-2\n  type: helm\n  url: https://localhost\n",
            "dex.config": "",
        }
        expected := map[string]string{
            "helm.repositories": "null\n",
			"repositories": "- name: test-1\n  type: helm\n  url: https://localhost\n- name: test-2\n  type: helm\n  url: https://localhost\n",
            "dex.config": "",
        }
        result, err := impl.removeRepoData(data, "test-repo")
        if err != nil {
            t.Errorf("Unexpected error: %v", err)
        }
        if !assert.Equal(t, result, expected) {
            t.Errorf("Expected %v, but got %v", expected, result)
        }
    })

    t.Run("Repository is not found", func(t *testing.T) {
        data := map[string]string{
            "helm.repositories": "null\n",
			"repositories": "- name: test-1\n  type: helm\n  url: https://localhost\n- name: test-2\n  type: helm\n  url: https://localhost\n",
        }
        _, err := impl.removeRepoData(data, "test-repo")
        if err == nil {
            t.Errorf("Expected an error but got nil")
        }
    })
}
