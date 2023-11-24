package cluster

import (
	util2 "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/k8s/informer"
	"go.uber.org/zap"
	"testing"
)

func TestClusterServiceImpl_CheckIfConfigIsValid(t *testing.T) {
	t.SkipNow()
	type fields struct {
		clusterRepository  repository.ClusterRepository
		logger             *zap.SugaredLogger
		K8sUtil            *util2.K8sUtil
		K8sInformerFactory informer.K8sInformerFactory
	}
	type args struct {
		cluster *ClusterBean
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			//incorrect server config
			args: args{
				cluster: &ClusterBean{
					Id:        4,
					ServerUrl: "",
					Config: map[string]string{
						"bearer_token": "",
					},
				},
			},
			wantErr: true,
		},
		{
			//correct server config
			args: args{
				cluster: &ClusterBean{
					Id:        5,
					ServerUrl: "",
					Config: map[string]string{
						"bearer_token": "",
					},
				},
			},
			wantErr: false,
		},
	}

	logger, _ := util.NewSugardLogger()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := ClusterServiceImpl{
				clusterRepository:  nil,
				logger:             logger,
				K8sUtil:            nil,
				K8sInformerFactory: nil,
			}
			if err := impl.CheckIfConfigIsValid(tt.args.cluster); (err != nil) != tt.wantErr {
				t.Errorf("ClusterServiceImpl.CheckIfConfigIsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
