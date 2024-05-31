/*
 * Copyright (c) 2024. Devtron Inc.
 */

package application

import (
	"context"
	"github.com/devtron-labs/authenticator/client"
	k8s3 "github.com/devtron-labs/common-lib-private/utils/k8s"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	repository4 "github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/cluster"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/k8s"
	"github.com/devtron-labs/devtron/pkg/k8s/informer"
	"github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs"
	repository10 "github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPortForwardManager(t *testing.T) {
	t.Run("base", func(t *testing.T) {
		sugaredLogger, err := util.NewSugardLogger()
		assert.Nil(t, err)
		config, err := sql.GetConfig()
		assert.Nil(t, err)
		db, err := sql.NewDbConnection(config, sugaredLogger)
		assert.Nil(t, err)
		runtimeConfig, err := client.GetRuntimeConfig()
		assert.Nil(t, err)
		v := informer.NewGlobalMapClusterNamespace()
		runTimeConfig, _ := client.GetRuntimeConfig()
		k8sUtil := k8s3.NewK8sUtilExtended(sugaredLogger, runTimeConfig, nil)
		k8sInformerFactoryImpl := informer.NewK8sInformerFactoryImpl(sugaredLogger, v, runtimeConfig, k8sUtil)
		clusterRepositoryImpl := repository2.NewClusterRepositoryImpl(db, sugaredLogger)
		defaultAuthPolicyRepositoryImpl := repository4.NewDefaultAuthPolicyRepositoryImpl(db, sugaredLogger)
		defaultAuthRoleRepositoryImpl := repository4.NewDefaultAuthRoleRepositoryImpl(db, sugaredLogger)
		userAuthRepositoryImpl := repository4.NewUserAuthRepositoryImpl(db, sugaredLogger, defaultAuthPolicyRepositoryImpl, defaultAuthRoleRepositoryImpl)
		userRepositoryImpl := repository4.NewUserRepositoryImpl(db, sugaredLogger)
		roleGroupRepositoryImpl := repository4.NewRoleGroupRepositoryImpl(db, sugaredLogger)
		userServiceImpl := user.NewUserServiceImpl(userAuthRepositoryImpl, sugaredLogger, userRepositoryImpl, roleGroupRepositoryImpl, nil, nil, nil, nil, nil, nil, nil, nil)
		clusterServiceImpl := cluster.NewClusterServiceImpl(clusterRepositoryImpl, sugaredLogger, k8sUtil, k8sInformerFactoryImpl, userAuthRepositoryImpl, userRepositoryImpl, roleGroupRepositoryImpl, nil, userServiceImpl)
		//k8sClientServiceImpl := application2.NewK8sClientServiceImpl(sugaredLogger, clusterServiceImpl, nil)
		//clusterServiceImpl := cluster2.NewClusterServiceImplExtended(clusterRepositoryImpl, nil, nil, sugaredLogger, nil, nil, nil, nil, nil)
		k8sResourceHistoryRepositoryImpl := repository10.NewK8sResourceHistoryRepositoryImpl(db, sugaredLogger)
		appRepositoryImpl := app.NewAppRepositoryImpl(db, sugaredLogger)
		environmentRepositoryImpl := repository2.NewEnvironmentRepositoryImpl(db, sugaredLogger, nil)
		k8sResourceHistoryServiceImpl := kubernetesResourceAuditLogs.Newk8sResourceHistoryServiceImpl(k8sResourceHistoryRepositoryImpl, sugaredLogger, appRepositoryImpl, environmentRepositoryImpl)
		//k8sApplicationService := application.NewK8sApplicationServiceImpl(sugaredLogger, clusterServiceImpl, nil, nil, nil, nil, k8sResourceHistoryServiceImpl, nil)
		K8sCommonService := k8s.NewK8sCommonServiceImpl(sugaredLogger, k8sUtil, nil, k8sResourceHistoryServiceImpl, clusterServiceImpl, nil, nil)
		portForwardManagerImpl, err := NewPortForwardManagerImpl(sugaredLogger, K8sCommonService)
		assert.Nil(t, err)
		//portForwardRequest := bean.PortForwardRequest{
		//	Namespace:   "monitoring",
		//	ServiceName: "scoop-stage-mon-service",
		//	ClusterId:   1,
		//	TargetPort:  "80",
		//}
		//_, err = portForwardManagerImpl.ForwardPort(context.Background(), portForwardRequest)
		proxyPort, err := portForwardManagerImpl.StartK8sProxy(context.Background(), 1)
		assert.Nil(t, err)
		assert.NotZero(t, proxyPort)
		time.Sleep(50 * time.Minute)
	})
}
