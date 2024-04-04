package in

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	v1alpha12 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/FullMode"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	"github.com/devtron-labs/devtron/pkg/workflow/dag"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"k8s.io/utils/pointer"
	"time"
)

type DeployedApplicationEventProcessorImpl struct {
	logger                    *zap.SugaredLogger
	pubSubClient              *pubsub.PubSubClientServiceImpl
	appService                app.AppService
	gitOpsConfigReadService   config.GitOpsConfigReadService
	installedAppService       FullMode.InstalledAppDBExtendedService
	workflowDagExecutor       dag.WorkflowDagExecutor
	cdWorkflowCommonService   cd.CdWorkflowCommonService
	pipelineBuilder           pipeline.PipelineBuilder
	appStoreDeploymentService service.AppStoreDeploymentService

	pipelineRepository     pipelineConfig.PipelineRepository
	installedAppRepository repository.InstalledAppRepository
}

func NewDeployedApplicationEventProcessorImpl(logger *zap.SugaredLogger,
	pubSubClient *pubsub.PubSubClientServiceImpl,
	appService app.AppService,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	installedAppService FullMode.InstalledAppDBExtendedService,
	workflowDagExecutor dag.WorkflowDagExecutor,
	cdWorkflowCommonService cd.CdWorkflowCommonService,
	pipelineBuilder pipeline.PipelineBuilder,
	appStoreDeploymentService service.AppStoreDeploymentService,
	pipelineRepository pipelineConfig.PipelineRepository,
	installedAppRepository repository.InstalledAppRepository) *DeployedApplicationEventProcessorImpl {
	deployedApplicationEventProcessorImpl := &DeployedApplicationEventProcessorImpl{
		logger:                    logger,
		pubSubClient:              pubSubClient,
		appService:                appService,
		gitOpsConfigReadService:   gitOpsConfigReadService,
		installedAppService:       installedAppService,
		workflowDagExecutor:       workflowDagExecutor,
		cdWorkflowCommonService:   cdWorkflowCommonService,
		pipelineBuilder:           pipelineBuilder,
		appStoreDeploymentService: appStoreDeploymentService,

		pipelineRepository:     pipelineRepository,
		installedAppRepository: installedAppRepository,
	}
	return deployedApplicationEventProcessorImpl
}

func (impl *DeployedApplicationEventProcessorImpl) SubscribeArgoAppUpdate() error {
	callback := func(msg *model.PubSubMsg) {
		applicationDetail := bean3.ApplicationDetail{}
		err := json.Unmarshal([]byte(msg.Data), &applicationDetail)
		if err != nil {
			impl.logger.Errorw("unmarshal error on app update status", "err", err)
			return
		}
		app := applicationDetail.Application
		if app == nil {
			return
		}
		if applicationDetail.StatusTime.IsZero() {
			applicationDetail.StatusTime = time.Now()
		}
		isAppStoreApplication := false
		_, err = impl.pipelineRepository.GetArgoPipelineByArgoAppName(app.ObjectMeta.Name)
		if err != nil && err == pg.ErrNoRows {
			impl.logger.Infow("this app not found in pipeline table looking in installed_apps table", "appName", app.ObjectMeta.Name)
			// if not found in pipeline table then search in installed_apps table
			installedAppModel, err := impl.installedAppRepository.GetInstalledAppByGitOpsAppName(app.ObjectMeta.Name)
			if err == pg.ErrNoRows {
				// no installed_apps found
				impl.logger.Errorw("no installed apps found", "err", err)
				return
			}
			if err != nil {
				impl.logger.Errorw("error in getting all gitops deployment app names from installed_apps ", "err", err)
				return
			}
			if installedAppModel.Id > 0 {
				// app found in installed_apps table hence setting flag to true
				isAppStoreApplication = true
			} else {
				// app neither found in installed_apps nor in pipeline table hence returning
				return
			}
		}
		isSucceeded, pipelineOverride, err := impl.appService.UpdateDeploymentStatusAndCheckIsSucceeded(app, applicationDetail.StatusTime, isAppStoreApplication)
		if err != nil {
			impl.logger.Errorw("error on application status update", "err", err, "msg", string(msg.Data))
			// TODO - check update for charts - fix this call
			if err == pg.ErrNoRows {
				// if not found in charts (which is for devtron apps) try to find in installed app (which is for devtron charts)
				_, err := impl.installedAppService.UpdateInstalledAppVersionStatus(app)
				if err != nil {
					impl.logger.Errorw("error on application status update", "err", err, "msg", string(msg.Data))
					return
				}
			}
			// return anyway whether updates or failure, no further processing for charts status update
			return
		}

		// invoke DagExecutor, for cd success which will trigger post stage if exist.
		if isSucceeded {
			impl.logger.Debugw("git hash history", "list", app.Status.History)
			triggerContext := bean2.TriggerContext{
				ReferenceId: pointer.String(msg.MsgId),
			}
			err = impl.workflowDagExecutor.HandleDeploymentSuccessEvent(triggerContext, pipelineOverride)
			if err != nil {
				impl.logger.Errorw("deployment success event error", "pipelineOverride", pipelineOverride, "err", err)
				return
			}
		}
		impl.logger.Debugw("application status update completed", "app", app.Name)
	}

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		return "", nil
	}

	validations := impl.cdWorkflowCommonService.GetTriggerValidateFuncs()
	err := impl.pubSubClient.Subscribe(pubsub.APPLICATION_STATUS_UPDATE_TOPIC, callback, loggerFunc, validations...)
	if err != nil {
		impl.logger.Error(err)
		return err
	}
	return nil
}

func (impl *DeployedApplicationEventProcessorImpl) SubscribeArgoAppDeleteStatus() error {
	callback := func(msg *model.PubSubMsg) {

		applicationDetail := bean3.ApplicationDetail{}
		err := json.Unmarshal([]byte(msg.Data), &applicationDetail)
		if err != nil {
			impl.logger.Errorw("unmarshal error on app delete status", "err", err)
			return
		}
		app := applicationDetail.Application
		if app == nil {
			return
		}
		impl.logger.Infow("argo delete event received", "appName", app.Name, "namespace", app.Namespace, "deleteTimestamp", app.DeletionTimestamp)

		err = impl.updateArgoAppDeleteStatus(app)
		if err != nil {
			impl.logger.Errorw("error in updating pipeline delete status", "err", err, "appName", app.Name)
		}
	}

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		applicationDetail := bean3.ApplicationDetail{}
		err := json.Unmarshal([]byte(msg.Data), &applicationDetail)
		if err != nil {
			return "unmarshal error on app delete status", []interface{}{"err", err}
		}
		return "got message for application status delete", []interface{}{"appName", applicationDetail.Application.Name, "namespace", applicationDetail.Application.Namespace, "deleteTimestamp", applicationDetail.Application.DeletionTimestamp}
	}

	err := impl.pubSubClient.Subscribe(pubsub.APPLICATION_STATUS_DELETE_TOPIC, callback, loggerFunc)
	if err != nil {
		impl.logger.Errorw("error in subscribing to argo application status delete topic", "err", err)
		return err
	}
	return nil
}

func (impl *DeployedApplicationEventProcessorImpl) updateArgoAppDeleteStatus(app *v1alpha12.Application) error {
	pipeline, err := impl.pipelineRepository.GetArgoPipelineByArgoAppName(app.ObjectMeta.Name)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching pipeline from Pipeline Repository", "err", err)
		return err
	}
	if pipeline.Deleted == true {
		impl.logger.Errorw("invalid nats message, pipeline already deleted")
		return errors.New("invalid nats message, pipeline already deleted")
	}
	if err == pg.ErrNoRows {
		// Helm app deployed using argocd
		var gitHash string
		if app.Operation != nil && app.Operation.Sync != nil {
			gitHash = app.Operation.Sync.Revision
		} else if app.Status.OperationState != nil && app.Status.OperationState.Operation.Sync != nil {
			gitHash = app.Status.OperationState.Operation.Sync.Revision
		}
		model, err := impl.installedAppRepository.GetInstalledAppByGitHash(gitHash)
		if err != nil {
			impl.logger.Errorw("error in fetching installed app by git hash from installed app repository", "err", err)
			return err
		}

		// Check to ensure that delete request for app was received
		installedApp, err := impl.installedAppService.CheckAppExistsByInstalledAppId(model.InstalledAppId)
		if err == pg.ErrNoRows {
			impl.logger.Errorw("App not found in database", "installedAppId", model.InstalledAppId, "err", err)
			return fmt.Errorf("app not found in database %s", err)
		} else if installedApp.DeploymentAppDeleteRequest == false {
			// TODO 4465 remove app from log after final RCA
			impl.logger.Infow("Deployment delete not requested for app, not deleting app from DB", "appName", app.Name, "app", app)
			return nil
		}

		deleteRequest := &appStoreBean.InstallAppVersionDTO{}
		deleteRequest.ForceDelete = false
		deleteRequest.NonCascadeDelete = false
		deleteRequest.AcdPartialDelete = false
		deleteRequest.InstalledAppId = model.InstalledAppId
		deleteRequest.AppId = model.AppId
		deleteRequest.AppName = model.AppName
		deleteRequest.Namespace = model.Namespace
		deleteRequest.ClusterId = model.ClusterId
		deleteRequest.EnvironmentId = model.EnvironmentId
		deleteRequest.AppOfferingMode = model.AppOfferingMode
		deleteRequest.UserId = 1
		_, err = impl.appStoreDeploymentService.DeleteInstalledApp(context.Background(), deleteRequest)
		if err != nil {
			impl.logger.Errorw("error in deleting installed app", "err", err)
			return err
		}
	} else {
		// devtron app
		_, err = impl.pipelineBuilder.DeleteCdPipeline(&pipeline, context.Background(), bean.FORCE_DELETE, false, 1)
		if err != nil {
			impl.logger.Errorw("error in deleting cd pipeline", "err", err)
			return err
		}
	}
	return nil
}
