package chartRepo

import (
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/stretchr/testify/mock"
	"testing"
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
