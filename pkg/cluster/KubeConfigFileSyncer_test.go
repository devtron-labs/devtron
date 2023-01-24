package cluster

import (
	"github.com/devtron-labs/authenticator/client"
	"github.com/devtron-labs/devtron/client/k8s/informer"
	"github.com/devtron-labs/devtron/internal/util"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/user"
	repository3 "github.com/devtron-labs/devtron/pkg/user/repository"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKubeConfigFileSyncer(t *testing.T) {
	t.SkipNow()
	t.Run("sync now", func(t *testing.T) {
		sugaredLogger, _ := util.InitLogger()
		runtimeConfig, err := client.GetRuntimeConfig()
		assert.Nil(t, err)
		v := informer.NewGlobalMapClusterNamespace()
		k8sInformerFactoryImpl := informer.NewK8sInformerFactoryImpl(sugaredLogger, v, runtimeConfig)
		clusterRepositoryImpl := repository2.NewClusterRepositoryFileBased(sugaredLogger)
		defaultAuthPolicyRepositoryImpl := repository3.NewDefaultAuthPolicyRepositoryImpl(nil, sugaredLogger)
		defaultAuthRoleRepositoryImpl := repository3.NewDefaultAuthRoleRepositoryImpl(nil, sugaredLogger)
		userAuthRepositoryImpl := repository3.NewUserAuthRepositoryImpl(nil, sugaredLogger, defaultAuthPolicyRepositoryImpl, defaultAuthRoleRepositoryImpl)
		userRepositoryImpl := repository3.NewUserRepositoryImpl(nil, sugaredLogger)
		roleGroupRepositoryImpl := repository3.NewRoleGroupRepositoryImpl(nil, sugaredLogger)
		k8sUtil := util.NewK8sUtil(sugaredLogger, runtimeConfig)
		userServiceImpl := user.NewUserServiceImpl(userAuthRepositoryImpl, sugaredLogger, userRepositoryImpl, roleGroupRepositoryImpl, nil, nil, nil)
		clusterServiceImpl := NewClusterServiceImpl(clusterRepositoryImpl, sugaredLogger, k8sUtil, k8sInformerFactoryImpl, userServiceImpl)
		fileSyncerImpl := NewKubeConfigFileSyncerImpl(sugaredLogger, clusterServiceImpl)
		fileSyncerImpl.SyncNow()
	})
}
