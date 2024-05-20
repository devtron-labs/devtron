package adapter

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	pipelineConfigBean "github.com/devtron-labs/devtron/pkg/pipeline/bean"
)

func GetRequirementReqForCreateRequest(reqBean *bean.DtResourceObjectCreateReqBean, objectDataPath string, skipJsonSchemaValidation bool) *bean.ResourceObjectRequirementRequest {
	return &bean.ResourceObjectRequirementRequest{
		ReqBean: &bean.DtResourceObjectInternalBean{
			DevtronResourceObjectDescriptorBean: reqBean.DevtronResourceObjectDescriptorBean,
			Overview:                            reqBean.Overview,
			ParentConfig:                        reqBean.ParentConfig,
		},
		ObjectDataPath:           objectDataPath,
		SkipJsonSchemaValidation: skipJsonSchemaValidation,
	}
}

func GetRequirementRequestForCatalogRequest(reqBean *bean.DtResourceObjectCatalogReqBean, skipJsonSchemaValidation bool) *bean.ResourceObjectRequirementRequest {
	return &bean.ResourceObjectRequirementRequest{
		ReqBean: &bean.DtResourceObjectInternalBean{
			DevtronResourceObjectDescriptorBean: reqBean.DevtronResourceObjectDescriptorBean,
			ObjectData:                          reqBean.ObjectData,
		},
		ObjectDataPath:           bean.ResourceObjectMetadataPath,
		SkipJsonSchemaValidation: skipJsonSchemaValidation,
	}
}

func GetRequirementRequestForDependenciesRequest(reqBean *bean.DtResourceObjectDependenciesReqBean, skipJsonSchemaValidation bool) (*bean.ResourceObjectRequirementRequest, error) {
	marshaledDependencies, err := json.Marshal(reqBean.Dependencies)
	if err != nil {
		return nil, err
	}
	return &bean.ResourceObjectRequirementRequest{
		ReqBean: &bean.DtResourceObjectInternalBean{
			DevtronResourceObjectDescriptorBean: reqBean.DevtronResourceObjectDescriptorBean,
			Dependencies:                        reqBean.Dependencies,
			ObjectData:                          string(marshaledDependencies),
		},
		ObjectDataPath:           bean.ResourceObjectDependenciesPath,
		SkipJsonSchemaValidation: skipJsonSchemaValidation,
	}, nil
}

func BuildConfigStatusSchemaData(status *bean.ConfigStatus) *bean.ReleaseConfigSchema {
	return &bean.ReleaseConfigSchema{
		Status:  status.Status.ToString(),
		Comment: status.Comment,
		Lock:    status.IsLocked,
	}
}

func CreateDependencyData(id, devtronResourceId, devtronResourceSchemaId int, maxIndex float64, dependencyType bean.DevtronResourceDependencyType, idType bean.IdType) *bean.DevtronResourceDependencyBean {
	return &bean.DevtronResourceDependencyBean{
		OldObjectId:             id,
		IdType:                  idType,
		Index:                   int(maxIndex),
		TypeOfDependency:        dependencyType,
		DevtronResourceId:       devtronResourceId,
		DevtronResourceSchemaId: devtronResourceSchemaId,
		Dependencies:            make([]*bean.DevtronResourceDependencyBean, 0),
	}
}

func GetSuccessPassResponse() *bean.SuccessResponse {
	return &bean.SuccessResponse{
		Success: true,
	}
}

func BuildUserSchemaData(id int32, emailId string) *bean.UserSchema {
	return &bean.UserSchema{
		Id:   id,
		Name: emailId,
		Icon: true,
	}
}

func BuildFilterCriteriaDecoder(resource, identifierType, value string) *bean.FilterCriteriaDecoder {
	return &bean.FilterCriteriaDecoder{
		Resource: bean.DevtronResourceKind(resource),
		Type:     bean.FilterCriteriaIdentifier(identifierType),
		Value:    value,
	}
}

func BuildSearchCriteriaDecoder(resource, value string) *bean.SearchCriteriaDecoder {
	return &bean.SearchCriteriaDecoder{
		SearchBy: bean.SearchPropertyBy(resource),
		Value:    value,
	}
}

func BuildDevtronResourceObjectGetAPIBean() *bean.DevtronResourceObjectGetAPIBean {
	return &bean.DevtronResourceObjectGetAPIBean{
		DevtronResourceObjectDescriptorBean: &bean.DevtronResourceObjectDescriptorBean{},
		DevtronResourceObjectBasicDataBean: &bean.DevtronResourceObjectBasicDataBean{
			Overview: &bean.ResourceOverview{},
		},
	}
}

func BuildDevtronResourceObjectDescriptorBean(id int, kind, subKind bean.DevtronResourceKind,
	version bean.DevtronResourceVersion, userId int32) *bean.DevtronResourceObjectDescriptorBean {
	reqBean := &bean.DevtronResourceObjectDescriptorBean{
		Kind:    kind.ToString(),
		SubKind: subKind.ToString(),
		Version: version.ToString(),
		UserId:  userId,
	}
	SetIdTypeAndResourceIdBasedOnKind(reqBean, id)
	return reqBean
}

func BuildCdPipelineEnvironmentBasicData(envName, deploymentAppType string, envId, pipelineId int) *bean.CdPipelineEnvironment {
	return &bean.CdPipelineEnvironment{
		Name:              envName,
		Id:                envId,
		PipelineId:        pipelineId,
		DeploymentAppType: deploymentAppType,
	}
}

func BuildChildObject(data interface{}, dataType bean.ChildObjectType) *bean.ChildObject {
	return &bean.ChildObject{
		Data: data,
		Type: dataType,
	}
}

func BuildGitCommit(author string, branch string, message string, modifiedTime string, revision string, tag string, webhookData *bean.WebHookMaterialInfo, url string) bean.GitCommitData {
	return bean.GitCommitData{
		Author: author, Branch: branch, ModifiedTime: modifiedTime, Message: message, Revision: revision, Tag: tag, Url: url, WebhookData: webhookData,
	}

}
func BuildWebHookMaterialInfo(id int, eventActionType string, data interface{}) *bean.WebHookMaterialInfo {
	return &bean.WebHookMaterialInfo{Id: id, EventActionType: eventActionType, Data: data}
}

func GetDefaultCdPipelineSelector() []string {
	return []string{bean.DefaultCdPipelineSelector}
}

func BuildTaskExecutionResponseBean(appId, envId int, appName, envName string, isVirtualEnv bool, feasibilityStatus, processingStatus error) *bean.TaskExecutionResponseBean {
	return &bean.TaskExecutionResponseBean{
		AppId:         appId,
		EnvId:         envId,
		AppName:       appName,
		EnvName:       envName,
		IsVirtualEnv:  isVirtualEnv,
		Feasibility:   feasibilityStatus,
		TriggerStatus: processingStatus,
	}
}

func BuildDevtronResourceTaskRunBean(descriptorBean *bean.DevtronResourceObjectDescriptorBean, overview *bean.ResourceOverview, actions []*bean.TaskRunAction) *bean.DevtronResourceTaskRunBean {
	return &bean.DevtronResourceTaskRunBean{
		DevtronResourceObjectDescriptorBean: descriptorBean,
		Overview:                            overview,
		Action:                              actions,
	}
}
func BuildRunSourceObject(id int, idType bean.IdType, drId, drSchemaId int, dependency *bean.DependencyDetail) *bean.RunSource {
	return &bean.RunSource{
		Id:                      id,
		IdType:                  idType,
		DevtronResourceId:       drId,
		DevtronResourceSchemaId: drSchemaId,
		DependencyDetail:        dependency,
	}
}

func BuildDependencyDetail(id int, idType bean.IdType, drId, drSchemaId int) *bean.DependencyDetail {
	return &bean.DependencyDetail{
		Id:                      id,
		IdType:                  idType,
		DevtronResourceId:       drId,
		DevtronResourceSchemaId: drSchemaId,
	}
}

func BuildActionObject(taskType bean.TaskType, cdWfrId int) *bean.TaskRunAction {
	return &bean.TaskRunAction{
		TaskType:           taskType,
		CdWorkflowRunnerId: cdWfrId,
	}
}

// MapDependenciesByDependentOnIndex will map the []*bean.DevtronResourceDependencyBean
// corresponding to DependentOnIndex OR DependentOnIndexes
// here Independent bean.DevtronResourceDependencyBean are mapped to Key -> 0
func MapDependenciesByDependentOnIndex(dependencies []*bean.DevtronResourceDependencyBean) map[int][]*bean.DevtronResourceDependencyBean {
	response := make(map[int][]*bean.DevtronResourceDependencyBean)
	for _, dependency := range dependencies {
		if dependency.DependentOnIndex != 0 {
			response[dependency.DependentOnIndex] = append(response[dependency.DependentOnIndex], dependency)
		} else if len(dependency.DependentOnIndexes) != 0 {
			for _, dependentOnIndex := range dependency.DependentOnIndexes {
				if dependentOnIndex != 0 {
					response[dependentOnIndex] = append(response[dependentOnIndex], dependency)
				}
			}
		} else {
			response[0] = append(response[0], dependency)
		}
	}
	return response
}

func NewCdPipelineReleaseInfo(appId, envId, pipelineId int, appName, envName string, deleteRequest bool) *bean.CdPipelineReleaseInfo {
	return &bean.CdPipelineReleaseInfo{
		AppId:                      appId,
		AppName:                    appName,
		EnvId:                      envId,
		EnvName:                    envName,
		PipelineId:                 pipelineId,
		DeploymentAppDeleteRequest: deleteRequest,
		DeployStatus:               pipelineConfigBean.NotTriggered,
		PreStatus:                  pipelineConfigBean.NotTriggered,
		PostStatus:                 pipelineConfigBean.NotTriggered,
	}
}

func GetReleaseConfigAndLockStatusChangeSuccessResponse(statusTo bean.ReleaseConfigStatus,
	isLocked bool) *bean.SuccessResponse {
	statusUserMessage, statusDetailMessage := getReleaseConfigStatusSuccessChangeMessage(statusTo)
	lockUserMessage := getReleaseLockStatusSuccessChangeMessage(isLocked)
	resp := &bean.SuccessResponse{
		Success:       true,
		UserMessage:   fmt.Sprintf("%s & %s", statusUserMessage, lockUserMessage),
		DetailMessage: statusDetailMessage,
	}
	return resp
}

func GetReleaseConfigStatusChangeSuccessResponse(statusTo bean.ReleaseConfigStatus) *bean.SuccessResponse {
	resp := &bean.SuccessResponse{
		Success: true,
	}
	resp.UserMessage, resp.DetailMessage = getReleaseConfigStatusSuccessChangeMessage(statusTo)
	return resp
}

func GetReleaseLockStatusChangeSuccessResponse(isLocked bool) *bean.SuccessResponse {
	resp := &bean.SuccessResponse{
		Success: true,
	}
	resp.UserMessage = getReleaseLockStatusSuccessChangeMessage(isLocked)
	return resp
}

func getReleaseConfigStatusSuccessChangeMessage(statusTo bean.ReleaseConfigStatus) (string, string) {
	statusMessage := ""
	detailMessage := ""
	switch statusTo {
	case bean.DraftReleaseConfigStatus:
		statusMessage = fmt.Sprintf("Status is changed to '%s'", "Draft")
	case bean.ReadyForReleaseConfigStatus:
		statusMessage = fmt.Sprintf("Status is changed to '%s'", "Ready for release")
	case bean.HoldReleaseConfigStatus:
		statusMessage = fmt.Sprintf("Release is '%s'", "On hold")
		detailMessage = bean.ReleaseHoldStatusChangeSuccessDetailMessage
	case bean.RescindReleaseConfigStatus:
		statusMessage = fmt.Sprintf("Release is '%s'", "Rescinded")
		detailMessage = bean.ReleaseRescindStatusChangeSuccessDetailMessage
	case bean.CorruptedReleaseConfigStatus:
		statusMessage = fmt.Sprintf("Release is '%s'", "Corrupted")
	}
	return statusMessage, detailMessage
}

func getReleaseLockStatusSuccessChangeMessage(isLocked bool) string {
	statusMessage := ""
	if isLocked {
		statusMessage = bean.ReleaseLockStatusChangeSuccessMessage
	} else {
		statusMessage = bean.ReleaseUnLockStatusChangeSuccessMessage
	}
	return statusMessage
}

func BuildReleaseDeploymentStatus(ongoing, yetToTrigger, failed, completed int) *bean.ReleaseDeploymentStatusCount {
	return &bean.ReleaseDeploymentStatusCount{
		AllDeployment: yetToTrigger + failed + completed + ongoing,
		YetToTrigger:  yetToTrigger,
		Failed:        failed,
		Completed:     completed,
		Ongoing:       ongoing,
	}
}

func BuildTaskInfoCount(releaseDeploymentStatus *bean.ReleaseDeploymentStatusCount, stageWiseStatusCount *bean.StageWiseStatusCount) *bean.TaskInfoCount {
	return &bean.TaskInfoCount{
		ReleaseDeploymentStatusCount: releaseDeploymentStatus,
		StageWiseStatusCount:         stageWiseStatusCount,
	}
}
func BuildStageWiseStatusCount(preStatusCount *bean.PrePostStatusCount, deployCount *bean.DeploymentCount, postStatusCount *bean.PrePostStatusCount) *bean.StageWiseStatusCount {
	return &bean.StageWiseStatusCount{
		PreStatusCount:  preStatusCount,
		DeploymentCount: deployCount,
		PostStatusCount: postStatusCount,
	}
}

func BuildDeploymentCount(notTriggered, failed, succeeded, queued, inProgress, others int) *bean.DeploymentCount {
	return &bean.DeploymentCount{
		NotTriggered: notTriggered,
		Failed:       failed,
		Succeeded:    succeeded,
		Queued:       queued,
		InProgress:   inProgress,
		Others:       others,
	}
}

func BuildPreOrPostDeploymentCount(notTriggered, failed, succeeded, inProgress, others int) *bean.PrePostStatusCount {
	return &bean.PrePostStatusCount{
		NotTriggered: notTriggered,
		Failed:       failed,
		Succeeded:    succeeded,
		InProgress:   inProgress,
		Others:       others,
	}
}
