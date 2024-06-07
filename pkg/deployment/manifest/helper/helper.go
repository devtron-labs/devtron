/*
 * Copyright (c) 2024. Devtron Inc.
 */

package helper

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	bean3 "github.com/devtron-labs/devtron/pkg/deployment/manifest/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	util4 "github.com/devtron-labs/devtron/util"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func ResolveDeploymentTypeAndUpdate(overrideRequest *bean.ValuesOverrideRequest) {
	if overrideRequest.DeploymentType == models.DEPLOYMENTTYPE_UNKNOWN {
		overrideRequest.DeploymentType = models.DEPLOYMENTTYPE_DEPLOY
	}
	if len(overrideRequest.DeploymentWithConfig) == 0 {
		overrideRequest.DeploymentWithConfig = bean.DEPLOYMENT_CONFIG_TYPE_LAST_SAVED
	}
}

func GetDeploymentTemplateType(overrideRequest *bean.ValuesOverrideRequest) chartRepoRepository.DeploymentStrategy {
	var deploymentTemplate chartRepoRepository.DeploymentStrategy
	if overrideRequest.DeploymentTemplate == "ROLLING" {
		deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_ROLLING
	} else if overrideRequest.DeploymentTemplate == "BLUE-GREEN" {
		deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_BLUE_GREEN
	} else if overrideRequest.DeploymentTemplate == "CANARY" {
		deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_CANARY
	} else if overrideRequest.DeploymentTemplate == "RECREATE" {
		deploymentTemplate = chartRepoRepository.DEPLOYMENT_STRATEGY_RECREATE
	}
	return deploymentTemplate
}

func ExtractParamValue(inputMap map[string]interface{}, key string, merged []byte) (float64, error) {
	if _, ok := inputMap[key]; !ok {
		return 0, errors.New("empty-val-err")
	}
	return util4.ParseFloatNumber(gjson.Get(string(merged), inputMap[key].(string)).Value())
}

func SetScalingValues(templateMap map[string]interface{}, customScalingKey string, merged []byte, value interface{}) ([]byte, error) {
	autoscalingJsonPath := templateMap[customScalingKey]
	autoscalingJsonPathKey := autoscalingJsonPath.(string)
	mergedRes, err := sjson.Set(string(merged), autoscalingJsonPathKey, value)
	if err != nil {
		return []byte{}, err
	}
	return []byte(mergedRes), nil
}

func FetchRequiredReplicaCount(currentReplicaCount float64, reqMaxReplicas float64, reqMinReplicas float64) float64 {
	var reqReplicaCount float64
	if currentReplicaCount <= reqMaxReplicas && currentReplicaCount >= reqMinReplicas {
		reqReplicaCount = currentReplicaCount
	} else if currentReplicaCount > reqMaxReplicas {
		reqReplicaCount = reqMaxReplicas
	} else if currentReplicaCount < reqMinReplicas {
		reqReplicaCount = reqMinReplicas
	}
	return reqReplicaCount
}

func GetAutoScalingReplicaCount(templateMap map[string]interface{}, appName string) *util4.HpaResourceRequest {
	hasOverride := false
	if _, ok := templateMap[bean3.FullnameOverride]; ok {
		appNameOverride := templateMap[bean3.FullnameOverride].(string)
		if len(appNameOverride) > 0 {
			appName = appNameOverride
			hasOverride = true
		}
	}
	if !hasOverride {
		if _, ok := templateMap[bean3.NameOverride]; ok {
			nameOverride := templateMap[bean3.NameOverride].(string)
			if len(nameOverride) > 0 {
				appName = fmt.Sprintf("%s-%s", appName, nameOverride)
			}
		}
	}
	hpaResourceRequest := &util4.HpaResourceRequest{}
	hpaResourceRequest.Version = ""
	hpaResourceRequest.Group = autoscaling.ServiceName
	hpaResourceRequest.Kind = bean3.HorizontalPodAutoscaler
	if _, ok := templateMap[bean3.KedaAutoscaling]; ok {
		as := templateMap[bean3.KedaAutoscaling]
		asd := as.(map[string]interface{})
		if _, ok := asd[bean3.Enabled]; ok {
			enable := asd[bean3.Enabled].(bool)
			if enable {
				hpaResourceRequest.IsEnable = enable
				hpaResourceRequest.ReqReplicaCount = templateMap[bean3.ReplicaCount].(float64)
				hpaResourceRequest.ReqMaxReplicas = asd["maxReplicaCount"].(float64)
				hpaResourceRequest.ReqMinReplicas = asd["minReplicaCount"].(float64)
				hpaResourceRequest.ResourceName = fmt.Sprintf("%s-%s-%s", "keda-hpa", appName, "keda")
				return hpaResourceRequest
			}
		}
	}

	if _, ok := templateMap[autoscaling.ServiceName]; ok {
		as := templateMap[autoscaling.ServiceName]
		asd := as.(map[string]interface{})
		if _, ok := asd[bean3.Enabled]; ok {
			enable := asd[bean3.Enabled].(bool)
			if enable {
				hpaResourceRequest.IsEnable = asd[bean3.Enabled].(bool)
				hpaResourceRequest.ReqReplicaCount = templateMap[bean3.ReplicaCount].(float64)
				hpaResourceRequest.ReqMaxReplicas = asd["MaxReplicas"].(float64)
				hpaResourceRequest.ReqMinReplicas = asd["MinReplicas"].(float64)
				hpaResourceRequest.ResourceName = fmt.Sprintf("%s-%s", appName, "hpa")
				return hpaResourceRequest
			}
		}
	}
	return hpaResourceRequest

}

func CreateConfigMapAndSecretJsonRequest(overrideRequest *bean.ValuesOverrideRequest, envOverride *chartConfig.EnvConfigOverride, chartVersion string, scope resourceQualifiers.Scope) bean3.ConfigMapAndSecretJsonV2 {
	request := bean3.ConfigMapAndSecretJsonV2{
		AppId:                                 overrideRequest.AppId,
		EnvId:                                 envOverride.TargetEnvironment,
		PipeLineId:                            overrideRequest.PipelineId,
		ChartVersion:                          chartVersion,
		DeploymentWithConfig:                  overrideRequest.DeploymentWithConfig,
		WfrIdForDeploymentWithSpecificTrigger: overrideRequest.WfrIdForDeploymentWithSpecificTrigger,
		Scope:                                 scope,
	}
	return request
}

func GetScopeForVariables(overrideRequest *bean.ValuesOverrideRequest, envOverride *chartConfig.EnvConfigOverride) resourceQualifiers.Scope {
	scope := resourceQualifiers.Scope{
		AppId:     overrideRequest.AppId,
		EnvId:     envOverride.TargetEnvironment,
		ClusterId: envOverride.Environment.Id,
		SystemMetadata: &resourceQualifiers.SystemMetadata{
			EnvironmentName: envOverride.Environment.Name,
			ClusterName:     envOverride.Environment.Cluster.ClusterName,
			Namespace:       envOverride.Environment.Namespace,
			ImageTag:        util3.GetImageTagFromImage(overrideRequest.Image),
			AppName:         overrideRequest.AppName,
			Image:           overrideRequest.Image,
		},
	}
	return scope
}
