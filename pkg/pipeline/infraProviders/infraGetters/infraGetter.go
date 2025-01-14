/*
 * Copyright (c) 2024. Devtron Inc.
 */

package infraGetters

import (
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
)

type InfraGetter interface {
	GetConfigurationsByScopeAndTargetPlatforms(request *InfraRequest) (map[string]*v1.InfraConfig, error)
	SaveInfraConfigHistorySnapshot(workflowId int, triggeredBy int32, infraConfigs map[string]*v1.InfraConfig) error
}

type InfraRequest struct {
	scope           resourceQualifiers.Scope
	appId           int
	envId           int
	targetPlatforms []string
}

func NewInfraRequest(scope resourceQualifiers.Scope) *InfraRequest {
	return &InfraRequest{
		scope: scope,
	}
}

func (infraRequest *InfraRequest) WithAppId(appId int) *InfraRequest {
	infraRequest.appId = appId
	return infraRequest
}

func (infraRequest *InfraRequest) WithEnvId(envId int) *InfraRequest {
	infraRequest.envId = envId
	return infraRequest
}

func (infraRequest *InfraRequest) WithPlatform(platforms ...string) *InfraRequest {
	for _, platform := range platforms {
		if platform == "" {
			platform = v1.RUNNER_PLATFORM
		} else {
			infraRequest.targetPlatforms = append(infraRequest.targetPlatforms, platform)
		}
	}
	return infraRequest
}

func (infraRequest *InfraRequest) GetCiScope() *v1.Scope {
	return &v1.Scope{
		AppId: infraRequest.appId,
	}
}

func (infraRequest *InfraRequest) GetWorkflowScope() resourceQualifiers.Scope {
	return infraRequest.scope
}

func (infraRequest *InfraRequest) GetAppId() int {
	return infraRequest.appId
}

func (infraRequest *InfraRequest) GetEnvId() int {
	return infraRequest.envId
}

func (infraRequest *InfraRequest) GetTargetPlatforms() []string {
	return infraRequest.targetPlatforms
}

func GetInfraConfigScope(workflowScope resourceQualifiers.Scope) *v1.Scope {
	return &v1.Scope{
		AppId: workflowScope.AppId,
	}
}
