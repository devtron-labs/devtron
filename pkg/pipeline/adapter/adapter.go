/*
 * Copyright (c) 2024. Devtron Inc.
 */

package adapter

import (
	"encoding/json"
	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean"
	pipelineConfigBean "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean/CiPipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/constants"
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

func ConvertBuildConfigBeanToDbEntity(templateId int, overrideTemplateId int, ciBuildConfigBean *CiPipeline.CiBuildConfigBean, userId int32) (*pipelineConfig.CiBuildConfig, error) {
	buildMetadata := ""
	ciBuildType := ciBuildConfigBean.CiBuildType
	if ciBuildType == CiPipeline.BUILDPACK_BUILD_TYPE {
		buildPackConfigMetadataBytes, err := json.Marshal(ciBuildConfigBean.BuildPackConfig)
		if err != nil {
			return nil, err
		}
		buildMetadata = string(buildPackConfigMetadataBytes)
	} else if ciBuildType == CiPipeline.SELF_DOCKERFILE_BUILD_TYPE || ciBuildType == CiPipeline.MANAGED_DOCKERFILE_BUILD_TYPE {
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

func ConvertDbBuildConfigToBean(dbBuildConfig *pipelineConfig.CiBuildConfig) (*CiPipeline.CiBuildConfigBean, error) {
	var buildPackConfig *CiPipeline.BuildPackConfig
	var dockerBuildConfig *CiPipeline.DockerBuildConfig
	var err error
	if dbBuildConfig == nil {
		return nil, nil
	}
	ciBuildType := CiPipeline.CiBuildType(dbBuildConfig.Type)
	if ciBuildType == CiPipeline.BUILDPACK_BUILD_TYPE {
		buildPackConfig, err = convertMetadataToBuildPackConfig(dbBuildConfig.BuildMetadata)
		if err != nil {
			return nil, err
		}
	} else if ciBuildType == CiPipeline.SELF_DOCKERFILE_BUILD_TYPE || ciBuildType == CiPipeline.MANAGED_DOCKERFILE_BUILD_TYPE {
		dockerBuildConfig, err = convertMetadataToDockerBuildConfig(dbBuildConfig.BuildMetadata)
		if err != nil {
			return nil, err
		}
	}
	useRootBuildContext := false
	// dbBuildConfig.UseRootContext will be nil if the entry in db never updated before
	if dbBuildConfig.UseRootContext == nil || *(dbBuildConfig.UseRootContext) {
		useRootBuildContext = true
	}
	ciBuildConfigBean := &CiPipeline.CiBuildConfigBean{
		Id:                  dbBuildConfig.Id,
		CiBuildType:         ciBuildType,
		BuildPackConfig:     buildPackConfig,
		DockerBuildConfig:   dockerBuildConfig,
		UseRootBuildContext: useRootBuildContext,
	}
	return ciBuildConfigBean, nil
}

func convertMetadataToBuildPackConfig(buildConfMetadata string) (*CiPipeline.BuildPackConfig, error) {
	buildPackConfig := &CiPipeline.BuildPackConfig{}
	err := json.Unmarshal([]byte(buildConfMetadata), buildPackConfig)
	return buildPackConfig, err
}

func convertMetadataToDockerBuildConfig(dockerBuildMetadata string) (*CiPipeline.DockerBuildConfig, error) {
	dockerBuildConfig := &CiPipeline.DockerBuildConfig{}
	err := json.Unmarshal([]byte(dockerBuildMetadata), dockerBuildConfig)
	return dockerBuildConfig, err
}

func OverrideCiBuildConfig(dockerfilePath string, oldArgs string, ciLevelArgs string, dockerBuildOptions string, targetPlatform string, ciBuildConfigBean *CiPipeline.CiBuildConfigBean) (*CiPipeline.CiBuildConfigBean, error) {
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
	// no entry found in ci_build_config table, construct with requested data
	if ciBuildConfigBean == nil {
		dockerArgs := mergeMap(oldDockerArgs, ciLevelDockerArgs)
		ciBuildConfigBean = &CiPipeline.CiBuildConfigBean{
			CiBuildType: CiPipeline.SELF_DOCKERFILE_BUILD_TYPE,
			DockerBuildConfig: &CiPipeline.DockerBuildConfig{
				DockerfilePath:     dockerfilePath,
				Args:               dockerArgs,
				TargetPlatform:     targetPlatform,
				DockerBuildOptions: dockerBuildOptionsMap,
				BuildContext:       "",
			},
			// setting true as default
			UseRootBuildContext: true,
		}
	} else if ciBuildConfigBean.CiBuildType == CiPipeline.SELF_DOCKERFILE_BUILD_TYPE || ciBuildConfigBean.CiBuildType == CiPipeline.MANAGED_DOCKERFILE_BUILD_TYPE {
		dockerBuildConfig := ciBuildConfigBean.DockerBuildConfig
		dockerArgs := mergeMap(dockerBuildConfig.Args, ciLevelDockerArgs)
		// dockerBuildConfig.DockerfilePath = dockerfilePath
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
func IsLinkedCD(ci *pipelineConfig.CiPipeline) bool {
	if ci == nil {
		return false
	}
	return ci.ParentCiPipeline != 0 && ci.PipelineType == string(constants.LINKED_CD)
}

// IsLinkedCI will return if the pipelineConfig.CiPipeline is a Linked CI
func IsLinkedCI(ci *pipelineConfig.CiPipeline) bool {
	if ci == nil {
		return false
	}
	return ci.ParentCiPipeline != 0 &&
		ci.PipelineType == string(constants.LINKED)
}

// IsCIJob will return if the pipelineConfig.CiPipeline is a CI JOB
func IsCIJob(ci *pipelineConfig.CiPipeline) bool {
	if ci == nil {
		return false
	}
	return ci.PipelineType == string(constants.CI_JOB)
}

// GetSourceCiDownStreamResponse will take the models []bean.LinkedCIDetails and []pipelineConfig.CdWorkflowRunner (for last deployment status) and generate the []CiPipeline.SourceCiDownStreamResponse
func GetSourceCiDownStreamResponse(linkedCIDetails []bean.LinkedCIDetails, latestWfrs ...pipelineConfig.CdWorkflowRunner) []CiPipeline.SourceCiDownStreamResponse {
	response := make([]CiPipeline.SourceCiDownStreamResponse, 0)
	cdWfrStatusMap := make(map[int]string)
	for _, latestWfr := range latestWfrs {
		cdWfrStatusMap[latestWfr.CdWorkflow.PipelineId] = latestWfr.Status
	}
	for _, item := range linkedCIDetails {
		linkedCIDetailsRes := CiPipeline.SourceCiDownStreamResponse{
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
