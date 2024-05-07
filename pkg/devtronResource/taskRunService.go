package devtronResource

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/helper"
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/pkg/workflow/cd/util"
	"github.com/go-pg/pg"
	"github.com/tidwall/sjson"
	"time"
)

func (impl *DevtronResourceServiceImpl) getOverviewForTaskRunObject(userId int32, recordedTime time.Time, id int, idType bean.IdType, drId, drSchemaId int, drDetails *bean.DependencyDetail) (*bean.ResourceOverview, error) {
	createdByDetails, err := impl.getUserSchemaDataById(userId)
	// considering the user details are already verified; this error indicates to an internal db error.
	if err != nil {
		impl.logger.Errorw("error encountered in populateDefaultValuesForCreateReleaseTrackRequest", "userId", userId, "err", err)
		return nil, err
	}

	overviewObject := &bean.ResourceOverview{
		CreatedBy: createdByDetails,
		CreatedOn: recordedTime,
		RunSource: adapter.BuildRunSourceObject(id, idType, drId, drSchemaId, drDetails),
	}
	return overviewObject, nil

}

func (impl *DevtronResourceServiceImpl) getObjectDataJsonForTaskRun(dRTaskRunBean *bean.DevtronResourceTaskRunBean) (objectData string, err error) {
	reqBean := &bean.DtResourceObjectInternalBean{
		DevtronResourceObjectDescriptorBean: dRTaskRunBean.DevtronResourceObjectDescriptorBean,
		Overview:                            dRTaskRunBean.Overview,
	}
	objectData, err = impl.setDevtronManagedFieldsInObjectData(objectData, reqBean)
	if err != nil {
		impl.logger.Errorw("error, setDevtronManagedFieldsInObjectData", "err", err, "req", reqBean)
		return objectData, err
	}
	objectData, err = impl.setUserProvidedFieldsInObjectData(objectData, reqBean)
	if err != nil {
		impl.logger.Errorw("error, setUserProvidedFieldsInObjectData", "err", err, "req", reqBean)
		return objectData, err
	}
	objectData, err = setActionForTaskRun(objectData, dRTaskRunBean.Action)
	if err != nil {
		impl.logger.Errorw("error, setActionForTaskRun", "err", err, "action", dRTaskRunBean.Action)
		return objectData, err
	}

	return objectData, nil
}

// bulkCreateDevtronResourceTaskRunObjects bulk creates devtron resource task run objects for every task run
func (impl *DevtronResourceServiceImpl) bulkCreateDevtronResourceTaskRunObjects(tx *pg.Tx, req *bean.DevtronResourceTaskExecutionBean, appVsArtifactIdMap map[int]int, pipelineCiArtifactKeyVsWorkflowIdMap map[string]int, cdWorkflowIdVsRunnerIdMap map[int]int, appIdVsDrSchemaDetails map[int]*bean.DependencyDetail, existingObject *repository.DevtronResourceObject) ([]*repository.DevtronResourceTaskRun, error) {
	tasks := req.Tasks
	taskRuns := make([]*repository.DevtronResourceTaskRun, 0, len(tasks))
	// get devtron schema id, ignoring it for now as of not creating schema , can be done in future
	//taskRunSchema, err := impl.devtronResourceSchemaRepository.FindSchemaByKindSubKindAndVersion(bean.DevtronResourceTaskRun.ToString(), "", bean.DevtronResourceVersionAlpha1.ToString())
	//if err != nil {
	//	impl.logger.Errorw("error encountered in bulkCreateDevtronResourceTaskRunObjects", "err", err)
	//	return err
	//}

	for _, task := range tasks {
		//fetching  cd workflow runner id from logic -
		// STEP 1: get artifact id from task app id
		// STEP 2: get cd workflow id from task pipeline id and artifact id
		// STEP 3: get cd workflow runner id from cd workflow id
		cdWorkflowId, found := pipelineCiArtifactKeyVsWorkflowIdMap[util2.GetKeyForPipelineIdAndArtifact(task.PipelineId, appVsArtifactIdMap[task.AppId])]
		if !found {
			continue
		}
		cdWorkflowRunnerId := cdWorkflowIdVsRunnerIdMap[cdWorkflowId]
		dependencyDetail := appIdVsDrSchemaDetails[task.AppId]

		// building objects from adapters with req data set
		//sending id as zero as task run object is still not created and we dont keep id in schema task run resource object
		descriptorBean := adapter.BuildDevtronResourceObjectDescriptorBean(0, bean.DevtronResourceTaskRun, "", bean.DevtronResourceVersionAlpha1, req.UserId)
		overview, err := impl.getOverviewForTaskRunObject(req.UserId, req.TriggeredTime, req.Id, req.IdType, existingObject.DevtronResourceId, existingObject.DevtronResourceSchemaId, dependencyDetail)
		if err != nil {
			impl.logger.Errorw("error encountered in bulkCreateDevtronResourceTaskRunObjects", "err", err, "task", task)
			return nil, err
		}
		action := adapter.BuildActionObject(helper.GetTaskTypeBasedOnWorkflowType(task.CdWorkflowType), cdWorkflowRunnerId)
		drTaskRunBean := adapter.BuildDevtronResourceTaskRunBean(descriptorBean, overview, []*bean.TaskRunAction{action})
		// got required data.
		taskJson, err := impl.getObjectDataJsonForTaskRun(drTaskRunBean)
		if err != nil {
			impl.logger.Errorw("error encountered in bulkCreateDevtronResourceTaskRunObjects", "err", err, "drTaskRunBean", drTaskRunBean)
			return nil, err
		}
		taskRuns = append(taskRuns, &repository.DevtronResourceTaskRun{
			TaskJson:                      taskJson,
			RunSourceIdentifier:           helper.GetTaskRunSourceIdentifier(req.Id, req.IdType, existingObject.DevtronResourceId, existingObject.DevtronResourceSchemaId),
			RunSourceDependencyIdentifier: helper.GetTaskRunSourceDependencyIdentifier(dependencyDetail.Id, dependencyDetail.IdType, dependencyDetail.DevtronResourceId, dependencyDetail.DevtronResourceSchemaId),
			TaskType:                      helper.GetTaskTypeBasedOnWorkflowType(task.CdWorkflowType),
			TaskTypeIdentifier:            cdWorkflowRunnerId,
			//DevtronResourceSchemaId:       taskRunSchema.Id,
			AuditLog: sql.AuditLog{CreatedOn: req.TriggeredTime, UpdatedOn: req.TriggeredTime, CreatedBy: req.UserId, UpdatedBy: req.UserId},
		})
	}
	err := impl.dtResourceTaskRunRepository.BulkCreate(tx, taskRuns)
	if err != nil {
		impl.logger.Errorw("error encountered in bulkCreateDevtronResourceTaskRunObjects", "taskRuns", taskRuns, "err", err)
		return nil, err
	}
	return taskRuns, nil
}

func (impl *DevtronResourceServiceImpl) updateUserProvidedDataInTaskRunObj(objectData string, reqBean *bean.DtResourceObjectInternalBean) (string, error) {
	var err error
	if reqBean.Overview != nil {
		objectData, err = impl.setTaskRunOverviewFieldsInObjectData(objectData, reqBean.Overview)
		if err != nil {
			impl.logger.Errorw("error in setting overview data in schema", "err", err, "request", reqBean)
			return objectData, err
		}
	}
	return objectData, nil
}

func (impl *DevtronResourceServiceImpl) setTaskRunOverviewFieldsInObjectData(objectData string, overview *bean.ResourceOverview) (string, error) {
	var err error
	if overview.CreatedBy != nil && overview.CreatedBy.Id > 0 {
		objectData, err = sjson.Set(objectData, bean.ResourceObjectCreatedByPath, overview.CreatedBy)
		if err != nil {
			impl.logger.Errorw("error in setting createdBy in schema", "err", err, "overview", overview)
			return objectData, err
		}
		objectData, err = sjson.Set(objectData, bean.ResourceObjectCreatedOnPath, overview.CreatedOn)
		if err != nil {
			impl.logger.Errorw("error in setting createdOn in schema", "err", err, "overview", overview)
			return objectData, err
		}
	}
	if overview.RunSource != nil {
		objectData, err = sjson.Set(objectData, bean.ResourceObjectRunSourcePath, overview.RunSource)
		if err != nil {
			impl.logger.Errorw("error in setting run source in schema", "err", err, "overview", overview)
			return objectData, err
		}
	}
	return objectData, nil
}

func setActionForTaskRun(objectData string, actions []*bean.TaskRunAction) (string, error) {
	return sjson.Set(objectData, bean.ResourceTaskRunActionPath, actions)

}
