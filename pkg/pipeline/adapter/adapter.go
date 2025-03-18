/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package adapter

import (
	"encoding/json"
	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/ciPipeline"
	"github.com/devtron-labs/devtron/pkg/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/build/pipeline/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/cluster/environment/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	pipelineConfigBean "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"
)

func GetDockerConfigBean(dockerRegistry *dockerRegistryRepository.DockerArtifactStore) *types.DockerArtifactStoreBean {
	return &types.DockerArtifactStoreBean{
		Id:                 dockerRegistry.Id,
		RegistryType:       dockerRegistry.RegistryType,
		RegistryURL:        dockerRegistry.RegistryURL,
		Username:           dockerRegistry.Username,
		Password:           dockerRegistry.Password,
		AWSRegion:          dockerRegistry.AWSRegion,
		Connection:         dockerRegistry.Connection,
		Cert:               dockerRegistry.Cert,
		AWSAccessKeyId:     dockerRegistry.AWSAccessKeyId,
		AWSSecretAccessKey: dockerRegistry.AWSSecretAccessKey,
	}
}

func UpdateRegistryDetailsToWrfReq(cdStageWorkflowRequest *types.WorkflowRequest, dockerRegistry *types.DockerArtifactStoreBean) {
	cdStageWorkflowRequest.DockerUsername = dockerRegistry.Username
	cdStageWorkflowRequest.DockerPassword = dockerRegistry.Password
	cdStageWorkflowRequest.AwsRegion = dockerRegistry.AWSRegion
	cdStageWorkflowRequest.DockerConnection = dockerRegistry.Connection
	cdStageWorkflowRequest.DockerCert = dockerRegistry.Cert
	cdStageWorkflowRequest.AccessKey = dockerRegistry.AWSAccessKeyId
	cdStageWorkflowRequest.SecretKey = dockerRegistry.AWSSecretAccessKey
	cdStageWorkflowRequest.DockerRegistryType = string(dockerRegistry.RegistryType)
	cdStageWorkflowRequest.DockerRegistryURL = dockerRegistry.RegistryURL
	cdStageWorkflowRequest.DockerRegistryId = dockerRegistry.Id
}

func ConvertBuildConfigBeanToDbEntity(templateId int, overrideTemplateId int, ciBuildConfigBean *bean2.CiBuildConfigBean, userId int32) (*pipelineConfig.CiBuildConfig, error) {
	buildMetadata := ""
	ciBuildType := ciBuildConfigBean.CiBuildType
	if ciBuildType == bean2.BUILDPACK_BUILD_TYPE {
		buildPackConfigMetadataBytes, err := json.Marshal(ciBuildConfigBean.BuildPackConfig)
		if err != nil {
			return nil, err
		}
		buildMetadata = string(buildPackConfigMetadataBytes)
	} else if ciBuildType == bean2.SELF_DOCKERFILE_BUILD_TYPE || ciBuildType == bean2.MANAGED_DOCKERFILE_BUILD_TYPE {
		dockerBuildMetadataBytes, err := json.Marshal(ciBuildConfigBean.DockerBuildConfig)
		if err != nil {
			return nil, err
		}
		buildMetadata = string(dockerBuildMetadataBytes)
	}
	ciBuildConfigEntity := &pipelineConfig.CiBuildConfig{
		Id:                   ciBuildConfigBean.Id,
		Type:                 string(ciBuildType),
		CiTemplateId:         templateId,
		CiTemplateOverrideId: overrideTemplateId,
		BuildMetadata:        buildMetadata,
		AuditLog:             sql.AuditLog{UpdatedOn: time.Now(), UpdatedBy: userId},
		UseRootContext:       &ciBuildConfigBean.UseRootBuildContext,
	}
	return ciBuildConfigEntity, nil
}

func ConvertDbBuildConfigToBean(dbBuildConfig *pipelineConfig.CiBuildConfig) (*bean2.CiBuildConfigBean, error) {
	var buildPackConfig *bean2.BuildPackConfig
	var dockerBuildConfig *bean2.DockerBuildConfig
	var err error
	if dbBuildConfig == nil {
		return nil, nil
	}
	ciBuildType := bean2.CiBuildType(dbBuildConfig.Type)
	if ciBuildType == bean2.BUILDPACK_BUILD_TYPE {
		buildPackConfig, err = convertMetadataToBuildPackConfig(dbBuildConfig.BuildMetadata)
		if err != nil {
			return nil, err
		}
	} else if ciBuildType == bean2.SELF_DOCKERFILE_BUILD_TYPE || ciBuildType == bean2.MANAGED_DOCKERFILE_BUILD_TYPE {
		dockerBuildConfig, err = convertMetadataToDockerBuildConfig(dbBuildConfig.BuildMetadata)
		if err != nil {
			return nil, err
		}
	}
	useRootBuildContext := false
	//dbBuildConfig.UseRootContext will be nil if the entry in db never updated before
	if dbBuildConfig.UseRootContext == nil || *(dbBuildConfig.UseRootContext) {
		useRootBuildContext = true
	}
	ciBuildConfigBean := &bean2.CiBuildConfigBean{
		Id:                  dbBuildConfig.Id,
		CiBuildType:         ciBuildType,
		BuildPackConfig:     buildPackConfig,
		DockerBuildConfig:   dockerBuildConfig,
		UseRootBuildContext: useRootBuildContext,
	}
	return ciBuildConfigBean, nil
}

func convertMetadataToBuildPackConfig(buildConfMetadata string) (*bean2.BuildPackConfig, error) {
	buildPackConfig := &bean2.BuildPackConfig{}
	err := json.Unmarshal([]byte(buildConfMetadata), buildPackConfig)
	return buildPackConfig, err
}

func convertMetadataToDockerBuildConfig(dockerBuildMetadata string) (*bean2.DockerBuildConfig, error) {
	dockerBuildConfig := &bean2.DockerBuildConfig{}
	err := json.Unmarshal([]byte(dockerBuildMetadata), dockerBuildConfig)
	return dockerBuildConfig, err
}

func OverrideCiBuildConfig(dockerfilePath string, oldArgs string, ciLevelArgs string, dockerBuildOptions string, targetPlatform string, ciBuildConfigBean *bean2.CiBuildConfigBean) (*bean2.CiBuildConfigBean, error) {
	oldDockerArgs := map[string]string{}
	ciLevelDockerArgs := map[string]string{}
	dockerBuildOptionsMap := map[string]string{}
	if oldArgs != "" {
		if err := json.Unmarshal([]byte(oldArgs), &oldDockerArgs); err != nil {
			return nil, err
		}
	}
	if ciLevelArgs != "" {
		if err := json.Unmarshal([]byte(ciLevelArgs), &ciLevelDockerArgs); err != nil {
			return nil, err
		}
	}
	if dockerBuildOptions != "" {
		if err := json.Unmarshal([]byte(dockerBuildOptions), &dockerBuildOptionsMap); err != nil {
			return nil, err
		}
	}
	//no entry found in ci_build_config table, construct with requested data
	if ciBuildConfigBean == nil {
		dockerArgs := mergeMap(oldDockerArgs, ciLevelDockerArgs)
		ciBuildConfigBean = &bean2.CiBuildConfigBean{
			CiBuildType: bean2.SELF_DOCKERFILE_BUILD_TYPE,
			DockerBuildConfig: &bean2.DockerBuildConfig{
				DockerfilePath:     dockerfilePath,
				Args:               dockerArgs,
				TargetPlatform:     targetPlatform,
				DockerBuildOptions: dockerBuildOptionsMap,
				BuildContext:       "",
			},
			//setting true as default
			UseRootBuildContext: true,
		}
	} else if ciBuildConfigBean.CiBuildType == bean2.SELF_DOCKERFILE_BUILD_TYPE || ciBuildConfigBean.CiBuildType == bean2.MANAGED_DOCKERFILE_BUILD_TYPE {
		dockerBuildConfig := ciBuildConfigBean.DockerBuildConfig
		dockerArgs := mergeMap(dockerBuildConfig.Args, ciLevelDockerArgs)
		//dockerBuildConfig.DockerfilePath = dockerfilePath
		dockerBuildConfig.Args = dockerArgs
	}
	return ciBuildConfigBean, nil
}

func mergeMap(oldDockerArgs map[string]string, ciLevelDockerArgs map[string]string) map[string]string {
	dockerArgs := make(map[string]string)
	for key, value := range oldDockerArgs {
		dockerArgs[key] = value
	}
	for key, value := range ciLevelDockerArgs {
		dockerArgs[key] = value
	}
	return dockerArgs
}

// IsLinkedCD will return if the pipelineConfig.CiPipeline is a Linked CD
func IsLinkedCD(ci pipelineConfig.CiPipeline) bool {
	return ci.ParentCiPipeline != 0 && ci.PipelineType == string(bean2.LINKED_CD)
}

// IsLinkedCI will return if the pipelineConfig.CiPipeline is a Linked CI
func IsLinkedCI(ci *pipelineConfig.CiPipeline) bool {
	if ci == nil {
		return false
	}
	return ci.ParentCiPipeline != 0 &&
		ci.PipelineType == string(bean2.LINKED)
}

// IsCIJob will return if the pipelineConfig.CiPipeline is a CI JOB
func IsCIJob(ci *pipelineConfig.CiPipeline) bool {
	if ci == nil {
		return false
	}
	return ci.PipelineType == string(bean2.CI_JOB)
}

// GetSourceCiDownStreamResponse will take the models []bean.LinkedCIDetails and []pipelineConfig.CdWorkflowRunner (for last deployment status) and generate the []CiPipeline.SourceCiDownStreamResponse
func GetSourceCiDownStreamResponse(linkedCIDetails []ciPipeline.LinkedCIDetails, latestWfrs ...pipelineConfig.CdWorkflowRunner) []bean2.SourceCiDownStreamResponse {
	response := make([]bean2.SourceCiDownStreamResponse, 0)
	cdWfrStatusMap := make(map[int]string)
	for _, latestWfr := range latestWfrs {
		cdWfrStatusMap[latestWfr.CdWorkflow.PipelineId] = latestWfr.Status
	}
	for _, item := range linkedCIDetails {
		linkedCIDetailsRes := bean2.SourceCiDownStreamResponse{
			AppName: item.AppName,
			AppId:   item.AppId,
		}
		if item.PipelineId != 0 {
			linkedCIDetailsRes.EnvironmentName = item.EnvironmentName
			linkedCIDetailsRes.EnvironmentId = item.EnvironmentId
			linkedCIDetailsRes.TriggerMode = item.TriggerMode
			linkedCIDetailsRes.DeploymentStatus = pipelineConfigBean.NotDeployed
			if status, ok := cdWfrStatusMap[item.PipelineId]; ok {
				linkedCIDetailsRes.DeploymentStatus = status
			}
		}
		response = append(response, linkedCIDetailsRes)
	}
	return response
}

func ConvertConfigDataToPipelineConfigData(r *bean.ConfigData) *pipelineConfigBean.ConfigData {
	if r != nil {
		return &pipelineConfigBean.ConfigData{
			Name:                  r.Name,
			Type:                  r.Type,
			External:              r.External,
			MountPath:             r.MountPath,
			Data:                  r.Data,
			DefaultData:           r.DefaultData,
			DefaultMountPath:      r.DefaultMountPath,
			Global:                r.Global,
			ExternalSecretType:    r.ExternalSecretType,
			ESOSecretData:         ConvertESOSecretDataToPipelineESOSecretData(r.ESOSecretData),
			DefaultESOSecretData:  ConvertESOSecretDataToPipelineESOSecretData(r.DefaultESOSecretData),
			ExternalSecret:        ConvertExternalSecretToPipelineExternalSecret(r.ExternalSecret),
			DefaultExternalSecret: ConvertExternalSecretToPipelineExternalSecret(r.DefaultExternalSecret),
			RoleARN:               r.RoleARN,
			SubPath:               r.SubPath,
			ESOSubPath:            r.ESOSubPath,
			FilePermission:        r.FilePermission,
			Overridden:            r.Overridden,
		}
	}
	return &pipelineConfigBean.ConfigData{}
}

func ConvertESOSecretDataToPipelineESOSecretData(r bean.ESOSecretData) pipelineConfigBean.ESOSecretData {
	return pipelineConfigBean.ESOSecretData{
		SecretStore:     r.SecretStore,
		SecretStoreRef:  r.SecretStoreRef,
		ESOData:         ConvertEsoDataToPipelineEsoData(r.ESOData),
		RefreshInterval: r.RefreshInterval,
	}
}

func ConvertExternalSecretToPipelineExternalSecret(r []bean.ExternalSecret) []pipelineConfigBean.ExternalSecret {
	extSec := make([]pipelineConfigBean.ExternalSecret, 0, len(r))
	for _, item := range r {
		newItem := pipelineConfigBean.ExternalSecret{
			Key:      item.Key,
			Name:     item.Name,
			Property: item.Property,
			IsBinary: item.IsBinary,
		}
		extSec = append(extSec, newItem)
	}
	return extSec
}

func ConvertEsoDataToPipelineEsoData(r []bean.ESOData) []pipelineConfigBean.ESOData {
	newEsoData := make([]pipelineConfigBean.ESOData, 0, len(r))
	for _, item := range r {
		newItem := pipelineConfigBean.ESOData{
			SecretKey: item.SecretKey,
			Key:       item.Key,
			Property:  item.Property,
		}
		newEsoData = append(newEsoData, newItem)
	}
	return newEsoData
}

// reverse adapter for the above adapters

func ConvertPipelineConfigDataToConfigData(r *pipelineConfigBean.ConfigData) *bean.ConfigData {
	if r != nil {
		return &bean.ConfigData{
			Name:                  r.Name,
			Type:                  r.Type,
			External:              r.External,
			MountPath:             r.MountPath,
			Data:                  r.Data,
			DefaultData:           r.DefaultData,
			DefaultMountPath:      r.DefaultMountPath,
			Global:                r.Global,
			ExternalSecretType:    r.ExternalSecretType,
			ESOSecretData:         ConvertPipelineESOSecretDataToESOSecretData(r.ESOSecretData),
			DefaultESOSecretData:  ConvertPipelineESOSecretDataToESOSecretData(r.DefaultESOSecretData),
			ExternalSecret:        ConvertPipelineExternalSecretToExternalSecret(r.ExternalSecret),
			DefaultExternalSecret: ConvertPipelineExternalSecretToExternalSecret(r.DefaultExternalSecret),
			RoleARN:               r.RoleARN,
			SubPath:               r.SubPath,
			ESOSubPath:            r.ESOSubPath,
			FilePermission:        r.FilePermission,
			Overridden:            r.Overridden,
		}
	}
	return &bean.ConfigData{}

}

func ConvertPipelineESOSecretDataToESOSecretData(r pipelineConfigBean.ESOSecretData) bean.ESOSecretData {
	return bean.ESOSecretData{
		SecretStore:     r.SecretStore,
		SecretStoreRef:  r.SecretStoreRef,
		ESOData:         ConvertPipelineEsoDataToEsoData(r.ESOData),
		RefreshInterval: r.RefreshInterval,
	}
}

func ConvertPipelineExternalSecretToExternalSecret(r []pipelineConfigBean.ExternalSecret) []bean.ExternalSecret {
	extSec := make([]bean.ExternalSecret, 0, len(r))
	for _, item := range r {
		newItem := bean.ExternalSecret{
			Key:      item.Key,
			Name:     item.Name,
			Property: item.Property,
			IsBinary: item.IsBinary,
		}
		extSec = append(extSec, newItem)
	}
	return extSec
}

func ConvertPipelineEsoDataToEsoData(r []pipelineConfigBean.ESOData) []bean.ESOData {
	newEsoData := make([]bean.ESOData, 0, len(r))
	for _, item := range r {
		newItem := bean.ESOData{
			SecretKey: item.SecretKey,
			Key:       item.Key,
			Property:  item.Property,
		}
		newEsoData = append(newEsoData, newItem)
	}
	return newEsoData
}

func GetStepVariableDto(variable *repository.PipelineStageStepVariable) (*pipelineConfigBean.StepVariableDto, error) {
	variableDto := &pipelineConfigBean.StepVariableDto{
		Id:                        variable.Id,
		Name:                      variable.Name,
		Format:                    variable.Format,
		Description:               variable.Description,
		AllowEmptyValue:           variable.AllowEmptyValue,
		DefaultValue:              variable.DefaultValue,
		Value:                     variable.Value,
		ValueType:                 variable.ValueType,
		PreviousStepIndex:         variable.PreviousStepIndex,
		ReferenceVariableName:     variable.ReferenceVariableName,
		ReferenceVariableStage:    variable.ReferenceVariableStage,
		VariableStepIndexInPlugin: variable.VariableStepIndexInPlugin,
	}
	return variableDto, nil
}

func NewMigrateExternalAppValidationRequest(pipeline *bean.CDPipelineConfigObject, env *repository2.Environment) *pipelineConfigBean.MigrateReleaseValidationRequest {
	request := &pipelineConfigBean.MigrateReleaseValidationRequest{
		AppId:             pipeline.AppId,
		DeploymentAppName: pipeline.DeploymentAppName,
		DeploymentAppType: pipeline.DeploymentAppType,
	}
	if pipeline.DeploymentAppType == bean3.PIPELINE_DEPLOYMENT_TYPE_ACD {
		request.ApplicationMetadataRequest = pipelineConfigBean.ApplicationMetadataRequest{
			ApplicationObjectClusterId: pipeline.ApplicationObjectClusterId,
			ApplicationObjectNamespace: pipeline.ApplicationObjectNamespace,
		}
	} else if pipeline.DeploymentAppType == bean3.PIPELINE_DEPLOYMENT_TYPE_HELM {
		request.HelmReleaseMetadataRequest = pipelineConfigBean.HelmReleaseMetadataRequest{
			ReleaseClusterId: env.ClusterId,
			ReleaseNamespace: env.Namespace,
		}
	}
	return request
}
