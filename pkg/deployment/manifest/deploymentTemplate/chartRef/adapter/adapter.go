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
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"
)

func ConvertChartRefDbObjToBean(ref *chartRepoRepository.ChartRef) *bean.ChartRefDto {
	dto := &bean.ChartRefDto{
		Id:                     ref.Id,
		Location:               ref.Location,
		Version:                ref.Version,
		Default:                ref.Default,
		ChartData:              ref.ChartData,
		ChartDescription:       ref.ChartDescription,
		UserUploaded:           ref.UserUploaded,
		IsAppMetricsSupported:  ref.IsAppMetricsSupported,
		DeploymentStrategyPath: ref.DeploymentStrategyPath,
		JsonPathForStrategy:    ref.JsonPathForStrategy,
	}
	if len(ref.Name) == 0 {
		dto.Name = bean.RolloutChartType
	} else {
		dto.Name = ref.Name
	}
	return dto
}

func ConvertCustomChartRefDtoToDbObj(ref *bean.CustomChartRefDto) *chartRepoRepository.ChartRef {
	return &chartRepoRepository.ChartRef{
		Id:                     ref.Id,
		Location:               ref.Location,
		Version:                ref.Version,
		Active:                 ref.Active,
		Default:                ref.Default,
		Name:                   ref.Name,
		ChartData:              ref.ChartData,
		ChartDescription:       ref.ChartDescription,
		UserUploaded:           ref.UserUploaded,
		IsAppMetricsSupported:  ref.IsAppMetricsSupported,
		DeploymentStrategyPath: ref.DeploymentStrategyPath,
		JsonPathForStrategy:    ref.JsonPathForStrategy,
		AuditLog:               ref.AuditLog,
	}
}
