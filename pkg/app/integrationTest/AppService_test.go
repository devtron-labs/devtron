package integrationTest

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	client "github.com/devtron-labs/devtron/api/helm-app"
	client1 "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	app2 "github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/app/status"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/cluster"
	repository1 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/module"
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
	serverEnvConfig "github.com/devtron-labs/devtron/pkg/server/config"
	"github.com/devtron-labs/devtron/pkg/sql"
	repository2 "github.com/devtron-labs/devtron/pkg/user/repository"
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestAppServiceImpl_UpdateDeploymentStatusAndCheckIsSucceeded(t *testing.T) {

	appService := InitAppService()
	type args struct {
		app        *v1alpha1.Application
		statusTime time.Time
	}
	type testCase struct {
		name        string
		args        args
		isSucceeded bool
		wantErr     bool
	}
	testCasesFile, err := os.Open("testCases.csv")
	if err != nil {
		fmt.Println("err", err)
	}
	defer testCasesFile.Close()
	testCaseReader := csv.NewReader(testCasesFile)
	data, err := testCaseReader.ReadAll()
	if err != nil {
		fmt.Println("error in reading testCases: ", err)
	}
	for _, line := range data {
		testCaseName := line[0]
		application := &v1alpha1.Application{}
		err = json.Unmarshal([]byte(line[1]), application)
		if err != nil {
			fmt.Println("error in unmarshal: ", err)
		}
		statusTimeArg := line[2]
		statusTime, err := time.Parse(time.RFC3339, statusTimeArg)
		if err != nil {
			fmt.Println("error in parsing time: ", err)
		}
		isSucceededArg := line[3]
		isSucceeded, err := strconv.ParseBool(isSucceededArg)
		if err != nil {
			fmt.Println("error in parsing bool isSucceededArg: ", err)
		}
		wantErrArg := line[4]
		wantErr, err := strconv.ParseBool(wantErrArg)
		if err != nil {
			fmt.Println("error in parsing bool wantErrArg: ", err)
		}
		tt := &testCase{
			name: testCaseName,
			args: args{
				app:        application,
				statusTime: statusTime,
			},
			isSucceeded: isSucceeded,
			wantErr:     wantErr,
		}
		t.Run(tt.name, func(t *testing.T) {
			got, err := appService.UpdateDeploymentStatusAndCheckIsSucceeded(tt.args.app, tt.args.statusTime, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateDeploymentStatusAndCheckIsSucceeded() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.isSucceeded {
				t.Errorf("UpdateDeploymentStatusAndCheckIsSucceeded() got = %v, want %v", got, tt.isSucceeded)
			}
		})

	}
}

func InitAppService() *app2.AppServiceImpl {
	logger, err := util.NewSugardLogger()
	if err != nil {
		log.Fatal("error in getting logger, AppService_test", "err", err)
	}

	sqlConfig, err := sql.GetConfig()
	if err != nil {
		log.Fatal("error in getting sql config, AppService_test", "err", err)
	}
	dbConnection, err := sql.NewDbConnection(sqlConfig, logger)
	if err != nil {
		log.Fatal("error in getting db connection, AppService_test", "err", err)
	}

	pipelineOverrideRepository := chartConfig.NewPipelineOverrideRepository(dbConnection)
	pipelineRepository := pipelineConfig.NewPipelineRepositoryImpl(dbConnection, logger)
	httpClient := util.NewHttpClient()
	eventClientConfig, err := client1.GetEventClientConfig()
	pubSubClient := pubsub.NewPubSubClientServiceImpl(logger)
	ciPipelineRepositoryImpl := pipelineConfig.NewCiPipelineRepositoryImpl(dbConnection, logger)
	attributesRepositoryImpl := repository.NewAttributesRepositoryImpl(dbConnection)
	serverEnvConfig, err := serverEnvConfig.ParseServerEnvConfig()
	if err != nil {
		log.Fatal("error in getting server env config, AppService_test", "err", err)
	}
	moduleRepositoryImpl := moduleRepo.NewModuleRepositoryImpl(dbConnection)
	moduleActionAuditLogRepository := module.NewModuleActionAuditLogRepositoryImpl(dbConnection)
	clusterRepository := repository1.NewClusterRepositoryImpl(dbConnection, logger)
	clusterService := cluster.NewClusterServiceImplExtended(clusterRepository, nil, nil, logger, nil, nil, nil, nil, nil, nil, nil, nil)
	helmClientConfig, err := client.GetConfig()
	if err != nil {
		log.Fatal("error in getting server helm client config, AppService_test", "err", err)
	}
	helmAppClient := client.NewHelmAppClientImpl(logger, helmClientConfig)
	helmAppService := client.NewHelmAppServiceImpl(logger, clusterService, helmAppClient, nil, nil, nil, serverEnvConfig, nil, nil, nil, nil, nil, nil, nil, nil)
	moduleService := module.NewModuleServiceImpl(logger, serverEnvConfig, moduleRepositoryImpl, moduleActionAuditLogRepository, helmAppService, nil, nil, nil, nil, nil, nil, nil)
	eventClient := client1.NewEventRESTClientImpl(logger, httpClient, eventClientConfig, pubSubClient, ciPipelineRepositoryImpl,
		pipelineRepository, attributesRepositoryImpl, moduleService)
	cdWorkflowRepository := pipelineConfig.NewCdWorkflowRepositoryImpl(dbConnection, logger)
	ciWorkflowRepository := pipelineConfig.NewCiWorkflowRepositoryImpl(dbConnection, logger)
	ciPipelineMaterialRepository := pipelineConfig.NewCiPipelineMaterialRepositoryImpl(dbConnection, logger)
	userRepository := repository2.NewUserRepositoryImpl(dbConnection, logger)
	eventFactory := client1.NewEventSimpleFactoryImpl(logger, cdWorkflowRepository, pipelineOverrideRepository, ciWorkflowRepository,
		ciPipelineMaterialRepository, ciPipelineRepositoryImpl, pipelineRepository, userRepository, nil, nil, nil, nil, nil)
	appListingRepositoryQueryBuilder := helper.NewAppListingRepositoryQueryBuilder(logger)
	appListingRepository := repository.NewAppListingRepositoryImpl(logger, dbConnection, appListingRepositoryQueryBuilder, nil)
	appRepository := app.NewAppRepositoryImpl(dbConnection, logger)
	chartRepository := chartRepoRepository.NewChartRepository(dbConnection)
	pipelineStatusTimelineResourcesRepository := pipelineConfig.NewPipelineStatusTimelineResourcesRepositoryImpl(dbConnection, logger)
	pipelineStatusTimelineResourcesService := status.NewPipelineStatusTimelineResourcesServiceImpl(dbConnection, logger, pipelineStatusTimelineResourcesRepository)
	pipelineStatusTimelineRepository := pipelineConfig.NewPipelineStatusTimelineRepositoryImpl(dbConnection, logger)
	pipelineStatusSyncDetailRepository := pipelineConfig.NewPipelineStatusSyncDetailRepositoryImpl(dbConnection, logger)
	pipelineStatusSyncDetailService := status.NewPipelineStatusSyncDetailServiceImpl(logger, pipelineStatusSyncDetailRepository)
	pipelineStatusTimelineService := status.NewPipelineStatusTimelineServiceImpl(logger, pipelineStatusTimelineRepository, cdWorkflowRepository, nil, pipelineStatusTimelineResourcesService, pipelineStatusSyncDetailService, nil, nil)
	refChartDir := chartRepoRepository.RefChartDir("scripts/devtron-reference-helm-charts")
	appService := app2.NewAppService(nil, pipelineOverrideRepository, nil, logger, nil,
		pipelineRepository, nil, eventClient, eventFactory, nil, nil, nil, nil, nil, nil,
		appListingRepository, appRepository, nil, nil, nil, nil, nil,
		chartRepository, nil, cdWorkflowRepository, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, refChartDir, nil,
		nil, nil, nil, pipelineStatusTimelineRepository, nil, nil, nil,
		nil, nil, pipelineStatusTimelineResourcesService, pipelineStatusSyncDetailService, pipelineStatusTimelineService,
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	return appService
}
