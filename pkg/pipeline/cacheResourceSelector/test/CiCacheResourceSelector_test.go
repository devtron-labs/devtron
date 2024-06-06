package test

import (
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/devtron-labs/authenticator/client"
	k8s3 "github.com/devtron-labs/common-lib-private/utils/k8s"
	"github.com/devtron-labs/devtron/enterprise/pkg/expressionEvaluators"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster"
	repository1 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	k8s2 "github.com/devtron-labs/devtron/pkg/k8s"
	"github.com/devtron-labs/devtron/pkg/k8s/application"
	"github.com/devtron-labs/devtron/pkg/k8s/informer"
	"github.com/devtron-labs/devtron/pkg/pipeline/cacheResourceSelector"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCacheSelector(t *testing.T) {
	t.Run("cache selector", func(tt *testing.T) {
		logger, err := util.NewSugardLogger()
		assert.Nil(tt, err)
		celServiceImpl := expressionEvaluators.NewCELServiceImpl(logger)
		runtimeConfig, err := client.GetRuntimeConfig()
		assert.Nil(tt, err)
		k8sUtil := k8s3.NewK8sUtilExtended(logger, runtimeConfig, nil)
		sqlConfig, err := sql.GetConfig()
		assert.Nil(tt, err)
		dbConnection, err := sql.NewDbConnection(sqlConfig, logger)
		assert.Nil(tt, err)
		informerFactoryImpl := informer.NewK8sInformerFactoryImpl(logger, map[string]map[string]bool{}, runtimeConfig, k8sUtil)
		clusterRepository := repository1.NewClusterRepositoryImpl(dbConnection, logger)
		clusterServiceImpl := cluster.NewClusterServiceImpl(clusterRepository, logger, k8sUtil, informerFactoryImpl, nil, nil, nil, nil, nil, nil)
		k8sCommonServiceImpl := k8s2.NewK8sCommonServiceImpl(logger, k8sUtil, nil, nil, clusterServiceImpl, nil, nil)
		k8sAppService, err := application.NewK8sApplicationServiceImpl(logger, clusterServiceImpl, nil, nil, k8sUtil, nil, nil, k8sCommonServiceImpl, nil, nil, nil, nil, nil, nil, nil, nil)
		assert.Nil(tt, err)
		cacheResourceSelectorImpl := cacheResourceSelector.NewCiCacheResourceSelectorImpl(logger, celServiceImpl, k8sAppService)
		scope := resourceQualifiers.Scope{}
		appLabels := make(map[string]string)
		appLabels["devtron.ai/language"] = "python"
		time.Sleep(3 * time.Second)
		ciCacheResource, err := cacheResourceSelectorImpl.GetAvailResource(scope, appLabels, 1)
		assert.Nil(tt, err)
		fmt.Println("ci cache resource", ciCacheResource)
		cacheResourceSelectorImpl.UpdateResourceStatus(1, string(v1alpha1.NodeRunning))
		ciCacheResource, err = cacheResourceSelectorImpl.GetAvailResource(scope, appLabels, 1)
		assert.Nil(tt, err)
		fmt.Println("ci cache resource", ciCacheResource)
	})
}
