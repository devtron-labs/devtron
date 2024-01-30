package service

import (
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	"testing"

	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	repository2 "github.com/devtron-labs/devtron/client/argocdServer/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/util"
	repository5 "github.com/devtron-labs/devtron/pkg/appStore/chartGroup/repository"
	appStoreDeploymentFullMode "github.com/devtron-labs/devtron/pkg/appStore/deployment/fullMode"
	repository4 "github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/values/service"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/team"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	"go.uber.org/zap"
)

func TestInstalledAppServiceImpl_DeployDefaultChartOnCluster(t *testing.T) {
	type fields struct {
		logger                               *zap.SugaredLogger
		installedAppRepository               repository4.InstalledAppRepository
		chartTemplateService                 util.ChartTemplateService
		repositoryService                    repository2.ServiceClient
		appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository
		environmentRepository                repository.EnvironmentRepository
		teamRepository                       team.TeamRepository
		appRepository                        app.AppRepository
		acdClient                            application.ServiceClient
		appStoreValuesService                service.AppStoreValuesService
		pubsubClient                         *pubsub.PubSubClientServiceImpl
		tokenCache                           *util2.TokenCache
		chartGroupDeploymentRepository       repository5.ChartGroupDeploymentRepository
		envService                           cluster.EnvironmentService
		ArgoK8sClient                        argocdServer.ArgoK8sClient
		gitFactory                           *git.GitFactory
		aCDAuthConfig                        *util2.ACDAuthConfig
		userService                          user.UserService
		appStoreDeploymentService            AppStoreDeploymentService
		appStoreDeploymentFullModeService    appStoreDeploymentFullMode.AppStoreDeploymentFullModeService
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
				appStoreApplicationVersionRepository: tt.fields.appStoreApplicationVersionRepository,
				environmentRepository:                tt.fields.environmentRepository,
				teamRepository:                       tt.fields.teamRepository,
				appRepository:                        tt.fields.appRepository,
				acdClient:                            tt.fields.acdClient,
				appStoreValuesService:                tt.fields.appStoreValuesService,
				pubsubClient:                         tt.fields.pubsubClient,
				chartGroupDeploymentRepository:       tt.fields.chartGroupDeploymentRepository,
				envService:                           tt.fields.envService,
				aCDAuthConfig:                        tt.fields.aCDAuthConfig,
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
