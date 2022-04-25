package service

import (
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	repository2 "github.com/devtron-labs/devtron/client/argocdServer/repository"
	"github.com/devtron-labs/devtron/client/pubsub"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDeploymentFullMode "github.com/devtron-labs/devtron/pkg/appStore/deployment/fullMode"
	repository4 "github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/values/service"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	"go.uber.org/zap"
	"testing"
)

func TestInstalledAppServiceImpl_DeployDefaultChartOnCluster(t *testing.T) {
	type fields struct {
		logger                               *zap.SugaredLogger
		installedAppRepository               repository4.InstalledAppRepository
		chartTemplateService                 util.ChartTemplateService
		refChartDir                          appStoreBean.RefChartProxyDir
		repositoryService                    repository2.ServiceClient
		appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository
		environmentRepository                repository.EnvironmentRepository
		teamRepository                       team.TeamRepository
		appRepository                        app.AppRepository
		acdClient                            application.ServiceClient
		appStoreValuesService             service.AppStoreValuesService
		pubsubClient                      *pubsub.PubSubClient
		tokenCache                        *util2.TokenCache
		chartGroupDeploymentRepository    repository4.ChartGroupDeploymentRepository
		envService                        cluster.EnvironmentService
		ArgoK8sClient                     argocdServer.ArgoK8sClient
		gitFactory                        *util.GitFactory
		aCDAuthConfig                     *util2.ACDAuthConfig
		gitOpsRepository                  repository3.GitOpsConfigRepository
		userService                       user.UserService
		appStoreDeploymentService         AppStoreDeploymentService
		appStoreDeploymentFullModeService appStoreDeploymentFullMode.AppStoreDeploymentFullModeService
	}
	type args struct {
		bean   *cluster.ClusterBean
		userId int32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &InstalledAppServiceImpl{
				logger:                               tt.fields.logger,
				installedAppRepository:               tt.fields.installedAppRepository,
				chartTemplateService:                 tt.fields.chartTemplateService,
				refChartDir:                          tt.fields.refChartDir,
				repositoryService:                    tt.fields.repositoryService,
				appStoreApplicationVersionRepository: tt.fields.appStoreApplicationVersionRepository,
				environmentRepository:                tt.fields.environmentRepository,
				teamRepository:                       tt.fields.teamRepository,
				appRepository:                        tt.fields.appRepository,
				acdClient:                            tt.fields.acdClient,
				appStoreValuesService:                tt.fields.appStoreValuesService,
				pubsubClient:                         tt.fields.pubsubClient,
				tokenCache:                           tt.fields.tokenCache,
				chartGroupDeploymentRepository:       tt.fields.chartGroupDeploymentRepository,
				envService:                           tt.fields.envService,
				ArgoK8sClient:                        tt.fields.ArgoK8sClient,
				gitFactory:                           tt.fields.gitFactory,
				aCDAuthConfig:                        tt.fields.aCDAuthConfig,
				gitOpsRepository:                     tt.fields.gitOpsRepository,
				userService:                          tt.fields.userService,
				appStoreDeploymentService:            tt.fields.appStoreDeploymentService,
				appStoreDeploymentFullModeService:    tt.fields.appStoreDeploymentFullModeService,
			}
			got, err := impl.DeployDefaultChartOnCluster(tt.args.bean, tt.args.userId)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeployDefaultChartOnCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DeployDefaultChartOnCluster() got = %v, want %v", got, tt.want)
			}
		})
	}
}
